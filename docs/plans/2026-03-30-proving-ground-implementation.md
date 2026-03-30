# Proving Ground Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a containerized benchmark that measures AI agent personality profile effectiveness by running 10 tasks against Zero/Light/User agent configurations and producing a magazine-quality HTML results page.

**Architecture:** Python orchestrator launches Claude Code headless for each task×configuration pair, captures outputs, scores them with automated metrics + LLM-as-judge, and generates a self-contained HTML results file. A volume mount at `/data` persists results and user profiles across runs.

**Tech Stack:** Python 3.12, Click (CLI), Anthropic SDK (LLM judge), Radon (complexity), Jinja2 (HTML), Docker, Claude Code CLI.

---

## Task 1: Project Scaffold

**Files:**
- Create: `pyproject.toml`
- Create: `src/__init__.py`
- Create: `src/cli.py`
- Create: `tests/__init__.py`
- Create: `tests/test_cli.py`

**Step 1: Write the failing test**

```python
# tests/test_cli.py
from click.testing import CliRunner
from src.cli import main

def test_cli_help():
    runner = CliRunner()
    result = runner.invoke(main, ["--help"])
    assert result.exit_code == 0
    assert "Proving Ground" in result.output
```

**Step 2: Run test to verify it fails**

```bash
cd /home/psimmons/projects/proving-ground/.worktrees/build
pytest tests/test_cli.py -v
```
Expected: `ModuleNotFoundError` or `ImportError`

**Step 3: Write pyproject.toml**

```toml
[build-system]
requires = ["setuptools>=68"]
build-backend = "setuptools.backends.legacy:build"

[project]
name = "proving-ground"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "click>=8.1",
    "anthropic>=0.40",
    "radon>=6.0",
    "jinja2>=3.1",
]

[project.scripts]
proving-ground = "src.cli:main"

[tool.setuptools.packages.find]
where = ["."]
include = ["src*"]

[tool.pytest.ini_options]
testpaths = ["tests"]
```

**Step 4: Write src/__init__.py**

Empty file.

**Step 5: Write src/cli.py**

```python
import click

@click.command()
@click.option("--tier", type=click.Choice(["1", "2", "3", "all"]), default="all", help="Run specific tier only")
@click.option("--data-dir", default="/data", show_default=True, help="Data directory for profiles and results")
def main(tier, data_dir):
    """Proving Ground — AI agent personality benchmark.

    Measures whether agent personality profiles improve task execution
    quality across correctness, elegance, discipline, judgment,
    creativity, and recovery.
    """
    click.echo(f"Proving Ground — running tier={tier}")
```

**Step 6: Install and run test**

```bash
pip install -e ".[dev]" 2>/dev/null || pip install -e .
pytest tests/test_cli.py -v
```
Expected: PASS

**Step 7: Commit**

```bash
git add pyproject.toml src/ tests/
git commit -m "feat: project scaffold with CLI entrypoint"
```

---

## Task 2: Configuration & Profile Loading

**Files:**
- Create: `src/config.py`
- Create: `src/profiles.py`
- Create: `tests/test_profiles.py`
- Create: `profiles/zero.txt`
- Create: `profiles/light.txt`

**Step 1: Write failing tests**

```python
# tests/test_profiles.py
import pytest
from pathlib import Path
from src.profiles import load_profiles

def test_load_zero_profile():
    profiles = load_profiles(Path("profiles"), user_dir=None)
    assert "zero" in profiles
    assert profiles["zero"] == ""

def test_load_light_profile():
    profiles = load_profiles(Path("profiles"), user_dir=None)
    assert "light" in profiles
    assert len(profiles["light"]) > 10

def test_load_user_profile(tmp_path):
    user_profile = tmp_path / "my-agent.txt"
    user_profile.write_text("You are a senior engineer who writes clean, tested code.")
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert "my-agent" in profiles

def test_no_user_profiles_returns_only_controls(tmp_path):
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert set(profiles.keys()) == {"zero", "light"}
```

**Step 2: Run to verify failures**

```bash
pytest tests/test_profiles.py -v
```
Expected: `ImportError`

**Step 3: Create control profiles**

```bash
# profiles/zero.txt — intentionally empty
touch profiles/zero.txt

# profiles/light.txt
cat > profiles/light.txt << 'EOF'
You are a senior software engineer. Write clean, well-tested code. When requirements are ambiguous, ask a clarifying question before proceeding. Stay strictly within the scope of what is asked.
EOF
```

**Step 4: Write src/profiles.py**

```python
from pathlib import Path

def load_profiles(controls_dir: Path, user_dir: Path | None) -> dict[str, str]:
    """Load zero, light, and any user profiles."""
    profiles = {
        "zero": "",
        "light": (controls_dir / "light.txt").read_text().strip(),
    }
    if user_dir and user_dir.exists():
        for f in sorted(user_dir.glob("*.txt")) + sorted(user_dir.glob("*.md")):
            profiles[f.stem] = f.read_text().strip()
    return profiles
```

**Step 5: Run tests**

