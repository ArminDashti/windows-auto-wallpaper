package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the gin context key for authenticated user id.
	ContextUserID = "userID"
	// ContextUsername is the gin context key for authenticated username.
	ContextUsername = "username"
)

// RequireAuth validates Bearer session tokens against the sessions table.
func RequireAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var userID int64
		var username string
		err := db.QueryRowContext(ctx, `
			SELECT u.id, u.username
			FROM sessions s
			JOIN users u ON u.id = s.user_id
			WHERE s.token = ? AND s.expires_at > datetime('now')
		`, token).Scan(&userID, &username)
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextUsername, username)
		c.Next()
	}
}
