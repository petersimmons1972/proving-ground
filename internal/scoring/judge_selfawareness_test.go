package scoring

// TestJudgeRegexHasNoSelfAwareness verifies that parseJudgeOutput does not
// extract SELF_AWARENESS from judge output. This test is in package scoring
// (not scoring_test) to access the unexported parseJudgeOutput function.
// It must fail until SELF_AWARENESS is removed from judgeRe.
import (
	"os"
	"strings"
	"testing"
)

func TestJudgeRegexHasNoSelfAwareness(t *testing.T) {
	raw := parseJudgeOutput("REQUIREMENT_INTERPRETATION: 8\nSELF_AWARENESS: 9\nDECISION_COMMUNICATION: 7\n")
	if _, found := raw.dims["SELF_AWARENESS"]; found {
		t.Error("dims map contains SELF_AWARENESS key — it must be removed from judgeRe")
	}
}

func TestRubricHasNoSelfAwareness(t *testing.T) {
	// Walk up from the package directory to find the repo root and the rubric.
	rubricPath := findRubricPath(t)
	data, err := os.ReadFile(rubricPath)
	if err != nil {
		t.Fatalf("reading rubric: %v", err)
	}
	if strings.Contains(string(data), "SELF_AWARENESS") {
		t.Error("judge-rubric.md still contains SELF_AWARENESS — it must be removed")
	}
}

// findRubricPath locates prompts/judge-rubric.md relative to this test file.
func findRubricPath(t *testing.T) string {
	t.Helper()
	// Tests run from internal/scoring/ — prompts/ is 2 levels up.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd + "/../../prompts/judge-rubric.md"
}
