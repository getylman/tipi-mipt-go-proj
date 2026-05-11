package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cloud-pricer/ingestion/config"
	"github.com/cloud-pricer/ingestion/internal/client"
	"github.com/cloud-pricer/ingestion/internal/handler"
	"github.com/cloud-pricer/ingestion/internal/invalid"
	"github.com/cloud-pricer/ingestion/internal/validator"
	"github.com/cloud-pricer/shared/logger"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustValidate()

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	log.Info("starting ingestion service", "port", cfg.HTTPPort)

	db, err := connectAndMigrate(cfg.DatabaseURL, cfg.MigrationFile)
	if err != nil {
		log.Error("db", "err", err)
		return
	}
	defer db.Close()
	log.Info("database ready")

	h := handler.New(
		log,
		validator.New(cfg.MaxItems),
		client.New(cfg.PricingEngineURL, cfg.PricingEngineTimeout),
		invalid.NewRepository(db),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/usage", h.Usage)
	mux.HandleFunc("POST /v1/estimate", h.Estimate)
	mux.HandleFunc("GET /health", h.Health)

	log.Info("server ready", "addr", ":"+cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, requestLogger(log, mux)); err != nil {
		log.Error("server", "err", err)
	}
}

func connectAndMigrate(dsn, migrationFile string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	db.SetMaxOpenConns(5)
	data, err := os.ReadFile(migrationFile)
	if err != nil {
		return nil, fmt.Errorf("read migration %q: %w", migrationFile, err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		return nil, fmt.Errorf("run migration: %w", err)
	}
	return db, nil
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
