package auth

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"yingce/backend/internal/model"
)

// A passwordless user with a verified phone and captured SMS, so tests can drive
// OTP reauthentication without any external delivery.
func passwordlessProfile(t *testing.T) (*Service, *gorm.DB, *model.User, string, *registrationTestSMS) {
	t.Helper()
	s, db, u, cookie := newProfileTestService(t)
	now := time.Now().UTC()
	if err := db.Model(u).Updates(map[string]any{"password_hash": "", "phone": "+8613800138000", "phone_verified_at": now}).Error; err != nil {
		t.Fatal(err)
	}
	u.PasswordHash, u.Phone, u.PhoneVerifiedAt = "", "+8613800138000", &now
	sms := &registrationTestSMS{}
	s.SetSMSDelivery(sms)
	s.SetMailSender(func(EmailSettingValue, string, string, string) error { return nil })
	return s, db, u, cookie, sms
}

// The OTP marker only authorizes set-password. Email change takes the current
// password alone (see the repository test for the transaction-level check).
func TestProfileOTPMarkerAuthorizesPasswordButNotEmailChange(t *testing.T) {
	s, _, u, cookie, sms := passwordlessProfile(t)
	ticket, err := s.StartProfileReauthentication(context.Background(), u, cookie, "sms")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmProfileReauthentication(u, cookie, VerificationConfirm{Ticket: ticket.Ticket, SMSCode: sms.code}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SendProfileEmailCode(context.Background(), u, cookie, ProfileEmailCodeRequest{Email: "new@external.example"}); err == nil {
		t.Fatal("OTP marker authorized sending an email-change code")
	}
	if _, err := s.ChangeProfileEmail(u, cookie, ProfileEmailChangeRequest{Ticket: "x.y"}); err == nil {
		t.Fatal("OTP marker authorized an email change")
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err != nil {
		t.Fatalf("OTP marker no longer authorizes set-password: %v", err)
	}
}

// A bare marker with no OTP challenge behind it must not authorize email change either.
func TestProfileEmailChangeIgnoresBareMarker(t *testing.T) {
	s, db, u, cookie, _ := passwordlessProfile(t)
	db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("reauthenticated_at", time.Now())
	if _, err := s.SendProfileEmailCode(context.Background(), u, cookie, ProfileEmailCodeRequest{Email: "new@external.example"}); err == nil {
		t.Fatal("bare marker authorized sending an email-change code")
	}
}
