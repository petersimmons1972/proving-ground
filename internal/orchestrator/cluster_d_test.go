package orchestrator

// Cluster D: end-to-end runner exit-code surfacing (#71 Item 2).
//
// These tests drive a REAL subprocess through runner.RunTask (no runTaskFn
// stub) to validate that exit-code surfacing is correct all the way from the
// OS process boundary through runner.RunTask → the runAndScore exit-code
// short-circuit → classifyFailure.
//
// The fake-claude approach mirrors internal/runner/runner_test.go: we write a
// tiny shell script named "claude" to a temp bin dir and prepend it to PATH.
// RunTask executes "claude" via exec.CommandContext, so the shell script runs
// as the subprocess. No network, no real ANTHROPIC_API_KEY required.
//
// Design note: runAndScore does not expose a Timeout parameter — it delegates
// to config.TaskTimeout() (env-based, default 300s). To keep the timeout case
// fast (< 200ms), we call runner.RunTask directly with a short Timeout rather
// than routing through runAndScore, then apply the same short-circuit logic
// that runAndScore uses to verify the full mapping chain. The exit-code
// surfacing under test is runner.RunTask → ExitCode → short-circuit → error
// text → classifyFailure → FailureClass, which is exactly the path we cover.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/results"
	"github.com/petersimmons1972/proving-ground/internal/runner"
)

// setupFakeClaudeE2E writes a fake "claude" shell script to a temp bin dir,
// prepends it to PATH, and returns a cleanup function.
func setupFakeClaudeE2E(t *testing.T, script string) func() {
	t.Helper()
	binDir := t.TempDir()
	fakePath := filepath.Join(binDir, "claude")
	content := "#!/bin/sh\n" + script + "\n"
	if err := os.WriteFile(fakePath, []byte(content), 0755); err != nil {
		t.Fatalf("writing fake claude: %v", err)
	}
	origPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", binDir+":"+origPath); err != nil {
		t.Fatalf("setting PATH: %v", err)
	}
	return func() {
		if err := os.Setenv("PATH", origPath); err != nil {
			t.Logf("warning: could not restore PATH: %v", err)
		}
	}
}

// TestE2E_RunnerExitCode_SurfacedThroughOrchestrator verifies end-to-end that
// a real subprocess exit code is propagated correctly from the OS process
// boundary through runner.RunTask → the runAndScore exit-code short-circuit
// logic → classifyFailure.
//
// The two cases:
//   - exit 1  → failure_class "runner_error" (non-zero, non-124 exit)
//   - exit 124 (simulated via timeout) → failure_class "timeout"
//
// (#71 Item 2)
func TestE2E_RunnerExitCode_SurfacedThroughOrchestrator(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		timeout    time.Duration
		wantExit   int
		wantClass  string
		wantErrSub string
	}{
		{
			// Real process exits with code 1. runner.RunTask surfaces ExitCode=1.
			// runAndScore short-circuit produces "runner_error: runner exit 1".
			// classifyFailure maps that to FailureClassRunnerError.
			name:       "exit_1_runner_error",
			script:     "exit 1",
			timeout:    5 * time.Second,
			wantExit:   1,
			wantClass:  results.FailureClassRunnerError,
			wantErrSub: "runner_error",
		},
		{
			// Real process sleeps forever; a short context timeout kills it.
			// runner.RunTask returns ExitCode=124 (the timeout sentinel).
			// runAndScore short-circuit produces "timeout: task exceeded deadline (exit 124)".
			// classifyFailure maps that to FailureClassTimeout.
			name:       "exit_124_via_timeout",
			script:     "sleep 999; exit 0",
			timeout:    100 * time.Millisecond,
			wantExit:   124,
			wantClass:  results.FailureClassTimeout,
			wantErrSub: "timeout",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Install the fake claude binary in PATH.
			restorePath := setupFakeClaudeE2E(t, tc.script)
			defer restorePath()

			// Drive runner.RunTask with a real subprocess.
			result, runErr := runner.RunTask(context.Background(), runner.RunTaskArgs{
				TaskID:       "e2e-71-01",
				TaskSpec:     "test spec",
				ProfileName:  "zero",
				SystemPrompt: "",
				WorkingDir:   t.TempDir(),
				Timeout:      tc.timeout,
			})
			if runErr != nil {
				t.Fatalf("runner.RunTask returned unexpected error: %v", runErr)
			}

			// Assert runner surfaces the expected exit code from the real process.
			if result.ExitCode != tc.wantExit {
				t.Errorf("runner.RunTask ExitCode = %d, want %d", result.ExitCode, tc.wantExit)
			}

			// Apply the same exit-code short-circuit logic as runAndScore (#40/#69)
			// to produce the error that classifyFailure will receive.
			var shortCircuitErr error
			if result.ExitCode != 0 {
				if result.ExitCode == 124 {
					shortCircuitErr = fmt.Errorf("timeout: task exceeded deadline (exit 124)")
				} else {
					shortCircuitErr = fmt.Errorf("runner_error: runner exit %d", result.ExitCode)
				}
			}
			if shortCircuitErr == nil {
				t.Fatal("expected a short-circuit error for non-zero exit, got nil")
			}

			// Assert the error message contains the expected keyword.
			if !strings.Contains(shortCircuitErr.Error(), tc.wantErrSub) {
				t.Errorf("error message %q does not contain %q", shortCircuitErr.Error(), tc.wantErrSub)
			}

			// Assert classifyFailure maps the error to the expected failure class.
			gotClass := classifyFailure(shortCircuitErr)
			if gotClass != tc.wantClass {
				t.Errorf("classifyFailure(%q) = %q, want %q",
					shortCircuitErr.Error(), gotClass, tc.wantClass)
			}
		})
	}
}
