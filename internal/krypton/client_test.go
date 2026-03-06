package krypton

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestFetchHTTPRejectsNon2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"decision":"Keep"}`))
	}))
	defer srv.Close()

	cfg := config.Config{
		Server: config.ServerConfig{Addr: "127.0.0.1:8080"},
		Krypton: config.KryptonConfig{
			Mode:       config.KryptonModeHTTP,
			URL:        srv.URL,
			BinaryPath: "entropy_health",
		},
	}

	if _, err := fetchHTTP(cfg); err == nil {
		t.Fatalf("expected fetchHTTP to fail on non-2xx status")
	}
}

func TestFetchHTTPRequiresValidExplicitDecision(t *testing.T) {
	t.Run("missing decision is rejected", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"samples":123}`))
		}))
		defer srv.Close()

		cfg := config.Config{
			Server: config.ServerConfig{Addr: "127.0.0.1:8080"},
			Krypton: config.KryptonConfig{
				Mode:       config.KryptonModeHTTP,
				URL:        srv.URL,
				BinaryPath: "entropy_health",
			},
		}

		if _, err := fetchHTTP(cfg); err == nil {
			t.Fatalf("expected fetchHTTP to fail when decision is missing")
		}
	})

	t.Run("invalid decision is rejected", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"decision":"UnknownDecision"}`))
		}))
		defer srv.Close()

		cfg := config.Config{
			Server: config.ServerConfig{Addr: "127.0.0.1:8080"},
			Krypton: config.KryptonConfig{
				Mode:       config.KryptonModeHTTP,
				URL:        srv.URL,
				BinaryPath: "entropy_health",
			},
		}

		if _, err := fetchHTTP(cfg); err == nil {
			t.Fatalf("expected fetchHTTP to fail when decision is invalid")
		}
	})
}

func TestFetchHTTPNestedPayloadParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","krypton":{"samples":99,"mean":0.51,"variance":0.004,"jitter":0.03,"decision":"Throttle"}}`))
	}))
	defer srv.Close()

	cfg := config.Config{
		Server: config.ServerConfig{Addr: "127.0.0.1:8080"},
		Krypton: config.KryptonConfig{
			Mode:       config.KryptonModeHTTP,
			URL:        srv.URL,
			BinaryPath: "entropy_health",
		},
	}

	h, err := fetchHTTP(cfg)
	if err != nil {
		t.Fatalf("unexpected fetchHTTP error: %v", err)
	}
	if h.Decision != DecisionThrottle {
		t.Fatalf("expected decision %q, got %q", DecisionThrottle, h.Decision)
	}
	if h.Samples != 99 {
		t.Fatalf("expected samples 99, got %d", h.Samples)
	}
}

func TestFetchBinaryStrictParsing(t *testing.T) {
	t.Run("missing decision is rejected", func(t *testing.T) {
		// Write a tiny shell script that prints JSON without a decision field.
		script := "#!/bin/sh\necho '{\"samples\":10}'\n"
		f, err := os.CreateTemp(t.TempDir(), "fake-binary-*")
		if err != nil {
			t.Fatalf("could not create temp file: %v", err)
		}
		if _, err := f.WriteString(script); err != nil {
			t.Fatalf("could not write script: %v", err)
		}
		if err := f.Chmod(0o755); err != nil {
			t.Fatalf("could not chmod script: %v", err)
		}
		f.Close()

		cfg := config.Config{
			Server:  config.ServerConfig{Addr: "127.0.0.1:8080"},
			Krypton: config.KryptonConfig{Mode: config.KryptonModeBinary, BinaryPath: f.Name()},
		}

		if _, err := fetchBinary(cfg); err == nil {
			t.Fatalf("expected fetchBinary to fail when decision field is missing")
		}
	})

	t.Run("invalid decision is rejected", func(t *testing.T) {
		script := "#!/bin/sh\necho '{\"decision\":\"Bogus\"}'\n"
		f, err := os.CreateTemp(t.TempDir(), "fake-binary-*")
		if err != nil {
			t.Fatalf("could not create temp file: %v", err)
		}
		if _, err := f.WriteString(script); err != nil {
			t.Fatalf("could not write script: %v", err)
		}
		if err := f.Chmod(0o755); err != nil {
			t.Fatalf("could not chmod script: %v", err)
		}
		f.Close()

		cfg := config.Config{
			Server:  config.ServerConfig{Addr: "127.0.0.1:8080"},
			Krypton: config.KryptonConfig{Mode: config.KryptonModeBinary, BinaryPath: f.Name()},
		}

		if _, err := fetchBinary(cfg); err == nil {
			t.Fatalf("expected fetchBinary to fail when decision is invalid")
		}
	})

	t.Run("valid decision succeeds", func(t *testing.T) {
		for _, dec := range []string{"Keep", "Throttle", "Kill"} {
			dec := dec
			t.Run(dec, func(t *testing.T) {
				script := "#!/bin/sh\necho '{\"samples\":512,\"mean\":0.5,\"variance\":0.003,\"jitter\":0.05,\"decision\":\"" + dec + "\"}'\n"
				f, err := os.CreateTemp(t.TempDir(), "fake-binary-*")
				if err != nil {
					t.Fatalf("could not create temp file: %v", err)
				}
				if _, err := f.WriteString(script); err != nil {
					t.Fatalf("could not write script: %v", err)
				}
				if err := f.Chmod(0o755); err != nil {
					t.Fatalf("could not chmod script: %v", err)
				}
				f.Close()

				cfg := config.Config{
					Server:  config.ServerConfig{Addr: "127.0.0.1:8080"},
					Krypton: config.KryptonConfig{Mode: config.KryptonModeBinary, BinaryPath: f.Name()},
				}

				h, err := fetchBinary(cfg)
				if err != nil {
					t.Fatalf("unexpected fetchBinary error: %v", err)
				}
				if string(h.Decision) != dec {
					t.Fatalf("expected decision %q, got %q", dec, h.Decision)
				}
			})
		}
	})
}

func TestFetchHTTPModeFallsBackToStubOnStrictParseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"samples":10}`))
	}))
	defer srv.Close()

	cfg := config.Config{
		Server: config.ServerConfig{Addr: "127.0.0.1:8080"},
		Krypton: config.KryptonConfig{
			Mode:       config.KryptonModeHTTP,
			URL:        srv.URL,
			BinaryPath: "entropy_health",
		},
	}

	h := Fetch(cfg)
	if h.Source != "stub:http" {
		t.Fatalf("expected Fetch(http) to fall back to stub:http, got %q", h.Source)
	}
	if h.Decision != DecisionKeep {
		t.Fatalf("expected stub decision %q, got %q", DecisionKeep, h.Decision)
	}
}
