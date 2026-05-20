package scoring

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
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
	// Failed and FailureReason are internal sentinel fields — not serialized to JSON.
	// When Failed is true, all dimension scores are zero and must NOT be averaged
	// into benchmark results as if they were real scores.
	Failed        bool   `json:"-"`
	FailureReason string `json:"-"`
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

// BuildJudgeArgs returns the command-line arguments for a judge subprocess invocation.
// The judge is a read-only evaluator: it reads a transcript and outputs scores.
// It deliberately omits --dangerously-skip-permissions because feeding untrusted
// agent transcripts to a process with shell-execute permissions creates a
// prompt-injection risk (B-4, #21).
// When PROVING_GROUND_MODEL is set, --model <value> is appended so the judge
// uses the same model as the runner.
func BuildJudgeArgs(rubric string) []string {
	args := []string{
		"--print",
		"--no-session-persistence",
		"--system-prompt", rubric,
	}
	// Allow callers to pin a specific Claude model via environment variable.
	// Omitting the flag leaves the CLI to use its own default — no breaking change.
	if model := os.Getenv("PROVING_GROUND_MODEL"); model != "" {
		args = append(args, "--model", model)
	}
	return args
}

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

	for i := range runs {
		runCtx, cancel := context.WithTimeout(ctx, judgeTimeout)
		args := BuildJudgeArgs(rubric)
		cmd := exec.CommandContext(runCtx, "claude", args...)
		cmd.Stdin = strings.NewReader(prompt)
		// #58: stream stdout to temp file rather than unbounded bytes.Buffer.
		stdoutFile, ferr := os.CreateTemp("", "pg-judge-stdout-*")
		if ferr != nil {
			log.Printf("judge: run %d/%d temp file error: %v", i+1, runs, ferr)
			cancel()
			continue
		}
		var stderr bytes.Buffer
		cmd.Stdout = stdoutFile
		cmd.Stderr = &stderr

		runErr := cmd.Run()
		cancel()

		// Read back with cap.
		stdoutCapBytes := int64(50 * 1024 * 1024)
		stdoutText := ""
		if info, ierr := stdoutFile.Stat(); ierr == nil {
			limit := info.Size()
			if limit > stdoutCapBytes {
				limit = stdoutCapBytes
				log.Printf("judge: run %d/%d stdout TRUNCATED at %d bytes (original %d)",
					i+1, runs, stdoutCapBytes, info.Size())
			}
			stdoutFile.Seek(0, 0)
			buf := make([]byte, limit)
			io.ReadFull(stdoutFile, buf)
			stdoutText = string(buf)
		}
		stdoutFile.Close()
		os.Remove(stdoutFile.Name())

		if runCtx.Err() == context.DeadlineExceeded || runErr != nil {
			stderrStr := stderr.String()
			if stderrStr != "" {
				log.Printf("judge: run %d/%d failed (exit_code=%v): stderr=%q",
					i+1, runs, runErr, truncateJudge(stderrStr, 300))
			} else {
				log.Printf("judge: run %d/%d failed: %v", i+1, runs, runErr)
			}
			continue
		}

parsed := parseJudgeOutput(stdoutText)
		if !isCompleteJudgeRun(parsed) {
			log.Printf("judge: run %d/%d dropped — incomplete output (missing dimensions); not contributing to median",
				i+1, runs)
			continue
		}
		allScores = append(allScores, parsed)
	}

	if len(allScores) < runs && len(allScores) > 0 {
		log.Printf("judge: WARNING degraded quality — %d/%d runs succeeded; scores based on partial sample",
			len(allScores), runs)
	}

	// Quorum: at least ceil(runs/2) usable observations required (AG-4).
	quorum := (runs + 1) / 2
	if len(allScores) < quorum {
		reason := "ALL_JUDGE_RUNS_FAILED"
		if len(allScores) > 0 {
			reason = "INSUFFICIENT_JUDGE_QUORUM"
		}
		return &JudgeScores{
			Rationale:     reason,
			Failed:        true,
			FailureReason: reason,
		}, nil
	}

	return medianJudgeScores(allScores), nil
}

// parseJudgeOutput returns the dims that were present in the judge output.
// Missing dims are NOT defaulted to 5 (#39, AG-4 fail-the-run): a truncated
// judge response is one event, not four independent observations, so dropping
// the whole run downstream is more honest than averaging real scores with
// fabricated neutral defaults.
func parseJudgeOutput(text string) judgeRaw {
	dims := map[string]int{}
	for name, re := range judgeRe {
		m := re.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		v, _ := strconv.Atoi(m[1])
		if v < 0 {
			v = 0
		}
		if v > 10 {
			v = 10
		}
		dims[name] = v
	}
	rationale := ""
	if m := rationaleRe.FindStringSubmatch(text); m != nil {
		rationale = strings.TrimSpace(m[1])
	}
	return judgeRaw{dims: dims, rationale: rationale}
}

// isCompleteJudgeRun returns true iff all four scoring dimensions were
// parsed from the judge output. A partial run is dropped from the median
// pool entirely (AG-4 fail-the-run policy).
func isCompleteJudgeRun(r judgeRaw) bool {
	for name := range judgeRe {
		if _, ok := r.dims[name]; !ok {
			return false
		}
	}
	return true
}

// truncateJudge returns s truncated to n bytes with a suffix if it was cut.
// Used to keep log lines readable when judge subprocess stderr is verbose.
func truncateJudge(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
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
