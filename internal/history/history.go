package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// HistoryEntry is one benchmark run appended to history.jsonl.
type HistoryEntry struct {
	RunID  string             `json:"run_id"`
	Scores map[string]float64 `json:"scores"`
}

// AppendRun appends a run entry to the JSONL history file. Creates if needed.
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
	_, err = fmt.Fprintf(f, "%s\n", b)
	return err
}

// LoadHistory loads all history entries sorted by run_id ascending.
// Returns nil (not error) if the file does not exist.
func LoadHistory(historyFile string) ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history file: %w", err)
	}
	var entries []HistoryEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e HistoryEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("parsing history line: %w", err)
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RunID < entries[j].RunID
	})
	return entries, nil
}
