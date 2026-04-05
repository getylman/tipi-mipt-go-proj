.PHONY: run stop build tidy up down sqlc-generate \
        test test-pricing test-ingestion test-integration test-coverage \
        check-config check-config-ingestion check-config-pricing \
        curl-usage curl-estimate curl-invalid curl-products curl-health \
        logs-pricing logs-ingestion \
        db-pricing db-ingestion db-consumption db-invalid db-products \
        help

# ── Настройки ────────────────────────────────────────────────────

INGESTION_PORT ?= 8080
PRICING_PORT   ?= 8081

DB_USER      ?= pricer
DB_PASS      ?= pricer
DB_PRICING   ?= pricing
DB_INGESTION ?= ingestion

DATABASE_URL_PRICING   = postgres://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_PRICING)?sslmode=disable
DATABASE_URL_INGESTION = postgres://$(DB_USER):$(DB_PASS)@localhost:5433/$(DB_INGESTION)?sslmode=disable

DOCKER ?= sudo docker

BIN_DIR       = .bin
LOG_DIR       = .logs
PID_DIR       = .pids
BIN_PRICING   = $(BIN_DIR)/pricing
BIN_INGESTION = $(BIN_DIR)/ingestion
PID_PRICING   = $(PID_DIR)/pricing.pid
PID_INGESTION = $(PID_DIR)/ingestion.pid
LOG_PRICING   = $(LOG_DIR)/pricing.log
LOG_INGESTION = $(LOG_DIR)/ingestion.log

# ── Локальный запуск ──────────────────────────────────────────────

## Собрать и запустить локально (поднимет два Docker-контейнера с БД)
run: build _dirs _kill-old _db-start _db-wait _run-pricing _run-ingestion _status

_kill-old:
	@echo "  Проверка занятых портов..."
	@lsof -t -i :$(PRICING_PORT)   2>/dev/null | xargs kill 2>/dev/null || true
	@lsof -t -i :$(INGESTION_PORT) 2>/dev/null | xargs kill 2>/dev/null || true
	@rm -f $(PID_PRICING) $(PID_INGESTION)
	@sleep 1

## Остановить оба сервиса (БД не трогает)
stop: _stop-pricing _stop-ingestion
	@echo ""
	@echo "  Сервисы остановлены."
	@echo "  Остановить БД: $(DOCKER) rm -f pricer-postgres-pricing pricer-postgres-ingestion"

_db-start:
	@echo "  Запуск PostgreSQL (pricing)..."
	@$(DOCKER) run -d --name pricer-postgres-pricing \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASS) \
		-e POSTGRES_DB=$(DB_PRICING) \
		-p 5432:5432 \
		postgres:16-alpine 2>/dev/null \
		&& echo "  PostgreSQL (pricing) запущен" \
		|| echo "  PostgreSQL (pricing) уже запущен"
	@echo "  Запуск PostgreSQL (ingestion)..."
	@$(DOCKER) run -d --name pricer-postgres-ingestion \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASS) \
		-e POSTGRES_DB=$(DB_INGESTION) \
		-p 5433:5432 \
		postgres:16-alpine 2>/dev/null \
		&& echo "  PostgreSQL (ingestion) запущен" \
		|| echo "  PostgreSQL (ingestion) уже запущен"

_db-wait:
	@echo "  Ожидание готовности БД..."
	@sleep 3
	@for i in $$(seq 1 30); do \
		$(DOCKER) exec pricer-postgres-pricing   pg_isready -U $(DB_USER) -q 2>/dev/null && \
		$(DOCKER) exec pricer-postgres-ingestion pg_isready -U $(DB_USER) -q 2>/dev/null && \
		echo "  БД готовы" && exit 0; \
		echo "  Попытка $$i/30..."; \
		sleep 2; \
	done; \
	echo "  ОШИБКА: БД не ответила за 60 секунд."; \
	echo "  Проверь: $(DOCKER) logs pricer-postgres-pricing"; \
	exit 1