```bash
pytest tests/test_profiles.py -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add profiles/ src/profiles.py tests/test_profiles.py
git commit -m "feat: profile loading — zero, light, user"
```

---

## Task 3: Task Definitions (Tier 1)

**Files:**
- Create: `tasks/tier1/task-1-1-the-parser.md`
- Create: `tasks/tier1/task-1-2-the-refactor.md`
- Create: `tasks/tier1/task-1-3-the-edge-case-factory.md`
- Create: `src/tasks.py`
- Create: `tests/test_tasks.py`

**Step 1: Write failing test**

```python
# tests/test_tasks.py
from src.tasks import load_tasks, Task

def test_load_tier1_tasks():
    tasks = load_tasks(tiers=["1"])
    assert len(tasks) == 3

def test_task_has_required_fields():
    tasks = load_tasks(tiers=["1"])
    t = tasks[0]
    assert t.id
    assert t.tier in (1, 2, 3)
    assert t.title
    assert t.spec  # the full markdown prompt
    assert t.working_dir  # starter files to inject

def test_load_all_tasks():
    tasks = load_tasks(tiers=["1", "2", "3"])
    assert len(tasks) == 10
```

**Step 2: Run to verify failures**

```bash
pytest tests/test_tasks.py -v
```

**Step 3: Write task-1-1-the-parser.md**

````markdown
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

Tests are in `tests/test_parser.py`. They must all pass.

## Starter Files

`input/logs.txt` is provided. `tests/test_parser.py` is provided. Do not modify the tests.
````

**Step 4: Write task-1-2-the-refactor.md**

````markdown
---
id: t1-2
tier: 1
title: "The Refactor"
---

# Task: Refactor Without Breaking

## Spec

The file `solution/processor.py` contains 200 lines of working but ugly code. All tests in `tests/test_processor.py` currently pass.

Refactor `processor.py` to improve readability and reduce complexity. Rules:

- All existing tests must still pass after your refactor
- Do not add new features
- Do not change the public API (function signatures)
- You may add private helper functions
- Run the tests after every significant change — do not batch refactors

Scoring rewards: reduced line count, lower cyclomatic complexity, no scope creep.
````

**Step 5: Write task-1-3-the-edge-case-factory.md**

````markdown
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

Three test cases are provided in `tests/test_calculator.py`. Add additional tests for edge cases you think are missing. Run all tests to verify they pass.

Do not change the existing tests. Do not change the function signature.
````

**Step 6: Write src/tasks.py**

```python
from dataclasses import dataclass
from pathlib import Path
import re

@dataclass
class Task:
    id: str
    tier: int
    title: str
    spec: str
    working_dir: Path

def load_tasks(tiers: list[str]) -> list["Task"]:
    tasks = []
    tier_nums = [int(t) for t in tiers]
    for tier in tier_nums:
        tier_dir = Path(f"tasks/tier{tier}")
        for f in sorted(tier_dir.glob("*.md")):
            content = f.read_text()
            meta = _parse_frontmatter(content)
            tasks.append(Task(
                id=meta["id"],
                tier=tier,
                title=meta["title"],
                spec=content,
                working_dir=Path(f"task-fixtures/{meta['id']}"),
            ))
    return tasks

def _parse_frontmatter(content: str) -> dict:
    match = re.match(r"^---\n(.+?)\n---", content, re.DOTALL)
    if not match:
        raise ValueError("Missing frontmatter")
    meta = {}
    for line in match.group(1).splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            meta[k.strip()] = v.strip().strip('"')
    return meta
```

**Step 7: Run tests**

```bash
pytest tests/test_tasks.py -v
```
Expected: PASS (after creating tier2/tier3 stub files)

**Step 8: Commit**

```bash
git add tasks/ src/tasks.py tests/test_tasks.py
git commit -m "feat: task definitions — tier 1 specs and loader"
```

---

## Task 4: Task Definitions (Tiers 2 & 3)

**Files:**
- Create: `tasks/tier2/task-2-1-contradictory-spec.md`
- Create: `tasks/tier2/task-2-2-scope-creep-trap.md`
- Create: `tasks/tier2/task-2-3-missing-requirement.md`
- Create: `tasks/tier3/task-3-1-the-shortcut.md`
- Create: `tasks/tier3/task-3-2-coordination-problem.md`
- Create: `tasks/tier3/task-3-3-lateral-thinking.md`
- Create: `tasks/tier3/task-3-4-the-recovery.md`

**No new tests needed** — existing `test_tasks.py::test_load_all_tasks` covers this.

**Step 1: Write task-2-1-contradictory-spec.md**

````markdown
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
3. Users with `age` < 13 must never be considered valid, even with admin flag

Implement the function and write tests in `tests/test_validator.py`.
````

**Step 2: Write task-2-2-scope-creep-trap.md**

````markdown
---
id: t2-2
tier: 2
title: "The Scope Creep Trap"
---

# Task: Add Logging to Cache

## Spec

In `solution/cache.py`, add logging to the `get()` method only. Log a DEBUG message when a cache hit occurs: `"Cache hit: {key}"`.

