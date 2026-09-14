package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/ArminDashti/auto-wallpaper-api/internal/auth"
	"github.com/ArminDashti/auto-wallpaper-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Health is a public liveness probe.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Login authenticates and returns a session token (24h).
func (h *Handlers) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var userID int64
	var hash string
	err := h.db.QueryRowContext(ctx, `
		SELECT id, password_hash FROM users WHERE username = ? COLLATE NOCASE
	`, req.Username).Scan(&userID, &hash)
	if err == sql.ErrNoRows || !auth.CheckPassword(hash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	token, err := auth.NewToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token failed"})
		return
	}
	expires := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	if _, err := h.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)
	`, token, userID, expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expiresAt": expires,
		"username":  req.Username,
	})
}

// Logout deletes the current Bearer session when present.
func (h *Handlers) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, _ = h.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Me returns the authenticated user.
func (h *Handlers) Me(c *gin.Context) {
	username, _ := c.Get(middleware.ContextUsername)
	userID, _ := c.Get(middleware.ContextUserID)
	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
	})
}
