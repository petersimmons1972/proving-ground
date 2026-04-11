package results_test

import (
	"strings"
	"testing"

	"github.com/psimmons/proving-ground/internal/results"
)

func sampleRadarData() map[string][]float64 {
	return map[string][]float64{
		"zero":  {5.0, 3.0, 2.0, 3.5, 2.5, 4.0},
		"light": {6.5, 5.5, 5.0, 6.0, 5.0, 6.5},
		"user":  {8.5, 8.0, 8.5, 8.0, 9.0, 7.5},
	}
}

var sampleDimensions = []string{"correctness", "elegance", "discipline", "judgment", "creativity", "recovery"}
var sampleConfigs = []string{"zero", "light", "user"}
var sampleColors = map[string]string{
	"zero": "#64748b", "light": "#22D3EE", "user": "#F59E0B",
}

func TestRadarSVGProducesValidSVG(t *testing.T) {
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 400)
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("RadarSVG does not start with <svg")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("RadarSVG does not end with </svg>")
	}
}

func TestRadarSVGHasCorrectPolygonCount(t *testing.T) {
	// 5 grid rings + 3 data polygons = 8 polygons total
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 400)
	count := strings.Count(svg, "<polygon")
	if count != 8 {
		t.Errorf("polygon count = %d, want 8 (5 rings + 3 data)", count)
	}
}

func TestRadarSVGHasCorrectSpokeCount(t *testing.T) {
	// 6 dimensions = 6 spokes
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 400)
	count := strings.Count(svg, "<line")
	if count != 6 {
		t.Errorf("spoke count = %d, want 6", count)
	}
}

func TestRadarSVGContainsDimensionLabels(t *testing.T) {
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 400)
	for _, dim := range sampleDimensions {
		if !strings.Contains(svg, dim) {
			t.Errorf("dimension label %q missing from radar SVG", dim)
		}
	}
}

func TestRadarSVGContainsConfigColors(t *testing.T) {
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 400)
	for _, color := range sampleColors {
		if !strings.Contains(svg, color) {
			t.Errorf("config color %q missing from radar SVG", color)
		}
	}
}

func TestBarChartSVGProducesValidSVG(t *testing.T) {
	dimScores := map[string]map[string]float64{
		"zero":  {"correctness": 5.0, "elegance": 3.0, "discipline": 2.0, "judgment": 3.5, "creativity": 2.5, "recovery": 4.0},
		"light": {"correctness": 6.5, "elegance": 5.5, "discipline": 5.0, "judgment": 6.0, "creativity": 5.0, "recovery": 6.5},
		"user":  {"correctness": 8.5, "elegance": 8.0, "discipline": 8.5, "judgment": 8.0, "creativity": 9.0, "recovery": 7.5},
	}
	svg := results.BarChartSVG(dimScores, sampleDimensions, sampleConfigs, sampleColors, 600)
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("BarChartSVG does not start with <svg")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("BarChartSVG does not end with </svg>")
	}
}

func TestBarChartSVGContainsDimensions(t *testing.T) {
	dimScores := map[string]map[string]float64{
		"zero": {"correctness": 5.0, "elegance": 3.0, "discipline": 2.0, "judgment": 3.5, "creativity": 2.5, "recovery": 4.0},
	}
	svg := results.BarChartSVG(dimScores, sampleDimensions, []string{"zero"}, sampleColors, 600)
	for _, dim := range sampleDimensions {
		if !strings.Contains(svg, dim) {
			t.Errorf("dimension %q missing from bar chart SVG", dim)
		}
	}
}

func TestBarChartSVGBarWidthsProportional(t *testing.T) {
	// config with score 10.0 should have a bar at max width, 5.0 at half width
	dimScores := map[string]map[string]float64{
		"full": {"correctness": 10.0},
		"half": {"correctness": 5.0},
	}
	svg := results.BarChartSVG(dimScores, []string{"correctness"}, []string{"full", "half"}, sampleColors, 600)
	// Max bar width = 600 - 100 - 40 - 20 = 440
	// Full bar: width="440.0", half bar: width="220.0"
	if !strings.Contains(svg, `width="440.0"`) {
		t.Errorf("full bar width not 440.0 in: %s", svg[:500])
	}
	if !strings.Contains(svg, `width="220.0"`) {
		t.Errorf("half bar width not 220.0 in: %s", svg[:500])
	}
}

func TestSparklineSVGTwoValues(t *testing.T) {
	svg := results.SparklineSVG([]float64{5.0, 8.0}, "#D4A574", 120, 40)
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("SparklineSVG does not start with <svg")
	}
	if !strings.Contains(svg, "<polyline") {
		t.Error("SparklineSVG missing polyline element")
	}
}

func TestSparklineSVGFewerThanTwoValuesReturnsEmpty(t *testing.T) {
	if svg := results.SparklineSVG([]float64{5.0}, "#D4A574", 120, 40); svg != "" {
		t.Errorf("expected empty string for 1 value, got %q", svg)
	}
	if svg := results.SparklineSVG(nil, "#D4A574", 120, 40); svg != "" {
		t.Errorf("expected empty string for nil values, got %q", svg)
	}
}

func TestSparklineSVGEqualValuesNoDiv0(t *testing.T) {
	// All equal values: range becomes 1.0 (protected division)
	svg := results.SparklineSVG([]float64{7.0, 7.0, 7.0}, "#D4A574", 120, 40)
	if svg == "" {
		t.Error("SparklineSVG returned empty for equal values")
	}
	if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
		t.Errorf("SparklineSVG contains NaN/Inf for equal values: %s", svg)
	}
}

func TestRadarSVGDefaultSize(t *testing.T) {
	// size=0 should default to 400
	svg := results.RadarSVG(sampleRadarData(), sampleDimensions, sampleColors, sampleConfigs, 0)
	if !strings.Contains(svg, `width="400"`) {
		t.Error("default size not 400")
	}
}
