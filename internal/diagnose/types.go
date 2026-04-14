package diagnose

// Severity indicates how critical a diagnostic check result is.
type Severity string

const (
	SeverityOK      Severity = "ok"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// CheckResult holds the outcome of a single diagnostic check.
type CheckResult struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}

// DiagnoseReport holds all check results for a single job.
type DiagnoseReport struct {
	Label  string        `json:"label"`
	Checks []CheckResult `json:"checks"`
}
