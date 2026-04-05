package config

import (
	"fmt"
	"net/url"
	"time"

	shared "github.com/cloud-pricer/shared/config"
)

type Config struct {
	shared.Base
	PricingEngineURL     string
	PricingEngineTimeout time.Duration
	MaxItems             int
	DatabaseURL          string
	MigrationFile        string
}

func Load() (*Config, error) {
	base, err := shared.LoadBase("8080")
	if err != nil {
		return nil, err
	}
	maxItems, err := shared.GetEnvInt("MAX_ITEMS", 50)
	if err != nil {
		return nil, err
	}
	timeout, err := shared.GetEnvDuration("PRICING_ENGINE_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Base:                 base,
		PricingEngineURL:     shared.GetEnv("PRICING_ENGINE_URL", ""),
		PricingEngineTimeout: timeout,
		MaxItems:             maxItems,
		DatabaseURL:          shared.GetEnv("DATABASE_URL", ""),
		MigrationFile:        shared.GetEnv("MIGRATION_FILE", "./migrations/001_init.sql"),
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if err := shared.ValidateBase(&c.Base); err != nil {
		return err
	}
	if c.PricingEngineURL == "" {
		return fmt.Errorf("PRICING_ENGINE_URL is required")
	}
	if _, err := url.ParseRequestURI(c.PricingEngineURL); err != nil {
		return fmt.Errorf("PRICING_ENGINE_URL is not a valid URL: %w", err)
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if u, err := url.Parse(c.DatabaseURL); err != nil || u.Scheme != "postgres" {
		return fmt.Errorf("DATABASE_URL must be a valid postgres DSN (got %q)", shared.MaskPassword(c.DatabaseURL))
	}
	if c.MaxItems <= 0 || c.MaxItems > 1000 {
		return fmt.Errorf("MAX_ITEMS must be between 1 and 1000 (got %d)", c.MaxItems)
	}
	if c.PricingEngineTimeout < time.Second {
		return fmt.Errorf("PRICING_ENGINE_TIMEOUT must be >= 1s (got %s)", c.PricingEngineTimeout)
	}
	if c.MigrationFile == "" {
		return fmt.Errorf("MIGRATION_FILE is required")
	}
	return nil
}
