// check-config — утилита проверки конфигов обоих сервисов.
// Запускается командой: make check-config
// Читает те же ENV-переменные что и сами сервисы,
// выполняет Load() + Validate() и печатает отчёт.
// Не запускает HTTP-серверы и не подключается к БД.
package main

import (
	"fmt"
	"os"
	"strings"

	ingestionConfig "github.com/cloud-pricer/ingestion/config"
	pricingConfig "github.com/cloud-pricer/pricing/config"
)

// результат проверки одного сервиса
type result struct {
	service string
	fields  []field
	err     error
}

type field struct {
	key      string
	value    string
	source   string // "env" или "default"
	redacted bool   // true для паролей/DSN
}

func main() {
	fmt.Println()
	fmt.Println("  Cloud Pricer — проверка конфигов")
	fmt.Println("  " + strings.Repeat("─", 46))

	results := []result{
		checkIngestion(),
		checkPricing(),
	}

	allOK := true
	for _, r := range results {
		printResult(r)
		if r.err != nil {
			allOK = false
		}
	}

	fmt.Println("  " + strings.Repeat("─", 46))
	if allOK {
		fmt.Println("  Результат: все конфиги валидны ✓")
		fmt.Println()
		os.Exit(0)
	} else {
		fmt.Println("  Результат: найдены ошибки ✗")
		fmt.Println()
		os.Exit(1)
	}
}

// checkIngestion проверяет конфиг Ingestion Service.
func checkIngestion() result {
	r := result{service: "Ingestion Service (:8080)"}

	// Собираем поля до валидации — чтобы показать их даже при ошибке
	r.fields = []field{
		envField("HTTP_PORT", "8080"),
		envField("LOG_LEVEL", "info"),
		envField("LOG_FORMAT", "text"),
		envField("ENVIRONMENT", "dev"),
		envField("PRICING_ENGINE_URL", "http://localhost:8081"),
		envField("PRICING_ENGINE_TIMEOUT", "5s"),
		envField("MAX_RESOURCES", "50"),
	}

	// Запускаем полную Load() + Validate()
	_, err := ingestionConfig.Load()
	r.err = err
	return r
}

// checkPricing проверяет конфиг Pricing Engine.
func checkPricing() result {
	r := result{service: "Pricing Engine (:8081)"}

	backend := getEnvOr("STORE_BACKEND", "file")

	r.fields = []field{
		envField("HTTP_PORT", "8081"),
		envField("LOG_LEVEL", "info"),
		envField("LOG_FORMAT", "text"),
		envField("ENVIRONMENT", "dev"),
		envField("STORE_BACKEND", "file"),
	}

	// Показываем только релевантные поля в зависимости от бэкенда
	switch backend {
	case "file":
		r.fields = append(r.fields, envField("PRICES_FILE_PATH", "./prices/prices.json"))
	case "postgres":
		r.fields = append(r.fields, redactedField("DATABASE_URL"))
	default:
		r.fields = append(r.fields,
			envField("PRICES_FILE_PATH", "./prices/prices.json"),
			redactedField("DATABASE_URL"),
		)
	}

	_, err := pricingConfig.Load()
	r.err = err
	return r
}

// ── Форматирование вывода ─────────────────────────────────────────

func printResult(r result) {
	fmt.Println()

	// Заголовок сервиса
	status := "✓"
	if r.err != nil {
		status = "✗"
	}
	fmt.Printf("  %s  %s\n", status, r.service)
	fmt.Println()

	// Таблица полей
	for _, f := range r.fields {
		val := f.value
		if f.redacted && val != "" && val != "(не задан)" {
			val = redact(val)
		}

		src := ""
		if f.source == "default" {
			src = " (default)"
		}

		// Выравнивание: ключ занимает 28 символов
		fmt.Printf("    %-28s %s%s\n", f.key, val, src)
	}

	// Ошибка валидации
	if r.err != nil {
		fmt.Println()
		fmt.Printf("    ОШИБКА: %s\n", r.err)
	}
}

// ── Хелперы ───────────────────────────────────────────────────────

func envField(key, defaultVal string) field {
	v := os.Getenv(key)
	if v == "" {
		return field{key: key, value: defaultVal, source: "default"}
	}
	return field{key: key, value: v, source: "env"}
}

func redactedField(key string) field {
	v := os.Getenv(key)
	if v == "" {
		return field{key: key, value: "(не задан)", source: "default", redacted: true}
	}
	return field{key: key, value: v, source: "env", redacted: true}
}

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// redact скрывает пароль в DSN: postgres://user:PASS@host/db
func redact(dsn string) string {
	// Ищем "://" и следующий "@" — между ними user:pass
	start := strings.Index(dsn, "://")
	if start == -1 {
		return "***"
	}
	rest := dsn[start+3:]
	atIdx := strings.LastIndex(rest, "@")
	if atIdx == -1 {
		return dsn[:start+3] + "***"
	}
	userpass := rest[:atIdx]
	colonIdx := strings.Index(userpass, ":")
	if colonIdx == -1 {
		return dsn
	}
	user := userpass[:colonIdx]
	host := rest[atIdx:]
	return dsn[:start+3] + user + ":***" + host
}
