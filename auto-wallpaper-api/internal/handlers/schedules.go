package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type scheduleRow struct {
	Screen         string `json:"screen"`
	Mode           string `json:"mode"`
	IntervalHours  int    `json:"intervalHours"`
	StartHour      int    `json:"startHour"`
	Enabled        bool   `json:"enabled"`
}

type scheduleUpdate struct {
	Mode          *string `json:"mode"`
	IntervalHours *int    `json:"intervalHours"`
	StartHour     *int    `json:"startHour"`
	Enabled       *bool   `json:"enabled"`
}

func validScreen(s string) bool {
	return s == "lock" || s == "home"
}

// GetSchedule returns schedule for lock or home.
func (h *Handlers) GetSchedule(c *gin.Context) {
	screen := c.Param("screen")
	if !validScreen(screen) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screen must be lock or home"})
		return
	}
	row, err := h.loadSchedule(c.Request.Context(), screen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load schedule failed"})
		return
	}
	c.JSON(http.StatusOK, row)
}

// PutSchedule updates schedule for a screen.
func (h *Handlers) PutSchedule(c *gin.Context) {
	screen := c.Param("screen")
	if !validScreen(screen) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screen must be lock or home"})
		return
	}
	var req scheduleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	cur, err := h.loadSchedule(c.Request.Context(), screen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load schedule failed"})
		return
	}
	if req.Mode != nil {
		if *req.Mode != "daily" && *req.Mode != "hourly" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be daily or hourly"})
			return
		}
		cur.Mode = *req.Mode
	}
	if req.IntervalHours != nil {
		ih := *req.IntervalHours
		if ih != 1 && ih != 2 && ih != 4 && ih != 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "intervalHours must be 1, 2, 4, or 8"})
			return
		}
		cur.IntervalHours = ih
	}
	if req.StartHour != nil {
		if *req.StartHour < 0 || *req.StartHour > 23 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "startHour must be 0-23"})
			return
		}
		cur.StartHour = *req.StartHour
	}
	if req.Enabled != nil {
		cur.Enabled = *req.Enabled
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	en := 0
	if cur.Enabled {
		en = 1
	}
	_, err = h.db.ExecContext(ctx, `
		UPDATE schedules
		SET mode = ?, interval_hours = ?, start_hour = ?, enabled = ?
		WHERE screen = ?
	`, cur.Mode, cur.IntervalHours, cur.StartHour, en, screen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update schedule failed"})
		return
	}
	c.JSON(http.StatusOK, cur)
}

func (h *Handlers) loadSchedule(ctx context.Context, screen string) (scheduleRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var row scheduleRow
	var en int
	err := h.db.QueryRowContext(ctx, `
		SELECT screen, mode, interval_hours, start_hour, enabled
		FROM schedules WHERE screen = ?
	`, screen).Scan(&row.Screen, &row.Mode, &row.IntervalHours, &row.StartHour, &en)
	if err == sql.ErrNoRows {
		return scheduleRow{
			Screen: screen, Mode: "daily", IntervalHours: 1, StartHour: 0, Enabled: true,
		}, nil
	}
	if err != nil {
		return row, err
	}
	row.Enabled = en == 1
	return row, nil
}
