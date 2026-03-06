package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GOLD_DUST_ADDR", "")
	t.Setenv("GOLD_DUST_KRYPTON_MODE", "")
	t.Setenv("GOLD_DUST_KRYPTON_URL", "")
	t.Setenv("GOLD_DUST_KRYPTON_BIN", "")

	cfg := Load()

	if cfg.Server.Addr != "127.0.0.1:8080" {
		t.Fatalf("expected default addr 127.0.0.1:8080, got %q", cfg.Server.Addr)
	}
	if cfg.Krypton.Mode != KryptonModeNone {
		t.Fatalf("expected default mode %q, got %q", KryptonModeNone, cfg.Krypton.Mode)
	}
	if cfg.Krypton.URL != "http://127.0.0.1:3000/health" {
		t.Fatalf("expected default URL http://127.0.0.1:3000/health, got %q", cfg.Krypton.URL)
	}
	if cfg.Krypton.BinaryPath != "entropy_health" {
		t.Fatalf("expected default binary path entropy_health, got %q", cfg.Krypton.BinaryPath)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("GOLD_DUST_ADDR", "0.0.0.0:9090")
	t.Setenv("GOLD_DUST_KRYPTON_MODE", "binary")
	t.Setenv("GOLD_DUST_KRYPTON_URL", "http://remote:3000/health")
	t.Setenv("GOLD_DUST_KRYPTON_BIN", "/usr/bin/entropy_health")

	cfg := Load()

	if cfg.Server.Addr != "0.0.0.0:9090" {
		t.Fatalf("expected addr 0.0.0.0:9090, got %q", cfg.Server.Addr)
	}
	if cfg.Krypton.Mode != KryptonModeBinary {
		t.Fatalf("expected mode %q, got %q", KryptonModeBinary, cfg.Krypton.Mode)
	}
	if cfg.Krypton.URL != "http://remote:3000/health" {
		t.Fatalf("expected URL http://remote:3000/health, got %q", cfg.Krypton.URL)
	}
	if cfg.Krypton.BinaryPath != "/usr/bin/entropy_health" {
		t.Fatalf("expected binary path /usr/bin/entropy_health, got %q", cfg.Krypton.BinaryPath)
	}
}

func TestLoadAllValidModes(t *testing.T) {
	cases := []struct {
		input    string
		expected KryptonMode
	}{
		{"none", KryptonModeNone},
		{"http", KryptonModeHTTP},
		{"binary", KryptonModeBinary},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Setenv("GOLD_DUST_KRYPTON_MODE", tc.input)
			cfg := Load()
			if cfg.Krypton.Mode != tc.expected {
				t.Fatalf("expected mode %q, got %q", tc.expected, cfg.Krypton.Mode)
			}
		})
	}
}

func TestLoadUnknownModeFallsBackToNone(t *testing.T) {
	t.Setenv("GOLD_DUST_KRYPTON_MODE", "turbo-mode")

	cfg := Load()

	if cfg.Krypton.Mode != KryptonModeNone {
		t.Fatalf("expected unknown mode to fall back to %q, got %q", KryptonModeNone, cfg.Krypton.Mode)
	}
}

func TestGetenvDefaultTreatsEmptyStringAsUnset(t *testing.T) {
	t.Setenv("GOLD_DUST_ADDR", "")

	cfg := Load()

	if cfg.Server.Addr != "127.0.0.1:8080" {
		t.Fatalf("expected empty env var to use default, got %q", cfg.Server.Addr)
	}
}
