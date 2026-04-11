package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var controlNames = map[string]bool{"zero": true, "light": true}

// LoadProfiles loads the zero, light, and any user-supplied profiles.
// controlsDir must contain light.txt.
// userDir may be empty string or non-existent (both are silently skipped).
// Returns map of profile name -> system prompt text.
func LoadProfiles(controlsDir, userDir string) (map[string]string, error) {
	lightPath := filepath.Join(controlsDir, "light.txt")
	lightBytes, err := os.ReadFile(lightPath)
	if err != nil {
		return nil, fmt.Errorf("reading light profile: %w", err)
	}

	profiles := map[string]string{
		"zero":  "",
		"light": strings.TrimSpace(string(lightBytes)),
	}

	if userDir == "" {
		return profiles, nil
	}
	info, err := os.Stat(userDir)
	if err != nil || !info.IsDir() {
		return profiles, nil
	}

	// Collect *.txt and *.md files, sort them for determinism.
	var files []string
	for _, pat := range []string{"*.txt", "*.md"} {
		matches, _ := filepath.Glob(filepath.Join(userDir, pat))
		files = append(files, matches...)
	}
	sort.Strings(files)

	for _, path := range files {
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if controlNames[stem] {
			fmt.Printf("WARNING: skipping user profile '%s' — collides with control baseline\n", filepath.Base(path))
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading user profile %s: %w", path, err)
		}
		profiles[stem] = strings.TrimSpace(string(content))
	}
	return profiles, nil
}
