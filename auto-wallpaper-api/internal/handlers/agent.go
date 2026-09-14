package handlers

import (
	"net/http"
	"time"

	"github.com/ArminDashti/auto-wallpaper-api/internal/schedule"
	"github.com/gin-gonic/gin"
)

type agentScreenState struct {
	Screen            string         `json:"screen"`
	Schedule          scheduleRow    `json:"schedule"`
	Wallpapers        []wallpaperRow `json:"wallpapers"`
	CurrentWallpaper  *wallpaperRow  `json:"currentWallpaper"`
	SlotIndex         int64          `json:"slotIndex"`
}

// AgentState returns both screens' schedules, enabled wallpapers, and current pick.
func (h *Handlers) AgentState(c *gin.Context) {
	now := time.Now()
	screens := []string{"lock", "home"}
	out := make([]agentScreenState, 0, 2)
	for _, screen := range screens {
		sched, err := h.loadSchedule(c.Request.Context(), screen)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "load schedule failed"})
			return
		}
		walls, err := h.listWallpapers(c.Request.Context(), screen, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list wallpapers failed"})
			return
		}
		slot := schedule.SlotIndex(now, sched.Mode, sched.IntervalHours, sched.StartHour)
		st := agentScreenState{
			Screen:     screen,
			Schedule:   sched,
			Wallpapers: walls,
			SlotIndex:  slot,
		}
		if sched.Enabled && len(walls) > 0 {
			idx := schedule.PickIndex(slot, len(walls))
			if idx >= 0 && idx < len(walls) {
				w := walls[idx]
				st.CurrentWallpaper = &w
			}
		}
		out = append(out, st)
	}
	c.JSON(http.StatusOK, gin.H{
		"now":     now.Format(time.RFC3339),
		"screens": out,
	})
}
