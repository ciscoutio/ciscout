package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port        int    `env:"PORT"                 envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL"            envDefault:"info"`
	Environment string `env:"ENVIRONMENT,required"`
	DatabaseURL string `env:"DATABASE_URL,required"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	switch cfg.Environment {
	case "dev", "staging", "prod":
	default:
		return nil, fmt.Errorf("ENVIRONMENT must be dev, staging, or prod; got %q", cfg.Environment)
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}
