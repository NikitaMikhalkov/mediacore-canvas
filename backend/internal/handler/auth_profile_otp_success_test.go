package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"yingce/backend/internal/model"
)

// otpCaptureSMS 记录验证码以代替真实短信；验证码只留在测试内存里。
type otpCaptureSMS struct{ code string }

func (*otpCaptureSMS) Available(string) (bool, error) { return true, nil }
func (s *otpCaptureSMS) SendCode(_ context.Context, _, _, code string) error {
	s.code = code
	return nil
}

func TestProfileOTPReauthenticationHTTPSuccessThenSetPassword(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	sms := &otpCaptureSMS{}
	runtimeService.SetAuthSMSDelivery(sms)
	var logs bytes.Buffer
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	if err := db.Model(&model.User{}).Where("id = ?", "profile-user").Updates(map[string]any{"password_hash": "", "phone": "+8613800138000", "phone_verified_at": time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	start := profileRequest(router, "POST", "/api/auth/profile/reauth-code", `{"method":"sms"}`, cookie)
	if start.Code != http.StatusOK || sms.code == "" {
		t.Fatalf("start status %d: %s (code captured: %v)", start.Code, start.Body.String(), sms.code != "")
	}
	var started struct {
		Data struct {
			Ticket string `json:"ticket"`
		} `json:"data"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &started); err != nil || started.Data.Ticket == "" {
		t.Fatalf("start body %s: %v", start.Body.String(), err)
	}
	code := sms.code
	confirm := profileRequest(router, "POST", "/api/auth/profile/reauth", `{"ticket":"`+started.Data.Ticket+`","smsCode":"`+code+`"}`, cookie)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm status %d: %s", confirm.Code, confirm.Body.String())
	}
	profile := profileRequest(router, "GET", "/api/auth/profile", "", cookie)
	if profile.Code != http.StatusOK || !strings.Contains(profile.Body.String(), `"reauthenticatedUntil"`) {
		t.Fatalf("profile after confirm %d: %s", profile.Code, profile.Body.String())
	}

	// OTP 标记只授权设置密码；换邮箱的发码与变更都必须被拒绝。
	for _, tc := range []struct{ path, body string }{
		{"/api/auth/profile/email-code", `{"email":"new@example.com"}`},
		{"/api/auth/profile/email", `{"ticket":"fake.ticket","emailCode":"123456"}`},
	} {
		if got := profileRequest(router, "POST", tc.path, tc.body, cookie); got.Code != http.StatusUnauthorized {
			t.Fatalf("%s after OTP returned %d: %s", tc.path, got.Code, got.Body.String())
		}
	}
	var verifications int64
	if err := db.Model(&model.AuthVerification{}).Where("purpose = ?", "profile_email").Count(&verifications).Error; err != nil || verifications != 0 {
		t.Fatalf("email change challenge created after OTP: %d %v", verifications, err)
	}

	setPassword := profileRequest(router, "POST", "/api/auth/profile/password", `{"newPassword":"new-password"}`, cookie)
	if setPassword.Code != http.StatusOK {
		t.Fatalf("set password status %d: %s", setPassword.Code, setPassword.Body.String())
	}
	var user model.User
	if err := db.First(&user, "id = ?", "profile-user").Error; err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("new-password")) != nil {
		t.Fatal("password was not persisted after OTP reauthentication")
	}
	for name, text := range map[string]string{"start": start.Body.String(), "confirm": confirm.Body.String(), "profile": profile.Body.String(), "password": setPassword.Body.String(), "logs": logs.String()} {
		if strings.Contains(text, code) {
			t.Errorf("%s output contains the verification code", name)
		}
	}
}
