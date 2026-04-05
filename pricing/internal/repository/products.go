package repository

import (
	"context"
	"fmt"

	"github.com/cloud-pricer/pricing/db/sqlc"
	"github.com/cloud-pricer/shared/types"
)

type ProductRepository struct {
	q *sqlcdb.Queries
}

func NewProductRepository(q *sqlcdb.Queries) *ProductRepository {
	return &ProductRepository{q: q}
}

func (r *ProductRepository) GetByIDs(ctx context.Context, ids []string) (map[string]types.Product, error) {
	if len(ids) == 0 {
		return map[string]types.Product{}, nil
	}
	rows, err := r.q.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get products by ids: %w", err)
	}
	result := make(map[string]types.Product, len(rows))
	for _, row := range rows {
		result[row.ID] = toProduct(row)
	}
	return result, nil
}

func (r *ProductRepository) ListAll(ctx context.Context) ([]types.Product, error) {
	rows, err := r.q.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	result := make([]types.Product, len(rows))
	for i, row := range rows {
		result[i] = toProduct(row)
	}
	return result, nil
}

func toProduct(row sqlcdb.Product) types.Product {
	return types.Product{
		ID:           row.ID,
		Name:         row.Name,
		PricePerUnit: row.PricePerUnit,
		Unit:         row.Unit,
		UpdatedAt:    row.UpdatedAt,
	}
}
