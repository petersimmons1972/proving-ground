package tasks_test

import (
	"testing"

	"github.com/psimmons/proving-ground/internal/tasks"
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
