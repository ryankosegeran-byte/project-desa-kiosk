package models

import (
	"encoding/json"
	"time"
)

// Surat represents a letter/document created at the kiosk.
type Surat struct {
	ID             string          `json:"id"`
	NomorSurat     string          `json:"nomor_surat,omitempty"`
	JenisSuratID   string          `json:"jenis_surat_id"`
	JenisSuratKode string          `json:"jenis_surat_kode"`
	JenisSuratNama string          `json:"jenis_surat_nama"`
	WargaID        string          `json:"warga_id,omitempty"`
	NIKPemohon     string          `json:"nik_pemohon"`
	NamaPemohon    string          `json:"nama_pemohon"`
	DataSurat      json.RawMessage `json:"data_surat"`
	Status         string          `json:"status"` // DRAFT, PRINTED, SYNCED, FAILED
	PDFPath        string          `json:"pdf_path,omitempty"`
	DesaID         string          `json:"desa_id"`
	CreatedAt      time.Time       `json:"created_at"`
	PrintedAt      *time.Time      `json:"printed_at,omitempty"`
	Synced         bool            `json:"synced"`
	SyncedAt       *time.Time      `json:"synced_at,omitempty"`
}

// NomorSuratBatch represents a batch of assigned surat numbers.
type NomorSuratBatch struct {
	JenisSuratID  string    `json:"jenis_surat_id"`
	NomorTerakhir int       `json:"nomor_terakhir"`
	BatasAtas     int       `json:"batas_atas"`
	FormatNomor   string    `json:"format_nomor"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SuratStatus constants.
const (
	SuratStatusDraft   = "DRAFT"
	SuratStatusPrinted = "PRINTED"
	SuratStatusSynced  = "SYNCED"
	SuratStatusFailed  = "FAILED"
)

// CreateSuratRequest is the payload for creating a new surat at the kiosk.
type CreateSuratRequest struct {
	JenisSuratID string          `json:"jenis_surat_id"`
	WargaID      string          `json:"warga_id,omitempty"`
	NIKPemohon   string          `json:"nik_pemohon"`
	NamaPemohon  string          `json:"nama_pemohon"`
	DataSurat    json.RawMessage `json:"data_surat"`
}
