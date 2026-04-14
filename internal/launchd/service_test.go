package launchd

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/A404coder/launchboard/internal/plist"
)

// execCall records a single command invocation for test assertions.
type execCall struct {
	name string
	args []string
}

func argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

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

// ---------------------------------------------------------------------------
// S05: validateLabel
// ---------------------------------------------------------------------------

func TestValidateLabel(t *testing.T) {
	tests := []struct {
		label string
		valid bool
	}{
		// Valid labels
		{"com.apple.Finder", true},
		{"com.example.my-app_v2", true},
		{"com.example.myapp", true},
		{"myapp", true},
		{"a", true},
		{"A.B.C-d_e.123", true},

		// Invalid labels — injection attempts
		{"", false},
		{"com.example;rm -rf /", false},
		{"com.example.$(evil)", false},
		{"com.example evil", false},
		{"com.example/../etc/passwd", false},
		{"com.example.`evil`", false},
		{"com.example|evil", false},
		{"com.example&evil", false},
		{"com.example\nevil", false},
		{"com.example\tevil", false},
		{"label with spaces", false},
	}

	for _, tt := range tests {
		name := tt.label
		if name == "" {
			name = "(empty)"
		}
		t.Run(name, func(t *testing.T) {
			err := ValidateLabel(tt.label)
			if tt.valid && err != nil {
				t.Errorf("ValidateLabel(%q) unexpected error: %v", tt.label, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateLabel(%q) expected error, got nil", tt.label)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// S05: Reload
// ---------------------------------------------------------------------------

// newTestServiceWithExec extends newTestService with a runExec capture.
func newTestServiceWithExec(listOutput string, plists []plist.ScanResult, calls *[]execCall, execErr func(int) error) *Service {
	svc := newTestService(listOutput, nil, plists)
	callIdx := 0
	svc.runExec = func(name string, args ...string) (*ExecResult, error) {
		*calls = append(*calls, execCall{name, args})
		idx := callIdx
		callIdx++
		if execErr != nil {
			if err := execErr(idx); err != nil {
				return &ExecResult{}, err
			}
		}
		return &ExecResult{}, nil
	}
	return svc
}

func TestReload_Success(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{Label: "com.example.myapp"},
		},
	}

	var calls []execCall
	svc := newTestServiceWithExec(listOutput, plists, &calls, nil)

	err := svc.Reload("com.example.myapp")
	if err != nil {
		t.Fatalf("Reload() error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(calls))
	}

	// Call 1: bootout
	if calls[0].name != "launchctl" {
		t.Errorf("call 0 name: got %q", calls[0].name)
	}
	wantBootout := []string{"bootout", "gui/501/com.example.myapp"}
	if !argsEqual(calls[0].args, wantBootout) {
		t.Errorf("bootout args: got %v, want %v", calls[0].args, wantBootout)
	}

	// Call 2: bootstrap
	if calls[1].name != "launchctl" {
		t.Errorf("call 1 name: got %q", calls[1].name)
	}
	wantBootstrap := []string{"bootstrap", "gui/501", "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist"}
	if !argsEqual(calls[1].args, wantBootstrap) {
		t.Errorf("bootstrap args: got %v, want %v", calls[1].args, wantBootstrap)
	}
}

func TestReload_IgnoresBootoutError(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{Label: "com.example.myapp"},
		},
	}

	var calls []execCall
	svc := newTestServiceWithExec(listOutput, plists, &calls, func(idx int) error {
		if idx == 0 {
			return errors.New("not loaded")
		}
		return nil
	})

	err := svc.Reload("com.example.myapp")
	if err != nil {
		t.Fatalf("Reload() should succeed when bootout fails: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(calls))
	}
}

func TestReload_BootstrapFails(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{Label: "com.example.myapp"},
		},
	}

	var calls []execCall
	svc := newTestServiceWithExec(listOutput, plists, &calls, func(idx int) error {
		if idx == 1 {
			return errors.New("bootstrap failed: path not found")
		}
		return nil
	})

	err := svc.Reload("com.example.myapp")
	if err == nil {
		t.Fatal("Reload() should fail when bootstrap fails")
	}
}

func TestReload_JobNotFound(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.other\n"
	var calls []execCall
	svc := newTestServiceWithExec(listOutput, nil, &calls, nil)

	err := svc.Reload("com.example.nonexistent")
	if err == nil {
		t.Fatal("Reload() should fail when job not found")
	}
	if len(calls) != 0 {
		t.Errorf("should not exec any commands, got %d calls", len(calls))
	}
}

func TestReload_NoPlistPath(t *testing.T) {
	// Job exists in launchctl list but has no matching plist file.
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.noplist\n"
	var calls []execCall
	svc := newTestServiceWithExec(listOutput, nil, &calls, nil)

	err := svc.Reload("com.example.noplist")
	if err == nil {
		t.Fatal("Reload() should fail when plist path is empty")
	}
	if len(calls) != 0 {
		t.Errorf("should not exec any commands, got %d calls", len(calls))
	}
}

func TestReload_InvalidLabel(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, nil)

	err := svc.Reload("invalid;label")
	if err == nil {
		t.Fatal("Reload() should reject invalid label")
	}
	if len(calls) != 0 {
		t.Errorf("should not exec any commands for invalid label, got %d calls", len(calls))
	}
}

// ---------------------------------------------------------------------------
// S05: Start
// ---------------------------------------------------------------------------

func TestStart_Success(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, nil)

	err := svc.Start("com.example.myapp")
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(calls))
	}

	if calls[0].name != "launchctl" {
		t.Errorf("command: got %q", calls[0].name)
	}
	want := []string{"kickstart", "gui/501/com.example.myapp"}
	if !argsEqual(calls[0].args, want) {
		t.Errorf("args: got %v, want %v", calls[0].args, want)
	}
}

func TestStart_InvalidLabel(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, nil)

	err := svc.Start("$(evil)")
	if err == nil {
		t.Fatal("Start() should reject invalid label")
	}
	if len(calls) != 0 {
		t.Errorf("should not exec any commands, got %d calls", len(calls))
	}
}

func TestStart_ExecFails(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, func(_ int) error {
		return fmt.Errorf("kickstart failed")
	})

	err := svc.Start("com.example.myapp")
	if err == nil {
		t.Fatal("Start() should propagate exec error")
	}
}

// ---------------------------------------------------------------------------
// S05: Stop
// ---------------------------------------------------------------------------

func TestStop_Success(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, nil)

	err := svc.Stop("com.example.myapp")
	if err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(calls))
	}

	if calls[0].name != "launchctl" {
		t.Errorf("command: got %q", calls[0].name)
	}
	want := []string{"kill", "SIGTERM", "gui/501/com.example.myapp"}
	if !argsEqual(calls[0].args, want) {
		t.Errorf("args: got %v, want %v", calls[0].args, want)
	}
}

func TestStop_InvalidLabel(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, nil)

	err := svc.Stop("label with spaces")
	if err == nil {
		t.Fatal("Stop() should reject invalid label")
	}
	if len(calls) != 0 {
		t.Errorf("should not exec any commands, got %d calls", len(calls))
	}
}

func TestStop_ExecFails(t *testing.T) {
	var calls []execCall
	svc := newTestServiceWithExec("", nil, &calls, func(_ int) error {
		return fmt.Errorf("kill failed")
	})

	err := svc.Stop("com.example.myapp")
	if err == nil {
		t.Fatal("Stop() should propagate exec error")
	}
}
