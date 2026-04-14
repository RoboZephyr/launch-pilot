package plist

import (
	"os"
	"path/filepath"
	"testing"
)

// samplePlistXML is a minimal valid LaunchAgent plist in XML format.
const samplePlistXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.example.myapp</string>
	<key>Program</key>
	<string>/usr/local/bin/myapp</string>
	<key>ProgramArguments</key>
	<array>
		<string>/usr/local/bin/myapp</string>
		<string>--daemon</string>
	</array>
	<key>StandardOutPath</key>
	<string>/tmp/myapp.stdout.log</string>
	<key>StandardErrorPath</key>
	<string>/tmp/myapp.stderr.log</string>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
	<key>WorkingDirectory</key>
	<string>/usr/local</string>
	<key>StartInterval</key>
	<integer>300</integer>
</dict>
</plist>`

// minimalPlistXML has only Label — all other fields are zero-values.
const minimalPlistXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.example.minimal</string>
</dict>
</plist>`

func writeTempPlist(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp plist: %v", err)
	}
	return path
}

func TestReadPlist(t *testing.T) {
	dir := t.TempDir()

	t.Run("full plist decodes all fields", func(t *testing.T) {
		path := writeTempPlist(t, dir, "com.example.myapp.plist", samplePlistXML)

		data, err := ReadPlist(path)
		if err != nil {
			t.Fatalf("ReadPlist: %v", err)
		}
		if data.Label != "com.example.myapp" {
			t.Errorf("Label = %q, want %q", data.Label, "com.example.myapp")
		}
		if data.Program != "/usr/local/bin/myapp" {
			t.Errorf("Program = %q, want %q", data.Program, "/usr/local/bin/myapp")
		}
		wantArgs := []string{"/usr/local/bin/myapp", "--daemon"}
		if len(data.ProgramArguments) != len(wantArgs) {
			t.Fatalf("ProgramArguments length = %d, want %d", len(data.ProgramArguments), len(wantArgs))
		}
		for i, arg := range wantArgs {
			if data.ProgramArguments[i] != arg {
				t.Errorf("ProgramArguments[%d] = %q, want %q", i, data.ProgramArguments[i], arg)
			}
		}
		if data.StandardOutPath != "/tmp/myapp.stdout.log" {
			t.Errorf("StandardOutPath = %q, want %q", data.StandardOutPath, "/tmp/myapp.stdout.log")
		}
		if data.StandardErrorPath != "/tmp/myapp.stderr.log" {
			t.Errorf("StandardErrorPath = %q, want %q", data.StandardErrorPath, "/tmp/myapp.stderr.log")
		}
		if !data.RunAtLoad {
			t.Error("RunAtLoad = false, want true")
		}
		if data.KeepAlive {
			t.Error("KeepAlive = true, want false")
		}
		if data.WorkingDirectory != "/usr/local" {
			t.Errorf("WorkingDirectory = %q, want %q", data.WorkingDirectory, "/usr/local")
		}
		if data.StartInterval != 300 {
			t.Errorf("StartInterval = %d, want 300", data.StartInterval)
		}
	})

	t.Run("minimal plist has zero-value defaults", func(t *testing.T) {
		path := writeTempPlist(t, dir, "com.example.minimal.plist", minimalPlistXML)

		data, err := ReadPlist(path)
		if err != nil {
			t.Fatalf("ReadPlist: %v", err)
		}
		if data.Label != "com.example.minimal" {
			t.Errorf("Label = %q, want %q", data.Label, "com.example.minimal")
		}
		if data.Program != "" {
			t.Errorf("Program = %q, want empty", data.Program)
		}
		if len(data.ProgramArguments) != 0 {
			t.Errorf("ProgramArguments = %v, want empty", data.ProgramArguments)
		}
		if data.RunAtLoad {
			t.Error("RunAtLoad = true, want false")
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := ReadPlist(filepath.Join(dir, "nonexistent.plist"))
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})

	t.Run("invalid plist content returns error", func(t *testing.T) {
		path := writeTempPlist(t, dir, "bad.plist", "this is not a plist")

		_, err := ReadPlist(path)
		if err == nil {
			t.Fatal("expected error for invalid plist, got nil")
		}
	})
}

func TestScanDirs(t *testing.T) {
	dirs := ScanDirs()
	if len(dirs) == 0 {
		t.Fatal("ScanDirs returned empty slice")
	}

	// Must include user LaunchAgents dir
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	userDir := filepath.Join(home, "Library", "LaunchAgents")
	found := false
	for _, d := range dirs {
		if d == userDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ScanDirs missing user dir %q, got %v", userDir, dirs)
	}

	// Must include global LaunchAgents dir
	globalDir := "/Library/LaunchAgents"
	found = false
	for _, d := range dirs {
		if d == globalDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ScanDirs missing global dir %q, got %v", globalDir, dirs)
	}
}

func TestScanAll(t *testing.T) {
	t.Run("reads plist files from directory", func(t *testing.T) {
		dir := t.TempDir()
		writeTempPlist(t, dir, "com.example.one.plist", samplePlistXML)
		writeTempPlist(t, dir, "com.example.two.plist", minimalPlistXML)
		// Non-plist file should be ignored
		os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("ignore me"), 0644)

		results := ScanAll([]string{dir})
		if len(results) != 2 {
			t.Fatalf("ScanAll returned %d results, want 2", len(results))
		}

		labels := map[string]bool{}
		for _, r := range results {
			labels[r.Data.Label] = true
			if r.Path == "" {
				t.Error("result has empty Path")
			}
		}
		if !labels["com.example.myapp"] {
			t.Error("missing label com.example.myapp")
		}
		if !labels["com.example.minimal"] {
			t.Error("missing label com.example.minimal")
		}
	})

	t.Run("skips non-existent directories", func(t *testing.T) {
		results := ScanAll([]string{"/nonexistent/path/12345"})
		if len(results) != 0 {
			t.Fatalf("ScanAll returned %d results for non-existent dir, want 0", len(results))
		}
	})

	t.Run("skips unparseable plist files", func(t *testing.T) {
		dir := t.TempDir()
		writeTempPlist(t, dir, "good.plist", samplePlistXML)
		writeTempPlist(t, dir, "bad.plist", "not a plist at all")

		results := ScanAll([]string{dir})
		if len(results) != 1 {
			t.Fatalf("ScanAll returned %d results, want 1 (bad file should be skipped)", len(results))
		}
		if results[0].Data.Label != "com.example.myapp" {
			t.Errorf("Label = %q, want %q", results[0].Data.Label, "com.example.myapp")
		}
	})

	t.Run("multiple directories merged", func(t *testing.T) {
		dir1 := t.TempDir()
		dir2 := t.TempDir()
		writeTempPlist(t, dir1, "one.plist", samplePlistXML)
		writeTempPlist(t, dir2, "two.plist", minimalPlistXML)

		results := ScanAll([]string{dir1, dir2})
		if len(results) != 2 {
			t.Fatalf("ScanAll returned %d results, want 2", len(results))
		}
	})
}
