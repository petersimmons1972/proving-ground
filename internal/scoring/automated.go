package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AutomatedScores holds scores derived from running pytest independently.
type AutomatedScores struct {
	TestsPass float64
}

// ScoreTests runs pytest independently against solutionDir using testCmd.
// Agent stdout is ignored — the harness always computes the ground-truth score.
// testCmd is the raw command string from the task spec (e.g. "python -m pytest solution/ --json-report").
// If testCmd is empty, returns 0.0.
func ScoreTests(ctx context.Context, solutionDir, testCmd string) float64 {
	if testCmd == "" {
		return 0.0
	}
	reportFile := filepath.Join(solutionDir, ".pytest_report.json")
	// Append --json-report flags; use %q for shell-safe path quoting.
	shellCmd := fmt.Sprintf("cd %s && %s --json-report --json-report-file=%s",
		shellQuote(solutionDir), testCmd, shellQuote(reportFile))
	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	cmd.Run() // non-zero exit is expected when tests fail — not an error here
	return parsePytestJSONReport(reportFile)
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

// shellQuote wraps s in single quotes, escaping any embedded single quotes.
// This is sufficient for paths passed to "sh -c".
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
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