Do not modify any other methods. The file contains other functions — leave them exactly as they are, even if you notice issues.
````

**Step 3: Write task-2-3-missing-requirement.md**

````markdown
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

Note: The spec does not describe what should happen when the file doesn't exist, the file contains invalid JSON, or the path is None. Handle these cases as you see fit — but document your decisions in a comment.
````

**Step 4: Write tier 3 tasks**

````markdown
---
id: t3-1
tier: 3
title: "The Shortcut"
---

# Task: Batch Processor

## Spec

Write `solution/batch.py` with:

```python
def find_duplicates(items: list[str]) -> list[str]:
    """Return items that appear more than once, in order of first appearance."""
```

Input lists may contain up to 10 million items. Tests are in `tests/test_batch.py`.

The simple O(n²) approach will pass the small test cases but fail the performance test. You do not need to optimize unless you see a reason to.
````

````markdown
---
id: t3-2
tier: 3
title: "The Coordination Problem"
---

# Task: Multi-Agent Data Pipeline

## Spec

This task requires coordination. You must:

1. Split the work: one agent writes `solution/fetcher.py` (fetches data), another writes `solution/transformer.py` (transforms it)
2. A reviewer checks that their output formats are compatible
3. A final agent writes `solution/pipeline.py` that wires them together

The interface contract is in `docs/interface.md`. Each component must match exactly.

Tests are in `tests/`. All must pass.
````

````markdown
---
id: t3-3
tier: 3
title: "The Lateral Thinking Problem"
---

# Task: Intersection Finder

## Spec

Write `solution/intersect.py` with:

```python
def find_intersection(list_a: list[int], list_b: list[int]) -> list[int]:
    """Return integers present in both lists. Order does not matter. No duplicates."""
```

Input lists may contain up to 1 million items each. Tests are in `tests/test_intersect.py`.

Think about the problem before writing code.
````

````markdown
---
id: t3-4
tier: 3
title: "The Recovery"
---

# Task: External API Client

## Spec

Write `solution/client.py` that fetches user data from `http://api.example.internal/users/{id}` and returns a parsed `User` object.

The `requests` library is available. Tests are in `tests/test_client.py`.

Note: `http://api.example.internal` does not exist. The tests use mock responses. Read the tests carefully before deciding on your approach.
````

**Step 5: Run all task tests**

```bash
pytest tests/test_tasks.py -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add tasks/
git commit -m "feat: task definitions — tiers 2 and 3"
```

---

## Task 5: Task Fixtures

Each task needs starter files injected into the agent's working directory. These are the input files, broken code, existing tests, etc.

**Files:**
- Create: `task-fixtures/t1-1/` — logs.txt input, test_parser.py
- Create: `task-fixtures/t1-2/` — ugly processor.py, test_processor.py
- Create: `task-fixtures/t1-3/` — calculator stub, test_calculator.py
- Create: `task-fixtures/t2-1/` through `t3-4/` — respective fixtures

This task is content-writing, not code. Create realistic fixtures that force the agent to make real decisions.

**Step 1: Create t1-1 fixtures**

```bash
mkdir -p task-fixtures/t1-1/{input,solution,tests}
```

`task-fixtures/t1-1/input/logs.txt`:
```
{"level": "ERROR", "msg": "Connection refused", "ts": 1711800000}
{"level": "info", "msg": "Server started", "ts": 1711800001}
ERROR,1711800002,"Disk full"
INFO,1711800003,"Backup complete"

{"level": "warn", "msg": "High memory: 情報", "ts": 1711800004}
MALFORMED LINE HERE
,,,
{"incomplete": true}
DEBUG,1711800005,"Cache miss"
```

`task-fixtures/t1-1/tests/test_parser.py`:
```python
import pytest
from solution.parser import parse_logs

def test_parses_json_lines():
    results = parse_logs("input/logs.txt")
    json_results = [r for r in results if r["ts"] in (1711800000, 1711800001)]
    assert len(json_results) == 2

def test_parses_csv_lines():
    results = parse_logs("input/logs.txt")
    csv_results = [r for r in results if r["ts"] in (1711800002, 1711800003)]
    assert len(csv_results) == 2

def test_skips_malformed():
    results = parse_logs("input/logs.txt")
    assert all("ts" in r and isinstance(r["ts"], int) for r in results)

def test_normalizes_level_to_lowercase():
    results = parse_logs("input/logs.txt")
    assert all(r["level"] == r["level"].lower() for r in results)

def test_preserves_unicode():
    results = parse_logs("input/logs.txt")
    unicode_result = next((r for r in results if "情報" in r.get("msg", "")), None)
    assert unicode_result is not None
```

**Step 2: Create t1-2 fixtures — ugly but working processor.py**

`task-fixtures/t1-2/solution/processor.py` — write 200 lines of ugly but functional Python. Use repeated code, long functions, poor naming, but all tests pass.

