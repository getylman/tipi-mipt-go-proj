package handler

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type PricingClient interface {
	Usage(ctx context.Context, req *types.UsageRequest) (*types.UsageResponse, error)
	Estimate(ctx context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error)
}

type InvalidStore interface {
	Save(rawPayload interface{}, errorReason string) error
}
