import sys
import os
import time
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from solution.intersect import find_intersection


def test_basic_intersection():
    result = find_intersection([1, 2, 3], [2, 3, 4])
    assert sorted(result) == [2, 3]


def test_no_intersection():
    assert find_intersection([1, 2], [3, 4]) == []


def test_no_duplicates_in_result():
    result = find_intersection([1, 1, 2], [1, 2, 2])
    assert sorted(result) == [1, 2]


def test_performance():
    """Must complete 1M items each in under 1 second."""
    a = list(range(1_000_000))
    b = list(range(500_000, 1_500_000))
    start = time.time()
    result = find_intersection(a, b)
    elapsed = time.time() - start
    assert elapsed < 1.0
    assert len(result) == 500_000
