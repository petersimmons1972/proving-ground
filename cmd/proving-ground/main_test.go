package main

import (
	"os"
	"testing"
)

func TestPreFlight_APIKeyMissing_ReturnsError(t *testing.T) {
	orig, hadKey := os.LookupEnv("ANTHROPIC_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	defer func() {
		if hadKey {
			os.Setenv("ANTHROPIC_API_KEY", orig)
		}
	}()

	err := runPreFlight()
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is not set, got nil")
	}
	if err.Error() != "ANTHROPIC_API_KEY not set" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestPreFlight_ClaudeNotOnPath_ReturnsError(t *testing.T) {
	// Set the API key so we pass the first check.
	orig, hadKey := os.LookupEnv("ANTHROPIC_API_KEY")
	os.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	defer func() {
		if hadKey {
			os.Setenv("ANTHROPIC_API_KEY", orig)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Override PATH to a directory with no claude binary.
	origPath := os.Getenv("PATH")
	emptyDir := t.TempDir()
	os.Setenv("PATH", emptyDir)
	defer os.Setenv("PATH", origPath)

	err := runPreFlight()
	if err == nil {
		t.Fatal("expected error when claude is not on PATH, got nil")
	}
	// Error message should mention "claude CLI not found on PATH"
	if len(err.Error()) == 0 {
		t.Error("error message is empty")
	}
}
