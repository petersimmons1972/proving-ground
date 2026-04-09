import re
import subprocess
from dataclasses import dataclass
from pathlib import Path


@dataclass
class AutomatedScores:
    tests_pass: float  # 0-10: test pass rate


def score_tests(exit_code: int, stdout: str) -> AutomatedScores:
    """Score based on pytest output. Parses 'N passed, M failed' pattern."""
    passed = _extract_count(stdout, r"(\d+) passed")
    failed = _extract_count(stdout, r"(\d+) failed")
    total = passed + failed

    if total == 0:
        rate = 0.0  # no tests found = no evidence of correctness
    else:
        rate = passed / total

    return AutomatedScores(tests_pass=round(rate * 10, 1))


def score_lines_of_code(
    solution_dir: str,
    reference_minimal: int,
    reference_verbose: int,
) -> float:
    """Score LOC: at or below minimal=10.0, at or above verbose=0.0, linear between."""
    py_files = list(Path(solution_dir).rglob("*.py"))
    if not py_files:
        return 5.0

    total_loc = sum(
        len([
            line for line in f.read_text().splitlines()
            if line.strip() and not line.strip().startswith("#")
        ])
        for f in py_files
    )

    if total_loc <= reference_minimal:
        return 10.0
    if total_loc >= reference_verbose:
        return 0.0

    ratio = (total_loc - reference_minimal) / (reference_verbose - reference_minimal)
    return round((1.0 - ratio) * 10, 1)


def score_complexity(solution_dir: str) -> float:
    """Score cyclomatic complexity using radon. Lower average complexity = higher score."""
    try:
        result = subprocess.run(
            ["radon", "cc", solution_dir, "-a", "-s"],
            capture_output=True,
            text=True,
            timeout=30,
        )
        match = re.search(r"Average complexity: \w \((\d+\.?\d*)\)", result.stdout)
        if not match:
            return 6.0  # neutral default -- no functions found

        avg = float(match.group(1))
        # Radon grades: A=1-5, B=6-10, C=11-15, D=16-20, E=21-25, F=25+
        if avg <= 5:    return 10.0
        if avg <= 10:   return 7.5
        if avg <= 15:   return 5.0
        if avg <= 20:   return 2.5
        return 0.0
    except Exception:
        return 5.0


def score_scope(allowed_files: list[str], files_written: list[str]) -> float:
    """Score scope discipline. Files written outside allowed prefixes reduce score."""
    if not files_written:
        return 10.0

    out_of_scope = [
        f for f in files_written
        if not any(f.startswith(allowed) for allowed in allowed_files)
    ]
    ratio = len(out_of_scope) / len(files_written)
    return round((1.0 - ratio) * 10, 1)


def _extract_count(text: str, pattern: str) -> int:
    match = re.search(pattern, text)
    return int(match.group(1)) if match else 0
