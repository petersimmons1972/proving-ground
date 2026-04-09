import re
from dataclasses import dataclass
from pathlib import Path


@dataclass
class Task:
    id: str
    tier: int
    title: str
    spec: str  # full markdown content including frontmatter


def load_tasks(tiers: list[str]) -> list[Task]:
    """Load task definitions for the specified tiers, sorted by id."""
    tasks = []
    for tier_str in tiers:
        tier = int(tier_str)
        tier_dir = Path(__file__).parent.parent / f"tasks/tier{tier}"
        for f in sorted(tier_dir.glob("*.md")):
            content = f.read_text()
            meta = _parse_frontmatter(content)
            tasks.append(Task(
                id=meta["id"],
                tier=tier,
                title=meta["title"],
                spec=content,
            ))
    return sorted(tasks, key=lambda t: t.id)


def _parse_frontmatter(content: str) -> dict[str, str]:
    """Extract YAML-ish frontmatter from between --- markers."""
    match = re.match(r"^---\n(.+?)\n---", content, re.DOTALL)
    if not match:
        raise ValueError("Missing frontmatter in task file")
    meta: dict[str, str] = {}
    for line in match.group(1).splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            meta[k.strip()] = v.strip().strip('"')
    return meta
