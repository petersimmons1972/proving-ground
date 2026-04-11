package config

import (
	"os"
	"strconv"
)

const TaskSuiteVersion = "v1"

// DimensionNames is the ordered list of scoring dimensions.
var DimensionNames = []string{
	"correctness", "elegance", "discipline", "judgment", "creativity", "recovery",
}

// TierWeights maps tier number to its weight in the overall score.
var TierWeights = map[int]float64{
	1: 0.25,
	2: 0.35,
	3: 0.40,
}

// LocRefs maps task ID to (minimal, verbose) LOC thresholds.
// minimal = fewest non-comment lines a correct solution needs.
// verbose = threshold above which elegance score drops to zero.
var LocRefs = map[string][2]int{
	"t1-1": {15, 80}, "t1-2": {20, 80}, "t1-3": {5, 40},
	"t2-1": {10, 60}, "t2-2": {5, 30}, "t2-3": {10, 50},
	"t3-1": {5, 40}, "t3-2": {30, 120}, "t3-3": {5, 30}, "t3-4": {15, 60},
}

// MaxWorkers returns the max parallel profiles per task.
// Controlled by PROVING_GROUND_MAX_WORKERS env var, default 4.
func MaxWorkers() int {
	if v := os.Getenv("PROVING_GROUND_MAX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 4
}

// JudgeRuns returns how many times to invoke the judge per task.
// Controlled by PROVING_GROUND_JUDGE_RUNS env var, default 5.
func JudgeRuns() int {
	if v := os.Getenv("PROVING_GROUND_JUDGE_RUNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 5
}
