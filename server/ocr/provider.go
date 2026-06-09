package ocr

import (
	"context"
)

// KTPData holds the structured resident information extracted from KTP image.
type KTPData struct {
	NIK             string  `json:"nik"`
	Nama            string  `json:"nama"`
	TempatLahir     string  `json:"tempat_lahir"`
	TanggalLahir    string  `json:"tanggal_lahir"` // YYYY-MM-DD
	JenisKelamin    string  `json:"jenis_kelamin"` // L/P
	Alamat          string  `json:"alamat"`
	RT              string  `json:"rt"`
	RW              string  `json:"rw"`
	Kelurahan       string  `json:"kelurahan"`
	Kecamatan       string  `json:"kecamatan"`
	Agama           string  `json:"agama"`
	StatusKawin     string  `json:"status_kawin"`
	Pekerjaan       string  `json:"pekerjaan"`
	Kewarganegaraan string  `json:"kewarganegaraan"`
	Confidence      float64 `json:"confidence"` // 0-1
}

// OCRProvider represents an interface for AI OCR services.
type OCRProvider interface {
	Name() string
	ExtractKTP(ctx context.Context, imageData []byte) (*KTPData, error)
}
