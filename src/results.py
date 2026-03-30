"""
results.py — Generate HTML and JSON output from a ResultsReport.

Takes scored data, renders a self-contained HTML page and a machine-readable
JSON file. No external dependencies in the HTML output — inline CSS, inline SVG.
"""

import json
import math
from dataclasses import dataclass, asdict
from pathlib import Path

from jinja2 import Environment, FileSystemLoader


@dataclass
class ResultsReport:
    run_id: str
    task_suite_version: str
    configurations: list[str]
    scores: dict[str, dict[str, float]]            # config -> {overall, tier1, tier2, tier3}
    dimension_scores: dict[str, dict[str, float]]  # config -> {correctness, elegance, ...}
    history: list | None = None                     # optional list of previous run summaries


# Default color palette: dim for zero-context, muted blue for light, gold for user
DEFAULT_CONFIG_COLORS = {
    "zero": "#475569",
    "light": "#64748b",
    "user": "#D4A574",
}


def _radar_svg(
    radar_data: dict[str, list[float]],
    dimensions: list[str],
    config_colors: dict[str, str],
    configurations: list[str],
    size: int = 400,
) -> str:
    """Generate an inline SVG radar chart. Returns raw SVG string (mark safe before rendering)."""
    cx, cy = size // 2, size // 2
    radius = size // 2 - 50  # extra padding for labels
    n = len(dimensions)
    # Start at top (-pi/2) and go clockwise
    angles = [-(math.pi / 2) + 2 * math.pi * i / n for i in range(n)]

    def point(value: float, angle: float) -> tuple[float, float]:
        r = (value / 10.0) * radius
        return cx + r * math.cos(angle), cy + r * math.sin(angle)

    lines = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{size}" height="{size}" '
        f'viewBox="0 0 {size} {size}">'
    ]

    # Grid rings at 20%, 40%, 60%, 80%, 100%
    for ring_pct in range(1, 6):
        ring_val = ring_pct * 2.0  # 2, 4, 6, 8, 10
        pts = [point(ring_val, a) for a in angles]
        pts_str = " ".join(f"{x:.1f},{y:.1f}" for x, y in pts)
        lines.append(
            f'<polygon points="{pts_str}" fill="none" '
            f'stroke="#1e293b" stroke-width="1"/>'
        )

    # Ring percentage labels along the rightmost axis (angle index 0, which is top)
    # Place them along the right horizontal axis for readability
    right_angle = 0  # 0 radians = rightward
    for ring_pct in range(1, 6):
        ring_val = ring_pct * 2.0
        lx, ly = point(ring_val, right_angle)
        pct_label = ring_pct * 20
        lines.append(
            f'<text x="{lx + 4:.1f}" y="{ly - 4:.1f}" '
            f'font-size="8" fill="#475569" font-family="sans-serif">'
            f'{pct_label}</text>'
        )

    # Axis spokes
    for angle in angles:
        x, y = point(10.0, angle)
        lines.append(
            f'<line x1="{cx}" y1="{cy}" x2="{x:.1f}" y2="{y:.1f}" '
            f'stroke="#334155" stroke-width="1"/>'
        )

    # Axis labels — positioned outside the chart boundary
    for dim, angle in zip(dimensions, angles):
        lx, ly = point(11.8, angle)
        if lx < cx - 10:
            anchor = "end"
        elif lx > cx + 10:
            anchor = "start"
        else:
            anchor = "middle"
        # Adjust vertical position for top/bottom labels
        if ly < cy - radius * 0.7:
            ly -= 4
        elif ly > cy + radius * 0.7:
            ly += 8
        lines.append(
            f'<text x="{lx:.1f}" y="{ly:.1f}" text-anchor="{anchor}" '
            f'dominant-baseline="middle" font-size="10" fill="#f8fafc" '
            f'font-family="-apple-system, BlinkMacSystemFont, sans-serif">'
            f'{dim}</text>'
        )

    # Data polygons — one per configuration
    for config in configurations:
        values = radar_data.get(config, [0.0] * n)
        color = config_colors.get(config, "#888888")
        pts = [point(v, a) for v, a in zip(values, angles)]
        pts_str = " ".join(f"{x:.1f},{y:.1f}" for x, y in pts)
        lines.append(
            f'<polygon points="{pts_str}" fill="{color}" fill-opacity="0.15" '
            f'stroke="{color}" stroke-width="2.5"/>'
        )

    lines.append("</svg>")
    return "\n".join(lines)


