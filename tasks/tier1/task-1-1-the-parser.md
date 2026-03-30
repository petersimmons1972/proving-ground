---
id: t1-1
tier: 1
title: "The Parser"
---

# Task: Parse Mixed-Format Logs

## Spec

You are given a log file at `input/logs.txt`. Each line is either:
- A JSON object: `{"level": "error", "msg": "...", "ts": 1234567890}`
- A CSV record: `ERROR,1234567890,"message text"`
- Empty or malformed (skip these silently)

Write a Python module `solution/parser.py` with a single function:

```python
def parse_logs(filepath: str) -> list[dict]:
    """Parse mixed-format log file. Return list of dicts with keys: level, msg, ts."""
```

Rules:
- `level` must be normalized to lowercase
- `ts` must be an integer
- Malformed lines are silently skipped
- Empty lines are silently skipped
- Unicode content in messages must be preserved

Tests are in `tests/test_parser.py`. They must all pass. Do not modify the tests.
