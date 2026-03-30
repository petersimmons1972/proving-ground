from pathlib import Path
from typing import Optional


def load_profiles(controls_dir: Path, user_dir: Optional[Path]) -> dict:
    """Load zero, light, and any user-supplied profiles.

    Returns a dict mapping profile name to system prompt text.
    Zero profile is always an empty string.
    Light profile is the contents of controls_dir/light.txt.
    User profiles are loaded from user_dir (*.txt and *.md files).
    """
    profiles: dict = {
        "zero": "",
        "light": (controls_dir / "light.txt").read_text().strip(),
    }
    if user_dir and user_dir.exists():
        for f in sorted(user_dir.glob("*.txt")) + sorted(user_dir.glob("*.md")):
            profiles[f.stem] = f.read_text().strip()
    return profiles
