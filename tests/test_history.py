import json
from pathlib import Path
from src.history import append_run, load_history, HistoryEntry


def test_append_creates_history_file(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    entry = HistoryEntry(run_id="2026-03-30T12:00:00", scores={"zero": 4.2, "user": 8.3})
    append_run(hist_file, entry)
    assert hist_file.exists()


def test_append_writes_valid_json(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    entry = HistoryEntry(run_id="2026-03-30T12:00:00", scores={"user": 8.3})
    append_run(hist_file, entry)
    lines = hist_file.read_text().strip().splitlines()
    assert len(lines) == 1
    data = json.loads(lines[0])
    assert data["run_id"] == "2026-03-30T12:00:00"
    assert data["scores"]["user"] == 8.3


def test_append_multiple_runs(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    for i in range(3):
        append_run(hist_file, HistoryEntry(run_id=f"run-{i}", scores={"user": float(i + 5)}))
    entries = load_history(hist_file)
    assert len(entries) == 3


def test_load_history_sorted_by_run_id(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    for run_id in ["2026-03-31", "2026-03-29", "2026-03-30"]:
        append_run(hist_file, HistoryEntry(run_id=run_id, scores={"user": 7.0}))
    entries = load_history(hist_file)
    assert entries[0].run_id == "2026-03-29"
    assert entries[1].run_id == "2026-03-30"
    assert entries[-1].run_id == "2026-03-31"


def test_load_history_empty_file_returns_empty_list(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    assert load_history(hist_file) == []


def test_load_history_missing_file_returns_empty_list(tmp_path):
    hist_file = tmp_path / "nonexistent.jsonl"
    assert load_history(hist_file) == []


def test_history_entry_preserves_all_scores(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    scores = {"zero": 4.2, "light": 6.1, "user": 8.3}
    append_run(hist_file, HistoryEntry(run_id="2026-03-30", scores=scores))
    entries = load_history(hist_file)
    assert entries[0].scores == scores
