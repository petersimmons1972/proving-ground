// Package results contains the report data model and HTML/JSON renderers
// for a Proving Ground benchmark run.
package results

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/psimmons/proving-ground/internal/config"
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

// MarshalJSON produces a JSON encoding of ResultsReport with explicit key
// ordering to match Python's json.dumps(asdict(report)) output:
//   - top-level keys appear in the struct field order defined here
//   - "scores" and "dimension_scores" config keys follow Configurations order
//   - "dimension_scores" inner keys follow config.DimensionNames order
//   - "scores" inner keys (overall, tier1, tier2, tier3) are alphabetical,
//     which also matches Python's insertion order
func (r ResultsReport) MarshalJSON() ([]byte, error) {
	scoreInnerOrder := []string{"overall", "tier1", "tier2", "tier3"}

	scoresJSON, err := orderedConfigMap(r.Configurations, r.Scores, scoreInnerOrder)
	if err != nil {
		return nil, fmt.Errorf("marshal scores: %w", err)
	}

	dimScoresJSON, err := orderedConfigMap(r.Configurations, r.DimensionScores, config.DimensionNames)
	if err != nil {
		return nil, fmt.Errorf("marshal dimension_scores: %w", err)
	}

	histJSON, err := json.Marshal(r.History)
	if err != nil {
		return nil, fmt.Errorf("marshal history: %w", err)
	}

	configsJSON, err := json.Marshal(r.Configurations)
	if err != nil {
		return nil, fmt.Errorf("marshal configurations: %w", err)
	}

	runIDJSON, _ := json.Marshal(r.RunID)
	versionJSON, _ := json.Marshal(r.TaskSuiteVersion)

	var buf bytes.Buffer
	buf.WriteString(`{"run_id":`)
	buf.Write(runIDJSON)
	buf.WriteString(`,"task_suite_version":`)
	buf.Write(versionJSON)
	buf.WriteString(`,"configurations":`)
	buf.Write(configsJSON)
	buf.WriteString(`,"scores":`)
	buf.Write(scoresJSON)
	buf.WriteString(`,"dimension_scores":`)
	buf.Write(dimScoresJSON)
	buf.WriteString(`,"history":`)
	buf.Write(histJSON)
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

// formatFloat formats a float64 to match Python's json.dumps output.
// Python keeps the decimal point for whole numbers (10.0 → "10.0"),
// and strips trailing zeros for fractional values (8.30 → "8.3").
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// orderedConfigMap serializes map[config]map[key]float64 with configs in the
// provided configs order and inner keys in innerOrder order.
func orderedConfigMap(configs []string, m map[string]map[string]float64, innerOrder []string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, cfg := range configs {
		if i > 0 {
			buf.WriteString(",")
		}
		cfgJSON, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		buf.Write(cfgJSON)
		buf.WriteString(":{")
		inner := m[cfg]
		first := true
		for _, key := range innerOrder {
			val, ok := inner[key]
			if !ok {
				continue
			}
			if !first {
				buf.WriteString(",")
			}
			first = false
			keyJSON, _ := json.Marshal(key)
			buf.Write(keyJSON)
			buf.WriteString(":")
			buf.WriteString(formatFloat(val))
		}
		buf.WriteString("}")
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
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
