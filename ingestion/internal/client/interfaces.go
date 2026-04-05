package client

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type PricingAPI interface {
	Usage(ctx context.Context, req *types.UsageRequest) (*types.UsageResponse, error)
	Estimate(ctx context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error)
}
