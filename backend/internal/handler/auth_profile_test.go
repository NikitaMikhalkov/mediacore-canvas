package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"yingce/backend/internal/auth"
	"yingce/backend/internal/database"
	"yingce/backend/internal/model"
	"yingce/backend/internal/platform"
	"yingce/backend/internal/repository"
	"yingce/backend/internal/service"
)

func profileHandlerFixture(t *testing.T) (*gin.Engine, *gorm.DB, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "fixture.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateSchema(db); err != nil {
		t.Fatal(err)
	}
	hash, err := auth.HashPassword("old-password")
	if err != nil {
		t.Fatal(err)
	}
	user := model.User{ID: "profile-user", Username: "profile-user", DisplayName: "Creator", Role: model.UserRoleUser, Status: model.UserStatusActive, PasswordHash: hash}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]model.AuthSession{
		{ID: "profile-session", UserID: user.ID, TokenHash: auth.HashToken("fixture-token"), ExpiresAt: time.Now().Add(time.Hour)},
		{ID: "profile-other-session", UserID: user.ID, TokenHash: auth.HashToken("other-token"), ExpiresAt: time.Now().Add(time.Hour)},
	}).Error; err != nil {
		t.Fatal(err)
	}
	svc := service.New(repository.New(db), dir)
	ConfigureRuntime(svc)
	t.Cleanup(func() { ConfigureRuntime(nil); svc.Close(); sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	router := gin.New()
	RegisterAuthRoutes(router.Group("/api"), svc)
	RegisterFinanceRoutes(router.Group("/api"), svc)
	return router, db, "profile-session.fixture-token"
}

func profileRequest(router http.Handler, method, path, body, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestProfileRoutesRejectAnonymousBeforeMutation(t *testing.T) {
	router, _, _ := profileHandlerFixture(t)
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/auth/profile"}, {"PUT", "/api/auth/profile/avatar"},
		{"DELETE", "/api/auth/profile/avatar"}, {"GET", "/api/auth/profile/avatar/profile-user"},
		{"POST", "/api/auth/profile/avatar/default"},
		{"POST", "/api/auth/profile/password"}, {"POST", "/api/auth/profile/email-code"},
		{"POST", "/api/auth/profile/email"},
	} {
		rec := profileRequest(router, route.method, route.path, "{}", "")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s returned %d", route.method, route.path, rec.Code)
		}
	}
}

func TestProfilePasswordHTTPRevokesSessionsAndClearsCookie(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	wrong := profileRequest(router, "POST", "/api/auth/profile/password", `{"currentPassword":"wrong-password","newPassword":"new-password"}`, cookie)
	if wrong.Code != 401 {
		t.Fatalf("wrong password status %d: %s", wrong.Code, wrong.Body.String())
	}
	var count int64
	db.Model(&model.AuthSession{}).Count(&count)
	if count != 2 {
		t.Fatalf("failed change revoked sessions: %d", count)
	}
	rec := profileRequest(router, "POST", "/api/auth/profile/password", `{"currentPassword":"old-password","newPassword":"new-password"}`, cookie)
	if rec.Code != 200 {
		t.Fatalf("change status %d: %s", rec.Code, rec.Body.String())
	}
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookieName && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("successful change did not clear login cookie")
	}
	var user model.User
	db.First(&user, "id = ?", "profile-user")
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("new-password")) != nil {
		t.Fatal("new password not persisted")
	}
	db.Model(&model.AuthSession{}).Count(&count)
	if count != 0 {
		t.Fatalf("sessions remain after change: %d", count)
	}
	if status := profileRequest(router, "GET", "/api/auth/profile", "", cookie).Code; status != 401 {
		t.Fatalf("old cookie still accepted: %d", status)
	}
}

func TestProfilePasswordRejectsOversizedUTF8WithoutRevokingSessions(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	body, err := json.Marshal(map[string]string{"currentPassword": "old-password", "newPassword": strings.Repeat("密", 25)})
	if err != nil {
		t.Fatal(err)
	}
	rec := profileRequest(router, "POST", "/api/auth/profile/password", string(body), cookie)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized UTF-8 password status %d, want 400", rec.Code)
	}
	var count int64
	if err := db.Model(&model.AuthSession{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("invalid password revoked sessions: %d", count)
	}
}

func TestProfileDefaultAvatarSelectionPersistsAndRejectsUnknownPreset(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	invalid := profileRequest(router, "POST", "/api/auth/profile/avatar/default", `{"preset":"unsupported"}`, cookie)
	if invalid.Code != 400 {
		t.Fatalf("unknown preset status %d", invalid.Code)
	}
	valid := profileRequest(router, "POST", "/api/auth/profile/avatar/default", `{"preset":"cat"}`, cookie)
	if valid.Code != 200 {
		t.Fatalf("selection status %d: %s", valid.Code, valid.Body.String())
	}
	var user model.User
	if err := db.First(&user, "id = ?", "profile-user").Error; err != nil {
		t.Fatal(err)
	}
	if user.AvatarMode != "default" || user.AvatarPreset != "cat" || user.AvatarVersion != "" {
		t.Fatal("selected default was not persisted")
	}
	reloaded := profileRequest(router, "GET", "/api/auth/profile", "", cookie)
	if reloaded.Code != 200 || !strings.Contains(reloaded.Body.String(), `"avatarPreset":"cat"`) {
		t.Fatalf("reloaded profile lost selection: %d", reloaded.Code)
	}
}

