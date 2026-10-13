package handler

import (
	"net/http"
	"testing"
	"time"

	"yingce/backend/internal/auth"
	"yingce/backend/internal/model"
)

func TestLoginHTTPKeepsSessionForSevenDaysWithoutActivityRenewal(t *testing.T) {
	router, _, _ := profileHandlerFixture(t)
	login := profileRequest(router, "POST", "/api/auth/login", `{"username":"profile-user","password":"old-password"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status %d: %s", login.Code, login.Body.String())
	}
	var session *http.Cookie
	for _, cookie := range login.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName {
			session = cookie
		}
	}
	if session == nil || session.MaxAge != 604800 || !session.HttpOnly || session.Path != "/" || session.SameSite != http.SameSiteLaxMode {
		t.Fatalf("login must issue a protected seven-day cookie: %v", session)
	}
	for _, path := range []string{"/api/auth/session", "/api/auth/profile"} {
		response := profileRequest(router, "GET", path, "", session.Value)
		if response.Code != http.StatusOK {
			t.Fatalf("session read %s: %d %s", path, response.Code, response.Body.String())
		}
		if response.Header().Get("Set-Cookie") != "" {
			t.Fatalf("activity at %s renewed the login cookie", path)
		}
	}
}

func TestSessionHTTPRejectsExpiredRevokedAndDisabledAccounts(t *testing.T) {
	for _, state := range []string{"expired", "logout", "disabled"} {
		t.Run(state, func(t *testing.T) {
			router, db, cookie := profileHandlerFixture(t)
			switch state {
			case "expired":
				if err := db.Model(&model.AuthSession{}).Where("id = ?", "profile-session").Update("expires_at", time.Now().Add(-time.Second)).Error; err != nil {
					t.Fatal(err)
				}
			case "logout":
				response := profileRequest(router, "POST", "/api/auth/logout", "", cookie)
				if response.Code != http.StatusOK {
					t.Fatalf("logout: %d %s", response.Code, response.Body.String())
				}
				cleared := false
				for _, value := range response.Result().Cookies() {
					cleared = cleared || (value.Name == auth.SessionCookieName && value.MaxAge < 0)
				}
				if !cleared {
					t.Fatal("logout did not clear login cookie")
				}
			case "disabled":
				if err := db.Model(&model.User{}).Where("id = ?", "profile-user").Update("status", model.UserStatusDisabled).Error; err != nil {
					t.Fatal(err)
				}
			}
			response := profileRequest(router, "GET", "/api/auth/profile", "", cookie)
			if response.Code != http.StatusUnauthorized && response.Code != http.StatusForbidden {
				t.Fatalf("%s session still grants account access: %d %s", state, response.Code, response.Body.String())
			}
		})
	}
}
