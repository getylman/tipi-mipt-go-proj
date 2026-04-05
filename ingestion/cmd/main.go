package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/cloud-pricer/ingestion/config"
	"github.com/cloud-pricer/ingestion/internal/client"
	"github.com/cloud-pricer/ingestion/internal/handler"
	"github.com/cloud-pricer/ingestion/internal/invalid"
	"github.com/cloud-pricer/ingestion/internal/validator"
	sharedlogger "github.com/cloud-pricer/shared/logger"
	"github.com/cloud-pricer/shared/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := sharedlogger.New(cfg.LogLevel, cfg.LogFormat)
	logger.Info("starting ingestion service", "port", cfg.HTTPPort, "env", cfg.Environment)

	shutdownTracing, _ := tracing.Init("ingestion-service", cfg.TracingEndpoint)
	defer shutdownTracing()

	// БД Ingestion (только для invalid_metrics)
	db, err := connectAndMigrate(cfg.DatabaseURL, cfg.MigrationFile)
	if err != nil {
		log.Fatalf("ingestion db: %v", err)
	}
	defer db.Close()
	logger.Info("ingestion database ready")

	// Зависимости
	invalidRepo   := invalid.NewRepository(db)
	v             := validator.New(cfg.MaxItems)
	pricingClient := client.New(cfg.PricingEngineURL, cfg.PricingEngineTimeout)
	h             := handler.New(logger, v, pricingClient, invalidRepo)

	// Роутинг
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/usage",    h.Usage)
	mux.HandleFunc("POST /v1/estimate", h.Estimate)
	mux.HandleFunc("GET /health",       h.Health)

	addr := ":" + cfg.HTTPPort
	logger.Info("server ready", "addr", addr)
	if err := http.ListenAndServe(addr, loggingMiddleware(logger, mux)); err != nil {
		log.Fatalf("server error: %v", err)
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

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("incoming", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