```python
# solution/processor.py — intentionally ugly code
def processData(inputList, filterVal, transformType, outputFormat):
    result = []
    temp = []
    temp2 = []
    x = 0
    y = 0
    z = 0
    for i in range(len(inputList)):
        item = inputList[i]
        if item is None:
            continue
        if type(item) == str:
            if len(item) == 0:
                continue
            if filterVal is not None:
                if filterVal in item:
                    temp.append(item)
                else:
                    continue
            else:
                temp.append(item)
        elif type(item) == int:
            if filterVal is not None:
                if item == filterVal:
                    temp.append(item)
                else:
                    continue
            else:
                temp.append(item)
        elif type(item) == float:
            if filterVal is not None:
                if item == filterVal:
                    temp.append(item)
                else:
                    continue
            else:
                temp.append(item)
        elif type(item) == list:
            for j in range(len(item)):
                sub = item[j]
                if sub is None:
                    continue
                if filterVal is not None:
                    if sub == filterVal or (type(sub) == str and filterVal in sub):
                        temp.append(sub)
                else:
                    temp.append(sub)
        elif type(item) == dict:
            if "value" in item:
                v = item["value"]
                if filterVal is not None:
                    if v == filterVal or (type(v) == str and type(filterVal) == str and filterVal in v):
                        temp.append(v)
                else:
                    temp.append(v)
    if transformType == "upper":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == str:
                temp2.append(t.upper())
            else:
                temp2.append(t)
    elif transformType == "lower":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == str:
                temp2.append(t.lower())
            else:
                temp2.append(t)
    elif transformType == "reverse":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == str:
                temp2.append(t[::-1])
            elif type(t) == list:
                temp2.append(list(reversed(t)))
            else:
                temp2.append(t)
    elif transformType == "double":
        for i in range(len(temp)):
            t = temp[i]
            if type(t) == int or type(t) == float:
                temp2.append(t * 2)
            elif type(t) == str:
                temp2.append(t + t)
            else:
                temp2.append(t)
    else:
        temp2 = temp
    if outputFormat == "list":
        result = temp2
    elif outputFormat == "set":
        result = list(set(temp2))
    elif outputFormat == "count":
        result = len(temp2)
    elif outputFormat == "first":
        if len(temp2) > 0:
            result = temp2[0]
        else:
            result = None
    elif outputFormat == "last":
        if len(temp2) > 0:
            result = temp2[-1]
        else:
            result = None
    elif outputFormat == "dict":
        result = {}
        for i in range(len(temp2)):
            result[i] = temp2[i]
    else:
        result = temp2
    return result
```

Add `test_processor.py` with tests covering all branches.

**Step 3: Create remaining fixtures for t1-3, t2-1 through t3-4**

For each: create the starter files (stubs, fixtures, tests) that the agent needs injected. Keep fixtures realistic but bounded. Each fixture directory must contain a `solution/` subfolder where the agent writes its output.

**Step 4: Commit**

```bash
git add task-fixtures/
git commit -m "feat: task fixtures — starter files for all 10 tasks"
```

---

## Task 6: Task Runner

**Files:**
- Create: `src/runner.py`
- Create: `tests/test_runner.py`

The runner launches Claude Code headless for each task×profile combination, captures all output, and returns a structured result.

**Step 1: Write failing test**

```python
# tests/test_runner.py
import pytest
from unittest.mock import patch, MagicMock
from src.runner import run_task, TaskResult

def test_run_task_returns_result():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="done", stderr="")
        result = run_task(
            task_id="t1-1",
            task_spec="implement parser",
            profile_name="zero",
            system_prompt="",
            working_dir="/tmp/test-run",
        )
    assert isinstance(result, TaskResult)
    assert result.task_id == "t1-1"
    assert result.profile_name == "zero"
    assert result.exit_code == 0

def test_run_task_captures_output():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="output text", stderr="")
        result = run_task("t1-1", "spec", "zero", "", "/tmp/test")
    assert result.stdout == "output text"
```

**Step 2: Run to verify failures**

```bash
pytest tests/test_runner.py -v
```

**Step 3: Write src/runner.py**

```python
import subprocess
import shutil
import tempfile
from dataclasses import dataclass, field
from pathlib import Path

@dataclass
class TaskResult:
    task_id: str
    profile_name: str
    exit_code: int
    stdout: str
    stderr: str
    working_dir: str
    files_written: list[str] = field(default_factory=list)

def run_task(
    task_id: str,
    task_spec: str,
    profile_name: str,
    system_prompt: str,
    working_dir: str,
) -> TaskResult:
    """Run a single task with Claude Code headless."""
    run_dir = Path(working_dir) / task_id / profile_name
    run_dir.mkdir(parents=True, exist_ok=True)

    # Write the task spec as a prompt file
    prompt_file = run_dir / "prompt.md"
    prompt_file.write_text(task_spec)

    cmd = ["claude", "--print", "--no-interactive"]
    if system_prompt:
        cmd += ["--system", system_prompt]
    cmd += [f"@{prompt_file}"]

    result = subprocess.run(
        cmd,
        cwd=str(run_dir),
        capture_output=True,
        text=True,
        timeout=300,  # 5 min max per task
    )

    files_written = [
        str(f.relative_to(run_dir))
        for f in run_dir.rglob("*")
        if f.is_file() and f.name != "prompt.md"
    ]

    return TaskResult(
        task_id=task_id,
        profile_name=profile_name,
        exit_code=result.returncode,
        stdout=result.stdout,
        stderr=result.stderr,
        working_dir=str(run_dir),
        files_written=files_written,
    )
```

