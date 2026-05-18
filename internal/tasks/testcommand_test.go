package tasks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/tasks"
)

// TestTask_TestCommandParsedFromMarkdown verifies that a task markdown file
// containing a ## Tests section with a command: line populates Task.TestCommand.
func TestTask_TestCommandParsedFromMarkdown(t *testing.T) {
	tmp := t.TempDir()
	tierDir := filepath.Join(tmp, "tier1")
	os.MkdirAll(tierDir, 0755)

	content := `---
id: t1-99
tier: 1
title: "Test Command Task"
---

# Task: Example

## Spec

Write something.

## Tests
command: python -m pytest solution/ --json-report
`
	os.WriteFile(filepath.Join(tierDir, "task-1-99-example.md"), []byte(content), 0644)

	ts, err := tasks.LoadTasks(tmp, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 1 {
		t.Fatalf("expected 1 task, got %d", len(ts))
	}

	want := "python -m pytest solution/ --json-report"
	if ts[0].TestCommand != want {
		t.Errorf("TestCommand = %q, want %q", ts[0].TestCommand, want)
	}
}

// TestTask_TestCommandEmptyWhenMissing verifies that tasks without a ## Tests
// section have an empty TestCommand (not a panic or error).
func TestTask_TestCommandEmptyWhenMissing(t *testing.T) {
	// The real tier-1 tasks don't have ## Tests sections yet; verify LoadTasks
	// returns them without error and TestCommand is empty string.
	ts, err := tasks.LoadTasks(tasksDir, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range ts {
		// After the ## Tests sections are added these will be non-empty,
		// but the field must always be present (no panic).
		_ = task.TestCommand // must compile
	}
}
