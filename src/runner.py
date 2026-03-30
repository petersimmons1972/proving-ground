"""Task runner: launches Claude Code headless for a single task × profile combination.

The A-0 compiler didn't need to be elegant. It needed to work.
Each run gets its own directory so parallel runs don't step on each other.
"""

import subprocess
from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class TaskResult:
    """Structured result from a single headless Claude Code execution."""

    task_id: str
    profile_name: str
    exit_code: int
    stdout: str
    stderr: str
    working_dir: str
    # Files written by Claude during the run
    files_written: list[str] = field(default_factory=list)


def run_task(
    task_id: str,
    task_spec: str,
    profile_name: str,
    system_prompt: str,
    working_dir: str,
    timeout: int = 300,
) -> TaskResult:
    """Run a single task with Claude Code headless.

    Launches `claude --print` with the task spec piped via stdin.
    If system_prompt is non-empty, passes it via --system-prompt.
    Captures stdout, stderr, exit code, and files written to the run directory.

    Each call creates an isolated subdirectory: working_dir/task_id/profile_name.
    """
    run_dir = Path(working_dir) / task_id / profile_name
    run_dir.mkdir(parents=True, exist_ok=True)

    cmd = ["claude", "--print", "--no-interactive"]
    if system_prompt:
        cmd += ["--system-prompt", system_prompt]

    result = subprocess.run(
        cmd,
        input=task_spec,        # pipe spec via stdin — avoids CLI arg length limits
        cwd=str(run_dir),
        capture_output=True,
        text=True,
        timeout=timeout,
    )

    # Collect every file Claude wrote in the run directory.
    files_written = [
        str(f.relative_to(run_dir))
        for f in run_dir.rglob("*")
        if f.is_file()
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
