package launchd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A404coder/launch-pilot/internal/plist"
)

// ---------------------------------------------------------------------------
// tailFile
// ---------------------------------------------------------------------------

func TestTailFile_LastNLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := tailFile(path, 3)
	if err != nil {
		t.Fatalf("tailFile() error: %v", err)
	}

	want := "line3\nline4\nline5"
	if got != want {
		t.Errorf("tailFile(3) = %q, want %q", got, want)
	}
}

func TestTailFile_AllLinesWhenNExceedsTotal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := tailFile(path, 100)
	if err != nil {
		t.Fatalf("tailFile() error: %v", err)
	}

	want := "line1\nline2\nline3"
	if got != want {
		t.Errorf("tailFile(100) = %q, want %q", got, want)
	}
}

func TestTailFile_NonExistentFile(t *testing.T) {
	_, err := tailFile("/nonexistent/path/file.log", 10)
	if err == nil {
		t.Fatal("tailFile() should return error for non-existent file")
	}
}

func TestTailFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.log")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := tailFile(path, 10)
	if err != nil {
		t.Fatalf("tailFile() error: %v", err)
	}

	if got != "" {
		t.Errorf("tailFile(empty) = %q, want empty string", got)
	}
}

func TestTailFile_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noeol.log")
	content := "line1\nline2\nline3"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := tailFile(path, 2)
	if err != nil {
		t.Fatalf("tailFile() error: %v", err)
	}

	want := "line2\nline3"
	if got != want {
		t.Errorf("tailFile(2) = %q, want %q", got, want)
	}
}

func TestTailFile_ExactLineCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exact.log")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := tailFile(path, 3)
	if err != nil {
		t.Fatalf("tailFile() error: %v", err)
	}

	want := "line1\nline2\nline3"
	if got != want {
		t.Errorf("tailFile(3) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// ReadLogs
// ---------------------------------------------------------------------------

func TestReadLogs_NoLogPathsConfigured(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{Label: "com.example.myapp"},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	out, err := svc.ReadLogs("com.example.myapp", 200)
	if err != nil {
		t.Fatalf("ReadLogs() error: %v", err)
	}

	if out.Label != "com.example.myapp" {
		t.Errorf("Label: got %q", out.Label)
	}
	if out.Stdout != nil {
		t.Errorf("Stdout: expected nil, got %q", *out.Stdout)
	}
	if out.Stderr != nil {
		t.Errorf("Stderr: expected nil, got %q", *out.Stderr)
	}
	if out.StdoutAvailable {
		t.Error("StdoutAvailable: expected false")
	}
	if out.StderrAvailable {
		t.Error("StderrAvailable: expected false")
	}
	if out.Message == "" {
		t.Error("Message: expected non-empty message when no paths configured")
	}
}

func TestReadLogs_LogFilesDoNotExist(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{
				Label:             "com.example.myapp",
				StandardOutPath:   "/nonexistent/path/stdout.log",
				StandardErrorPath: "/nonexistent/path/stderr.log",
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	out, err := svc.ReadLogs("com.example.myapp", 200)
	if err != nil {
		t.Fatalf("ReadLogs() error: %v", err)
	}

	if out.Stdout != nil {
		t.Errorf("Stdout: expected nil, got %q", *out.Stdout)
	}
	if out.Stderr != nil {
		t.Errorf("Stderr: expected nil, got %q", *out.Stderr)
	}
	if out.StdoutPath != "/nonexistent/path/stdout.log" {
		t.Errorf("StdoutPath: got %q", out.StdoutPath)
	}
	if out.StderrPath != "/nonexistent/path/stderr.log" {
		t.Errorf("StderrPath: got %q", out.StderrPath)
	}
	if out.StdoutAvailable {
		t.Error("StdoutAvailable: expected false")
	}
	if out.StderrAvailable {
		t.Error("StderrAvailable: expected false")
	}
}

func TestReadLogs_ReadsLogFiles(t *testing.T) {
	dir := t.TempDir()
	stdoutPath := filepath.Join(dir, "stdout.log")
	stderrPath := filepath.Join(dir, "stderr.log")

	stdoutLines := make([]string, 10)
	for i := range stdoutLines {
		stdoutLines[i] = strings.Repeat("x", i+1)
	}
	os.WriteFile(stdoutPath, []byte(strings.Join(stdoutLines, "\n")+"\n"), 0644)
	os.WriteFile(stderrPath, []byte("err1\nerr2\n"), 0644)

	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{
				Label:             "com.example.myapp",
				StandardOutPath:   stdoutPath,
				StandardErrorPath: stderrPath,
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	out, err := svc.ReadLogs("com.example.myapp", 3)
	if err != nil {
		t.Fatalf("ReadLogs() error: %v", err)
	}

	if out.Stdout == nil {
		t.Fatal("Stdout: expected non-nil")
	}
	// Last 3 lines of stdout
	wantStdout := strings.Join(stdoutLines[7:], "\n")
	if *out.Stdout != wantStdout {
		t.Errorf("Stdout: got %q, want %q", *out.Stdout, wantStdout)
	}
	if !out.StdoutAvailable {
		t.Error("StdoutAvailable: expected true")
	}

	if out.Stderr == nil {
		t.Fatal("Stderr: expected non-nil")
	}
	if *out.Stderr != "err1\nerr2" {
		t.Errorf("Stderr: got %q, want %q", *out.Stderr, "err1\nerr2")
	}
	if !out.StderrAvailable {
		t.Error("StderrAvailable: expected true")
	}

	if out.StdoutPath != stdoutPath {
		t.Errorf("StdoutPath: got %q", out.StdoutPath)
	}
	if out.StderrPath != stderrPath {
		t.Errorf("StderrPath: got %q", out.StderrPath)
	}
}

func TestReadLogs_JobNotFound(t *testing.T) {
	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.other\n"
	svc := newTestService(listOutput, nil, nil)

	_, err := svc.ReadLogs("com.example.nonexistent", 200)
	if err == nil {
		t.Fatal("ReadLogs() should return error for non-existent job")
	}
}

func TestReadLogs_OnlyStdoutConfigured(t *testing.T) {
	dir := t.TempDir()
	stdoutPath := filepath.Join(dir, "stdout.log")
	os.WriteFile(stdoutPath, []byte("hello\nworld\n"), 0644)

	listOutput := "PID\tStatus\tLabel\n584\t0\tcom.example.myapp\n"
	plists := []plist.ScanResult{
		{
			Path: "/Users/testuser/Library/LaunchAgents/com.example.myapp.plist",
			Data: plist.PlistData{
				Label:           "com.example.myapp",
				StandardOutPath: stdoutPath,
				// No StandardErrorPath
			},
		},
	}

	svc := newTestService(listOutput, nil, plists)
	out, err := svc.ReadLogs("com.example.myapp", 200)
	if err != nil {
		t.Fatalf("ReadLogs() error: %v", err)
	}

	if out.Stdout == nil {
		t.Fatal("Stdout: expected non-nil")
	}
	if *out.Stdout != "hello\nworld" {
		t.Errorf("Stdout: got %q", *out.Stdout)
	}
	if !out.StdoutAvailable {
		t.Error("StdoutAvailable: expected true")
	}

	if out.Stderr != nil {
		t.Errorf("Stderr: expected nil, got %q", *out.Stderr)
	}
	if out.StderrAvailable {
		t.Error("StderrAvailable: expected false")
	}
}
