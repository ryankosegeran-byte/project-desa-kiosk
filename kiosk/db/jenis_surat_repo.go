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
//
// The server hub is the source of truth for the row id. A jenis_surat is
// identified by its business key `kode`, so when a row with the same `kode`
// already exists locally with a different `id`, the local `id` is realigned to
// the server id. Child rows that reference the old id (surat_template, surat,
// nomor_surat_batch) are repointed to the new id inside the same transaction so
// foreign keys stay valid.
func (r *JenisSuratRepository) Upsert(ctx context.Context, js *models.JenisSurat) error {
	schemaBytes, err := json.Marshal(js.FieldsSchema)
	if err != nil {
		return fmt.Errorf("gagal marshal fields_schema: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi upsert jenis_surat: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Defer foreign key checks until commit so we can realign the parent id and
	// repoint child rows within the same transaction regardless of statement
	// order. Checks still run at COMMIT, so a genuinely broken graph is rejected.
	if _, err := tx.ExecContext(ctx, `PRAGMA defer_foreign_keys = ON`); err != nil {
		return fmt.Errorf("gagal mengaktifkan defer_foreign_keys: %w", err)
	}

	// Look up the existing local id for this kode (business key).
	var existingID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM jenis_surat WHERE kode = ?`, js.Kode).Scan(&existingID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		existingID = ""
	case err != nil:
		return fmt.Errorf("gagal cek jenis_surat berdasarkan kode: %w", err)
	}

	// If the kode exists with a different id, realign children before changing
	// the parent id so foreign key references remain valid.
	if existingID != "" && existingID != js.ID {
		if err := repointJenisSuratChildren(ctx, tx, existingID, js.ID); err != nil {
			return err
		}
	}

	query := `
		INSERT INTO jenis_surat (
			id, kode, nama, deskripsi, fields_schema, aktif, urutan, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(kode) DO UPDATE SET
			id = excluded.id,
			nama = excluded.nama,
			deskripsi = excluded.deskripsi,
			fields_schema = excluded.fields_schema,
			aktif = excluded.aktif,
			urutan = excluded.urutan,
			updated_at = excluded.updated_at
	`
	if _, err := tx.ExecContext(ctx, query,
		js.ID, js.Kode, js.Nama, js.Deskripsi, string(schemaBytes), js.Aktif, js.Urutan, js.UpdatedAt,
	); err != nil {
		return fmt.Errorf("gagal upsert jenis_surat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal commit upsert jenis_surat: %w", err)
	}
	return nil
}

// repointJenisSuratChildren updates child tables that reference jenis_surat(id)
// from oldID to newID. Run inside a transaction before changing the parent id.
func repointJenisSuratChildren(ctx context.Context, tx *sql.Tx, oldID, newID string) error {
	stmts := []struct {
		table string
		query string
	}{
		{"surat_template", `UPDATE surat_template SET jenis_surat_id = ? WHERE jenis_surat_id = ?`},
		{"surat", `UPDATE surat SET jenis_surat_id = ? WHERE jenis_surat_id = ?`},
		{"nomor_surat_batch", `UPDATE nomor_surat_batch SET jenis_surat_id = ? WHERE jenis_surat_id = ?`},
	}
	for _, s := range stmts {
		if _, err := tx.ExecContext(ctx, s.query, newID, oldID); err != nil {
			return fmt.Errorf("gagal repoint %s ke id jenis_surat baru: %w", s.table, err)
		}
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
//
// The sync puller normalises desa_id to the kiosk's own desa before calling
// this method, so the incoming template may carry a different desa_id than
// the row already stored locally (which was saved with the server's original
// desa_id). To avoid a PRIMARY KEY clash we first DELETE any stale row that
// has the same id but a different (jenis_surat_id, desa_id) pair, then use
// INSERT … ON CONFLICT to handle the normal upsert path.
func (r *JenisSuratRepository) UpsertTemplate(ctx context.Context, t *models.SuratTemplate) error {
	if t.FormatKertas == "" {
		t.FormatKertas = "A4"
	}

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

	// Remove any stale row that already has this id but with a different
	// (jenis_surat_id, desa_id) combination. This prevents a PK collision
	// when the puller re-maps desa_id to the kiosk's own desa.
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM surat_template WHERE id = ? AND NOT (jenis_surat_id = ? AND desa_id = ?)`,
		t.ID, t.JenisSuratID, t.DesaID,
	)
	if err != nil {
		return fmt.Errorf("gagal hapus template stale: %w", err)
	}

	isGeneral := 0
	if t.IsGeneral {
		isGeneral = 1
	}

	query := `
		INSERT INTO surat_template (
			id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(jenis_surat_id, desa_id) DO UPDATE SET
			id = excluded.id,
			template_html = excluded.template_html,
			template_docx = excluded.template_docx,
			placeholders = excluded.placeholders,
			is_general = excluded.is_general,
			format_kertas = excluded.format_kertas,
			version = excluded.version,
			updated_at = excluded.updated_at
	`
	args := []interface{}{t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, docxArg, placeholdersArg, isGeneral, t.FormatKertas, t.Version, t.UpdatedAt}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("gagal upsert template: %w", err)
	}

	return nil
}
