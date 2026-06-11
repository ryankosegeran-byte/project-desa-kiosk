package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type WargaRepository struct {
	db *DB
}

func NewWargaRepository(db *DB) *WargaRepository {
	return &WargaRepository{db: db}
}

const wargaSelectCols = `id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
	alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
	agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
	foto_ktp_path, status, draft_token,
	created_at, updated_at, synced_at`

// FindByRFID looks up a resident by their RFID UID (case-insensitive).
func (r *WargaRepository) FindByRFID(ctx context.Context, rfidUID string) (*models.Warga, error) {
	query := `SELECT ` + wargaSelectCols + ` FROM warga WHERE LOWER(rfid_uid) = LOWER(?)`
	row := r.db.QueryRowContext(ctx, query, rfidUID)
	return r.scanRow(row)
}

// FindByNIK looks up a resident by their NIK.
func (r *WargaRepository) FindByNIK(ctx context.Context, NIK string) (*models.Warga, error) {
	query := `SELECT ` + wargaSelectCols + ` FROM warga WHERE NIK = ?`
	row := r.db.QueryRowContext(ctx, query, NIK)
	return r.scanRow(row)
}

// Search searches for residents by name or NIK.
func (r *WargaRepository) Search(ctx context.Context, query string) ([]models.Warga, error) {
	sqlQuery := `SELECT ` + wargaSelectCols + ` FROM warga WHERE nama LIKE ? OR nik LIKE ? LIMIT 20`
	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari warga: %w", err)
	}
	defer rows.Close()

	var result []models.Warga
	for rows.Next() {
		w, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Upsert inserts or updates a resident record (synced from online backend).
func (r *WargaRepository) Upsert(ctx context.Context, w *models.Warga) error {
	query := `
		INSERT INTO warga (
			id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
			alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
			agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
			foto_ktp_path, status, draft_token,
			created_at, updated_at, synced_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			nik = excluded.nik,
			rfid_uid = excluded.rfid_uid,
			nama = excluded.nama,
			tempat_lahir = excluded.tempat_lahir,
			tanggal_lahir = excluded.tanggal_lahir,
			jenis_kelamin = excluded.jenis_kelamin,
			alamat = excluded.alamat,
			rt = excluded.rt,
			rw = excluded.rw,
			kelurahan = excluded.kelurahan,
			kecamatan = excluded.kecamatan,
			kabupaten = excluded.kabupaten,
			provinsi = excluded.provinsi,
			agama = excluded.agama,
			status_kawin = excluded.status_kawin,
			pekerjaan = excluded.pekerjaan,
			kewarganegaraan = excluded.kewarganegaraan,
			desa_id = excluded.desa_id,
			foto_ktp_path = excluded.foto_ktp_path,
			status = excluded.status,
			draft_token = excluded.draft_token,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			synced_at = excluded.synced_at
	`

	var rfid interface{}
	if w.RFIDUID != "" {
		rfid = w.RFIDUID
	}

	var syncedAt interface{}
	if w.SyncedAt != nil {
		syncedAt = *w.SyncedAt
	} else {
		now := time.Now()
		syncedAt = now
	}

	var fotoKTP interface{}
	if w.FotoKTPPath != "" {
		fotoKTP = w.FotoKTPPath
	}

	var status interface{}
	if w.Status != "" {
		status = w.Status
	}

	var draftToken interface{}
	if w.DraftToken != "" {
		draftToken = w.DraftToken
	}

	_, err := r.db.ExecContext(ctx, query,
		w.ID, w.NIK, rfid, w.Nama, w.TempatLahir, w.TanggalLahir, w.JenisKelamin,
		w.Alamat, w.RT, w.RW, w.Kelurahan, w.Kecamatan, w.Kabupaten, w.Provinsi,
		w.Agama, w.StatusKawin, w.Pekerjaan, w.Kewarganegaraan, w.DesaID,
		fotoKTP, status, draftToken,
		w.CreatedAt, w.UpdatedAt, syncedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert warga: %w", err)
	}

	return nil
}

// Helpers for scanning row
func (r *WargaRepository) scanRow(row *sql.Row) (*models.Warga, error) {
	var w models.Warga
	var rfid, fotoKTP, draftToken sql.NullString
	var createdAt, updatedAt time.Time
	var syncedAt sql.NullTime

	err := row.Scan(
		&w.ID, &w.NIK, &rfid, &w.Nama, &w.TempatLahir, &w.TanggalLahir, &w.JenisKelamin,
		&w.Alamat, &w.RT, &w.RW, &w.Kelurahan, &w.Kecamatan, &w.Kabupaten, &w.Provinsi,
		&w.Agama, &w.StatusKawin, &w.Pekerjaan, &w.Kewarganegaraan, &w.DesaID,
		&fotoKTP, &w.Status, &draftToken,
		&createdAt, &updatedAt, &syncedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan warga: %w", err)
	}

	w.RFIDUID = rfid.String
	w.FotoKTPPath = fotoKTP.String
	w.DraftToken = draftToken.String
	w.CreatedAt = createdAt
	w.UpdatedAt = updatedAt
	if syncedAt.Valid {
		w.SyncedAt = &syncedAt.Time
	}

	return &w, nil
}

func (r *WargaRepository) scanRows(rows *sql.Rows) (*models.Warga, error) {
	var w models.Warga
	var rfid, fotoKTP, draftToken sql.NullString
	var createdAt, updatedAt time.Time
	var syncedAt sql.NullTime

	err := rows.Scan(
		&w.ID, &w.NIK, &rfid, &w.Nama, &w.TempatLahir, &w.TanggalLahir, &w.JenisKelamin,
		&w.Alamat, &w.RT, &w.RW, &w.Kelurahan, &w.Kecamatan, &w.Kabupaten, &w.Provinsi,
		&w.Agama, &w.StatusKawin, &w.Pekerjaan, &w.Kewarganegaraan, &w.DesaID,
		&fotoKTP, &w.Status, &draftToken,
		&createdAt, &updatedAt, &syncedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("gagal scan warga row: %w", err)
	}

	w.RFIDUID = rfid.String
	w.FotoKTPPath = fotoKTP.String
	w.DraftToken = draftToken.String
	w.CreatedAt = createdAt
	w.UpdatedAt = updatedAt
	if syncedAt.Valid {
		w.SyncedAt = &syncedAt.Time
	}

	return &w, nil
}
