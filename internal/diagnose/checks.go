package diagnose

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/A404coder/launch-pilot/internal/launchd"
)

// checkExitCode maps the job's last exit status to a human-readable explanation.
func checkExitCode(job *launchd.Job) CheckResult {
	meaning, suggestion := ExplainExitCode(job.LastExitStatus)
	sev := SeverityOK
	if job.LastExitStatus != 0 {
		sev = SeverityError
	}
	return CheckResult{
		ID:         "exit-code",
		Name:       "Exit Code Analysis",
		Severity:   sev,
		Message:    fmt.Sprintf("Exit code %d: %s", job.LastExitStatus, meaning),
		Suggestion: suggestion,
	}
}

// resolveProgram returns the effective executable path for a job.
// It prefers Program, then falls back to ProgramArgs[0].
func resolveProgram(job *launchd.Job) string {
	if job.Program != "" {
		return job.Program
	}
	if len(job.ProgramArgs) > 0 {
		return job.ProgramArgs[0]
	}
	return ""
}

// checkProgramExists verifies that the job's executable path exists on disk.
func checkProgramExists(job *launchd.Job) CheckResult {
	prog := resolveProgram(job)
	if prog == "" {
		return CheckResult{
			ID:         "program-exists",
			Name:       "Program Path",
			Severity:   SeverityWarning,
			Message:    "No Program or ProgramArguments configured",
			Suggestion: "Add Program or ProgramArguments to the plist",
		}
	}

	if _, err := os.Stat(prog); err != nil {
		return CheckResult{
			ID:         "program-exists",
			Name:       "Program Path",
			Severity:   SeverityError,
			Message:    fmt.Sprintf("%s does not exist", prog),
			Suggestion: "Verify the path in Program or ProgramArguments",
		}
	}

	return CheckResult{
		ID:       "program-exists",
		Name:     "Program Path",
		Severity: SeverityOK,
		Message:  fmt.Sprintf("%s exists", prog),
	}
}

// checkProgramExecutable verifies that the job's executable has the execute bit set.
func checkProgramExecutable(job *launchd.Job) CheckResult {
	prog := resolveProgram(job)
	if prog == "" {
		return CheckResult{
			ID:       "program-executable",
			Name:     "Program Executable",
			Severity: SeverityWarning,
			Message:  "No program path to check",
		}
	}

	info, err := os.Stat(prog)
	if err != nil {
		return CheckResult{
			ID:       "program-executable",
			Name:     "Program Executable",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Cannot stat %s: %v", prog, err),
		}
	}

	if info.Mode().Perm()&0111 == 0 {
		return CheckResult{
			ID:         "program-executable",
			Name:       "Program Executable",
			Severity:   SeverityError,
			Message:    fmt.Sprintf("%s lacks execute permission", prog),
			Suggestion: fmt.Sprintf("chmod +x %s", prog),
		}
	}

	return CheckResult{
		ID:       "program-executable",
		Name:     "Program Executable",
		Severity: SeverityOK,
		Message:  fmt.Sprintf("%s is executable", prog),
	}
}

// checkPlistOwner verifies that the plist file is owned by the current user.
func checkPlistOwner(job *launchd.Job) CheckResult {
	if job.PlistPath == "" {
		return CheckResult{
			ID:       "plist-owner",
			Name:     "Plist Owner",
			Severity: SeverityWarning,
			Message:  "No plist path known",
		}
	}

	info, err := os.Stat(job.PlistPath)
	if err != nil {
		return CheckResult{
			ID:       "plist-owner",
			Name:     "Plist Owner",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Cannot stat %s: %v", job.PlistPath, err),
		}
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return CheckResult{
			ID:       "plist-owner",
			Name:     "Plist Owner",
			Severity: SeverityWarning,
			Message:  "Cannot determine file owner on this platform",
		}
	}

	uid := os.Getuid()
	if int(stat.Uid) != uid {
		return CheckResult{
			ID:         "plist-owner",
			Name:       "Plist Owner",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("%s owned by UID %d, not current user UID %d", job.PlistPath, stat.Uid, uid),
			Suggestion: fmt.Sprintf("chown %d %s", uid, job.PlistPath),
		}
	}

	return CheckResult{
		ID:       "plist-owner",
		Name:     "Plist Owner",
		Severity: SeverityOK,
		Message:  fmt.Sprintf("%s owned by current user", job.PlistPath),
	}
}

// checkPlistPerms verifies that the plist file has no group/world write bits set.
func checkPlistPerms(job *launchd.Job) CheckResult {
	if job.PlistPath == "" {
		return CheckResult{
			ID:       "plist-perms",
			Name:     "Plist Permissions",
			Severity: SeverityWarning,
			Message:  "No plist path known",
		}
	}

	info, err := os.Stat(job.PlistPath)
	if err != nil {
		return CheckResult{
			ID:       "plist-perms",
			Name:     "Plist Permissions",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Cannot stat %s: %v", job.PlistPath, err),
		}
	}

	perm := info.Mode().Perm()
	if perm&0022 != 0 {
		return CheckResult{
			ID:         "plist-perms",
			Name:       "Plist Permissions",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("%s has permissions %o (group/world writable)", job.PlistPath, perm),
			Suggestion: fmt.Sprintf("chmod 644 %s", job.PlistPath),
		}
	}

	return CheckResult{
		ID:       "plist-perms",
		Name:     "Plist Permissions",
		Severity: SeverityOK,
		Message:  fmt.Sprintf("%s permissions %o OK", job.PlistPath, perm),
	}
}

// checkLogPathExists verifies that the parent directories for log files exist.
func checkLogPathExists(job *launchd.Job) CheckResult {
	if job.StandardOutPath == "" && job.StandardErrPath == "" {
		return CheckResult{
			ID:       "log-path-exists",
			Name:     "Log Path",
			Severity: SeverityOK,
			Message:  "No log paths configured",
		}
	}

	var missing []string
	for _, p := range []string{job.StandardOutPath, job.StandardErrPath} {
		if p == "" {
			continue
		}
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); err != nil {
			missing = append(missing, dir)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:         "log-path-exists",
			Name:       "Log Path",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("Log parent directory missing: %v", missing),
			Suggestion: fmt.Sprintf("mkdir -p %s", missing[0]),
		}
	}

	return CheckResult{
		ID:       "log-path-exists",
		Name:     "Log Path",
		Severity: SeverityOK,
		Message:  "Log parent directories exist",
	}
}
