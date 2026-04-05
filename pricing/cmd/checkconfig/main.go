// checkconfig — утилита проверки конфига Pricing Engine.
// Запускается отдельно без поднятия HTTP-сервера и без подключения к БД.
//
// Использование:
//
//	DATABASE_URL=postgres://... go run ./cmd/checkconfig
package main

import (
	"fmt"
	"os"

	"github.com/cloud-pricer/pricing/config"
	shared "github.com/cloud-pricer/shared/config"
)

func main() {
	fmt.Println("=== Pricing Engine — проверка конфига ===")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  [FAIL] %v\n\n", err)
		printHints()
		os.Exit(1)
	}

	fmt.Println("  [OK] Конфиг валиден")
	fmt.Println()
	fmt.Println("  Сервер:")
	fmt.Printf("    HTTP_PORT          = %s\n", cfg.HTTPPort)
	fmt.Printf("    HTTP_READ_TIMEOUT  = %s\n", cfg.ReadTimeout)
	fmt.Printf("    HTTP_WRITE_TIMEOUT = %s\n", cfg.WriteTimeout)
	fmt.Printf("    SHUTDOWN_TIMEOUT   = %s\n", cfg.ShutdownTimeout)
	fmt.Println()
	fmt.Println("  Логирование:")
	fmt.Printf("    LOG_LEVEL          = %s\n", cfg.LogLevel)
	fmt.Printf("    LOG_FORMAT         = %s\n", cfg.LogFormat)
	fmt.Printf("    ENVIRONMENT        = %s\n", cfg.Environment)
	fmt.Println()
	fmt.Println("  База данных:")
	fmt.Printf("    DATABASE_URL       = %s\n", shared.MaskPassword(cfg.DatabaseURL))
	fmt.Printf("    MIGRATION_FILE     = %s\n", cfg.MigrationFile)

	if cfg.TracingEndpoint != "" {
		fmt.Printf("\n  TRACING_ENDPOINT   = %s\n", cfg.TracingEndpoint)
	} else {
		fmt.Println("\n  TRACING_ENDPOINT   = (выключен)")
	}

	if cfg.IsDebug() {
		fmt.Println()
		fmt.Println("  [!] Debug-режим активен (ENVIRONMENT=dev или LOG_LEVEL=debug)")
	}
}

func printHints() {
	fmt.Println("  Обязательные переменные:")
	fmt.Println("    DATABASE_URL       — postgres DSN")
	fmt.Println("                        postgres://user:pass@host:5432/dbname?sslmode=disable")
	fmt.Println()
	fmt.Println("  Опциональные (есть дефолты):")
	fmt.Println("    HTTP_PORT          [8081]")
	fmt.Println("    LOG_LEVEL          debug|info|warn|error  [info]")
	fmt.Println("    LOG_FORMAT         text|json               [text]")
	fmt.Println("    ENVIRONMENT        dev|prod                [dev]")
	fmt.Println("    MIGRATION_FILE                             [./migrations/001_init.sql]")
	fmt.Println("    HTTP_READ_TIMEOUT                          [10s]")
	fmt.Println("    HTTP_WRITE_TIMEOUT                         [10s]")
	fmt.Println("    SHUTDOWN_TIMEOUT                           [15s]")
	fmt.Println("    TRACING_ENDPOINT   OTLP endpoint           []")
}
