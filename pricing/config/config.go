package config

import (
	"fmt"
	"net/url"

	shared "github.com/cloud-pricer/shared/config"
)

type Config struct {
	shared.Base
	DatabaseURL   string
	MigrationFile string
}

func Load() (*Config, error) {
	base, err := shared.LoadBase("8081")
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Base:          base,
		DatabaseURL:   shared.GetEnv("DATABASE_URL", ""),
		MigrationFile: shared.GetEnv("MIGRATION_FILE", "./migrations/001_init.sql"),
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if err := shared.ValidateBase(&c.Base); err != nil {
		return err
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if u, err := url.Parse(c.DatabaseURL); err != nil || u.Scheme != "postgres" {
		return fmt.Errorf("DATABASE_URL must be a valid postgres DSN (got %q)", shared.MaskPassword(c.DatabaseURL))
	}
	if c.MigrationFile == "" {
		return fmt.Errorf("MIGRATION_FILE is required")
	}
	return nil
}
