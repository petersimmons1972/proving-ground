package tasks_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/tasks"
)

// tasksDir anchored to the worktree root tasks/ directory.
const tasksDir = "../../tasks"

func TestLoadTier1Tasks(t *testing.T) {
	ts, err := tasks.LoadTasks(tasksDir, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 3 {
		t.Errorf("expected 3 tier-1 tasks, got %d", len(ts))
	}
}

func TestTaskHasRequiredFields(t *testing.T) {
	ts, err := tasks.LoadTasks(tasksDir, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	task := ts[0]
	if task.ID == "" {
		t.Error("task.ID is empty")
	}
	if task.Tier != 1 {
		t.Errorf("task.Tier = %d, want 1", task.Tier)
	}
	if task.Title == "" {
		t.Error("task.Title is empty")
	}
	if task.Spec == "" {
		t.Error("task.Spec is empty")
	}
}

func TestTasksOrderedByID(t *testing.T) {
	ts, err := tasks.LoadTasks(tasksDir, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(ts); i++ {
		if ts[i].ID < ts[i-1].ID {
			t.Errorf("tasks not sorted: %s before %s", ts[i-1].ID, ts[i].ID)
		}
	}
}

func TestLoadAllTasks(t *testing.T) {
	ts, err := tasks.LoadTasks(tasksDir, []string{"1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 10 {
		t.Errorf("expected 10 total tasks, got %d", len(ts))
	}
}

// TestLoadTasks_MaliciousID_Sanitized verifies that a task file with a path-traversal
// ID (e.g. "../../etc/passwd") is sanitized via filepath.Base. The resulting ID must
// not contain any path separators — it is either the basename ("passwd") or rejected
// entirely if the basename itself fails the safe-pattern check.
func TestLoadTasks_MaliciousID_Sanitized(t *testing.T) {
	// Build a minimal task tree: <tmp>/tier1/<file>.md
	tmp := t.TempDir()
	tier1 := filepath.Join(tmp, "tier1")
	if err := os.MkdirAll(tier1, 0o755); err != nil {
		t.Fatal(err)
	}

	maliciousContent := "---\nid: ../../etc/passwd\ntitle: Evil Task\n---\n\n# Task\ndo something\n"
	if err := os.WriteFile(filepath.Join(tier1, "evil.md"), []byte(maliciousContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ts, err := tasks.LoadTasks(tmp, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	// The ID must have been sanitized to the basename "passwd" (path traversal stripped),
	// or rejected entirely. Either outcome is safe — no task with the raw traversal ID.
	for _, task := range ts {
		if task.ID == "../../etc/passwd" {
			t.Errorf("path traversal ID was not sanitized: got raw ID %q", task.ID)
		}
		if filepath.IsAbs(task.ID) || strings.Contains(task.ID, "..") || strings.Contains(task.ID, "/") {
			t.Errorf("task ID contains path components after sanitization: %q", task.ID)
		}
	}
	// At most one task can survive (the sanitized basename), and if present its ID
	// must be "passwd" — not a path.
	if len(ts) == 1 && ts[0].ID != "passwd" {
		t.Errorf("expected sanitized ID %q, got %q", "passwd", ts[0].ID)
	}
}

// TestLoadTasks_MaliciousID_WithValidTask verifies that the traversal attempt does not
// block a legitimate task in the same tier directory.
func TestLoadTasks_MaliciousID_WithValidTask(t *testing.T) {
	tmp := t.TempDir()
	tier1 := filepath.Join(tmp, "tier1")
	if err := os.MkdirAll(tier1, 0o755); err != nil {
		t.Fatal(err)
	}

	malicious := "---\nid: ../../etc/passwd\ntitle: Evil\n---\n\n# Task\nbad\n"
	valid := "---\nid: safe-task\ntitle: Safe Task\n---\n\n# Task\ngood\n"
	if err := os.WriteFile(filepath.Join(tier1, "bad.md"), []byte(malicious), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tier1, "good.md"), []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}

	ts, err := tasks.LoadTasks(tmp, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	// "safe-task" must always be present.
	found := false
	for _, task := range ts {
		if task.ID == "safe-task" {
			found = true
		}
		// No task may carry a raw path-traversal ID.
		if strings.Contains(task.ID, "..") || strings.Contains(task.ID, "/") {
			t.Errorf("task ID contains path components: %q", task.ID)
		}
	}
	if !found {
		t.Errorf("safe-task was missing from results (IDs: %v)", taskIDs(ts))
	}
}

func taskIDs(ts []tasks.Task) []string {
	ids := make([]string, len(ts))
	for i, t := range ts {
		ids[i] = t.ID
	}
	return ids
}
