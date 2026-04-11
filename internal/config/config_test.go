package config_test

import (
	"os"
	"testing"

	"github.com/psimmons/proving-ground/internal/config"
)

func TestDefaults(t *testing.T) {
	os.Unsetenv("PROVING_GROUND_MAX_WORKERS")
	os.Unsetenv("PROVING_GROUND_JUDGE_RUNS")

	if got := config.MaxWorkers(); got != 4 {
		t.Errorf("MaxWorkers() = %d, want 4", got)
	}
	if got := config.JudgeRuns(); got != 5 {
		t.Errorf("JudgeRuns() = %d, want 5", got)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("PROVING_GROUND_MAX_WORKERS", "8")
	t.Setenv("PROVING_GROUND_JUDGE_RUNS", "3")

	if got := config.MaxWorkers(); got != 8 {
		t.Errorf("MaxWorkers() = %d, want 8", got)
	}
	if got := config.JudgeRuns(); got != 3 {
		t.Errorf("JudgeRuns() = %d, want 3", got)
	}
}

func TestTierWeights(t *testing.T) {
	total := config.TierWeights[1] + config.TierWeights[2] + config.TierWeights[3]
	if total < 0.999 || total > 1.001 {
		t.Errorf("TierWeights sum = %f, want 1.0", total)
	}
}

func TestLocRefsHasAllTasks(t *testing.T) {
	expected := []string{"t1-1", "t1-2", "t1-3", "t2-1", "t2-2", "t2-3", "t3-1", "t3-2", "t3-3", "t3-4"}
	for _, id := range expected {
		if _, ok := config.LocRefs[id]; !ok {
			t.Errorf("LocRefs missing task %s", id)
		}
	}
}

func TestDimensionNames(t *testing.T) {
	if len(config.DimensionNames) != 6 {
		t.Errorf("DimensionNames len = %d, want 6", len(config.DimensionNames))
	}
}
