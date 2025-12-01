package jobs

import (
	"testing"

	"github.com/hellocripsis/gold-dust-go/internal/krypton"
)

func TestDecideMapsKryptonDecision(t *testing.T) {
	tests := []struct {
		name     string
		input    krypton.Health
		expected JobDecision
	}{
		{
			name: "Kill becomes denied",
			input: krypton.Health{
				Decision: krypton.DecisionKill,
			},
			expected: DecisionDenied,
		},
		{
			name: "Throttle becomes throttled",
			input: krypton.Health{
				Decision: krypton.DecisionThrottle,
			},
			expected: DecisionThrottled,
		},
		{
			name: "Keep becomes accepted",
			input: krypton.Health{
				Decision: krypton.DecisionKeep,
			},
			expected: DecisionAccepted,
		},
		{
			name: "Unknown decision defaults to accepted",
			input: krypton.Health{
				Decision: krypton.Decision("weird"),
			},
			expected: DecisionAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.input)
			if got != tt.expected {
				t.Fatalf("Decide(%v) = %q, want %q", tt.input.Decision, got, tt.expected)
			}
		})
	}
}
