package launchd

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/A404coder/launch-pilot/internal/plist"
)

// Sentinel errors returned by Service methods.
var (
	ErrNotFound     = errors.New("job not found")
	ErrInvalidLabel = errors.New("invalid label")
)

// LabelRe validates launchd job labels: alphanumeric, dots, hyphens, underscores.
var LabelRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidLabel reports whether label contains only safe characters for launchctl args.
func ValidLabel(label string) bool {
	return LabelRe.MatchString(label)
}

// ValidateLabel checks that a label contains only safe characters for launchctl args.
func ValidateLabel(label string) error {
	if !ValidLabel(label) {
		return fmt.Errorf("%w: %q", ErrInvalidLabel, label)
	}
	return nil
}

// JobStatus represents the runtime state of a launchd job.
type JobStatus string

const (
	StatusRunning   JobStatus = "running"
	StatusScheduled JobStatus = "scheduled"
	StatusCompleted JobStatus = "completed"
	StatusStopped   JobStatus = "stopped"
	StatusError     JobStatus = "error"
	StatusOffline   JobStatus = "offline"
)

// Job holds merged data from launchctl list + plist file for a single launchd job.
type Job struct {
	Label                 string                `json:"label"`
	PID                   int                   `json:"pid"`
	LastExitStatus        int                   `json:"lastExitStatus"`
	Status                JobStatus             `json:"status"`
	PlistPath             string                `json:"plistPath"`
	Program               string                `json:"program"`
	ProgramArgs           []string              `json:"programArgs"`
	StandardOutPath       string                `json:"standardOutPath"`
	StandardErrPath       string                `json:"standardErrPath"`
	RunAtLoad             bool                  `json:"runAtLoad"`
	KeepAlive             bool                  `json:"keepAlive"`
	Domain                string                `json:"domain"`
	NextRunAt             *time.Time            `json:"nextRunAt,omitempty"`
	LastRunAt             *time.Time            `json:"lastRunAt,omitempty"`
	StartInterval         int                   `json:"startInterval,omitempty"`
	StartCalendarInterval []plist.CalendarEntry `json:"startCalendarInterval,omitempty"`
}

// DefaultRecentWindow is the default --recent-window value: how long after
// lastRunAt a scheduled job is still considered "completed".
const DefaultRecentWindow = 10 * time.Minute

// DeriveStatus determines the JobStatus for an online (launchctl-listed) job
// from PID, exit code, plist schedule shape, last-run heuristic, and the
// configured recent-completion window.
//
// Offline (plist present but not in `launchctl list`) is assigned by the
// service during the offline merge pass, not here.
func DeriveStatus(pid, exitStatus int, p plist.PlistData, lastRunAt *time.Time, now time.Time, window time.Duration) JobStatus {
	if pid > 0 {
		return StatusRunning
	}
	if exitStatus != 0 {
		return StatusError
	}
	if lastRunAt != nil {
		delta := now.Sub(*lastRunAt)
		if delta >= 0 && delta <= window {
			return StatusCompleted
		}
	}
	if p.StartInterval > 0 || len(p.StartCalendarInterval) > 0 || p.RunAtLoad {
		return StatusScheduled
	}
	return StatusStopped
}

// LogOutput holds the result of reading a job's stdout/stderr log files.
type LogOutput struct {
	Label           string  `json:"label"`
	Stdout          *string `json:"stdout"`
	Stderr          *string `json:"stderr"`
	StdoutPath      string  `json:"stdoutPath"`
	StderrPath      string  `json:"stderrPath"`
	StdoutAvailable bool    `json:"stdoutAvailable"`
	StderrAvailable bool    `json:"stderrAvailable"`
	Message         string  `json:"message,omitempty"`
}
