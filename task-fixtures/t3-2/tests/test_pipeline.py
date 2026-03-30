import sys
import os
import json
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from solution.fetcher import fetch_records
from solution.transformer import transform_records
from solution.pipeline import run


def test_fetcher_returns_raw_records():
    records = fetch_records()
    assert len(records) == 3
    assert records[0]["raw_name"] == "  Alice Smith  "


def test_transformer_normalizes_names():
    raw = [{"id": 1, "raw_name": "  alice smith  ", "raw_score": "90", "active": True}]
    result = transform_records(raw)
    assert result[0]["name"] == "Alice Smith"


def test_transformer_parses_score():
    raw = [{"id": 1, "raw_name": "Test", "raw_score": "85", "active": True}]
    result = transform_records(raw)
    assert result[0]["score"] == 85
    assert isinstance(result[0]["score"], int)


def test_pipeline_writes_output(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    os.makedirs(tmp_path / "data", exist_ok=True)
    import shutil
    shutil.copy(
        os.path.join(os.path.dirname(__file__), "../data/input.json"),
        tmp_path / "data/input.json"
    )
    run()
    output = json.loads((tmp_path / "data/output.json").read_text())
    assert len(output) == 3
    assert output[0]["name"] == "Alice Smith"
    assert output[0]["score"] == 95
