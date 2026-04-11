package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/psimmons/proving-ground/internal/runner"
	"github.com/psimmons/proving-ground/internal/scoring"
	"github.com/psimmons/proving-ground/internal/tasks"
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

// TestPartialTierRunRenormalizesWeights verifies that when only 2 of 3 tiers have
// results, the overall score is computed using renormalized weights (summing only
// the present tiers' weights to 1.0), and that a WARNING is written to stderr.
func TestPartialTierRunRenormalizesWeights(t *testing.T) {
	dataDir := t.TempDir()

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	// All tasks score 8.0 overall (each dimension = 8.0 for simplicity).
	// With tiers 1+2 only (weights 0.25 and 0.35), renormalized:
	//   totalWeight = 0.25+0.35 = 0.60
	//   tier1 contribution = 8.0 * (0.25/0.60) ≈ 3.333
	//   tier2 contribution = 8.0 * (0.35/0.60) ≈ 4.667
	//   overall ≈ 8.0 (a flat score distribution stays at 8.0 regardless of weights)
	// Actually any uniform score stays at the same value. Let's use a non-uniform
	// mock: tier1 tasks score 6.0, tier2 tasks score 9.0.
	// With renormalized weights: overall = 6.0*(0.25/0.60) + 9.0*(0.35/0.60)
	//   = 6.0*0.4167 + 9.0*0.5833 = 2.5 + 5.25 = 7.75 → rounds to 7.8
	// Without renormalization (wrong): 6.0*0.25 + 9.0*0.35 + 0*0.40 = 1.5+3.15 = 4.65 → rounds to 4.7
	//
	// To get tier1 tasks overall=6.0 and tier2 tasks overall=9.0, we need the
	// dimension means to equal those values. Since correctness=(auto+RI)/2,
	// and we can set RI and auto via mocks, it's easiest to return a uniform score.
	// Simplest: use uniform judge scores so overall ≈ mean of 6 dims.
	// We'll just set mockJudge for all and verify the overall > 5 (vs 4.65 without renorm).

	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		r := *mockTaskResult
		r.WorkingDir = t.TempDir()
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
		return mockJudge, nil
	}

	// Capture stderr to verify the WARNING message.
	oldStderr := os.Stderr
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = stderrW

	cfg := Config{
		DataDir:     dataDir,
		Tiers:       []string{"1", "2"}, // 2-tier run — no tier 3 tasks
		TasksDir:    findTasksDir(t),
		ControlsDir: findControlsDir(t),
		TemplateDir: findTemplateDir(t),
		PromptDir:   findPromptDir(t),
	}
	runErr := Run(context.Background(), cfg)

	// Restore stderr and read captured output.
	stderrW.Close()
	os.Stderr = oldStderr
	var stderrBuf strings.Builder
	buf := make([]byte, 4096)
	for {
		n, readErr := stderrR.Read(buf)
		if n > 0 {
			stderrBuf.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}
	stderrR.Close()

	if runErr != nil {
		t.Fatal(runErr)
	}

	// Verify WARNING was emitted.
	stderrOutput := stderrBuf.String()
	if !strings.Contains(stderrOutput, "WARNING") {
		t.Errorf("expected WARNING in stderr for partial-tier run, got: %q", stderrOutput)
	}

	// Read back results.json and verify overall score is renormalized.
	jsonData, err := os.ReadFile(filepath.Join(dataDir, "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Scores map[string]map[string]float64 `json:"scores"`
	}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatal(err)
	}

	// With uniform mockJudge scores and a 2-tier run:
	// - Without renormalization: overall = tier1_score*0.25 + tier2_score*0.35 (max 0.6 of full)
	// - With renormalization: overall = tier1_score*(0.25/0.60) + tier2_score*(0.35/0.60) = full score
	// Since all scores are the same value X, renormalized overall = X, non-renormalized = X*0.60.
	// We just need to verify that for any config, overall > tier1 (which would be capped at 0.60 otherwise).
	for configName, scores := range parsed.Scores {
		overall := scores["overall"]
		tier1 := scores["tier1"]
		// With renorm: overall should equal tier scores (since they're the same flat value).
		// Without renorm: overall would be tier1*0.25 + tier2*0.35 = 0.6 * tier_score < tier_score.
		if overall < tier1*0.9 {
			t.Errorf("config %q: overall %f < 90%% of tier1 %f; suggests weights not renormalized", configName, overall, tier1)
		}
	}
}

