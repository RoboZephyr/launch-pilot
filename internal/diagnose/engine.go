package diagnose

import "github.com/A404coder/launchboard/internal/launchd"

// Engine runs read-only diagnostic checks against launchd jobs.
type Engine struct{}

// Diagnose runs all diagnostic checks against the given job and returns a report.
func (e *Engine) Diagnose(job *launchd.Job) *DiagnoseReport {
	return &DiagnoseReport{
		Label: job.Label,
		Checks: []CheckResult{
			checkExitCode(job),
			checkProgramExists(job),
			checkProgramExecutable(job),
			checkPlistOwner(job),
			checkPlistPerms(job),
			checkLogPathExists(job),
		},
	}
}
