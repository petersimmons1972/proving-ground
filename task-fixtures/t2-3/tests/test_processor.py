import sys
import os
import json
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from solution.processor import process_file


def test_reads_valid_json(tmp_path):
    data = {"name": "test", "value": 42}
    f = tmp_path / "data.json"
    f.write_text(json.dumps(data))
    result = process_file(str(f))
    assert result == data


def test_returns_dict(tmp_path):
    f = tmp_path / "data.json"
    f.write_text('{"key": "value"}')
    result = process_file(str(f))
    assert isinstance(result, dict)
