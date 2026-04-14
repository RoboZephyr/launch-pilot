package plist

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_Get(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "com.example.cached.plist")
	if err := os.WriteFile(path, []byte(samplePlistXML), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	cache := NewCache()

	t.Run("first call parses file (cache miss)", func(t *testing.T) {
		data, err := cache.Get(path)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if data.Label != "com.example.myapp" {
			t.Errorf("Label = %q, want %q", data.Label, "com.example.myapp")
		}
	})

	t.Run("second call returns cached entry (cache hit)", func(t *testing.T) {
		data, err := cache.Get(path)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if data.Label != "com.example.myapp" {
			t.Errorf("Label = %q, want %q", data.Label, "com.example.myapp")
		}
	})

	t.Run("modified file triggers re-parse", func(t *testing.T) {
		// Ensure mtime differs — some filesystems have 1s resolution
		time.Sleep(10 * time.Millisecond)

		updated := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.example.updated</string>
</dict>
</plist>`
		// Touch with a future mtime to guarantee cache invalidation
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			t.Fatalf("rewrite plist: %v", err)
		}
		// Force distinct mtime
		future := time.Now().Add(1 * time.Hour)
		os.Chtimes(path, future, future)

		data, err := cache.Get(path)
		if err != nil {
			t.Fatalf("Get after modify: %v", err)
		}
		if data.Label != "com.example.updated" {
			t.Errorf("Label = %q, want %q (cache should have been invalidated)", data.Label, "com.example.updated")
		}
	})

	t.Run("deleted file returns error", func(t *testing.T) {
		gone := filepath.Join(dir, "gone.plist")
		_, err := cache.Get(gone)
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})
}

func TestCache_Concurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.plist")
	if err := os.WriteFile(path, []byte(samplePlistXML), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	cache := NewCache()
	done := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			data, err := cache.Get(path)
			if err != nil {
				t.Errorf("concurrent Get: %v", err)
				return
			}
			if data.Label != "com.example.myapp" {
				t.Errorf("concurrent Label = %q, want %q", data.Label, "com.example.myapp")
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
