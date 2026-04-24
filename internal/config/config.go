// Package config provides primitives for loading and parsing
// environment-based configuration for the broker.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port           int `env:"PORT" envDefault:"9092"`
	TimeoutSeconds int `env:"TIMEOUTSECONDS" envDefault:"10"`
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file")
	}
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse .env file content into config struct")
	}
	return cfg, nil
}
