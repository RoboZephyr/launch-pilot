package diagnose

import "fmt"

type exitInfo struct {
	meaning    string
	suggestion string
}

var exitCodeMessages = map[int]exitInfo{
	0:   {"Normal exit", ""},
	1:   {"General error", "Check program logs for details"},
	2:   {"Command usage error", "Verify ProgramArguments in plist"},
	5:   {"I/O error", "Check plist encoding (UTF-8); ensure Full Disk Access for Terminal"},
	13:  {"Permission denied", "Check executable and working directory permissions"},
	78:  {"Configuration error (EX_CONFIG)", "Validate plist: plutil -lint <path>"},
	126: {"Command not executable", "Add execute permission: chmod +x <path>"},
	127: {"Command not found", "Verify Program/ProgramArguments path exists"},
}

var signalNames = map[int]string{
	1:  "SIGHUP",
	2:  "SIGINT",
	9:  "SIGKILL",
	15: "SIGTERM",
}

// ExplainExitCode returns a human-readable meaning and fix suggestion for a
// launchd job exit code. Negative values are interpreted as signal terminations.
func ExplainExitCode(code int) (meaning, suggestion string) {
	if code < 0 {
		sigNum := -code
		if name, ok := signalNames[sigNum]; ok {
			return fmt.Sprintf("Killed by signal %d (%s)", sigNum, name), ""
		}
		return fmt.Sprintf("Killed by signal %d", sigNum), ""
	}

	if info, ok := exitCodeMessages[code]; ok {
		return info.meaning, info.suggestion
	}

	return fmt.Sprintf("Unknown exit code %d", code), ""
}
