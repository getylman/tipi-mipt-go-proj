// Package tracing provides OpenTelemetry initialisation.
// Текущая реализация — no-op заглушка без внешних зависимостей.
// Для продакшна установите:
//   go get go.opentelemetry.io/otel
//   go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
//   go get go.opentelemetry.io/otel/sdk
// и раскомментируйте полную реализацию.
package tracing

// Init инициализирует трейсинг.
// Если endpoint пустой или трейсинг не подключён — возвращает no-op.
func Init(serviceName, endpoint string) (shutdown func(), err error) {
	_ = serviceName
	_ = endpoint
	return func() {}, nil
}
