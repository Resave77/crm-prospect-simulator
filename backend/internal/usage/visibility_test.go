package usage

import "testing"

func TestHistoryVisibilityDefaultsToActive(t *testing.T) {
	for _, input := range []string{"", "unknown", "active", " ACTIVE "} {
		if got := historyVisibility(input); got != "ACTIVE" {
			t.Fatalf("historyVisibility(%q) = %q, want ACTIVE", input, got)
		}
	}
}

func TestHistoryVisibilityModes(t *testing.T) {
	if got := historyVisibility("hidden"); got != "HIDDEN" {
		t.Fatalf("hidden mode = %q, want HIDDEN", got)
	}
	if got := historyVisibility("ALL"); got != "ALL" {
		t.Fatalf("all mode = %q, want ALL", got)
	}
	if got := visibilityPredicate("e.hidden_at", "ACTIVE"); got != "e.hidden_at IS NULL" {
		t.Fatalf("active predicate = %q", got)
	}
	if got := visibilityPredicate("e.hidden_at", "HIDDEN"); got != "e.hidden_at IS NOT NULL" {
		t.Fatalf("hidden predicate = %q", got)
	}
	if got := visibilityPredicate("e.hidden_at", "ALL"); got != "TRUE" {
		t.Fatalf("all predicate = %q", got)
	}
}
