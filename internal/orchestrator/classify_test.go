package orchestrator

import (
	"errors"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// TestClassifyFailure_TableDriven pins the mapping for every documented input
// class. This is a characterization test: it must pass on the current code,
// then remain byte-identical after the explicit runner_error branch is added.
// Any change to these assertions indicates a behavior regression. (#71 Item 1)
func TestClassifyFailure_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		errMsg    string
		wantClass string
	}{
		// nil error → empty string (no failure)
		{
			name:      "nil_error",
			errMsg:    "", // special-cased below
			wantClass: "",
		},
		// runner_error prefix → FailureClassRunnerError
		{
			name:      "runner_error_exit_1",
			errMsg:    "runner_error: runner exit 1",
			wantClass: results.FailureClassRunnerError,
		},
		{
			name:      "runner_error_exit_2",
			errMsg:    "runner_error: runner exit 2",
			wantClass: results.FailureClassRunnerError,
		},
		{
			name:      "runner_error_exit_99",
			errMsg:    "runner_error: runner exit 99",
			wantClass: results.FailureClassRunnerError,
		},
		// timeout keyword → FailureClassTimeout
		{
			name:      "timeout_keyword",
			errMsg:    "timeout: task exceeded deadline (exit 124)",
			wantClass: results.FailureClassTimeout,
		},
		{
			name:      "timeout_in_middle",
			errMsg:    "some prefix: timeout occurred",
			wantClass: results.FailureClassTimeout,
		},
		// judge keywords → FailureClassJudgeFailure
		{
			name:      "score_with_judge_keyword",
			errMsg:    "score_with_judge: all runs failed",
			wantClass: results.FailureClassJudgeFailure,
		},
		{
			name:      "ALL_JUDGE_RUNS_FAILED_keyword",
			errMsg:    "ALL_JUDGE_RUNS_FAILED: no quorum",
			wantClass: results.FailureClassJudgeFailure,
		},
		{
			name:      "INSUFFICIENT_JUDGE_QUORUM_keyword",
			errMsg:    "INSUFFICIENT_JUDGE_QUORUM: only 1 of 3",
			wantClass: results.FailureClassJudgeFailure,
		},
		// unknown string → default maps to FailureClassRunnerError (documented default)
		{
			name:      "unknown_defaults_to_runner_error",
			errMsg:    "something completely unexpected",
			wantClass: results.FailureClassRunnerError,
		},
		{
			name:      "empty_error_string_defaults_to_runner_error",
			errMsg:    "non-nil error with empty-ish message",
			wantClass: results.FailureClassRunnerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.name == "nil_error" {
				err = nil
			} else {
				err = errors.New(tc.errMsg)
			}
			got := classifyFailure(err)
			if got != tc.wantClass {
				t.Errorf("classifyFailure(%q) = %q, want %q", tc.errMsg, got, tc.wantClass)
			}
		})
	}
}