def _bar_chart_svg(
    dimension_scores: dict[str, dict[str, float]],
    dimensions: list[str],
    configurations: list[str],
    config_colors: dict[str, str],
    width: int = 600,
) -> str:
    """Horizontal grouped bar chart for dimension scores.

    Each dimension is a row. Within each row, one bar per configuration.
    Bar width is proportional to value/10.
    """
    label_width = 100
    score_width = 40
    max_bar_width = width - label_width - score_width - 20
    bar_height = 12
    bar_gap = 4
    group_gap = 20
    n_configs = len(configurations)

    # Calculate total height
    row_height = n_configs * bar_height + (n_configs - 1) * bar_gap
    total_height = len(dimensions) * (row_height + group_gap) - group_gap + 20

    lines = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" '
        f'height="{total_height}" viewBox="0 0 {width} {total_height}">'
    ]

    y_cursor = 10
    for dim in dimensions:
        # Dimension label — left aligned, vertically centered in the group
        label_y = y_cursor + row_height / 2
        lines.append(
            f'<text x="0" y="{label_y:.1f}" dominant-baseline="middle" '
            f'font-size="11" fill="#94a3b8" '
            f'font-family="-apple-system, BlinkMacSystemFont, sans-serif" '
            f'text-transform="capitalize">{dim}</text>'
        )

        for i, config in enumerate(configurations):
            value = dimension_scores.get(config, {}).get(dim, 0.0)
            bar_w = (value / 10.0) * max_bar_width
            bar_y = y_cursor + i * (bar_height + bar_gap)
            color = config_colors.get(config, "#888888")

            # Bar background track
            lines.append(
                f'<rect x="{label_width}" y="{bar_y:.1f}" '
                f'width="{max_bar_width}" height="{bar_height}" '
                f'rx="2" fill="#1e293b"/>'
            )

            # Filled bar
            if bar_w > 0:
                lines.append(
                    f'<rect x="{label_width}" y="{bar_y:.1f}" '
                    f'width="{bar_w:.1f}" height="{bar_height}" '
                    f'rx="2" fill="{color}" opacity="0.85"/>'
                )

            # Score value — right aligned
            score_x = label_width + max_bar_width + 12
            score_y = bar_y + bar_height / 2
            lines.append(
                f'<text x="{score_x:.1f}" y="{score_y:.1f}" '
                f'dominant-baseline="middle" font-size="10" fill="{color}" '
                f'font-family="-apple-system, BlinkMacSystemFont, sans-serif">'
                f'{value:.1f}</text>'
            )

        y_cursor += row_height + group_gap

    lines.append("</svg>")
    return "\n".join(lines)


def _history_sparkline_svg(
    values: list[float],
    color: str = "#D4A574",
    width: int = 120,
    height: int = 40,
) -> str:
    """Tiny inline sparkline SVG for a sequence of score values."""
    if len(values) < 2:
        return ""
    min_v = min(values) - 0.5
    max_v = max(values) + 0.5
    range_v = max_v - min_v if max_v > min_v else 1.0
    padding = 4
    usable_w = width - 2 * padding
    usable_h = height - 2 * padding
    step = usable_w / (len(values) - 1)

    points = []
    for i, v in enumerate(values):
        x = padding + i * step
        y = padding + usable_h - ((v - min_v) / range_v) * usable_h
        points.append(f"{x:.1f},{y:.1f}")

    pts_str = " ".join(points)
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" '
        f'viewBox="0 0 {width} {height}">'
        f'<polyline points="{pts_str}" fill="none" stroke="{color}" '
        f'stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>'
        f'</svg>'
    )


def generate_results(
    report: ResultsReport,
    html_path: Path,
    json_path: Path,
) -> None:
    """Generate results.html and results.json from a ResultsReport."""

    # Write JSON — straightforward serialisation of the dataclass
    json_path.write_text(json.dumps(asdict(report), indent=2))

    # Build verdict sentence comparing best config to zero-context baseline
    configs = report.configurations
    if len(configs) >= 2:
        best = max(configs, key=lambda c: report.scores[c]["overall"])
        baseline = "zero"
        if baseline in report.scores and best != baseline:
            delta = report.scores[best]["overall"] - report.scores[baseline]["overall"]
            verdict = (
                f"Your {best} agent outperformed the baseline by "
                f"{delta:.1f} points overall."
            )
        else:
            verdict = f"Best configuration: {best}."
    else:
        verdict = "Run complete."

    # Build dimension data for the radar chart
    dimensions = ["correctness", "elegance", "discipline", "judgment", "creativity", "recovery"]
    radar_data = {
        config: [report.dimension_scores[config].get(d, 0.0) for d in dimensions]
        for config in configs
    }

    config_colors = dict(DEFAULT_CONFIG_COLORS)

    # Render Jinja2 template — template directory is ../templates relative to this file
    template_dir = Path(__file__).parent.parent / "templates"
    env = Environment(loader=FileSystemLoader(str(template_dir)), autoescape=True)
    # Register SVG generators as globals so the template can call them directly
    env.globals["radar_svg"] = _radar_svg
    env.globals["bar_chart_svg"] = _bar_chart_svg
    env.globals["sparkline_svg"] = _history_sparkline_svg
    template = env.get_template("results.html.j2")
    html = template.render(
        report=report,
        verdict=verdict,
        dimensions=dimensions,
        radar_data=radar_data,
        config_colors=config_colors,
    )
    html_path.write_text(html)
