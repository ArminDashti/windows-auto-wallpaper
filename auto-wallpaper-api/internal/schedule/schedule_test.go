package schedule_test

import (
	"testing"
	"time"

	"github.com/ArminDashti/auto-wallpaper-api/internal/schedule"
)

func TestPickIndex(t *testing.T) {
	if schedule.PickIndex(0, 0) != -1 {
		t.Fatal("empty")
	}
	if schedule.PickIndex(5, 3) != 2 {
		t.Fatal("wrap")
	}
}

func TestSlotIndexDaily(t *testing.T) {
	loc := time.Local
	// 10:00 with start 8 → same day slot
	now := time.Date(2026, 3, 10, 10, 0, 0, 0, loc)
	a := schedule.SlotIndex(now, "daily", 1, 8)
	before := time.Date(2026, 3, 10, 7, 0, 0, 0, loc)
	b := schedule.SlotIndex(before, "daily", 1, 8)
	if a != b+1 {
		t.Fatalf("daily before start should be previous day: a=%d b=%d", a, b)
	}
}

func TestSlotIndexHourly(t *testing.T) {
	loc := time.Local
	// start 8, interval 4: 8,12,16,20
	t8 := schedule.SlotIndex(time.Date(2026, 3, 10, 8, 0, 0, 0, loc), "hourly", 4, 8)
	t12 := schedule.SlotIndex(time.Date(2026, 3, 10, 12, 0, 0, 0, loc), "hourly", 4, 8)
	t11 := schedule.SlotIndex(time.Date(2026, 3, 10, 11, 59, 0, 0, loc), "hourly", 4, 8)
	if t12 != t8+1 {
		t.Fatalf("expected next slot at 12: t8=%d t12=%d", t8, t12)
	}
	if t11 != t8 {
		t.Fatalf("11:59 still first slot: t8=%d t11=%d", t8, t11)
	}
}
