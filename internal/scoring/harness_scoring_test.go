package scoring_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

// pytestAvailable returns true if the pytest binary found in PATH supports
// --json-report (i.e. pytest-json-report is installed for that pytest).
func pytestAvailable() bool {
	if _, err := exec.LookPath("pytest"); err != nil {
		return false
	}
	// Verify the plugin is loaded by asking pytest to list plugins
	out, err := exec.Command("pytest", "--co", "--json-report", "--help").Output()
	_ = out
	// If --json-report is unknown, exit code is non-zero
	return err == nil
}

// TestScoreTests_HarnessRunsPytest_IgnoresAgentStdout verifies that ScoreTests runs
// pytest independently and returns a score reflecting actual results, not fabricated
// agent stdout ("100 passed 0 failed" would score 10.0 under the old scheme).
func TestScoreTests_HarnessRunsPytest_IgnoresAgentStdout(t *testing.T) {
	if !pytestAvailable() {
		t.Skip("pytest or pytest-json-report not available")
	}

	tmp := t.TempDir()
	solutionDir := filepath.Join(tmp, "solution")
	os.MkdirAll(solutionDir, 0755)
	testsDir := filepath.Join(tmp, "tests")
	os.MkdirAll(testsDir, 0755)

	// Write a trivial module for import
	os.WriteFile(filepath.Join(solutionDir, "add.py"), []byte("def add(a, b): return a + b\n"), 0644)

	// 2 tests pass, 1 fails intentionally — actual pass rate is 2/3 ≈ 6.7
	testCode := `import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'solution'))
from add import add

def test_add_positive():
    assert add(1, 2) == 3

def test_add_zero():
    assert add(0, 0) == 0

def test_add_intentional_failure():
    assert add(1, 1) == 99  # intentionally wrong
`
	os.WriteFile(filepath.Join(testsDir, "test_add.py"), []byte(testCode), 0644)

	// Pass a fabricated "agent stdout" value — the new ScoreTests ignores it.
	// Old code: "100 passed 0 failed" → 10.0
	// New code: runs pytest, 2/3 pass → ~6.7
	score := scoring.ScoreTests(context.Background(), tmp, "pytest tests/ --json-report")

	if score >= 10.0 {
		t.Errorf("ScoreTests = %f, must NOT be 10.0 — fabricated stdout must be ignored", score)
	}
	if score <= 0.0 {
		t.Errorf("ScoreTests = %f, want > 0 (2 of 3 tests pass)", score)
	}
	// 2/3 * 10 = 6.666... rounds to 6.7
	if score < 6.0 || score > 7.5 {
		t.Errorf("ScoreTests = %f, want ~6.7 (2 of 3 tests pass)", score)
	}
}

// TestScoreTests_FabricatedStdout_DoesNotScorePerfect verifies that when there are
// no test files to run, the score is 0.0 regardless of what the agent said.
func TestScoreTests_FabricatedStdout_DoesNotScorePerfect(t *testing.T) {
	if !pytestAvailable() {
		t.Skip("pytest or pytest-json-report not available")
	}

	tmp := t.TempDir()
	// No test files — pytest finds nothing to run; total == 0 → score 0.0

	score := scoring.ScoreTests(context.Background(), tmp, "pytest . --json-report")
	if score != 0.0 {
		t.Errorf("ScoreTests = %f, want 0.0 — empty dir must score 0 regardless of agent stdout", score)
	}
}
