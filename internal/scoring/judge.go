package scoring

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const judgeTimeout = 300 * time.Second

// JudgeScores holds the four LLM-as-judge dimension scores.
type JudgeScores struct {
	RequirementInterpretation float64
	DecisionCommunication     float64
	RecoveryQuality           float64
	UnconventionalThinking    float64
	Rationale                 string
}

// judgeRaw holds one parsed judge run result.
type judgeRaw struct {
	dims      map[string]int
	rationale string
}

var judgeRe = map[string]*regexp.Regexp{
	"REQUIREMENT_INTERPRETATION": regexp.MustCompile(`REQUIREMENT_INTERPRETATION:\s*(\d+)`),
	"DECISION_COMMUNICATION":     regexp.MustCompile(`DECISION_COMMUNICATION:\s*(\d+)`),
	"RECOVERY_QUALITY":           regexp.MustCompile(`RECOVERY_QUALITY:\s*(\d+)`),
	"UNCONVENTIONAL_THINKING":    regexp.MustCompile(`UNCONVENTIONAL_THINKING:\s*(\d+)`),
}

var rationaleRe = regexp.MustCompile(`RATIONALE:\s*(.+)`)

// ScoreWithJudge invokes claude as LLM judge `runs` times and returns median scores.
// rubricPath is the absolute path to prompts/judge-rubric.md.
func ScoreWithJudge(ctx context.Context, transcript, taskSpec, rubricPath string, runs int) (*JudgeScores, error) {
	rubricBytes, err := os.ReadFile(rubricPath)
	if err != nil {
		return nil, fmt.Errorf("reading judge rubric: %w", err)
	}
	rubric := string(rubricBytes)

	prompt := fmt.Sprintf("## Task Spec\n\n%s\n\n## Agent Session\n\n%s", taskSpec, transcript)

	var allScores []judgeRaw

	for i := 0; i < runs; i++ {
		runCtx, cancel := context.WithTimeout(ctx, judgeTimeout)
		cmd := exec.CommandContext(runCtx, "claude",
			"--print",
			"--dangerously-skip-permissions",
			"--no-session-persistence",
			"--system-prompt", rubric,
		)
		cmd.Stdin = strings.NewReader(prompt)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		cancel()

		if runCtx.Err() == context.DeadlineExceeded || runErr != nil {
			continue // skip failed run
		}

		allScores = append(allScores, parseJudgeOutput(stdout.String()))
	}

	if len(allScores) == 0 {
		return &JudgeScores{
			RequirementInterpretation: 0,
			DecisionCommunication:     0,
			RecoveryQuality:           0,
			UnconventionalThinking:    0,
			Rationale:                 "ALL_JUDGE_RUNS_FAILED",
		}, nil
	}

	return medianJudgeScores(allScores), nil
}

func parseJudgeOutput(text string) judgeRaw {
	dims := map[string]int{}
	for name, re := range judgeRe {
		m := re.FindStringSubmatch(text)
		if m != nil {
			v, _ := strconv.Atoi(m[1])
			if v < 0 {
				v = 0
			}
			if v > 10 {
				v = 10
			}
			dims[name] = v
		} else {
			dims[name] = 5 // default to neutral when dimension is missing
		}
	}
	rationale := ""
	if m := rationaleRe.FindStringSubmatch(text); m != nil {
		rationale = strings.TrimSpace(m[1])
	}
	return judgeRaw{dims: dims, rationale: rationale}
}

func medianInt(vals []int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]int, len(vals))
	copy(sorted, vals)
	sort.Ints(sorted)
	n := len(sorted)
	if n%2 == 1 {
		return float64(sorted[n/2])
	}
	return float64(sorted[n/2-1]+sorted[n/2]) / 2.0
}

func medianJudgeScores(all []judgeRaw) *JudgeScores {
	extract := func(dim string) []int {
		vals := make([]int, len(all))
		for i, s := range all {
			vals[i] = s.dims[dim]
		}
		return vals
	}

	// Sort by REQUIREMENT_INTERPRETATION ascending and take the middle element's
	// rationale — so the rationale comes from the median-scoring run, not always run 0.
	sorted := make([]judgeRaw, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].dims["REQUIREMENT_INTERPRETATION"] < sorted[j].dims["REQUIREMENT_INTERPRETATION"]
	})
	medianRationale := sorted[len(sorted)/2].rationale

	return &JudgeScores{
		RequirementInterpretation: medianInt(extract("REQUIREMENT_INTERPRETATION")),
		DecisionCommunication:     medianInt(extract("DECISION_COMMUNICATION")),
		RecoveryQuality:           medianInt(extract("RECOVERY_QUALITY")),
		UnconventionalThinking:    medianInt(extract("UNCONVENTIONAL_THINKING")),
		Rationale:                 medianRationale,
	}
}
