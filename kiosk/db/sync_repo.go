package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type SyncRepository struct {
	db *DB
}

func NewSyncRepository(db *DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// ListPendingSync retrieves all unprocessed items in the sync queue.
func (r *SyncRepository) ListPendingSync(ctx context.Context) ([]models.SyncQueueItem, error) {
	query := `
		SELECT id, entity_type, entity_id, operation, payload, attempts, max_attempts, last_error, created_at
		FROM sync_queue
		WHERE processed_at IS NULL AND attempts < max_attempts
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal query pending sync queue: %w", err)
	}
	defer rows.Close()

	var result []models.SyncQueueItem
	for rows.Next() {
		var item models.SyncQueueItem
		var lastErr sql.NullString
		var payloadStr string

		err := rows.Scan(
			&item.ID, &item.EntityType, &item.EntityID, &item.Operation, &payloadStr,
			&item.Attempts, &item.MaxAttempts, &lastErr, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan sync queue row: %w", err)
		}

		item.Payload = json.RawMessage(payloadStr)
		item.LastError = lastErr.String
		result = append(result, item)
	}

	return result, nil
}

// MarkProcessed marks a sync queue item as successfully processed.
func (r *SyncRepository) MarkProcessed(ctx context.Context, id int64) error {
	query := `
		UPDATE sync_queue
		SET processed_at = ?
		WHERE id = ?
	`
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("gagal update sync queue processed: %w", err)
	}
	return nil
}

// IncrementAttempts increments the retry counter and records the last error.
func (r *SyncRepository) IncrementAttempts(ctx context.Context, id int64, errMessage string) error {
	query := `
		UPDATE sync_queue
		SET attempts = attempts + 1, last_error = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, errMessage, id)
	if err != nil {
		return fmt.Errorf("gagal increment attempts sync queue: %w", err)
	}
	return nil
}
