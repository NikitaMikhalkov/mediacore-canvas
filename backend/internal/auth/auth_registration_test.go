package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"yingce/backend/internal/model"

	"gorm.io/gorm"
)

func newRegistrationTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	svc, db := newPasswordResetTestService(t)
	if err := db.AutoMigrate(&model.UserIdentity{}, &model.OAuthState{}, &model.CreditAccount{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return svc, db
}

func TestVerificationPolicyDefaultsAndSavedSettings(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	p, err := svc.verificationPolicy()
	if err != nil || !p.SMSLogin || !p.SMSRegistration || !p.EmailRegistration {
		t.Fatalf("unexpected default policy: %+v, %v", p, err)
	}
	if err := db.Create(&model.SystemSetting{Key: authPolicyKey, ValueJSON: `{"emailLogin":true}`}).Error; err != nil {
		t.Fatal(err)
	}
	p, err = svc.verificationPolicy()
	if err != nil || !p.EmailLogin || p.SMSLogin || p.SMSRegistration {
		t.Fatalf("saved policy changed: %+v, %v", p, err)
	}
	if _, err := svc.UpdateVerificationPolicy(&model.User{ID: "admin"}, VerificationPolicy{SMSLogin: true}); err == nil {
		t.Fatal("unconfigured SMS was accepted")
	}
}

type registrationTestHost struct{ nopHost }

func (registrationTestHost) SettingsEncryptionKey() ([]byte, error) {
	return []byte("test-only-encryption-key-32-bytes"), nil
}

type registrationTestSMS struct{ code string }

func (*registrationTestSMS) Available(string) (bool, error) { return true, nil }
func (s *registrationTestSMS) SendCode(_ context.Context, _, _, code string) error {
	s.code = code
	return nil
}

func TestRegisterVerifiedContactWithoutUsernameOrPassword(t *testing.T) {
	for _, method := range []string{"sms", "email"} {
		t.Run(method, func(t *testing.T) {
			svc, db := newRegistrationTestService(t)
			if err := db.AutoMigrate(&model.AuthVerification{}, &model.NotificationQuota{}); err != nil {
				t.Fatal(err)
			}
			svc.host = registrationTestHost{}
			sms := &registrationTestSMS{}
			svc.SetSMSDelivery(sms)
			if err := db.Create(&model.User{ID: "admin", Username: "admin", Role: model.UserRoleAdmin, Status: model.UserStatusActive}).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&model.SystemSetting{Key: registrationSettingKey, ValueJSON: `{"enabled":true}`}).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&model.SystemSetting{Key: authPolicyKey, ValueJSON: `{"smsLogin":true,"emailLogin":true,"smsRegistration":true,"emailRegistration":true}`}).Error; err != nil {
				t.Fatal(err)
			}
			var emailCode string
			svc.SetMailSender(func(_ EmailSettingValue, _, _, body string) error {
				_, rest, _ := strings.Cut(body, "验证码是：")
				emailCode, _, _ = strings.Cut(rest, "\n")
				return nil
			})
			req := VerificationRequest{Purpose: "register", Method: method}
			if method == "sms" {
				req.Phone = "13800138000"
			} else {
				req.Email = "member@example.com"
			}
			ticket, err := svc.StartVerification(context.Background(), nil, req)
			if err != nil {
				t.Fatal(err)
			}
			result, err := svc.Register(RegisterRequest{Phone: req.Phone, Email: req.Email, Ticket: ticket.Ticket, SMSCode: sms.code, EmailCode: emailCode, AcceptedTerms: true})
			if err != nil {
				t.Fatal(err)
			}
			if result.Session == "" || result.User.Username == "" || result.User.PasswordHash != "" {
				t.Fatalf("unexpected registration: %+v", result.User)
			}
			if method == "sms" && result.User.PhoneVerifiedAt == nil {
				t.Fatal("phone not marked verified")
			}
			if method == "email" && result.User.EmailVerifiedAt == nil {
				t.Fatal("email not marked verified")
			}
		})
	}
}

