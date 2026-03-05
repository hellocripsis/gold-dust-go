package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hellocripsis/gold-dust-go/internal/config"
	"github.com/hellocripsis/gold-dust-go/internal/jobs"
	"github.com/hellocripsis/gold-dust-go/internal/krypton"
)

type HealthResponse struct {
	Status      string             `json:"status"`
	Message     string             `json:"message"`
	Krypton     krypton.Health     `json:"krypton"`
	Addr        string             `json:"addr"`
	KryptonMode config.KryptonMode `json:"krypton_mode"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func makeHealthHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			if err := writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"}); err != nil {
				log.Printf("error encoding health method error response: %v", err)
			}
			return
		}

		h := krypton.Fetch(cfg)

		resp := HealthResponse{
			Status:      "ok",
			Message:     "gold-dust-go gateway alive",
			Krypton:     h,
			Addr:        cfg.Server.Addr,
			KryptonMode: cfg.Krypton.Mode,
		}

		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			log.Printf("error encoding health response: %v", err)
		}
	}
}

func makeJobsHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			if err := writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"}); err != nil {
				log.Printf("jobs: error encoding method error response: %v", err)
			}
			return
		}

		defer r.Body.Close()

		var req jobs.JobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("jobs: bad request body: %v", err)
			if err := writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON body"}); err != nil {
				log.Printf("jobs: error encoding bad request response: %v", err)
			}
			return
		}

		if req.JobID == "" {
			if err := writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "job_id is required"}); err != nil {
				log.Printf("jobs: error encoding validation response: %v", err)
			}
			return
		}

		// Fetch Krypton health and map to a job decision.
		h := krypton.Fetch(cfg)
		decision := jobs.Decide(h)

		resp := jobs.JobResponse{
			JobID:    req.JobID,
			Decision: decision,
			Krypton:  h,
		}

		if err := writeJSON(w, http.StatusOK, resp); err != nil {
			log.Printf("jobs: error encoding response: %v", err)
		}
	}
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", makeHealthHandler(cfg))
	mux.HandleFunc("/jobs", makeJobsHandler(cfg))

	log.Printf("gold-dust-go gateway listening on http://%s (krypton mode: %s)", cfg.Server.Addr, cfg.Krypton.Mode)

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
