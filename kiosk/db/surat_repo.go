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

// Create inserts a new surat record in DRAFT status.
func (r *SuratRepository) Create(ctx context.Context, s *models.Surat) error {
	query := `
		INSERT INTO surat (
			id, nomor_surat, jenis_surat_id, jenis_surat_kode, jenis_surat_nama,
			warga_id, nik_pemohon, nama_pemohon, data_surat, status,
			pdf_path, desa_id, created_at, printed_at, synced, synced_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var printedAt, syncedAt interface{}
	if s.PrintedAt != nil {
		printedAt = s.PrintedAt.Format("2006-01-02 15:04:05")
	}
	if s.SyncedAt != nil {
		syncedAt = s.SyncedAt.Format("2006-01-02 15:04:05")
	}

	dataBytes, err := json.Marshal(s.DataSurat)
	if err != nil {
		return fmt.Errorf("gagal marshal data_surat: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		s.ID, s.NomorSurat, s.JenisSuratID, s.JenisSuratKode, s.JenisSuratNama,
		s.WargaID, s.NIKPemohon, s.NamaPemohon, string(dataBytes), s.Status,
		s.PDFPath, s.DesaID, s.CreatedAt.Format("2006-01-02 15:04:05"), printedAt, s.Synced, syncedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan surat: %w", err)
	}

	return nil
}

// FindByID retrieves a surat by its ID.
func (r *SuratRepository) FindByID(ctx context.Context, id string) (*models.Surat, error) {
	query := `
		SELECT id, nomor_surat, jenis_surat_id, jenis_surat_kode, jenis_surat_nama,
		       warga_id, nik_pemohon, nama_pemohon, data_surat, status,
		       pdf_path, desa_id, created_at, printed_at, synced, synced_at
		FROM surat
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

// ListToday lists all surat created today.
func (r *SuratRepository) ListToday(ctx context.Context, desaID string) ([]models.Surat, error) {
	query := `
		SELECT id, nomor_surat, jenis_surat_id, jenis_surat_kode, jenis_surat_nama,
		       warga_id, nik_pemohon, nama_pemohon, data_surat, status,
		       pdf_path, desa_id, created_at, printed_at, synced, synced_at
		FROM surat
		WHERE desa_id = ? AND date(created_at) = ?
		ORDER BY created_at DESC
	`
	todayStr := time.Now().Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx, query, desaID, todayStr)
	if err != nil {
		return nil, fmt.Errorf("gagal query surat hari ini: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// MarkPrinted updates the status of a surat to PRINTED and sets printed_at.
// It also queues the surat in sync_queue.
func (r *SuratRepository) MarkPrinted(ctx context.Context, id string, pdfPath string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	// 1. Update status in surat table
	updateQuery := `
		UPDATE surat
		SET status = ?, pdf_path = ?, printed_at = ?
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, updateQuery, models.SuratStatusPrinted, pdfPath, now.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("gagal update status printed: %w", err)
	}

	// Retrieve updated surat to queue it
	var s models.Surat
	var dataSuratStr string
	var wargaID sql.NullString
	var nomorSurat sql.NullString
	var printedAtVal sql.NullTime
	var syncedAtVal sql.NullTime
	var createdAtVal time.Time

	selectQuery := `
		SELECT id, nomor_surat, jenis_surat_id, jenis_surat_kode, jenis_surat_nama,
		       warga_id, nik_pemohon, nama_pemohon, data_surat, status,
		       pdf_path, desa_id, created_at, printed_at, synced, synced_at
		FROM surat
		WHERE id = ?
	`
	err = tx.QueryRowContext(ctx, selectQuery, id).Scan(
		&s.ID, &nomorSurat, &s.JenisSuratID, &s.JenisSuratKode, &s.JenisSuratNama,
		&wargaID, &s.NIKPemohon, &s.NamaPemohon, &dataSuratStr, &s.Status,
		&s.PDFPath, &s.DesaID, &createdAtVal, &printedAtVal, &s.Synced, &syncedAtVal,
	)
	if err != nil {
		return fmt.Errorf("gagal query surat setelah update: %w", err)
	}

	s.NomorSurat = nomorSurat.String
	s.WargaID = wargaID.String
	s.CreatedAt = createdAtVal
	s.DataSurat = json.RawMessage(dataSuratStr)
	if printedAtVal.Valid {
		s.PrintedAt = &printedAtVal.Time
	}

	// 2. Queue in sync_queue
	payloadBytes, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("gagal marshal payload sync: %w", err)
	}

	queueQuery := `
		INSERT INTO sync_queue (entity_type, entity_id, operation, payload)
		VALUES ('surat', ?, 'CREATE', ?)
	`
	_, err = tx.ExecContext(ctx, queueQuery, id, string(payloadBytes))
	if err != nil {
		return fmt.Errorf("gagal memasukkan surat ke sync_queue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal commit transaksi print: %w", err)
	}

	return nil
}

// MarkSynced updates the status of a surat to SYNCED and sets synced_at.
func (r *SuratRepository) MarkSynced(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE surat
		SET status = ?, synced = 1, synced_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, models.SuratStatusSynced, now.Format("2006-01-02 15:04:05"), id)
	if err != nil {
		return fmt.Errorf("gagal update status synced: %w", err)
	}
	return nil
}

// ListUnsynced retrieves all surat records that have not been synced yet.
func (r *SuratRepository) ListUnsynced(ctx context.Context, desaID string) ([]models.Surat, error) {
	query := `
		SELECT id, nomor_surat, jenis_surat_id, jenis_surat_kode, jenis_surat_nama,
		       warga_id, nik_pemohon, nama_pemohon, data_surat, status,
		       pdf_path, desa_id, created_at, printed_at, synced, synced_at
		FROM surat
		WHERE desa_id = ? AND synced = 0 AND status = ?
	`
	rows, err := r.db.QueryContext(ctx, query, desaID, models.SuratStatusPrinted)
	if err != nil {
		return nil, fmt.Errorf("gagal query surat belum synced: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Helpers
func (r *SuratRepository) scanRow(row *sql.Row) (*models.Surat, error) {
	var s models.Surat
	var nomorSurat, wargaID, pdfPath sql.NullString
	var dataSuratStr string
	var printedAt, syncedAt sql.NullTime

	err := row.Scan(
		&s.ID, &nomorSurat, &s.JenisSuratID, &s.JenisSuratKode, &s.JenisSuratNama,
		&wargaID, &s.NIKPemohon, &s.NamaPemohon, &dataSuratStr, &s.Status,
		&pdfPath, &s.DesaID, &s.CreatedAt, &printedAt, &s.Synced, &syncedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan surat: %w", err)
	}

	s.NomorSurat = nomorSurat.String
	s.WargaID = wargaID.String
	s.PDFPath = pdfPath.String
	s.DataSurat = json.RawMessage(dataSuratStr)
	if printedAt.Valid {
		s.PrintedAt = &printedAt.Time
	}
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}

func (r *SuratRepository) scanRows(rows *sql.Rows) (*models.Surat, error) {
	var s models.Surat
	var nomorSurat, wargaID, pdfPath sql.NullString
	var dataSuratStr string
	var printedAt, syncedAt sql.NullTime

	err := rows.Scan(
		&s.ID, &nomorSurat, &s.JenisSuratID, &s.JenisSuratKode, &s.JenisSuratNama,
		&wargaID, &s.NIKPemohon, &s.NamaPemohon, &dataSuratStr, &s.Status,
		&pdfPath, &s.DesaID, &s.CreatedAt, &printedAt, &s.Synced, &syncedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("gagal scan surat row: %w", err)
	}

	s.NomorSurat = nomorSurat.String
	s.WargaID = wargaID.String
	s.PDFPath = pdfPath.String
	s.DataSurat = json.RawMessage(dataSuratStr)
	if printedAt.Valid {
		s.PrintedAt = &printedAt.Time
	}
	if syncedAt.Valid {
		s.SyncedAt = &syncedAt.Time
	}

	return &s, nil
}
