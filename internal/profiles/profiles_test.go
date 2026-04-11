package profiles_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/psimmons/proving-ground/internal/profiles"
)

// controlsDir is the profiles/ directory in the worktree (contains light.txt).
const controlsDir = "../../profiles"

func TestLoadZeroProfile(t *testing.T) {
	p, err := profiles.LoadProfiles(controlsDir, "")
	if err != nil {
		t.Fatal(err)
	}
	if p["zero"] != "" {
		t.Errorf("zero profile should be empty, got %q", p["zero"])
	}
}

func TestLoadLightProfile(t *testing.T) {
	p, err := profiles.LoadProfiles(controlsDir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p["light"]) < 10 {
		t.Errorf("light profile too short: %q", p["light"])
	}
}

func TestLoadUserProfile(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "my-agent.txt"), []byte("You are a senior engineer."), 0644)
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p["my-agent"]; !ok {
		t.Error("user profile 'my-agent' not loaded")
	}
}

func TestNoUserProfilesReturnsOnlyControls(t *testing.T) {
	tmp := t.TempDir()
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 2 {
		t.Errorf("expected 2 profiles, got %d: %v", len(p), p)
	}
}

func TestUserProfileContentLoaded(t *testing.T) {
	tmp := t.TempDir()
	text := "You are a senior engineer who writes clean, tested code."
	os.WriteFile(filepath.Join(tmp, "my-agent.txt"), []byte(text), 0644)
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if p["my-agent"] != text {
		t.Errorf("got %q, want %q", p["my-agent"], text)
	}
}

func TestLoadMdUserProfile(t *testing.T) {
	tmp := t.TempDir()
	text := "You are a careful engineer who documents decisions."
	os.WriteFile(filepath.Join(tmp, "my-agent.md"), []byte(text), 0644)
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if p["my-agent"] != text {
		t.Errorf("got %q, want %q", p["my-agent"], text)
	}
}

func TestUserProfileCannotOverwriteZero(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "zero.txt"), []byte("I am not the zero profile"), 0644)
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if p["zero"] != "" {
		t.Error("zero control was overwritten by user profile")
	}
}

func TestUserProfileCannotOverwriteLight(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "light.txt"), []byte("I am not the light profile"), 0644)
	p, err := profiles.LoadProfiles(controlsDir, tmp)
	if err != nil {
		t.Fatal(err)
	}
	if p["light"] == "I am not the light profile" {
		t.Error("light control was overwritten by user profile")
	}
}
