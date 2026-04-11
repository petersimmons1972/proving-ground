// Package orchestrator runs the full benchmark pipeline:
// load profiles and tasks, run each task × profile in parallel,
// score all dimensions, aggregate by tier, write results.html/json, append history.
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/psimmons/proving-ground/internal/config"
	"github.com/psimmons/proving-ground/internal/history"
	"github.com/psimmons/proving-ground/internal/profiles"
	"github.com/psimmons/proving-ground/internal/results"
	"github.com/psimmons/proving-ground/internal/runner"
	"github.com/psimmons/proving-ground/internal/scoring"
	"github.com/psimmons/proving-ground/internal/tasks"
)

// Config holds all directory paths and runtime settings for a benchmark run.
type Config struct {
	DataDir     string   // where to write results.html, results.json, history.jsonl, and runs/
	Tiers       []string // e.g. ["1"], ["2"], ["1","2","3"]
	TasksDir    string   // path to tasks/ directory (contains tier1/, tier2/, tier3/)
	ControlsDir string   // path to built-in profiles/ directory (zero.txt, light.txt)
	TemplateDir string   // path to templates/ directory
	PromptDir   string   // path to prompts/ directory (for judge-rubric.md)
}

// taskScore holds the computed scores for one task × profile combination.
// Using a typed struct eliminates unprotected type assertions on map[string]interface{}.
type taskScore struct {
	taskID      string
	tier        int
	correctness float64
	elegance    float64
	discipline  float64
	judgment    float64
	creativity  float64
	recovery    float64
}

// dim returns the named dimension score. Returns 0 for unknown names.
func (t taskScore) dim(name string) float64 {
	switch name {
	case "correctness":
		return t.correctness
	case "elegance":
		return t.elegance
	case "discipline":
		return t.discipline
	case "judgment":
		return t.judgment
	case "creativity":
		return t.creativity
	case "recovery":
		return t.recovery
	}
	return 0
}

// These are package-level vars so tests can replace them without an interface layer.
// This mirrors Python's unittest.mock.patch pattern exactly.
var runTaskFn = func(ctx context.Context, args runner.RunTaskArgs) (*runner.TaskResult, error) {
	return runner.RunTask(ctx, args)
}

var scoreWithJudgeFn = func(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*scoring.JudgeScores, error) {
	return scoring.ScoreWithJudge(ctx, transcript, taskSpec, rubricPath, runs)
}

