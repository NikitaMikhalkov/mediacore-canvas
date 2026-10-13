package handler

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"yingce/backend/internal/auth"
	"yingce/backend/internal/database"
	"yingce/backend/internal/model"
	"yingce/backend/internal/repository"
	"yingce/backend/internal/service"
)

func TestPluginDownloadUsesNeutralZIPNameAndPreservesPackageBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "fixture.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateSchema(db); err != nil {
		t.Fatal(err)
	}
	svc := service.New(repository.New(db), dir)
	t.Cleanup(func() { svc.Close(); sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	admin := model.User{ID: "download-admin", Username: "download-admin", Role: model.UserRoleAdmin, Status: model.UserStatusActive}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.AuthSession{ID: "download-session", UserID: admin.ID, TokenHash: auth.HashToken("download-token"), ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entry, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(`{"apiVersion":"yingce.plugin/v1","id":"brand-download","name":"Brand Download","version":"1.0.0","contributes":{"commands":[{"id":"brand-download/open","label":"Open"}]}}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	packageBytes := buffer.Bytes()
	if _, err := svc.InstallPluginForAdmin(&admin, packageBytes, "brand-download.yingce-plugin"); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	RegisterPluginRoutes(router.Group("/api"), svc)
	req := httptest.NewRequest(http.MethodGet, "/api/plugins/brand-download/package", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "download-session.download-token"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Disposition"), `filename="brand-download.zip"`) {
		t.Fatalf("download = HTTP %d, disposition %q", rec.Code, rec.Header().Get("Content-Disposition"))
	}
	if rec.Header().Get("Content-Type") != "application/zip" || !bytes.Equal(rec.Body.Bytes(), packageBytes) {
		t.Fatal("ZIP download changed the plugin format or payload")
	}
	if _, err := svc.InstallPluginForAdmin(&admin, rec.Body.Bytes(), "brand-download.zip"); err != nil {
		t.Fatalf("downloaded ZIP cannot be reinstalled: %v", err)
	}
	reinstalled := httptest.NewRecorder()
	router.ServeHTTP(reinstalled, req)
	if reinstalled.Code != http.StatusOK || !bytes.Equal(reinstalled.Body.Bytes(), packageBytes) {
		t.Fatal("reinstalled ZIP changed the compatible package")
	}
}
