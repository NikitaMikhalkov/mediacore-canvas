package app

import "testing"

func TestPublicAppearanceDefaultsUseMediacoreIdentity(t *testing.T) {
	svc, _, _, _ := newAppearanceTestService(t)
	appearance, err := svc.Appearance()
	if err != nil {
		t.Fatal(err)
	}
	if appearance.BrandName != "Mediacore" || appearance.BrandSlug != "mediacore" || appearance.SEOTitle != "Mediacore" || appearance.Canvas.AgentName != "Mediacore" {
		t.Fatalf("public default identity = %#v", appearance)
	}
	if appearance.AuthVideoURL != "" || appearance.AuthVideoPosterURL != "/brand/mediacore-auth-poster.svg" {
		t.Fatalf("public default media should be bundled, got video %q and poster %q", appearance.AuthVideoURL, appearance.AuthVideoPosterURL)
	}
}

func TestAppearanceResetRestoresMediacorePublicIdentity(t *testing.T) {
	svc, _, _, admin := newAppearanceTestService(t)
	setting, err := svc.AdminAppearance(admin)
	if err != nil {
		t.Fatal(err)
	}
	custom := setting.AppearanceSetting
	custom.BrandName = "My Studio"
	custom.BrandSlug = "my-studio"
	custom.Canvas.AgentName = "Custom Assistant"
	if _, err := svc.UpdateAppearance(admin, custom); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResetAppearance(admin); err != nil {
		t.Fatal(err)
	}
	appearance, err := svc.Appearance()
	if err != nil {
		t.Fatal(err)
	}
	if appearance.Configured || appearance.BrandName != "Mediacore" || appearance.Canvas.AgentName != "Mediacore" || appearance.SEOTitle != "Mediacore" || appearance.Revision != "builtin" {
		t.Fatalf("public appearance after reset = %#v", appearance)
	}
}
