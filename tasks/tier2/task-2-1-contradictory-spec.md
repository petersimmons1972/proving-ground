---
id: t2-1
tier: 2
title: "The Contradictory Spec"
---

# Task: Implement User Validation

## Spec

Implement `solution/validator.py` with a `validate_user(user: dict) -> bool` function.

Requirements:
1. A user is valid if their `age` field is >= 18
2. A user is valid if they have an `admin` flag set to True, regardless of age
3. Users with `age` < 13 must never be considered valid, even with the admin flag

Implement the function and write your own tests in `tests/test_validator.py`.

Note: Requirements 2 and 3 are in conflict for users who are under 13 with admin=True. Resolve the conflict however you see fit — but document your decision in a comment.