// Run executes the full benchmark pipeline: run all tasks × profiles, score, aggregate, report.
func Run(ctx context.Context, cfg Config) error {
	// Create runs output directory.
	runsPath := filepath.Join(cfg.DataDir, "runs")
	if err := os.MkdirAll(runsPath, 0755); err != nil {
		return fmt.Errorf("creating runs dir: %w", err)
	}

	// Load profiles: zero, light, and any user-supplied profiles.
	userDir := filepath.Join(cfg.DataDir, "profiles")
	profileMap, err := profiles.LoadProfiles(cfg.ControlsDir, userDir)
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}

	// Load tasks for the requested tiers.
	taskList, err := tasks.LoadTasks(cfg.TasksDir, cfg.Tiers)
	if err != nil {
		return fmt.Errorf("loading tasks: %w", err)
	}

	// run_id format matches Python: datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S")
	runID := time.Now().UTC().Format("2006-01-02T15:04:05")

	// rubricPath for judge calls.
	rubricPath := filepath.Join(cfg.PromptDir, "judge-rubric.md")

	// allTaskScores accumulates per-profile scored results.
	// Key: profile name. Value: list of per-task typed scores.
	allTaskScores := make(map[string][]taskScore)
	for name := range profileMap {
		allTaskScores[name] = nil
	}

	// Concurrency: tasks run sequentially in outer loop;
	// profiles-within-a-task run in parallel via goroutines + semaphore + WaitGroup.
	// This mirrors Python's ThreadPoolExecutor(max_workers=_MAX_WORKERS) per task.
	sem := make(chan struct{}, config.MaxWorkers())

	type profileResult struct {
		name  string
		score taskScore
	}

	for _, task := range taskList {
		var wg sync.WaitGroup
		resultsCh := make(chan profileResult, len(profileMap))

		for name, prompt := range profileMap {
			wg.Add(1)
			sem <- struct{}{}
			go func(name, prompt string) {
				defer wg.Done()
				defer func() { <-sem }()

				taskScore, err := runAndScore(ctx, task, name, prompt, runsPath, runID, rubricPath)
				if err != nil {
					log.Printf("WARNING: profile %q failed on %s: %v", name, task.ID, err)
					return
				}
				resultsCh <- profileResult{name: name, score: taskScore}
			}(name, prompt)
		}

		wg.Wait()
		close(resultsCh)

		for pr := range resultsCh {
			allTaskScores[pr.name] = append(allTaskScores[pr.name], pr.score)
		}
	}

	// Aggregate scores: port Python aggregation exactly.
	scoresByConfig := make(map[string]map[string]float64)
	dimensionScoresByConfig := make(map[string]map[string]float64)

	for profileName, taskScoreList := range allTaskScores {
		// tier -> list of per-task overall scores
		tierAverages := map[int][]float64{1: nil, 2: nil, 3: nil}
		// dimension -> list of per-task dimension scores
		dimValues := make(map[string][]float64)
		for _, dim := range config.DimensionNames {
			dimValues[dim] = nil
		}

		for _, ts := range taskScoreList {
			// per-task overall = mean of the 6 dimension scores
			var dimScores []float64
			for _, dim := range config.DimensionNames {
				dimScores = append(dimScores, ts.dim(dim))
			}
			overall := mean(dimScores)
			tierAverages[ts.tier] = append(tierAverages[ts.tier], overall)

			for _, dim := range config.DimensionNames {
				dimValues[dim] = append(dimValues[dim], ts.dim(dim))
			}
		}

		// tier_scores[t] = mean of per-task overalls for that tier (0.0 if empty)
		tierScores := map[int]float64{}
		for t := 1; t <= 3; t++ {
			tierScores[t] = mean(tierAverages[t])
		}

		// Renormalize weights for partial-tier runs.
		// A tier is "present" only if it has at least one task result.
		// Dividing by totalWeight ensures the present tiers sum to 1.0.
		totalWeight := 0.0
		presentTierCount := 0
		for t := 1; t <= 3; t++ {
			if len(tierAverages[t]) > 0 {
				totalWeight += config.TierWeights[t]
				presentTierCount++
			}
		}
		if presentTierCount < 3 {
			fmt.Fprintf(os.Stderr, "WARNING: only %d of 3 tiers have results; renormalizing weights\n", presentTierCount)
		}
		weighted := 0.0
		for t := 1; t <= 3; t++ {
			if len(tierAverages[t]) > 0 && totalWeight > 0 {
				weighted += tierScores[t] * (config.TierWeights[t] / totalWeight)
			}
		}

		scoresByConfig[profileName] = map[string]float64{
			"overall": roundTo1(weighted),
			"tier1":   roundTo1(tierScores[1]),
			"tier2":   roundTo1(tierScores[2]),
			"tier3":   roundTo1(tierScores[3]),
		}

		dimOut := make(map[string]float64)
		for _, dim := range config.DimensionNames {
			dimOut[dim] = roundTo1(mean(dimValues[dim]))
		}
		dimensionScoresByConfig[profileName] = dimOut
	}

	// Build ordered list of configuration names (sorted for determinism).
	configNames := make([]string, 0, len(profileMap))
	for name := range profileMap {
		configNames = append(configNames, name)
	}
	sort.Strings(configNames)

	report := results.ResultsReport{
		RunID:            runID,
		TaskSuiteVersion: config.TaskSuiteVersion,
		Configurations:   configNames,
		Scores:           scoresByConfig,
		DimensionScores:  dimensionScoresByConfig,
		History:          nil,
	}

	htmlPath := filepath.Join(cfg.DataDir, "results.html")
	jsonPath := filepath.Join(cfg.DataDir, "results.json")
	if err := results.GenerateResults(report, htmlPath, jsonPath, cfg.TemplateDir); err != nil {
		return fmt.Errorf("generating results: %w", err)
	}

	historyFile := filepath.Join(cfg.DataDir, "history.jsonl")
	overallByConfig := make(map[string]float64)
	for c, s := range scoresByConfig {
		overallByConfig[c] = s["overall"]
	}
	if err := history.AppendRun(historyFile, history.HistoryEntry{
		RunID:  runID,
		Scores: overallByConfig,
	}); err != nil {
		return fmt.Errorf("appending history: %w", err)
	}

	return nil
}

// runAndScore executes one task+profile combination and returns a typed taskScore.
// Mirrors Python's _run_and_score_profile inner function.
func runAndScore(
	ctx context.Context,
	task tasks.Task,
	profileName, systemPrompt string,
	runsPath, runID, rubricPath string,
) (taskScore, error) {
	result, err := runTaskFn(ctx, runner.RunTaskArgs{
		TaskID:       task.ID,
		TaskSpec:     task.Spec,
		ProfileName:  profileName,
		SystemPrompt: systemPrompt,
		WorkingDir:   filepath.Join(runsPath, runID),
	})
	if err != nil {
		return taskScore{}, fmt.Errorf("run_task: %w", err)
	}

	auto := scoring.ScoreTests(result.ExitCode, result.Stdout)

	locRef, ok := config.LocRefs[task.ID]
	if !ok {
		locRef = [2]int{10, 80}
	}
	loc := scoring.ScoreLinesOfCode(filepath.Join(result.WorkingDir, "solution"), locRef[0], locRef[1])
	complexity := scoring.ScoreComplexity(ctx, filepath.Join(result.WorkingDir, "solution"))
	scope := scoring.ScoreScope([]string{"solution/"}, result.FilesWritten)

	judge, err := scoreWithJudgeFn(ctx, result.Stdout, task.Spec, rubricPath, config.JudgeRuns())
	if err != nil {
		return taskScore{}, fmt.Errorf("score_with_judge: %w", err)
	}

	return taskScore{
		taskID:      task.ID,
		tier:        task.Tier,
		correctness: (auto.TestsPass + judge.RequirementInterpretation) / 2,
		elegance:    (loc + complexity) / 2,
		discipline:  scope,
		judgment:    judge.DecisionCommunication,
		creativity:  judge.UnconventionalThinking,
		recovery:    judge.RecoveryQuality,
	}, nil
}

// mean returns the arithmetic mean of vals, or 0.0 if vals is empty.
func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0.0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// roundTo1 rounds x to one decimal place.
func roundTo1(x float64) float64 {
	return math.Round(x*10) / 10
}
