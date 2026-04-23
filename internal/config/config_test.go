package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "dev")
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
	})

	t.Run("env overrides", func(t *testing.T) {
		t.Setenv("PORT", "9090")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("ENVIRONMENT", "staging")
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
		_, err := Load()
		if err == nil {
			t.Error("expected error for missing ENVIRONMENT, got nil")
		}
	})
}
