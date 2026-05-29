package orchestrator

import (
	"strings"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// classifyFailure maps a runAndScore error to a FailureClass enum (#60).
// Falls back to runner_error rather than judge_failure when the cause is
// ambiguous; over-classification of judge issues would corrupt downstream
// stats about judge reliability.
//
// Branch order is significant: timeout is checked first (exit 124 is the
// timeout sentinel), then judge keywords, then the explicit runner_error
// prefix (produced by runAndScore for any other non-zero exit), and finally
// the catch-all default. The explicit runner_error branch (#71) ensures a
// future contributor adding a new keyword branch above the default cannot
// accidentally re-route "runner_error: runner exit N" messages.
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
	case strings.Contains(msg, "runner_error"):
		// Explicit branch for runner exit codes (#71). runAndScore produces
		// "runner_error: runner exit N" for any non-zero, non-124 exit.
		// Matching this before the default prevents accidental re-routing if
		// new keyword branches are inserted above the default in the future.
		return results.FailureClassRunnerError
	default:
		return results.FailureClassRunnerError
	}
}
