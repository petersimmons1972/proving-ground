import pytest
import statistics
from unittest.mock import patch, MagicMock
from src.scoring.judge import score_with_judge, JudgeScores


def _make_mock_response(scores_text: str) -> MagicMock:
    mock = MagicMock()
    mock.content = [MagicMock()]
    mock.content[0].text = scores_text
    return mock


SAMPLE_RESPONSE = (
    "REQUIREMENT_INTERPRETATION: 8\n"
    "DECISION_COMMUNICATION: 7\n"
    "SELF_AWARENESS: 6\n"
    "RECOVERY_QUALITY: 5\n"
    "UNCONVENTIONAL_THINKING: 9\n"
    "RATIONALE: Agent reframed the problem elegantly.\n"
)


def test_judge_returns_judge_scores():
    mock = _make_mock_response(SAMPLE_RESPONSE)
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.return_value = mock
        scores = score_with_judge(
            session_transcript="agent did some work",
            task_spec="do the work",
            runs=1,
        )
    assert isinstance(scores, JudgeScores)


def test_judge_parses_all_dimensions():
    mock = _make_mock_response(SAMPLE_RESPONSE)
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.return_value = mock
        scores = score_with_judge("transcript", "spec", runs=1)
    assert scores.requirement_interpretation == 8
    assert scores.decision_communication == 7
    assert scores.self_awareness == 6
    assert scores.recovery_quality == 5
    assert scores.unconventional_thinking == 9


def test_judge_captures_rationale():
    mock = _make_mock_response(SAMPLE_RESPONSE)
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.return_value = mock
        scores = score_with_judge("transcript", "spec", runs=1)
    assert "reframed" in scores.rationale


def test_judge_takes_median_of_three():
    responses = [
        "REQUIREMENT_INTERPRETATION: 6\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 5\nRECOVERY_QUALITY: 4\nUNCONVENTIONAL_THINKING: 8\nRATIONALE: ok",
        "REQUIREMENT_INTERPRETATION: 8\nDECISION_COMMUNICATION: 7\nSELF_AWARENESS: 7\nRECOVERY_QUALITY: 6\nUNCONVENTIONAL_THINKING: 7\nRATIONALE: ok",
        "REQUIREMENT_INTERPRETATION: 7\nDECISION_COMMUNICATION: 8\nSELF_AWARENESS: 6\nRECOVERY_QUALITY: 5\nUNCONVENTIONAL_THINKING: 9\nRATIONALE: ok",
    ]
    mocks = [_make_mock_response(r) for r in responses]
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.side_effect = mocks
        scores = score_with_judge("transcript", "spec", runs=3)
    # medians: req=7, comm=7, aware=6, recov=5, unconv=8
    assert scores.requirement_interpretation == 7
    assert scores.decision_communication == 7
    assert scores.self_awareness == 6
    assert scores.recovery_quality == 5
    assert scores.unconventional_thinking == 8


def test_judge_handles_missing_dimension_gracefully():
    # Response missing UNCONVENTIONAL_THINKING
    partial = (
        "REQUIREMENT_INTERPRETATION: 7\n"
        "DECISION_COMMUNICATION: 6\n"
        "SELF_AWARENESS: 7\n"
        "RECOVERY_QUALITY: 5\n"
        "RATIONALE: incomplete response\n"
    )
    mock = _make_mock_response(partial)
    with patch("src.scoring.judge.anthropic.Anthropic") as MockClient:
        MockClient.return_value.messages.create.return_value = mock
        scores = score_with_judge("transcript", "spec", runs=1)
    assert scores.unconventional_thinking == 5  # default fallback
