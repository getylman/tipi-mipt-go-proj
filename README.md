# cloud-pricer

Система расчёта и хранения стоимости потребления облачных ресурсов.
Два микросервиса на Go, две отдельные PostgreSQL БД.

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
    ├── POST /v1/usage      → считает + сохраняет потребление
    ├── POST /v1/estimate   → только считает, ничего не пишет
    └── GET  /v1/products   → справочник продуктов

  Pricing DB (:5432)         Ingestion DB (:5433)
    ├── products               └── invalid_metrics
    ├── users
    └── consumption
```

## Запуск

```bash
# Запустить (сборка + PostgreSQL в Docker + оба сервиса)
./run.sh

# Остановить сервисы
./run.sh stop

# Только собрать бинарники
./run.sh build

# Посмотреть логи
./run.sh logs
```

Если Docker требует sudo — скрипт определит это автоматически.
Для постоянного решения: `sudo usermod -aG docker $USER` (затем перелогиниться).

## Демонстрация

```bash
# Полная демонстрация: юнит-тесты + HTTP запросы + интеграционные тесты
./demo.sh

# Только HTTP запросы (нужен ./run.sh)
./demo.sh http

# Только юнит-тесты (без сервисов)
./demo.sh unit

# Только интеграционные тесты (нужен ./run.sh)
./demo.sh integ
```

## Продукты

| ID | Описание | Цена | Единица |
|---|---|---|---|
| `vcpu` | Virtual CPU | 2.50 | core |
| `ram_gb` | RAM | 0.80 | gb |
| `disk_gb` | SSD Disk | 0.15 | gb |
| `network_mbps` | Network | 0.05 | mbps |

Пример: 4 vCPU + 16 GB RAM + 100 GB disk = **37.80**

## Структура

```
cloud-pricer/
├── run.sh                     # запуск и остановка
├── demo.sh                    # демонстрация и тесты
├── go.work
│
├── shared/                    # общие типы, логгер, ошибки
├── ingestion/                 # Ingestion Service (:8080)
│   ├── cmd/main.go            # точка входа; config.MustValidate() в начале
│   ├── config/                # Load() + MustValidate() (валидация при старте)
│   ├── migrations/
│   └── internal/
│       ├── handler/           # HTTP-хендлеры + интерфейсы PricingClient, InvalidStore
│       ├── validator/         # валидация запросов + тесты
│       ├── client/            # реализация HTTP-клиента к Pricing Engine
│       ├── invalid/           # репозиторий невалидных метрик
│       └── mocks/             # моки для юнит-тестов handler
│
└── pricing/                   # Pricing Engine (:8081)
    ├── cmd/main.go            # точка входа; config.MustValidate() в начале
    ├── config/                # Load() + MustValidate() (валидация при старте)
    ├── migrations/
    ├── integration_test.go    # интеграционные тесты (пропускаются без живых сервисов)
    ├── db/queries/            # SQL запросы (sqlc)
    ├── db/sqlc/               # сгенерированный код
    └── internal/
        ├── handler/           # HTTP-хендлеры + интерфейс ProductLister
        ├── usage/             # бизнес-логика usage + интерфейсы + тесты
        ├── estimate/          # бизнес-логика estimate + интерфейс + тесты
        ├── repository/        # реализации репозиториев (конкретные типы)
        └── mocks/             # моки для юнит-тестов usage/estimate
```

## Переменные окружения

| Сервис | Переменная | По умолчанию |
|---|---|---|
| Оба | `LOG_LEVEL` | `info` |
| Оба | `LOG_FORMAT` | `text` |
| Оба | `ENVIRONMENT` | `dev` |
| Оба | `DATABASE_URL` | — обязательный |
| Ingestion | `PRICING_ENGINE_URL` | — обязательный |
| Ingestion | `MAX_ITEMS` | `50` |
| Pricing | `HTTP_PORT` | `8081` |
| Ingestion | `HTTP_PORT` | `8080` |
