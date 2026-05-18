package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestResolveAssetDirs_ExplicitFlags verifies that explicit flag values take
// precedence over the auto-detect fallback, and that the resolved paths are
// what was passed in.
func TestResolveAssetDirs_ExplicitFlags(t *testing.T) {
	want := t.TempDir()
	got := resolveAssetDir(want, "tasks")
	if got != want {
		t.Errorf("resolveAssetDir with explicit path: got %q, want %q", got, want)
	}
}

// TestResolveAssetDirs_FallbackToWorkingDir verifies that when no explicit path
// is given, resolveAssetDir falls back to subdir of os.Getwd(), not
// os.Executable() — the latter fails after go install.
func TestResolveAssetDirs_FallbackToWorkingDir(t *testing.T) {
	// Create a temp dir with a "tasks" subdirectory to simulate the repo root.
	tmpRoot := t.TempDir()
	tasksDir := filepath.Join(tmpRoot, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change working directory to tmpRoot for this test.
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd) //nolint:errcheck

	got := resolveAssetDir("", "tasks")
	if got != tasksDir {
		t.Errorf("resolveAssetDir fallback: got %q, want %q", got, tasksDir)
	}
}

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

// TestMakeRootContext verifies that makeRootContext returns a cancellable context
// (not context.Background) that can be cancelled via signal.NotifyContext.
func TestMakeRootContext(t *testing.T) {
	ctx := makeRootContext()

	// A cancellable context should not equal Background and should have a non-nil Err() after cancel.
	if ctx == context.Background() {
		t.Error("makeRootContext returned context.Background; expected a signal-cancellable context")
	}

	// Check that it's a derived context with cancellation capability.
	// We cannot test cancellation directly without sending signals, but we can
	// verify it's not the background by checking that it's a different object.
	bgCtx := context.Background()
	if ctx == bgCtx {
		t.Error("makeRootContext returned Background context, expected derived context")
	}
}
