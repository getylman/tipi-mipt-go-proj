package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cloud-pricer/pricing/db/sqlc"
	"github.com/cloud-pricer/shared/types"
)

type ConsumptionRepository struct {
	q  *sqlcdb.Queries
	db *sql.DB // нужен для Begin() транзакции
}

func NewConsumptionRepository(db *sql.DB, q *sqlcdb.Queries) *ConsumptionRepository {
	return &ConsumptionRepository{q: q, db: db}
}

// SaveBatch сохраняет все строки потребления в одной транзакции.
// Если один INSERT падает — откатываем все.
// SQL в db/queries/consumption.sql.
func (r *ConsumptionRepository) SaveBatch(ctx context.Context, userID string, items []types.UsageLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// WithTx — sqlc переключает Queries на транзакцию
	qtx := r.q.WithTx(tx)

	for _, item := range items {
		err := qtx.InsertConsumption(ctx, sqlcdb.InsertConsumptionParams{
			UserID:     userID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
		if err != nil {
			return fmt.Errorf("insert consumption for %s: %w", item.ProductID, err)
		}
	}

	return tx.Commit()
}

// GetByUser возвращает историю потребления пользователя.
func (r *ConsumptionRepository) GetByUser(ctx context.Context, userID string) ([]types.UsageLineItem, error) {
	rows, err := r.q.GetConsumptionByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get consumption for %s: %w", userID, err)
	}

	result := make([]types.UsageLineItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.UsageLineItem{
			ProductID:  row.ProductID,
			Quantity:   row.Quantity,
			UnitPrice:  row.UnitPrice,
			TotalPrice: row.TotalPrice,
		})
	}
	return result, nil
}

// SumByUser возвращает накопленную сумму затрат пользователя.
func (r *ConsumptionRepository) SumByUser(ctx context.Context, userID string) (float64, error) {
	total, err := r.q.SumConsumptionByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("sum consumption for %s: %w", userID, err)
	}
	return total, nil
}
