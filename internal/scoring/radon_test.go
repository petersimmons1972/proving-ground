package scoring_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/psimmons/proving-ground/internal/scoring"
)

func TestScoreComplexityNoPyFilesReturnsFive(t *testing.T) {
	tmp := t.TempDir()
	score := scoring.ScoreComplexity(context.Background(), tmp)
	if score != 5.0 {
		t.Errorf("score = %f, want 5.0", score)
	}
}

func TestScoreComplexitySimpleFunction(t *testing.T) {
	if _, err := exec.LookPath("radon"); err != nil {
		t.Skip("radon not in PATH")
	}
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "solution.py"), []byte("def add(a, b):\n    return a + b\n"), 0644)
	score := scoring.ScoreComplexity(context.Background(), tmp)
	if score != 10.0 {
		t.Errorf("score = %f, want 10.0 (simple function = grade A)", score)
	}
}
