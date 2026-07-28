package results_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// makeSampleReport mirrors _make_sample_report from tests/test_results.py.
func makeSampleReport() results.ResultsReport {
	return results.ResultsReport{
		RunID:            "2026-03-30T12:00:00",
		TaskSuiteVersion: "v1",
		Configurations:   []string{"zero", "light", "user"},
		Scores: map[string]map[string]float64{
			"zero":  {"overall": 4.2, "tier1": 5.0, "tier2": 4.0, "tier3": 3.5},
			"light": {"overall": 6.1, "tier1": 6.5, "tier2": 6.0, "tier3": 5.8},
			"user":  {"overall": 8.3, "tier1": 8.0, "tier2": 8.5, "tier3": 8.4},
		},
		DimensionScores: map[string]map[string]float64{
			"zero":  {"correctness": 5.0, "elegance": 3.0, "discipline": 2.0, "judgment": 3.5, "creativity": 2.5, "recovery": 4.0},
			"light": {"correctness": 6.5, "elegance": 5.5, "discipline": 5.0, "judgment": 6.0, "creativity": 5.0, "recovery": 6.5},
			"user":  {"correctness": 8.5, "elegance": 8.0, "discipline": 8.5, "judgment": 8.0, "creativity": 9.0, "recovery": 7.5},
		},
	}
}

// findTemplateDir returns the absolute path to the templates/ directory.
// Tests run from internal/results/, templates live two levels up.
func findTemplateDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Join(wd, "..", "..", "templates")
}

// generate runs GenerateResults against a temp dir and returns (htmlPath, jsonPath).
func generate(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	htmlPath := filepath.Join(tmp, "results.html")
	jsonPath := filepath.Join(tmp, "results.json")
	report := makeSampleReport()
	if err := results.GenerateResults(report, htmlPath, jsonPath, findTemplateDir(t)); err != nil {
		t.Fatalf("GenerateResults: %v", err)
	}
	return htmlPath, jsonPath
}

