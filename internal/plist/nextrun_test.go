package plist

import (
	"testing"
	"time"
)

func intPtr(v int) *int { return &v }

func TestNextCalendarFire(t *testing.T) {
	tests := []struct {
		name    string
		entries []CalendarEntry
		now     time.Time
		want    time.Time
	}{
		{
			name:    "every_hour_at_minute_0",
			entries: []CalendarEntry{{Minute: intPtr(0)}},
			now:     time.Date(2026, 4, 17, 14, 30, 45, 0, time.UTC),
			want:    time.Date(2026, 4, 17, 15, 0, 0, 0, time.UTC),
		},
		{
			name:    "daily_at_9_30",
			entries: []CalendarEntry{{Hour: intPtr(9), Minute: intPtr(30)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC), // past today's 9:30
			want:    time.Date(2026, 4, 18, 9, 30, 0, 0, time.UTC),
		},
		{
			name:    "friday_11am",
			entries: []CalendarEntry{{Weekday: intPtr(5), Hour: intPtr(11), Minute: intPtr(0)}},
			// 2026-04-17 is a Friday. Now = Friday 10:59 → next fire = Friday 11:00.
			now:  time.Date(2026, 4, 17, 10, 59, 0, 0, time.UTC),
			want: time.Date(2026, 4, 17, 11, 0, 0, 0, time.UTC),
		},
		{
			name:    "monthly_first_at_9am",
			entries: []CalendarEntry{{Day: intPtr(1), Hour: intPtr(9), Minute: intPtr(0)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			name:    "cross_year_dec_31_to_jan_1",
			entries: []CalendarEntry{{Month: intPtr(1), Day: intPtr(1), Hour: intPtr(0), Minute: intPtr(0)}},
			now:     time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC),
			want:    time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "impossible_feb_31_returns_zero",
			entries: []CalendarEntry{{Month: intPtr(2), Day: intPtr(31), Hour: intPtr(0), Minute: intPtr(0)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Time{},
		},
		{
			name: "multiple_entries_return_earliest",
			entries: []CalendarEntry{
				{Hour: intPtr(17), Minute: intPtr(0)}, // later today
				{Hour: intPtr(12), Minute: intPtr(0)}, // earlier today
			},
			now:  time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC),
		},
		{
			name:    "empty_entries_returns_zero",
			entries: nil,
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Time{},
		},
		{
			name:    "all_nil_fields_returns_zero",
			entries: []CalendarEntry{{}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Time{},
		},
		{
			name:    "invalid_minute_returns_zero",
			entries: []CalendarEntry{{Minute: intPtr(60)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Time{},
		},
		{
			name:    "invalid_month_returns_zero",
			entries: []CalendarEntry{{Month: intPtr(13), Minute: intPtr(0)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Time{},
		},
		{
			name:    "sunday_weekday_0",
			entries: []CalendarEntry{{Weekday: intPtr(0), Hour: intPtr(6), Minute: intPtr(0)}},
			// 2026-04-17 is Friday. Next Sunday is 2026-04-19.
			now:  time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want: time.Date(2026, 4, 19, 6, 0, 0, 0, time.UTC),
		},
		{
			name:    "sunday_weekday_7_alias",
			entries: []CalendarEntry{{Weekday: intPtr(7), Hour: intPtr(6), Minute: intPtr(0)}},
			now:     time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
			want:    time.Date(2026, 4, 19, 6, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextCalendarFire(tt.entries, tt.now)
			if !got.Equal(tt.want) {
				t.Errorf("NextCalendarFire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextIntervalFire(t *testing.T) {
	now := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		intervalSec int
		lastRun     time.Time
		want        time.Time
	}{
		{
			name:        "zero_interval_returns_zero",
			intervalSec: 0,
			lastRun:     time.Time{},
			want:        time.Time{},
		},
		{
			name:        "negative_interval_returns_zero",
			intervalSec: -60,
			lastRun:     time.Time{},
			want:        time.Time{},
		},
		{
			name:        "no_lastrun_uses_now",
			intervalSec: 300,
			lastRun:     time.Time{},
			want:        now.Add(300 * time.Second),
		},
		{
			name:        "future_schedule_from_lastrun",
			intervalSec: 600,
			lastRun:     now.Add(-60 * time.Second), // last ran 1 min ago
			want:        now.Add(-60*time.Second + 600*time.Second),
		},
		{
			name:        "stale_lastrun_in_past_bumps_to_now",
			intervalSec: 60,
			lastRun:     now.Add(-3600 * time.Second), // last ran 1 hour ago; interval 60s → overdue
			want:        now.Add(60 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextIntervalFire(tt.intervalSec, tt.lastRun, now)
			if !got.Equal(tt.want) {
				t.Errorf("NextIntervalFire(%d, %v, %v) = %v, want %v",
					tt.intervalSec, tt.lastRun, now, got, tt.want)
			}
		})
	}
}
