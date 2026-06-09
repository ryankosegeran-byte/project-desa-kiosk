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
	Options     []string   `json:"options,omitempty"`   // for select, radio
	SubFields   []FieldDef `json:"sub_fields,omitempty"` // for repeater
}

// SuratTemplate represents an HTML template for a specific jenis surat per desa.
type SuratTemplate struct {
	ID           string    `json:"id"`
	JenisSuratID string    `json:"jenis_surat_id"`
	DesaID       string    `json:"desa_id"`
	TemplateHTML string    `json:"template_html"`
	Version      int       `json:"version"`
	CreatedBy    string    `json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DesaJenisSurat represents the activation of a jenis surat for a specific desa.
type DesaJenisSurat struct {
	DesaID       string `json:"desa_id"`
	JenisSuratID string `json:"jenis_surat_id"`
	Aktif        bool   `json:"aktif"`
	Urutan       int    `json:"urutan"`
}
