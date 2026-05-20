package results_test

import (
	"encoding/json"
	"testing"

	"github.com/petersimmons1972/proving-ground/internal/results"
)

// Issue #60: TaskFailure.Reason is free-form. Add FailureClass enum
// (timeout|judge_failure|panic|runner_error) so downstream tooling can
// classify failures without brittle string matching against Go error
// messages whose format has no stability guarantee.
func TestTaskFailure_HasFailureClassField(t *testing.T) {
	tf := results.TaskFailure{
		TaskID:       "t1-1",
		Profile:      "zero",
		Reason:       "context deadline exceeded",
		FailureClass: "timeout",
	}
	b, err := json.Marshal(tf)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["failure_class"] != "timeout" {
		t.Errorf("failure_class = %q, want %q", got["failure_class"], "timeout")
	}
}

// Schema bump to v3 in this release.
func TestRunMetadata_SchemaVersionV3(t *testing.T) {
	const want = "3"
	if results.SchemaVersionV3 != want {
		t.Errorf("SchemaVersionV3 = %q, want %q", results.SchemaVersionV3, want)
	}
}
