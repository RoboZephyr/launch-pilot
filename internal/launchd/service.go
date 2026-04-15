package launchd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/A404coder/launch-pilot/internal/plist"
)

// tailFile reads the last n lines from a file using a circular buffer,
// avoiding loading the entire file into memory for large log files.
func tailFile(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	ring := make([]string, n)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	idx := 0
	count := 0
	for scanner.Scan() {
		ring[idx] = scanner.Text()
		idx = (idx + 1) % n
		count++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if count == 0 {
		return "", nil
	}
	if count <= n {
		return strings.Join(ring[:count], "\n"), nil
	}
	result := make([]string, 0, n)
	result = append(result, ring[idx:]...)
	result = append(result, ring[:idx]...)
	return strings.Join(result, "\n"), nil
}

// ReadLogs reads the stdout and stderr log files for a job, returning the last
// n lines of each. Returns a message when no log paths are configured in the
// plist. Returns null stdout/stderr when log files don't exist on disk.
func (s *Service) ReadLogs(label string, lines int) (*LogOutput, error) {
	job, err := s.GetJob(label)
	if err != nil {
		return nil, err
	}

	out := &LogOutput{
		Label:      job.Label,
		StdoutPath: job.StandardOutPath,
		StderrPath: job.StandardErrPath,
	}

	// No log paths configured at all.
	if job.StandardOutPath == "" && job.StandardErrPath == "" {
		out.Message = "No log paths configured in plist"
		return out, nil
	}

	// Read stdout if configured.
	if job.StandardOutPath != "" {
		content, err := tailFile(job.StandardOutPath, lines)
		if err == nil {
			out.Stdout = &content
			out.StdoutAvailable = true
		}
	}

	// Read stderr if configured.
	if job.StandardErrPath != "" {
		content, err := tailFile(job.StandardErrPath, lines)
		if err == nil {
			out.Stderr = &content
			out.StderrAvailable = true
		}
	}

	return out, nil
}

// Service provides high-level operations on launchd user-domain jobs.
// It merges data from launchctl list (Layer 1) with plist files (Layer 2).
type Service struct {
	uid     int
	domain  string // "gui/<UID>"
	homeDir string
	cache   *plist.Cache

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
		cache:   plist.NewCache(),
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
	return plist.ScanAll(plist.ScanDirs(), s.cache)
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

// GetJob returns a single job by label. Returns an error wrapping ErrNotFound
// if the label is not present in the current job list.
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

	return nil, fmt.Errorf("%w: %s", ErrNotFound, label)
}

// Reload unloads and reloads a job (bootout then bootstrap).
func (s *Service) Reload(label string) error {
	if err := ValidateLabel(label); err != nil {
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
	if err := ValidateLabel(label); err != nil {
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
	if err := ValidateLabel(label); err != nil {
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
