package auth

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"yingce/backend/internal/model"
)

func TestProfileOTPReauthenticationSetsPassword(t *testing.T) {
	s, db, user, cookie := newProfileTestService(t)
	now := time.Now().UTC()
	if err := db.Model(user).Updates(map[string]any{"password_hash": "", "phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	user.PasswordHash, user.Phone, user.PhoneVerifiedAt = "", "+8613800138000", &now
	delivery := &registrationTestSMS{}
	s.SetSMSDelivery(delivery)
	if err := s.ChangeProfilePassword(user, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("password changed without reauthentication")
	}
	ticket, err := s.StartProfileReauthentication(context.Background(), user, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	req := VerificationConfirm{Ticket: ticket.Ticket, SMSCode: delivery.code}
	if err := s.ConfirmProfileReauthentication(user, "session-two.token", req); err == nil {
		t.Fatal("another session used the ticket")
	}
	if err := s.ConfirmProfileReauthentication(user, cookie, req); err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmProfileReauthentication(user, cookie, req); err == nil {
		t.Fatal("replayed ticket accepted")
	}
	if err := s.ChangeProfilePassword(user, "session-two.token", ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("another session used reauthentication")
	}
	if err := s.ChangeProfilePassword(user, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err != nil {
		t.Fatal(err)
	}
	var remaining int64
	if err := db.Model(&model.AuthSession{}).Where("user_id = ?", user.ID).Count(&remaining).Error; err != nil || remaining != 0 {
		t.Fatalf("sessions after password change: %d, %v", remaining, err)
	}
	if _, err := s.Login(LoginRequest{Username: user.Username, Password: "new-password"}); err != nil {
		t.Fatalf("new password did not work: %v", err)
	}
}

func TestProfileOTPRejectsChangedState(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*testing.T, *gorm.DB, *model.User, *VerificationConfirm)
	}{
		{"wrong code", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			if r.SMSCode[0] == '0' {
				r.SMSCode = "1" + r.SMSCode[1:]
			} else {
				r.SMSCode = "0" + r.SMSCode[1:]
			}
		}},
		{"expired", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(&model.AuthVerification{}).Where("id = ?", strings.Split(r.Ticket, ".")[0]).Update("expires_at", time.Now().Add(-time.Minute))
		}},
		{"revoked session", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Where("id = ?", "session-one").Delete(&model.AuthSession{})
		}},
		{"expired session", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("expires_at", time.Now().Add(-time.Minute))
		}},
		{"disabled", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(u).Update("status", model.UserStatusDisabled)
		}},
		{"changed phone", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(u).Update("phone", "+8613900139000")
		}},
		{"unverified phone", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(u).Update("phone_verified_at", nil)
		}},
		{"new verification", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(u).Update("phone_verified_at", time.Now().Add(time.Minute))
		}},
		{"password changed", func(t *testing.T, db *gorm.DB, u *model.User, r *VerificationConfirm) {
			db.Model(u).Update("password_hash", "changed")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, db, u, cookie := newProfileTestService(t)
			now := time.Now().UTC()
			if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
				t.Fatal(err)
			}
			sms := &registrationTestSMS{}
			s.SetSMSDelivery(sms)
			ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
			if err != nil {
				t.Fatal(err)
			}
			req := VerificationConfirm{Ticket: ticket.Ticket, SMSCode: sms.code}
			tc.mutate(t, db, u, &req)
			if err := s.ConfirmProfileReauthentication(u, cookie, req); err == nil {
				t.Fatal("invalid challenge accepted")
			}
			var count int64
			if err := db.Model(&model.AuthSession{}).Where("reauthenticated_at IS NOT NULL").Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("unexpected marker count %d: %v", count, err)
			}
		})
	}
}

func TestProfileOTPConcurrentConfirmationConsumesOnce(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	delivery := &registrationTestSMS{}
	s.SetSMSDelivery(delivery)
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	req := VerificationConfirm{Ticket: ticket.Ticket, SMSCode: delivery.code}
	results := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; results <- s.ConfirmProfileReauthentication(u, cookie, req) }()
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful confirmations = %d", success)
	}
	var count int64
	if err := db.Model(&model.AuthSession{}).Where("reauthenticated_at IS NOT NULL").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("marker count %d: %v", count, err)
	}
}

