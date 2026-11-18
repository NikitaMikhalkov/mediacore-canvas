package auth

import (
	"testing"

	"yingce/backend/internal/model"
	"yingce/backend/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPublicAuthUserKeepsLocalUserWithoutIdentity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.UserIdentity{}); err != nil {
		t.Fatal(err)
	}
	user := model.User{ID: "user-1", Username: "local-user", DisplayName: "Local User"}

	result, err := (New(repository.New(db), nil, nil)).PublicAuthUser(&user)
	if err != nil {
		t.Fatal(err)
	}
	if result.Username != user.Username || result.AvatarURL != "" {
		t.Fatalf("PublicAuthUser() = %#v", result)
	}
}
