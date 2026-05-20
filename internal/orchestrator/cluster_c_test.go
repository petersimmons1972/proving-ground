package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/petersimmons1972/proving-ground/internal/runner"
	"github.com/petersimmons1972/proving-ground/internal/scoring"
)

func extractIntField(t *testing.T, json string, field string) int {
	t.Helper()
	key := `"` + field + `":`
	idx := indexAfter(json, key)
	if idx < 0 {
		t.Fatalf("field %q not found in json", field)
	}
	// skip whitespace
	for idx < len(json) && (json[idx] == ' ' || json[idx] == '\t' || json[idx] == '\n') {
		idx++
	}
	end := idx
	for end < len(json) && json[end] >= '0' && json[end] <= '9' {
		end++
	}
	if idx == end {
		t.Fatalf("no integer after field %q", field)
	}
	n := 0
	for i := idx; i < end; i++ {
		n = n*10 + int(json[i]-'0')
	}
	return n
}

func indexAfter(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i + len(sub)
		}
	}
	return -1
}

// Issue #59: active Claude model must be logged at startup so runs are
// reproducible. Cluster C scope: confirm pre-flight emits a model= log line.
func TestRunBenchmark_LogsActiveModelAtStartup(t *testing.T) {
	dataDir := t.TempDir()
	orig := os.Getenv("PROVING_GROUND_MODEL")
	os.Setenv("PROVING_GROUND_MODEL", "claude-test-model-v1")
	defer os.Setenv("PROVING_GROUND_MODEL", orig)

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

	// Capture log output.
	logFile := filepath.Join(t.TempDir(), "log")
	f, _ := os.Create(logFile)
	logSink := captureLog(f)
	defer logSink()

	cfg := Config{
		DataDir:     dataDir,
		Tiers:       []string{"1"},
		TasksDir:    findTasksDir(t),
		ControlsDir: findControlsDir(t),
		TemplateDir: findTemplateDir(t),
		PromptDir:   findPromptDir(t),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := Run(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	f.Sync()
	f.Close()

	logged, _ := os.ReadFile(logFile)
	if !containsString(string(logged), "claude-test-model-v1") {
		t.Errorf("model identity not logged at startup; log=%q", string(logged))
	}
}

func containsString(haystack, needle string) bool {
	return indexAfter(haystack, needle) >= 0
}
