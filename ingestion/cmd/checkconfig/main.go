package main

import (
	"fmt"
	"os"

	"github.com/cloud-pricer/ingestion/config"
	shared "github.com/cloud-pricer/shared/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n\n", err)
		fmt.Print(`required:
  DATABASE_URL         postgres://user:pass@host:5433/db?sslmode=disable
  PRICING_ENGINE_URL   http://localhost:8081

optional (defaults):
  HTTP_PORT                [8080]
  LOG_LEVEL                debug|info|warn|error  [info]
  LOG_FORMAT               text|json              [text]
  ENVIRONMENT              dev|prod               [dev]
  MAX_ITEMS                1..1000                [50]
  PRICING_ENGINE_TIMEOUT   >=1s                   [5s]
  MIGRATION_FILE                                  [./migrations/001_init.sql]
  SHUTDOWN_TIMEOUT                                [15s]
`)
		os.Exit(1)
	}

	fmt.Printf("ingestion service config OK\n\n  HTTP_PORT = %s  LOG = %s/%s  ENV = %s\n  PRICING_ENGINE_URL = %s  (timeout: %s)\n  MAX_ITEMS = %d\n  DATABASE_URL = %s\n  MIGRATION_FILE = %s\n  SHUTDOWN_TIMEOUT = %s\n",
		cfg.HTTPPort,
		cfg.LogLevel, cfg.LogFormat, cfg.Environment,
		cfg.PricingEngineURL, cfg.PricingEngineTimeout,
		cfg.MaxItems,
		shared.MaskPassword(cfg.DatabaseURL),
		cfg.MigrationFile,
		cfg.ShutdownTimeout,
	)
}
