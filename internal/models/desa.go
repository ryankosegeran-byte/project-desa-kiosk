package models

import "time"

// Desa represents a village entity.
type Desa struct {
	ID            string    `json:"id"`
	Nama          string    `json:"nama"`
	KodeDesa      string    `json:"kode_desa"`
	Kecamatan     string    `json:"kecamatan,omitempty"`
	Kabupaten     string    `json:"kabupaten,omitempty"`
	Provinsi      string    `json:"provinsi,omitempty"`
	KepalaDesa    string    `json:"kepala_desa,omitempty"`
	NIPKepalaDesa string    `json:"nip_kepala_desa,omitempty"`
	AlamatKantor  string    `json:"alamat_kantor,omitempty"`
	LogoPath      string    `json:"logo_path,omitempty"`
	Theme         string    `json:"theme,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
