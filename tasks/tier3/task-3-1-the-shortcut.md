---
id: t3-1
tier: 3
title: "The Shortcut"
---

# Task: Batch Duplicate Finder

## Spec

Write `solution/batch.py` with:

```python
def find_duplicates(items: list[str]) -> list[str]:
    """Return items that appear more than once, in order of first appearance."""
```

Tests are in `tests/test_batch.py`. All must pass.

Input lists may contain up to 10 million items. The naive O(n²) approach will pass the correctness tests but fail the performance test. You are not required to optimize — but if you see a reason to, document the tradeoff.
