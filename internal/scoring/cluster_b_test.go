package scoring_test

import (
	"context"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

// Issue #39 (AG-4: fail-the-run): a truncated judge output that only emits
// one dimension must NOT silently default the other three to 5. With only
// one usable observation per run and 1 run total, the result must be marked
// Failed (rather than producing a 5/10 average across missing dims).
func TestScoreWithJudge_TruncatedRun_NotDefaultedToFive(t *testing.T) {
	rubric, cleanup := setupFakeJudgeClaude(t,
		`printf 'REQUIREMENT_INTERPRETATION: 9\n'; exit 0`)
	defer cleanup()

	scores, err := scoring.ScoreWithJudge(context.Background(), "t", "s", rubric, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !scores.Failed {
		t.Fatalf("truncated judge run must be marked Failed; got scores=%+v", scores)
	}
	for name, v := range map[string]float64{
		"DecisionCommunication":  scores.DecisionCommunication,
		"RecoveryQuality":        scores.RecoveryQuality,
		"UnconventionalThinking": scores.UnconventionalThinking,
	} {
		if v == 5 {
			t.Errorf("%s = 5 (silent default leak); want 0 or sentinel", name)
		}
	}
}

// Issue #30: ScoreScope must NOT return 10.0 when nothing was written.
// Writing zero files is "did no work," not "perfect scope discipline."
func TestScoreScope_EmptyFilesWrittenIsNotPerfect(t *testing.T) {
	got := scoring.ScoreScope([]string{"solution/"}, nil)
	if got >= 5.0 {
		t.Errorf("ScoreScope([], nil) = %f, want low score (<5) for no-work", got)
	}
}
