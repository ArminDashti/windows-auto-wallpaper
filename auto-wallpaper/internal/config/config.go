package config

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config for autowall agent.
type Config struct {
	APIURL       string
	Username     string
	Password     string
	Token        string
	PollSeconds  int
	DataDir      string
	CacheDir     string
	InstallDir   string
	ServiceName  string
	DisplayName  string
}

func defaultDataDir() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("ProgramData")
		if base == "" {
			base = `C:\ProgramData`
		}
		return filepath.Join(base, "AutoWallpaper")
	}
	return filepath.Join(os.TempDir(), "AutoWallpaper")
}

// LoadDotEnv loads KEY=VALUE from path when key not already set.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

// Load reads config from env and optional ProgramData config.env.
func Load() Config {
	data := envOr("AUTOWALL_DATA_DIR", defaultDataDir())
	LoadDotEnv(filepath.Join(data, "config.env"))
	LoadDotEnv(".env")

	poll := 60
	if v := os.Getenv("AUTOWALL_POLL_SECONDS"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			poll = n
		}
	}

	return Config{
		APIURL:      strings.TrimRight(envOr("AUTOWALL_API_URL", envOr("API_URL", "http://127.0.0.1:8101")), "/"),
		Username:    envOr("AUTOWALL_USERNAME", "armin"),
		Password:    envOr("AUTOWALL_PASSWORD", "dopadopa123"),
		Token:       os.Getenv("AUTOWALL_TOKEN"),
		PollSeconds: poll,
		DataDir:     data,
		CacheDir:    filepath.Join(data, "cache"),
		InstallDir:  filepath.Join(data, "bin"),
		ServiceName: envOr("AUTOWALL_SERVICE_NAME", "AutoWallpaper"),
		DisplayName: "Auto Wallpaper",
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
