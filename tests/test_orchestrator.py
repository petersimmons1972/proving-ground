import pytest
from pathlib import Path
from unittest.mock import patch, MagicMock, call
from src.orchestrator import run_benchmark


def test_run_benchmark_produces_html(tmp_path):
    profiles_dir = tmp_path / "profiles"
    profiles_dir.mkdir()
    (profiles_dir / "my-agent.txt").write_text("You are a careful engineer.")

    mock_task_result = MagicMock()
    mock_task_result.exit_code = 0
    mock_task_result.stdout = "1 passed"
    mock_task_result.stderr = ""
    mock_task_result.files_written = ["solution/parser.py"]
    mock_task_result.working_dir = str(tmp_path)

    mock_judge = MagicMock()
    mock_judge.requirement_interpretation = 7
    mock_judge.decision_communication = 6
    mock_judge.self_awareness = 7
    mock_judge.recovery_quality = 6
    mock_judge.unconventional_thinking = 5
    mock_judge.rationale = "ok"

    with patch("src.orchestrator.run_task", return_value=mock_task_result), \
         patch("src.orchestrator.score_with_judge", return_value=mock_judge):
        run_benchmark(data_dir=str(tmp_path), tiers=["1"])

    assert (tmp_path / "results.html").exists()
    assert (tmp_path / "results.json").exists()


def test_run_benchmark_runs_each_config(tmp_path):
    """run_task should be called once per task x configuration."""
    profiles_dir = tmp_path / "profiles"
    profiles_dir.mkdir()
    (profiles_dir / "user.txt").write_text("I am an agent.")

    mock_result = MagicMock(exit_code=0, stdout="1 passed", stderr="", files_written=[], working_dir=str(tmp_path))
    mock_judge = MagicMock(
        requirement_interpretation=7, decision_communication=6,
        self_awareness=7, recovery_quality=6, unconventional_thinking=5, rationale="ok"
    )

    with patch("src.orchestrator.run_task", return_value=mock_result) as mock_run, \
         patch("src.orchestrator.score_with_judge", return_value=mock_judge):
        run_benchmark(data_dir=str(tmp_path), tiers=["1"])

    # 3 tier-1 tasks x 3 configs (zero, light, user) = 9 calls
    assert mock_run.call_count == 9


def test_run_benchmark_appends_history(tmp_path):
    """Each run should append to history.jsonl."""
    profiles_dir = tmp_path / "profiles"
    profiles_dir.mkdir()

    mock_result = MagicMock(exit_code=0, stdout="1 passed", stderr="", files_written=[], working_dir=str(tmp_path))
    mock_judge = MagicMock(
        requirement_interpretation=7, decision_communication=6,
        self_awareness=7, recovery_quality=6, unconventional_thinking=5, rationale="ok"
    )

    with patch("src.orchestrator.run_task", return_value=mock_result), \
         patch("src.orchestrator.score_with_judge", return_value=mock_judge):
        run_benchmark(data_dir=str(tmp_path), tiers=["1"])

    assert (tmp_path / "history.jsonl").exists()
