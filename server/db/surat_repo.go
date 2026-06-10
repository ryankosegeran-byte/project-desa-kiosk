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

type SuratRepository struct {
	db *DB
}

func NewSuratRepository(db *DB) *SuratRepository {
	return &SuratRepository{db: db}
}

func (r *SuratRepository) Create(ctx context.Context, s *models.Surat) error {
	query := `
		INSERT INTO surat (
			id, nomor_surat, jenis_surat_id, warga_id, nik_pemohon, nama_pemohon,
			data_surat, status, desa_id, kiosk_id, kiosk_created_at, synced_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	var wargaID, kioskID interface{}
	if s.WargaID != "" {
		wargaID = s.WargaID
	}
	// We will handle kioskID mapping if synced from kiosk. For direct server creations it is null.

	now := time.Now()
	s.CreatedAt = now
	s.SyncedAt = &now

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.NomorSurat, s.JenisSuratID, wargaID, s.NIKPemohon, s.NamaPemohon,
		string(s.DataSurat), s.Status, s.DesaID, kioskID, s.CreatedAt, s.SyncedAt, now,
	)
	if err != nil {
		return fmt.Errorf("gagal create surat server: %w", err)
	}
	return nil
}

func (r *SuratRepository) FindByID(ctx context.Context, id string) (*models.Surat, error) {
	query := `
		SELECT id, nomor_surat, jenis_surat_id, warga_id, nik_pemohon, nama_pemohon,
		       data_surat, status, desa_id, kiosk_created_at, synced_at, created_at
		FROM surat
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *SuratRepository) List(ctx context.Context, desaID string, limit, offset int) ([]models.Surat, error) {
	query := `
		SELECT s.id, s.nomor_surat, s.jenis_surat_id, s.warga_id, s.nik_pemohon, s.nama_pemohon,
		       s.data_surat, s.status, s.desa_id, s.kiosk_created_at, s.synced_at, s.created_at,
		       js.kode as jenis_surat_kode, js.nama as jenis_surat_nama
		FROM surat s
		JOIN jenis_surat js ON s.jenis_surat_id = js.id
		WHERE ($1 = '' OR s.desa_id = $1::uuid)
		ORDER BY s.kiosk_created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, desaID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("gagal list surat server: %w", err)
	}
	defer rows.Close()

	var result []models.Surat
	for rows.Next() {
		s, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *s)
	}
	return result, nil
}

func (r *SuratRepository) Upsert(ctx context.Context, s *models.Surat, kioskID string) error {
	query := `
		INSERT INTO surat (
			id, nomor_surat, jenis_surat_id, warga_id, nik_pemohon, nama_pemohon,
			data_surat, status, desa_id, kiosk_id, kiosk_created_at, synced_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			nomor_surat = excluded.nomor_surat,
			status = excluded.status,
			synced_at = excluded.synced_at
	`
	var wargaID interface{}
	if s.WargaID != "" {
		wargaID = s.WargaID
	}
	var kID interface{}
	if kioskID != "" {
		kID = kioskID
	}

	now := time.Now()
	s.SyncedAt = &now

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.NomorSurat, s.JenisSuratID, wargaID, s.NIKPemohon, s.NamaPemohon,
		string(s.DataSurat), s.Status, s.DesaID, kID, s.CreatedAt, s.SyncedAt, now,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert surat server: %w", err)
	}
	return nil
}

func (r *SuratRepository) scanRow(row *sql.Row) (*models.Surat, error) {
	var s models.Surat
	var wargaID sql.NullString
	var dataSuratStr string
	var syncedAt sql.NullTime

	err := row.Scan(
		&s.ID, &s.NomorSurat, &s.JenisSuratID, &wargaID, &s.NIKPemohon, &s.NamaPemohon,
		&dataSuratStr, &s.Status, &s.DesaID, &s.CreatedAt, &syncedAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan surat row server: %w", err)
	}

	s.WargaID = wargaID.String
	s.DataSurat = json.RawMessage(dataSuratStr)
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}

func (r *SuratRepository) scanRows(rows *sql.Rows) (*models.Surat, error) {
	var s models.Surat
	var wargaID sql.NullString
	var dataSuratStr string
	var syncedAt sql.NullTime

	err := rows.Scan(
		&s.ID, &s.NomorSurat, &s.JenisSuratID, &wargaID, &s.NIKPemohon, &s.NamaPemohon,
		&dataSuratStr, &s.Status, &s.DesaID, &s.CreatedAt, &syncedAt, &s.CreatedAt,
		&s.JenisSuratKode, &s.JenisSuratNama,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan surat rows server: %w", err)
	}

	s.WargaID = wargaID.String
	s.DataSurat = json.RawMessage(dataSuratStr)
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}