_run-pricing:
	@echo "  Запуск Pricing Engine на :$(PRICING_PORT)..."
	@HTTP_PORT=$(PRICING_PORT) \
	 DATABASE_URL="$(DATABASE_URL_PRICING)" \
	 MIGRATION_FILE=./pricing/migrations/001_init.sql \
	 LOG_LEVEL=debug \
	 LOG_FORMAT=text \
	 ENVIRONMENT=dev \
	 $(BIN_PRICING) >> $(LOG_PRICING) 2>&1 & echo $$! > $(PID_PRICING)
	@sleep 2
	@if kill -0 $$(cat $(PID_PRICING)) 2>/dev/null; then \
		echo "    OK (PID $$(cat $(PID_PRICING)))"; \
	else \
		echo "    ОШИБКА: Pricing Engine не запустился."; \
		tail -30 $(LOG_PRICING); \
		exit 1; \
	fi

_run-ingestion:
	@echo "  Запуск Ingestion Service на :$(INGESTION_PORT)..."
	@HTTP_PORT=$(INGESTION_PORT) \
	 PRICING_ENGINE_URL=http://localhost:$(PRICING_PORT) \
	 DATABASE_URL="$(DATABASE_URL_INGESTION)" \
	 MIGRATION_FILE=./ingestion/migrations/001_init.sql \
	 LOG_LEVEL=debug \
	 LOG_FORMAT=text \
	 ENVIRONMENT=dev \
	 $(BIN_INGESTION) >> $(LOG_INGESTION) 2>&1 & echo $$! > $(PID_INGESTION)
	@sleep 2
	@if kill -0 $$(cat $(PID_INGESTION)) 2>/dev/null; then \
		echo "    OK (PID $$(cat $(PID_INGESTION)))"; \
	else \
		echo "    ОШИБКА: Ingestion Service не запустился."; \
		tail -30 $(LOG_INGESTION); \
		exit 1; \
	fi

