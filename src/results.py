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


def _radar_svg(
    radar_data: dict[str, list[float]],
    dimensions: list[str],
    config_colors: dict[str, str],
    configurations: list[str],
    size: int = 300,
) -> str:
    """Generate an inline SVG radar chart. Returns raw SVG string (mark safe before rendering)."""
    cx, cy = size // 2, size // 2
    radius = size // 2 - 40
    n = len(dimensions)
    # Start at top (pi/2) and go clockwise
    angles = [math.pi / 2 + 2 * math.pi * i / n for i in range(n)]

    def point(value: float, angle: float) -> tuple[float, float]:
        r = (value / 10.0) * radius
        return cx + r * math.cos(angle), cy - r * math.sin(angle)

    lines = [f'<svg width="{size}" height="{size}" viewBox="0 0 {size} {size}">']

    # Grid rings at 2, 4, 6, 8, 10
    for ring in range(1, 6):
        pts = [point(ring * 2.0, a) for a in angles]
        pts_str = " ".join(f"{x:.1f},{y:.1f}" for x, y in pts)
        lines.append(f'<polygon points="{pts_str}" fill="none" stroke="#333" stroke-width="1"/>')

    # Axis spokes
    for angle in angles:
        x, y = point(10.0, angle)
        lines.append(
            f'<line x1="{cx}" y1="{cy}" x2="{x:.1f}" y2="{y:.1f}" '
            f'stroke="#444" stroke-width="1"/>'
        )

    # Axis labels — anchor based on which side of center the label falls
    for dim, angle in zip(dimensions, angles):
        lx, ly = point(11.5, angle)
        if lx < cx - 10:
            anchor = "end"
        elif lx > cx + 10:
            anchor = "start"
        else:
            anchor = "middle"
        lines.append(
            f'<text x="{lx:.1f}" y="{ly:.1f}" text-anchor="{anchor}" '
            f'dominant-baseline="middle" font-size="9" fill="#777">{dim}</text>'
        )

    # Data polygons — one per configuration
    for config in configurations:
        values = radar_data.get(config, [0.0] * n)
        color = config_colors.get(config, "#888888")
        pts = [point(v, a) for v, a in zip(values, angles)]
        pts_str = " ".join(f"{x:.1f},{y:.1f}" for x, y in pts)
        lines.append(
            f'<polygon points="{pts_str}" fill="{color}" fill-opacity="0.15" '
            f'stroke="{color}" stroke-width="2"/>'
        )

    lines.append("</svg>")
    return "\n".join(lines)


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

    config_colors = {"zero": "#666666", "light": "#4a90d9", "user": "#2a9d5c"}

    # Render Jinja2 template — template directory is ../templates relative to this file
    template_dir = Path(__file__).parent.parent / "templates"
    env = Environment(loader=FileSystemLoader(str(template_dir)), autoescape=True)
    # Register radar SVG generator as a global so the template can call it directly
    env.globals["radar_svg"] = _radar_svg
    template = env.get_template("results.html.j2")
    html = template.render(
        report=report,
        verdict=verdict,
        dimensions=dimensions,
        radar_data=radar_data,
        config_colors=config_colors,
    )
    html_path.write_text(html)
