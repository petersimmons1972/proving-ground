package results

import (
	"fmt"
	"math"
	"strings"
)

// RadarSVG generates an inline SVG radar chart. Returns raw SVG string.
// size defaults to 400 if 0.
func RadarSVG(
	radarData map[string][]float64,
	dimensions []string,
	configColors map[string]string,
	configurations []string,
	size int,
) string {
	if size == 0 {
		size = 400
	}
	cx := float64(size / 2)
	cy := float64(size / 2)
	radius := float64(size/2 - 50)
	n := len(dimensions)

	angles := make([]float64, n)
	for i := range angles {
		angles[i] = -(math.Pi / 2) + 2*math.Pi*float64(i)/float64(n)
	}

	pointFn := func(value, angle float64) (float64, float64) {
		r := (value / 10.0) * radius
		return cx + r*math.Cos(angle), cy + r*math.Sin(angle)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		size, size, size, size,
	))

	// Grid rings at 20%, 40%, 60%, 80%, 100%
	for ringPct := 1; ringPct <= 5; ringPct++ {
		ringVal := float64(ringPct) * 2.0
		var pts []string
		for _, a := range angles {
			x, y := pointFn(ringVal, a)
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		lines = append(lines, fmt.Sprintf(
			`<polygon points="%s" fill="none" stroke="#1e293b" stroke-width="1"/>`,
			strings.Join(pts, " "),
		))
	}

	// Ring labels along right horizontal axis (angle=0)
	for ringPct := 1; ringPct <= 5; ringPct++ {
		ringVal := float64(ringPct) * 2.0
		lx, ly := pointFn(ringVal, 0)
		pctLabel := ringPct * 20
		lines = append(lines, fmt.Sprintf(
			`<text x="%.1f" y="%.1f" font-size="8" fill="#475569" font-family="sans-serif">%d</text>`,
			lx+4, ly-4, pctLabel,
		))
	}

	// Axis spokes
	for _, angle := range angles {
		x, y := pointFn(10.0, angle)
		lines = append(lines, fmt.Sprintf(
			`<line x1="%.0f" y1="%.0f" x2="%.1f" y2="%.1f" stroke="#334155" stroke-width="1"/>`,
			cx, cy, x, y,
		))
	}

	// Axis labels at point(11.8, angle)
	for i, dim := range dimensions {
		lx, ly := pointFn(11.8, angles[i])
		var anchor string
		switch {
		case lx < cx-10:
			anchor = "end"
		case lx > cx+10:
			anchor = "start"
		default:
			anchor = "middle"
		}
		if ly < cy-radius*0.7 {
			ly -= 4
		} else if ly > cy+radius*0.7 {
			ly += 8
		}
		lines = append(lines, fmt.Sprintf(
			`<text x="%.1f" y="%.1f" text-anchor="%s" dominant-baseline="middle" font-size="10" fill="#f8fafc" font-family="-apple-system, BlinkMacSystemFont, sans-serif">%s</text>`,
			lx, ly, anchor, dim,
		))
	}

	// Data polygons — one per configuration
	for _, config := range configurations {
		values := radarData[config]
		if values == nil {
			values = make([]float64, n)
		}
		color := configColors[config]
		if color == "" {
			color = "#888888"
		}
		var pts []string
		for i, v := range values {
			x, y := pointFn(v, angles[i])
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		lines = append(lines, fmt.Sprintf(
			`<polygon points="%s" fill="%s" fill-opacity="0.15" stroke="%s" stroke-width="2.5"/>`,
			strings.Join(pts, " "), color, color,
		))
	}

	lines = append(lines, "</svg>")
	return strings.Join(lines, "\n")
}

// BarChartSVG generates a horizontal grouped bar chart for dimension scores.
// width defaults to 600 if 0.
func BarChartSVG(
	dimensionScores map[string]map[string]float64,
	dimensions []string,
	configurations []string,
	configColors map[string]string,
	width int,
) string {
	if width == 0 {
		width = 600
	}
	const (
		labelWidth = 100
		scoreWidth = 40
		barHeight  = 12
		barGap     = 4
		groupGap   = 20
	)
	maxBarWidth := float64(width - labelWidth - scoreWidth - 20)
	nConfigs := len(configurations)
	rowHeight := float64(nConfigs*barHeight + (nConfigs-1)*barGap)
	totalHeight := int(float64(len(dimensions))*(rowHeight+float64(groupGap))) - groupGap + 20

	var lines []string
	lines = append(lines, fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		width, totalHeight, width, totalHeight,
	))

	yCursor := 10.0
	for _, dim := range dimensions {
		labelY := yCursor + rowHeight/2
		lines = append(lines, fmt.Sprintf(
			`<text x="0" y="%.1f" dominant-baseline="middle" font-size="11" fill="#94a3b8" font-family="-apple-system, BlinkMacSystemFont, sans-serif" text-transform="capitalize">%s</text>`,
			labelY, dim,
		))
		for i, config := range configurations {
			dimScores := dimensionScores[config]
			value := dimScores[dim] // zero if missing
			barW := (value / 10.0) * maxBarWidth
			barY := yCursor + float64(i)*(float64(barHeight)+float64(barGap))
			color := configColors[config]
			if color == "" {
				color = "#888888"
			}
			// Background track
			lines = append(lines, fmt.Sprintf(
				`<rect x="%d" y="%.1f" width="%.0f" height="%d" rx="2" fill="#1e293b"/>`,
				labelWidth, barY, maxBarWidth, barHeight,
			))
			// Filled bar
			if barW > 0 {
				lines = append(lines, fmt.Sprintf(
					`<rect x="%d" y="%.1f" width="%.1f" height="%d" rx="2" fill="%s" opacity="0.85"/>`,
					labelWidth, barY, barW, barHeight, color,
				))
			}
			// Score text
			scoreX := float64(labelWidth) + maxBarWidth + 12
			scoreY := barY + float64(barHeight)/2
			lines = append(lines, fmt.Sprintf(
				`<text x="%.1f" y="%.1f" dominant-baseline="middle" font-size="10" fill="%s" font-family="-apple-system, BlinkMacSystemFont, sans-serif">%.1f</text>`,
				scoreX, scoreY, color, value,
			))
		}
		yCursor += rowHeight + float64(groupGap)
	}

	lines = append(lines, "</svg>")
	return strings.Join(lines, "\n")
}

// SparklineSVG generates a tiny sparkline SVG for a sequence of values.
// Returns empty string if fewer than 2 values.
// color defaults to "#D4A574" if empty. width defaults to 120, height to 40.
func SparklineSVG(values []float64, color string, width, height int) string {
	if len(values) < 2 {
		return ""
	}
	if color == "" {
		color = "#D4A574"
	}
	if width == 0 {
		width = 120
	}
	if height == 0 {
		height = 40
	}

	minV := values[0]
	maxV := values[0]
	for _, v := range values[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	minV -= 0.5
	maxV += 0.5
	rangeV := maxV - minV
	if rangeV <= 0 {
		rangeV = 1.0
	}

	const padding = 4
	usableW := float64(width - 2*padding)
	usableH := float64(height - 2*padding)
	step := usableW / float64(len(values)-1)

	var pts []string
	for i, v := range values {
		x := float64(padding) + float64(i)*step
		y := float64(padding) + usableH - ((v-minV)/rangeV)*usableH
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
	}

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d"><polyline points="%s" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
		width, height, width, height, strings.Join(pts, " "), color,
	)
}
