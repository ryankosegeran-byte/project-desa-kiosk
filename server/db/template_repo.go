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

type TemplateRepository struct {
	db *DB
}

func NewTemplateRepository(db *DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// detailColumns includes the heavy DOCX bytes + placeholder mapping; used when a
// single full template is fetched (detail / render / kiosk).
const detailColumns = `id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, created_by, created_at, updated_at`

// listColumns omits template_docx (can be hundreds of KB) so list endpoints stay light.
const listColumns = `id, jenis_surat_id, desa_id, template_html, is_general, format_kertas, version, created_by, created_at, updated_at`

// GetTemplate retrieves template with hierarchy: per-desa > general > fallback
// Priority: 1. Per-desa template, 2. General template, 3. Error if none exists
func (r *TemplateRepository) GetTemplate(ctx context.Context, jenisSuratID string, desaID string) (*models.SuratTemplate, error) {
	// First try to get per-desa template
	query := `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE jenis_surat_id = $1 AND desa_id = $2 AND is_general = false
	`
	row := r.db.QueryRowContext(ctx, query, jenisSuratID, desaID)
	template, err := r.scanRow(row)
	if err == nil {
		return template, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Fallback to general template
	query = `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE jenis_surat_id = $1 AND is_general = true
		LIMIT 1
	`
	row = r.db.QueryRowContext(ctx, query, jenisSuratID)
	return r.scanRow(row)
}

// GetTemplateWithFallback retrieves template, returning optional (no error if not found)
// Useful for checking if per-desa override exists
func (r *TemplateRepository) GetTemplateWithFallback(ctx context.Context, jenisSuratID string, desaID string) (*models.SuratTemplate, bool, error) {
	// Check per-desa template first
	query := `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE jenis_surat_id = $1 AND desa_id = $2 AND is_general = false
	`
	row := r.db.QueryRowContext(ctx, query, jenisSuratID, desaID)
	template, err := r.scanRow(row)
	if err == nil {
		return template, true, nil // found per-desa
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	// Check general template
	query = `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE jenis_surat_id = $1 AND is_general = true
		LIMIT 1
	`
	row = r.db.QueryRowContext(ctx, query, jenisSuratID)
	template, err = r.scanRow(row)
	if err == nil {
		return template, false, nil // found general
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil // not found
	}
	return nil, false, err
}

// UpsertTemplate creates or updates a template (general or per-desa).
// Per-desa templates are unique on (jenis_surat_id, desa_id); general templates
// are unique on (jenis_surat_id) via a partial index — so the ON CONFLICT target
// differs per case (a single statement cannot carry two ON CONFLICT clauses).
func (r *TemplateRepository) UpsertTemplate(ctx context.Context, t *models.SuratTemplate) error {
	if t.FormatKertas == "" {
		t.FormatKertas = models.FormatKertasA4
	}

	conflictTarget := "(jenis_surat_id, desa_id)"
	if t.IsGeneral {
		conflictTarget = "(jenis_surat_id) WHERE is_general = true"
	}

	query := `
		INSERT INTO surat_template (id, jenis_surat_id, desa_id, template_html, template_docx, placeholders, is_general, format_kertas, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT ` + conflictTarget + ` DO UPDATE SET
			template_html = excluded.template_html,
			template_docx = excluded.template_docx,
			placeholders = excluded.placeholders,
			is_general = excluded.is_general,
			format_kertas = excluded.format_kertas,
			version = surat_template.version + 1,
			created_by = excluded.created_by,
			updated_at = excluded.updated_at
	`

	var createdBy interface{}
	if t.CreatedBy != "" {
		createdBy = t.CreatedBy
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
		placeholdersArg = b
	}

	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.JenisSuratID, t.DesaID, t.TemplateHTML, docxArg, placeholdersArg, t.IsGeneral, t.FormatKertas, t.Version, createdBy, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert template: %w", err)
	}
	return nil
}

// UpdatePlaceholders replaces the placeholder mapping for a template by id,
// leaving the DOCX bytes and HTML untouched.
func (r *TemplateRepository) UpdatePlaceholders(ctx context.Context, templateID string, placeholders []models.PlaceholderDef) error {
	b, err := json.Marshal(placeholders)
	if err != nil {
		return fmt.Errorf("gagal marshal placeholders: %w", err)
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE surat_template SET placeholders = $1, version = version + 1, updated_at = now() WHERE id = $2`,
		b, templateID,
	)
	if err != nil {
		return fmt.Errorf("gagal update placeholders: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListTemplatesForDesa lists all templates for a specific desa (including general ones)
func (r *TemplateRepository) ListTemplatesForDesa(ctx context.Context, desaID string) ([]models.SuratTemplate, error) {
	query := `
		SELECT ` + listColumns + `
		FROM surat_template
		WHERE desa_id = $1 OR is_general = true
		ORDER BY is_general ASC, updated_at DESC
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

// ListTemplatesForDesaFull is like ListTemplatesForDesa but includes the DOCX
// bytes and placeholder mapping — used by kiosk sync so templates render offline.
func (r *TemplateRepository) ListTemplatesForDesaFull(ctx context.Context, desaID string) ([]models.SuratTemplate, error) {
	query := `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE desa_id = $1 OR is_general = true
		ORDER BY is_general ASC, updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal list templates (full): %w", err)
	}
	defer rows.Close()

	var result []models.SuratTemplate
	for rows.Next() {
		t, err := r.scanRowsFull(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *t)
	}
	return result, nil
}

// ListGeneralTemplates lists all general templates
func (r *TemplateRepository) ListGeneralTemplates(ctx context.Context) ([]models.SuratTemplate, error) {
	query := `
		SELECT ` + listColumns + `
		FROM surat_template
		WHERE is_general = true
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal list general templates: %w", err)
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

// SaveTemplateVersion saves a version history entry
func (r *TemplateRepository) SaveTemplateVersion(ctx context.Context, history *models.TemplateVersionHistory) error {
	query := `
		INSERT INTO template_version_history (id, template_id, version, template_html, format_kertas, change_note, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var createdBy interface{}
	if history.CreatedBy != "" {
		createdBy = history.CreatedBy
	}

	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.TemplateID, history.Version, history.TemplateHTML, history.FormatKertas, history.ChangeNote, createdBy, history.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal save template version: %w", err)
	}
	return nil
}

// GetTemplateVersions gets version history for a template
func (r *TemplateRepository) GetTemplateVersions(ctx context.Context, templateID string) ([]models.TemplateVersionHistory, error) {
	query := `
		SELECT id, template_id, version, template_html, format_kertas, change_note, created_by, created_at
		FROM template_version_history
		WHERE template_id = $1
		ORDER BY version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("gagal get template versions: %w", err)
	}
	defer rows.Close()

	var result []models.TemplateVersionHistory
	for rows.Next() {
		var v models.TemplateVersionHistory
		var changeNote, createdBy sql.NullString
		err := rows.Scan(&v.ID, &v.TemplateID, &v.Version, &v.TemplateHTML, &v.FormatKertas, &changeNote, &createdBy, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("gagal scan version row: %w", err)
		}
		v.ChangeNote = changeNote.String
		v.CreatedBy = createdBy.String
		result = append(result, v)
	}
	return result, nil
}

// GetGeneralTemplate retrieves a general template for a jenis surat
func (r *TemplateRepository) GetGeneralTemplate(ctx context.Context, jenisSuratID string) (*models.SuratTemplate, error) {
	query := `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE jenis_surat_id = $1 AND is_general = true
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, jenisSuratID)
	return r.scanRow(row)
}

// GetTemplateByID retrieves a single template by its primary key (UUID),
// including the DOCX blob and placeholders.
func (r *TemplateRepository) GetTemplateByID(ctx context.Context, id string) (*models.SuratTemplate, error) {
	query := `
		SELECT ` + detailColumns + `
		FROM surat_template
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

// scanRow scans a full template row including DOCX bytes and placeholder mapping.
func (r *TemplateRepository) scanRow(row *sql.Row) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var createdBy sql.NullString
	var docx []byte
	var placeholders []byte

	err := row.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &docx, &placeholders, &t.IsGeneral, &t.FormatKertas, &t.Version, &createdBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan template row: %w", err)
	}

	t.CreatedBy = createdBy.String
	t.TemplateDocx = docx
	if len(placeholders) > 0 {
		if err := json.Unmarshal(placeholders, &t.Placeholders); err != nil {
			return nil, fmt.Errorf("gagal unmarshal placeholders: %w", err)
		}
	}
	return &t, nil
}

// scanRows scans a template row for list endpoints (no DOCX bytes / placeholders).
func (r *TemplateRepository) scanRows(rows *sql.Rows) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var createdBy sql.NullString

	err := rows.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &t.IsGeneral, &t.FormatKertas, &t.Version, &createdBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan template rows: %w", err)
	}

	t.CreatedBy = createdBy.String
	return &t, nil
}

// scanRowsFull scans a full template row (incl. DOCX bytes + placeholders) from *sql.Rows.
func (r *TemplateRepository) scanRowsFull(rows *sql.Rows) (*models.SuratTemplate, error) {
	var t models.SuratTemplate
	var createdBy sql.NullString
	var docx []byte
	var placeholders []byte

	err := rows.Scan(
		&t.ID, &t.JenisSuratID, &t.DesaID, &t.TemplateHTML, &docx, &placeholders, &t.IsGeneral, &t.FormatKertas, &t.Version, &createdBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan template rows (full): %w", err)
	}

	t.CreatedBy = createdBy.String
	t.TemplateDocx = docx
	if len(placeholders) > 0 {
		if err := json.Unmarshal(placeholders, &t.Placeholders); err != nil {
			return nil, fmt.Errorf("gagal unmarshal placeholders: %w", err)
		}
	}
	return &t, nil
}
