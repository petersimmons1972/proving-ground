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
