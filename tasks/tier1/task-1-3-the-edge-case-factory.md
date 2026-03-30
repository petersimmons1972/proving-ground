---
id: t1-3
tier: 1
title: "The Edge Case Factory"
---

# Task: Implement and Test

## Spec

Implement the function in `solution/calculator.py`:

```python
def safe_divide(a: float, b: float, precision: int = 2) -> float | None:
    """
    Divide a by b, rounded to precision decimal places.
    Returns None if division is not possible.
    """
```

Three basic test cases are provided in `tests/test_calculator.py`. Add additional tests for edge cases you think are missing. Run all tests to verify they pass.

Do not change the existing tests. Do not change the function signature.