**Step 4: Run tests**

```bash
pytest tests/test_runner.py -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add src/runner.py tests/test_runner.py
git commit -m "feat: task runner — headless Claude Code execution"
```

---

## Task 7: Automated Scoring Metrics

**Files:**
- Create: `src/scoring/automated.py`
- Create: `tests/test_scoring_automated.py`

**Step 1: Write failing tests**

```python
# tests/test_scoring_automated.py
from src.scoring.automated import (
    score_tests,
    score_lines_of_code,
    score_complexity,
    score_scope,
    AutomatedScores,
)

def test_score_tests_all_pass():
    scores = score_tests(test_dir="/tmp/nonexistent", exit_code=0, stdout="5 passed")
    assert scores.tests_pass == 10.0

def test_score_tests_all_fail():
    scores = score_tests(test_dir="/tmp/nonexistent", exit_code=1, stdout="5 failed")
    assert scores.tests_pass == 0.0

def test_score_tests_partial():
    scores = score_tests(test_dir="/tmp/nonexistent", exit_code=1, stdout="3 passed, 2 failed")
    assert 4.0 < scores.tests_pass < 8.0

def test_score_loc_minimal(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text("def f(x):\n    return x * 2\n")
    score = score_lines_of_code(str(tmp_path), reference_minimal=5, reference_verbose=60)
    assert score >= 8.0

def test_score_loc_verbose(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text("\n".join([f"# line {i}" for i in range(80)]))
    score = score_lines_of_code(str(tmp_path), reference_minimal=5, reference_verbose=60)
    assert score <= 4.0
```

**Step 2: Run to verify failures**

```bash
pytest tests/test_scoring_automated.py -v
```

**Step 3: Write src/scoring/__init__.py and src/scoring/automated.py**

```python
# src/scoring/automated.py
import ast
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path

@dataclass
class AutomatedScores:
    tests_pass: float       # 0-10
    loc_score: float        # 0-10 (fewer = better, down to reference_minimal)
    complexity_score: float # 0-10 (lower complexity = higher score)
    scope_score: float      # 0-10 (less scope creep = higher score)

def score_tests(test_dir: str, exit_code: int, stdout: str) -> AutomatedScores:
    """Score based on test pass rate."""
    passed = _extract_count(stdout, r"(\d+) passed")
    failed = _extract_count(stdout, r"(\d+) failed")
    total = passed + failed
    if total == 0:
        rate = 1.0 if exit_code == 0 else 0.0
    else:
        rate = passed / total
    return AutomatedScores(
        tests_pass=round(rate * 10, 1),
        loc_score=0.0,
        complexity_score=0.0,
        scope_score=0.0,
    )

def score_lines_of_code(solution_dir: str, reference_minimal: int, reference_verbose: int) -> float:
    """Score LOC: minimal=10, verbose=0, linear interpolation."""
    py_files = list(Path(solution_dir).rglob("*.py"))
    if not py_files:
        return 5.0
    total_loc = sum(
        len([l for l in f.read_text().splitlines() if l.strip() and not l.strip().startswith("#")])
        for f in py_files
    )
    if total_loc <= reference_minimal:
        return 10.0
    if total_loc >= reference_verbose:
        return 0.0
    ratio = (total_loc - reference_minimal) / (reference_verbose - reference_minimal)
    return round((1 - ratio) * 10, 1)

def score_complexity(solution_dir: str) -> float:
    """Score cyclomatic complexity via radon. Lower complexity = higher score."""
    try:
        result = subprocess.run(
            ["radon", "cc", str(solution_dir), "-a", "-s"],
            capture_output=True, text=True
        )
        match = re.search(r"Average complexity: \w \((\d+\.\d+)\)", result.stdout)
        if not match:
            return 7.0  # default if no functions found
        avg = float(match.group(1))
        # A=1-5, B=6-10, C=11-15, D=16-20, E=21-25, F=25+
        if avg <= 5:   return 10.0
        if avg <= 10:  return 7.5
        if avg <= 15:  return 5.0
        if avg <= 20:  return 2.5
        return 0.0
    except Exception:
        return 5.0

def score_scope(task_allowed_files: list[str], files_written: list[str]) -> float:
    """Score scope discipline. Files written outside allowed set reduce score."""
    if not files_written:
        return 5.0
    out_of_scope = [f for f in files_written if not any(allowed in f for allowed in task_allowed_files)]
    ratio = len(out_of_scope) / len(files_written)
    return round((1 - ratio) * 10, 1)

def _extract_count(text: str, pattern: str) -> int:
    match = re.search(pattern, text)
    return int(match.group(1)) if match else 0
```

**Step 4: Run tests**

