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

// GetTemplate retrieves the template for a specific jenis_surat and desa.
// Implements hierarchy: 1. Per-desa template, 2. General template, 3. Error if none.
// Returns DOCX bytes + placeholder mapping (Strategi B) when present.
func (r *JenisSuratRepository) GetTemplate(ctx context.Context, jenisSuratID string, desaID string) (*models.SuratTemplate, error) {
	const cols = `id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, updated_at`

	// First try per-desa template
	query := `
		SELECT ` + cols + `
		FROM surat_template
		WHERE jenis_surat_id = ? AND desa_id = ? AND (is_general = 0 OR is_general IS NULL)
	`
	t, err := scanTemplateRow(r.db.QueryRowContext(ctx, query, jenisSuratID, desaID))
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("gagal query template per-desa: %w", err)
	}

	// Fallback to general template
	query = `
		SELECT ` + cols + `
		FROM surat_template
		WHERE jenis_surat_id = ? AND is_general = 1
		LIMIT 1
	`
	t, err = scanTemplateRow(r.db.QueryRowContext(ctx, query, jenisSuratID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal query template umum: %w", err)
	}
	return t, nil
}

// scanTemplateRow scans a surat_template row (kiosk SQLite) into a model,
// including DOCX bytes and the placeholder mapping.
func scanTemplateRow(row *sql.Row) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var updatedAt time.Time
	var isGeneral int
	var formatKertas string
	var docx []byte
	var placeholders sql.NullString

	err := row.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &docx, &placeholders, &isGeneral, &formatKertas, &t.Version, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.IsGeneral = isGeneral == 1
	t.FormatKertas = formatKertas
	t.UpdatedAt = updatedAt
	t.TemplateDocx = docx
	if placeholders.Valid && placeholders.String != "" {
		if err := json.Unmarshal([]byte(placeholders.String), &t.Placeholders); err != nil {
			return nil, fmt.Errorf("gagal unmarshal placeholders: %w", err)
		}
	}
	return &t, nil
}

// UpsertTemplate inserts or updates a template (synced from online backend).
// Supports both per-desa and general templates.
func (r *JenisSuratRepository) UpsertTemplate(ctx context.Context, t *models.SuratTemplate) error {
	// Build query based on whether it's a general template
	var query string
	var args []interface{}

	// Set default format_kertas
	if t.FormatKertas == "" {
		t.FormatKertas = "A4"
	}

	// DOCX bytes (BLOB) and placeholder mapping (JSON TEXT); NULL when absent.
	var docxArg interface{}
	if len(t.TemplateDocx) > 0 {
		docxArg = t.TemplateDocx
	}
	var placeholdersArg interface{}
	if len(t.Placeholders) > 0 {
		b, err := json.Marshal(t.Placeholders)
		if err != nil {
			return fmt.Errorf("gagal marshal placeholders: %w", err)
		}
		placeholdersArg = string(b)
	}

	if t.IsGeneral {
		// General template: unique by jenis_surat_id where is_general = 1
		query = `
			INSERT INTO surat_template (
				id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?)
			ON CONFLICT(jenis_surat_id, desa_id) WHERE is_general = 1 DO UPDATE SET
				id = excluded.id,
				template_html = excluded.template_html,
				template_docx = excluded.template_docx,
				placeholders = excluded.placeholders,
				is_general = 1,
				format_kertas = excluded.format_kertas,
				version = excluded.version,
				updated_at = excluded.updated_at
		`
		args = []interface{}{t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, docxArg, placeholdersArg, t.FormatKertas, t.Version, t.UpdatedAt}
	} else {
		// Per-desa template: unique by (jenis_surat_id, desa_id)
		query = `
			INSERT INTO surat_template (
				id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?)
			ON CONFLICT(jenis_surat_id, desa_id) DO UPDATE SET
				id = excluded.id,
				template_html = excluded.template_html,
				template_docx = excluded.template_docx,
				placeholders = excluded.placeholders,
				is_general = 0,
				format_kertas = excluded.format_kertas,
				version = excluded.version,
				updated_at = excluded.updated_at
		`
		args = []interface{}{t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, docxArg, placeholdersArg, t.FormatKertas, t.Version, t.UpdatedAt}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("gagal upsert template: %w", err)
	}

	return nil
}
