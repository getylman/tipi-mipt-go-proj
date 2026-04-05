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

// GetByIDs возвращает продукты по списку id.
// SQL живёт в db/queries/products.sql — здесь только вызов.
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
		result[row.ID] = types.Product{
			ID:           row.ID,
			Name:         row.Name,
			PricePerUnit: row.PricePerUnit,
			Unit:         row.Unit,
			UpdatedAt:    row.UpdatedAt,
		}
	}
	return result, nil
}

// ListAll возвращает все продукты для GET /v1/products.
func (r *ProductRepository) ListAll(ctx context.Context) ([]types.Product, error) {
	rows, err := r.q.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	result := make([]types.Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.Product{
			ID:           row.ID,
			Name:         row.Name,
			PricePerUnit: row.PricePerUnit,
			Unit:         row.Unit,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return result, nil
}
