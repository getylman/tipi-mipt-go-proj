package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/cloud-pricer/pricing/config"
	sqlcdb "github.com/cloud-pricer/pricing/db/sqlc"
	"github.com/cloud-pricer/pricing/internal/estimate"
	"github.com/cloud-pricer/pricing/internal/handler"
	"github.com/cloud-pricer/pricing/internal/repository"
	"github.com/cloud-pricer/pricing/internal/usage"
	"github.com/cloud-pricer/shared/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	log.Info("starting pricing engine", "port", cfg.HTTPPort)

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Error("db connect", "err", err)
		return
	}
	defer db.Close()

	if err := repository.Migrate(db, cfg.MigrationFile); err != nil {
		log.Error("migration", "err", err)
		return
	}
	log.Info("database ready")

	q := sqlcdb.New(db)
	productRepo := repository.NewProductRepository(q)
	userRepo := repository.NewUserRepository(q)
	consumptionRepo := repository.NewConsumptionRepository(db, q)

	usageSvc := usage.NewService(productRepo, userRepo, consumptionRepo)
	estimateSvc := estimate.NewService(productRepo)
	h := handler.New(log, usageSvc, estimateSvc, productRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/usage", h.Usage)
	mux.HandleFunc("POST /v1/estimate", h.Estimate)
	mux.HandleFunc("GET /v1/products", h.ListProducts)
	mux.HandleFunc("GET /health", h.Health)

	log.Info("server ready", "addr", ":"+cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, requestLogger(log, mux)); err != nil {
		log.Error("server", "err", err)
	}
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