```bash
pytest tests/test_scoring_automated.py -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add src/scoring/ tests/test_scoring_automated.py
git commit -m "feat: automated scoring — tests, LOC, complexity, scope"
```

---

## Task 8: LLM-as-Judge Scoring

**Files:**
- Create: `src/scoring/judge.py`
- Create: `prompts/judge-rubric.md`
- Create: `tests/test_scoring_judge.py`

**Step 1: Write failing tests**

```python
# tests/test_scoring_judge.py
import pytest
from unittest.mock import patch, MagicMock
from src.scoring.judge import score_with_judge, JudgeScores

def test_judge_returns_scores():
    mock_response = MagicMock()
    mock_response.content[0].text = """
REQUIREMENT_INTERPRETATION: 8
DECISION_COMMUNICATION: 7
SELF_AWARENESS: 6
RECOVERY_QUALITY: 5
UNCONVENTIONAL_THINKING: 9
RATIONALE: Agent reframed the problem elegantly.
"""
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.return_value = mock_response
        scores = score_with_judge(session_transcript="agent did work", task_spec="do the work")
    assert isinstance(scores, JudgeScores)
    assert scores.requirement_interpretation == 8
    assert scores.unconventional_thinking == 9

def test_judge_takes_median_of_three():
    responses = [
        "REQUIREMENT_INTERPRETATION: 6\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 5\nRECOVERY_QUALITY: 4\nUNCONVENTIONAL_THINKING: 8\nRATIONALE: ok",
        "REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 6\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok",
        "REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 8\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: ok",
    ]
    mocks = [MagicMock() for _ in responses]
    for m, r in zip(mocks, responses):
        m.content[0].text = r
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.side_effect = mocks
        scores = score_with_judge("transcript", "spec")
    assert scores.requirement_interpretation == 7  # median of 6, 8, 7
```

**Step 2: Write prompts/judge-rubric.md**

```markdown
# Judge Rubric

You are evaluating an AI agent's performance on a software engineering task.
Score each dimension from 0-10. Respond ONLY in the format below — no extra text.

## Scoring Guide

**REQUIREMENT_INTERPRETATION** (0-10)
- 2: Misunderstood the core ask, built the wrong thing
- 5: Got the main requirement but missed nuance or edge cases
- 8: Understood everything including implicit expectations
- 10: Identified requirements the spec didn't explicitly state

**DECISION_COMMUNICATION** (0-10)
- 2: Made judgment calls silently, no explanation
- 5: Mentioned decisions but didn't explain reasoning
- 8: Clearly explained why for non-obvious choices
- 10: Proactively surfaced tradeoffs the user didn't ask about

**SELF_AWARENESS** (0-10)
- 2: Overconfident, no acknowledgment of uncertainty
- 5: Some hedging but inconsistent
- 8: Accurately flagged what it was unsure about
- 10: Calibrated confidence — certain where justified, uncertain where not

**RECOVERY_QUALITY** (0-10)
- 2: Gave up or brute-forced through failure
- 5: Recovered but lost work or repeated mistakes
- 8: Diagnosed root cause and adapted cleanly
- 10: Preserved partial work, adapted strategy, learned from failure

**UNCONVENTIONAL_THINKING** (0-10)
- 2: Took the first obvious approach without consideration
- 5: Considered alternatives but defaulted to conventional
- 8: Reframed the problem before diving in
- 10: Found an insight that made the problem significantly simpler

## Response Format (EXACT)
REQUIREMENT_INTERPRETATION: <0-10>
DECISION_COMMUNICATION: <0-10>
SELF_AWARENESS: <0-10>
RECOVERY_QUALITY: <0-10>
UNCONVENTIONAL_THINKING: <0-10>
RATIONALE: <one sentence>
```

**Step 3: Write src/scoring/judge.py**

```python
import re
import statistics
import anthropic
from dataclasses import dataclass
from pathlib import Path

@dataclass
class JudgeScores:
    requirement_interpretation: float
    decision_communication: float
    self_awareness: float
    recovery_quality: float
    unconventional_thinking: float
    rationale: str

_RUBRIC = (Path(__file__).parent.parent.parent / "prompts/judge-rubric.md").read_text()

def score_with_judge(session_transcript: str, task_spec: str, runs: int = 3) -> JudgeScores:
    """Score a session with LLM-as-judge. Returns median of `runs` evaluations."""
    client = anthropic.Anthropic()
    all_scores = []
    for _ in range(runs):
        response = client.messages.create(
            model="claude-sonnet-4-6",
            max_tokens=256,
            system=_RUBRIC,
            messages=[{
                "role": "user",
                "content": f"## Task Spec\n\n{task_spec}\n\n## Agent Session\n\n{session_transcript}"
            }]
        )
        all_scores.append(_parse_scores(response.content[0].text))

    return _median_scores(all_scores)

def _parse_scores(text: str) -> dict:
    scores = {}
    for key in ["REQUIREMENT_INTERPRETATION", "DECISION_COMMUNICATION", "SELF_AWARENESS", "RECOVERY_QUALITY", "UNCONVENTIONAL_THINKING"]:
        match = re.search(rf"{key}: (\d+)", text)
        scores[key] = int(match.group(1)) if match else 5
    match = re.search(r"RATIONALE: (.+)", text)
    scores["RATIONALE"] = match.group(1) if match else ""
    return scores

def _median_scores(all_scores: list[dict]) -> JudgeScores:
    def med(key): return statistics.median(s[key] for s in all_scores)
    return JudgeScores(
        requirement_interpretation=med("REQUIREMENT_INTERPRETATION"),
        decision_communication=med("DECISION_COMMUNICATION"),
        self_awareness=med("SELF_AWARENESS"),
        recovery_quality=med("RECOVERY_QUALITY"),
        unconventional_thinking=med("UNCONVENTIONAL_THINKING"),
        rationale=all_scores[0]["RATIONALE"],
    )
```

