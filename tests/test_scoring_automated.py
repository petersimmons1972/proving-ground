import pytest
from pathlib import Path
from src.scoring.automated import (
    score_tests,
    score_lines_of_code,
    score_complexity,
    score_scope,
    AutomatedScores,
)


def test_score_tests_all_pass():
    scores = score_tests(exit_code=0, stdout="5 passed")
    assert scores.tests_pass == 10.0


def test_score_tests_all_fail():
    scores = score_tests(exit_code=1, stdout="5 failed")
    assert scores.tests_pass == 0.0


def test_score_tests_partial():
    scores = score_tests(exit_code=1, stdout="3 passed, 2 failed")
    assert 5.0 < scores.tests_pass < 7.0  # 3/5 = 0.6 -> 6.0


def test_score_tests_exact_partial():
    scores = score_tests(exit_code=1, stdout="3 passed, 2 failed")
    assert scores.tests_pass == 6.0


def test_score_loc_minimal(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text("def f(x):\n    return x * 2\n")
    score = score_lines_of_code(str(tmp_path), reference_minimal=5, reference_verbose=60)
    assert score >= 8.0


def test_score_loc_verbose(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text("\n".join([f"x_{i} = {i}" for i in range(80)]))
    score = score_lines_of_code(str(tmp_path), reference_minimal=5, reference_verbose=60)
    assert score <= 2.0


def test_score_loc_at_minimal_returns_10(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text("x = 1\ny = 2\nz = 3\n")
    score = score_lines_of_code(str(tmp_path), reference_minimal=3, reference_verbose=60)
    assert score == 10.0


def test_score_scope_no_creep():
    score = score_scope(
        allowed_files=["solution/"],
        files_written=["solution/parser.py", "solution/__init__.py"],
    )
    assert score == 10.0


def test_score_scope_full_creep():
    score = score_scope(
        allowed_files=["solution/"],
        files_written=["tests/new_test.py", "README.md"],
    )
    assert score == 0.0


def test_score_scope_partial_creep():
    score = score_scope(
        allowed_files=["solution/"],
        files_written=["solution/parser.py", "tests/new_test.py"],
    )
    assert score == 5.0


def test_score_complexity_no_files(tmp_path):
    # Empty dir -- should return a neutral default
    score = score_complexity(str(tmp_path))
    assert 4.0 <= score <= 8.0


def test_score_complexity_simple_function(tmp_path):
    f = tmp_path / "solution.py"
    f.write_text(
        "def add(a, b):\n"
        "    return a + b\n"
    )
    score = score_complexity(str(tmp_path))
    # Simple function with no branches = complexity 1 (grade A) = score 10.0
    assert score == 10.0


def test_score_tests_does_not_carry_dead_fields():
    """score_tests should only return tests_pass — no dead zero fields that corrupt downstream math."""
    scores = score_tests(exit_code=0, stdout="5 passed")
    assert scores.tests_pass == 10.0
    # These fields should not exist or should not be zero if used in arithmetic
    assert not hasattr(scores, 'scope_score') or scores.scope_score is None


def test_score_tests_no_tests_found_exit_zero():
    """No tests detected with exit 0 should NOT award perfect score."""
    scores = score_tests(exit_code=0, stdout="All done, no tests to run")
    assert scores.tests_pass == 0.0  # absence of tests != perfect score
