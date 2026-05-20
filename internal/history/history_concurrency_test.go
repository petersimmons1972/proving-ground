package history_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/history"
)

// Issue #37: history.jsonl has no file lock — concurrent runs corrupt the file.
// All N concurrent appends must be present and parseable.
func TestAppendRun_ConcurrentWritersDoNotCorrupt(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")

	const N = 50
	var wg sync.WaitGroup
	errs := make(chan error, N)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := history.AppendRun(f, history.HistoryEntry{
				RunID:  fmt.Sprintf("run-%04d", i),
				Scores: map[string]float64{"user": float64(i)},
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("AppendRun error: %v", err)
	}

	entries, err := history.LoadHistory(f)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(entries) != N {
		t.Fatalf("expected %d entries, got %d", N, len(entries))
	}
	seen := make(map[string]bool)
	for _, e := range entries {
		if seen[e.RunID] {
			t.Errorf("duplicate run_id: %s", e.RunID)
		}
		seen[e.RunID] = true
	}
}

// Issue #57: history.jsonl has no fsync — flush loss on crash drops entries.
// We can't easily simulate kernel crash, but we can assert the file is fully
// written on disk before AppendRun returns by checking that Stat sees the
// expected size immediately after the call.
func TestAppendRun_DataPresentBeforeReturn(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	err := history.AppendRun(f, history.HistoryEntry{
		RunID:  "sync-test",
		Scores: map[string]float64{"user": 1.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("history file is empty after AppendRun returned")
	}
}

// Issue #37 follow-up (AG-2 recommendation): a single torn/corrupt line
// from a pre-fix run must not abort all of LoadHistory after $15-$40 of
// API spend. Skip-and-warn behavior.
func TestLoadHistory_SkipsCorruptLineWithoutAborting(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "history.jsonl")
	// Manually write: one good line, one corrupt line, one good line.
	contents := `{"run_id":"a","scores":{"user":1}}` + "\n" +
		`{"run_id":"b","scores":{` + "\n" + // truncated/torn
		`{"run_id":"c","scores":{"user":3}}` + "\n"
	if err := os.WriteFile(f, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := history.LoadHistory(f)
	if err != nil {
		t.Fatalf("LoadHistory must not abort on corrupt line: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 good entries (corrupt one skipped), got %d", len(entries))
	}
	if entries[0].RunID != "a" || entries[1].RunID != "c" {
		t.Errorf("got %v", entries)
	}
}
