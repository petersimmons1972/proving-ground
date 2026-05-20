package results_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// Issue #37 follow-up: results.json must be written atomically via temp+rename.
// Verified by absence of leftover .tmp files post-success and valid JSON output.
func TestGenerateResults_NoTempFileLeftBehind(t *testing.T) {
	tmp := t.TempDir()
	jsonPath := filepath.Join(tmp, "results.json")
	htmlPath := filepath.Join(tmp, "results.html")

	tplDir := filepath.Join(tmp, "templates")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tplDir, "results.html.tmpl"),
		[]byte(`<html>{{.RunDate}}</html>`), 0o644); err != nil {
		t.Fatal(err)
	}

	report := results.ResultsReport{RunID: "test-run"}
	if err := results.GenerateResults(report, htmlPath, jsonPath, tplDir); err != nil {
		t.Fatalf("GenerateResults: %v", err)
	}

	entries, _ := os.ReadDir(tmp)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("results.json not valid JSON: %v", err)
	}
}

// Pre-existing partial file at the destination (simulating a crashed prior
// run that wrote half a JSON document) must be cleanly replaced, not appended.
func TestGenerateResults_OverwritesPartialPreviousFile(t *testing.T) {
	tmp := t.TempDir()
	jsonPath := filepath.Join(tmp, "results.json")
	htmlPath := filepath.Join(tmp, "results.html")

	if err := os.WriteFile(jsonPath, []byte(`{"partial":`), 0o644); err != nil {
		t.Fatal(err)
	}

	tplDir := filepath.Join(tmp, "templates")
	os.MkdirAll(tplDir, 0o755)
	os.WriteFile(filepath.Join(tplDir, "results.html.tmpl"), []byte(`x`), 0o644)

	if err := results.GenerateResults(results.ResultsReport{RunID: "r"},
		htmlPath, jsonPath, tplDir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(jsonPath)
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("results.json corrupt after overwrite: %v", err)
	}
}
