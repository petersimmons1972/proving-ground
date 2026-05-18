package runner_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/runner"
)

// TestRunner_ModelEnvVar_AppendsModelFlag confirms that when PROVING_GROUND_MODEL
// is set, the args passed to claude include --model <value>.
func TestRunner_ModelEnvVar_AppendsModelFlag(t *testing.T) {
	// Script echoes its arguments so we can inspect them.
	cleanup := setupFakeClaude(t, `echo "$@"; exit 0`)
	defer cleanup()

	os.Setenv("PROVING_GROUND_MODEL", "claude-opus-4-5")
	defer os.Unsetenv("PROVING_GROUND_MODEL")

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero", WorkingDir: tmp,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Stdout, "--model") {
		t.Errorf("expected --model flag in args when PROVING_GROUND_MODEL is set; got stdout: %q", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "claude-opus-4-5") {
		t.Errorf("expected model value in args; got stdout: %q", res.Stdout)
	}
}

// TestRunner_ModelEnvVar_NotSet_NoModelFlag confirms that when PROVING_GROUND_MODEL
// is absent, no --model flag is passed (no breaking change to current behavior).
func TestRunner_ModelEnvVar_NotSet_NoModelFlag(t *testing.T) {
	cleanup := setupFakeClaude(t, `echo "$@"; exit 0`)
	defer cleanup()

	os.Unsetenv("PROVING_GROUND_MODEL")

	tmp := t.TempDir()
	res, err := runner.RunTask(context.Background(), runner.RunTaskArgs{
		TaskID: "t1-1", TaskSpec: "spec", ProfileName: "zero", WorkingDir: tmp,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Stdout, "--model") {
		t.Errorf("--model flag should not be present when PROVING_GROUND_MODEL is not set; got stdout: %q", res.Stdout)
	}
}
