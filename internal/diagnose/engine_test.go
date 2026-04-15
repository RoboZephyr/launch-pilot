package diagnose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/A404coder/launch-pilot/internal/launchd"
)

// --- checkExitCode ---

func TestCheckExitCode_Zero(t *testing.T) {
	job := &launchd.Job{Label: "com.test.ok", LastExitStatus: 0}
	r := checkExitCode(job)
	if r.ID != "exit-code" {
		t.Fatalf("ID = %q, want %q", r.ID, "exit-code")
	}
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q", r.Severity, SeverityOK)
	}
	if r.Suggestion != "" {
		t.Errorf("Suggestion = %q, want empty for exit 0", r.Suggestion)
	}
}

func TestCheckExitCode_NonZero(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"general error", 1},
		{"config error", 78},
		{"not found", 127},
		{"signal kill", -9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &launchd.Job{Label: "com.test.fail", LastExitStatus: tt.code}
			r := checkExitCode(job)
			if r.Severity != SeverityError {
				t.Errorf("code %d: Severity = %q, want %q", tt.code, r.Severity, SeverityError)
			}
			if r.Message == "" {
				t.Errorf("code %d: Message should not be empty", tt.code)
			}
		})
	}
}

// --- checkProgramExists ---

func TestCheckProgramExists_BinSh(t *testing.T) {
	job := &launchd.Job{Label: "com.test.sh", Program: "/bin/sh"}
	r := checkProgramExists(job)
	if r.ID != "program-exists" {
		t.Fatalf("ID = %q, want %q", r.ID, "program-exists")
	}
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q for /bin/sh", r.Severity, SeverityOK)
	}
}

func TestCheckProgramExists_Missing(t *testing.T) {
	job := &launchd.Job{Label: "com.test.missing", Program: "/nonexistent/path/to/binary"}
	r := checkProgramExists(job)
	if r.Severity != SeverityError {
		t.Errorf("Severity = %q, want %q for missing program", r.Severity, SeverityError)
	}
	if r.Suggestion == "" {
		t.Error("Suggestion should not be empty for missing program")
	}
}

func TestCheckProgramExists_EmptyProgram(t *testing.T) {
	// When no Program is set, fall back to ProgramArgs[0]
	job := &launchd.Job{Label: "com.test.args", ProgramArgs: []string{"/bin/sh", "-c", "echo hi"}}
	r := checkProgramExists(job)
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q when ProgramArgs[0] = /bin/sh", r.Severity, SeverityOK)
	}
}

func TestCheckProgramExists_NoProgramAtAll(t *testing.T) {
	job := &launchd.Job{Label: "com.test.none"}
	r := checkProgramExists(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q when no program configured", r.Severity, SeverityWarning)
	}
}

// --- checkProgramExecutable ---

func TestCheckProgramExecutable_BinSh(t *testing.T) {
	job := &launchd.Job{Label: "com.test.sh", Program: "/bin/sh"}
	r := checkProgramExecutable(job)
	if r.ID != "program-executable" {
		t.Fatalf("ID = %q, want %q", r.ID, "program-executable")
	}
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q for /bin/sh", r.Severity, SeverityOK)
	}
}

func TestCheckProgramExecutable_NoExecPerm(t *testing.T) {
	// Create a temp file without execute permission.
	tmp := t.TempDir()
	f := filepath.Join(tmp, "nox")
	if err := os.WriteFile(f, []byte("#!/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}

	job := &launchd.Job{Label: "com.test.nox", Program: f}
	r := checkProgramExecutable(job)
	if r.Severity != SeverityError {
		t.Errorf("Severity = %q, want %q for non-executable file", r.Severity, SeverityError)
	}
	if r.Suggestion == "" {
		t.Error("Suggestion should not be empty for non-executable file")
	}
}

func TestCheckProgramExecutable_Missing(t *testing.T) {
	// If the file doesn't exist, skip with warning — checkProgramExists already covers it.
	job := &launchd.Job{Label: "com.test.missing", Program: "/nonexistent/binary"}
	r := checkProgramExecutable(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q when file doesn't exist", r.Severity, SeverityWarning)
	}
}

// --- checkPlistOwner ---

func TestCheckPlistOwner_OwnedByCurrentUser(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "com.test.plist")
	if err := os.WriteFile(f, []byte("<plist/>"), 0644); err != nil {
		t.Fatal(err)
	}
	job := &launchd.Job{Label: "com.test.owner", PlistPath: f}
	r := checkPlistOwner(job)
	if r.ID != "plist-owner" {
		t.Fatalf("ID = %q, want %q", r.ID, "plist-owner")
	}
	// File we just created should be owned by current user → ok.
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q for own file", r.Severity, SeverityOK)
	}
}

func TestCheckPlistOwner_NoPlistPath(t *testing.T) {
	job := &launchd.Job{Label: "com.test.nopath"}
	r := checkPlistOwner(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q when no plist path", r.Severity, SeverityWarning)
	}
}

// --- checkPlistPerms ---

