package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type DesaRepository struct {
	db *DB
}

func NewDesaRepository(db *DB) *DesaRepository {
	return &DesaRepository{db: db}
}

func (r *DesaRepository) FindByID(ctx context.Context, id string) (*models.Desa, error) {
	query := `
		SELECT id, nama, kode_desa, kecamatan, kabupaten, provinsi, kepala_desa, nip_kepala_desa, alamat_kantor, logo_path, created_at, updated_at
		FROM desa
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *DesaRepository) FindByKode(ctx context.Context, kode string) (*models.Desa, error) {
	query := `
		SELECT id, nama, kode_desa, kecamatan, kabupaten, provinsi, kepala_desa, nip_kepala_desa, alamat_kantor, logo_path, created_at, updated_at
		FROM desa
		WHERE kode_desa = $1
	`
	row := r.db.QueryRowContext(ctx, query, kode)
	return r.scanRow(row)
}

func (r *DesaRepository) Create(ctx context.Context, d *models.Desa) error {
	query := `
		INSERT INTO desa (
			id, nama, kode_desa, kecamatan, kabupaten, provinsi, kepala_desa, nip_kepala_desa, alamat_kantor, logo_path, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.Nama, d.KodeDesa, d.Kecamatan, d.Kabupaten, d.Provinsi, d.KepalaDesa, d.NIPKepalaDesa, d.AlamatKantor, d.LogoPath, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal create desa: %w", err)
	}
	return nil
}

func (r *DesaRepository) List(ctx context.Context) ([]models.Desa, error) {
	query := `
		SELECT id, nama, kode_desa, kecamatan, kabupaten, provinsi, kepala_desa, nip_kepala_desa, alamat_kantor, logo_path, created_at, updated_at
		FROM desa
		ORDER BY nama ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal list desa: %w", err)
	}
	defer rows.Close()

	var result []models.Desa
	for rows.Next() {
		d, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, nil
}

// ==========================================
// Kiosk Registry Methods
// ==========================================

func (r *DesaRepository) RegisterKiosk(ctx context.Context, k *models.Kiosk) error {
	query := `
		INSERT INTO kiosks (id, desa_id, nama, api_key, last_seen_at, last_sync_at, status, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	var lastSeen, lastSync interface{}
	if k.LastSeenAt != nil {
		lastSeen = *k.LastSeenAt
	}
	if k.LastSyncAt != nil {
		lastSync = *k.LastSyncAt
	}

	now := time.Now()
	k.CreatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		k.ID, k.DesaID, k.Nama, k.APIKey, lastSeen, lastSync, k.Status, k.IPAddress, k.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal register kiosk: %w", err)
	}
	return nil
}

func (r *DesaRepository) FindKioskByID(ctx context.Context, id string) (*models.Kiosk, error) {
	query := `
		SELECT id, desa_id, nama, api_key, last_seen_at, last_sync_at, status, ip_address, created_at
		FROM kiosks
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanKioskRow(row)
}

func (r *DesaRepository) FindKioskByKey(ctx context.Context, apiKey string) (*models.Kiosk, error) {
	query := `
		SELECT id, desa_id, nama, api_key, last_seen_at, last_sync_at, status, ip_address, created_at
		FROM kiosks
		WHERE api_key = $1
	`
	row := r.db.QueryRowContext(ctx, query, apiKey)
	return r.scanKioskRow(row)
}

func (r *DesaRepository) ListKiosks(ctx context.Context, desaID string) ([]models.Kiosk, error) {
	query := `
		SELECT id, desa_id, nama, api_key, last_seen_at, last_sync_at, status, ip_address, created_at
		FROM kiosks
		WHERE ($1 = '' OR desa_id = $1)
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, desaID)
	if err != nil {
		return nil, fmt.Errorf("gagal list kiosks: %w", err)
	}
	defer rows.Close()

	var result []models.Kiosk
	for rows.Next() {
		k, err := r.scanKioskRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *k)
	}
	return result, nil
}

func (r *DesaRepository) UpdateKioskSyncTime(ctx context.Context, kioskID string) error {
	query := `
		UPDATE kiosks
		SET last_sync_at = $1, last_seen_at = $2
		WHERE id = $3
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, kioskID)
	if err != nil {
		return fmt.Errorf("gagal update kiosk sync time: %w", err)
	}
	return nil
}

func (r *DesaRepository) UpdateKioskStatus(ctx context.Context, kioskID string, ipAddress string) error {
	query := `
		UPDATE kiosks
		SET last_seen_at = $1, ip_address = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), ipAddress, kioskID)
	if err != nil {
		return fmt.Errorf("gagal update kiosk status: %w", err)
	}
	return nil
}

// ==========================================
// Helpers
// ==========================================

func (r *DesaRepository) scanRow(row *sql.Row) (*models.Desa, error) {
	var d models.Desa
	var logo sql.NullString

	err := row.Scan(
		&d.ID, &d.Nama, &d.KodeDesa, &d.Kecamatan, &d.Kabupaten, &d.Provinsi, &d.KepalaDesa, &d.NIPKepalaDesa, &d.AlamatKantor, &logo, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan desa row: %w", err)
	}
	d.LogoPath = logo.String
	return &d, nil
}

func (r *DesaRepository) scanRows(rows *sql.Rows) (*models.Desa, error) {
	var d models.Desa
	var logo sql.NullString

	err := rows.Scan(
		&d.ID, &d.Nama, &d.KodeDesa, &d.Kecamatan, &d.Kabupaten, &d.Provinsi, &d.KepalaDesa, &d.NIPKepalaDesa, &d.AlamatKantor, &logo, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan desa rows: %w", err)
	}
	d.LogoPath = logo.String
	return &d, nil
}

func (r *DesaRepository) scanKioskRow(row *sql.Row) (*models.Kiosk, error) {
	var k models.Kiosk
	var nama, apiKey, ip sql.NullString
	var lastSeen, lastSync sql.NullTime

	err := row.Scan(
		&k.ID, &k.DesaID, &nama, &apiKey, &lastSeen, &lastSync, &k.Status, &ip, &k.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan kiosk row: %w", err)
	}

	k.Nama = nama.String
	k.APIKey = apiKey.String
	k.IPAddress = ip.String
	if lastSeen.Valid {
		k.LastSeenAt = &lastSeen.Time
	}
	if lastSync.Valid {
		k.LastSyncAt = &lastSync.Time
	}

	return &k, nil
}

func (r *DesaRepository) scanKioskRows(rows *sql.Rows) (*models.Kiosk, error) {
	var k models.Kiosk
	var nama, apiKey, ip sql.NullString
	var lastSeen, lastSync sql.NullTime

	err := rows.Scan(
		&k.ID, &k.DesaID, &nama, &apiKey, &lastSeen, &lastSync, &k.Status, &ip, &k.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan kiosk rows: %w", err)
	}

	k.Nama = nama.String
	k.APIKey = apiKey.String
	k.IPAddress = ip.String
	if lastSeen.Valid {
		k.LastSeenAt = &lastSeen.Time
	}
	if lastSync.Valid {
		k.LastSyncAt = &lastSync.Time
	}

	return &k, nil
}
