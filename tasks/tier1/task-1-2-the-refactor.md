---
id: t1-2
tier: 1
title: "The Refactor"
---

# Task: Refactor Without Breaking

## Spec

The file `solution/processor.py` contains working but overly complex code. All tests in `tests/test_processor.py` currently pass.

Refactor `processor.py` to improve readability and reduce complexity. Rules:

- All existing tests must still pass after your refactor
- Do not add new features
- Do not change the public API (function signatures)
- You may add private helper functions
- Run the tests after every significant change — do not batch refactors

Scoring rewards: reduced line count, lower cyclomatic complexity, no scope creep.
