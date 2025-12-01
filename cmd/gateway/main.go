package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hellocripsis/gold-dust-go/internal/config"
	"github.com/hellocripsis/gold-dust-go/internal/krypton"
)

type HealthResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Krypton krypton.Health     `json:"krypton"`
	Addr    string             `json:"addr"`
	KMode   config.KryptonMode `json:"krypton_mode"`
}

func makeHealthHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := krypton.Fetch(cfg)

		resp := HealthResponse{
			Status:  "ok",
			Message: "gold-dust-go gateway alive",
			Krypton: h,
			Addr:    cfg.Server.Addr,
			KMode:   cfg.Krypton.Mode,
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(resp); err != nil {
			log.Printf("error encoding health response: %v", err)
		}
	}
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", makeHealthHandler(cfg))

	log.Printf("gold-dust-go gateway listening on http://%s", cfg.Server.Addr)

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
