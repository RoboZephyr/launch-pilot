package launchd

import (
	"errors"
	"fmt"
	"regexp"
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
	StatusRunning JobStatus = "running" // PID > 0
	StatusStopped JobStatus = "stopped" // PID == 0, exit == 0
	StatusError   JobStatus = "error"   // PID == 0, exit != 0
)

// Job holds merged data from launchctl list + plist file for a single launchd job.
type Job struct {
	Label           string    `json:"label"`
	PID             int       `json:"pid"`
	LastExitStatus  int       `json:"lastExitStatus"`
	Status          JobStatus `json:"status"`
	PlistPath       string    `json:"plistPath"`
	Program         string    `json:"program"`
	ProgramArgs     []string  `json:"programArgs"`
	StandardOutPath string    `json:"standardOutPath"`
	StandardErrPath string    `json:"standardErrPath"`
	RunAtLoad       bool      `json:"runAtLoad"`
	KeepAlive       bool      `json:"keepAlive"`
	Domain          string    `json:"domain"`
}

// DeriveStatus determines the JobStatus from PID and last exit code.
func DeriveStatus(pid, exitStatus int) JobStatus {
	if pid > 0 {
		return StatusRunning
	}
	if exitStatus != 0 {
		return StatusError
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
