package history_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/psimmons/proving-ground/internal/history"
)

func TestAppendCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	err := history.AppendRun(f, history.HistoryEntry{
		RunID:  "2026-03-30T12:00:00",
		Scores: map[string]float64{"zero": 4.2, "user": 8.3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f); err != nil {
		t.Error("history file not created")
	}
}

func TestAppendWritesValidJSON(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	history.AppendRun(f, history.HistoryEntry{
		RunID:  "2026-03-30T12:00:00",
		Scores: map[string]float64{"user": 8.3},
	})
	data, _ := os.ReadFile(f)
	var e history.HistoryEntry
	if err := json.Unmarshal(data[:len(data)-1], &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.RunID != "2026-03-30T12:00:00" {
		t.Errorf("run_id = %q", e.RunID)
	}
	if e.Scores["user"] != 8.3 {
		t.Errorf("score = %f", e.Scores["user"])
	}
}

func TestAppendMultipleRuns(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	for i := 0; i < 3; i++ {
		history.AppendRun(f, history.HistoryEntry{
			RunID:  fmt.Sprintf("run-%d", i),
			Scores: map[string]float64{"user": float64(i + 5)},
		})
	}
	entries, err := history.LoadHistory(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLoadHistorySortedByRunID(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	for _, id := range []string{"2026-03-31", "2026-03-29", "2026-03-30"} {
		history.AppendRun(f, history.HistoryEntry{RunID: id, Scores: map[string]float64{"user": 7.0}})
	}
	entries, err := history.LoadHistory(f)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].RunID != "2026-03-29" || entries[2].RunID != "2026-03-31" {
		t.Errorf("wrong order: %v", entries)
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	entries, err := history.LoadHistory(filepath.Join(tmp, "nonexistent.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d entries", len(entries))
	}
}

func TestHistoryEntryPreservesAllScores(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	scores := map[string]float64{"zero": 4.2, "light": 6.1, "user": 8.3}
	history.AppendRun(f, history.HistoryEntry{RunID: "2026-03-30", Scores: scores})
	entries, _ := history.LoadHistory(f)
	for k, v := range scores {
		if entries[0].Scores[k] != v {
			t.Errorf("score[%s] = %f, want %f", k, entries[0].Scores[k], v)
		}
	}
}
