package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type NomorSuratRepository struct {
	db *DB
}

func NewNomorSuratRepository(db *DB) *NomorSuratRepository {
	return &NomorSuratRepository{db: db}
}

// GetNextNumber mengambil nomor berikutnya, increment di database lokal, dan return formatted string.
func (r *NomorSuratRepository) GetNextNumber(ctx context.Context, jenisSuratID string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var batch models.NomorSuratBatch
	query := `SELECT nomor_terakhir, batas_atas, COALESCE(format_nomor, '') FROM nomor_surat_batch WHERE jenis_surat_id = ?`

	err = tx.QueryRowContext(ctx, query, jenisSuratID).Scan(&batch.NomorTerakhir, &batch.BatasAtas, &batch.FormatNomor)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("penomoran surat belum dikonfigurasi atau belum disinkronisasi dari server. Hubungi Sekdes untuk mengatur penomoran")
	} else if err != nil {
		return 0, err
	}

	if batch.NomorTerakhir >= batch.BatasAtas {
		return 0, fmt.Errorf("penomoran surat sudah penuh (terpakai %d/%d). Silahkan hubungi Sekdes untuk mengatur penomoran", batch.NomorTerakhir, batch.BatasAtas)
	}

	newNumber := batch.NomorTerakhir + 1
	updateQuery := `UPDATE nomor_surat_batch SET nomor_terakhir = ?, updated_at = ? WHERE jenis_surat_id = ?`
	_, err = tx.ExecContext(ctx, updateQuery, newNumber, time.Now(), jenisSuratID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newNumber, nil
}

// FormatNomorSurat memformat nomor surat berdasarkan pattern yang tersimpan.
// Pattern contoh: "{nomor}/{kode_surat}/{kode_desa}/{bulan_romawi}/{tahun}"
func (r *NomorSuratRepository) FormatNomorSurat(ctx context.Context, jenisSuratID string, nomor int, kodeSurat string, kodeDesa string) (string, error) {
	var formatNomor string
	query := `SELECT COALESCE(format_nomor, '') FROM nomor_surat_batch WHERE jenis_surat_id = ?`
	err := r.db.QueryRowContext(ctx, query, jenisSuratID).Scan(&formatNomor)
	if err != nil || formatNomor == "" {
		// Fallback: plain number
		return fmt.Sprintf("%d", nomor), nil
	}

	now := time.Now()
	bulanRomawi := toRomanMonth(int(now.Month()))

	result := formatNomor
	result = strings.ReplaceAll(result, "{nomor}", fmt.Sprintf("%d", nomor))
	result = strings.ReplaceAll(result, "{kode_surat}", kodeSurat)
	result = strings.ReplaceAll(result, "{kode_desa}", kodeDesa)
	result = strings.ReplaceAll(result, "{bulan_romawi}", bulanRomawi)
	result = strings.ReplaceAll(result, "{tahun}", fmt.Sprintf("%d", now.Year()))
	result = strings.ReplaceAll(result, "{bulan}", fmt.Sprintf("%02d", int(now.Month())))

	return result, nil
}

// GetBatch returns the current batch info for monitoring.
func (r *NomorSuratRepository) GetBatch(ctx context.Context, jenisSuratID string) (*models.NomorSuratBatch, error) {
	query := `SELECT jenis_surat_id, nomor_terakhir, batas_atas, COALESCE(format_nomor, ''), updated_at FROM nomor_surat_batch WHERE jenis_surat_id = ?`
	var batch models.NomorSuratBatch
	err := r.db.QueryRowContext(ctx, query, jenisSuratID).Scan(
		&batch.JenisSuratID, &batch.NomorTerakhir, &batch.BatasAtas, &batch.FormatNomor, &batch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// ListAllBatches returns all batch statuses.
func (r *NomorSuratRepository) ListAllBatches(ctx context.Context) ([]models.NomorSuratBatch, error) {
	query := `SELECT jenis_surat_id, nomor_terakhir, batas_atas, COALESCE(format_nomor, ''), updated_at FROM nomor_surat_batch`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.NomorSuratBatch
	for rows.Next() {
		var b models.NomorSuratBatch
		if err := rows.Scan(&b.JenisSuratID, &b.NomorTerakhir, &b.BatasAtas, &b.FormatNomor, &b.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

// UpdateBatch digunakan untuk menyinkronkan jatah nomor dari server.
func (r *NomorSuratRepository) UpdateBatch(ctx context.Context, batch models.NomorSuratBatch) error {
	query := `
		INSERT INTO nomor_surat_batch (jenis_surat_id, nomor_terakhir, batas_atas, format_nomor, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(jenis_surat_id) DO UPDATE SET
		nomor_terakhir = excluded.nomor_terakhir,
		batas_atas = excluded.batas_atas,
		format_nomor = excluded.format_nomor,
		updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, batch.JenisSuratID, batch.NomorTerakhir, batch.BatasAtas, batch.FormatNomor, time.Now())
	return err
}

// toRomanMonth converts month number to Roman numeral.
func toRomanMonth(month int) string {
	romans := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	if month >= 1 && month <= 12 {
		return romans[month]
	}
	return fmt.Sprintf("%d", month)
}
