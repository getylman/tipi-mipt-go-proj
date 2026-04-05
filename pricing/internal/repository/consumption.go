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

func (r *ConsumptionRepository) SaveBatch(ctx context.Context, userID string, items []types.UsageLineItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

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
