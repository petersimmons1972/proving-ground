package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

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
