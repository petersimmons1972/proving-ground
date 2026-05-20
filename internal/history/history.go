package history

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"syscall"
)

// HistoryEntry is one benchmark run appended to history.jsonl.
type HistoryEntry struct {
	RunID  string             `json:"run_id"`
	Scores map[string]float64 `json:"scores"`
}

// AppendRun appends a run entry to the JSONL history file. Creates if needed.
//
// Concurrency: an exclusive advisory lock (flock LOCK_EX) is held for the
// duration of the write to prevent cross-process interleaving (#37). The
// lock is released automatically when the fd closes, including on panic
// or SIGKILL — no stale-lock cleanup is needed.
//
// Durability: f.Sync() is invoked before Close to flush the write through
// the OS page cache, so a process crash immediately after AppendRun returns
// does not lose the entry (#57). Directory fsync is intentionally omitted
// because this is an append, not a rename — the directory entry already
// existed once the file was created.
//
// AppendRun is the ONLY supported way to write history.jsonl. Direct
// OpenFile callers bypass the advisory lock.
func AppendRun(historyFile string, entry HistoryEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling history entry: %w", err)
	}
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening history file: %w", err)
	}
	defer f.Close()

	// Exclusive advisory lock; blocks if another process holds it.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquiring history file lock: %w", err)
	}
	// Lock released implicitly on Close; no explicit LOCK_UN needed.

	if _, err := fmt.Fprintf(f, "%s\n", b); err != nil {
		return fmt.Errorf("writing history entry: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("syncing history file: %w", err)
	}
	return nil
}

// LoadHistory loads all history entries sorted by run_id ascending.
// Returns nil (not error) if the file does not exist.
//
// A corrupt or torn line (e.g. from a pre-fix concurrent-write race) is
// logged as a warning and skipped, not treated as a fatal error. Aborting
// here would lose all valid history after $15–$40 of API spend has
// already produced a fresh entry that depends on history-based delta
// scoring.
func LoadHistory(historyFile string) ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history file: %w", err)
	}
	var entries []HistoryEntry
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e HistoryEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			log.Printf("warning: history.jsonl line %d unparseable, skipping: %v", i+1, err)
			continue
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RunID < entries[j].RunID
	})
	return entries, nil
}