func TestRegistrationWithoutLoginMethodRequiresPassword(t *testing.T) {
	for _, method := range []string{"sms", "email"} {
		t.Run(method, func(t *testing.T) {
			svc, db := newRegistrationTestService(t)
			if err := db.AutoMigrate(&model.AuthVerification{}, &model.NotificationQuota{}); err != nil {
				t.Fatal(err)
			}
			svc.host = registrationTestHost{}
			delivery := &registrationTestSMS{}
			svc.SetSMSDelivery(delivery)
			if err := db.Create(&model.User{ID: "admin", Username: "admin", Status: model.UserStatusActive}).Error; err != nil {
				t.Fatal(err)
			}
			for _, setting := range []model.SystemSetting{{Key: registrationSettingKey, ValueJSON: `{"enabled":true}`}, {Key: authPolicyKey, ValueJSON: `{"smsRegistration":true,"emailRegistration":true}`}} {
				if err := db.Create(&setting).Error; err != nil {
					t.Fatal(err)
				}
			}
			var emailCode string
			svc.SetMailSender(func(_ EmailSettingValue, _, _, body string) error {
				_, rest, _ := strings.Cut(body, "验证码是：")
				emailCode, _, _ = strings.Cut(rest, "\n")
				return nil
			})
			request := VerificationRequest{Purpose: "register", Method: method}
			if method == "sms" {
				request.Phone = "13800138000"
			} else {
				request.Email = "member@example.com"
			}
			ticket, err := svc.StartVerification(context.Background(), nil, request)
			if err != nil {
				t.Fatal(err)
			}
			reg := RegisterRequest{Phone: request.Phone, Email: request.Email, Ticket: ticket.Ticket, SMSCode: delivery.code, EmailCode: emailCode, AcceptedTerms: true}
			if _, err := svc.Register(reg); err == nil {
				t.Fatal("passwordless registration accepted without a login method")
			}
			reg.Password = "password123"
			if _, err := svc.Register(reg); err != nil {
				t.Fatalf("password registration rejected: %v", err)
			}
		})
	}
}

