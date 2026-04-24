package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/ciscoutio/ciscout/internal/config"
)

func openClosedDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("postgres", "postgres://ciscout:ciscout@localhost:5432/ciscout?sslmode=disable")
	if err != nil {
		t.Fatalf("sqlx.Open: %v", err)
	}
	db.Close()
	return db
}

func TestHealthz(t *testing.T) {
	cfg := &config.Config{Port: 8080, LogLevel: "info", Environment: "dev"}
	db := openClosedDB(t)
	srv := New(cfg, db)

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
	t.Run("db unavailable", func(t *testing.T) {
		cfg := &config.Config{Port: 8080, LogLevel: "info", Environment: "dev"}
		db := openClosedDB(t)
		srv := New(cfg, db)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("readyz: want 503, got %d", w.Code)
		}
	})

	t.Run("db available", func(t *testing.T) {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			t.Skip("DATABASE_URL not set")
		}
		cfg := &config.Config{Port: 8080, LogLevel: "info", Environment: "dev"}
		db, err := sqlx.Open("postgres", dbURL)
		if err != nil {
			t.Fatalf("sqlx.Open: %v", err)
		}
		defer db.Close()
		srv := New(cfg, db)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("readyz: want 200, got %d", w.Code)
		}
	})
}