func TestProfileAvatarHTTPPersistsServesAndRestoresDefault(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(part, image.NewNRGBA(image.Rect(0, 0, 40, 20))); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("PUT", "/api/auth/profile/avatar", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: cookie})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("upload status %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data struct {
			User struct {
				AvatarURL string `json:"avatarUrl"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(envelope.Data.User.AvatarURL, "/api/auth/profile/avatar/profile-user?v=") {
		t.Fatalf("missing versioned avatar URL: %s", envelope.Data.User.AvatarURL)
	}
	content := profileRequest(router, "GET", envelope.Data.User.AvatarURL, "", cookie)
	if content.Code != 200 || content.Header().Get("Content-Type") != "image/png" || content.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("avatar response %d %v", content.Code, content.Header())
	}
	if foreign := profileRequest(router, "GET", "/api/auth/profile/avatar/other-user", "", cookie); foreign.Code != 403 {
		t.Fatalf("foreign avatar status %d", foreign.Code)
	}
	reset := profileRequest(router, "DELETE", "/api/auth/profile/avatar", "", cookie)
	if reset.Code != 200 {
		t.Fatalf("reset avatar status %d: %s", reset.Code, reset.Body.String())
	}
	var user model.User
	db.First(&user, "id = ?", "profile-user")
	if user.AvatarMode != "default" || user.AvatarVersion != "" {
		t.Fatalf("default not persisted: %q", user.AvatarMode)
	}
}

func TestProfileOTPReauthenticationEndpointsRequireSession(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	for _, path := range []string{"/api/auth/profile/reauth-code", "/api/auth/profile/reauth"} {
		if got := profileRequest(router, "POST", path, `{}`, ""); got.Code != http.StatusUnauthorized {
			t.Fatalf("%s anonymous: %d", path, got.Code)
		}
		if got := profileRequest(router, "POST", path, `{}`, "missing-session.invalid"); got.Code != http.StatusUnauthorized {
			t.Fatalf("%s bad cookie: %d", path, got.Code)
		}
	}
	got := profileRequest(router, "POST", "/api/auth/profile/reauth-code", `{"method":"sms","phone":"13800138000"}`, cookie)
	if got.Code == http.StatusOK {
		t.Fatal("arbitrary client contact accepted")
	}
	var count int64
	if err := db.Model(&model.AuthVerification{}).Where("purpose = ?", "profile_reauth").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("unexpected challenge count %d: %v", count, err)
	}
	if err := db.Model(&model.User{}).Where("id = ?", "profile-user").Updates(map[string]any{"phone": "+8613900139000", "phone_verified_at": time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	got = profileRequest(router, "POST", "/api/auth/profile/reauth-code", `{"method":"sms","phone":"13800138000"}`, cookie)
	if got.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured login SMS route: %d %s", got.Code, got.Body.String())
	}
	got = profileRequest(router, "POST", "/api/auth/profile/reauth", `{"ticket":"invalid","smsCode":"654321"}`, cookie)
	if got.Code == http.StatusOK || strings.Contains(got.Body.String(), "654321") {
		t.Fatal("invalid confirmation accepted or code reflected")
	}
	got = profileRequest(router, "POST", "/api/auth/verification", `{"purpose":"profile_reauth","method":"sms","phone":"13800138000"}`, cookie)
	if got.Code == http.StatusOK {
		t.Fatal("public profile reauthentication accepted")
	}
}

func TestProfileOTPReauthenticationRespectsRateLimit(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	policy := platform.DefaultRuntimePolicy()
	policy.Request.LoginAccountPerTenMinutes = 2
	raw, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.SystemSetting{Key: "runtime_policy", ValueJSON: string(raw)}).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		rec := profileRequest(router, "POST", "/api/auth/profile/reauth", `{"ticket":"invalid"}`, cookie)
		if i < 2 && rec.Code == http.StatusTooManyRequests {
			t.Fatalf("limited too early: %d", i)
		}
		if i == 2 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("limit missing: %d", rec.Code)
		}
	}
}

func TestLinuxDORoutesAndFieldsAreRemoved(t *testing.T) {
	router, db, cookie := profileHandlerFixture(t)
	if err := db.Model(&model.User{}).Where("id = ?", "profile-user").Update("role", model.UserRoleAdmin).Error; err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/auth/linuxdo/start"}, {"GET", "/api/auth/linuxdo/callback"},
		{"GET", "/api/auth/profile/linuxdo/start"},
		{"GET", "/api/admin/settings/linuxdo"}, {"PATCH", "/api/admin/settings/linuxdo"},
	} {
		if rec := profileRequest(router, route.method, route.path, "{}", cookie); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s returned %d, want 404", route.method, route.path, rec.Code)
		}
	}
	for _, path := range []string{"/api/auth/settings", "/api/auth/profile"} {
		rec := profileRequest(router, "GET", path, "", cookie)
		if rec.Code != http.StatusOK || strings.Contains(strings.ToLower(rec.Body.String()), "linuxdo") {
			t.Errorf("GET %s: %d %s", path, rec.Code, rec.Body.String())
		}
	}
}
