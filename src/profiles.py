from pathlib import Path


def load_profiles(controls_dir: Path, user_dir: Path | None) -> dict[str, str]:
    """Load zero, light, and any user-supplied profiles.

    Returns a dict mapping profile name to system prompt text.
    Zero profile is always an empty string.
    Light profile is the contents of controls_dir/light.txt.
    User profiles are loaded from user_dir (*.txt and *.md files).
    Control name collisions (zero.txt, light.txt) are skipped with a warning.
    """
    _CONTROL_NAMES = {"zero", "light"}

    profiles: dict = {
        "zero": "",
        "light": (controls_dir / "light.txt").read_text().strip(),
    }
    if user_dir and user_dir.exists():
        for f in sorted(user_dir.glob("*.txt")) + sorted(user_dir.glob("*.md")):
            if f.stem in _CONTROL_NAMES:
                print(f"WARNING: skipping user profile '{f.name}' — collides with control baseline")
                continue
            profiles[f.stem] = f.read_text().strip()
    return profiles
