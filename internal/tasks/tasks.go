package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Task represents one benchmark task.
type Task struct {
	ID    string
	Tier  int
	Title string
	Spec  string // full markdown content including frontmatter
}

// LoadTasks loads task definitions for the requested tiers, sorted by ID.
// tasksDir is the root tasks/ directory (e.g., "/path/to/proving-ground/tasks").
func LoadTasks(tasksDir string, tiers []string) ([]Task, error) {
	var tasks []Task
	for _, tierStr := range tiers {
		tier := 0
		if _, err := fmt.Sscanf(tierStr, "%d", &tier); err != nil {
			return nil, fmt.Errorf("invalid tier %q: %w", tierStr, err)
		}
		tierDir := filepath.Join(tasksDir, fmt.Sprintf("tier%d", tier))
		entries, err := os.ReadDir(tierDir)
		if err != nil {
			return nil, fmt.Errorf("reading tier dir %s: %w", tierDir, err)
		}
		// Collect .md files, sorted.
		var mdFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				mdFiles = append(mdFiles, filepath.Join(tierDir, e.Name()))
			}
		}
		sort.Strings(mdFiles)

		for _, path := range mdFiles {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading task file %s: %w", path, err)
			}
			meta, err := parseFrontmatter(string(content))
			if err != nil {
				return nil, fmt.Errorf("parsing frontmatter in %s: %w", path, err)
			}
			tasks = append(tasks, Task{
				ID:    meta["id"],
				Tier:  tier,
				Title: meta["title"],
				Spec:  string(content),
			})
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

var frontmatterRe = regexp.MustCompile(`(?s)^---\n(.+?)\n---`)

func parseFrontmatter(content string) (map[string]string, error) {
	m := frontmatterRe.FindStringSubmatch(content)
	if m == nil {
		return nil, fmt.Errorf("missing frontmatter")
	}
	meta := map[string]string{}
	for _, line := range strings.Split(m[1], "\n") {
		if idx := strings.Index(line, ":"); idx >= 0 {
			k := strings.TrimSpace(line[:idx])
			v := strings.TrimSpace(line[idx+1:])
			v = strings.Trim(v, `"`)
			meta[k] = v
		}
	}
	return meta, nil
}