func TestCheckPlistPerms_Safe(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "safe.plist")
	if err := os.WriteFile(f, []byte("<plist/>"), 0644); err != nil {
		t.Fatal(err)
	}
	job := &launchd.Job{Label: "com.test.safe", PlistPath: f}
	r := checkPlistPerms(job)
	if r.ID != "plist-perms" {
		t.Fatalf("ID = %q, want %q", r.ID, "plist-perms")
	}
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q for 0644 perms", r.Severity, SeverityOK)
	}
}

func TestCheckPlistPerms_GroupWorldWritable(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "unsafe.plist")
	if err := os.WriteFile(f, []byte("<plist/>"), 0644); err != nil {
		t.Fatal(err)
	}
	// Explicitly set group/world writable bits (bypasses umask).
	if err := os.Chmod(f, 0666); err != nil {
		t.Fatal(err)
	}
	job := &launchd.Job{Label: "com.test.unsafe", PlistPath: f}
	r := checkPlistPerms(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q for group/world writable", r.Severity, SeverityWarning)
	}
	if r.Suggestion == "" {
		t.Error("Suggestion should not be empty for writable plist")
	}
}

func TestCheckPlistPerms_NoPlistPath(t *testing.T) {
	job := &launchd.Job{Label: "com.test.nopath"}
	r := checkPlistPerms(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q when no plist path", r.Severity, SeverityWarning)
	}
}

// --- checkLogPathExists ---

func TestCheckLogPathExists_BothExist(t *testing.T) {
	tmp := t.TempDir()
	stdout := filepath.Join(tmp, "stdout.log")
	stderr := filepath.Join(tmp, "stderr.log")
	os.WriteFile(stdout, []byte("log"), 0644)
	os.WriteFile(stderr, []byte("err"), 0644)

	job := &launchd.Job{Label: "com.test.logs", StandardOutPath: stdout, StandardErrPath: stderr}
	r := checkLogPathExists(job)
	if r.ID != "log-path-exists" {
		t.Fatalf("ID = %q, want %q", r.ID, "log-path-exists")
	}
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q when both log dirs exist", r.Severity, SeverityOK)
	}
}

func TestCheckLogPathExists_ParentMissing(t *testing.T) {
	job := &launchd.Job{
		Label:           "com.test.nologdir",
		StandardOutPath: "/nonexistent/dir/stdout.log",
	}
	r := checkLogPathExists(job)
	if r.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want %q when log parent dir missing", r.Severity, SeverityWarning)
	}
}

func TestCheckLogPathExists_NoPaths(t *testing.T) {
	job := &launchd.Job{Label: "com.test.nologs"}
	r := checkLogPathExists(job)
	if r.Severity != SeverityOK {
		t.Errorf("Severity = %q, want %q when no log paths configured", r.Severity, SeverityOK)
	}
}

// --- Diagnose (integration) ---

func TestDiagnose_ReturnsExactly6Checks(t *testing.T) {
	job := &launchd.Job{
		Label:   "com.test.full",
		Program: "/bin/sh",
	}
	e := &Engine{}
	report := e.Diagnose(job)

	if report.Label != job.Label {
		t.Errorf("Label = %q, want %q", report.Label, job.Label)
	}
	if len(report.Checks) != 6 {
		t.Fatalf("len(Checks) = %d, want 6", len(report.Checks))
	}

	// Verify all 6 check IDs are present.
	wantIDs := map[string]bool{
		"exit-code":          false,
		"program-exists":     false,
		"program-executable": false,
		"plist-owner":        false,
		"plist-perms":        false,
		"log-path-exists":    false,
	}
	for _, c := range report.Checks {
		if _, ok := wantIDs[c.ID]; !ok {
			t.Errorf("unexpected check ID %q", c.ID)
		}
		wantIDs[c.ID] = true
	}
	for id, seen := range wantIDs {
		if !seen {
			t.Errorf("missing check ID %q", id)
		}
	}
}

func TestDiagnose_HealthyJob(t *testing.T) {
	job := &launchd.Job{
		Label:          "com.test.healthy",
		PID:            1234,
		LastExitStatus: 0,
		Program:        "/bin/sh",
	}
	e := &Engine{}
	report := e.Diagnose(job)

	for _, c := range report.Checks {
		if c.ID == "exit-code" && c.Severity != SeverityOK {
			t.Errorf("exit-code: Severity = %q, want ok for running job", c.Severity)
		}
		if c.ID == "program-exists" && c.Severity != SeverityOK {
			t.Errorf("program-exists: Severity = %q, want ok for /bin/sh", c.Severity)
		}
		if c.ID == "program-executable" && c.Severity != SeverityOK {
			t.Errorf("program-executable: Severity = %q, want ok for /bin/sh", c.Severity)
		}
	}
}

func TestDiagnose_BrokenJob(t *testing.T) {
	job := &launchd.Job{
		Label:          "com.test.broken",
		PID:            0,
		LastExitStatus: 78,
		Program:        "/nonexistent/binary",
	}
	e := &Engine{}
	report := e.Diagnose(job)

	for _, c := range report.Checks {
		if c.ID == "exit-code" && c.Severity != SeverityError {
			t.Errorf("exit-code: Severity = %q, want error for exit 78", c.Severity)
		}
		if c.ID == "program-exists" && c.Severity != SeverityError {
			t.Errorf("program-exists: Severity = %q, want error for missing binary", c.Severity)
		}
	}
}
