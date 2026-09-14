package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ArminDashti/auto-wallpaper/internal/config"
	"github.com/ArminDashti/auto-wallpaper/internal/doctor"
	"github.com/ArminDashti/auto-wallpaper/internal/service"
)

func main() {
	cfg := config.Load()

	if len(os.Args) == 1 {
		if isWindowsService() {
			if err := service.RunService(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "service: %v\n", err)
				os.Exit(1)
			}
			return
		}
		printUsage()
		os.Exit(2)
	}

	if os.Args[1] == "service" && len(os.Args) == 2 {
		if err := service.RunService(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "service: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "doctor":
		if !doctor.Run(cfg) {
			os.Exit(1)
		}
	case "service":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: autowall service start|stop|restart")
			os.Exit(2)
		}
		mgr := service.New(cfg)
		var err error
		switch os.Args[2] {
		case "start":
			err = mgr.Start()
		case "stop":
			err = mgr.Stop()
		case "restart":
			err = mgr.Restart()
		default:
			fmt.Fprintln(os.Stderr, "usage: autowall service start|stop|restart")
			os.Exit(2)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("ok")
	case "uninstall":
		if err := service.New(cfg).Uninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("uninstalled")
	case "run":
		if err := service.RunWorker(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println(strings.TrimSpace(`
autowall — Windows auto wallpaper agent (CLI, no GUI)

Commands:
  autowall doctor
  autowall service start
  autowall service stop
  autowall service restart
  autowall uninstall
  autowall run          # foreground poll loop (debug)

Env / %ProgramData%\AutoWallpaper\config.env:
  AUTOWALL_API_URL   (default http://127.0.0.1:8101)
  AUTOWALL_USERNAME  (default armin)
  AUTOWALL_PASSWORD  (default dopadopa123)
  AUTOWALL_TOKEN     (optional; skips login)
  AUTOWALL_POLL_SECONDS (default 60)
`))
}
