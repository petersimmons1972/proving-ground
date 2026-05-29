package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/results"
	"github.com/petersimmons1972/proving-ground/internal/runner"
	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

// Issue #40: when RunTask reports a timeout (ExitCode == 124, empty Stdout),
// the orchestrator MUST NOT send the empty transcript to the judge. Doing so
// produces neutral-5 defaults that conflate "timed out" with "mediocre."
//
// Contract: scoreWithJudgeFn is not called for a timed-out task. The task is
// recorded as a failure with a timeout reason.
func TestRunAndScore_TimedOutTask_SkipsJudge(t *testing.T) {
	dataDir := t.TempDir()
	profilesDir := filepath.Join(dataDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profilesDir, "user.txt"), []byte("p"), 0644); err != nil {
		t.Fatal(err)
	}

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		return &runner.TaskResult{
			TaskID:       args.TaskID,
			ProfileName:  args.ProfileName,
			ExitCode:     124, // timeout sentinel
			Stdout:       "",
			Stderr:       "TIMEOUT: task exceeded deadline",
			WorkingDir:   t.TempDir(),
			FilesWritten: []string{"solution/x.py"},
		}, nil
	}
	var judgeCalls atomic.Int64
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
		judgeCalls.Add(1)
		return mockJudge, nil
	}

	cfg := Config{
		DataDir:     dataDir,
		Tiers:       []string{"1"},
		TasksDir:    findTasksDir(t),
		ControlsDir: findControlsDir(t),
		TemplateDir: findTemplateDir(t),
		PromptDir:   findPromptDir(t),
	}
	// All tasks timing out means Run() returns an aggregate failure error —
	// that is expected and not the property under test. The property under
	// test is that scoreWithJudgeFn was never invoked.
	_ = Run(context.Background(), cfg)

	if judgeCalls.Load() != 0 {
		t.Errorf("expected 0 judge invocations for timed-out tasks, got %d", judgeCalls.Load())
	}
}

// Issue #69: a runner exit with any non-zero code (e.g. exit 1) must be
// classified as runner_error, not judge_failure.
//
// Root cause: the #40 short-circuit only checked ExitCode == 124. Any other
// non-zero exit fell through to scoreWithJudgeFn, which fails on an
// empty/bad transcript and returns an error containing "score_with_judge",
// which classifyFailure then maps to judge_failure — masking the real cause.
//
// Contract:
//   - ExitCode == 1 (or any non-zero, non-124) → failure_class=runner_error, judge never called.
//   - ExitCode == 124 → failure_class=timeout (existing #40 contract unchanged).
//   - ExitCode == 0  → success path unchanged.
func TestRunAndScore_NonZeroExitCode_ClassifiesAsRunnerError(t *testing.T) {
	tests := []struct {
		name           string
		exitCode       int
		wantClass      string
		wantJudgeCalls int64
	}{
		{
			name:           "exit_1_runner_error",
			exitCode:       1,
			wantClass:      results.FailureClassRunnerError,
			wantJudgeCalls: 0,
		},
		{
			name:           "exit_2_runner_error",
			exitCode:       2,
			wantClass:      results.FailureClassRunnerError,
			wantJudgeCalls: 0,
		},
		{
			name:           "exit_124_timeout_unchanged",
			exitCode:       124,
			wantClass:      results.FailureClassTimeout,
			wantJudgeCalls: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dataDir := t.TempDir()
			profilesDir := filepath.Join(dataDir, "profiles")
			if err := os.MkdirAll(profilesDir, 0755); err != nil {
				t.Fatal(err)
			}

			orig1, orig2 := runTaskFn, scoreWithJudgeFn
			defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

			runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
				return &runner.TaskResult{
					TaskID:      args.TaskID,
					ProfileName: args.ProfileName,
					ExitCode:    tc.exitCode,
					Stdout:      "",
					Stderr:      "",
					WorkingDir:  t.TempDir(),
				}, nil
			}
			var judgeCalls atomic.Int64
			scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
				judgeCalls.Add(1)
				return mockJudge, nil
			}

			cfg := Config{
				DataDir:     dataDir,
				Tiers:       []string{"1"},
				TasksDir:    findTasksDir(t),
				ControlsDir: findControlsDir(t),
				TemplateDir: findTemplateDir(t),
				PromptDir:   findPromptDir(t),
			}
			// All tasks fail with runner exit — Run() returns aggregate error, expected.
			_ = Run(context.Background(), cfg)

			// Judge must never be called for a non-zero exit.
			if got := judgeCalls.Load(); got != tc.wantJudgeCalls {
				t.Errorf("judge calls = %d, want %d", got, tc.wantJudgeCalls)
			}

			// Read results.json and verify failure_class for every task_failure entry.
			jsonData, err := os.ReadFile(filepath.Join(dataDir, "results.json"))
			if err != nil {
				t.Fatalf("results.json not written: %v", err)
			}
			var parsed struct {
				Metadata struct {
					TaskFailures []struct {
						FailureClass string `json:"failure_class"`
					} `json:"task_failures"`
				} `json:"metadata"`
			}
			if err := json.Unmarshal(jsonData, &parsed); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if len(parsed.Metadata.TaskFailures) == 0 {
				t.Fatal("expected task_failures in metadata, got none")
			}
			for i, tf := range parsed.Metadata.TaskFailures {
				if tf.FailureClass != tc.wantClass {
					t.Errorf("task_failure[%d].failure_class = %q, want %q", i, tf.FailureClass, tc.wantClass)
				}
			}
		})
	}
}
