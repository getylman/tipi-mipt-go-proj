package types

import "time"

type Product struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	PricePerUnit float64   `json:"price_per_unit"`
	Unit         string    `json:"unit"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UsageItem struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

type UsageRequest struct {
	UserID string      `json:"user_id"`
	Items  []UsageItem `json:"items"`
}

type UsageLineItem struct {
	ProductID  string  `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type UsageResponse struct {
	Status     string          `json:"status"`
	UserID     string          `json:"user_id"`
	TotalPrice float64         `json:"total_price"`
	Breakdown  []UsageLineItem `json:"breakdown"`
}

type EstimateRequest struct {
	Items []UsageItem `json:"items"`
}

type EstimateResponse struct {
	TotalPrice float64         `json:"total_price"`
	Breakdown  []UsageLineItem `json:"breakdown"`
}
