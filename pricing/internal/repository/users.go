package repository

import (
	"context"

	"github.com/cloud-pricer/pricing/db/sqlc"
)

type UserRepository struct {
	q *sqlcdb.Queries
}

func NewUserRepository(q *sqlcdb.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Upsert(ctx context.Context, userID string) error {
	return r.q.UpsertUser(ctx, userID)
}
