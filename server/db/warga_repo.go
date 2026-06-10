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

func (r *WargaRepository) FindByID(ctx context.Context, id string) (*models.Warga, error) {
	query := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *WargaRepository) FindByNIK(ctx context.Context, NIK string) (*models.Warga, error) {
	query := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE NIK = $1
	`
	row := r.db.QueryRowContext(ctx, query, NIK)
	return r.scanRow(row)
}

func (r *WargaRepository) FindByRFID(ctx context.Context, rfidUID string) (*models.Warga, error) {
	query := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE LOWER(rfid_uid) = LOWER($1)
	`
	row := r.db.QueryRowContext(ctx, query, rfidUID)
	return r.scanRow(row)
}

func (r *WargaRepository) Search(ctx context.Context, query string, desaID string) ([]models.Warga, error) {
	sqlQuery := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE (nama ILIKE $1 OR NIK LIKE $1) AND ($2 = '' OR desa_id = $2::uuid)
		LIMIT 50
	`
	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchTerm, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari warga server: %w", err)
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

	return result, nil
}

func (r *WargaRepository) List(ctx context.Context, desaID string) ([]models.Warga, error) {
	query := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE ($1 = '' OR desa_id = $1::uuid)
		ORDER BY nama ASC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal list warga server: %w", err)
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

	return result, nil
}

func (r *WargaRepository) Create(ctx context.Context, w *models.Warga) error {
	query := `
		INSERT INTO warga (
			id, NIK, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
			alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
			agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`
	var rfid interface{}
	if w.RFIDUID != "" {
		rfid = w.RFIDUID
	}

	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		w.ID, w.NIK, rfid, w.Nama, w.TempatLahir, w.TanggalLahir, w.JenisKelamin,
		w.Alamat, w.RT, w.RW, w.Kelurahan, w.Kecamatan, w.Kabupaten, w.Provinsi,
		w.Agama, w.StatusKawin, w.Pekerjaan, w.Kewarganegaraan, w.DesaID,
		w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal create warga server: %w", err)
	}
	return nil
}

func (r *WargaRepository) Update(ctx context.Context, w *models.Warga) error {
	query := `
		UPDATE warga SET
			NIK = $1,
			rfid_uid = $2,
			nama = $3,
			tempat_lahir = $4,
			tanggal_lahir = $5,
			jenis_kelamin = $6,
			alamat = $7,
			rt = $8,
			rw = $9,
			kelurahan = $10,
			kecamatan = $11,
			kabupaten = $12,
			provinsi = $13,
			agama = $14,
			status_kawin = $15,
			pekerjaan = $16,
			kewarganegaraan = $17,
			updated_at = $18
		WHERE id = $19
	`
	var rfid interface{}
	if w.RFIDUID != "" {
		rfid = w.RFIDUID
	}

	w.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		w.NIK, rfid, w.Nama, w.TempatLahir, w.TanggalLahir, w.JenisKelamin,
		w.Alamat, w.RT, w.RW, w.Kelurahan, w.Kecamatan, w.Kabupaten, w.Provinsi,
		w.Agama, w.StatusKawin, w.Pekerjaan, w.Kewarganegaraan, w.UpdatedAt,
		w.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update warga server: %w", err)
	}
	return nil
}

func (r *WargaRepository) LinkRFID(ctx context.Context, id string, rfidUID string) error {
	query := `
		UPDATE warga
		SET rfid_uid = $1, updated_at = $2
		WHERE id = $3
	`
	var rfid interface{}
	if rfidUID != "" {
		rfid = rfidUID
	}

	_, err := r.db.ExecContext(ctx, query, rfid, time.Now(), id)
	if err != nil {
		return fmt.Errorf("gagal link rfid warga server: %w", err)
	}
	return nil
}

func (r *WargaRepository) scanRow(row *sql.Row) (*models.Warga, error) {
	var w models.Warga
	var rfid sql.NullString
	var tglLahir string

	err := row.Scan(
		&w.ID, &w.NIK, &rfid, &w.Nama, &w.TempatLahir, &tglLahir, &w.JenisKelamin,
		&w.Alamat, &w.RT, &w.RW, &w.Kelurahan, &w.Kecamatan, &w.Kabupaten, &w.Provinsi,
		&w.Agama, &w.StatusKawin, &w.Pekerjaan, &w.Kewarganegaraan, &w.DesaID,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan warga row server: %w", err)
	}

	w.RFIDUID = rfid.String
	// Parse date layout: format date string from postgres
	if parsedTime, err := time.Parse("2006-01-02", tglLahir[:10]); err == nil {
		w.TanggalLahir = parsedTime.Format("2006-01-02")
	} else {
		w.TanggalLahir = tglLahir
	}

	return &w, nil
}

func (r *WargaRepository) scanRows(rows *sql.Rows) (*models.Warga, error) {
	var w models.Warga
	var rfid sql.NullString
	var tglLahir string

	err := rows.Scan(
		&w.ID, &w.NIK, &rfid, &w.Nama, &w.TempatLahir, &tglLahir, &w.JenisKelamin,
		&w.Alamat, &w.RT, &w.RW, &w.Kelurahan, &w.Kecamatan, &w.Kabupaten, &w.Provinsi,
		&w.Agama, &w.StatusKawin, &w.Pekerjaan, &w.Kewarganegaraan, &w.DesaID,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan warga rows server: %w", err)
	}

	w.RFIDUID = rfid.String
	if parsedTime, err := time.Parse("2006-01-02", tglLahir[:10]); err == nil {
		w.TanggalLahir = parsedTime.Format("2006-01-02")
	} else {
		w.TanggalLahir = tglLahir
	}

	return &w, nil
}

func (r *WargaRepository) ListUpdatedSince(ctx context.Context, desaID string, since time.Time) ([]models.Warga, error) {
	query := `
		SELECT id, nik, rfid_uid, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		       alamat, rt, rw, kelurahan, kecamatan, kabupaten, provinsi,
		       agama, status_kawin, pekerjaan, kewarganegaraan, desa_id,
		       created_at, updated_at
		FROM warga
		WHERE ($1 = '' OR desa_id = $1::uuid) AND updated_at > $2
		ORDER BY updated_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID, since)
	if err != nil {
		return nil, fmt.Errorf("gagal list warga updated since: %w", err)
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

	return result, nil
}
