import subprocess
from unittest.mock import patch, MagicMock
from src.scoring.judge import score_with_judge, JudgeScores, _parse_scores

SAMPLE_RESPONSE = (
    "REQUIREMENT_INTERPRETATION: 8\n"
    "DECISION_COMMUNICATION: 7\n"
    "SELF_AWARENESS: 6\n"
    "RECOVERY_QUALITY: 5\n"
    "UNCONVENTIONAL_THINKING: 9\n"
    "RATIONALE: Agent reframed the problem elegantly.\n"
)

def _make_completed_process(stdout: str, returncode: int = 0):
    return subprocess.CompletedProcess(
        args=["claude"], returncode=returncode, stdout=stdout, stderr=""
    )

def test_judge_returns_judge_scores():
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process(SAMPLE_RESPONSE)):
        scores = score_with_judge("transcript", "spec", runs=1)
    assert isinstance(scores, JudgeScores)

def test_judge_parses_all_dimensions():
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process(SAMPLE_RESPONSE)):
        scores = score_with_judge("transcript", "spec", runs=1)
    assert scores.requirement_interpretation == 8
    assert scores.decision_communication == 7
    assert scores.self_awareness == 6
    assert scores.recovery_quality == 5
    assert scores.unconventional_thinking == 9

def test_judge_captures_rationale():
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process(SAMPLE_RESPONSE)):
        scores = score_with_judge("transcript", "spec", runs=1)
    assert "reframed" in scores.rationale

def test_judge_takes_median_of_three():
    responses = [
        _make_completed_process("REQUIREMENT_INTERPRETATION: 6\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 5\nRECOVERY_QUALITY: 4\nUNCONVENTIONAL_THINKING: 8\nRATIONALE: ok"),
        _make_completed_process("REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 6\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok"),
        _make_completed_process("REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 8\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: ok"),
    ]
    with patch("src.scoring.judge.subprocess.run", side_effect=responses):
        scores = score_with_judge("transcript", "spec", runs=3)
    assert scores.requirement_interpretation == 7
    assert scores.decision_communication == 7
    assert scores.self_awareness == 6
    assert scores.recovery_quality == 5
    assert scores.unconventional_thinking == 8

def test_judge_handles_missing_dimension():
    partial = (
        "REQUIREMENT_INTERPRETATION: 7\n"
        "DECISION_COMMUNICATION: 6\n"
        "SELF_AWARENESS: 7\n"
        "RECOVERY_QUALITY: 5\n"
        "RATIONALE: incomplete response\n"
    )
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process(partial)):
        scores = score_with_judge("transcript", "spec", runs=1)
    assert scores.unconventional_thinking == 5  # default fallback

def test_judge_skips_failed_subprocess():
    """Non-zero exit code should be skipped, not parsed as valid scores."""
    responses = [
        _make_completed_process("", returncode=1),  # failed
        _make_completed_process(SAMPLE_RESPONSE),     # success
    ]
    with patch("src.scoring.judge.subprocess.run", side_effect=responses):
        scores = score_with_judge("transcript", "spec", runs=2)
    assert scores.requirement_interpretation == 8  # from the successful run

def test_judge_all_runs_fail_returns_zeros():
    """When all judge runs fail, return zero scores with failure rationale."""
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process("", returncode=1)):
        scores = score_with_judge("transcript", "spec", runs=3)
    assert scores.requirement_interpretation == 0.0
    assert scores.rationale == "ALL_JUDGE_RUNS_FAILED"

def test_judge_timeout_skips_run():
    """TimeoutExpired should be caught and that run skipped."""
    responses = [
        subprocess.TimeoutExpired(cmd="claude", timeout=300),
        _make_completed_process(SAMPLE_RESPONSE),
    ]
    with patch("src.scoring.judge.subprocess.run", side_effect=responses):
        scores = score_with_judge("transcript", "spec", runs=2)
    assert scores.requirement_interpretation == 8

def test_judge_clamps_scores():
    """Scores outside 0-10 must be clamped."""
    wild = "REQUIREMENT_INTERPRETATION: 85\nDECISION_COMMUNICATION: -3\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: wild\n"
    with patch("src.scoring.judge.subprocess.run", return_value=_make_completed_process(wild)):
        scores = score_with_judge("transcript", "spec", runs=1)
    assert scores.requirement_interpretation == 10  # clamped from 85
    # Note: -3 won't match \d+ regex, so defaults to 5

def test_parse_scores_unit():
    """Direct test of _parse_scores parsing logic."""
    scores = _parse_scores(SAMPLE_RESPONSE)
    assert scores["REQUIREMENT_INTERPRETATION"] == 8
    assert scores["RATIONALE"] == "Agent reframed the problem elegantly."
