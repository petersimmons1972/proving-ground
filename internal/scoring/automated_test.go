package scoring_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/psimmons/proving-ground/internal/scoring"
)

func TestScoreTestsAllPass(t *testing.T) {
	s := scoring.ScoreTests(0, "5 passed")
	if s.TestsPass != 10.0 {
		t.Errorf("TestsPass = %f, want 10.0", s.TestsPass)
	}
}

func TestScoreTestsAllFail(t *testing.T) {
	s := scoring.ScoreTests(1, "5 failed")
	if s.TestsPass != 0.0 {
		t.Errorf("TestsPass = %f, want 0.0", s.TestsPass)
	}
}

func TestScoreTestsExactPartial(t *testing.T) {
	s := scoring.ScoreTests(1, "3 passed, 2 failed")
	if s.TestsPass != 6.0 {
		t.Errorf("TestsPass = %f, want 6.0", s.TestsPass)
	}
}

func TestScoreTestsNoTestsFound(t *testing.T) {
	s := scoring.ScoreTests(0, "All done, no tests to run")
	if s.TestsPass != 0.0 {
		t.Errorf("TestsPass = %f, want 0.0", s.TestsPass)
	}
}

func TestScoreLOCAtMinimalReturns10(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "solution.py"), []byte("x = 1\ny = 2\nz = 3\n"), 0644)
	score := scoring.ScoreLinesOfCode(tmp, 3, 60)
	if score != 10.0 {
		t.Errorf("score = %f, want 10.0", score)
	}
}

func TestScoreLOCVerbose(t *testing.T) {
	tmp := t.TempDir()
	var lines []string
	for i := 0; i < 80; i++ {
		lines = append(lines, "x_"+string(rune('0'+i%10))+" = "+string(rune('0'+i%10)))
	}
	os.WriteFile(filepath.Join(tmp, "solution.py"), []byte(strings.Join(lines, "\n")), 0644)
	score := scoring.ScoreLinesOfCode(tmp, 5, 60)
	if score > 2.0 {
		t.Errorf("score = %f, want <= 2.0", score)
	}
}

func TestScoreScopeNoCreep(t *testing.T) {
	score := scoring.ScoreScope(
		[]string{"solution/"},
		[]string{"solution/parser.py", "solution/__init__.py"},
	)
	if score != 10.0 {
		t.Errorf("score = %f, want 10.0", score)
	}
}

func TestScoreScopeFullCreep(t *testing.T) {
	score := scoring.ScoreScope(
		[]string{"solution/"},
		[]string{"tests/new_test.py", "README.md"},
	)
	if score != 0.0 {
		t.Errorf("score = %f, want 0.0", score)
	}
}

func TestScoreScopePartialCreep(t *testing.T) {
	score := scoring.ScoreScope(
		[]string{"solution/"},
		[]string{"solution/parser.py", "tests/new_test.py"},
	)
	if score != 5.0 {
		t.Errorf("score = %f, want 5.0", score)
	}
}

func TestScoreLOCNoPyFilesReturnsFivePointZero(t *testing.T) {
	tmp := t.TempDir()
	score := scoring.ScoreLinesOfCode(tmp, 10, 80)
	if score != 5.0 {
		t.Errorf("score = %f, want 5.0", score)
	}
}
