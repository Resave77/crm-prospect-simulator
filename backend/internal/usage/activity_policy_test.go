package usage

import "testing"

func TestShouldPersistActivityPolicy(t *testing.T) {
	tests := []struct {
		name string
		trace map[string]any
		status int
		authenticated bool
		want bool
	}{
		{"meaningful", map[string]any{"action": "SEARCH_PROSPECT"}, 200, true, true},
		{"provider", map[string]any{"provider_attempted": true}, 500, true, true},
		{"cache hit meaningful", map[string]any{"action": "VIEW_PROSPECT_DETAIL", "provider_hit_count": 0}, 200, true, true},
		{"auth me noise", map[string]any{}, 200, true, false},
		{"background noise", map[string]any{}, 200, true, false},
		{"authenticated server error", map[string]any{}, 500, true, true},
		{"client error noise", map[string]any{}, 404, true, false},
		{"security audit", map[string]any{"action": "LOGOUT"}, 204, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldPersistActivity("GET", "/api/v1/auth/me", tt.status, tt.trace, tt.authenticated); got != tt.want {
				t.Fatalf("ShouldPersistActivity()=%v, want %v", got, tt.want)
			}
		})
	}
}
