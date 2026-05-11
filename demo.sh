#!/usr/bin/env bash
# demo.sh — демонстрация работы системы и запуск тестов
#
# Использование:
#   ./demo.sh          — полная демонстрация (HTTP + юнит-тесты)
#   ./demo.sh http     — только HTTP запросы к запущенным сервисам
#   ./demo.sh unit     — только юнит-тесты (без сервисов)
#   ./demo.sh integ    — только интеграционные тесты (нужен ./run.sh)

set -euo pipefail

INGESTION_URL="http://localhost:8080"
PRICING_URL="http://localhost:8081"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
fail() { echo -e "${RED}✗${NC} $*"; }
info() { echo -e "${BOLD}$*${NC}"; }
sep()  { echo "────────────────────────────────────────"; }

# ── Проверка доступности сервисов ────────────────────────────────

check_services() {
    if ! curl -s --max-time 2 "${INGESTION_URL}/health" > /dev/null 2>&1; then
        echo ""
        fail "Сервисы не запущены. Сначала выполни: ./run.sh"
        echo ""
        exit 1
    fi
}

# ── HTTP демонстрация ─────────────────────────────────────────────

demo_http() {
    info "HTTP демонстрация"
    sep

    # Health check
    echo ""
    info "1. Health check обоих сервисов"
    echo -n "   Ingestion  → "
    curl -s "${INGESTION_URL}/health" | python3 -m json.tool 2>/dev/null | tr -d '\n' | sed 's/  */ /g'
    echo ""
    echo -n "   Pricing    → "
    curl -s "${PRICING_URL}/health" | python3 -m json.tool 2>/dev/null | tr -d '\n' | sed 's/  */ /g'
    echo ""
    ok "Оба сервиса живы"

    # Справочник продуктов
    echo ""
    info "2. Справочник продуктов (GET /v1/products)"
    curl -s "${PRICING_URL}/v1/products" | python3 -m json.tool
    ok "Справочник получен"

    # Ручка 2 — estimate (только расчёт, без сохранения)
    echo ""
    info "3. Расчёт стоимости без сохранения (POST /v1/estimate)"
    echo "   Запрос: 4 vCPU + 16 GB RAM + 100 GB disk"
    ESTIMATE_RESP=$(curl -s -X POST "${INGESTION_URL}/v1/estimate" \
        -H "Content-Type: application/json" \
        -d '{"items":[{"product_id":"vcpu","quantity":4},{"product_id":"ram_gb","quantity":16},{"product_id":"disk_gb","quantity":100}]}')
    echo "$ESTIMATE_RESP" | python3 -m json.tool
    ESTIMATE_TOTAL=$(echo "$ESTIMATE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['total_price'])")
    ok "Расчёт: total_price = ${ESTIMATE_TOTAL} (в БД ничего не записано)"

    # Ручка 1 — usage (расчёт + сохранение)
    echo ""
    info "4. Расчёт и сохранение затрат (POST /v1/usage)"
    echo "   Запрос: user_id + 4 vCPU + 16 GB RAM + 100 GB disk"
    USAGE_RESP=$(curl -s -X POST "${INGESTION_URL}/v1/usage" \
        -H "Content-Type: application/json" \
        -d '{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[{"product_id":"vcpu","quantity":4},{"product_id":"ram_gb","quantity":16},{"product_id":"disk_gb","quantity":100}]}')
    echo "$USAGE_RESP" | python3 -m json.tool
    USAGE_STATUS=$(echo "$USAGE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")
    USAGE_TOTAL=$(echo "$USAGE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['total_price'])")
    ok "Сохранено: status=${USAGE_STATUS}, total_price=${USAGE_TOTAL}"

    # Ручка 3 — невалидный запрос
    echo ""
    info "5. Невалидный запрос — пустой user_id (POST /v1/usage)"
    INVALID_RESP=$(curl -s -X POST "${INGESTION_URL}/v1/usage" \
        -H "Content-Type: application/json" \
        -d '{"user_id":"","items":[]}')
    echo "$INVALID_RESP" | python3 -m json.tool
    ERR_CODE=$(echo "$INVALID_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['error']['code'])")
    ok "Валидация сработала: error.code = ${ERR_CODE} (записано в invalid_metrics)"

    sep
    echo ""
    ok "HTTP демонстрация завершена"
    echo ""
    echo "  Посмотреть данные в БД:"
    echo "    docker exec pricer-postgres-pricing   psql -U pricer -d pricing   -c 'SELECT product_id, quantity, total_price FROM consumption ORDER BY calculated_at DESC LIMIT 5;'"
    echo "    docker exec pricer-postgres-ingestion psql -U pricer -d ingestion -c 'SELECT error_reason, received_at FROM invalid_metrics ORDER BY received_at DESC LIMIT 5;'"
}

# ── Юнит-тесты ───────────────────────────────────────────────────

demo_unit() {
    info "Юнит-тесты"
    sep
    echo ""

    FAILED=0

    echo "  Pricing Engine..."
    if (cd pricing && go test ./internal/... -v 2>&1 | grep -E "PASS|FAIL|ok|---"); then
        ok "Pricing Engine — все тесты прошли"
    else
        fail "Pricing Engine — есть упавшие тесты"
        FAILED=1
    fi

    echo ""
    echo "  Ingestion Service..."
    if (cd ingestion && go test ./internal/... -v 2>&1 | grep -E "PASS|FAIL|ok|---"); then
        ok "Ingestion Service — все тесты прошли"
    else
        fail "Ingestion Service — есть упавшие тесты"
        FAILED=1
    fi

    echo ""
    sep
    if [ "$FAILED" -eq 0 ]; then
        ok "Все юнит-тесты прошли"
    else
        fail "Есть упавшие тесты"
        exit 1
    fi
    echo ""
}

# ── Интеграционные тесты ─────────────────────────────────────────

demo_integ() {
    info "Интеграционные тесты"
    sep
    echo ""
    check_services

    if (cd pricing && go test . -v -run TestIntegration 2>&1 | grep -E "PASS|FAIL|SKIP|ok|---"); then
        ok "Интеграционные тесты прошли"
    else
        fail "Интеграционные тесты упали"
        exit 1
    fi
    echo ""
}

# ── Точка входа ───────────────────────────────────────────────────

ACTION="${1:-all}"

case "$ACTION" in
    all|"")
        echo ""
        info "=== Демонстрация cloud-pricer ==="
        echo ""
        demo_unit
        echo ""
        check_services
        demo_http
        echo ""
        demo_integ
        ;;
    http)
        check_services
        demo_http
        ;;
    unit)
        demo_unit
        ;;
    integ)
        demo_integ
        ;;
    *)
        echo "Использование: $0 [all|http|unit|integ]"
        exit 1
        ;;
esac
