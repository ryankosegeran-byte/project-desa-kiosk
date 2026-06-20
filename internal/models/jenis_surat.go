package models

import (
	"encoding/json"
	"time"
)

// JenisSurat represents a type of letter that can be issued.
type JenisSurat struct {
	ID           string          `json:"id"`
	Kode         string          `json:"kode"`
	Nama         string          `json:"nama"`
	Deskripsi    string          `json:"deskripsi,omitempty"`
	FieldsSchema json.RawMessage `json:"fields_schema"`
	Aktif        bool            `json:"aktif"`
	Urutan       int             `json:"urutan"`
	CreatedAt    time.Time       `json:"created_at,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// FieldsSchema defines the dynamic form fields for a jenis surat.
type FieldsSchema struct {
	Fields []FieldDef `json:"fields"`
}

// FieldDef defines a single form field.
type FieldDef struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Type        string     `json:"type"` // text, textarea, number, date, select, radio, checkbox, repeater
	Required    bool       `json:"required"`
	Placeholder string     `json:"placeholder,omitempty"`
	Options     []string   `json:"options,omitempty"`    // for select, radio
	SubFields   []FieldDef `json:"sub_fields,omitempty"` // for repeater
}

// SuratTemplate represents an HTML template for a specific jenis surat per desa or general.
type SuratTemplate struct {
	ID           string           `json:"id"`
	JenisSuratID string           `json:"jenis_surat_id"`
	DesaID       string           `json:"desa_id"`
	TemplateHTML string           `json:"template_html"`
	TemplateDocx []byte           `json:"template_docx,omitempty"` // DOCX master (Strategi B); nil untuk template HTML lama
	Placeholders []PlaceholderDef `json:"placeholders,omitempty"`  // pemetaan token {{...}} -> sumber nilai, per-template
	IsGeneral    bool             `json:"is_general"`              // true = template umum untuk semua desa
	FormatKertas string           `json:"format_kertas"`           // "A4" atau "F4"
	Version      int              `json:"version"`
	CreatedBy    string           `json:"created_by,omitempty"`
	CreatedAt    time.Time        `json:"created_at,omitempty"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// FormatKertas constants
const (
	FormatKertasA4 = "A4"
	FormatKertasF4 = "F4"
)

// PlaceholderDef mendefinisikan satu token {{...}} di template DOCX dan dari mana
// nilainya diambil saat surat dibuat. Disimpan per-template karena tiap desa
// punya kumpulan token sendiri.
type PlaceholderDef struct {
	Key         string   `json:"key"`                    // nama token tanpa kurung, mis. "jenis_usaha"
	Label       string   `json:"label"`                  // label yang ditampilkan di form kiosk
	Source      string   `json:"source"`                 // PlaceholderSource*: warga | manual | sistem
	WargaField  string   `json:"warga_field,omitempty"`  // jika Source=warga: nama field Warga, mis. "Nama"
	SistemField string   `json:"sistem_field,omitempty"` // jika Source=sistem: NomorSurat|DateToday|DesaKepalaDesa|DesaNIP
	Type        string   `json:"type,omitempty"`         // jika Source=manual: text|textarea|select|date|number
	Options     []string `json:"options,omitempty"`      // jika Type=select
	Required    bool     `json:"required,omitempty"`
	Urutan      int      `json:"urutan,omitempty"` // urutan tampil di form kiosk
}

// Nilai untuk PlaceholderDef.Source.
const (
	PlaceholderSourceWarga  = "warga"
	PlaceholderSourceManual = "manual"
	PlaceholderSourceSistem = "sistem"
)

// TemplateVersionHistory represents version history of a template.
type TemplateVersionHistory struct {
	ID           string    `json:"id"`
	TemplateID   string    `json:"template_id"`
	Version      int       `json:"version"`
	TemplateHTML string    `json:"template_html"`
	FormatKertas string    `json:"format_kertas"`
	ChangeNote   string    `json:"change_note,omitempty"`
	CreatedBy    string    `json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// DesaJenisSurat represents the activation of a jenis surat for a specific desa.
type DesaJenisSurat struct {
	DesaID       string `json:"desa_id"`
	JenisSuratID string `json:"jenis_surat_id"`
	Aktif        bool   `json:"aktif"`
	Urutan       int    `json:"urutan"`
}
