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

func (r *JenisSuratRepository) FindByID(ctx context.Context, id string) (*models.JenisSurat, error) {
	query := `
		SELECT id, kode, nama, deskripsi, fields_schema, aktif_global, urutan, created_at, updated_at
		FROM jenis_surat
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *JenisSuratRepository) FindByKode(ctx context.Context, kode string) (*models.JenisSurat, error) {
	query := `
		SELECT id, kode, nama, deskripsi, fields_schema, aktif_global, urutan, created_at, updated_at
		FROM jenis_surat
		WHERE kode = $1
	`
	row := r.db.QueryRowContext(ctx, query, kode)
	return r.scanRow(row)
}

func (r *JenisSuratRepository) List(ctx context.Context) ([]models.JenisSurat, error) {
	query := `
		SELECT id, kode, nama, deskripsi, fields_schema, aktif_global, urutan, created_at, updated_at
		FROM jenis_surat
		ORDER BY urutan ASC, nama ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal list jenis surat server: %w", err)
	}
	defer rows.Close()

	var result []models.JenisSurat
	for rows.Next() {
		js, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *js)
	}
	return result, nil
}

func (r *JenisSuratRepository) ListActiveForDesa(ctx context.Context, desaID string) ([]models.JenisSurat, error) {
	query := `
		SELECT js.id, js.kode, js.nama, js.deskripsi, js.fields_schema, djs.aktif as aktif_global, djs.urutan, js.created_at, js.updated_at
		FROM jenis_surat js
		JOIN desa_jenis_surat djs ON js.id = djs.jenis_surat_id
		WHERE djs.desa_id = $1 AND djs.aktif = true
		ORDER BY djs.urutan ASC, js.nama ASC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal list jenis surat active desa server: %w", err)
	}
	defer rows.Close()

	var result []models.JenisSurat
	for rows.Next() {
		js, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *js)
	}
	return result, nil
}

func (r *JenisSuratRepository) Create(ctx context.Context, js *models.JenisSurat) error {
	query := `
		INSERT INTO jenis_surat (
			id, kode, nama, deskripsi, fields_schema, aktif_global, urutan, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	js.CreatedAt = now
	js.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		js.ID, js.Kode, js.Nama, js.Deskripsi, string(js.FieldsSchema), js.Aktif, js.Urutan, js.CreatedAt, js.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal create jenis surat: %w", err)
	}
	return nil
}

func (r *JenisSuratRepository) Update(ctx context.Context, js *models.JenisSurat) error {
	query := `
		UPDATE jenis_surat SET
			kode = $1,
			nama = $2,
			deskripsi = $3,
			fields_schema = $4,
			aktif_global = $5,
			urutan = $6,
			updated_at = $7
		WHERE id = $8
	`
	js.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		js.Kode, js.Nama, js.Deskripsi, string(js.FieldsSchema), js.Aktif, js.Urutan, js.UpdatedAt, js.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update jenis surat: %w", err)
	}
	return nil
}

func (r *JenisSuratRepository) ToggleForDesa(ctx context.Context, desaID string, jenisSuratID string, aktif bool, urutan int) error {
	query := `
		INSERT INTO desa_jenis_surat (desa_id, jenis_surat_id, aktif, urutan)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (desa_id, jenis_surat_id) DO UPDATE SET
			aktif = excluded.aktif,
			urutan = excluded.urutan
	`
	_, err := r.db.ExecContext(ctx, query, desaID, jenisSuratID, aktif, urutan)
	if err != nil {
		return fmt.Errorf("gagal toggle jenis surat desa: %w", err)
	}
	return nil
}

func (r *JenisSuratRepository) scanRow(row *sql.Row) (*models.JenisSurat, error) {
	var js models.JenisSurat
	var deskripsi sql.NullString
	var schemaStr string

	err := row.Scan(
		&js.ID, &js.Kode, &js.Nama, &deskripsi, &schemaStr, &js.Aktif, &js.Urutan, &js.CreatedAt, &js.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan jenis surat row: %w", err)
	}

	js.Deskripsi = deskripsi.String
	js.FieldsSchema = json.RawMessage(schemaStr)

	return &js, nil
}

func (r *JenisSuratRepository) scanRows(rows *sql.Rows) (*models.JenisSurat, error) {
	var js models.JenisSurat
	var deskripsi sql.NullString
	var schemaStr string

	err := rows.Scan(
		&js.ID, &js.Kode, &js.Nama, &deskripsi, &schemaStr, &js.Aktif, &js.Urutan, &js.CreatedAt, &js.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan jenis surat rows: %w", err)
	}

	js.Deskripsi = deskripsi.String
	js.FieldsSchema = json.RawMessage(schemaStr)

	return &js, nil
}
