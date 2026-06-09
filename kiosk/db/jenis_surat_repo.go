package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type JenisSuratRepository struct {
	db *DB
}

func NewJenisSuratRepository(db *DB) *JenisSuratRepository {
	return &JenisSuratRepository{db: db}
}

// ListAktif lists all active jenis_surat for the kiosk.
func (r *JenisSuratRepository) ListAktif(ctx context.Context) ([]models.JenisSurat, error) {
	query := `
		SELECT id, kode, nama, deskripsi, fields_schema, aktif, urutan, updated_at
		FROM jenis_surat
		WHERE aktif = 1
		ORDER BY urutan ASC, nama ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal query jenis_surat aktif: %w", err)
	}
	defer rows.Close()

	var result []models.JenisSurat
	for rows.Next() {
		var js models.JenisSurat
		var fieldsSchemaStr string
		var updatedAt time.Time

		err := rows.Scan(
			&js.ID, &js.Kode, &js.Nama, &js.Deskripsi, &fieldsSchemaStr, &js.Aktif, &js.Urutan, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan jenis_surat: %w", err)
		}

		js.FieldsSchema = json.RawMessage(fieldsSchemaStr)
		js.UpdatedAt = updatedAt
		result = append(result, js)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// FindByID retrieves a jenis_surat by its ID.
func (r *JenisSuratRepository) FindByID(ctx context.Context, id string) (*models.JenisSurat, error) {
	query := `
		SELECT id, kode, nama, deskripsi, fields_schema, aktif, urutan, updated_at
		FROM jenis_surat
		WHERE id = ?
	`
	var js models.JenisSurat
	var fieldsSchemaStr string
	var updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&js.ID, &js.Kode, &js.Nama, &js.Deskripsi, &fieldsSchemaStr, &js.Aktif, &js.Urutan, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal query jenis_surat: %w", err)
	}

	js.FieldsSchema = json.RawMessage(fieldsSchemaStr)
	js.UpdatedAt = updatedAt

	return &js, nil
}

// Upsert inserts or updates a jenis_surat (synced from online backend).
func (r *JenisSuratRepository) Upsert(ctx context.Context, js *models.JenisSurat) error {
	query := `
		INSERT INTO jenis_surat (
			id, kode, nama, deskripsi, fields_schema, aktif, urutan, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kode = excluded.kode,
			nama = excluded.nama,
			deskripsi = excluded.deskripsi,
			fields_schema = excluded.fields_schema,
			aktif = excluded.aktif,
			urutan = excluded.urutan,
			updated_at = excluded.updated_at
	`
	schemaBytes, err := json.Marshal(js.FieldsSchema)
	if err != nil {
		return fmt.Errorf("gagal marshal fields_schema: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		js.ID, js.Kode, js.Nama, js.Deskripsi, string(schemaBytes), js.Aktif, js.Urutan, js.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert jenis_surat: %w", err)
	}

	return nil
}

// GetTemplate retrieves the HTML template for a specific jenis_surat and desa.
func (r *JenisSuratRepository) GetTemplate(ctx context.Context, jenisSuratID string, desaID string) (*models.SuratTemplate, error) {
	query := `
		SELECT id, jenis_surat_id, desa_id, template_html, version, updated_at
		FROM surat_template
		WHERE jenis_surat_id = ? AND desa_id = ?
	`
	var t models.SuratTemplate
	var updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, jenisSuratID, desaID).Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &t.Version, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal query template: %w", err)
	}

	t.UpdatedAt = updatedAt
	return &t, nil
}

// UpsertTemplate inserts or updates a template (synced from online backend).
func (r *JenisSuratRepository) UpsertTemplate(ctx context.Context, t *models.SuratTemplate) error {
	query := `
		INSERT INTO surat_template (
			id, jenis_surat_id, desa_id, template_html, version, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(jenis_surat_id, desa_id) DO UPDATE SET
			id = excluded.id,
			template_html = excluded.template_html,
			version = excluded.version,
			updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, t.Version, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert template: %w", err)
	}

	return nil
}
