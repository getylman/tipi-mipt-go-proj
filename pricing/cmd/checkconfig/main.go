package main

import (
	"fmt"
	"os"

	"github.com/cloud-pricer/pricing/config"
	shared "github.com/cloud-pricer/shared/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n\n", err)
		fmt.Print(`required:
  DATABASE_URL   postgres://user:pass@host:5432/db?sslmode=disable

optional (defaults):
  HTTP_PORT          [8081]
  LOG_LEVEL          debug|info|warn|error  [info]
  LOG_FORMAT         text|json              [text]
  ENVIRONMENT        dev|prod               [dev]
  MIGRATION_FILE                            [./migrations/001_init.sql]
  SHUTDOWN_TIMEOUT                          [15s]
`)
		os.Exit(1)
	}

	fmt.Printf("pricing engine config OK\n\n  HTTP_PORT = %s  LOG = %s/%s  ENV = %s\n  DATABASE_URL = %s\n  MIGRATION_FILE = %s\n  SHUTDOWN_TIMEOUT = %s\n",
		cfg.HTTPPort,
		cfg.LogLevel, cfg.LogFormat, cfg.Environment,
		shared.MaskPassword(cfg.DatabaseURL),
		cfg.MigrationFile,
		cfg.ShutdownTimeout,
	)
}
