package types

import "time"

// ─── Продукты ─────────────────────────────────────────────────────────────────

// Product — компонент машины (vcpu, ram_gb, disk_gb, network_mbps).
// Каждый продукт имеет цену за единицу потребления.
type Product struct {
	ID           string    `json:"id"`             // "vcpu", "ram_gb", "disk_gb"
	Name         string    `json:"name"`
	PricePerUnit float64   `json:"price_per_unit"` // цена за 1 единицу
	Unit         string    `json:"unit"`           // "core", "gb", "mbps"
	UpdatedAt    time.Time `json:"updated_at"`
}

// ─── Ручка 1: POST /v1/usage ──────────────────────────────────────────────────

// UsageItem — пара (продукт, количество единиц).
type UsageItem struct {
	ProductID string  `json:"product_id"` // "vcpu", "ram_gb", ...
	Quantity  float64 `json:"quantity"`   // количество единиц
}

// UsageRequest — запрос расчёта и сохранения затрат пользователя.
type UsageRequest struct {
	UserID string      `json:"user_id"` // UUID пользователя
	Items  []UsageItem `json:"items"`
}

// UsageLineItem — строка разбивки по продукту в ответе.
type UsageLineItem struct {
	ProductID  string  `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

// UsageResponse — ответ ручки /v1/usage.
type UsageResponse struct {
	Status     string          `json:"status"`      // "ok"
	UserID     string          `json:"user_id"`
	TotalPrice float64         `json:"total_price"`
	Breakdown  []UsageLineItem `json:"breakdown"`
}

// ─── Ручка 2: POST /v1/estimate ───────────────────────────────────────────────

// EstimateRequest — запрос расчёта стоимости без сохранения (без user_id).
type EstimateRequest struct {
	Items []UsageItem `json:"items"`
}

// EstimateResponse — ответ ручки /v1/estimate.
type EstimateResponse struct {
	TotalPrice float64         `json:"total_price"`
	Breakdown  []UsageLineItem `json:"breakdown"`
}

// ─── Невалидные метрики (пишет Ingestion) ─────────────────────────────────────

// InvalidMetricRequest — то что Ingestion пишет в БД при плохой валидации.
type InvalidMetricRequest struct {
	RawPayload  interface{} `json:"raw_payload"`  // исходный запрос как есть
	ErrorReason string      `json:"error_reason"` // описание ошибки
}
