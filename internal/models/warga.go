package models

import "time"

// Warga represents a resident registered in the system.
type Warga struct {
	ID              string     `json:"id"`
	NIK             string     `json:"nik"`
	RFIDUID         string     `json:"rfid_uid,omitempty"`
	Nama            string     `json:"nama"`
	TempatLahir     string     `json:"tempat_lahir,omitempty"`
	TanggalLahir    string     `json:"tanggal_lahir,omitempty"` // YYYY-MM-DD
	JenisKelamin    string     `json:"jenis_kelamin,omitempty"` // L or P
	Alamat          string     `json:"alamat,omitempty"`
	RT              string     `json:"rt,omitempty"`
	RW              string     `json:"rw,omitempty"`
	Kelurahan       string     `json:"kelurahan,omitempty"`
	Kecamatan       string     `json:"kecamatan,omitempty"`
	Kabupaten       string     `json:"kabupaten,omitempty"`
	Provinsi        string     `json:"provinsi,omitempty"`
	Agama           string     `json:"agama,omitempty"`
	StatusKawin     string     `json:"status_kawin,omitempty"`
	Pekerjaan       string     `json:"pekerjaan,omitempty"`
	Kewarganegaraan string     `json:"kewarganegaraan,omitempty"`
	FotoKTPPath     string     `json:"foto_ktp_path,omitempty"`
	DesaID          string     `json:"desa_id"`
	RegisteredBy    string     `json:"registered_by,omitempty"`
	Status          string     `json:"status,omitempty"`      // "draft" atau "complete"
	DraftToken      string     `json:"draft_token,omitempty"` // UUID untuk link draft
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	SyncedAt        *time.Time `json:"synced_at,omitempty"` // nullable: kiosk local field
}

// KTPData represents OCR-extracted data from a KTP photo.
type KTPData struct {
	NIK             string      `json:"nik"`
	Nama            string      `json:"nama"`
	TempatLahir     string      `json:"tempat_lahir"`
	TanggalLahir    string      `json:"tanggal_lahir"`
	JenisKelamin    string      `json:"jenis_kelamin"`
	Alamat          string      `json:"alamat"`
	RT              string      `json:"rt"`
	RW              string      `json:"rw"`
	Kelurahan       string      `json:"kelurahan"`
	Kecamatan       string      `json:"kecamatan"`
	Agama           string      `json:"agama"`
	StatusKawin     string      `json:"status_kawin"`
	Pekerjaan       string      `json:"pekerjaan"`
	Kewarganegaraan string      `json:"kewarganegaraan"`
	Confidence      interface{} `json:"confidence"`
}
