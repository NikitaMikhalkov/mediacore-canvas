package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"yingce/backend/internal/auth"
	"yingce/backend/internal/model"
)

func TestRegistrationHTTPStartsWithZeroCreditsAndKeepsPublicRegistrationClosed(t *testing.T) {
	t.Setenv("CANVAS_REGISTRATION_ENABLED", "false")
	router, db, _ := profileHandlerFixture(t)
	// This isolated fixture starts before the first administrator is created.
	for _, entity := range []any{&model.AuthSession{}, &model.User{}} {
		if err := db.Where("1 = 1").Delete(entity).Error; err != nil {
			t.Fatal(err)
		}
	}
	registration := profileRequest(router, "POST", "/api/auth/register", `{"username":"new-admin","password":"strong-password","acceptedTerms":true}`, "")
	if registration.Code != http.StatusOK {
		t.Fatalf("registration: %d %s", registration.Code, registration.Body.String())
	}
	var cookie string
	for _, value := range registration.Result().Cookies() {
		if value.Name == auth.SessionCookieName {
			cookie = value.Value
		}
	}
	if cookie == "" {
		t.Fatal("registration did not create a login session")
	}
	settings := profileRequest(router, "GET", "/api/auth/settings", "", "")
	var publicSettings struct {
		Data struct {
			FirstUser           bool `json:"firstUser"`
			RegistrationEnabled bool `json:"registrationEnabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(settings.Body.Bytes(), &publicSettings); err != nil {
		t.Fatal(err)
	}
	if settings.Code != http.StatusOK || publicSettings.Data.FirstUser || publicSettings.Data.RegistrationEnabled {
		t.Fatalf("new installation publicly opened ordinary registration: %d %s", settings.Code, settings.Body.String())
	}
	wallet := profileRequest(router, "GET", "/api/wallet", "", cookie)
	var summary struct {
		Data struct {
			Account struct {
				UserID    string `json:"userId"`
				Available int64 `json:"availableMicrocredits"`
				Reserved  int64 `json:"reservedMicrocredits"`
			} `json:"account"`
			Policy struct {
				SignupBonus int64 `json:"signupBonusMicrocredits"`
			} `json:"policy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(wallet.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if wallet.Code != http.StatusOK || summary.Data.Account.UserID == "" || summary.Data.Account.Available != 0 || summary.Data.Account.Reserved != 0 || summary.Data.Policy.SignupBonus != 0 {
		t.Fatalf("new account must start at zero credits: %d %s", wallet.Code, wallet.Body.String())
	}
	blocked := profileRequest(router, "POST", "/api/auth/register", `{"username":"next-user","password":"strong-password","acceptedTerms":true}`, "")
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("ordinary registration must stay closed: %d %s", blocked.Code, blocked.Body.String())
	}
}
