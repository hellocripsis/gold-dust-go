package krypton

import (
	"time"

	"github.com/hellocripsis/gold-dust-go/internal/config"
)

type Decision string

const (
	DecisionKeep     Decision = "Keep"
	DecisionThrottle Decision = "Throttle"
	DecisionKill     Decision = "Kill"
)

// Health mirrors the basic Krypton entropy_health JSON shape.
type Health struct {
	Samples  int       `json:"samples"`
	Mean     float64   `json:"mean"`
	Variance float64   `json:"variance"`
	Jitter   float64   `json:"jitter"`
	Decision Decision  `json:"decision"`
	Source   string    `json:"source"`
	At       time.Time `json:"at"`
}

// Fetch returns a Health snapshot.
//
// For MVP this is a stub that ignores the Krypton config and returns
// a constant "Keep" decision. Later we can:
//
// - If cfg.Krypton.Mode == KryptonModeHTTP:  call the HTTP /health endpoint.
// - If cfg.Krypton.Mode == KryptonModeBinary: exec entropy_health.
// - If cfg.Krypton.Mode == KryptonModeNone:  keep using a local stub.
func Fetch(cfg config.Config) Health {
	return Health{
		Samples:  2048,
		Mean:     0.5,
		Variance: 0.003,
		Jitter:   0.05,
		Decision: DecisionKeep,
		Source:   string(cfg.Krypton.Mode),
		At:       time.Now().UTC(),
	}
}
