package launchd

import (
	"encoding/json"
	"testing"
)

func TestJobStatus_Values(t *testing.T) {
	// Verify the three status constants exist and have expected string values.
	tests := []struct {
		status JobStatus
		want   string
	}{
		{StatusRunning, "running"},
		{StatusStopped, "stopped"},
		{StatusError, "error"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("expected %q, got %q", tt.want, tt.status)
		}
	}
}

func TestJob_DeriveStatus(t *testing.T) {
	tests := []struct {
		name   string
		pid    int
		exit   int
		want   JobStatus
	}{
		{"running process", 584, 0, StatusRunning},
		{"running with prior error", 1234, 78, StatusRunning},
		{"stopped normally", 0, 0, StatusStopped},
		{"error exit code", 0, 78, StatusError},
		{"signal killed", 0, -9, StatusError},
		{"negative exit", 0, -15, StatusError},
		{"exit code 1", 0, 1, StatusError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveStatus(tt.pid, tt.exit)
			if got != tt.want {
				t.Errorf("DeriveStatus(%d, %d) = %q, want %q", tt.pid, tt.exit, got, tt.want)
			}
		})
	}
}

func TestJob_JSONRoundTrip(t *testing.T) {
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
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify all fields round-trip correctly.
	if decoded.Label != original.Label {
		t.Errorf("Label: got %q, want %q", decoded.Label, original.Label)
	}
	if decoded.PID != original.PID {
		t.Errorf("PID: got %d, want %d", decoded.PID, original.PID)
	}
	if decoded.LastExitStatus != original.LastExitStatus {
		t.Errorf("LastExitStatus: got %d, want %d", decoded.LastExitStatus, original.LastExitStatus)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.PlistPath != original.PlistPath {
		t.Errorf("PlistPath: got %q, want %q", decoded.PlistPath, original.PlistPath)
	}
	if decoded.Program != original.Program {
		t.Errorf("Program: got %q, want %q", decoded.Program, original.Program)
	}
	if len(decoded.ProgramArgs) != len(original.ProgramArgs) {
		t.Fatalf("ProgramArgs length: got %d, want %d", len(decoded.ProgramArgs), len(original.ProgramArgs))
	}
	for i, arg := range decoded.ProgramArgs {
		if arg != original.ProgramArgs[i] {
			t.Errorf("ProgramArgs[%d]: got %q, want %q", i, arg, original.ProgramArgs[i])
		}
	}
	if decoded.StandardOutPath != original.StandardOutPath {
		t.Errorf("StandardOutPath: got %q, want %q", decoded.StandardOutPath, original.StandardOutPath)
	}
	if decoded.StandardErrPath != original.StandardErrPath {
		t.Errorf("StandardErrPath: got %q, want %q", decoded.StandardErrPath, original.StandardErrPath)
	}
	if decoded.RunAtLoad != original.RunAtLoad {
		t.Errorf("RunAtLoad: got %v, want %v", decoded.RunAtLoad, original.RunAtLoad)
	}
	if decoded.KeepAlive != original.KeepAlive {
		t.Errorf("KeepAlive: got %v, want %v", decoded.KeepAlive, original.KeepAlive)
	}
	if decoded.Domain != original.Domain {
		t.Errorf("Domain: got %q, want %q", decoded.Domain, original.Domain)
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
}

func TestJob_EmptyProgramArgs(t *testing.T) {
	// ProgramArgs should marshal as empty array, not null.
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
