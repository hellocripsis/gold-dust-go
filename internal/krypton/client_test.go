package krypton

import (
	"testing"
	"time"

	"github.com/hellocripsis/gold-dust-go/internal/config"
)

func TestFromPayloadParsesFields(t *testing.T) {
	payload := map[string]any{
		"samples":  4096,
		"mean":     0.5123,
		"variance": 0.0042,
		"jitter":   0.031,
		"decision": "Throttle",
	}

	h := fromPayload(payload, "test-source")

	if h.Samples != 4096 {
		t.Fatalf("expected samples 4096, got %d", h.Samples)
	}
	if h.Mean != 0.5123 {
		t.Fatalf("expected mean 0.5123, got %f", h.Mean)
	}
	if h.Variance != 0.0042 {
		t.Fatalf("expected variance 0.0042, got %f", h.Variance)
	}
	if h.Jitter != 0.031 {
		t.Fatalf("expected jitter 0.031, got %f", h.Jitter)
	}
	if h.Decision != DecisionThrottle {
		t.Fatalf("expected decision %q, got %q", DecisionThrottle, h.Decision)
	}
	if h.Source != "test-source" {
		t.Fatalf("expected source test-source, got %q", h.Source)
	}
	if h.At.IsZero() {
		t.Fatalf("expected non-zero timestamp")
	}
}

func TestFromPayloadUnknownDecisionDefaultsToKeep(t *testing.T) {
	payload := map[string]any{
		"samples":  100,
		"mean":     0.5,
		"variance": 0.003,
		"jitter":   0.05,
		"decision": "WeirdDecision",
	}

	h := fromPayload(payload, "test-source")

	if h.Decision != DecisionKeep {
		t.Fatalf("expected unknown decision to default to %q, got %q", DecisionKeep, h.Decision)
	}
}

func TestFetchStubNoneMode(t *testing.T) {
	cfg := config.Config{
		Server: config.ServerConfig{
			Addr: "127.0.0.1:8080",
		},
		Krypton: config.KryptonConfig{
			Mode:       config.KryptonModeNone,
			URL:        "http://127.0.0.1:3000/health",
			BinaryPath: "entropy_health",
		},
	}

	h := Fetch(cfg)

	if h.Source != "stub:none" {
		t.Fatalf("expected source stub:none, got %q", h.Source)
	}
	if h.Decision != DecisionKeep {
		t.Fatalf("expected decision %q in stub none mode, got %q", DecisionKeep, h.Decision)
	}
	if h.Samples <= 0 {
		t.Fatalf("expected positive samples in stub health, got %d", h.Samples)
	}
}

func TestFetchUnknownModeFallsBackToStub(t *testing.T) {
	cfg := config.Config{
		Server: config.ServerConfig{
			Addr: "127.0.0.1:8080",
		},
		Krypton: config.KryptonConfig{
			Mode:       "bogus-mode",
			URL:        "http://127.0.0.1:3000/health",
			BinaryPath: "entropy_health",
		},
	}

	start := time.Now()
	h := Fetch(cfg)

	if h.Source != "stub:unknown" {
		t.Fatalf("expected source stub:unknown for unknown mode, got %q", h.Source)
	}
	if h.At.Before(start) {
		t.Fatalf("expected timestamp At to be set after call start")
	}
}
