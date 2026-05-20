package results

import "testing"

// Issue #49: profile color must be stable across runs even when the set of
// profiles changes. Currently colorFor(name, idx) uses idx → adding a
// profile that sorts before "user" shifts the user color. After fix, the
// color must be derived from the name itself (hash) so adding profiles
// never reassigns colors of existing ones.
func TestColorFor_StableAcrossProfileSetChanges(t *testing.T) {
	// "user" is in defaultConfigColors so it's always F59E0B — pick a name
	// that is NOT in defaults so we exercise the palette fallback.
	want := colorFor("alpha-team", 1)

	// Now imagine a second run where another profile "aardvark" was added
	// and sorts ahead of "alpha-team" — alpha-team's idx is now 2.
	got := colorFor("alpha-team", 2)

	if got != want {
		t.Errorf("color for %q changed when its index shifted: idx=1 got %q, idx=2 got %q",
			"alpha-team", want, got)
	}
}