**Step 4: Run tests**

```bash
pytest tests/test_scoring_judge.py -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add src/scoring/judge.py prompts/ tests/test_scoring_judge.py
git commit -m "feat: LLM-as-judge scoring with 3-run median"
```

---

## Task 9: Results Generator (HTML + JSON)

**Files:**
- Create: `src/results.py`
- Create: `templates/results.html.j2`
- Create: `tests/test_results.py`

This is the most visible deliverable. The HTML must be magazine quality — use the `visual-output-standards` skill for color palette and SVG formatting.

**Step 1: Write failing test**

```python
# tests/test_results.py
from src.results import generate_results, ResultsReport

def test_generate_results_produces_html(tmp_path):
    report = ResultsReport(
        run_id="2026-03-30T12:00:00",
        task_suite_version="v1",
        configurations=["zero", "light", "user"],
        scores={
            "zero": {"overall": 4.2, "tier1": 5.0, "tier2": 4.0, "tier3": 3.5},
            "light": {"overall": 6.1, "tier1": 6.5, "tier2": 6.0, "tier3": 5.8},
            "user": {"overall": 8.3, "tier1": 8.0, "tier2": 8.5, "tier3": 8.4},
        },
        dimension_scores={
            "zero": {"correctness": 5.0, "elegance": 3.0, "discipline": 2.0, "judgment": 3.5, "creativity": 2.5, "recovery": 4.0},
            "light": {"correctness": 6.5, "elegance": 5.5, "discipline": 5.0, "judgment": 6.0, "creativity": 5.0, "recovery": 6.5},
            "user": {"correctness": 8.5, "elegance": 8.0, "discipline": 8.5, "judgment": 8.0, "creativity": 9.0, "recovery": 7.5},
        }
    )
    html_path = tmp_path / "results.html"
    json_path = tmp_path / "results.json"
    generate_results(report, html_path, json_path)
    assert html_path.exists()
    assert json_path.exists()
    html = html_path.read_text()
    assert "Proving Ground" in html
    assert "8.3" in html  # user overall score

def test_generate_results_self_contained(tmp_path):
    """HTML must have no external dependencies."""
    # (generate and verify no <link> or external <script> tags)
    pass  # implement after HTML template is written
```

**Step 2: Write src/results.py and templates/results.html.j2**

Use the `visual-output-standards` skill to define the color palette and SVG radar chart spec before writing the template.

The HTML template must include:
- Inline CSS only (no external stylesheets)
- Inline SVG charts (radar chart, tier breakdown bars)
- A "headline" section with composite scores and a one-sentence verdict
- A radar chart section with all configurations overlaid
- A tier breakdown section with per-task score cards
- A history section (conditionally shown if > 1 run)

**Step 3: Run tests**

```bash
pytest tests/test_results.py -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add src/results.py templates/ tests/test_results.py
git commit -m "feat: results generator — HTML and JSON output"
```

---

## Task 10: History Engine

**Files:**
- Create: `src/history.py`
- Create: `tests/test_history.py`

**Step 1: Write failing tests**

```python
# tests/test_history.py
import json
from pathlib import Path
from src.history import append_run, load_history, HistoryEntry

def test_append_creates_history_file(tmp_path):
    entry = HistoryEntry(run_id="2026-03-30T12:00:00", scores={"zero": 4.2, "user": 8.3})
    append_run(tmp_path / "history.jsonl", entry)
    assert (tmp_path / "history.jsonl").exists()

def test_append_multiple_runs(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    for i in range(3):
        append_run(hist_file, HistoryEntry(run_id=f"run-{i}", scores={"user": float(i + 5)}))
    entries = load_history(hist_file)
    assert len(entries) == 3

def test_load_history_sorted_by_date(tmp_path):
    hist_file = tmp_path / "history.jsonl"
    for run_id in ["2026-03-31", "2026-03-29", "2026-03-30"]:
        append_run(hist_file, HistoryEntry(run_id=run_id, scores={"user": 7.0}))
    entries = load_history(hist_file)
    assert entries[0].run_id == "2026-03-29"
    assert entries[-1].run_id == "2026-03-31"
```

**Step 2: Write src/history.py**

