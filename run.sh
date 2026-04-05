#!/usr/bin/env bash
# run.sh — сборка и запуск сервисов
#
# Использование:
#   ./run.sh          — собрать и запустить
#   ./run.sh stop     — остановить сервисы
#   ./run.sh build    — только собрать бинарники
#   ./run.sh logs     — показать логи обоих сервисов

set -euo pipefail

# ── Настройки ────────────────────────────────────────────────────

INGESTION_PORT=8080
PRICING_PORT=8081

DB_USER=pricer
DB_PASS=pricer
DB_PRICING=pricing
DB_INGESTION=ingestion

DATABASE_URL_PRICING="postgres://${DB_USER}:${DB_PASS}@localhost:5432/${DB_PRICING}?sslmode=disable"
DATABASE_URL_INGESTION="postgres://${DB_USER}:${DB_PASS}@localhost:5433/${DB_INGESTION}?sslmode=disable"

BIN_DIR=".bin"
LOG_DIR=".logs"
PID_DIR=".pids"
BIN_PRICING="${BIN_DIR}/pricing"
BIN_INGESTION="${BIN_DIR}/ingestion"
PID_PRICING="${PID_DIR}/pricing.pid"
PID_INGESTION="${PID_DIR}/ingestion.pid"
LOG_PRICING="${LOG_DIR}/pricing.log"
LOG_INGESTION="${LOG_DIR}/ingestion.log"

# docker или sudo docker — определяется автоматически
if docker info > /dev/null 2>&1; then
    DOCKER="docker"
elif sudo docker info > /dev/null 2>&1; then
    DOCKER="sudo docker"
else
    echo "ОШИБКА: docker не найден или недоступен"
    exit 1
fi

# ── Команды ──────────────────────────────────────────────────────

cmd_build() {
    echo "  Сборка Pricing Engine..."
    mkdir -p "$BIN_DIR"
    (cd pricing   && go build -o "../${BIN_PRICING}"   ./cmd/main.go)
    echo "  Сборка Ingestion Service..."
    (cd ingestion && go build -o "../${BIN_INGESTION}" ./cmd/main.go)
    echo "  Готово."
}

cmd_db_start() {
    echo "  Запуск PostgreSQL (pricing)..."
    $DOCKER run -d --name pricer-postgres-pricing \
        -e POSTGRES_USER="$DB_USER" \
        -e POSTGRES_PASSWORD="$DB_PASS" \
        -e POSTGRES_DB="$DB_PRICING" \
        -p 5432:5432 \
        postgres:16-alpine > /dev/null 2>&1 \
        && echo "  PostgreSQL (pricing) запущен" \
        || echo "  PostgreSQL (pricing) уже запущен"

    echo "  Запуск PostgreSQL (ingestion)..."
    $DOCKER run -d --name pricer-postgres-ingestion \
        -e POSTGRES_USER="$DB_USER" \
        -e POSTGRES_PASSWORD="$DB_PASS" \
        -e POSTGRES_DB="$DB_INGESTION" \
        -p 5433:5432 \
        postgres:16-alpine > /dev/null 2>&1 \
        && echo "  PostgreSQL (ingestion) запущен" \
        || echo "  PostgreSQL (ingestion) уже запущен"
}

cmd_db_wait() {
    echo "  Ожидание готовности БД..."
    sleep 3
    for i in $(seq 1 30); do
        if $DOCKER exec pricer-postgres-pricing   pg_isready -U "$DB_USER" -q 2>/dev/null && \
           $DOCKER exec pricer-postgres-ingestion pg_isready -U "$DB_USER" -q 2>/dev/null; then
            echo "  БД готовы"
            return 0
        fi
        echo "  Попытка ${i}/30..."
        sleep 2
    done
    echo "  ОШИБКА: БД не ответила за 60 секунд."
    echo "  Проверь: $DOCKER logs pricer-postgres-pricing"
    exit 1
}

cmd_kill_old() {
    echo "  Освобождение портов..."
    lsof -t -i :"$PRICING_PORT"   2>/dev/null | xargs kill 2>/dev/null || true
    lsof -t -i :"$INGESTION_PORT" 2>/dev/null | xargs kill 2>/dev/null || true
    rm -f "$PID_PRICING" "$PID_INGESTION"
    sleep 1
}

