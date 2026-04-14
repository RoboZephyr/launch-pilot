package plist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

// PlistData holds the decoded fields from a LaunchAgent/LaunchDaemon plist file.
type PlistData struct {
	Label             string   `plist:"Label"`
	Program           string   `plist:"Program"`
	ProgramArguments  []string `plist:"ProgramArguments"`
	StandardOutPath   string   `plist:"StandardOutPath"`
	StandardErrorPath string   `plist:"StandardErrorPath"`
	RunAtLoad         bool     `plist:"RunAtLoad"`
	KeepAlive         bool     `plist:"KeepAlive"`
	WorkingDirectory  string   `plist:"WorkingDirectory"`
	StartInterval     int      `plist:"StartInterval"`
}

// ScanResult pairs a parsed plist with its file path on disk.
type ScanResult struct {
	Path string
	Data PlistData
}

// ReadPlist decodes a single plist file into PlistData.
func ReadPlist(path string) (*PlistData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plist %s: %w", path, err)
	}
	defer f.Close()

	var data PlistData
	decoder := plist.NewDecoder(f)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode plist %s: %w", path, err)
	}
	return &data, nil
}

// ScanDirs returns the standard LaunchAgent directories to scan.
// Includes ~/Library/LaunchAgents (user) and /Library/LaunchAgents (global).
func ScanDirs() []string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Library", "LaunchAgents"))
	}
	dirs = append(dirs, "/Library/LaunchAgents")
	return dirs
}

// ScanAll reads all .plist files from the given directories.
// Non-existent directories and unparseable files are silently skipped.
func ScanAll(dirs []string) []ScanResult {
	var results []ScanResult
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // skip non-existent or unreadable directories
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			data, err := ReadPlist(path)
			if err != nil {
				continue // skip unparseable files
			}
			results = append(results, ScanResult{Path: path, Data: *data})
		}
	}
	return results
}
