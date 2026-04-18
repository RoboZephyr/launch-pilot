package plist

import (
	"os"
	"time"
)

// LastRunAt returns the newer of the stdout/stderr file mtimes as a heuristic
// for when a launchd job last ran. Empty paths are skipped (statFn is not
// called). Stat failures are treated as "no info". Returns zero time if no
// usable mtime is obtained.
//
// statFn is injected for testability; production callers pass os.Stat.
func LastRunAt(stdoutPath, stderrPath string, statFn func(string) (os.FileInfo, error)) time.Time {
	var best time.Time
	consider := func(path string) {
		if path == "" {
			return
		}
		info, err := statFn(path)
		if err != nil || info == nil {
			return
		}
		mt := info.ModTime()
		if mt.After(best) {
			best = mt
		}
	}
	consider(stdoutPath)
	consider(stderrPath)
	return best
}
