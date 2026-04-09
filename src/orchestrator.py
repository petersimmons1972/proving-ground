"""Orchestrator: main benchmark pipeline.

For each task x profile configuration, run it, score it, aggregate the results,
and produce the final HTML/JSON report plus a history entry.

The compiler ran because someone built it. The tests pass because someone wired
the pieces together.
"""

import os
import statistics
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone
from pathlib import Path

from src.history import append_run, HistoryEntry, load_history
from src.profiles import load_profiles
from src.results import generate_results, ResultsReport
from src.runner import run_task
from src.scoring.automated import score_tests, score_lines_of_code, score_complexity, score_scope
from src.scoring.judge import score_with_judge
from src.tasks import load_tasks

# Max profiles run in parallel within a single task. Each profile makes one
# claude --print call for the task run plus three for the judge (12+ claude
# subprocesses per task at full parallelism). Cap via PROVING_GROUND_MAX_WORKERS
# env var if rate-limited. Default 4 matches the standard 4-profile suite
# (zero + light + grace-hopper + qa-army-group) so each task runs one batch
# of four simultaneously rather than four sequential batches of one.
_MAX_WORKERS = int(os.environ.get("PROVING_GROUND_MAX_WORKERS", "4"))

# Per-task LOC references: (minimal, verbose) — tuned to each task type.
# Minimal = fewest non-comment lines a correct solution needs.
# Verbose = threshold above which elegance score drops to zero.
_LOC_REFS: dict[str, tuple[int, int]] = {
    "t1-1": (15, 80),
    "t1-2": (20, 80),
    "t1-3": (5, 40),
    "t2-1": (10, 60),
    "t2-2": (5, 30),
    "t2-3": (10, 50),
    "t3-1": (5, 40),
    "t3-2": (30, 120),
    "t3-3": (5, 30),
    "t3-4": (15, 60),
}

# Tier weights: higher tiers are harder and count for more of the overall score.
_TIER_WEIGHTS = {1: 0.25, 2: 0.35, 3: 0.40}

_DIMENSION_NAMES = ["correctness", "elegance", "discipline", "judgment", "creativity", "recovery"]


def run_benchmark(data_dir: str, tiers: list[str]) -> None:
    """Main benchmark pipeline: run all tasks x configs, score, and generate results.

    Creates results.html, results.json, and appends to history.jsonl under data_dir.
    """
    data_path = Path(data_dir)
    runs_path = data_path / "runs"
    runs_path.mkdir(parents=True, exist_ok=True)

    # Load control profiles (zero + light) plus any user-supplied profiles.
    controls_dir = Path(__file__).parent.parent / "profiles"
    profiles = load_profiles(controls_dir, user_dir=data_path / "profiles")

    # Load task definitions for the requested tiers.
    tasks = load_tasks(tiers=tiers)

    run_id = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S")

    # Accumulate per-task dimension scores: config -> list of score dicts
    all_task_scores: dict[str, list[dict]] = {name: [] for name in profiles}

    def _run_and_score_profile(task, profile_name, system_prompt):
        """Run one profile on one task and return (profile_name, task_score_dict).

        Inner unit of work used by the per-task ThreadPoolExecutor. Each call
        performs one claude --print invocation for the task run plus three
        more for the judge. Tasks still run sequentially in the outer loop;
        only profiles-within-a-task are parallelized.
        """
        result = run_task(
            task_id=task.id,
            task_spec=task.spec,
            profile_name=profile_name,
            system_prompt=system_prompt,
            working_dir=str(runs_path / run_id),
        )

        auto = score_tests(exit_code=result.exit_code, stdout=result.stdout)
        loc_min, loc_verb = _LOC_REFS.get(task.id, (10, 80))
        loc = score_lines_of_code(result.working_dir, loc_min, loc_verb)
        complexity = score_complexity(result.working_dir)
        scope = score_scope(
            allowed_files=["solution/"],
            files_written=result.files_written,
        )
        judge = score_with_judge(
            session_transcript=result.stdout,
            task_spec=task.spec,
        )

        return profile_name, {
            "task_id": task.id,
            "tier": task.tier,
            "correctness": (auto.tests_pass + judge.requirement_interpretation) / 2,
            "elegance": (loc + complexity) / 2,
            "discipline": scope,
            "judgment": judge.decision_communication,
            "creativity": judge.unconventional_thinking,
            "recovery": judge.recovery_quality,
        }

    for task in tasks:
        # Run all profiles for this task in parallel. Tasks remain sequential
        # so log output stays readable and the orchestrator can still reason
        # about progress one task at a time. Within a task, a 4-profile suite
        # runs in roughly the time of a single profile.
        with ThreadPoolExecutor(max_workers=_MAX_WORKERS) as pool:
            futures = [
                pool.submit(_run_and_score_profile, task, name, prompt)
                for name, prompt in profiles.items()
            ]
            for future in futures:
                try:
                    profile_name, task_score = future.result()
                    all_task_scores[profile_name].append(task_score)
                except Exception as e:
                    # Log and continue — one failed profile should not kill the run
                    print(f"WARNING: profile failed on {task.id}: {e}")

    # Aggregate scores per config across all tasks, weighted by tier.
    scores: dict[str, dict[str, float]] = {}
    dimension_scores: dict[str, dict[str, float]] = {}

    for config, task_score_list in all_task_scores.items():
        tier_averages: dict[int, list[float]] = {1: [], 2: [], 3: []}
        dim_values: dict[str, list[float]] = {d: [] for d in _DIMENSION_NAMES}

        for ts in task_score_list:
            tier = ts["tier"]
            overall = statistics.mean(ts[d] for d in _DIMENSION_NAMES)
            tier_averages[tier].append(overall)
            for dim in _DIMENSION_NAMES:
                dim_values[dim].append(ts[dim])

        tier_scores = {
            t: (statistics.mean(vals) if vals else 0.0)
            for t, vals in tier_averages.items()
        }
        overall = sum(
            tier_scores[t] * _TIER_WEIGHTS[t]
            for t in (1, 2, 3)
        )

        scores[config] = {
            "overall": round(overall, 1),
            "tier1": round(tier_scores[1], 1),
            "tier2": round(tier_scores[2], 1),
            "tier3": round(tier_scores[3], 1),
        }
        dimension_scores[config] = {
            dim: round(statistics.mean(vals), 1) if vals else 0.0
            for dim, vals in dim_values.items()
        }

    # Generate the HTML and JSON results report.
    report = ResultsReport(
        run_id=run_id,
        task_suite_version="v1",
        configurations=list(profiles.keys()),
        scores=scores,
        dimension_scores=dimension_scores,
    )
    generate_results(report, data_path / "results.html", data_path / "results.json")

    # Append overall scores to history for trend tracking.
    append_run(data_path / "history.jsonl", HistoryEntry(
        run_id=run_id,
        scores={c: scores[c]["overall"] for c in scores},
    ))
