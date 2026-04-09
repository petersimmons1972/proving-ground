import pytest
from pathlib import Path
from src.profiles import load_profiles

def test_load_zero_profile():
    profiles = load_profiles(Path("profiles"), user_dir=None)
    assert "zero" in profiles
    assert profiles["zero"] == ""

def test_load_light_profile():
    profiles = load_profiles(Path("profiles"), user_dir=None)
    assert "light" in profiles
    assert len(profiles["light"]) > 10

def test_load_user_profile(tmp_path):
    user_profile = tmp_path / "my-agent.txt"
    user_profile.write_text("You are a senior engineer who writes clean, tested code.")
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert "my-agent" in profiles

def test_no_user_profiles_returns_only_controls(tmp_path):
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert set(profiles.keys()) == {"zero", "light"}

def test_user_profile_content_is_loaded(tmp_path):
    text = "You are a senior engineer who writes clean, tested code."
    (tmp_path / "my-agent.txt").write_text(text)
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert profiles["my-agent"] == text.strip()

def test_load_md_user_profile(tmp_path):
    text = "You are a careful engineer who documents decisions."
    (tmp_path / "my-agent.md").write_text(text)
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert "my-agent" in profiles
    assert profiles["my-agent"] == text.strip()


def test_user_profile_cannot_overwrite_controls(tmp_path):
    """User file named 'zero.txt' must not overwrite the zero control baseline."""
    (tmp_path / "zero.txt").write_text("I am not the zero profile")
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    assert profiles["zero"] == ""  # control must be preserved


def test_user_profile_cannot_overwrite_light(tmp_path):
    """User file named 'light.txt' must not overwrite the light control baseline."""
    (tmp_path / "light.txt").write_text("I am not the light profile")
    profiles = load_profiles(Path("profiles"), user_dir=tmp_path)
    # light should still be the control version from profiles/light.txt
    assert profiles["light"] != "I am not the light profile"
