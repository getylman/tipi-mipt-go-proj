package repository

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

// ProductStore — интерфейс хранилища продуктов.
// Реализации: ProductRepository (прод), MockProductStore (тесты).
type ProductStore interface {
	GetByIDs(ctx context.Context, ids []string) (map[string]types.Product, error)
	ListAll(ctx context.Context) ([]types.Product, error)
}

// UserStore — интерфейс хранилища пользователей.
type UserStore interface {
	Upsert(ctx context.Context, userID string) error
}

// ConsumptionStore — интерфейс хранилища потребления.
type ConsumptionStore interface {
	SaveBatch(ctx context.Context, userID string, items []types.UsageLineItem) error
	GetByUser(ctx context.Context, userID string) ([]types.UsageLineItem, error)
	SumByUser(ctx context.Context, userID string) (float64, error)
}
