package scoring_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

// setupFakeJudgeClaude writes a fake claude binary that runs the given shell script,
// prepends its bin dir to PATH, and returns the path to a temp rubric file plus a
// cleanup function that restores PATH.
func setupFakeJudgeClaude(t *testing.T, script string) (rubricPath string, cleanup func()) {
	t.Helper()

	rubricDir := t.TempDir()
	rubricPath = filepath.Join(rubricDir, "judge-rubric.md")
	if err := os.WriteFile(rubricPath, []byte("# Rubric\nScore things."), 0644); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	fakePath := filepath.Join(binDir, "claude")
	if err := os.WriteFile(fakePath, []byte("#!/bin/sh\n"+script+"\n"), 0755); err != nil {
		t.Fatal(err)
	}

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	return rubricPath, func() { os.Setenv("PATH", origPath) }
}

func TestJudgeReturnsJudgeScores(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: Agent reframed the problem elegantly.\n'; exit 0`)
	defer cleanup()

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if err != nil {
		t.Fatal(err)
	}
	if scores.RequirementInterpretation != 8 {
		t.Errorf("RequirementInterpretation = %f, want 8", scores.RequirementInterpretation)
	}
	if scores.DecisionCommunication != 7 {
		t.Errorf("DecisionCommunication = %f, want 7", scores.DecisionCommunication)
	}
	if scores.RecoveryQuality != 5 {
		t.Errorf("RecoveryQuality = %f, want 5", scores.RecoveryQuality)
	}
	if scores.UnconventionalThinking != 9 {
		t.Errorf("UnconventionalThinking = %f, want 9", scores.UnconventionalThinking)
	}
}

func TestJudgeCapturesRationale(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: Agent reframed the problem elegantly.\n'; exit 0`)
	defer cleanup()

	scores, _ := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if scores.Rationale == "" {
		t.Error("rationale is empty")
	}
	if !strings.Contains(scores.Rationale, "reframed") {
		t.Errorf("rationale = %q, expected it to contain 'reframed'", scores.Rationale)
	}
}

func TestJudgeMedianOfThree(t *testing.T) {
	// Three responses with different scores so we can verify median selection.
	responses := []string{
		"REQUIREMENT_INTERPRETATION: 6\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 5\nRECOVERY_QUALITY: 4\nUNCONVENTIONAL_THINKING: 8\nRATIONALE: ok",
		"REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 6\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok",
		"REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 8\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: ok",
	}

	rubricDir := t.TempDir()
	rubricPath := filepath.Join(rubricDir, "judge-rubric.md")
	os.WriteFile(rubricPath, []byte("# Rubric"), 0644)

	binDir := t.TempDir()
	for i, r := range responses {
		os.WriteFile(filepath.Join(binDir, fmt.Sprintf("resp%d.txt", i)), []byte(r), 0644)
	}

	counterFile := filepath.Join(binDir, "counter")
	os.WriteFile(counterFile, []byte("0"), 0644)

	script := fmt.Sprintf(`#!/bin/sh
N=$(cat %s)
cat %s/resp${N}.txt
echo $((N+1)) > %s
exit 0`, counterFile, binDir, counterFile)
	os.WriteFile(filepath.Join(binDir, "claude"), []byte(script), 0755)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 3)
	if err != nil {
		t.Fatal(err)
	}
	// Median of [6, 8, 7] = 7
	if scores.RequirementInterpretation != 7 {
		t.Errorf("RequirementInterpretation median = %f, want 7", scores.RequirementInterpretation)
	}
}

func TestJudgeAllRunsFailReturnsZeros(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t, `exit 1`)
	defer cleanup()

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 3)
	if err != nil {
		t.Fatal(err)
	}
	if scores.RequirementInterpretation != 0 {
		t.Errorf("RequirementInterpretation = %f, want 0", scores.RequirementInterpretation)
	}
	if scores.Rationale != "ALL_JUDGE_RUNS_FAILED" {
		t.Errorf("Rationale = %q, want 'ALL_JUDGE_RUNS_FAILED'", scores.Rationale)
	}
}

