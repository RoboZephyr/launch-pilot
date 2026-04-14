package launchd

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/A404coder/launchboard/internal/plist"
)

// newTestService creates a Service with injectable list output and plist scan results.
func newTestService(listOutput string, listErr error, plists []plist.ScanResult) *Service {
	return &Service{
		uid:     501,
		domain:  "gui/501",
		homeDir: "/Users/testuser",
		runList: func() (string, error) {
			return listOutput, listErr
		},
		scanPlist: func() []plist.ScanResult {
			return plists
		},
	}
}

func TestListJobs_MergesPlistData(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"584\t0\tcom.example.myapp\n" +
		"-\t0\tcom.example.stopped\n" +
		"-\t78\tcom.example.broken\n"

	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{
				Label:             "com.example.myapp",
				Program:           "/usr/local/bin/myapp",
				ProgramArguments:  []string{"/usr/local/bin/myapp", "--daemon"},
				StandardOutPath:   "/tmp/myapp.stdout.log",
				StandardErrorPath: "/tmp/myapp.stderr.log",
				RunAtLoad:         true,
				KeepAlive:         false,
			},
		},
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.stopped.plist",
			Data: plist.PlistData{
				Label:            "com.example.stopped",
				ProgramArguments: []string{"/usr/local/bin/stopped"},
				RunAtLoad:        false,
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	jobs, err := svc.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs() error: %v", err)
	}

	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}

	// Build label→job map for easier assertions.
	byLabel := make(map[string]Job)
	for _, j := range jobs {
		byLabel[j.Label] = j
	}

	// com.example.myapp — running, merged with plist
	app := byLabel["com.example.myapp"]
	if app.PID != 584 {
		t.Errorf("myapp PID: got %d, want 584", app.PID)
	}
	if app.Status != StatusRunning {
		t.Errorf("myapp Status: got %q, want %q", app.Status, StatusRunning)
	}
	if app.Program != "/usr/local/bin/myapp" {
		t.Errorf("myapp Program: got %q, want %q", app.Program, "/usr/local/bin/myapp")
	}
	if len(app.ProgramArgs) != 2 || app.ProgramArgs[1] != "--daemon" {
		t.Errorf("myapp ProgramArgs: got %v", app.ProgramArgs)
	}
	if app.StandardOutPath != "/tmp/myapp.stdout.log" {
		t.Errorf("myapp StandardOutPath: got %q", app.StandardOutPath)
	}
	if app.StandardErrPath != "/tmp/myapp.stderr.log" {
		t.Errorf("myapp StandardErrPath: got %q", app.StandardErrPath)
	}
	if !app.RunAtLoad {
		t.Error("myapp RunAtLoad: expected true")
	}
	if app.KeepAlive {
		t.Error("myapp KeepAlive: expected false")
	}
	if app.PlistPath != "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist" {
		t.Errorf("myapp PlistPath: got %q", app.PlistPath)
	}
	if app.Domain != "user" {
		t.Errorf("myapp Domain: got %q, want %q", app.Domain, "user")
	}

	// com.example.stopped — stopped, merged with plist (no Program field, uses ProgramArguments[0])
	stopped := byLabel["com.example.stopped"]
	if stopped.Status != StatusStopped {
		t.Errorf("stopped Status: got %q, want %q", stopped.Status, StatusStopped)
	}
	if stopped.Program != "/usr/local/bin/stopped" {
		t.Errorf("stopped Program: got %q, want %q", stopped.Program, "/usr/local/bin/stopped")
	}

	// com.example.broken — error, no matching plist
	broken := byLabel["com.example.broken"]
	if broken.Status != StatusError {
		t.Errorf("broken Status: got %q, want %q", broken.Status, StatusError)
	}
	if broken.LastExitStatus != 78 {
		t.Errorf("broken LastExitStatus: got %d, want 78", broken.LastExitStatus)
	}
	if broken.PlistPath != "" {
		t.Errorf("broken PlistPath: expected empty, got %q", broken.PlistPath)
	}
}

func TestListJobs_NoMatchingPlist(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"100\t0\tcom.example.noplist\n"

	svc := newTestService(listOutput, nil, nil)
	jobs, err := svc.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs() error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.Label != "com.example.noplist" {
		t.Errorf("Label: got %q", job.Label)
	}
	if job.PID != 100 {
		t.Errorf("PID: got %d, want 100", job.PID)
	}
	if job.Status != StatusRunning {
		t.Errorf("Status: got %q, want %q", job.Status, StatusRunning)
	}
	if job.Program != "" {
		t.Errorf("Program: expected empty, got %q", job.Program)
	}
	if job.PlistPath != "" {
		t.Errorf("PlistPath: expected empty, got %q", job.PlistPath)
	}
}

