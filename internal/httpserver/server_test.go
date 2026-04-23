package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ciscoutio/ciscout/internal/config"
)

func TestHealthz(t *testing.T) {
	cfg := &config.Config{Port: 8080, LogLevel: "info", Environment: "dev"}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("healthz: want 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("healthz: want Content-Type application/json, got %s", ct)
	}
}

func TestReadyz(t *testing.T) {
	cfg := &config.Config{Port: 8080, LogLevel: "info", Environment: "dev"}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("readyz: want 200, got %d", w.Code)
	}
}
