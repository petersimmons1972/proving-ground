// Package results contains the report data model and HTML/JSON renderers
// for a Proving Ground benchmark run.
package results

import (
	"encoding/json"
	"fmt"
	"os"
)

// ResultsReport holds all scored data for a benchmark run.
// JSON field names match Python's dataclass asdict() output exactly,
// so JSON written by the Go port is byte-compatible with the Python port.
type ResultsReport struct {
	RunID            string                        `json:"run_id"`
	TaskSuiteVersion string                        `json:"task_suite_version"`
	Configurations   []string                      `json:"configurations"`
	Scores           map[string]map[string]float64 `json:"scores"`
	DimensionScores  map[string]map[string]float64 `json:"dimension_scores"`
	History          interface{}                   `json:"history"`
}

// defaultConfigColors mirrors DEFAULT_CONFIG_COLORS from the Python port.
// These are the brand colors used for each configuration in the HTML report.
var defaultConfigColors = map[string]string{
	"zero":  "#64748b",
	"light": "#22D3EE",
	"user":  "#F59E0B",
}

// colorFor returns the brand color for a configuration name,
// falling back to neutral gray for unknown configurations.
func colorFor(config string) string {
	if c, ok := defaultConfigColors[config]; ok {
		return c
	}
	return "#888888"
}

// buildVerdict produces the one-sentence verdict line shown at the top
// of the results HTML. Logic ported verbatim from Python generate_results().
func buildVerdict(report ResultsReport) string {
	configs := report.Configurations
	if len(configs) < 2 {
		return "Run complete."
	}

	best := configs[0]
	bestScore := report.Scores[best]["overall"]
	for _, c := range configs[1:] {
		if s := report.Scores[c]["overall"]; s > bestScore {
			best = c
			bestScore = s
		}
	}

	const baseline = "zero"
	if _, ok := report.Scores[baseline]; ok && best != baseline {
		delta := report.Scores[best]["overall"] - report.Scores[baseline]["overall"]
		return fmt.Sprintf("Your %s agent outperformed the baseline by %.1f points overall.", best, delta)
	}
	return fmt.Sprintf("Best configuration: %s.", best)
}

// GenerateResults writes results.json and results.html for a benchmark run.
// htmlPath and jsonPath are absolute file paths for the two outputs.
// templateDir is the directory containing results.html.tmpl.
func GenerateResults(report ResultsReport, htmlPath, jsonPath, templateDir string) error {
	// JSON output — matches Python json.dumps(asdict(report), indent=2).
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}

	// HTML output — build the template data then render.
	verdict := buildVerdict(report)
	data := buildTemplateData(report, verdict)
	html, err := renderHTML(data, templateDir)
	if err != nil {
		return fmt.Errorf("render html: %w", err)
	}
	if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write html: %w", err)
	}
	return nil
}
