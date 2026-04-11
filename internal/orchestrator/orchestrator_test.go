package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/psimmons/proving-ground/internal/runner"
	"github.com/psimmons/proving-ground/internal/scoring"
)

// findWorktreeRoot walks up from the test directory to find the worktree root
// (the directory containing go.mod).
func findWorktreeRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Tests run from internal/orchestrator/ — root is 2 levels up.
	return filepath.Join(wd, "..", "..")
}

func findTasksDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(findWorktreeRoot(t), "tasks")
}

func findControlsDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(findWorktreeRoot(t), "profiles")
}

func findTemplateDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(findWorktreeRoot(t), "templates")
}

func findPromptDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(findWorktreeRoot(t), "prompts")
}

// mockTaskResult is a baseline successful TaskResult used across tests.
// WorkingDir must be set per-call because it needs a real (even empty) directory
// for the scoring functions to walk safely.
var mockTaskResult = &runner.TaskResult{
	ExitCode:     0,
	Stdout:       "1 passed",
	Stderr:       "",
	FilesWritten: []string{"solution/parser.py"},
}

var mockJudge = &scoring.JudgeScores{
	RequirementInterpretation: 7,
	DecisionCommunication:     6,
	SelfAwareness:             7,
	RecoveryQuality:           6,
	UnconventionalThinking:    5,
	Rationale:                 "ok",
}

// TestRunBenchmarkProducesHTML verifies that Run writes results.html and results.json.
func TestRunBenchmarkProducesHTML(t *testing.T) {
	dataDir := t.TempDir()
	profilesDir := filepath.Join(dataDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profilesDir, "my-agent.txt"), []byte("You are a careful engineer."), 0644); err != nil {
		t.Fatal(err)
	}

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		r := *mockTaskResult
		r.WorkingDir = t.TempDir()
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
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
	if err := Run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "results.html")); err != nil {
		t.Error("results.html missing")
	}
	if _, err := os.Stat(filepath.Join(dataDir, "results.json")); err != nil {
		t.Error("results.json missing")
	}
}

// TestRunBenchmarkRunsEachConfig verifies RunTask is called once per task × profile.
// Tier 1 has 3 tasks; profiles are zero, light, and one user profile = 3.
// Expected calls: 3 tasks × 3 profiles = 9.
func TestRunBenchmarkRunsEachConfig(t *testing.T) {
	dataDir := t.TempDir()
	profilesDir := filepath.Join(dataDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profilesDir, "user.txt"), []byte("You are a Go engineer."), 0644); err != nil {
		t.Fatal(err)
	}

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	var callCount atomic.Int64
	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		callCount.Add(1)
		r := *mockTaskResult
		r.WorkingDir = t.TempDir()
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
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
	if err := Run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	got := callCount.Load()
	if got != 9 {
		t.Errorf("expected 9 RunTask calls (3 tasks × 3 profiles), got %d", got)
	}
}

// TestRunBenchmarkAppendsHistory verifies that history.jsonl is written after a run.
func TestRunBenchmarkAppendsHistory(t *testing.T) {
	dataDir := t.TempDir()

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		r := *mockTaskResult
		r.WorkingDir = t.TempDir()
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
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
	if err := Run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "history.jsonl")); err != nil {
		t.Error("history.jsonl missing")
	}
}

// TestRunBenchmarkSurvivesTaskFailure verifies that a single profile failure does not
// abort the whole run. One of the N RunTask calls returns an error; results.html
// must still be produced.
func TestRunBenchmarkSurvivesTaskFailure(t *testing.T) {
	dataDir := t.TempDir()

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	var callCount atomic.Int64
	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		n := callCount.Add(1)
		// Fail the first call; all subsequent succeed.
		if n == 1 {
			return nil, &mockError{"simulated task failure"}
		}
		r := *mockTaskResult
		r.WorkingDir = t.TempDir()
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
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
	if err := Run(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "results.html")); err != nil {
		t.Error("results.html missing after partial task failure")
	}
}

// mockError is a simple error type for test stubs.
type mockError struct{ msg string }

func (e *mockError) Error() string { return e.msg }
