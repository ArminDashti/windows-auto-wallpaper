package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ArminDashti/auto-wallpaper-api/internal/config"
	appdb "github.com/ArminDashti/auto-wallpaper-api/internal/db"
	"github.com/ArminDashti/auto-wallpaper-api/internal/handlers"
	"github.com/ArminDashti/auto-wallpaper-api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadDotEnv(".env")

	cfg := config.Load()
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "wallpapers", "lock"), 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "wallpapers", "home"), 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}

	sqlDB, err := appdb.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer sqlDB.Close()

	if err := appdb.Migrate(sqlDB, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	h := handlers.New(sqlDB, cfg.DataDir)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.MaxMultipartMemory = 32 << 20

	r.GET("/health", h.Health)
	r.POST("/api/auth/login", h.Login)
	r.POST("/api/auth/logout", h.Logout)

	api := r.Group("/api")
	api.Use(middleware.RequireAuth(sqlDB))
	{
		api.GET("/me", h.Me)
		api.GET("/schedules/:screen", h.GetSchedule)
		api.PUT("/schedules/:screen", h.PutSchedule)
		// screen list/upload under /screens to avoid Gin :screen vs :id conflict
		api.GET("/screens/:screen/wallpapers", h.ListWallpapers)
		api.POST("/screens/:screen/wallpapers", h.UploadWallpaper)
		api.PATCH("/wallpapers/:id", h.PatchWallpaper)
		api.DELETE("/wallpapers/:id", h.DeleteWallpaper)
		api.GET("/wallpapers/:id/file", h.GetWallpaperFile)
		api.GET("/agent/state", h.AgentState)
	}

	log.Printf("auto-wallpaper-api listening on %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
