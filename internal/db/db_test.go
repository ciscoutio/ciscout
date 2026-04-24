package db

import (
	"context"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping DB connection test")
	}

	ctx := context.Background()
	db, err := New(ctx, dbURL)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Errorf("PingContext: unexpected error: %v", err)
	}
}

func TestNewInvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := New(ctx, "postgres://invalid:invalid@nonexistent:5432/nonexistent")
	if err == nil {
		t.Error("New: expected error for invalid URL, got nil")
	}
}
