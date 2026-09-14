package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ArminDashti/auto-wallpaper/internal/client"
	"github.com/ArminDashti/auto-wallpaper/internal/config"
	"github.com/ArminDashti/auto-wallpaper/internal/wallpaper"
)

func runWorker(cfg config.Config) error {
	if err := os.MkdirAll(cfg.CacheDir, 0o755); err != nil {
		return err
	}
	cli := client.New(cfg.APIURL, cfg.Token)
	if err := cli.EnsureToken(cfg.Username, cfg.Password); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	ticker := time.NewTicker(time.Duration(cfg.PollSeconds) * time.Second)
	defer ticker.Stop()

	applyOnce := func() {
		if err := cli.EnsureToken(cfg.Username, cfg.Password); err != nil {
			log.Printf("auth: %v", err)
			return
		}
		st, err := cli.AgentState()
		if err != nil {
			log.Printf("agent state: %v", err)
			return
		}
		for _, screen := range st.Screens {
			if !screen.Schedule.Enabled || screen.CurrentWallpaper == nil {
				continue
			}
			w := screen.CurrentWallpaper
			data, err := cli.DownloadFile(w.ID)
			if err != nil {
				log.Printf("download %s #%d: %v", screen.Screen, w.ID, err)
				continue
			}
			ext := filepath.Ext(w.Filename)
			if ext == "" {
				ext = ".jpg"
			}
			path := filepath.Join(cfg.CacheDir, screen.Screen+ext)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				log.Printf("write cache: %v", err)
				continue
			}
			var applyErr error
			switch screen.Screen {
			case "home":
				applyErr = wallpaper.ApplyHome(path)
			case "lock":
				applyErr = wallpaper.ApplyLock(path)
			}
			if applyErr != nil {
				log.Printf("apply %s: %v", screen.Screen, applyErr)
			} else {
				log.Printf("applied %s wallpaper id=%d", screen.Screen, w.ID)
			}
		}
	}

	applyOnce()
	for range ticker.C {
		applyOnce()
	}
	return nil
}