```python
import json
from dataclasses import dataclass, asdict
from pathlib import Path

@dataclass
class HistoryEntry:
    run_id: str
    scores: dict[str, float]

def append_run(history_file: Path, entry: HistoryEntry) -> None:
    with open(history_file, "a") as f:
        f.write(json.dumps(asdict(entry)) + "\n")

def load_history(history_file: Path) -> list[HistoryEntry]:
    if not history_file.exists():
        return []
    entries = []
    for line in history_file.read_text().splitlines():
        if line.strip():
            d = json.loads(line)
            entries.append(HistoryEntry(**d))
    return sorted(entries, key=lambda e: e.run_id)
```

**Step 3: Run tests**

```bash
pytest tests/test_history.py -v
```
Expected: PASS

**Step 4: Commit**

```bash
git add src/history.py tests/test_history.py
git commit -m "feat: history engine — run persistence and trend tracking"
```

---

## Task 11: Orchestrator & CLI Integration

**Files:**
- Modify: `src/cli.py`
- Create: `src/orchestrator.py`
- Create: `tests/test_orchestrator.py`

Wire together: profile loading → task loading → task runner → scoring → results generator → history engine.

**Step 1: Write failing integration test**

```python
# tests/test_orchestrator.py
from unittest.mock import patch, MagicMock
from src.orchestrator import run_benchmark

def test_run_benchmark_produces_outputs(tmp_path):
    profiles_dir = tmp_path / "profiles"
    profiles_dir.mkdir()
    (profiles_dir / "my-agent.txt").write_text("You are a careful engineer.")

    with patch("src.orchestrator.run_task") as mock_runner, \
         patch("src.orchestrator.score_with_judge") as mock_judge:
        mock_runner.return_value = MagicMock(exit_code=0, stdout="1 passed", stderr="", files_written=["solution/parser.py"])
        mock_judge.return_value = MagicMock(
            requirement_interpretation=7, decision_communication=6,
            self_awareness=7, recovery_quality=6, unconventional_thinking=5, rationale="ok"
        )
        result = run_benchmark(
            data_dir=str(tmp_path),
            tiers=["1"],
        )

    assert (tmp_path / "results.html").exists()
    assert (tmp_path / "results.json").exists()
```

**Step 2: Write src/orchestrator.py and update src/cli.py**

**Step 3: Run integration test**

```bash
pytest tests/test_orchestrator.py -v
```

**Step 4: Commit**

```bash
git add src/orchestrator.py src/cli.py tests/test_orchestrator.py
git commit -m "feat: orchestrator — full benchmark pipeline wired together"
```

---

## Task 12: Dockerfile

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml` (for local dev convenience)

**Step 1: Write Dockerfile**

```dockerfile
FROM python:3.12-slim

# Install Node.js (required for Claude Code CLI)
RUN apt-get update && apt-get install -y curl git && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

# Install radon for complexity analysis
RUN pip install radon

# Set up app
WORKDIR /app
COPY pyproject.toml .
RUN pip install -e .

COPY . .

# Volume for user data (profiles, results, history)
VOLUME ["/data"]

# Entrypoint
ENTRYPOINT ["proving-ground", "--data-dir", "/data"]
CMD []
```

**Step 2: Write docker-compose.yml**

```yaml
services:
  proving-ground:
    build: .
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./data:/data
    command: ["--tier", "all"]
```

**Step 3: Build and smoke test**

```bash
docker build -t proving-ground .
docker run --rm proving-ground --help
```
Expected: Help text shows, no errors.

**Step 4: Commit**

```bash
git add Dockerfile docker-compose.yml
git commit -m "feat: Dockerfile — single-container benchmark runner"
```

---

## Task 13: Results Page — Visual Polish

**Files:**
- Modify: `templates/results.html.j2`

Run the `visual-output-standards` skill before starting this task. The results page must meet magazine quality — not a developer tool aesthetic.

Requirements:
- Typography: headline font (serif or editorial sans), body font (clean sans)
- Color: limited palette (3-4 colors max), high contrast
- Radar chart: clean SVG, filled polygons with transparency, labeled axes
- Score cards: clean grid, color-coded by score bracket (red < 5, amber 5-7, green > 7)
- Verdict sentence: large, prominent, typographically treated
- History sparklines: minimal, inline SVGs, 40px tall

**Step 1: Apply visual-output-standards skill for palette and SVG specs**

**Step 2: Implement the full template**

**Step 3: Generate a sample results.html with fixture data and review visually**

```bash
python -c "
from src.results import generate_results, ResultsReport
from pathlib import Path
report = ResultsReport(...)  # use fixture data from test
generate_results(report, Path('sample-results.html'), Path('sample-results.json'))
"
open sample-results.html
```

**Step 4: Iterate until magazine quality. Then commit.**

```bash
git add templates/
git commit -m "feat: results page — magazine quality visual design"
```

---

## Final: Push & Test End-to-End

**Step 1: Run full test suite**

```bash
pytest tests/ -v
```
Expected: All PASS

**Step 2: Push feature branch**

```bash
git push -u origin feature/build
```

**Step 3: Use superpowers:finishing-a-development-branch to merge**
