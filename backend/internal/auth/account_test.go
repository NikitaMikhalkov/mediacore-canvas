package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
	"yingce/backend/internal/model"
)

type profileTestHost struct{ nopHost }

func (profileTestHost) SettingsEncryptionKey() ([]byte, error) {
	return []byte("profile-test-key-32-bytes-long!!!"), nil
}

func newProfileTestService(t *testing.T) (*Service, *gorm.DB, *model.User, string) {
	t.Helper()
	s, db := newPasswordResetTestService(t)
	s.host = profileTestHost{}
	if err := db.AutoMigrate(&model.AuthVerification{}, &model.NotificationQuota{}, &model.UserIdentity{}, &model.OAuthState{}); err != nil {
		t.Fatal(err)
	}
	hash, err := HashPassword("old-password")
	if err != nil {
		t.Fatal(err)
	}
	u := &model.User{ID: "profile-user", Username: "profile-user", Email: "old@example.com", Status: model.UserStatusActive, PasswordHash: hash}
	if err := db.Create(u).Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"session-one", "session-two"} {
		if err := db.Create(&model.AuthSession{ID: id, UserID: u.ID, TokenHash: HashToken("token"), ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return s, db, u, "session-one.token"
}

func TestProfilePasswordRequiresCurrentCredentialsAndRevokesAllSessions(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{CurrentPassword: "wrong", NewPassword: "new-password"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{CurrentPassword: "old-password", NewPassword: "new-password"}); err != nil {
		t.Fatal(err)
	}
	var updated model.User
	db.First(&updated, "id = ?", u.ID)
	if !verifyPassword("new-password", updated.PasswordHash) {
		t.Fatal("password not changed")
	}
	var count int64
	db.Model(&model.AuthSession{}).Where("user_id = ?", u.ID).Count(&count)
	if count != 0 {
		t.Fatal("sessions not revoked")
	}
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{CurrentPassword: "old-password", NewPassword: "third-password"}); err == nil {
		t.Fatal("stale actor/session accepted")
	}
}

func TestProfileOAuthAccountNeedsRecentSameSessionReauthentication(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	db.Model(u).Update("password_hash", "")
	u.PasswordHash = ""
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("OAuth account accepted without reauthentication")
	}
	now := time.Now()
	db.Model(&model.AuthSession{}).Where("id = ?", "session-two").Update("reauthenticated_at", now)
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("another session reauthentication accepted")
	}
	expired := now.Add(-6 * time.Minute)
	db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("reauthenticated_at", expired)
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err == nil {
		t.Fatal("expired reauthentication accepted")
	}
	db.Model(&model.AuthSession{}).Where("id = ?", "session-one").Update("reauthenticated_at", now)
	if err := s.ChangeProfilePassword(u, cookie, ProfilePasswordRequest{NewPassword: "new-password"}); err != nil {
		t.Fatal(err)
	}
}

func TestProfileEmailTicketBoundToSessionAndReportsOldMailFailure(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	var code string
	s.SetMailSender(func(_ EmailSettingValue, to, subject, body string) error {
		if to == "new@external.example" {
			code = codeFromEmailBody(body)
			return nil
		}
		return errors.New("mail unavailable")
	})
	ticket, err := s.SendProfileEmailCode(context.Background(), u, cookie, ProfileEmailCodeRequest{Email: "new@external.example", CurrentPassword: "old-password"})
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatal("missing code")
	}
	req := ProfileEmailChangeRequest{Ticket: ticket.Ticket, EmailCode: code, CurrentPassword: "old-password"}
	if _, err := s.ChangeProfileEmail(u, "session-two.token", req); err == nil {
		t.Fatal("another session used ticket")
	}
	result, err := s.ChangeProfileEmail(u, cookie, req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.NotificationStatus != "failed" {
		t.Fatalf("result %#v", result)
	}
	var updated model.User
	db.First(&updated, "id = ?", u.ID)
	if updated.Email != "new@external.example" || updated.EmailVerifiedAt == nil {
		t.Fatalf("updated user %#v", updated)
	}
	var count int64
	db.Model(&model.AuthSession{}).Where("user_id = ?", u.ID).Count(&count)
	if count != 0 {
		t.Fatal("sessions remain")
	}
	if _, err := s.ChangeProfileEmail(u, cookie, req); err == nil {
		t.Fatal("reused ticket accepted")
	}
}

func TestProfileEmailDisabledAndDuplicateAreRejected(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	db.Create(&model.User{ID: "other", Username: "other", Email: "used@example.com", Status: model.UserStatusActive})
	if _, err := s.SendProfileEmailCode(context.Background(), u, cookie, ProfileEmailCodeRequest{Email: "used@example.com", CurrentPassword: "old-password"}); err == nil {
		t.Fatal("duplicate email accepted")
	}
	db.Model(&model.SystemSetting{}).Where("key = ?", emailSettingKey).Update("value_json", `{"enabled":false}`)
	if _, err := s.SendProfileEmailCode(context.Background(), u, cookie, ProfileEmailCodeRequest{Email: "new@example.com", CurrentPassword: "old-password"}); err == nil {
		t.Fatal("disabled SMTP accepted")
	}
}
