package scoring

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var radonRe = regexp.MustCompile(`Average complexity: \w \((\d+\.?\d*)\)`)

// ScoreComplexity scores cyclomatic complexity via radon subprocess.
// Returns 5.0 on timeout/parse-fail (graceful degradation matching Python).
// Returns 6.0 when no functions found (matching Python).
// Returns 5.0 when no .py files present (neutral, matching Python's except clause).
func ScoreComplexity(ctx context.Context, solutionDir string) float64 {
	hasPy, err := hasPythonFiles(solutionDir)
	if err != nil || !hasPy {
		return 5.0
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cmdCtx, "radon", "cc", solutionDir, "-a", "-s").Output()
	if err != nil {
		return 5.0
	}

	m := radonRe.FindStringSubmatch(string(out))
	if m == nil {
		return 6.0 // no functions found
	}

	avg, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 5.0
	}

	switch {
	case avg <= 5:
		return 10.0
	case avg <= 10:
		return 7.5
	case avg <= 15:
		return 5.0
	case avg <= 20:
		return 2.5
	default:
		return 0.0
	}
}

func hasPythonFiles(dir string) (bool, error) {
	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".py") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found, err
}
