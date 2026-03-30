import json
from pathlib import Path
from src.results import generate_results, ResultsReport


def _make_sample_report() -> ResultsReport:
    return ResultsReport(
        run_id="2026-03-30T12:00:00",
        task_suite_version="v1",
        configurations=["zero", "light", "user"],
        scores={
            "zero":  {"overall": 4.2, "tier1": 5.0, "tier2": 4.0, "tier3": 3.5},
            "light": {"overall": 6.1, "tier1": 6.5, "tier2": 6.0, "tier3": 5.8},
            "user":  {"overall": 8.3, "tier1": 8.0, "tier2": 8.5, "tier3": 8.4},
        },
        dimension_scores={
            "zero":  {"correctness": 5.0, "elegance": 3.0, "discipline": 2.0, "judgment": 3.5, "creativity": 2.5, "recovery": 4.0},
            "light": {"correctness": 6.5, "elegance": 5.5, "discipline": 5.0, "judgment": 6.0, "creativity": 5.0, "recovery": 6.5},
            "user":  {"correctness": 8.5, "elegance": 8.0, "discipline": 8.5, "judgment": 8.0, "creativity": 9.0, "recovery": 7.5},
        },
    )


def test_generate_results_produces_html(tmp_path):
    report = _make_sample_report()
    html_path = tmp_path / "results.html"
    json_path = tmp_path / "results.json"
    generate_results(report, html_path, json_path)
    assert html_path.exists()
    assert json_path.exists()


def test_html_contains_proving_ground(tmp_path):
    generate_results(_make_sample_report(), tmp_path / "r.html", tmp_path / "r.json")
    html = (tmp_path / "r.html").read_text()
    assert "Proving Ground" in html


def test_html_contains_all_scores(tmp_path):
    generate_results(_make_sample_report(), tmp_path / "r.html", tmp_path / "r.json")
    html = (tmp_path / "r.html").read_text()
    assert "8.3" in html   # user overall
    assert "6.1" in html   # light overall
    assert "4.2" in html   # zero overall


def test_html_contains_all_configs(tmp_path):
    generate_results(_make_sample_report(), tmp_path / "r.html", tmp_path / "r.json")
    html = (tmp_path / "r.html").read_text()
    assert "zero" in html
    assert "light" in html
    assert "user" in html


def test_json_output_is_valid(tmp_path):
    generate_results(_make_sample_report(), tmp_path / "r.html", tmp_path / "r.json")
    data = json.loads((tmp_path / "r.json").read_text())
    assert data["run_id"] == "2026-03-30T12:00:00"
    assert "scores" in data
    assert data["scores"]["user"]["overall"] == 8.3


def test_html_is_self_contained(tmp_path):
    """HTML must not reference external CSS, fonts, or scripts."""
    generate_results(_make_sample_report(), tmp_path / "r.html", tmp_path / "r.json")
    html = (tmp_path / "r.html").read_text()
    import re
    # No external stylesheet links
    assert not re.search(r'<link[^>]+rel=["\']stylesheet["\'][^>]+href=["\']http', html)
    # No external script src
    assert not re.search(r'<script[^>]+src=["\']http', html)
    # No Google Fonts
    assert "fonts.googleapis" not in html