func TestProfileOTPMarkerFailureRollsBackConsumption(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	delivery := &registrationTestSMS{}
	s.SetSMSDelivery(delivery)
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER reject_reauth BEFORE UPDATE OF reauthenticated_at ON auth_sessions BEGIN SELECT RAISE(FAIL, 'test'); END`).Error; err != nil {
		t.Fatal(err)
	}
	req := VerificationConfirm{Ticket: ticket.Ticket, SMSCode: delivery.code}
	if err := s.ConfirmProfileReauthentication(u, cookie, req); err == nil {
		t.Fatal("marker failure accepted")
	}
	var v model.AuthVerification
	if err := db.First(&v, "id = ?", strings.Split(ticket.Ticket, ".")[0]).Error; err != nil || v.UsedAt != nil {
		t.Fatalf("consumption not rolled back: %v", err)
	}
	var count int64
	if err := db.Model(&model.AuthSession{}).Where("reauthenticated_at IS NOT NULL").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("marker survived failed transaction: %d, %v", count, err)
	}
	if err := db.Exec("DROP TRIGGER reject_reauth").Error; err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, req); err != nil {
		t.Fatalf("retry after rollback: %v", err)
	}
}

func TestProfileOTPEmailAndPasswordExpiry(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"password_hash": "", "email": "Old@Example.com", "email_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	var code string
	s.SetMailSender(func(_ EmailSettingValue, _, _, body string) error {
		_, rest, _ := strings.Cut(body, "验证码是：")
		code, _, _ = strings.Cut(rest, "\n")
		return nil
	})
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "email")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: ticket.Ticket, EmailCode: code}); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("reauthenticated_at", time.Now().Add(-6*time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("expired marker accepted")
	}
	if err := db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("reauthenticated_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err != nil {
		t.Fatal(err)
	}
}

type profileSceneSMS struct{ available, sent, code, target string }

func (s *profileSceneSMS) Available(scene string) (bool, error) {
	s.available = scene
	return scene == "login", nil
}
func (s *profileSceneSMS) SendCode(_ context.Context, scene, target, code string) error {
	s.sent, s.target, s.code = scene, target, code
	return nil
}
func TestProfileOTPUsesCurrentVerifiedContactAndLoginSMSScene(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	delivery := &profileSceneSMS{}
	s.SetSMSDelivery(delivery)
	if err := db.Create(&model.SystemSetting{Key: authPolicyKey, ValueJSON: `{}`}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms"); err == nil {
		t.Fatal("unverified phone accepted")
	}
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartProfileReauthentication(context.Background(), u, "", "sms"); err == nil {
		t.Fatal("missing session accepted")
	}
	if _, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms"); err != nil {
		t.Fatal(err)
	}
	if delivery.available != "login" || delivery.sent != "login" || delivery.target != "+8613800138000" || len(delivery.code) != 6 {
		t.Fatal("wrong SMS scene or destination")
	}
	if _, err := s.StartVerification(context.Background(), u, VerificationRequest{Purpose: "profile_reauth", Method: "sms", Phone: "+8613800138000"}); err == nil {
		t.Fatal("public reauth purpose accepted")
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: "invalid", SMSCode: delivery.code}); err == nil {
		t.Fatal("invalid ticket accepted")
	}
}

func TestProfileOTPDoesNotReplaceCurrentPassword(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	delivery := &registrationTestSMS{}
	s.SetSMSDelivery(delivery)
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: ticket.Ticket, SMSCode: delivery.code}); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("OTP bypassed current password")
	}
}

type unavailableProfileSMS struct{}

func (unavailableProfileSMS) Available(string) (bool, error) { return false, nil }
func (unavailableProfileSMS) SendCode(context.Context, string, string, string) error {
	panic("disabled route sent a code")
}
func TestProfileOTPUnavailableAndAttemptLimit(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	now := time.Now()
	if err := db.Model(u).Updates(map[string]any{"phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	s.SetSMSDelivery(unavailableProfileSMS{})
	if _, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms"); err == nil {
		t.Fatal("unavailable delivery accepted")
	}
	var count int64
	if err := db.Model(&model.AuthVerification{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unexpected ticket count %d: %v", count, err)
	}
	delivery := &registrationTestSMS{}
	s.SetSMSDelivery(delivery)
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	wrong := "0" + delivery.code[1:]
	if wrong == delivery.code {
		wrong = "1" + delivery.code[1:]
	}
	for i := 0; i < 5; i++ {
		if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: ticket.Ticket, SMSCode: wrong}); err == nil {
			t.Fatal("wrong code accepted")
		}
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: ticket.Ticket, SMSCode: delivery.code}); err == nil {
		t.Fatal("exhausted ticket accepted")
	}
	if err := db.Model(&model.AuthSession{}).Where("reauthenticated_at IS NOT NULL").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unexpected markers %d: %v", count, err)
	}
}
