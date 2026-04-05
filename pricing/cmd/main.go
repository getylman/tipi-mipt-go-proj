package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/cloud-pricer/pricing/config"
	"github.com/cloud-pricer/pricing/db/sqlc"
	"github.com/cloud-pricer/pricing/internal/estimate"
	"github.com/cloud-pricer/pricing/internal/handler"
	"github.com/cloud-pricer/pricing/internal/repository"
	"github.com/cloud-pricer/pricing/internal/usage"
	sharedlogger "github.com/cloud-pricer/shared/logger"
	"github.com/cloud-pricer/shared/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := sharedlogger.New(cfg.LogLevel, cfg.LogFormat)
	logger.Info("starting pricing engine", "port", cfg.HTTPPort, "env", cfg.Environment)

	shutdownTracing, _ := tracing.Init("pricing-engine", cfg.TracingEndpoint)
	defer shutdownTracing()

	// Подключение к БД и миграция
	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	if err := repository.Migrate(db, cfg.MigrationFile); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	logger.Info("database ready")

	// Создаём sqlc.Queries — единственный объект знающий о SQL
	q := sqlcdb.New(db)

	// Репозитории — тонкая обёртка над sqlc без SQL строк
	productRepo     := repository.NewProductRepository(q)
	userRepo        := repository.NewUserRepository(q)
	consumptionRepo := repository.NewConsumptionRepository(db, q) // db нужен для Begin()

	// Сервисы — бизнес-логика
	usageSvc    := usage.NewService(productRepo, userRepo, consumptionRepo)
	estimateSvc := estimate.NewService(productRepo)

	// Handler
	h := handler.New(logger, usageSvc, estimateSvc, productRepo)

	// Роутинг
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/usage",    h.Usage)
	mux.HandleFunc("POST /v1/estimate", h.Estimate)
	mux.HandleFunc("GET /v1/products",  h.ListProducts)
	mux.HandleFunc("GET /health",       h.Health)

	addr := ":" + cfg.HTTPPort
	logger.Info("server ready", "addr", addr)
	if err := http.ListenAndServe(addr, loggingMiddleware(logger, mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("incoming", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
