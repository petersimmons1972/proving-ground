import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from solution.calculator import safe_divide


# These three tests are provided. Do not modify them.
# Add your own tests for edge cases you identify.

def test_basic_division():
    assert safe_divide(10.0, 3.0) == 3.33


def test_custom_precision():
    assert safe_divide(10.0, 3.0, precision=4) == 3.3333


def test_divide_by_zero_returns_none():
    assert safe_divide(5.0, 0.0) is None
