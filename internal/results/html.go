package results

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// dimensionOrder defines the canonical dimension ordering used by the
// Dimension Analysis section. Must match the Python port exactly.
var dimensionOrder = []string{
	"correctness",
	"elegance",
	"discipline",
	"judgment",
	"creativity",
	"recovery",
}

// tierKeys pairs the score map key with its human-readable label.
// Order is significant — this drives the Tier Breakdown section.
var tierKeys = []struct {
	Key   string
	Label string
}{
	{"tier1", "Tier 1 — Craft"},
	{"tier2", "Tier 2 — Judgment"},
	{"tier3", "Tier 3 — Pressure"},
}

// ConfigRow is one row in the Overall Results section (section 01).
type ConfigRow struct {
	Name       string
	Color      string
	ColorAlpha string // color + "88" (semi-transparent) for the dim header
	Overall    string // formatted "%.1f"
	OverallPct int    // int(overall * 10) — width percentage for the bar
}

// TierRow is one config row inside a single tier block (section 02).
type TierRow struct {
	Name     string
	Color    string
	Score    string
	ScorePct int
}

// TierSection is a complete tier block with its label and config rows.
type TierSection struct {
	Label string
	Rows  []TierRow
}

// DimensionCell is one config cell within a dimension row (section 03).
// Opacity is "1" for the max score in the row, otherwise "0.55".
type DimensionCell struct {
	Score    string
	ScorePct int
	Color    string
	Opacity  string
}

// DimensionRow is one dimension across all configurations.
type DimensionRow struct {
	Name  string
	Cells []DimensionCell
}

// TemplateData is the fully pre-computed view model for results.html.tmpl.
// Every value the template interpolates lives here — the template itself
// performs only simple range/field lookups, no logic.
type TemplateData struct {
	RunDate      string
	RunYear      string
	SuiteVersion string
	SuiteRaw     string // original (non-uppercased) suite version
	ConfigCount  int
	GridColumns  template.CSS // dynamic grid-template-columns for dim section
	Verdict      string
	Configs      []ConfigRow
	Tiers        []TierSection
	Dimensions   []DimensionRow
}

// buildTemplateData pre-computes every value the HTML template needs.
// Putting all the formatting here (instead of in the template) keeps
// the template dumb and avoids having to implement Jinja2-style filters
// inside html/template's funcmap.
func buildTemplateData(report ResultsReport, verdict string) TemplateData {
	configCount := len(report.Configurations)

	// Overall Results rows.
	configs := make([]ConfigRow, 0, configCount)
	for _, name := range report.Configurations {
		overall := report.Scores[name]["overall"]
		color := colorFor(name)
		configs = append(configs, ConfigRow{
			Name:       name,
			Color:      color,
			ColorAlpha: color + "88",
			Overall:    fmt.Sprintf("%.1f", overall),
			OverallPct: int(overall * 10),
		})
	}

	// Tier Breakdown sections.
	tiers := make([]TierSection, 0, len(tierKeys))
	for _, tk := range tierKeys {
		rows := make([]TierRow, 0, configCount)
		for _, name := range report.Configurations {
			score := report.Scores[name][tk.Key]
			color := colorFor(name)
			rows = append(rows, TierRow{
				Name:     name,
				Color:    color,
				Score:    fmt.Sprintf("%.1f", score),
				ScorePct: int(score * 10),
			})
		}
		tiers = append(tiers, TierSection{Label: tk.Label, Rows: rows})
	}

	// Dimension rows. For each dimension, the max score across configs
	// gets opacity "1"; everyone else gets "0.55". This highlights
	// which configuration leads on each dimension.
	dims := make([]DimensionRow, 0, len(dimensionOrder))
	for _, dim := range dimensionOrder {
		// Find max for this dimension across all configs.
		maxScore := 0.0
		for _, name := range report.Configurations {
			if s := report.DimensionScores[name][dim]; s > maxScore {
				maxScore = s
			}
		}

		cells := make([]DimensionCell, 0, configCount)
		for _, name := range report.Configurations {
			score := report.DimensionScores[name][dim]
			opacity := "0.55"
			if score == maxScore {
				opacity = "1"
			}
			cells = append(cells, DimensionCell{
				Score:    fmt.Sprintf("%.1f", score),
				ScorePct: int(score * 10),
				Color:    colorFor(name),
				Opacity:  opacity,
			})
		}
		dims = append(dims, DimensionRow{Name: dim, Cells: cells})
	}

	// Grid columns for the dimension section: 160px label column + one
	// flexible column per configuration. Wrapped in template.CSS so
	// html/template knows it's safe to drop into an inline style attribute.
	gridColumns := template.CSS("160px " + strings.Repeat("1fr ", configCount))

	// Safe string slicing — all run_ids are ISO timestamps, but guard
	// against shorter values just in case.
	runDate := report.RunID
	if len(runDate) > 10 {
		runDate = runDate[:10]
	}
	runYear := report.RunID
	if len(runYear) > 4 {
		runYear = runYear[:4]
	}

	return TemplateData{
		RunDate:      runDate,
		RunYear:      runYear,
		SuiteVersion: strings.ToUpper(report.TaskSuiteVersion),
		SuiteRaw:     report.TaskSuiteVersion,
		ConfigCount:  configCount,
		GridColumns:  gridColumns,
		Verdict:      verdict,
		Configs:      configs,
		Tiers:        tiers,
		Dimensions:   dims,
	}
}

// renderHTML loads results.html.tmpl from templateDir and executes it
// against the prepared TemplateData. Returns the rendered HTML string.
func renderHTML(data TemplateData, templateDir string) (string, error) {
	tmplPath := filepath.Join(templateDir, "results.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
