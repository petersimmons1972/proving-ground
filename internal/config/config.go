package config

import (
	"log"
	"os"
	"strconv"
	"time"
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
	const defaultVal = 4
	if v := os.Getenv("PROVING_GROUND_MAX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		log.Printf("WARNING: invalid PROVING_GROUND_MAX_WORKERS=%q, using default %d", v, defaultVal)
	}
	return defaultVal
}

// JudgeRuns returns how many times to invoke the judge per task.
// Controlled by PROVING_GROUND_JUDGE_RUNS env var, default 5.
func JudgeRuns() int {
	const defaultVal = 5
	if v := os.Getenv("PROVING_GROUND_JUDGE_RUNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
		log.Printf("WARNING: invalid PROVING_GROUND_JUDGE_RUNS=%q, using default %d", v, defaultVal)
	}
	return defaultVal
}

// TaskTimeout returns the per-task subprocess timeout (#43).
// Controlled by PROVING_GROUND_TASK_TIMEOUT env var (e.g. "180s", "5m"),
// default 300s.
func TaskTimeout() time.Duration {
	const defaultVal = 300 * time.Second
	if v := os.Getenv("PROVING_GROUND_TASK_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		log.Printf("WARNING: invalid PROVING_GROUND_TASK_TIMEOUT=%q, using default %s", v, defaultVal)
	}
	return defaultVal
}

// StdoutCapBytes returns the maximum bytes of subprocess stdout to retain
// in memory after the run completes (#58). Output is streamed to a temp
// file during execution and read back via io.LimitReader at this cap.
// Default 50 MB. Override with PROVING_GROUND_STDOUT_CAP_MB.
func StdoutCapBytes() int64 {
	const defaultMB = 50
	mb := int64(defaultMB)
	if v := os.Getenv("PROVING_GROUND_STDOUT_CAP_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			mb = int64(n)
		} else {
			log.Printf("WARNING: invalid PROVING_GROUND_STDOUT_CAP_MB=%q, using default %d", v, defaultMB)
		}
	}
	return mb * 1024 * 1024
}
