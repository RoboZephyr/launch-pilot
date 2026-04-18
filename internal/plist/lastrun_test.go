package plist

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"
)

// fakeFileInfo implements os.FileInfo with a controllable ModTime.
type fakeFileInfo struct {
	name    string
	mtime   time.Time
	size    int64
	mode    os.FileMode
	isDir   bool
	sys     any
}

func (f *fakeFileInfo) Name() string       { return f.name }
func (f *fakeFileInfo) Size() int64        { return f.size }
func (f *fakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f *fakeFileInfo) ModTime() time.Time { return f.mtime }
func (f *fakeFileInfo) IsDir() bool        { return f.isDir }
func (f *fakeFileInfo) Sys() any           { return f.sys }

func TestLastRunAt(t *testing.T) {
	t0 := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(1 * time.Hour)
	t2 := t0.Add(2 * time.Hour)

	t.Run("newer of stdout and stderr wins", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			switch path {
			case "/tmp/out.log":
				return &fakeFileInfo{name: "out.log", mtime: t1}, nil
			case "/tmp/err.log":
				return &fakeFileInfo{name: "err.log", mtime: t2}, nil
			}
			return nil, errors.New("unexpected path")
		}
		got := LastRunAt("/tmp/out.log", "/tmp/err.log", stat)
		if !got.Equal(t2) {
			t.Errorf("got %v, want %v (stderr newer)", got, t2)
		}
	})

	t.Run("only stdout exists", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			if path == "/tmp/out.log" {
				return &fakeFileInfo{name: "out.log", mtime: t1}, nil
			}
			return nil, fs.ErrNotExist
		}
		got := LastRunAt("/tmp/out.log", "/tmp/err.log", stat)
		if !got.Equal(t1) {
			t.Errorf("got %v, want %v", got, t1)
		}
	})

	t.Run("only stderr exists", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			if path == "/tmp/err.log" {
				return &fakeFileInfo{name: "err.log", mtime: t2}, nil
			}
			return nil, fs.ErrNotExist
		}
		got := LastRunAt("/tmp/out.log", "/tmp/err.log", stat)
		if !got.Equal(t2) {
			t.Errorf("got %v, want %v", got, t2)
		}
	})

	t.Run("both paths empty returns zero time", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			t.Fatalf("statFn should not be called for empty paths: got %q", path)
			return nil, nil
		}
		got := LastRunAt("", "", stat)
		if !got.IsZero() {
			t.Errorf("got %v, want zero time", got)
		}
	})

	t.Run("both paths fail stat returns zero time", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			return nil, fs.ErrNotExist
		}
		got := LastRunAt("/tmp/missing.out", "/tmp/missing.err", stat)
		if !got.IsZero() {
			t.Errorf("got %v, want zero time", got)
		}
	})

	t.Run("empty stdout path but valid stderr path", func(t *testing.T) {
		stat := func(path string) (os.FileInfo, error) {
			if path == "/tmp/err.log" {
				return &fakeFileInfo{name: "err.log", mtime: t1}, nil
			}
			t.Fatalf("statFn called with unexpected path: %q", path)
			return nil, nil
		}
		got := LastRunAt("", "/tmp/err.log", stat)
		if !got.Equal(t1) {
			t.Errorf("got %v, want %v", got, t1)
		}
	})
}
