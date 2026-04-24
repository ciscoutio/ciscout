package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		t.Setenv("DATABASE_URL", "postgres://localhost/ciscout")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 8080 {
			t.Errorf("Port: want 8080, got %d", cfg.Port)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("LogLevel: want info, got %s", cfg.LogLevel)
		}
		if cfg.Environment != "dev" {
			t.Errorf("Environment: want dev, got %s", cfg.Environment)
		}
		if cfg.DatabaseURL != "postgres://localhost/ciscout" {
			t.Errorf("DatabaseURL: want postgres://localhost/ciscout, got %s", cfg.DatabaseURL)
		}
	})

	t.Run("env overrides", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("ENVIRONMENT", "staging")
		t.Setenv("DATABASE_URL", "postgres://prod-db/ciscout")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 9090 {
			t.Errorf("Port: want 9090, got %d", cfg.Port)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("LogLevel: want debug, got %s", cfg.LogLevel)
		}
		if cfg.Environment != "staging" {
			t.Errorf("Environment: want staging, got %s", cfg.Environment)
		}
		if cfg.DatabaseURL != "postgres://prod-db/ciscout" {
			t.Errorf("DatabaseURL: want postgres://prod-db/ciscout, got %s", cfg.DatabaseURL)
		}
	})

	t.Run("invalid environment", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "local")
		_, err := Load()
		if err == nil {
			t.Error("expected error for invalid ENVIRONMENT, got nil")
		}
	})

	t.Run("missing environment", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "")
		t.Setenv("DATABASE_URL", "postgres://localhost/ciscout")
		_, err := Load()
		if err == nil {
			t.Error("expected error for missing ENVIRONMENT, got nil")
		}
	})

	t.Run("missing database url", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
		t.Setenv("DATABASE_URL", "")
		_, err := Load()
		if err == nil {
			t.Error("expected error for missing DATABASE_URL, got nil")
		}
	})
}
