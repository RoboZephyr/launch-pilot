package launchd

import (
	"strconv"
	"strings"
)

// ListEntry holds a single row from `launchctl list` output.
// This is the raw Layer-1 data before merging with plist info.
type ListEntry struct {
	Label          string
	PID            int // 0 = not running
	LastExitStatus int
}

// ParseListOutput parses the tab-separated output of `launchctl list`.
//
// Expected format (stable for 10+ years):
//
//	PID	Status	Label
//	584	0	com.example.myapp
//	-	0	com.example.stopped
//	-	78	com.example.broken
//
// Returns nil for empty input or header-only input.
// Malformed lines are silently skipped.
func ParseListOutput(output string) []ListEntry {
	if output == "" {
		return nil
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) == 0 {
		return nil
	}

	start := 0
	// Skip header line if present.
	if strings.HasPrefix(lines[0], "PID") {
		start = 1
	}

	if start >= len(lines) {
		return nil
	}

	var entries []ListEntry
	for _, line := range lines[start:] {
		if line == "" {
			continue
		}

		cols := strings.SplitN(line, "\t", 3)
		if len(cols) != 3 {
			continue
		}

		pid, err := parsePID(cols[0])
		if err != nil {
			continue
		}

		exitStatus, err := strconv.Atoi(cols[1])
		if err != nil {
			continue
		}

		entries = append(entries, ListEntry{
			Label:          cols[2],
			PID:            pid,
			LastExitStatus: exitStatus,
		})
	}

	if len(entries) == 0 {
		return nil
	}
	return entries
}

// parsePID converts the PID column: "-" means not running (0),
// otherwise parse as integer.
func parsePID(s string) (int, error) {
	if s == "-" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
