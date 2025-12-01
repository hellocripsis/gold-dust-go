package krypton

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
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

func stubHealth(source string) Health {
	return Health{
		Samples:  2048,
		Mean:     0.5,
		Variance: 0.003,
		Jitter:   0.05,
		Decision: DecisionKeep,
		Source:   source,
		At:       time.Now().UTC(),
	}
}

func fromPayload(payload map[string]any, source string) Health {
	getFloat := func(key string) float64 {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case float64:
				return t
			case int:
				return float64(t)
			case int64:
				return float64(t)
			}
		}
		return 0.0
	}

	getInt := func(key string) int {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case float64:
				return int(t)
			case int:
				return t
			case int64:
				return int(t)
			}
		}
		return 0
	}

	decision := DecisionKeep
	if v, ok := payload["decision"]; ok {
		if s, ok := v.(string); ok {
			switch s {
			case "Keep", "Throttle", "Kill":
				decision = Decision(s)
			default:
				log.Printf("[krypton] unknown decision %q, defaulting to Keep", s)
			}
		}
	}

	return Health{
		Samples:  getInt("samples"),
		Mean:     getFloat("mean"),
		Variance: getFloat("variance"),
		Jitter:   getFloat("jitter"),
		Decision: decision,
		Source:   source,
		At:       time.Now().UTC(),
	}
}

// fetchHTTP calls an HTTP /health endpoint and parses JSON.
//
// Supported shapes:
//
// 1) Flat JSON:
//    { "samples": ..., "mean": ..., "variance": ..., "jitter": ..., "decision": ... }
//
// 2) Nested JSON (this service or others):
//    { "krypton": { "samples": ..., "mean": ..., "variance": ..., "jitter": ..., "decision": ... }, ... }
func fetchHTTP(cfg config.Config) (Health, error) {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(cfg.Krypton.URL)
	if err != nil {
		return Health{}, err
	}
	defer resp.Body.Close()

	var payload any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Health{}, err
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		return Health{}, ErrBadJSONShape("top-level not an object")
	}

	if inner, ok := obj["krypton"]; ok {
		if m, ok := inner.(map[string]any); ok {
			return fromPayload(m, "http:"+cfg.Krypton.URL), nil
		}
	}

	return fromPayload(obj, "http:"+cfg.Krypton.URL), nil
}

// fetchBinary execs the entropy_health binary and parses JSON.
//
// Assumes the binary prints either a single JSON object or a JSON line
// with the expected fields.
func fetchBinary(cfg config.Config) (Health, error) {
	cmd := exec.Command(cfg.Krypton.BinaryPath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return Health{}, err
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return Health{}, ErrBadJSONShape("empty stdout from entropy_health")
	}

	lastLine := out
	if idx := strings.LastIndex(out, "\n"); idx != -1 {
		lastLine = out[idx+1:]
	}

	var payload any
	if err := json.Unmarshal([]byte(lastLine), &payload); err != nil {
		return Health{}, err
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		return Health{}, ErrBadJSONShape("entropy_health output not an object")
	}

	return fromPayload(obj, "binary:"+cfg.Krypton.BinaryPath), nil
}

// ErrBadJSONShape is a soft error for unexpected JSON structures.
type ErrBadJSONShape string

func (e ErrBadJSONShape) Error() string {
	return "bad JSON shape: " + string(e)
}

// Fetch returns a Health snapshot using the configured Krypton mode.
//
// Modes:
// - none   -> stub
// - http   -> HTTP /health (flat or nested)
// - binary -> exec entropy_health
//
// On any error, logs and returns a stub.
func Fetch(cfg config.Config) Health {
	switch cfg.Krypton.Mode {
	case config.KryptonModeNone:
		return stubHealth("stub:none")

	case config.KryptonModeHTTP:
		h, err := fetchHTTP(cfg)
		if err != nil {
			log.Printf("[krypton] HTTP mode error: %v; falling back to stub", err)
			return stubHealth("stub:http")
		}
		return h

	case config.KryptonModeBinary:
		h, err := fetchBinary(cfg)
		if err != nil {
			log.Printf("[krypton] binary mode error: %v; falling back to stub", err)
			return stubHealth("stub:binary")
		}
		return h

	default:
		log.Printf("[krypton] unknown mode %q; falling back to stub", cfg.Krypton.Mode)
		return stubHealth("stub:unknown")
	}
}
