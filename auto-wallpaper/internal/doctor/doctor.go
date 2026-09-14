package doctor

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ArminDashti/auto-wallpaper/internal/client"
	"github.com/ArminDashti/auto-wallpaper/internal/config"
	"github.com/ArminDashti/auto-wallpaper/internal/service"
	"github.com/ArminDashti/auto-wallpaper/internal/wallpaper"
)

type Check struct {
	Name   string
	OK     bool
	Detail string
}

// Run executes diagnostic checks and prints results. Returns false if any failed.
func Run(cfg config.Config) bool {
	checks := []Check{}

	checks = append(checks, Check{
		Name:   "os",
		OK:     runtime.GOOS == "windows",
		Detail: fmt.Sprintf("GOOS=%s (Windows required for apply/service)", runtime.GOOS),
	})

	checks = append(checks, Check{
		Name:   "wallpaper_api",
		OK:     wallpaper.Supported(),
		Detail: map[bool]string{true: "SystemParametersInfo + PersonalizationCSP available", false: "stubs only on this OS"}[wallpaper.Supported()],
	})

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		checks = append(checks, Check{Name: "data_dir", OK: false, Detail: err.Error()})
	} else {
		checks = append(checks, Check{Name: "data_dir", OK: true, Detail: cfg.DataDir})
	}

	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		checks = append(checks, Check{Name: "cache_dir", OK: false, Detail: err.Error()})
	} else {
		test := cfg.CacheDir + string(os.PathSeparator) + ".writetest"
		if err := os.WriteFile(test, []byte("ok"), 0o644); err != nil {
			checks = append(checks, Check{Name: "cache_writable", OK: false, Detail: err.Error()})
		} else {
			_ = os.Remove(test)
			checks = append(checks, Check{Name: "cache_writable", OK: true, Detail: cfg.CacheDir})
		}
	}

	cli := client.New(cfg.APIURL, cfg.Token)
	if err := cli.Health(); err != nil {
		checks = append(checks, Check{Name: "api_health", OK: false, Detail: fmt.Sprintf("%s: %v", cfg.APIURL, err)})
	} else {
		checks = append(checks, Check{Name: "api_health", OK: true, Detail: cfg.APIURL})
	}

	if err := cli.EnsureToken(cfg.Username, cfg.Password); err != nil {
		checks = append(checks, Check{Name: "api_auth", OK: false, Detail: err.Error()})
	} else {
		checks = append(checks, Check{Name: "api_auth", OK: true, Detail: "login ok"})
		if _, err := cli.AgentState(); err != nil {
			checks = append(checks, Check{Name: "agent_state", OK: false, Detail: err.Error()})
		} else {
			checks = append(checks, Check{Name: "agent_state", OK: true, Detail: "ok"})
		}
	}

	mgr := service.New(cfg)
	st := mgr.Status()
	okSvc := st == "running" || st == "stopped" || st == "not installed" || strings.Contains(st, "not windows")
	checks = append(checks, Check{Name: "service", OK: okSvc, Detail: st})

	allOK := true
	for _, c := range checks {
		mark := "PASS"
		if !c.OK {
			mark = "FAIL"
			allOK = false
		}
		fmt.Printf("[%s] %s — %s\n", mark, c.Name, c.Detail)
	}
	return allOK
}
