package launchd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

const execTimeout = 10 * time.Second

// ExecResult holds the stdout, stderr, and any error from a command execution.
type ExecResult struct {
	Stdout string
	Stderr string
}

// runCmd executes a command with a 10-second timeout and returns its output.
func runCmd(name string, args ...string) (*ExecResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %s: %s %v", execTimeout, name, args)
	}

	if err != nil {
		return result, fmt.Errorf("command failed: %s %v: %w (stderr: %s)", name, args, err, result.Stderr)
	}

	return result, nil
}
