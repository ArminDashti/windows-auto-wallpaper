package schedule

import "time"

// SlotIndex returns how many schedule slots have elapsed from epoch alignment,
// used to pick which wallpaper in a sequential rotation is current.
//
// daily: one slot per calendar day after start_hour on that day.
// hourly: slots every intervalHours, aligned to startHour within each day.
func SlotIndex(now time.Time, mode string, intervalHours, startHour int) int64 {
	now = now.Local()
	if startHour < 0 {
		startHour = 0
	}
	if startHour > 23 {
		startHour = 23
	}

	switch mode {
	case "hourly":
		if intervalHours != 1 && intervalHours != 2 && intervalHours != 4 && intervalHours != 8 {
			intervalHours = 1
		}
		// Days since local epoch midnight, then hour slots within day from startHour.
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dayIndex := midnight.Unix() / 86400
		hour := now.Hour()
		// Hours since startHour today (may be negative → treat as previous day's last slots).
		offset := hour - startHour
		if offset < 0 {
			dayIndex--
			offset += 24
		}
		slotsToday := offset / intervalHours
		slotsPerDay := 24 / intervalHours
		return dayIndex*int64(slotsPerDay) + int64(slotsToday)
	default: // daily
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dayIndex := midnight.Unix() / 86400
		if now.Hour() < startHour {
			dayIndex--
		}
		return dayIndex
	}
}

// PickIndex maps slot index onto wallpaper count (sequential wrap).
func PickIndex(slot int64, count int) int {
	if count <= 0 {
		return -1
	}
	if slot < 0 {
		slot = -slot
	}
	return int(slot % int64(count))
}