cmd_run_pricing() {
    echo "  Запуск Pricing Engine на :${PRICING_PORT}..."
    mkdir -p "$LOG_DIR" "$PID_DIR"
    > "$LOG_PRICING"

    HTTP_PORT="$PRICING_PORT" \
    DATABASE_URL="$DATABASE_URL_PRICING" \
    MIGRATION_FILE=./pricing/migrations/001_init.sql \
    LOG_LEVEL=debug \
    LOG_FORMAT=text \
    ENVIRONMENT=dev \
    "$BIN_PRICING" >> "$LOG_PRICING" 2>&1 &

    echo $! > "$PID_PRICING"
    sleep 2

    if kill -0 "$(cat "$PID_PRICING")" 2>/dev/null; then
        echo "    OK (PID $(cat "$PID_PRICING"))"
    else
        echo "    ОШИБКА: Pricing Engine не запустился."
        tail -30 "$LOG_PRICING"
        exit 1
    fi
}

cmd_run_ingestion() {
    echo "  Запуск Ingestion Service на :${INGESTION_PORT}..."
    > "$LOG_INGESTION"

    HTTP_PORT="$INGESTION_PORT" \
    PRICING_ENGINE_URL="http://localhost:${PRICING_PORT}" \
    DATABASE_URL="$DATABASE_URL_INGESTION" \
    MIGRATION_FILE=./ingestion/migrations/001_init.sql \
    LOG_LEVEL=debug \
    LOG_FORMAT=text \
    ENVIRONMENT=dev \
    "$BIN_INGESTION" >> "$LOG_INGESTION" 2>&1 &

    echo $! > "$PID_INGESTION"
    sleep 2

    if kill -0 "$(cat "$PID_INGESTION")" 2>/dev/null; then
        echo "    OK (PID $(cat "$PID_INGESTION"))"
    else
        echo "    ОШИБКА: Ingestion Service не запустился."
        tail -30 "$LOG_INGESTION"
        exit 1
    fi
}

cmd_stop() {
    if [ -f "$PID_PRICING" ]; then
        PID=$(cat "$PID_PRICING")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" && echo "  Pricing Engine (PID $PID) остановлен"
        else
            echo "  Pricing Engine уже не запущен"
        fi
        rm -f "$PID_PRICING"
    else
        echo "  Pricing Engine не запущен"
    fi

    if [ -f "$PID_INGESTION" ]; then
        PID=$(cat "$PID_INGESTION")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" && echo "  Ingestion Service (PID $PID) остановлен"
        else
            echo "  Ingestion Service уже не запущен"
        fi
        rm -f "$PID_INGESTION"
    else
        echo "  Ingestion Service не запущен"
    fi

    echo ""
    echo "  Остановить БД:"
    echo "    $DOCKER rm -f pricer-postgres-pricing pricer-postgres-ingestion"
}

cmd_logs() {
    echo "=== Pricing Engine ==="
    cat "$LOG_PRICING" 2>/dev/null || echo "(лог пуст)"
    echo ""
    echo "=== Ingestion Service ==="
    cat "$LOG_INGESTION" 2>/dev/null || echo "(лог пуст)"
}

cmd_status() {
    echo ""
    echo "  Сервисы запущены:"
    echo "    Ingestion  →  http://localhost:${INGESTION_PORT}"
    echo "    Pricing    →  http://localhost:${PRICING_PORT}"
    echo ""
    echo "  ./run.sh stop   — остановить"
    echo "  ./demo.sh       — запустить демо-запросы"
}

# ── Точка входа ───────────────────────────────────────────────────

ACTION="${1:-start}"

case "$ACTION" in
    start|"")
        cmd_build
        cmd_kill_old
        cmd_db_start
        cmd_db_wait
        cmd_run_pricing
        cmd_run_ingestion
        cmd_status
        ;;
    stop)
        cmd_stop
        ;;
    build)
        cmd_build
        ;;
    logs)
        cmd_logs
        ;;
    *)
        echo "Использование: $0 [start|stop|build|logs]"
        exit 1
        ;;
esac
