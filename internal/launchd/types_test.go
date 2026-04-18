package launchd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/A404coder/launch-pilot/internal/plist"
)

func intPtr(v int) *int { return &v }

func TestJobStatus_Values(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   string
	}{
		{StatusRunning, "running"},
		{StatusScheduled, "scheduled"},
		{StatusCompleted, "completed"},
		{StatusStopped, "stopped"},
		{StatusError, "error"},
		{StatusOffline, "offline"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("expected %q, got %q", tt.want, tt.status)
		}
	}
}

func TestJob_DeriveStatus(t *testing.T) {
	now := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	recent := now.Add(-2 * time.Minute)
	old := now.Add(-2 * time.Hour)

	recentPtr := &recent
	oldPtr := &old

	calSchedule := plist.PlistData{
		StartCalendarInterval: plist.CalendarEntries{
			{Hour: intPtr(9), Minute: intPtr(0)},
		},
	}
	intervalSchedule := plist.PlistData{StartInterval: 300}
	runAtLoadOnly := plist.PlistData{RunAtLoad: true}
	emptyPlist := plist.PlistData{}

	tests := []struct {
		name    string
		pid     int
		exit    int
		plist   plist.PlistData
		lastRun *time.Time
		want    JobStatus
	}{
		{
			name:    "pid>0 always running even with schedule",
			pid:     584,
			exit:    0,
			plist:   calSchedule,
			lastRun: nil,
			want:    StatusRunning,
		},
		{
			name:    "pid>0 running overrides prior nonzero exit",
			pid:     1234,
			exit:    78,
			plist:   emptyPlist,
			lastRun: nil,
			want:    StatusRunning,
		},
		{
			name:    "pid=0 with nonzero exit is error",
			pid:     0,
			exit:    78,
			plist:   calSchedule,
			lastRun: recentPtr,
			want:    StatusError,
		},
		{
			name:    "pid=0 exit=0 with recent lastRun is completed (calendar schedule)",
			pid:     0,
			exit:    0,
			plist:   calSchedule,
			lastRun: recentPtr,
			want:    StatusCompleted,
		},
		{
			name:    "pid=0 exit=0 with recent lastRun is completed (interval schedule)",
			pid:     0,
			exit:    0,
			plist:   intervalSchedule,
			lastRun: recentPtr,
			want:    StatusCompleted,
		},
		{
			name:    "pid=0 exit=0 with calendar schedule and no lastRun is scheduled",
			pid:     0,
			exit:    0,
			plist:   calSchedule,
			lastRun: nil,
			want:    StatusScheduled,
		},
		{
			name:    "pid=0 exit=0 with interval schedule and no lastRun is scheduled",
			pid:     0,
			exit:    0,
			plist:   intervalSchedule,
			lastRun: nil,
			want:    StatusScheduled,
		},
		{
			name:    "pid=0 exit=0 with stale lastRun and schedule is scheduled",
			pid:     0,
			exit:    0,
			plist:   calSchedule,
			lastRun: oldPtr,
			want:    StatusScheduled,
		},
		{
			name:    "pid=0 exit=0 with RunAtLoad only is scheduled",
			pid:     0,
			exit:    0,
			plist:   runAtLoadOnly,
			lastRun: nil,
			want:    StatusScheduled,
		},
		{
			name:    "pid=0 exit=0 with stale lastRun and no schedule is stopped",
			pid:     0,
			exit:    0,
			plist:   emptyPlist,
			lastRun: oldPtr,
			want:    StatusStopped,
		},
		{
			name:    "pid=0 exit=0 with no plist and no lastRun is stopped",
			pid:     0,
			exit:    0,
			plist:   emptyPlist,
			lastRun: nil,
			want:    StatusStopped,
		},
		{
			name:    "signal-killed exit is error",
			pid:     0,
			exit:    -9,
			plist:   emptyPlist,
			lastRun: nil,
			want:    StatusError,
		},
	}

	window := 10 * time.Minute
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveStatus(tt.pid, tt.exit, tt.plist, tt.lastRun, now, window)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeriveStatus_WindowBoundary(t *testing.T) {
	now := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	window := 10 * time.Minute
	exactlyWindow := now.Add(-window)
	justOver := now.Add(-window - time.Second)

	schedule := plist.PlistData{StartInterval: 300}

	if got := DeriveStatus(0, 0, schedule, &exactlyWindow, now, window); got != StatusCompleted {
		t.Errorf("lastRun exactly at window should be completed, got %q", got)
	}
	if got := DeriveStatus(0, 0, schedule, &justOver, now, window); got != StatusScheduled {
		t.Errorf("lastRun just beyond window should be scheduled, got %q", got)
	}
}

func TestJob_JSONRoundTrip(t *testing.T) {
	next := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	last := time.Date(2026, 4, 17, 9, 0, 0, 0, time.UTC)

	original := Job{
		Label:           "com.example.myapp",
		PID:             584,
		LastExitStatus:  0,
		Status:          StatusRunning,
		PlistPath:       "/Users/me/Library/LaunchAgents/com.example.myapp.plist",
		Program:         "/usr/local/bin/myapp",
		ProgramArgs:     []string{"/usr/local/bin/myapp", "--daemon"},
		StandardOutPath: "/tmp/myapp.stdout.log",
		StandardErrPath: "/tmp/myapp.stderr.log",
		RunAtLoad:       true,
		KeepAlive:       false,
		Domain:          "user",
		NextRunAt:       &next,
		LastRunAt:       &last,
		StartInterval:   300,
		StartCalendarInterval: []plist.CalendarEntry{
			{Hour: intPtr(9), Minute: intPtr(0)},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Label != original.Label {
		t.Errorf("Label: got %q, want %q", decoded.Label, original.Label)
	}
	if decoded.PID != original.PID {
		t.Errorf("PID: got %d, want %d", decoded.PID, original.PID)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.NextRunAt == nil || !decoded.NextRunAt.Equal(next) {
		t.Errorf("NextRunAt: got %v, want %v", decoded.NextRunAt, next)
	}
	if decoded.LastRunAt == nil || !decoded.LastRunAt.Equal(last) {
		t.Errorf("LastRunAt: got %v, want %v", decoded.LastRunAt, last)
	}
	if decoded.StartInterval != 300 {
		t.Errorf("StartInterval: got %d, want 300", decoded.StartInterval)
	}
	if len(decoded.StartCalendarInterval) != 1 {
		t.Fatalf("StartCalendarInterval length: got %d, want 1", len(decoded.StartCalendarInterval))
	}
	got0 := decoded.StartCalendarInterval[0]
	if got0.Hour == nil || *got0.Hour != 9 {
		t.Errorf("entry[0].Hour: got %v, want 9", got0.Hour)
	}
}

func TestJob_JSONFieldNames(t *testing.T) {
	job := Job{
		Label:           "com.test.app",
		PID:             100,
		LastExitStatus:  0,
		Status:          StatusRunning,
		PlistPath:       "/path/to/plist",
		Program:         "/bin/test",
		ProgramArgs:     []string{"/bin/test"},
		StandardOutPath: "/tmp/out.log",
		StandardErrPath: "/tmp/err.log",
		RunAtLoad:       true,
		KeepAlive:       true,
		Domain:          "user",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	expectedFields := []string{
		"label", "pid", "lastExitStatus", "status",
		"plistPath", "program", "programArgs",
		"standardOutPath", "standardErrPath",
		"runAtLoad", "keepAlive", "domain",
	}

	for _, field := range expectedFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected JSON field %q not found", field)
		}
	}

	for _, field := range []string{"nextRunAt", "lastRunAt", "startInterval", "startCalendarInterval"} {
		if _, ok := raw[field]; ok {
			t.Errorf("expected nil/zero field %q to be omitted, but it was present", field)
		}
	}
}

func TestJob_EmptyProgramArgs(t *testing.T) {
	job := Job{
		Label:       "com.test.empty",
		ProgramArgs: []string{},
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if string(raw["programArgs"]) != "[]" {
		t.Errorf("empty ProgramArgs should be [], got %s", raw["programArgs"])
	}
}
