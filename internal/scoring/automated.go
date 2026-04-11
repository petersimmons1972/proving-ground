package scoring

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// AutomatedScores holds scores derived from pytest output.
type AutomatedScores struct {
	TestsPass float64
}

var passedRe = regexp.MustCompile(`(\d+) passed`)
var failedRe = regexp.MustCompile(`(\d+) failed`)

// ScoreTests parses pytest stdout and returns a 0–10 pass rate score.
func ScoreTests(exitCode int, stdout string) AutomatedScores {
	passed := extractCount(passedRe, stdout)
	failed := extractCount(failedRe, stdout)
	total := passed + failed
	var rate float64
	if total > 0 {
		rate = float64(passed) / float64(total)
	}
	return AutomatedScores{TestsPass: roundTo1(rate * 10)}
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

func extractCount(re *regexp.Regexp, text string) int {
	m := re.FindStringSubmatch(text)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func roundTo1(x float64) float64 {
	return math.Round(x*10) / 10
}
