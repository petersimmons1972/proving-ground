package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidateTestCmd checks that testCmd contains no shell metacharacters that
// could enable injection. Only pytest/python invocations with flags are permitted.
// Rejected: ; & | ` $ > < ! { } ( ) ' " \  (shell control, substitution, and
// word-boundary chars that can break out of quoting contexts).
func ValidateTestCmd(cmd string) error {
	if cmd == "" {
		return nil // empty is handled upstream
	}
	const forbidden = `';|&` + "`" + `$><!()"`
	for _, ch := range forbidden {
		if strings.ContainsRune(cmd, ch) {
			return fmt.Errorf("testCmd contains forbidden shell metacharacter %q: %q", ch, cmd)
		}
	}
	// Check backslash separately — it cannot appear in the raw string literal above.
	if strings.Contains(cmd, `\`) {
		return fmt.Errorf("testCmd contains forbidden character '\\': %q", cmd)
	}
	return nil
}

// ScoreTests runs pytest independently against solutionDir using testCmd.
// Agent stdout is ignored — the harness always computes the ground-truth score.
// testCmd is the raw command string from the task spec (e.g. "python -m pytest solution/").
// If testCmd is empty or contains forbidden characters, returns 0.0.
// No shell is spawned — testCmd is split into argv and executed directly via exec.Command.
func ScoreTests(ctx context.Context, solutionDir, testCmd string) float64 {
	if testCmd == "" {
		return 0.0
	}
	if err := ValidateTestCmd(testCmd); err != nil {
		// Log the rejection; return 0 to fail scoring on malicious input.
		log.Printf("scoring: rejected testCmd: %v", err)
		return 0.0
	}
	reportFile := filepath.Join(solutionDir, ".pytest_report.json")

	// Remove any pre-existing report file. An agent could write a fake
	// perfect-score report before the harness runs; removing it here ensures
	// pytest always writes a fresh result from actual test execution.
	os.Remove(reportFile)

	// Split testCmd into argv — no shell. Append json-report flags.
	parts := strings.Fields(testCmd)
	if len(parts) == 0 {
		return 0.0
	}
	args := append(parts[1:], "--json-report", "--json-report-file="+reportFile)

	cmd := exec.CommandContext(ctx, parts[0], args...)
	cmd.Dir = solutionDir // set working directory directly; no `cd` shell needed
	cmd.Run()             // non-zero exit is expected when tests fail — not an error here

	score := parsePytestJSONReport(reportFile)

	// Remove the harness-generated report so it never appears in FilesWritten
	// and does not penalize the agent's Discipline score.
	os.Remove(reportFile)

	return score
}

// parsePytestJSONReport reads the json-report output and returns a 0–10 pass rate.
// Returns 0.0 if the file is missing, unreadable, or has no tests.
func parsePytestJSONReport(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0.0
	}
	var report struct {
		Summary struct {
			Passed int `json:"passed"`
			Total  int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return 0.0
	}
	if report.Summary.Total == 0 {
		return 0.0
	}
	return roundTo1(10.0 * float64(report.Summary.Passed) / float64(report.Summary.Total))
}

// ScoreLinesOfCode scores solution LOC: at or below minimal=10.0, at or above verbose=0.0.
func ScoreLinesOfCode(solutionDir string, refMinimal, refVerbose int) float64 {
	var pyFiles []string
	_ = filepath.Walk(solutionDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".py") {
			pyFiles = append(pyFiles, path)
		}
		return nil
	})
	if len(pyFiles) == 0 {
		return 5.0
	}

	totalLOC := 0
	for _, f := range pyFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				totalLOC++
			}
		}
	}

	if totalLOC <= refMinimal {
		return 10.0
	}
	if totalLOC >= refVerbose {
		return 0.0
	}
	ratio := float64(totalLOC-refMinimal) / float64(refVerbose-refMinimal)
	return roundTo1((1.0 - ratio) * 10)
}

// ScoreScope scores how many files were written outside allowed prefixes.
func ScoreScope(allowedFiles, filesWritten []string) float64 {
	if len(filesWritten) == 0 {
		return 10.0
	}
	outOfScope := 0
	for _, f := range filesWritten {
		inScope := false
		for _, allowed := range allowedFiles {
			if strings.HasPrefix(f, allowed) {
				inScope = true
				break
			}
		}
		if !inScope {
			outOfScope++
		}
	}
	ratio := float64(outOfScope) / float64(len(filesWritten))
	return roundTo1((1.0 - ratio) * 10)
}

func roundTo1(x float64) float64 {
	return math.Round(x*10) / 10
}
