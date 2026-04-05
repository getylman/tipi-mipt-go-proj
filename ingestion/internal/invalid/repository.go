package invalid

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(rawPayload interface{}, errorReason string) error {
	data, _ := json.Marshal(rawPayload)
	_, err := r.db.Exec(
		`INSERT INTO invalid_metrics (raw_payload, error_reason) VALUES ($1, $2)`,
		string(data), errorReason,
	)
	if err != nil {
		return fmt.Errorf("insert invalid metric: %w", err)
	}
	return nil
}