// TestConfigNamesAreSorted verifies that the Configurations list in the report is
// sorted alphabetically, regardless of map iteration order.
func TestConfigNamesAreSorted(t *testing.T) {
	// Create three profiles whose names would NOT be in alpha order if returned
	// directly from map iteration: "zebra", "apple", "mango".
	dataDir := t.TempDir()
	profilesDir := filepath.Join(dataDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"zebra", "apple", "mango"} {
		if err := os.WriteFile(filepath.Join(profilesDir, name+".txt"), []byte("You are "+name+"."), 0644); err != nil {
			t.Fatal(err)
		}
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

	// Read back the JSON and verify configurations are sorted.
	jsonData, err := os.ReadFile(filepath.Join(dataDir, "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Configurations []string `json:"configurations"`
	}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(parsed.Configurations); i++ {
		if parsed.Configurations[i-1] > parsed.Configurations[i] {
			t.Errorf("configurations not sorted: %v", parsed.Configurations)
			break
		}
	}
}

// TestLOCScoringUsesCorrectSubdir verifies that runAndScore passes
// filepath.Join(result.WorkingDir, "solution") to ScoreLinesOfCode, not
// result.WorkingDir directly. The test writes a large .py file in the base dir
// and a minimal .py file in solution/ — the scores differ, so the test can
// tell which path was actually used.
func TestLOCScoringUsesCorrectSubdir(t *testing.T) {
	workDir := t.TempDir()

	// Write a large (verbose) .py file in the base working dir.
	// With refMinimal=10, refVerbose=80 and 50 lines, score = ~(1 - 40/70)*10 ≈ 4.3
	var bigLines []string
	for i := 0; i < 50; i++ {
		bigLines = append(bigLines, "x = "+string(rune('0'+i%10)))
	}
	bigContent := strings.Join(bigLines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(workDir, "big.py"), []byte(bigContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a small (minimal) .py file in the solution/ subdir — 5 lines → score = 10.0
	solutionDir := filepath.Join(workDir, "solution")
	if err := os.MkdirAll(solutionDir, 0755); err != nil {
		t.Fatal(err)
	}
	smallContent := "a = 1\nb = 2\nc = 3\nd = 4\ne = 5\n"
	if err := os.WriteFile(filepath.Join(solutionDir, "small.py"), []byte(smallContent), 0644); err != nil {
		t.Fatal(err)
	}

	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		r := *mockTaskResult
		r.WorkingDir = workDir
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
		return mockJudge, nil
	}

	task := taskStub{id: "parse-01", tier: 1}
	ts, err := runAndScore(context.Background(), task.toTask(), "zero", "sys", t.TempDir(), "run1", filepath.Join(findWorktreeRoot(t), "prompts", "judge-rubric.md"))
	if err != nil {
		t.Fatal(err)
	}

	// If LOC scoring used solution/ (5 lines ≤ refMinimal=10), loc score = 10.0.
	// elegance = (loc + complexity) / 2. complexity from an empty-ish dir ≈ 5.0 default.
	// So if loc=10.0 and complexity=5.0, elegance ≈ 7.5.
	// If LOC scoring used the base dir (50 lines, refMinimal=10, refVerbose=80), loc ≈ 4.3.
	// We just need to verify loc portion is 10.0 (solution/ path), not ≈ 4.3 (base dir path).
	// Since elegance = (loc + complexity)/2, and complexity for solution/ dir is 5.0 (no py tool run),
	// a loc of 10.0 gives elegance = 7.5, while loc ≈ 4.3 gives elegance ≈ 4.65.
	// We pick 6.0 as the discriminant threshold.
	if ts.elegance < 6.0 {
		t.Errorf("elegance = %f, want >= 6.0 (expected LOC from solution/ dir, not base dir)", ts.elegance)
	}
}

// TestRunAndScoreReturnsTypedResult verifies runAndScore returns a taskScore struct
// with populated tier and named dimension fields accessible without type assertions.
func TestRunAndScoreReturnsTypedResult(t *testing.T) {
	orig1, orig2 := runTaskFn, scoreWithJudgeFn
	defer func() { runTaskFn, scoreWithJudgeFn = orig1, orig2 }()

	workDir := t.TempDir()
	runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
		r := *mockTaskResult
		r.WorkingDir = workDir
		return &r, nil
	}
	scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
		return mockJudge, nil
	}

	task := taskStub{id: "task-001", tier: 2}
	ts, err := runAndScore(context.Background(), task.toTask(), "zero", "sys", t.TempDir(), "run1", filepath.Join(findWorktreeRoot(t), "prompts", "judge-rubric.md"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify struct fields are accessible without type assertions.
	if ts.tier != 2 {
		t.Errorf("tier = %d, want 2", ts.tier)
	}
	if ts.taskID != "task-001" {
		t.Errorf("taskID = %q, want %q", ts.taskID, "task-001")
	}
	// correctness = (auto.TestsPass + judge.RequirementInterpretation) / 2
	// auto.TestsPass from "1 passed" exit 0 = 10.0, judge.RequirementInterpretation = 7 → (10+7)/2 = 8.5
	if ts.correctness != 8.5 {
		t.Errorf("correctness = %f, want 8.5", ts.correctness)
	}

	// dim() accessor must work for each named dimension.
	for _, dim := range []string{"correctness", "elegance", "discipline", "judgment", "creativity", "recovery"} {
		got := ts.dim(dim)
		_ = got // just verify it compiles and doesn't panic
	}
	// Unknown dim returns 0.
	if got := ts.dim("nonexistent"); got != 0 {
		t.Errorf("dim(nonexistent) = %f, want 0", got)
	}
}

// taskStub is a simple helper to build a tasks.Task for tests.
type taskStub struct {
	id   string
	tier int
}

func (s taskStub) toTask() tasks.Task {
	return tasks.Task{ID: s.id, Tier: s.tier, Spec: "spec text"}
}

// mockError is a simple error type for test stubs.
type mockError struct{ msg string }

func (e *mockError) Error() string { return e.msg }
