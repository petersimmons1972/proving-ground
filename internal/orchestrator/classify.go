package orchestrator

import (
	"strings"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// classifyFailure maps a runAndScore error to a FailureClass enum (#60).
// Falls back to runner_error rather than judge_failure when the cause is
// ambiguous; over-classification of judge issues would corrupt downstream
// stats about judge reliability.
func classifyFailure(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "timeout"):
		return results.FailureClassTimeout
	case strings.Contains(msg, "score_with_judge") ||
		strings.Contains(msg, "ALL_JUDGE_RUNS_FAILED") ||
		strings.Contains(msg, "INSUFFICIENT_JUDGE_QUORUM"):
		return results.FailureClassJudgeFailure
	default:
		return results.FailureClassRunnerError
	}
}
