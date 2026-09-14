package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type wallpaperRow struct {
	ID         int64  `json:"id"`
	Screen     string `json:"screen"`
	Filename   string `json:"filename"`
	Enabled    bool   `json:"enabled"`
	SortOrder  int    `json:"sortOrder"`
	CreatedAt  string `json:"createdAt"`
	FileURL    string `json:"fileUrl"`
}

type wallpaperPatch struct {
	Enabled   *bool `json:"enabled"`
	SortOrder *int  `json:"sortOrder"`
}

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".bmp": true,
}

// ListWallpapers returns wallpapers for a screen.
func (h *Handlers) ListWallpapers(c *gin.Context) {
	screen := c.Param("screen")
	if !validScreen(screen) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screen must be lock or home"})
		return
	}
	rows, err := h.listWallpapers(c.Request.Context(), screen, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallpapers": rows})
}

// UploadWallpaper stores a multipart image for a screen.
func (h *Handlers) UploadWallpaper(c *gin.Context) {
	screen := c.Param("screen")
	if !validScreen(screen) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screen must be lock or home"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported image type"})
		return
	}
	if err := os.MkdirAll(h.wallpaperDir(screen), 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
		return
	}
	stored := uuid.NewString() + ext
	dst := h.wallpaperPath(screen, stored)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var maxOrder sql.NullInt64
	_ = h.db.QueryRowContext(ctx, `
		SELECT MAX(sort_order) FROM wallpapers WHERE screen = ?
	`, screen).Scan(&maxOrder)
	next := 0
	if maxOrder.Valid {
		next = int(maxOrder.Int64) + 1
	}

	res, err := h.db.ExecContext(ctx, `
		INSERT INTO wallpapers (screen, filename, stored_name, enabled, sort_order)
		VALUES (?, ?, ?, 1, ?)
	`, screen, filepath.Base(file.Filename), stored, next)
	if err != nil {
		_ = os.Remove(dst)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, wallpaperRow{
		ID: id, Screen: screen, Filename: filepath.Base(file.Filename),
		Enabled: true, SortOrder: next, CreatedAt: time.Now().UTC().Format(time.RFC3339),
		FileURL: fmt.Sprintf("/api/wallpapers/%d/file", id),
	})
}

// PatchWallpaper updates enabled or sort order.
func (h *Handlers) PatchWallpaper(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req wallpaperPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if req.Enabled == nil && req.SortOrder == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var screen, filename, stored, created string
	var enabled, sortOrder int
	err = h.db.QueryRowContext(ctx, `
		SELECT screen, filename, stored_name, enabled, sort_order, created_at
		FROM wallpapers WHERE id = ?
	`, id).Scan(&screen, &filename, &stored, &enabled, &sortOrder, &created)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	if req.Enabled != nil {
		if *req.Enabled {
			enabled = 1
		} else {
			enabled = 0
		}
	}
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	_, err = h.db.ExecContext(ctx, `
		UPDATE wallpapers SET enabled = ?, sort_order = ? WHERE id = ?
	`, enabled, sortOrder, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, wallpaperRow{
		ID: id, Screen: screen, Filename: filename, Enabled: enabled == 1,
		SortOrder: sortOrder, CreatedAt: created,
		FileURL: fmt.Sprintf("/api/wallpapers/%d/file", id),
	})
}

// DeleteWallpaper removes a wallpaper and its file.
func (h *Handlers) DeleteWallpaper(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var screen, stored string
	err = h.db.QueryRowContext(ctx, `
		SELECT screen, stored_name FROM wallpapers WHERE id = ?
	`, id).Scan(&screen, &stored)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	if _, err := h.db.ExecContext(ctx, `DELETE FROM wallpapers WHERE id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	_ = os.Remove(h.wallpaperPath(screen, stored))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetWallpaperFile streams the image file.
func (h *Handlers) GetWallpaperFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var screen, stored, filename string
	err = h.db.QueryRowContext(ctx, `
		SELECT screen, stored_name, filename FROM wallpapers WHERE id = ?
	`, id).Scan(&screen, &stored, &filename)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	path := h.wallpaperPath(screen, stored)
	f, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file missing"})
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	c.Header("Cache-Control", "private, max-age=60")
	http.ServeContent(c.Writer, c.Request, filename, stat.ModTime(), f)
}

func (h *Handlers) listWallpapers(ctx context.Context, screen string, enabledOnly bool) ([]wallpaperRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	q := `
		SELECT id, screen, filename, enabled, sort_order, created_at
		FROM wallpapers WHERE screen = ?
	`
	if enabledOnly {
		q += ` AND enabled = 1`
	}
	q += ` ORDER BY sort_order ASC, id ASC`
	rows, err := h.db.QueryContext(ctx, q, screen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]wallpaperRow, 0)
	for rows.Next() {
		var w wallpaperRow
		var en int
		if err := rows.Scan(&w.ID, &w.Screen, &w.Filename, &en, &w.SortOrder, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.Enabled = en == 1
		w.FileURL = fmt.Sprintf("/api/wallpapers/%d/file", w.ID)
		out = append(out, w)
	}
	return out, rows.Err()
}