func TestVerificationLoginUnknownPhoneReturnsSameReceipt(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	if err := db.AutoMigrate(&model.AuthVerification{}, &model.NotificationQuota{}); err != nil {
		t.Fatal(err)
	}
	svc.host = registrationTestHost{}
	delivery := &registrationTestSMS{}
	svc.SetSMSDelivery(delivery)
	now := time.Now()
	if err := db.Create(&model.User{ID: "member", Username: "member", Phone: "+8613800138000", PhoneVerifiedAt: &now, Status: model.UserStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	known, err := svc.StartVerification(context.Background(), nil, VerificationRequest{Purpose: "login", Method: "sms", Phone: "13800138000"})
	if err != nil || delivery.code == "" {
		t.Fatalf("known phone: %v", err)
	}
	delivery.code = ""
	unknown, err := svc.StartVerification(context.Background(), nil, VerificationRequest{Purpose: "login", Method: "sms", Phone: "13900139000"})
	if err != nil || unknown.Ticket == "" || known.ExpiresIn != unknown.ExpiresIn || known.RetryAfter != unknown.RetryAfter || delivery.code != "" {
		t.Fatalf("unknown phone receipt differs: %+v, %v", unknown, err)
	}
	if _, err := svc.LoginVerification(VerificationConfirm{Ticket: unknown.Ticket, SMSCode: "123456"}); err == nil {
		t.Fatal("unknown phone logged in")
	}
}

func TestFirstAdministratorRequiresPassword(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	if _, err := svc.Register(RegisterRequest{Username: "admin", AcceptedTerms: true}); err == nil {
		t.Fatal("first administrator created without password or verification")
	}
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unexpected user count %d: %v", count, err)
	}
}

func TestPhoneLoginDoesNotSelectCollidingUsername(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, user := range []model.User{
		{ID: "owner", Username: "owner", Phone: "+8613800138000", PhoneVerifiedAt: &now, PasswordHash: hash, Status: model.UserStatusActive},
		{ID: "other", Username: "13800138000", PasswordHash: hash, Status: model.UserStatusActive},
	} {
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	result, err := svc.Login(LoginRequest{Username: "13800138000", Password: "password123"})
	if err != nil || result.User.ID != "owner" {
		t.Fatalf("phone login selected wrong account: result=%+v err=%v", result, err)
	}
	if subject, err := svc.LoginRateLimitSubject("13800138000"); err != nil || subject != "user:owner" {
		t.Fatalf("rate limit subject: %q %v", subject, err)
	}
	if err := ValidateUsername("13800138000"); err == nil {
		t.Fatal("phone-shaped username accepted")
	}
}

func TestDisabledPhoneDoesNotRevealStatusWithoutPassword(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := db.Create(&model.User{ID: "disabled", Username: "disabled", Phone: "+8613800138000", PhoneVerifiedAt: &now, PasswordHash: hash, Status: model.UserStatusDisabled}).Error; err != nil {
		t.Fatal(err)
	}
	_, disabledErr := svc.Login(LoginRequest{Username: "13800138000", Password: "wrong-password"})
	_, missingErr := svc.Login(LoginRequest{Username: "13900139000", Password: "wrong-password"})
	var disabled, missing *AuthError
	if !errors.As(disabledErr, &disabled) || !errors.As(missingErr, &missing) || disabled.Status != 401 || disabled.Status != missing.Status || disabled.Message != missing.Message {
		t.Fatalf("login failures differ: disabled=%v missing=%v", disabledErr, missingErr)
	}
	if _, err := svc.Login(LoginRequest{Username: "13800138000", Password: "password123"}); err == nil {
		t.Fatal("disabled account logged in")
	}
}

func TestPasswordLoginSupportsPhone(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.User{ID: "member", Username: "member", Phone: "+8613800138000", PasswordHash: hash, Status: model.UserStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	result, err := svc.Login(LoginRequest{Username: "13800138000", Password: "password123"})
	if err != nil || result.Session == "" {
		t.Fatalf("phone password login: %v", err)
	}
}

func TestRegisterRequiresAcceptedTermsBeforeWriting(t *testing.T) {
	for _, payload := range []string{
		`{"username":"new-user","password":"password"}`,
		`{"username":"new-user","password":"password","acceptedTerms":false}`,
	} {
		svc, db := newRegistrationTestService(t)
		var req RegisterRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatal(err)
		}
		_, err := svc.Register(req)
		var authErr *AuthError
		if !errors.As(err, &authErr) || authErr.Status != 400 || authErr.Message != "请先同意Mediacore服务协议" {
			t.Fatalf("Register() error = %v", err)
		}
		for _, entity := range []any{&model.User{}, &model.AuthSession{}} {
			var count int64
			if err := db.Model(entity).Count(&count).Error; err != nil || count != 0 {
				t.Fatalf("unaccepted registration wrote %T: count=%d err=%v", entity, count, err)
			}
		}
	}
}

func TestRegisterAcceptedTermsCreatesFirstAdmin(t *testing.T) {
	svc, _ := newRegistrationTestService(t)
	result, err := svc.Register(RegisterRequest{Username: "new-user", Password: "password", AcceptedTerms: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.User.Username != "new-user" || result.User.Role != model.UserRoleAdmin || result.Session == "" {
		t.Fatalf("Register() result = %#v", result)
	}
}

func TestRegisterAcceptedTermsCreatesEmailUser(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	if err := db.Create(&model.User{ID: "admin", Username: "admin", Role: model.UserRoleAdmin, Status: model.UserStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SystemSetting{Key: registrationSettingKey, ValueJSON: `{"enabled":true}`}).Error; err != nil {
		t.Fatal(err)
	}
	var code string
	svc.SetMailSender(func(_ EmailSettingValue, _, _ string, body string) error {
		code = codeFromEmailBody(body)
		return nil
	})
	if err := svc.SendRegistrationEmailCode("member@example.com"); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Register(RegisterRequest{Username: "member", Email: "member@example.com", EmailCode: code, Password: "password", AcceptedTerms: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.User.Role != model.UserRoleUser || result.Session == "" {
		t.Fatalf("Register() result = %#v", result)
	}
}
