import pytest
import subprocess
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
        run_task("t1-1", "spec text", "light", "You are an engineer.", "/tmp/test")
    call_args = mock_run.call_args
    cmd = call_args[0][0]
    # System prompt goes via --system-prompt flag
    assert "--system-prompt" in cmd or "--system" in cmd
    # Spec is piped via stdin, not included as a positional CLI argument
    assert "spec text" not in cmd


def test_run_task_zero_profile_no_system_flag():
    """Zero profile (empty string) should not pass --system flag."""
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        run_task("t1-1", "spec", "zero", "", "/tmp/test")
    call_args = mock_run.call_args
    cmd = call_args[0][0]
    assert "--system-prompt" not in cmd
    assert "--system" not in cmd


def test_spec_piped_via_stdin(tmp_path):
    """Task spec must be passed as subprocess input (stdin), not as a CLI argument.

    The runner prepends a WORKING DIRECTORY header to anchor file operations
    (fix for stray-files-at-repo-root leak, issue #2). The original spec text
    must still appear in the piped stdin, just after the header.
    """
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        run_task("t1-1", "my task spec", "zero", "", str(tmp_path))
    call_kwargs = mock_run.call_args.kwargs
    piped = call_kwargs.get("input", "")
    # The spec text arrives via the input= keyword argument
    assert "my task spec" in piped
    # The WORKING DIRECTORY header is prepended before the spec
    assert piped.startswith("WORKING DIRECTORY:")
    assert "my task spec" in piped.split("---\n")[1]  # body after separator
    # And is not present as a positional element in the command list
    cmd = mock_run.call_args[0][0]
    assert "my task spec" not in cmd


def test_spec_header_contains_absolute_run_dir(tmp_path):
    """The prepended WORKING DIRECTORY header must contain the absolute run_dir.

    Relative paths would leave claude to resolve against whatever cwd it happens
    to have. The absolute path removes that ambiguity.
    """
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")
        run_task("t1-1", "the spec", "zero", "", str(tmp_path))
    piped = mock_run.call_args.kwargs["input"]
    expected_run_dir = str((tmp_path / "t1-1" / "zero").absolute())
    assert expected_run_dir in piped
    assert "Never write files outside this directory" in piped


def test_run_task_timeout_returns_result():
    """TimeoutExpired should produce a TaskResult with non-zero exit code, not crash."""
    with patch("src.runner.subprocess.run") as mock_run:
        mock_run.side_effect = subprocess.TimeoutExpired(cmd="claude", timeout=300)
        result = run_task("t1-1", "spec", "zero", "", "/tmp/test")
    assert result.exit_code != 0
    assert "timeout" in result.stderr.lower()
