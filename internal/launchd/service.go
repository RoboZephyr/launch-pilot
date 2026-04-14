package launchd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/A404coder/launchboard/internal/plist"
)

// labelRe validates launchd job labels: alphanumeric, dots, hyphens, underscores.
var labelRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Service provides high-level operations on launchd user-domain jobs.
// It merges data from launchctl list (Layer 1) with plist files (Layer 2).
type Service struct {
	uid     int
	domain  string // "gui/<UID>"
	homeDir string

	// Overridable for testing.
	runList   func() (string, error)
	scanPlist func() []plist.ScanResult
	runExec   func(name string, args ...string) (*ExecResult, error)
}

// NewService creates a Service configured for the current user.
func NewService() *Service {
	uid := os.Getuid()
	home, _ := os.UserHomeDir()
	s := &Service{
		uid:     uid,
		domain:  fmt.Sprintf("gui/%d", uid),
		homeDir: home,
	}
	s.runList = s.defaultRunList
	s.scanPlist = s.defaultScanPlist
	s.runExec = func(name string, args ...string) (*ExecResult, error) {
		return runCmd(name, args...)
	}
	return s
}

func (s *Service) defaultRunList() (string, error) {
	result, err := runCmd("launchctl", "list")
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func (s *Service) defaultScanPlist() []plist.ScanResult {
	return plist.ScanAll(plist.ScanDirs())
}

// ListJobs returns all launchd jobs visible to the current user, with plist
// data merged when a matching plist file is found.
func (s *Service) ListJobs() ([]Job, error) {
	output, err := s.runList()
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	entries := ParseListOutput(output)
	if len(entries) == 0 {
		return []Job{}, nil
	}

	// Build label → plist lookup from scanned plist files.
	plistMap := make(map[string]plist.ScanResult)
	for _, sr := range s.scanPlist() {
		plistMap[sr.Data.Label] = sr
	}

	jobs := make([]Job, 0, len(entries))
	for _, e := range entries {
		job := Job{
			Label:          e.Label,
			PID:            e.PID,
			LastExitStatus: e.LastExitStatus,
			Status:         DeriveStatus(e.PID, e.LastExitStatus),
		}

		if sr, ok := plistMap[e.Label]; ok {
			job.PlistPath = sr.Path
			job.Program = sr.Data.Program
			job.ProgramArgs = sr.Data.ProgramArguments
			job.StandardOutPath = sr.Data.StandardOutPath
			job.StandardErrPath = sr.Data.StandardErrorPath
			job.RunAtLoad = sr.Data.RunAtLoad
			job.KeepAlive = sr.Data.KeepAlive
			job.Domain = s.detectDomain(sr.Path)

			// Fallback: if Program is empty, use ProgramArguments[0].
			if job.Program == "" && len(job.ProgramArgs) > 0 {
				job.Program = job.ProgramArgs[0]
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJob returns a single job by label. Returns an error if the label is not found.
func (s *Service) GetJob(label string) (*Job, error) {
	jobs, err := s.ListJobs()
	if err != nil {
		return nil, err
	}

	for i := range jobs {
		if jobs[i].Label == label {
			return &jobs[i], nil
		}
	}

	return nil, fmt.Errorf("job not found: %s", label)
}

// validateLabel checks that a label contains only safe characters for launchctl args.
func validateLabel(label string) error {
	if !labelRe.MatchString(label) {
		return fmt.Errorf("invalid label: %q", label)
	}
	return nil
}

// Reload unloads and reloads a job (bootout then bootstrap).
func (s *Service) Reload(label string) error {
	if err := validateLabel(label); err != nil {
		return err
	}

	job, err := s.GetJob(label)
	if err != nil {
		return fmt.Errorf("reload %s: %w", label, err)
	}
	if job.PlistPath == "" {
		return fmt.Errorf("reload %s: no plist path available", label)
	}

	// Step 1: bootout — ignore error (job may not be loaded).
	s.runExec("launchctl", "bootout", s.domain+"/"+label)

	// Step 2: bootstrap.
	_, err = s.runExec("launchctl", "bootstrap", s.domain, job.PlistPath)
	if err != nil {
		return fmt.Errorf("reload %s: %w", label, err)
	}
	return nil
}

// Start kickstarts a job.
func (s *Service) Start(label string) error {
	if err := validateLabel(label); err != nil {
		return err
	}
	_, err := s.runExec("launchctl", "kickstart", s.domain+"/"+label)
	if err != nil {
		return fmt.Errorf("start %s: %w", label, err)
	}
	return nil
}

// Stop sends SIGTERM to a running job.
func (s *Service) Stop(label string) error {
	if err := validateLabel(label); err != nil {
		return err
	}
	_, err := s.runExec("launchctl", "kill", "SIGTERM", s.domain+"/"+label)
	if err != nil {
		return fmt.Errorf("stop %s: %w", label, err)
	}
	return nil
}

// detectDomain determines whether a plist path belongs to the user or global domain.
func (s *Service) detectDomain(plistPath string) string {
	userDir := filepath.Join(s.homeDir, "Library", "LaunchAgents")
	if strings.HasPrefix(plistPath, userDir) {
		return "user"
	}
	if strings.HasPrefix(plistPath, "/Library/LaunchAgents") {
		return "global"
	}
	return "user"
}
