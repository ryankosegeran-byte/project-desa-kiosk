package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ConfigRepository struct {
	db *DB
}

func NewConfigRepository(db *DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Get retrieves a config value by key.
func (r *ConfigRepository) Get(ctx context.Context, key string) (string, error) {
	query := "SELECT value FROM kiosk_config WHERE key = ?"
	var val string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("gagal query config key %s: %w", key, err)
	}
	return val, nil
}

// Set inserts or updates a config key-value pair.
func (r *ConfigRepository) Set(ctx context.Context, key, val string) error {
	query := `
		INSERT INTO kiosk_config (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`
	_, err := r.db.ExecContext(ctx, query, key, val)
	if err != nil {
		return fmt.Errorf("gagal set config key %s: %w", key, err)
	}
	return nil
}
