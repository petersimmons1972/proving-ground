import json
from dataclasses import dataclass, asdict
from pathlib import Path


@dataclass
class HistoryEntry:
    run_id: str
    scores: dict[str, float]


def append_run(history_file: Path, entry: HistoryEntry) -> None:
    """Append a run entry to the JSONL history file. Creates the file if needed."""
    with open(history_file, "a") as f:
        f.write(json.dumps(asdict(entry)) + "\n")


def load_history(history_file: Path) -> list[HistoryEntry]:
    """Load all history entries sorted by run_id ascending."""
    if not history_file.exists():
        return []

    entries: list[HistoryEntry] = []
    for line in history_file.read_text().splitlines():
        line = line.strip()
        if line:
            data = json.loads(line)
            entries.append(HistoryEntry(
                run_id=data["run_id"],
                scores=data["scores"],
            ))

    return sorted(entries, key=lambda e: e.run_id)
