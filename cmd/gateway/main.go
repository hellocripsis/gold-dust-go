package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hellocripsis/gold-dust-go/internal/config"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:  "ok",
		Message: "gold-dust-go gateway alive",
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		log.Printf("error encoding health response: %v", err)
	}
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	log.Printf("gold-dust-go gateway listening on http://%s", cfg.Server.Addr)

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
