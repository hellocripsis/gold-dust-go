package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hellocripsis/gold-dust-go/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		Server: config.ServerConfig{
			Addr: "127.0.0.1:8080",
		},
		Krypton: config.KryptonConfig{
			Mode:       config.KryptonModeNone,
			URL:        "http://127.0.0.1:3000/health",
			BinaryPath: "entropy_health",
		},
	}
}

func TestHealthRejectsNonGET(t *testing.T) {
	handler := makeHealthHandler(testConfig())
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response JSON: %v", err)
	}
	if errResp.Error != "method not allowed" {
		t.Fatalf("expected method-not-allowed error message, got %q", errResp.Error)
	}
}

func TestJobsErrorResponsesSetJSONContentType(t *testing.T) {
	handler := makeJobsHandler(testConfig())

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader("{"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
	})

	t.Run("missing job_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"payload":{"a":1}}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
	})
}
