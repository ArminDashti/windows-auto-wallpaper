package config_test

import (
	"os"
	"testing"

	"github.com/ArminDashti/auto-wallpaper/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	_ = os.Unsetenv("AUTOWALL_API_URL")
	_ = os.Unsetenv("API_URL")
	cfg := config.Load()
	if cfg.APIURL != "http://127.0.0.1:8101" {
		t.Fatalf("APIURL=%s", cfg.APIURL)
	}
	if cfg.ServiceName != "AutoWallpaper" {
		t.Fatalf("ServiceName=%s", cfg.ServiceName)
	}
}
