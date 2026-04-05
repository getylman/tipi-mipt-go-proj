package repository

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type ProductStore interface {
	GetByIDs(ctx context.Context, ids []string) (map[string]types.Product, error)
	ListAll(ctx context.Context) ([]types.Product, error)
}

type UserStore interface {
	Upsert(ctx context.Context, userID string) error
}

type ConsumptionStore interface {
	SaveBatch(ctx context.Context, userID string, items []types.UsageLineItem) error
}
