package repository

import (
	"context"
	"fmt"

	"github.com/cloud-pricer/pricing/db/sqlc"
)

type UserRepository struct {
	q *sqlcdb.Queries
}

func NewUserRepository(q *sqlcdb.Queries) *UserRepository {
	return &UserRepository{q: q}
}

// Upsert создаёт пользователя если не существует.
// SQL в db/queries/users.sql.
func (r *UserRepository) Upsert(ctx context.Context, userID string) error {
	if err := r.q.UpsertUser(ctx, userID); err != nil {
		return fmt.Errorf("upsert user %s: %w", userID, err)
	}
	return nil
}
