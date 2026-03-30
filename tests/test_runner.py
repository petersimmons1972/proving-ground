import pytest
from unittest.mock import patch, MagicMock
from src.runner import run_task, TaskResult


def test_run_task_returns_result():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="1 passed", stderr="")
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


def test_run_task_captures_stdout():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="output text", stderr="")
        result = run_task("t1-1", "spec", "zero", "", "/tmp/test")
    assert result.stdout == "output text"


def test_run_task_captures_stderr():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=1, stdout="", stderr="error msg")
        result = run_task("t1-1", "spec", "zero", "", "/tmp/test")
    assert result.exit_code == 1
    assert result.stderr == "error msg"


def test_run_task_with_system_prompt_passes_it():
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        run_task("t1-1", "spec", "light", "You are an engineer.", "/tmp/test")
    call_args = mock_run.call_args
    cmd = call_args[0][0]
    assert "--system-prompt" in cmd or "--system" in cmd


def test_run_task_zero_profile_no_system_flag():
    """Zero profile (empty string) should not pass --system flag."""
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        run_task("t1-1", "spec", "zero", "", "/tmp/test")
    call_args = mock_run.call_args
    cmd = call_args[0][0]
    assert "--system-prompt" not in cmd
    assert "--system" not in cmd


def test_files_written_excludes_prompt_md(tmp_path):
    """prompt.md is seeded by the runner itself and must not appear in files_written.

    Uses a real tmp directory so the actual rglob + name-filter logic is exercised,
    not a mock of the filesystem.
    """
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        result = run_task("t1-1", "my task spec", "zero", "", str(tmp_path))
    assert "prompt.md" not in result.files_written
    assert not any(f.endswith("prompt.md") for f in result.files_written)