func TestListJobs_CommandError(t *testing.T) {
	svc := newTestService("", errors.New("launchctl failed"), nil)
	_, err := svc.ListJobs()
	if err == nil {
		t.Fatal("expected error from ListJobs when command fails")
	}
}

func TestListJobs_EmptyOutput(t *testing.T) {
	svc := newTestService("PID\tStatus\tLabel\n", nil, nil)
	jobs, err := svc.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs() error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestListJobs_DomainDetection(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"100\t0\tcom.example.user\n" +
		"200\t0\tcom.example.global\n"

	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.user.plist",
			Data: plist.PlistData{Label: "com.example.user"},
		},
		{
			Path: "/Library/LaunchAgents/com.example.global.plist",
			Data: plist.PlistData{Label: "com.example.global"},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	jobs, err := svc.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs() error: %v", err)
	}

	byLabel := make(map[string]Job)
	for _, j := range jobs {
		byLabel[j.Label] = j
	}

	if byLabel["com.example.user"].Domain != "user" {
		t.Errorf("user domain: got %q, want %q", byLabel["com.example.user"].Domain, "user")
	}
	if byLabel["com.example.global"].Domain != "global" {
		t.Errorf("global domain: got %q, want %q", byLabel["com.example.global"].Domain, "global")
	}
}

func TestListJobs_ProgramFallbackToArgs(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"-\t0\tcom.example.argsonly\n"

	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.argsonly.plist",
			Data: plist.PlistData{
				Label:            "com.example.argsonly",
				ProgramArguments: []string{"/usr/bin/env", "python3", "script.py"},
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	jobs, err := svc.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs() error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	// When Program is empty, should use ProgramArguments[0]
	if jobs[0].Program != "/usr/bin/env" {
		t.Errorf("Program: got %q, want %q", jobs[0].Program, "/usr/bin/env")
	}
	if len(jobs[0].ProgramArgs) != 3 {
		t.Errorf("ProgramArgs length: got %d, want 3", len(jobs[0].ProgramArgs))
	}
}

func TestGetJob_Found(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"584\t0\tcom.example.myapp\n" +
		"-\t0\tcom.example.other\n"

	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{
				Label:   "com.example.myapp",
				Program: "/usr/local/bin/myapp",
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	job, err := svc.GetJob("com.example.myapp")
	if err != nil {
		t.Fatalf("GetJob() error: %v", err)
	}

	if job.Label != "com.example.myapp" {
		t.Errorf("Label: got %q", job.Label)
	}
	if job.PID != 584 {
		t.Errorf("PID: got %d, want 584", job.PID)
	}
	if job.Program != "/usr/local/bin/myapp" {
		t.Errorf("Program: got %q", job.Program)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n" +
		"584\t0\tcom.example.myapp\n"

	svc := newTestService(listOutput, nil, nil)
	_, err := svc.GetJob("com.example.nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent label")
	}
}

func TestGetJob_CommandError(t *testing.T) {
	svc := newTestService("", errors.New("launchctl failed"), nil)
	_, err := svc.GetJob("com.example.any")
	if err == nil {
		t.Fatal("expected error when command fails")
	}
}

func TestDetectDomain(t *testing.T) {
	svc := &Service{homeDir: "/Users/testuser"}

	tests := []struct {
		name      string
		plistPath string
		want      string
	}{
		{
			"user agent",
			"/Users/testuser/Library/LaunchAgents/com.example.plist",
			"user",
		},
		{
			"global agent",
			"/Library/LaunchAgents/com.example.plist",
			"global",
		},
		{
			"unknown path defaults to user",
			"/some/random/path/com.example.plist",
			"user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.detectDomain(tt.plistPath)
			if got != tt.want {
				t.Errorf("detectDomain(%q) = %q, want %q", tt.plistPath, got, tt.want)
			}
		})
	}
}

func TestDetectDomain_GlobalPathNotConfusedWithUser(t *testing.T) {
	// Ensure /Library/LaunchAgents is not matched as user when home is /Library
	svc := &Service{homeDir: "/Users/testuser"}
	userPath := filepath.Join(svc.homeDir, "Library", "LaunchAgents", "test.plist")
	globalPath := "/Library/LaunchAgents/test.plist"

	if got := svc.detectDomain(userPath); got != "user" {
		t.Errorf("user path: got %q, want %q", got, "user")
	}
	if got := svc.detectDomain(globalPath); got != "global" {
		t.Errorf("global path: got %q, want %q", got, "global")
	}
}
