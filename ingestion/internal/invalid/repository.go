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

// Save пишет невалидный запрос в таблицу invalid_metrics.
func (r *Repository) Save(rawPayload interface{}, errorReason string) error {
	data, err := json.Marshal(rawPayload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	_, err = r.db.Exec(`
		INSERT INTO invalid_metrics (raw_payload, error_reason)
		VALUES ($1, $2)
	`, string(data), errorReason)
	if err != nil {
		return fmt.Errorf("insert invalid metric: %w", err)
	}
	return nil
}
