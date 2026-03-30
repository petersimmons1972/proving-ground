import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from solution.parser import parse_logs


def test_parses_json_lines():
    results = parse_logs("input/logs.txt")
    json_results = [r for r in results if r.get("ts") in (1711800000, 1711800001)]
    assert len(json_results) == 2


def test_parses_csv_lines():
    results = parse_logs("input/logs.txt")
    csv_results = [r for r in results if r.get("ts") in (1711800002, 1711800003)]
    assert len(csv_results) == 2


def test_skips_malformed():
    results = parse_logs("input/logs.txt")
    assert all("ts" in r and isinstance(r["ts"], int) for r in results)
    assert len(results) == 8  # exactly 8 valid lines


def test_normalizes_level_to_lowercase():
    results = parse_logs("input/logs.txt")
    assert all(r["level"] == r["level"].lower() for r in results)


def test_preserves_unicode():
    results = parse_logs("input/logs.txt")
    unicode_result = next((r for r in results if "情報" in r.get("msg", "")), None)
    assert unicode_result is not None


def test_returns_list_of_dicts():
    results = parse_logs("input/logs.txt")
    assert isinstance(results, list)
    assert all(isinstance(r, dict) for r in results)
    assert all({"level", "msg", "ts"}.issubset(r.keys()) for r in results)
