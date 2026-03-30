import re
import statistics
import subprocess
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


_RUBRIC_PATH = Path(__file__).parent.parent.parent / "prompts" / "judge-rubric.md"


def score_with_judge(
    session_transcript: str,
    task_spec: str,
    runs: int = 3,
) -> JudgeScores:
    """Score a session with LLM-as-judge via claude --print. Returns median of `runs` evaluations."""
    rubric = _RUBRIC_PATH.read_text()
    prompt = (
        f"## Task Spec\n\n{task_spec}\n\n"
        f"## Agent Session\n\n{session_transcript}"
    )
    all_scores: list[dict] = []

    for _ in range(runs):
        result = subprocess.run(
            [
                "claude", "--print",
                "--dangerously-skip-permissions",
                "--no-session-persistence",
                "--system-prompt", rubric,
            ],
            input=prompt,
            capture_output=True,
            text=True,
            timeout=300,
        )
        all_scores.append(_parse_scores(result.stdout))

    return _median_scores(all_scores)


def _parse_scores(text: str) -> dict:
    """Extract dimension scores from judge response text."""
    dimensions = [
        "REQUIREMENT_INTERPRETATION",
        "DECISION_COMMUNICATION",
        "SELF_AWARENESS",
        "RECOVERY_QUALITY",
        "UNCONVENTIONAL_THINKING",
    ]
    scores: dict = {}
    for dim in dimensions:
        match = re.search(rf"{dim}:\s*(\d+)", text)
        scores[dim] = int(match.group(1)) if match else 5  # default to neutral

    rationale_match = re.search(r"RATIONALE:\s*(.+)", text)
    scores["RATIONALE"] = rationale_match.group(1).strip() if rationale_match else ""
    return scores


def _median_scores(all_scores: list[dict]) -> JudgeScores:
    """Return a JudgeScores with the median value for each dimension."""
    def med(key: str) -> float:
        return statistics.median(s[key] for s in all_scores)

    return JudgeScores(
        requirement_interpretation=med("REQUIREMENT_INTERPRETATION"),
        decision_communication=med("DECISION_COMMUNICATION"),
        self_awareness=med("SELF_AWARENESS"),
        recovery_quality=med("RECOVERY_QUALITY"),
        unconventional_thinking=med("UNCONVENTIONAL_THINKING"),
        rationale=all_scores[0]["RATIONALE"],
    )
