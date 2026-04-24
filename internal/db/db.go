package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(ctx context.Context, url string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", url)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return db, nil
}