// TestScoreWithJudge_AllRunsFail_ReturnsSentinel verifies B-2:
// when ALL judge invocations fail, Failed must be true so callers
// know the zero scores are not real scores — not silent contamination.
func TestScoreWithJudge_AllRunsFail_ReturnsSentinel(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t, `exit 1`)
	defer cleanup()

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !scores.Failed {
		t.Error("Failed = false, want true when all judge runs fail")
	}
	if scores.FailureReason != "ALL_JUDGE_RUNS_FAILED" {
		t.Errorf("FailureReason = %q, want %q", scores.FailureReason, "ALL_JUDGE_RUNS_FAILED")
	}
	// Scores must be zero AND Failed must be true — not ambiguously zero
	if scores.RequirementInterpretation != 0 {
		t.Errorf("RequirementInterpretation = %f, want 0 for failed sentinel", scores.RequirementInterpretation)
	}
}

// TestScoreWithJudge_SuccessfulRun_NotSentinel verifies that a successful
// run does NOT set Failed=true, so callers can distinguish real zeros from
// failed-run zeros.
func TestScoreWithJudge_SuccessfulRun_NotSentinel(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 0\nDECISION_COMMUNICATION: 0\nRECOVERY_QUALITY: 0\nUNCONVENTIONAL_THINKING: 0\nRATIONALE: legitimately bad\n'; exit 0`)
	defer cleanup()

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if err != nil {
		t.Fatal(err)
	}
	if scores.Failed {
		t.Error("Failed = true, want false when judge runs succeed (even with zero scores)")
	}
}

func TestJudgeMissingDimensionDefaultsFive(t *testing.T) {
	// Response omits UNCONVENTIONAL_THINKING — should default to 5.
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 6\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 5\nRATIONALE: incomplete response\n'; exit 0`)
	defer cleanup()

	scores, _ := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if scores.UnconventionalThinking != 5 {
		t.Errorf("UnconventionalThinking = %f, want 5 (default for missing)", scores.UnconventionalThinking)
	}
}

func TestJudgeSkipsFailedRun(t *testing.T) {
	// First run exits 1 (failure), second run succeeds — scores should come from run 2.
	rubricDir := t.TempDir()
	rubricPath := filepath.Join(rubricDir, "judge-rubric.md")
	os.WriteFile(rubricPath, []byte("# Rubric"), 0644)

	binDir := t.TempDir()
	counterFile := filepath.Join(binDir, "counter")
	os.WriteFile(counterFile, []byte("0"), 0644)

	script := fmt.Sprintf(`#!/bin/sh
N=$(cat %s)
echo $((N+1)) > %s
if [ "$N" = "0" ]; then
  exit 1
fi
printf 'REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: ok\n'
exit 0`, counterFile, counterFile)
	os.WriteFile(filepath.Join(binDir, "claude"), []byte(script), 0755)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	scores, _ := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 2)
	if scores.RequirementInterpretation != 8 {
		t.Errorf("RequirementInterpretation = %f, want 8 (from successful run)", scores.RequirementInterpretation)
	}
}

// TestJudgeRationaleFromMedianRun verifies that the rationale comes from the
// median-scoring run (by REQUIREMENT_INTERPRETATION), not always all[0].
// Three runs: scores 3, 7, 5 — median is 5, so rationale must be "rationale-mid".
func TestJudgeRationaleFromMedianRun(t *testing.T) {
	// Run 0: RI=3, rationale "rationale-low"
	// Run 1: RI=7, rationale "rationale-high"
	// Run 2: RI=5, rationale "rationale-mid"
	// After sorting by RI ascending: [3, 5, 7] → index 1 (len=3, 3/2=1) → RI=5 → "rationale-mid"
	responses := []string{
		"REQUIREMENT_INTERPRETATION: 3\nDECISION_COMMUNICATION: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: rationale-low\n",
		"REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: rationale-high\n",
		"REQUIREMENT_INTERPRETATION: 5\nDECISION_COMMUNICATION: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: rationale-mid\n",
	}

	rubricDir := t.TempDir()
	rubricPath := filepath.Join(rubricDir, "judge-rubric.md")
	os.WriteFile(rubricPath, []byte("# Rubric"), 0644)

	binDir := t.TempDir()
	for i, r := range responses {
		os.WriteFile(filepath.Join(binDir, fmt.Sprintf("resp%d.txt", i)), []byte(r), 0644)
	}

	counterFile := filepath.Join(binDir, "counter")
	os.WriteFile(counterFile, []byte("0"), 0644)

	script := fmt.Sprintf(`#!/bin/sh
N=$(cat %s)
cat %s/resp${N}.txt
echo $((N+1)) > %s
exit 0`, counterFile, binDir, counterFile)
	os.WriteFile(filepath.Join(binDir, "claude"), []byte(script), 0755)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	scores, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 3)
	if err != nil {
		t.Fatal(err)
	}
	if scores.Rationale != "rationale-mid" {
		t.Errorf("Rationale = %q, want %q (median run's rationale)", scores.Rationale, "rationale-mid")
	}
}

