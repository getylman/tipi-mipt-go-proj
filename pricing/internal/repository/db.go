package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db, nil
}

func Migrate(db *sql.DB, migrationFile string) error {
	data, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("read migration %q: %w", migrationFile, err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}
