import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from solution.processor import process_data


def test_filter_strings():
    result = process_data(["hello", "world", "help"], "hel", "upper", "list")
    assert "HELLO" in result
    assert "HELP" in result
    assert "WORLD" not in result


def test_filter_numbers():
    result = process_data([1, 2, 3, 2], 2, "double", "list")
    assert result == [4, 4]


def test_output_format_count():
    result = process_data(["a", "b", "c"], None, "upper", "count")
    assert result == 3


def test_output_format_first():
    result = process_data(["x", "y"], None, "upper", "first")
    assert result == "X"


def test_output_format_last():
    result = process_data(["x", "y"], None, "lower", "last")
    assert result == "y"


def test_none_items_skipped():
    result = process_data([None, "a", None, "b"], None, "upper", "list")
    assert result == ["A", "B"]


def test_nested_list_items():
    result = process_data([["a", "b"], ["c"]], None, "upper", "list")
    assert set(result) == {"A", "B", "C"}


def test_dict_items_extract_value():
    result = process_data([{"value": "hello"}, {"value": "world"}], None, "upper", "list")
    assert result == ["HELLO", "WORLD"]
