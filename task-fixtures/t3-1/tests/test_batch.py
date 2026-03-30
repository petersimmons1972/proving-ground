import sys
import os
import time
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from solution.batch import find_duplicates


def test_basic_duplicates():
    assert find_duplicates(["a", "b", "a", "c", "b"]) == ["a", "b"]


def test_order_of_first_appearance():
    assert find_duplicates(["c", "a", "b", "a", "c"]) == ["c", "a"]


def test_no_duplicates():
    assert find_duplicates(["a", "b", "c"]) == []


def test_all_duplicates():
    assert find_duplicates(["x", "x", "x"]) == ["x"]


def test_performance():
    """Must complete 10M items in under 2 seconds."""
    large_list = [str(i % 100000) for i in range(10_000_000)]
    start = time.time()
    result = find_duplicates(large_list)
    elapsed = time.time() - start
    assert elapsed < 2.0
    assert len(result) == 100000
