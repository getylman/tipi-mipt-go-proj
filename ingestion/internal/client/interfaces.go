package client

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

// PricingAPI — интерфейс клиента к Pricing Engine.
// Реализации: PricingClient (прод), MockPricingClient (тесты).
type PricingAPI interface {
	Usage(ctx context.Context, req *types.UsageRequest) (*types.UsageResponse, error)
	Estimate(ctx context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error)
}
