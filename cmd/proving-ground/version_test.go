package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Issue #50: proving-ground --version must print the TaskSuiteVersion.
func TestVersionFlagPrintsTaskSuiteVersion(t *testing.T) {
	// Build the binary into the test's temp dir, then exec it with --version.
	tmp := t.TempDir()
	binPath := filepath.Join(tmp, "proving-ground")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	cmd := exec.Command(binPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("--version exited with error: %v\n%s", err, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "v1") {
		t.Errorf("--version output missing task suite version 'v1': %q", got)
	}
}