// TestJudge_Args_DoNotContainDangerousPermissions verifies the judge subprocess
// does not run with --dangerously-skip-permissions.
// The judge is a read-only evaluator — it reads transcripts and outputs scores.
// Feeding it untrusted agent transcripts with shell-execute permissions in place
// creates a prompt-injection-into-executable-process risk (B-4, #21).
func TestJudge_Args_DoNotContainDangerousPermissions(t *testing.T) {
	args := scoring.BuildJudgeArgs("rubric text")
	for _, a := range args {
		if a == "--dangerously-skip-permissions" {
			t.Errorf("judge args contain --dangerously-skip-permissions; this flag must not be present on the read-only judge subprocess")
		}
	}
}

func TestJudgeClampsHighScore(t *testing.T) {
	// Score of 85 should be clamped to 10.
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 85\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: wild\n'; exit 0`)
	defer cleanup()

	scores, _ := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if scores.RequirementInterpretation != 10 {
		t.Errorf("RequirementInterpretation = %f, want 10 (clamped from 85)", scores.RequirementInterpretation)
	}
}

// TestScoreWithJudge_PartialFailures_LogsWarning verifies B-R3-2: when some
// (but not all) judge runs fail, a WARNING log is emitted indicating degraded
// quality so callers can detect that the sample size is smaller than requested.
func TestScoreWithJudge_PartialFailures_LogsWarning(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// First 2 runs fail (exit 1), last 3 succeed.
	rubricDir := t.TempDir()
	rubricPath := filepath.Join(rubricDir, "judge-rubric.md")
	os.WriteFile(rubricPath, []byte("# Rubric"), 0644)

	binDir := t.TempDir()
	counterFile := filepath.Join(binDir, "counter")
	os.WriteFile(counterFile, []byte("0"), 0644)

	script := fmt.Sprintf(`#!/bin/sh
N=$(cat %s)
echo $((N+1)) > %s
if [ "$N" -lt "2" ]; then
  exit 1
fi
printf 'REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 7\nRECOVERY_QUALITY: 7\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok\n'
exit 0`, counterFile, counterFile)
	os.WriteFile(filepath.Join(binDir, "claude"), []byte(script), 0755)

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	_, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubricPath, 5)
	if err != nil {
		t.Fatal(err)
	}

	logOut := buf.String()
	if !strings.Contains(logOut, "degraded") && !strings.Contains(logOut, "WARNING") {
		t.Errorf("expected log to contain 'degraded' or 'WARNING' for partial failures, got: %q", logOut)
	}
}

// TestScoreWithJudge_FailedRun_LogsStderr verifies B-R3-3: stderr from a
// failed judge invocation is logged, mirroring runner.go behavior.
func TestScoreWithJudge_FailedRun_LogsStderr(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// The fake judge writes to stderr and exits non-zero.
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'judge error: model overloaded\n' >&2; exit 1`)
	defer cleanup()

	_, err := scoring.ScoreWithJudge(context.Background(), "transcript", "spec", rubric, 1)
	if err != nil {
		t.Fatal(err)
	}

	logOut := buf.String()
	if !strings.Contains(logOut, "judge error: model overloaded") {
		t.Errorf("expected log to contain stderr text 'judge error: model overloaded', got: %q", logOut)
	}
}
