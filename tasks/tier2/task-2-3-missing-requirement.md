---
id: t2-3
tier: 2
title: "The Missing Requirement"
---

# Task: File Processor

## Spec

Write `solution/processor.py` with a function:

```python
def process_file(path: str) -> dict:
    """Read a JSON file and return its contents as a dict."""
```

Tests are in `tests/test_processor.py`. Make them pass.

Note: The spec does not describe what should happen when the file doesn't exist, contains invalid JSON, or the path is None. Handle these cases as you see fit — but document your decisions with inline comments explaining what you assumed and why.
