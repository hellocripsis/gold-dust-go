package config

import (
	"log"
	"os"
)

type KryptonMode string

const (
	KryptonModeNone   KryptonMode = "none"
	KryptonModeHTTP   KryptonMode = "http"
	KryptonModeBinary KryptonMode = "binary"
)

type KryptonConfig struct {
	Mode       KryptonMode
	URL        string
	BinaryPath string
}

type ServerConfig struct {
	Addr string
}

type Config struct {
	Server  ServerConfig
	Krypton KryptonConfig
}

// Load returns a Config with simple environment-based overrides.
//
// This is intentionally minimal for MVP:
// - GOLD_DUST_ADDR          -> server address (default "127.0.0.1:8080")
// - GOLD_DUST_KRYPTON_MODE  -> "none" | "http" | "binary" (default "none")
// - GOLD_DUST_KRYPTON_URL   -> HTTP endpoint (default "http://127.0.0.1:3000/health")
// - GOLD_DUST_KRYPTON_BIN   -> path to entropy_health binary (default "entropy_health")
func Load() Config {
	addr := getenvDefault("GOLD_DUST_ADDR", "127.0.0.1:8080")

	modeStr := getenvDefault("GOLD_DUST_KRYPTON_MODE", "none")
	mode := KryptonMode(modeStr)
	switch mode {
	case KryptonModeNone, KryptonModeHTTP, KryptonModeBinary:
	default:
		log.Printf("unknown GOLD_DUST_KRYPTON_MODE=%q, falling back to 'none'", modeStr)
		mode = KryptonModeNone
	}

	kryptonURL := getenvDefault("GOLD_DUST_KRYPTON_URL", "http://127.0.0.1:3000/health")
	kryptonBin := getenvDefault("GOLD_DUST_KRYPTON_BIN", "entropy_health")

	return Config{
		Server: ServerConfig{
			Addr: addr,
		},
		Krypton: KryptonConfig{
			Mode:       mode,
			URL:        kryptonURL,
			BinaryPath: kryptonBin,
		},
	}
}

func getenvDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
