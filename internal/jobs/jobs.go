package jobs

import "github.com/hellocripsis/gold-dust-go/internal/krypton"

// JobDecision is the gateway-level decision we expose over HTTP.
type JobDecision string

const (
	DecisionAccepted  JobDecision = "accepted"
	DecisionThrottled JobDecision = "throttled"
	DecisionDenied    JobDecision = "denied"
)

// JobRequest is the JSON shape accepted by POST /jobs.
type JobRequest struct {
	JobID   string         `json:"job_id"`
	Payload map[string]any `json:"payload,omitempty"`
}

// JobResponse is the JSON shape returned by POST /jobs.
type JobResponse struct {
	JobID    string         `json:"job_id"`
	Decision JobDecision    `json:"decision"`
	Krypton  krypton.Health `json:"krypton"`
}

// Decide maps a Krypton decision into a gateway-level job decision.
func Decide(h krypton.Health) JobDecision {
	switch h.Decision {
	case krypton.DecisionKill:
		return DecisionDenied
	case krypton.DecisionThrottle:
		return DecisionThrottled
	case krypton.DecisionKeep:
		fallthrough
	default:
		return DecisionAccepted
	}
}
