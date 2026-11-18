package auth

import (
	"testing"

	"yingce/backend/internal/model"

	"gorm.io/gorm"
)

// Commit the administrator's change after the login read, before its user update.
func disableBeforeLoginWrite(t *testing.T, db *gorm.DB, userID string) *bool {
	t.Helper()
	fired := false
	if err := db.Callback().Update().Before("gorm:begin_transaction").Register("test:disable_before_login_write", func(tx *gorm.DB) {
		if fired || tx.Statement.Schema == nil || tx.Statement.Schema.Table != "users" {
			return
		}
		fired = true
		if err := db.Exec("UPDATE users SET status = ?, display_name = ? WHERE id = ?", model.UserStatusDisabled, "Admin Edit", userID).Error; err != nil {
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	return &fired
}

func assertLoginPreservesDisable(t *testing.T, db *gorm.DB, fired *bool, loginErr error) {
	t.Helper()
	if !*fired {
		t.Fatal("concurrent disable was not injected")
	}
	var user model.User
	if err := db.First(&user, "id = ?", "member").Error; err != nil {
		t.Fatal(err)
	}
	if user.Status != model.UserStatusDisabled {
		t.Errorf("status after login = %q, want disabled", user.Status)
	}
	if user.DisplayName != "Admin Edit" {
		t.Errorf("display name after login = %q, want Admin Edit", user.DisplayName)
	}
	if user.LastLoginAt == nil || user.UpdatedAt.Before(*user.LastLoginAt) {
		t.Error("login timestamps were not updated")
	}
	if loginErr == nil {
		t.Error("login issued a session after concurrent disable")
	}
	var sessions int64
	if err := db.Model(&model.AuthSession{}).Where("user_id = ?", user.ID).Count(&sessions).Error; err != nil {
		t.Fatal(err)
	}
	if sessions != 0 {
		t.Error("disabled user received a login session")
	}
}

func TestPasswordLoginPreservesConcurrentDisable(t *testing.T) {
	svc, db := newRegistrationTestService(t)
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.User{ID: "member", Username: "member", DisplayName: "Before", PasswordHash: hash, Role: model.UserRoleUser, Status: model.UserStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	fired := disableBeforeLoginWrite(t, db, "member")
	_, err = svc.Login(LoginRequest{Username: "member", Password: "password"})
	assertLoginPreservesDisable(t, db, fired, err)
}
