package plist

import "time"

// hasAnyField reports whether the entry has at least one non-nil field,
// i.e. any actual constraint. An all-nil entry matches everything, which is
// meaningless as a schedule and returns zero time.
func (e CalendarEntry) hasAnyField() bool {
	return e.Minute != nil || e.Hour != nil || e.Day != nil || e.Weekday != nil || e.Month != nil
}

// valid reports whether every set field holds a launchd-acceptable value.
// Out-of-range integers (e.g. Minute=60, Month=13) can never match, so we
// bail out rather than scan to the 2-year horizon one minute at a time.
func (e CalendarEntry) valid() bool {
	if e.Minute != nil && (*e.Minute < 0 || *e.Minute > 59) {
		return false
	}
	if e.Hour != nil && (*e.Hour < 0 || *e.Hour > 23) {
		return false
	}
	if e.Day != nil && (*e.Day < 1 || *e.Day > 31) {
		return false
	}
	if e.Weekday != nil && (*e.Weekday < 0 || *e.Weekday > 7) {
		return false
	}
	if e.Month != nil && (*e.Month < 1 || *e.Month > 12) {
		return false
	}
	return true
}

// matches reports whether t satisfies every non-nil field of e. Weekday 7 is
// treated as Sunday (0), matching launchd's accepted aliasing.
func (e CalendarEntry) matches(t time.Time) bool {
	if e.Minute != nil && t.Minute() != *e.Minute {
		return false
	}
	if e.Hour != nil && t.Hour() != *e.Hour {
		return false
	}
	if e.Day != nil && t.Day() != *e.Day {
		return false
	}
	if e.Month != nil && int(t.Month()) != *e.Month {
		return false
	}
	if e.Weekday != nil {
		want := *e.Weekday
		if want == 7 {
			want = 0
		}
		if int(t.Weekday()) != want {
			return false
		}
	}
	return true
}

// nextFire returns the earliest time strictly after now at which entry matches,
// or zero time if no match occurs within the 2-year safety horizon (guards
// against impossible combinations like Day=31,Month=2).
func (e CalendarEntry) nextFire(now time.Time) time.Time {
	if !e.hasAnyField() || !e.valid() {
		return time.Time{}
	}

	horizon := now.AddDate(2, 0, 0)
	// Start at the next whole minute — launchd has minute granularity.
	candidate := now.Truncate(time.Minute).Add(time.Minute)

	for candidate.Before(horizon) || candidate.Equal(horizon) {
		// Advance by the coarsest unit that mismatches, resetting finer units.
		if e.Month != nil && int(candidate.Month()) != *e.Month {
			y, m := candidate.Year(), int(candidate.Month())
			nextMonth := m + 1
			nextYear := y
			if nextMonth > 12 {
				nextMonth = 1
				nextYear++
			}
			candidate = time.Date(nextYear, time.Month(nextMonth), 1, 0, 0, 0, 0, candidate.Location())
			continue
		}
		if e.Day != nil && candidate.Day() != *e.Day {
			candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, candidate.Location()).AddDate(0, 0, 1)
			continue
		}
		if e.Weekday != nil {
			want := *e.Weekday
			if want == 7 {
				want = 0
			}
			if int(candidate.Weekday()) != want {
				candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, candidate.Location()).AddDate(0, 0, 1)
				continue
			}
		}
		if e.Hour != nil && candidate.Hour() != *e.Hour {
			candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), candidate.Hour(), 0, 0, 0, candidate.Location()).Add(time.Hour)
			continue
		}
		if e.Minute != nil && candidate.Minute() != *e.Minute {
			candidate = candidate.Add(time.Minute)
			continue
		}
		return candidate
	}
	return time.Time{}
}

// NextCalendarFire returns the smallest t > now at which any of the given
// calendar entries matches. Returns zero time if entries is empty, if all
// entries are all-nil, or if no match occurs within a 2-year horizon.
func NextCalendarFire(entries []CalendarEntry, now time.Time) time.Time {
	var best time.Time
	for _, e := range entries {
		fire := e.nextFire(now)
		if fire.IsZero() {
			continue
		}
		if best.IsZero() || fire.Before(best) {
			best = fire
		}
	}
	return best
}

// NextIntervalFire computes the next fire time for a StartInterval job.
//   - intervalSec <= 0: no schedule, returns zero time.
//   - lastRun zero:    returns now + interval.
//   - lastRun+interval in the future: returns that.
//   - otherwise (overdue): returns now + interval.
func NextIntervalFire(intervalSec int, lastRun, now time.Time) time.Time {
	if intervalSec <= 0 {
		return time.Time{}
	}
	interval := time.Duration(intervalSec) * time.Second
	if lastRun.IsZero() {
		return now.Add(interval)
	}
	candidate := lastRun.Add(interval)
	if candidate.After(now) {
		return candidate
	}
	return now.Add(interval)
}