_stop-pricing:
	@if [ -f $(PID_PRICING) ]; then \
		PID=$$(cat $(PID_PRICING)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "  Pricing Engine (PID $$PID) остановлен"; \
		else \
			echo "  Pricing Engine уже не запущен"; \
		fi; \
		rm -f $(PID_PRICING); \
	else \
		echo "  Pricing Engine не запущен"; \
	fi

_stop-ingestion:
	@if [ -f $(PID_INGESTION) ]; then \
		PID=$$(cat $(PID_INGESTION)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID && echo "  Ingestion Service (PID $$PID) остановлен"; \
		else \
			echo "  Ingestion Service уже не запущен"; \
		fi; \
		rm -f $(PID_INGESTION); \
	else \
		echo "  Ingestion Service не запущен"; \
	fi

_status:
	@echo ""
	@echo "  Сервисы запущены:"
	@echo "    Ingestion  →  http://localhost:$(INGESTION_PORT)"
	@echo "    Pricing    →  http://localhost:$(PRICING_PORT)"
	@echo ""
	@echo "  make curl-usage      ручка 1: расчёт и сохранение"
	@echo "  make curl-estimate   ручка 2: только калькулятор"
	@echo "  make curl-invalid    ручка 3: невалидный запрос"
	@echo "  make stop            остановить"

_dirs:
	@mkdir -p $(BIN_DIR) $(LOG_DIR) $(PID_DIR)
	@> $(LOG_PRICING)
	@> $(LOG_INGESTION)

# ── Проверка конфига ─────────────────────────────────────────────

## Проверить конфиг обоих сервисов
check-config: check-config-ingestion check-config-pricing

## Проверить конфиг Ingestion Service
check-config-ingestion:
	@DATABASE_URL="$(DATABASE_URL_INGESTION)" \
	 PRICING_ENGINE_URL=http://localhost:$(PRICING_PORT) \
	 cd ingestion && go run ./cmd/checkconfig

## Проверить конфиг Pricing Engine
check-config-pricing:
	@DATABASE_URL="$(DATABASE_URL_PRICING)" \
	 cd pricing && go run ./cmd/checkconfig

# ── Docker Compose ────────────────────────────────────────────────

## Поднять всё в Docker Compose
up:
	$(DOCKER) compose up --build

## Остановить и удалить тома
down:
	$(DOCKER) compose down -v

# ── Сборка ───────────────────────────────────────────────────────

## Собрать оба бинарника в .bin/
build:
	@echo "  Сборка Pricing Engine..."
	@mkdir -p $(BIN_DIR)
	@cd pricing   && go build -o ../$(BIN_PRICING)   ./cmd/main.go
	@echo "  Сборка Ingestion Service..."
	@cd ingestion && go build -o ../$(BIN_INGESTION) ./cmd/main.go
	@echo "  Готово."

## Обновить Go-зависимости
tidy:
	@cd shared    && go mod tidy
	@cd ingestion && GOPROXY=direct GONOSUMDB='*' go mod tidy
	@cd pricing   && GOPROXY=direct GONOSUMDB='*' go mod tidy

# ── sqlc ──────────────────────────────────────────────────────────

## Перегенерировать Go-код из SQL файлов
## Установка: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc-generate:
	@echo "  Генерация кода из db/queries/*.sql..."
	@cd pricing && sqlc generate
	@echo "  Готово. Проверьте pricing/db/sqlc/"

# ── Тестовые запросы ─────────────────────────────────────────────

## Ручка 1: расчёт и сохранение затрат пользователя
curl-usage:
	@curl -s -X POST http://localhost:$(INGESTION_PORT)/v1/usage \
	  -H "Content-Type: application/json" \
	  -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[{"product_id":"vcpu","quantity":4},{"product_id":"ram_gb","quantity":16},{"product_id":"disk_gb","quantity":100}]}' \
	  | python3 -m json.tool

## Ручка 2: только калькулятор, без сохранения
curl-estimate:
	@curl -s -X POST http://localhost:$(INGESTION_PORT)/v1/estimate \
	  -H "Content-Type: application/json" \
	  -d '{"items":[{"product_id":"vcpu","quantity":4},{"product_id":"ram_gb","quantity":16},{"product_id":"disk_gb","quantity":100}]}' \
	  | python3 -m json.tool

## Невалидный запрос — проверить запись в invalid_metrics
curl-invalid:
	@curl -s -X POST http://localhost:$(INGESTION_PORT)/v1/usage \
	  -H "Content-Type: application/json" \
	  -d '{"user_id":"","items":[]}' \
	  | python3 -m json.tool

## Справочник продуктов с ценами
curl-products:
	@curl -s http://localhost:$(PRICING_PORT)/v1/products | python3 -m json.tool

## Проверить /health обоих сервисов
curl-health:
	@echo "--- Ingestion ---"
	@curl -s http://localhost:$(INGESTION_PORT)/health
	@echo ""
	@echo "--- Pricing ---"
	@curl -s http://localhost:$(PRICING_PORT)/health
	@echo ""

# ── Логи ──────────────────────────────────────────────────────────

## Следить за логом Pricing Engine (Ctrl+C для выхода)
logs-pricing:
	@tail -f $(LOG_PRICING)

## Следить за логом Ingestion Service (Ctrl+C для выхода)
logs-ingestion:
	@tail -f $(LOG_INGESTION)

# ── Просмотр БД ───────────────────────────────────────────────────

## psql в БД Pricing Engine
db-pricing:
	@$(DOCKER) exec -it pricer-postgres-pricing psql -U $(DB_USER) -d $(DB_PRICING)

## psql в БД Ingestion
db-ingestion:
	@$(DOCKER) exec -it pricer-postgres-ingestion psql -U $(DB_USER) -d $(DB_INGESTION)

## История потребления (последние 20 записей)
db-consumption:
	@$(DOCKER) exec pricer-postgres-pricing psql -U $(DB_USER) -d $(DB_PRICING) -c \
	  "SELECT user_id, product_id, quantity, unit_price, total_price, calculated_at \
	   FROM consumption ORDER BY calculated_at DESC LIMIT 20;"

## Невалидные запросы (последние 20)
db-invalid:
	@$(DOCKER) exec pricer-postgres-ingestion psql -U $(DB_USER) -d $(DB_INGESTION) -c \
	  "SELECT id, error_reason, received_at, raw_payload \
	   FROM invalid_metrics ORDER BY received_at DESC LIMIT 20;"

## Справочник продуктов
db-products:
	@$(DOCKER) exec pricer-postgres-pricing psql -U $(DB_USER) -d $(DB_PRICING) -c \
	  "SELECT id, name, price_per_unit, unit FROM products ORDER BY id;"


# ── Тесты ────────────────────────────────────────────────────────

## Запустить все юнит-тесты (без сервисов и БД)
test:
	@echo "  Запуск юнит-тестов..."
	@cd pricing   && go test ./internal/... -v
	@cd ingestion && go test ./internal/... -v
	@echo ""
	@echo "  Все тесты пройдены."

## Запустить юнит-тесты Pricing Engine
test-pricing:
	@cd pricing && go test ./internal/... -v

## Запустить юнит-тесты Ingestion Service
test-ingestion:
	@cd ingestion && go test ./internal/... -v

## Запустить интеграционные тесты (требуют make run)
test-integration:
	@echo "  Интеграционные тесты (сервисы должны быть запущены: make run)"
	@cd pricing && go test ./tests/... -v

## Все тесты с покрытием
test-coverage:
	@echo "  Покрытие Pricing Engine:"
	@cd pricing   && go test ./internal/... -coverprofile=coverage.out && go tool cover -func=coverage.out
	@echo "  Покрытие Ingestion Service:"
	@cd ingestion && go test ./internal/... -coverprofile=coverage.out && go tool cover -func=coverage.out

# ── Помощь ────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "  Запуск:"
	@echo "    make run                    локально (Docker только для БД)"
	@echo "    make stop                   остановить сервисы"
	@echo "    make up                     Docker Compose (всё в контейнерах)"
	@echo "    make down                   остановить Docker Compose"
	@echo ""
	@echo "  Если docker требует sudo:"
	@echo "    make run DOCKER='sudo docker'"
	@echo "    или: sudo usermod -aG docker \$$USER  (затем перелогиниться)"
	@echo ""
	@echo "  Проверка конфига:"
	@echo "    make check-config           проверить оба сервиса"
	@echo "    make check-config-pricing   только Pricing Engine"
	@echo "    make check-config-ingestion только Ingestion Service"
	@echo ""
	@echo "  Разработка:"
	@echo "    make build                  собрать бинарники"
	@echo "    make tidy                   обновить зависимости"
	@echo "    make sqlc-generate          перегенерировать код из SQL"
	@echo ""
	@echo "  Тесты:"
	@echo "    make curl-usage             ручка 1: расчёт + сохранение"
	@echo "    make curl-estimate          ручка 2: только калькулятор"
	@echo "    make curl-invalid           ручка 3: невалидный запрос"
	@echo "    make curl-products          справочник продуктов"
	@echo "    make curl-health            /health обоих сервисов"
	@echo ""
	@echo "  Логи:"
	@echo "    make logs-pricing           лог Pricing Engine"
	@echo "    make logs-ingestion         лог Ingestion Service"
	@echo ""
	@echo "  БД:"
	@echo "    make db-consumption         история затрат"
	@echo "    make db-invalid             невалидные запросы"
	@echo "    make db-products            справочник продуктов"
	@echo "    make db-pricing             psql в Pricing DB"
	@echo "    make db-ingestion           psql в Ingestion DB"
