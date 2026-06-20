package print

import (
	"fmt"

	"github.com/project-desa-kiosk/internal/models"
)

// SistemValues holds the auto-filled, system-sourced values for a letter
// (assigned by the kiosk, not typed by the resident).
type SistemValues struct {
	NomorSurat     string
	DateToday      string
	DesaKepalaDesa string
	DesaNIP        string
}

// ResolveValues maps every placeholder token key to its final string value,
// based on the placeholder's source (warga / manual / sistem). Every placeholder
// gets an entry (empty string when unknown) so no {{token}} is left unrendered.
func ResolveValues(placeholders []models.PlaceholderDef, warga *models.Warga, dataSurat map[string]interface{}, sys SistemValues) map[string]string {
	values := make(map[string]string, len(placeholders))
	for _, p := range placeholders {
		switch p.Source {
		case models.PlaceholderSourceWarga:
			values[p.Key] = wargaFieldValue(warga, p.WargaField)
		case models.PlaceholderSourceSistem:
			values[p.Key] = sistemFieldValue(sys, p.SistemField)
		case models.PlaceholderSourceManual:
			values[p.Key] = dataSuratValue(dataSurat, p.Key)
		default:
			values[p.Key] = ""
		}
	}
	return values
}

func dataSuratValue(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	v, ok := data[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func sistemFieldValue(sys SistemValues, field string) string {
	switch field {
	case "NomorSurat":
		return sys.NomorSurat
	case "DateToday":
		return sys.DateToday
	case "DesaKepalaDesa":
		return sys.DesaKepalaDesa
	case "DesaNIP":
		return sys.DesaNIP
	default:
		return ""
	}
}

func wargaFieldValue(w *models.Warga, field string) string {
	if w == nil {
		return ""
	}
	switch field {
	case "Nama":
		return w.Nama
	case "NIK":
		return w.NIK
	case "TempatLahir":
		return w.TempatLahir
	case "TanggalLahir":
		return w.TanggalLahir
	case "JenisKelamin":
		return expandJenisKelamin(w.JenisKelamin)
	case "Alamat":
		return w.Alamat
	case "RT":
		return w.RT
	case "RW":
		return w.RW
	case "Kelurahan":
		return w.Kelurahan
	case "Kecamatan":
		return w.Kecamatan
	case "Kabupaten":
		return w.Kabupaten
	case "Provinsi":
		return w.Provinsi
	case "Agama":
		return w.Agama
	case "StatusKawin":
		return w.StatusKawin
	case "Pekerjaan":
		return w.Pekerjaan
	case "Kewarganegaraan":
		return w.Kewarganegaraan
	default:
		return ""
	}
}

func expandJenisKelamin(jk string) string {
	switch jk {
	case "L":
		return "Laki-laki"
	case "P":
		return "Perempuan"
	default:
		return jk
	}
}
