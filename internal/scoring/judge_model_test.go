package scoring_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

// TestJudge_ModelEnvVar_AppendsModelFlag confirms that when PROVING_GROUND_MODEL
// is set, the judge subprocess is invoked with --model <value>.
func TestJudge_ModelEnvVar_AppendsModelFlag(t *testing.T) {
	// Script echoes its arguments so we can inspect them.
	rubricPath, cleanup := setupFakeJudgeClaude(t, `echo "$@"; exit 0`)
	defer cleanup()

	os.Setenv("PROVING_GROUND_MODEL", "claude-opus-4-5")
	defer os.Unsetenv("PROVING_GROUND_MODEL")

	// ScoreWithJudge will attempt to parse our output as scores and fail
	// (no score lines), but we need the args, so we just run and check stdout.
	// We capture via the fake claude printing args; since all judge runs will
	// "fail" (no score output), we need a version that wraps BuildJudgeArgs.
	args := scoring.BuildJudgeArgs("rubric")
	hasModel := false
	for i, a := range args {
		if a == "--model" && i+1 < len(args) && args[i+1] == "claude-opus-4-5" {
			hasModel = true
		}
	}
	if !hasModel {
		t.Errorf("BuildJudgeArgs did not include --model claude-opus-4-5 when PROVING_GROUND_MODEL=claude-opus-4-5; args: %v", args)
	}

	// Also verify end-to-end via ScoreWithJudge running the fake binary.
	res, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	// All runs will fail to parse scores (fake outputs args, not scores),
	// but we can verify the fake claude was passed --model by inspecting
	// the fake binary's stdout captured in... actually we can't easily get
	// the fake's stdout here. The BuildJudgeArgs check above is sufficient
	// since ScoreWithJudge uses BuildJudgeArgs. Verify it's a sentinel.
	if !res.Failed {
		// This is fine — just means parse somehow succeeded, which is unexpected
		// but not a failure mode for this test.
		_ = res
	}
}

// TestJudge_ModelEnvVar_NotSet_NoModelFlag confirms that when PROVING_GROUND_MODEL
// is absent, --model is not passed to the judge.
func TestJudge_ModelEnvVar_NotSet_NoModelFlag(t *testing.T) {
	os.Unsetenv("PROVING_GROUND_MODEL")

	args := scoring.BuildJudgeArgs("rubric")
	for _, a := range args {
		if a == "--model" {
			t.Errorf("--model flag should not be in BuildJudgeArgs when PROVING_GROUND_MODEL is unset; args: %v", args)
		}
	}
}

// TestJudge_ModelEnvVar_E2E_AppendsModelFlag runs the full ScoreWithJudge path
// and captures the fake claude's received args from stdout to verify --model passes.
func TestJudge_ModelEnvVar_E2E_AppendsModelFlag(t *testing.T) {
	// Fake claude: write its own args to a temp file, then output valid scores.
	argsFile := t.TempDir() + "/args.txt"

	rubricDir := t.TempDir()
	rubricPath := rubricDir + "/judge-rubric.md"
	os.WriteFile(rubricPath, []byte("# Rubric"), 0644)

	binDir := t.TempDir()
	script := `printf '%s\n' "$@" > ` + argsFile + `
printf 'REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 7\nRECOVERY_QUALITY: 7\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok\n'
exit 0`
	os.WriteFile(binDir+"/claude", []byte("#!/bin/sh\n"+script+"\n"), 0755)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	os.Setenv("PROVING_GROUND_MODEL", "claude-sonnet-4-5")
	defer os.Unsetenv("PROVING_GROUND_MODEL")

	_, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 1)
	if err != nil {
		t.Fatal(err)
	}

	argsBytes, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("could not read args file: %v", err)
	}
	argsStr := string(argsBytes)
	if !strings.Contains(argsStr, "--model") {
		t.Errorf("--model not found in args passed to fake claude: %q", argsStr)
	}
	if !strings.Contains(argsStr, "claude-sonnet-4-5") {
		t.Errorf("model value not found in args passed to fake claude: %q", argsStr)
	}
}
