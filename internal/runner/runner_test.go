package runner_test

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/runner"
)

// setupFakeClaude writes a fake claude script to a temp bin dir and prepends it to PATH.
// Returns a cleanup function.
func setupFakeClaude(t *testing.T, script string) func() {
	t.Helper()
	binDir := t.TempDir()
	fakePath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(fakePath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
		t.Fatalf("writing fake claude: %v", err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	return func() { os.Setenv("PATH", origPath) }
}

func TestRunTaskReturnsResult(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "1 passed"; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "implement parser",
		ProfileName: "zero", SystemPrompt: "", WorkingDir: tmp,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TaskID != "t1-1" {
		t.Errorf("TaskID = %q", res.TaskID)
	}
	if res.ProfileName != "zero" {
		t.Errorf("ProfileName = %q", res.ProfileName)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d", res.ExitCode)
	}
}

func TestRunTaskCapturesStdout(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "output text"; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero", WorkingDir: tmp,
	})
	if !strings.Contains(res.Stdout, "output text") {
		t.Errorf("stdout = %q", res.Stdout)
	}
}

func TestRunTaskCapturesNonZeroExit(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "error msg" >&2; exit 1`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero", WorkingDir: tmp,
	})
	if res.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", res.ExitCode)
	}
}

func TestRunTaskZeroProfileNoSystemFlag(t *testing.T) {
	// Script prints all args to stdout so we can inspect them.
	cleanup := setupFakeClaude(t, `echo "$@"; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero",
		SystemPrompt: "", WorkingDir: tmp,
	})
	if strings.Contains(res.Stdout, "--system-prompt") {
		t.Error("--system-prompt passed for empty system prompt")
	}
}

func TestRunTaskWithSystemPromptPassesFlag(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "$@"; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec text", ProfileName: "light",
		SystemPrompt: "You are an engineer.", WorkingDir: tmp,
	})
	if !strings.Contains(res.Stdout, "--system-prompt") {
		t.Error("--system-prompt not passed for non-empty system prompt")
	}
}

func TestSpecPipedViaStdin(t *testing.T) {
	// Script reads stdin and echoes it to stdout.
	cleanup := setupFakeClaude(t, `cat; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "my task spec", ProfileName: "zero", WorkingDir: tmp,
	})
	if !strings.Contains(res.Stdout, "my task spec") {
		t.Errorf("task spec not in stdin: %q", res.Stdout)
	}
	if !strings.HasPrefix(res.Stdout, "WORKING DIRECTORY:") {
		t.Error("WORKING DIRECTORY header not prepended")
	}
}

func TestSpecHeaderContainsAbsRunDir(t *testing.T) {
	cleanup := setupFakeClaude(t, `cat; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, _ := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "the spec", ProfileName: "zero", WorkingDir: tmp,
	})
	expectedDir := filepath.Join(tmp, "t1-1", "zero")
	abs, _ := filepath.Abs(expectedDir)
	if !strings.Contains(res.Stdout, abs) {
		t.Errorf("absolute run dir %q not in stdin header: %q", abs, res.Stdout[:200])
	}
	if !strings.Contains(res.Stdout, "Never write files outside this directory") {
		t.Error("missing 'Never write files outside this directory' directive")
	}
}

// TestRunner_Args_DoNotContainDangerousPermissions verifies the task runner
// subprocess does not run with --dangerously-skip-permissions.
// Without this flag, Claude Code enforces its permission system. Agents running
// in headless --print mode may be unable to write files or execute shell
// commands without permission approval — this is the intended security posture.
// If tasks produce empty stdout, verify the agent's tool-use permissions.
func TestRunner_Args_DoNotContainDangerousPermissions(t *testing.T) {
	args := runner.BuildRunnerArgs("system prompt")
	for _, a := range args {
		if a == "--dangerously-skip-permissions" {
			t.Errorf("runner args contain --dangerously-skip-permissions; this flag must not be present")
		}
	}
}

func TestRunTaskTimeoutReturnsResult(t *testing.T) {
	// Script sleeps indefinitely -- will be killed by timeout.
	cleanup := setupFakeClaude(t, `sleep 999; exit 0`)
	defer cleanup()

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero", WorkingDir: tmp,
		Timeout: 50 * time.Millisecond, // very short timeout for test speed
	})
	// Should not error -- timeout is handled gracefully.
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 124 {
		t.Errorf("ExitCode = %d, want 124", res.ExitCode)
	}
	if !strings.Contains(strings.ToLower(res.Stderr), "timeout") {
		t.Errorf("Stderr = %q, want timeout message", res.Stderr)
	}
}

// TestRunTask_LogsOnNonZeroExit verifies that a non-zero exit code produces
// a log line containing the task ID, profile name, and exit code.
func TestRunTask_LogsOnNonZeroExit(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "error output" >&2; exit 2`)
	defer cleanup()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "task-log-01", TaskSpec: "spec", ProfileName: "myprofile", WorkingDir: tmp,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 2 {
		t.Fatalf("ExitCode = %d, want 2", res.ExitCode)
	}

	logged := buf.String()
	if !strings.Contains(logged, "task-log-01") {
		t.Errorf("log output missing task ID %q: %s", "task-log-01", logged)
	}
	if !strings.Contains(logged, "myprofile") {
		t.Errorf("log output missing profile name %q: %s", "myprofile", logged)
	}
	if !strings.Contains(logged, strconv.Itoa(res.ExitCode)) {
		t.Errorf("log output missing exit code %d: %s", res.ExitCode, logged)
	}
}

// TestRunTask_LogsOnTimeout verifies that a timeout (ExitCode 124) produces
// a log line indicating the task timed out.
func TestRunTask_LogsOnTimeout(t *testing.T) {
	cleanup := setupFakeClaude(t, `sleep 999; exit 0`)
	defer cleanup()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "task-timeout-01", TaskSpec: "spec", ProfileName: "myprofile", WorkingDir: tmp,
		Timeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 124 {
		t.Fatalf("ExitCode = %d, want 124", res.ExitCode)
	}

	logged := buf.String()
	if !strings.Contains(strings.ToUpper(logged), "TIMEOUT") {
		t.Errorf("log output missing TIMEOUT keyword: %s", logged)
	}
	if !strings.Contains(logged, "task-timeout-01") {
		t.Errorf("log output missing task ID: %s", logged)
	}
}
