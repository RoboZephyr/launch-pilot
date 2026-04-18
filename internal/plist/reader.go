package plist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

// CalendarEntry mirrors launchd's StartCalendarInterval dict entry. Every field
// is an optional integer: absent (nil) means "match any value" for that field.
// See launchd.plist(5).
type CalendarEntry struct {
	Minute  *int `plist:"Minute"`
	Hour    *int `plist:"Hour"`
	Day     *int `plist:"Day"`
	Weekday *int `plist:"Weekday"`
	Month   *int `plist:"Month"`
}

// CalendarEntries is a slice of CalendarEntry with a custom UnmarshalPlist
// that transparently accepts either a single <dict> or an <array> of <dict>s,
// both of which are valid StartCalendarInterval shapes.
type CalendarEntries []CalendarEntry

// UnmarshalPlist implements howett.net/plist's Unmarshaler interface.
func (c *CalendarEntries) UnmarshalPlist(unmarshal func(any) error) error {
	var arr []CalendarEntry
	if err := unmarshal(&arr); err == nil {
		*c = arr
		return nil
	}
	var single CalendarEntry
	if err := unmarshal(&single); err != nil {
		return err
	}
	*c = CalendarEntries{single}
	return nil
}

// PlistData holds the decoded fields from a LaunchAgent/LaunchDaemon plist file.
type PlistData struct {
	Label                 string          `plist:"Label"`
	Program               string          `plist:"Program"`
	ProgramArguments      []string        `plist:"ProgramArguments"`
	StandardOutPath       string          `plist:"StandardOutPath"`
	StandardErrorPath     string          `plist:"StandardErrorPath"`
	RunAtLoad             bool            `plist:"RunAtLoad"`
	KeepAlive             bool            `plist:"KeepAlive"`
	WorkingDirectory      string          `plist:"WorkingDirectory"`
	StartInterval         int             `plist:"StartInterval"`
	StartCalendarInterval CalendarEntries `plist:"StartCalendarInterval"`
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

// ScanAll reads all .plist files from the given directories, using the provided
// cache for mtime-based caching. If cache is nil, files are parsed directly.
// Non-existent directories and unparseable files are silently skipped.
func ScanAll(dirs []string, cache *Cache) []ScanResult {
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

			var data *PlistData
			if cache != nil {
				data, err = cache.Get(path)
			} else {
				data, err = ReadPlist(path)
			}
			if err != nil {
				continue // skip unparseable files
			}
			results = append(results, ScanResult{Path: path, Data: *data})
		}
	}
	return results
}
