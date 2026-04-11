package results

// TestColorForUnknownConfigIsNotGray verifies that unknown config names receive
// palette-based colors (not the old #888888 gray fallback) and that two
// different unknown names at different indices get distinct colors.
import "testing"

func TestColorForUnknownConfigIsNotGray(t *testing.T) {
	c0 := colorFor("qa-army-group", 0)
	c1 := colorFor("grace-hopper", 1)

	if c0 == "#888888" {
		t.Errorf("colorFor(%q, 0) = %q, must not be #888888 (old gray fallback)", "qa-army-group", c0)
	}
	if c1 == "#888888" {
		t.Errorf("colorFor(%q, 1) = %q, must not be #888888 (old gray fallback)", "grace-hopper", c1)
	}
	if c0 == c1 {
		t.Errorf("colorFor(%q, 0) == colorFor(%q, 1) = %q; want distinct colors", "qa-army-group", "grace-hopper", c0)
	}
}

func TestColorForKnownConfigsPreservesBrandColors(t *testing.T) {
	// Brand colors must be preserved exactly — do not change these values.
	cases := []struct {
		name  string
		want  string
	}{
		{"zero", "#64748b"},
		{"light", "#22D3EE"},
		{"user", "#F59E0B"},
	}
	for _, tc := range cases {
		got := colorFor(tc.name, 99) // idx should be ignored for known names
		if got != tc.want {
			t.Errorf("colorFor(%q, 99) = %q, want %q", tc.name, got, tc.want)
		}
	}
}