func TestGenerateResultsProducesHTML(t *testing.T) {
	htmlPath, jsonPath := generate(t)
	if _, err := os.Stat(htmlPath); err != nil {
		t.Errorf("html file missing: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("json file missing: %v", err)
	}
}

func TestHTMLContainsProvingGround(t *testing.T) {
	htmlPath, _ := generate(t)
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	if !strings.Contains(string(data), "Proving Ground") {
		t.Error(`expected "Proving Ground" in html output`)
	}
}

func TestHTMLContainsAllScores(t *testing.T) {
	htmlPath, _ := generate(t)
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	html := string(data)
	for _, score := range []string{"8.3", "6.1", "4.2"} {
		if !strings.Contains(html, score) {
			t.Errorf("expected score %q in html output", score)
		}
	}
}

func TestHTMLContainsAllConfigs(t *testing.T) {
	htmlPath, _ := generate(t)
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	html := string(data)
	for _, name := range []string{"zero", "light", "user"} {
		if !strings.Contains(html, name) {
			t.Errorf("expected config name %q in html output", name)
		}
	}
}

func TestJSONOutputIsValid(t *testing.T) {
	_, jsonPath := generate(t)
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var parsed struct {
		RunID  string                        `json:"run_id"`
		Scores map[string]map[string]float64 `json:"scores"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if parsed.RunID != "2026-03-30T12:00:00" {
		t.Errorf("run_id = %q, want %q", parsed.RunID, "2026-03-30T12:00:00")
	}
	if got := parsed.Scores["user"]["overall"]; got != 8.3 {
		t.Errorf("scores.user.overall = %v, want 8.3", got)
	}

	// Verify "run_id" key is present.
	if !bytes.Contains(data, []byte(`"run_id"`)) {
		t.Error("results.json missing run_id key")
	}

	// Verify Configurations order in "scores": zero before light before user.
	// Scope the search to the scores section only (before dimension_scores)
	// so that keys appearing in dimension_scores don't produce false positives.
	rawJSON := data
	scoresStart := bytes.Index(rawJSON, []byte(`"scores":`))
	if scoresStart == -1 {
		t.Fatal("scores section not found")
	}
	dimStart := bytes.Index(rawJSON, []byte(`"dimension_scores":`))
	if dimStart == -1 {
		t.Fatal("dimension_scores section not found")
	}
	scoresSection := rawJSON[scoresStart:dimStart]

	zeroIdx := bytes.Index(scoresSection, []byte(`"zero"`))
	lightIdx := bytes.Index(scoresSection, []byte(`"light"`))
	userIdx := bytes.Index(scoresSection, []byte(`"user"`))
	if !(zeroIdx < lightIdx && lightIdx < userIdx) {
		t.Errorf("scores keys not in Configurations order: zero=%d light=%d user=%d", zeroIdx, lightIdx, userIdx)
	}

	// Verify DimensionNames order in "dimension_scores": correctness before elegance before discipline.
	// Find the dimension_scores section first, then check order within it.
	dimIdx := bytes.Index(data, []byte(`"dimension_scores"`))
	if dimIdx < 0 {
		t.Fatal("dimension_scores key not found in JSON")
	}
	dimSection := data[dimIdx:]
	idxCorrectness := bytes.Index(dimSection, []byte(`"correctness"`))
	idxElegance := bytes.Index(dimSection, []byte(`"elegance"`))
	idxDiscipline := bytes.Index(dimSection, []byte(`"discipline"`))
	if !(idxCorrectness < idxElegance && idxElegance < idxDiscipline) {
		t.Errorf("dimension_scores inner key order wrong: correctness=%d elegance=%d discipline=%d; want correctness < elegance < discipline",
			idxCorrectness, idxElegance, idxDiscipline)
	}
}

// TestResultsReport_JSON_ContainsSchemaVersionAndMetadata verifies that a
// ResultsReport with a populated RunMetadata marshals to JSON containing
// schema_version: "3" (v3 adds failure_class + judge_prompt_format) and the metadata block (tasks_failed, tasks_attempted).
func TestResultsReport_JSON_ContainsSchemaVersionAndMetadata(t *testing.T) {
	report := makeSampleReport()
	report.Metadata = results.RunMetadata{
		SchemaVersion:  "3",
		TasksAttempted: 9,
		TasksFailed:    2,
		TaskFailures: []results.TaskFailure{
			{TaskID: "parse-01", Profile: "zero", Reason: "ALL_JUDGE_RUNS_FAILED"},
			{TaskID: "sort-02", Profile: "light", Reason: "run_task: simulated failure"},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed struct {
		Metadata struct {
			SchemaVersion  string `json:"schema_version"`
			TasksAttempted int    `json:"tasks_attempted"`
			TasksFailed    int    `json:"tasks_failed"`
			TaskFailures   []struct {
				TaskID  string `json:"task_id"`
				Profile string `json:"profile"`
				Reason  string `json:"reason"`
			} `json:"task_failures"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.Metadata.SchemaVersion != "3" {
		t.Errorf("schema_version = %q, want %q", parsed.Metadata.SchemaVersion, "3")
	}
	if parsed.Metadata.TasksAttempted != 9 {
		t.Errorf("tasks_attempted = %d, want 9", parsed.Metadata.TasksAttempted)
	}
	if parsed.Metadata.TasksFailed != 2 {
		t.Errorf("tasks_failed = %d, want 2", parsed.Metadata.TasksFailed)
	}
	if len(parsed.Metadata.TaskFailures) != 2 {
		t.Errorf("task_failures len = %d, want 2", len(parsed.Metadata.TaskFailures))
	}
}

// TestResultsReport_JSON_MetadataOmittedWhenEmpty verifies that when TaskFailures
// is empty (all tasks succeeded), task_failures is omitted from JSON (omitempty).
func TestResultsReport_JSON_MetadataOmittedWhenEmpty(t *testing.T) {
	report := makeSampleReport()
	report.Metadata = results.RunMetadata{
		SchemaVersion:  "3",
		TasksAttempted: 9,
		TasksFailed:    0,
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if bytes.Contains(data, []byte(`"task_failures"`)) {
		t.Error("task_failures present in JSON when empty, want omitempty to suppress it")
	}
}

func TestHTMLIsSelfContained(t *testing.T) {
	htmlPath, _ := generate(t)
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read html: %v", err)
	}
	html := string(data)

	extCSS := regexp.MustCompile(`(?i)<link[^>]+rel=["']stylesheet["'][^>]+href=["']http`)
	if extCSS.MatchString(html) {
		t.Error("external stylesheet link found — html must be self-contained")
	}

	extScript := regexp.MustCompile(`(?i)<script[^>]+src=["']http`)
	if extScript.MatchString(html) {
		t.Error("external script src found — html must be self-contained")
	}

	if strings.Contains(html, "fonts.googleapis") {
		t.Error("google fonts reference found — html must be self-contained")
	}
}
