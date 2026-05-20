package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/runner"
)

// installFakeClaude writes /bin/sh script as 'claude' on PATH.
func installFakeClaude(t *testing.T, script string) func() {
	t.Helper()
	binDir := t.TempDir()
	fakePath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(fakePath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	orig := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+orig)
	return func() { os.Setenv("PATH", orig) }
}

// Issue #58: a verbose agent emitting tens of MB of stdout must not OOM the
// proving-ground process. The captured Stdout must be bounded; the run must
// complete; a truncation marker must be present when the cap was hit.
func TestRunTask_StdoutBoundedAgainstVerboseAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("verbose output test")
	}
	// Emit ~80MB of output (well above the 50MB read cap).
	cleanup := installFakeClaude(t,
		`yes 'x' | head -c 80000000; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	result, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID:      "verbose-task",
		TaskSpec:    "produce verbose output",
		ProfileName: "p",
		WorkingDir:  tmp,
		Timeout:     30 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	// Bound: result.Stdout must be ≤ 60MB (cap is 50MB + some headroom for truncation marker).
	if got := len(result.Stdout); got > 60*1024*1024 {
		t.Errorf("Stdout length = %d bytes, want bounded under 60MB", got)
	}
	// Truncation marker must be present when actual output exceeded cap.
	if !strings.Contains(result.Stdout, "TRUNCATED") {
		t.Errorf("Stdout missing TRUNCATED marker after 80MB output")
	}
}

// Issue #43: PROVING_GROUND_TASK_TIMEOUT env var must override the 300s default.
func TestRunTask_HonorsTaskTimeoutEnvVar(t *testing.T) {
	cleanup := installFakeClaude(t, `sleep 30; exit 0`)
	defer cleanup()

	orig := os.Getenv("PROVING_GROUND_TASK_TIMEOUT")
	os.Setenv("PROVING_GROUND_TASK_TIMEOUT", "1s")
	defer os.Setenv("PROVING_GROUND_TASK_TIMEOUT", orig)

	tmp := t.TempDir()
	start := time.Now()
	result, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID:      "slow-task",
		TaskSpec:    "sleep",
		ProfileName: "p",
		WorkingDir:  tmp,
		// Timeout left zero — must pick up env var.
	})
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if result.ExitCode != 124 {
		t.Errorf("ExitCode = %d, want 124 (timeout)", result.ExitCode)
	}
	if elapsed > 10*time.Second {
		t.Errorf("env var did not take effect: elapsed=%s, want ~1s", elapsed)
	}
}
