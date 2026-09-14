package handlers

import (
	"database/sql"
	"path/filepath"
)

// Handlers holds shared dependencies for HTTP handlers.
type Handlers struct {
	db      *sql.DB
	dataDir string
}

// New creates Handlers.
func New(db *sql.DB, dataDir string) *Handlers {
	return &Handlers{db: db, dataDir: dataDir}
}

func (h *Handlers) wallpaperDir(screen string) string {
	return filepath.Join(h.dataDir, "wallpapers", screen)
}

func (h *Handlers) wallpaperPath(screen, storedName string) string {
	return filepath.Join(h.wallpaperDir(screen), storedName)
}
