package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type TemplateRepository struct {
	db *DB
}

func NewTemplateRepository(db *DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) GetTemplate(ctx context.Context, jenisSuratID string, desaID string) (*models.SuratTemplate, error) {
	query := `
		SELECT id, jenis_surat_id, desa_id, template_html, version, created_by, created_at, updated_at
		FROM surat_template
		WHERE jenis_surat_id = $1 AND desa_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, jenisSuratID, desaID)
	return r.scanRow(row)
}

func (r *TemplateRepository) UpsertTemplate(ctx context.Context, t *models.SuratTemplate) error {
	query := `
		INSERT INTO surat_template (id, jenis_surat_id, desa_id, template_html, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (jenis_surat_id, desa_id) DO UPDATE SET
			template_html = excluded.template_html,
			version = surat_template.version + 1,
			created_by = excluded.created_by,
			updated_at = excluded.updated_at
	`
	var createdBy interface{}
	if t.CreatedBy != "" {
		createdBy = t.CreatedBy
	}

	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, t.Version, createdBy, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) ListTemplatesForDesa(ctx context.Context, desaID string) ([]models.SuratTemplate, error) {
	query := `
		SELECT id, jenis_surat_id, desa_id, template_html, version, created_by, created_at, updated_at
		FROM surat_template
		WHERE desa_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal list templates: %w", err)
	}
	defer rows.Close()

	var result []models.SuratTemplate
	for rows.Next() {
		t, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *t)
	}
	return result, nil
}

func (r *TemplateRepository) scanRow(row *sql.Row) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var createdBy sql.NullString

	err := row.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &t.Version, &createdBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan template row: %w", err)
	}

	t.CreatedBy = createdBy.String
	return &t, nil
}

func (r *TemplateRepository) scanRows(rows *sql.Rows) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var createdBy sql.NullString

	err := rows.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &t.Version, &createdBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan template rows: %w", err)
	}

	t.CreatedBy = createdBy.String
	return &t, nil
}
