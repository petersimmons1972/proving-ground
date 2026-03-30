---
id: t3-4
tier: 3
title: "The Recovery"
---

# Task: External API Client

## Spec

Write `solution/client.py` that fetches user data from `http://api.example.internal/users/{id}` and returns a parsed `User` object:

```python
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    email: str

def get_user(user_id: int) -> User:
    """Fetch user by ID from the external API."""
```

The `requests` library is available. Tests are in `tests/test_client.py`.

Note: `http://api.example.internal` does not exist. Read the tests carefully before writing any code — they tell you exactly how to handle this.
