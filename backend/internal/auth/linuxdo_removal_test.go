package auth

import (
	"encoding/json"
	"strings"
	"testing"

	"yingce/backend/internal/model"
)

// Linux.do login was removed: auth payloads carry no Linux.do state even when
// a legacy identity row still exists.
func TestAuthPayloadsCarryNoLinuxDOState(t *testing.T) {
	s, db, u, cookie := newProfileTestService(t)
	if err := db.Create(&model.UserIdentity{ID: "legacy", UserID: u.ID, Provider: "linuxdo", Subject: "42", ProviderUsername: "legacy-name", AvatarURL: "https://example.com/a.png"}).Error; err != nil {
		t.Fatal(err)
	}
	profile, err := s.UserProfile(u, cookie)
	if err != nil {
		t.Fatal(err)
	}
	settings, err := s.PublicAuthSettings()
	if err != nil {
		t.Fatal(err)
	}
	for name, payload := range map[string]any{"profile": profile, "settings": settings} {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(raw))
		for _, forbidden := range []string{"linuxdo", "identityprovider", "identityid", "identityusername", "legacy-name", "example.com/a.png"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%s payload contains %q: %s", name, forbidden, raw)
			}
		}
	}
}
