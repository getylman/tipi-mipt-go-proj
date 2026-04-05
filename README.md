# Cloud Pricer v5

Система расчёта и хранения стоимости потребления облачных ресурсов.
Два микросервиса на Go, две отдельные PostgreSQL БД, SQL-запросы через sqlc.

## Архитектура

```
Client
  │
  ▼
Ingestion Service (:8080)
  ├── валидация запроса
  ├── при ошибке → пишет в invalid_metrics (своя БД)
  └── пересылает в Pricing Engine
        │
        ▼
  Pricing Engine (:8081)
    ├── POST /v1/usage    → считает + upsert user + INSERT consumption
    ├── POST /v1/estimate → только считает, ничего не пишет
    └── GET  /v1/products → справочник продуктов

  Pricing DB (port 5432)     Ingestion DB (port 5433)
    ├── products               └── invalid_metrics
    ├── users
    └── consumption
```

## Быстрый старт

```bash
# Всё в Docker Compose
make up

# Или локально (Docker только для БД)
make run
```

## Тестовые запросы

```bash
make curl-usage      # ручка 1: расчёт + сохранение затрат
make curl-estimate   # ручка 2: только калькулятор
make curl-invalid    # ручка 3: невалидный запрос → invalid_metrics
make curl-products   # справочник продуктов с ценами
make curl-health     # /health обоих сервисов
```

## Просмотр данных в БД

```bash
make db-consumption  # история затрат пользователей
make db-invalid      # невалидные запросы
make db-pricing      # psql в БД Pricing Engine
make db-ingestion    # psql в БД Ingestion
```

## sqlc — SQL отдельно от Go

Все SQL запросы живут в `.sql` файлах:

```
pricing/db/queries/
  ├── products.sql     ← GetProductsByIDs, ListProducts
  ├── users.sql        ← UpsertUser
  └── consumption.sql  ← InsertConsumption, GetConsumptionByUser, SumConsumptionByUser
```

Go-код генерируется автоматически:

```bash
make sqlc-generate   # запускает sqlc generate в pricing/
```

Сгенерированные файлы (`pricing/db/sqlc/*.go`) не редактируются вручную.

## Переменные окружения

### Ingestion Service
| Переменная | По умолчанию |
|---|---|
| `HTTP_PORT` | `8080` |
| `LOG_LEVEL` | `info` |
| `LOG_FORMAT` | `text` |
| `ENVIRONMENT` | `dev` |
| `PRICING_ENGINE_URL` | `http://localhost:8081` |
| `DATABASE_URL` | `postgres://pricer:pricer@localhost:5433/ingestion?sslmode=disable` |
| `MIGRATION_FILE` | `./migrations/001_init.sql` |

### Pricing Engine
| Переменная | По умолчанию |
|---|---|
| `HTTP_PORT` | `8081` |
| `LOG_LEVEL` | `info` |
| `LOG_FORMAT` | `text` |
| `ENVIRONMENT` | `dev` |
| `DATABASE_URL` | `postgres://pricer:pricer@localhost:5432/pricing?sslmode=disable` |
| `MIGRATION_FILE` | `./migrations/001_init.sql` |

## Продукты (начальные данные)

| ID | Описание | Цена | Единица |
|---|---|---|---|
| `vcpu` | Virtual CPU | 2.50 | core |
| `ram_gb` | RAM | 0.80 | gb |
| `disk_gb` | SSD Disk | 0.15 | gb |
| `network_mbps` | Network bandwidth | 0.05 | mbps |

Пример: 4 vCPU + 16 GB RAM + 100 GB disk = 4×2.50 + 16×0.80 + 100×0.15 = **37.80**

## Структура проекта

```
cloud-pricer/
├── go.work
├── docker-compose.yml
├── Makefile
│
├── shared/                        # Общие типы, логгер, ошибки
│   ├── types/types.go             # UsageRequest, EstimateRequest, Product...
│   ├── apierror/apierror.go       # Типизированные HTTP-ошибки
│   ├── config/config.go           # Base, LoadBase, GetEnv*
│   ├── logger/logger.go           # slog wrapper
│   └── tracing/tracing.go         # OpenTelemetry no-op
│
├── ingestion/                     # Сервис 1 (:8080)
│   ├── cmd/main.go
│   ├── config/config.go
│   ├── migrations/001_init.sql    # invalid_metrics таблица
│   └── internal/
│       ├── handler/               # POST /v1/usage, POST /v1/estimate
│       ├── validator/             # валидация входящих запросов
│       ├── client/                # HTTP клиент к Pricing Engine
│       └── invalid/               # репозиторий invalid_metrics
│
└── pricing/                       # Сервис 2 (:8081)
    ├── cmd/main.go
    ├── config/config.go
    ├── sqlc.yaml                  # конфиг sqlc генератора
    ├── migrations/001_init.sql    # products, users, consumption таблицы
    ├── db/
    │   ├── queries/               # SQL файлы (редактируем здесь)
    │   │   ├── products.sql
    │   │   ├── users.sql
    │   │   └── consumption.sql
    │   └── sqlc/                  # сгенерированный Go-код (не трогаем)
    │       ├── db.go
    │       ├── models.go
    │       ├── products.sql.go
    │       ├── users.sql.go
    │       └── consumption.sql.go
    └── internal/
        ├── handler/               # POST /v1/usage, /v1/estimate, GET /v1/products
        ├── usage/                 # бизнес-логика ручки 1
        ├── estimate/              # бизнес-логика ручки 2
        └── repository/            # обёртки над sqlc (без SQL строк)
            ├── db.go
            ├── products.go
            ├── users.go
            └── consumption.go
```
