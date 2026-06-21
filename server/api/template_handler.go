package api

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/project-desa-kiosk/internal/docx"
	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListTemplates lists all letter templates for a village (including general templates).
func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := r.URL.Query().Get("desa_id")
	isGeneral := r.URL.Query().Get("is_general") == "true"

	// PIC can only see their own desa templates + general templates
	if claims.Role == models.RolePICDesa {
		desaID = claims.DesaID
	}

	var list []models.SuratTemplate
	var err error

	if isGeneral && claims.Role == models.RoleSuperAdmin {
		// Superadmin can list general templates
		list, err = s.templateRepo.ListGeneralTemplates(ctx)
	} else if desaID != "" {
		// List templates for specific desa (including general)
		list, err = s.templateRepo.ListTemplatesForDesa(ctx, desaID)
	} else {
		sendError(w, http.StatusBadRequest, "desa_id atau is_general=true diperlukan")
		return
	}

	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil daftar template: "+err.Error())
		return
	}

	if len(list) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, list)
}

// handleUpsertTemplate creates or updates an HTML template for a type of letter.
// Supports both general templates (superadmin only) and per-desa templates.
func (s *Server) handleUpsertTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	var req models.SuratTemplate
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.JenisSuratID == "" || req.TemplateHTML == "" {
		sendError(w, http.StatusBadRequest, "JenisSuratID dan TemplateHTML harus diisi")
		return
	}

	// Validate format_kertas
	if req.FormatKertas == "" {
		req.FormatKertas = models.FormatKertasA4
	}
	if req.FormatKertas != models.FormatKertasA4 && req.FormatKertas != models.FormatKertasF4 {
		sendError(w, http.StatusBadRequest, "FormatKertas harus 'A4' atau 'F4'")
		return
	}

	// Handle general template (superadmin only)
	if req.IsGeneral {
		if claims.Role != models.RoleSuperAdmin {
			sendError(w, http.StatusForbidden, "Hanya superadmin yang dapat membuat template umum")
			return
		}
		// General templates don't need desa_id
		req.DesaID = ""
	} else {
		// Per-desa template
		if claims.Role == models.RolePICDesa {
			req.DesaID = claims.DesaID
		} else if req.DesaID == "" {
			sendError(w, http.StatusBadRequest, "DesaID harus diisi untuk template per-desa")
			return
		}
	}

	// Verify jenis surat exists
	_, err := s.jenisSuratRepo.FindByID(ctx, req.JenisSuratID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusBadRequest, "Jenis surat tidak terdaftar")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal verifikasi jenis surat: "+err.Error())
		return
	}

	// Check if this is a new template or overwrite
	var existing *models.SuratTemplate
	if req.IsGeneral {
		// Get existing general template
		existing, err = s.templateRepo.GetGeneralTemplate(ctx, req.JenisSuratID)
	} else {
		existing, err = s.templateRepo.GetTemplate(ctx, req.JenisSuratID, req.DesaID)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			req.ID = uuid.New().String()
			req.Version = 1
		} else {
			sendError(w, http.StatusInternalServerError, "Gagal memeriksa template: "+err.Error())
			return
		}
	} else {
		// Save version history before updating
		historyID := uuid.New().String()
		history := &models.TemplateVersionHistory{
			ID:           historyID,
			TemplateID:   existing.ID,
			Version:      existing.Version,
			TemplateHTML: existing.TemplateHTML,
			FormatKertas: existing.FormatKertas,
			CreatedBy:    existing.CreatedBy,
			CreatedAt:    existing.UpdatedAt,
		}
		_ = s.templateRepo.SaveTemplateVersion(ctx, history)

		req.ID = existing.ID
		req.Version = existing.Version
	}

	req.CreatedBy = claims.UserID

	if err := s.templateRepo.UpsertTemplate(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan template: "+err.Error())
		return
	}

	// Log activity
	entityType := "template_general"
	if !req.IsGeneral {
		entityType = "template_desa"
	}
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "UPSERT_TEMPLATE",
		EntityType: entityType,
		EntityID:   req.ID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, req)
}

// handleGetTemplateVersionHistory returns version history for a template
func (s *Server) handleGetTemplateVersionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	templateID := r.URL.Query().Get("template_id")
	if templateID == "" {
		sendError(w, http.StatusBadRequest, "template_id diperlukan")
		return
	}

	versions, err := s.templateRepo.GetTemplateVersions(ctx, templateID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil history template: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, versions)
}

// handleUploadDocxTemplate accepts a marked-up .docx (Strategi B), stores it as
// the per-desa template, auto-detects {{token}} placeholders, and returns a
// suggested mapping for the admin to refine.
func (s *Server) handleUploadDocxTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		sendError(w, http.StatusBadRequest, "Ukuran file terlalu besar (maksimal 10MB)")
		return
	}

	jenisSuratID := r.FormValue("jenis_surat_id")
	if jenisSuratID == "" {
		sendError(w, http.StatusBadRequest, "jenis_surat_id diperlukan")
		return
	}
	if _, perr := uuid.Parse(jenisSuratID); perr != nil {
		sendError(w, http.StatusBadRequest, "jenis_surat_id tidak valid (harus UUID). Pastikan daftar jenis surat dimuat dari server, bukan data contoh.")
		return
	}
	if r.FormValue("is_general") == "true" {
		sendError(w, http.StatusNotImplemented, "Upload DOCX untuk template umum belum didukung — unggah per desa")
		return
	}

	desaID := r.FormValue("desa_id")
	if claims.Role == models.RolePICDesa {
		desaID = claims.DesaID
	} else if desaID == "" {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan untuk template per-desa")
		return
	}

	file, _, err := r.FormFile("docx")
	if err != nil {
		sendError(w, http.StatusBadRequest, "File docx diperlukan")
		return
	}
	defer file.Close()
	docxBytes, err := io.ReadAll(file)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membaca file: "+err.Error())
		return
	}

	if _, err := s.jenisSuratRepo.FindByID(ctx, jenisSuratID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusBadRequest, "Jenis surat tidak terdaftar")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal verifikasi jenis surat: "+err.Error())
		return
	}

	tokens, err := docx.DetectTokens(docxBytes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Gagal membaca DOCX: "+err.Error())
		return
	}

	// Preserve any existing per-desa mapping for re-uploads.
	var existingPlaceholders []models.PlaceholderDef
	var existingID string
	if tpl, isPerDesa, _ := s.templateRepo.GetTemplateWithFallback(ctx, jenisSuratID, desaID); tpl != nil && isPerDesa {
		existingPlaceholders = tpl.Placeholders
		existingID = tpl.ID
	}

	tpl := &models.SuratTemplate{
		ID:           existingID,
		JenisSuratID: jenisSuratID,
		DesaID:       desaID,
		IsGeneral:    false,
		TemplateHTML: "", // DOCX-only template
		TemplateDocx: docxBytes,
		Placeholders: buildPlaceholders(tokens, existingPlaceholders),
		Version:      1,
		CreatedBy:    claims.UserID,
	}
	if tpl.ID == "" {
		tpl.ID = uuid.New().String()
	}

	if err := s.templateRepo.UpsertTemplate(ctx, tpl); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan template: "+err.Error())
		return
	}

	// PDF pendamping (opsional) — hanya untuk preview tampilan di dashboard.
	if pdfFile, _, perr := r.FormFile("pdf"); perr == nil {
		defer pdfFile.Close()
		if pdfBytes, rerr := io.ReadAll(pdfFile); rerr == nil && len(pdfBytes) > 0 {
			if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
				sendError(w, http.StatusBadRequest, "File tampilan harus berformat PDF")
				return
			}
			if err := s.templateRepo.SetTemplatePDF(ctx, tpl.ID, pdfBytes); err != nil {
				sendError(w, http.StatusInternalServerError, "Gagal menyimpan PDF tampilan: "+err.Error())
				return
			}
		}
	}

	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "UPLOAD_TEMPLATE_DOCX",
		EntityType: "template_desa",
		EntityID:   tpl.ID,
		IPAddress:  r.RemoteAddr,
	})

	tpl.TemplateDocx = nil // do not echo the blob back
	sendJSON(w, http.StatusOK, map[string]interface{}{
		"template": tpl,
		"tokens":   tokens,
	})
}

// handleUpdateTemplatePlaceholders saves the admin's refined placeholder mapping.
func (s *Server) handleUpdateTemplatePlaceholders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	templateID := chi.URLParam(r, "id")
	if templateID == "" {
		sendError(w, http.StatusBadRequest, "id template diperlukan")
		return
	}

	var placeholders []models.PlaceholderDef
	if err := parseJSON(r, &placeholders); err != nil {
		sendError(w, http.StatusBadRequest, "Payload tidak valid: "+err.Error())
		return
	}

	if err := s.templateRepo.UpdatePlaceholders(ctx, templateID, placeholders); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan pemetaan: "+err.Error())
		return
	}

	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "UPDATE_TEMPLATE_PLACEHOLDERS",
		EntityType: "template_desa",
		EntityID:   templateID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// wargaKeyMap / sistemKeyMap drive smart suggestions: a detected token whose name
// matches a known key is pre-assigned to a KTP (warga) or system field; the admin
// always confirms. Unknown tokens default to a manual text field.
var wargaKeyMap = map[string]string{
	"nama":            "Nama",
	"nik":             "NIK",
	"tempat_lahir":    "TempatLahir",
	"tanggal_lahir":   "TanggalLahir",
	"jenis_kelamin":   "JenisKelamin",
	"alamat":          "Alamat",
	"rt":              "RT",
	"rw":              "RW",
	"kelurahan":       "Kelurahan",
	"kecamatan":       "Kecamatan",
	"kabupaten":       "Kabupaten",
	"provinsi":        "Provinsi",
	"agama":           "Agama",
	"status":          "StatusKawin",
	"status_kawin":    "StatusKawin",
	"pekerjaan":       "Pekerjaan",
	"kewarganegaraan": "Kewarganegaraan",
}

var sistemKeyMap = map[string]string{
	"nomor_surat":     "NomorSurat",
	"nomor":           "NomorSurat",
	"tanggal":         "DateToday",
	"tanggal_surat":   "DateToday",
	"kepala_desa":     "DesaKepalaDesa",
	"hukum_tua":       "DesaKepalaDesa",
	"nip_kepala_desa": "DesaNIP",
	"nip":             "DesaNIP",
}

// buildPlaceholders maps detected tokens to PlaceholderDefs, reusing any existing
// mapping for the same key and otherwise applying a sensible suggestion.
func buildPlaceholders(tokens []string, existing []models.PlaceholderDef) []models.PlaceholderDef {
	prev := make(map[string]models.PlaceholderDef, len(existing))
	for _, p := range existing {
		prev[p.Key] = p
	}
	out := make([]models.PlaceholderDef, 0, len(tokens))
	for i, key := range tokens {
		if p, ok := prev[key]; ok {
			p.Urutan = i
			out = append(out, p)
			continue
		}
		out = append(out, suggestPlaceholder(key, i))
	}
	return out
}

// suggestPlaceholder picks a default source for a token from its name.
func suggestPlaceholder(key string, urutan int) models.PlaceholderDef {
	p := models.PlaceholderDef{Key: key, Label: labelize(key), Urutan: urutan}
	if wf, ok := wargaKeyMap[key]; ok {
		p.Source = models.PlaceholderSourceWarga
		p.WargaField = wf
		return p
	}
	if sf, ok := sistemKeyMap[key]; ok {
		p.Source = models.PlaceholderSourceSistem
		p.SistemField = sf
		return p
	}
	p.Source = models.PlaceholderSourceManual
	p.Type = "text"
	p.Required = true
	return p
}

// labelize turns "jenis_usaha" into "Jenis Usaha".
func labelize(key string) string {
	parts := strings.Split(key, "_")
	for i, wd := range parts {
		if wd == "" {
			continue
		}
		parts[i] = strings.ToUpper(wd[:1]) + wd[1:]
	}
	return strings.Join(parts, " ")
}

// handlePreviewTemplatePost fills a DOCX template with caller-supplied dummy
// values (merged on top of auto-generated defaults) and returns the filled DOCX
// as an attachment so the browser triggers a download → Word opens automatically.
func (s *Server) handlePreviewTemplatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	templateID := chi.URLParam(r, "id")
	if templateID == "" {
		sendError(w, http.StatusBadRequest, "id template diperlukan")
		return
	}

	var req struct {
		DummyValues map[string]string `json:"dummy_values"`
	}
	_ = parseJSON(r, &req) // ignore error — custom values are optional

	tpl, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil template: "+err.Error())
		return
	}

	if len(tpl.TemplateDocx) == 0 {
		sendError(w, http.StatusBadRequest, "Template ini tidak memiliki konten DOCX")
		return
	}

	// Start from auto-generated defaults then override with caller-supplied values.
	values := buildDummyValues(tpl.Placeholders)
	for k, v := range req.DummyValues {
		values[k] = v
	}

	filled, err := docx.FillDocx(tpl.TemplateDocx, values)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengisi template: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", `attachment; filename="preview_surat.docx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(filled)
}

// handleGetTemplatePDF menyajikan PDF pendamping (tampilan) sebuah template.
// 404 bila belum ada — frontend menampilkan "belum ada tampilan".
func (s *Server) handleGetTemplatePDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if middleware.GetClaims(ctx) == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}
	templateID := chi.URLParam(r, "id")
	if templateID == "" {
		sendError(w, http.StatusBadRequest, "id template diperlukan")
		return
	}
	pdf, err := s.templateRepo.GetTemplatePDF(ctx, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil PDF: "+err.Error())
		return
	}
	if len(pdf) == 0 {
		sendError(w, http.StatusNotFound, "Belum ada tampilan PDF untuk template ini")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="tampilan.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

// handleUploadTemplatePDF mengunggah/mengganti PDF pendamping untuk template yang sudah ada.
func (s *Server) handleUploadTemplatePDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}
	templateID := chi.URLParam(r, "id")
	if templateID == "" {
		sendError(w, http.StatusBadRequest, "id template diperlukan")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		sendError(w, http.StatusBadRequest, "Ukuran file terlalu besar (maksimal 10MB)")
		return
	}
	file, _, err := r.FormFile("pdf")
	if err != nil {
		sendError(w, http.StatusBadRequest, "File PDF diperlukan")
		return
	}
	defer file.Close()
	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membaca file: "+err.Error())
		return
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		sendError(w, http.StatusBadRequest, "File tampilan harus berformat PDF")
		return
	}
	if err := s.templateRepo.SetTemplatePDF(ctx, templateID, pdfBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan PDF tampilan: "+err.Error())
		return
	}

	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "UPLOAD_TEMPLATE_PDF",
		EntityType: "template_desa",
		EntityID:   templateID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// buildDummyValues generates sample values for each placeholder in the template,
// and also detects any remaining tokens in the DOCX directly to fill them with
// sensible defaults. This ensures tokens work even if placeholder mapping is
// incomplete or uses different naming conventions.
func buildDummyValues(placeholders []models.PlaceholderDef) map[string]string {
	values := make(map[string]string, len(placeholders)+20)

	// Comprehensive token→value lookup. Keys use both dash and underscore forms.
	// This covers common Indonesian administrative letter fields.
	knownDefaults := map[string]string{
		// Warga identity
		"nama":              "BUDI SANTOSO",
		"nama-lengkap":      "BUDI SANTOSO",
		"nama_lengkap":      "BUDI SANTOSO",
		"nik":               "3201234567890001",
		"tempat-lahir":      "Manado",
		"tempat_lahir":      "Manado",
		"tanggal-lahir":     "17 Juni 1990",
		"tanggal_lahir":     "17 Juni 1990",
		"tempat-tanggal-lahir": "Manado, 17 Juni 1990",
		"tempat_tanggal_lahir": "Manado, 17 Juni 1990",
		"jenis-kelamin":     "Laki-laki",
		"jenis_kelamin":     "Laki-laki",
		"alamat":            "Jl. Sudirman No. 10",
		"rt":                "001",
		"rw":                "002",
		"kelurahan":         "Kalawat",
		"kecamatan":         "Kalawat",
		"kabupaten":         "Minahasa Utara",
		"provinsi":          "Sulawesi Utara",
		"agama":             "Kristen",
		"status":            "Kawin",
		"status-kawin":      "Kawin",
		"status_kawin":      "Kawin",
		"pekerjaan":         "Wiraswasta",
		"kewarganegaraan":   "Indonesia",

		// System / letter
		"nomor-surat":       "001/SKU/08.10/VI/2026",
		"nomor_surat":       "001/SKU/08.10/VI/2026",
		"tanggal-surat":     formatTanggal(time.Now()),
		"tanggal_surat":     formatTanggal(time.Now()),
		"tanggal":           formatTanggal(time.Now()),
		"kepala-desa":       "ALFRIDA, A.Md.Kes",
		"kepala_desa":       "ALFRIDA, A.Md.Kes",
		"nip":               "196902061993032004",
		"desa":              "Kalawat",
		"nama-desa":         "Kalawat",
		"nama_desa":         "Kalawat",

		// SKU-specific
		"jenis-usaha":       "Toko Kelontong",
		"jenis_usaha":       "Toko Kelontong",
		"merk-usaha":        "Toko Budi Makmur",
		"merk_usaha":        "Toko Budi Makmur",
		"nama-usaha":        "Toko Budi Makmur",
		"nama_usaha":        "Toko Budi Makmur",
		"alamat-usaha":      "Jl. Desa Kalawat No. 5",
		"alamat_usaha":      "Jl. Desa Kalawat No. 5",
		"batas-utara":       "Toko Sari",
		"batas_utara":       "Toko Sari",
		"batas-selatan":     "Jalan Desa",
		"batas_selatan":     "Jalan Desa",
		"batas-timur":       "Sungai Kecil",
		"batas_timur":       "Sungai Kecil",
		"batas-barat":       "Tanah Kosong",
		"batas_barat":       "Tanah Kosong",
		"sifat-tempat-usaha": "Permanen",
		"sifat_tempat_usaha": "Permanen",
		"tahun-kewajiban":   "2026",
		"tahun_kewajiban":   "2026",
		"tahun-mulai-usaha": "2020",
		"tahun_mulai_usaha": "2020",
		"tahun":             "2026",

		// SKTM / poverty
		"penghasilan":       "Rp 1.500.000",
		"hubungan":          "Anak Kandung",
		"nama-orang-tua":    "SUHARTONO",
		"nama_orang_tua":    "SUHARTONO",

		// Generic
		"keterangan":        "Digunakan untuk keperluan administrasi.",
		"keperluan":         "Administrasi",
		"nomor":             "001/SKU/08.10/VI/2026",
	}

	// Map from placeholder definitions first (these take priority).
	wargaFieldMap := map[string]string{
		"Nama": "BUDI SANTOSO", "NIK": "3201234567890001",
		"TempatLahir": "Manado", "TanggalLahir": "17 Juni 1990",
		"JenisKelamin": "Laki-laki", "Alamat": "Jl. Sudirman No. 10",
		"RT": "001", "RW": "002", "Kelurahan": "Kalawat", "Kecamatan": "Kalawat",
		"Kabupaten": "Minahasa Utara", "Provinsi": "Sulawesi Utara",
		"Agama": "Kristen", "StatusKawin": "Kawin", "Pekerjaan": "Wiraswasta",
		"Kewarganegaraan": "Indonesia",
	}
	systemFieldMap := map[string]string{
		"NomorSurat": "001/SKU/08.10/VI/2026", "DateToday": formatTanggal(time.Now()),
		"DesaKepalaDesa": "ALFRIDA, A.Md.Kes", "DesaNIP": "196902061993032004",
	}

	for _, p := range placeholders {
		switch p.Source {
		case models.PlaceholderSourceWarga:
			if v, ok := wargaFieldMap[p.WargaField]; ok {
				values[p.Key] = v
			} else {
				values[p.Key] = fmt.Sprintf("[%s]", p.Label)
			}
		case models.PlaceholderSourceSistem:
			if v, ok := systemFieldMap[p.SistemField]; ok {
				values[p.Key] = v
			} else {
				values[p.Key] = fmt.Sprintf("[%s]", p.Label)
			}
		case models.PlaceholderSourceManual:
			if v, ok := knownDefaults[p.Key]; ok {
				values[p.Key] = v
			} else {
				switch p.Type {
				case "number":
					values[p.Key] = "2026"
				case "date":
					values[p.Key] = formatTanggal(time.Now())
				case "select":
					if len(p.Options) > 0 {
						values[p.Key] = p.Options[0]
					} else {
						values[p.Key] = fmt.Sprintf("[%s]", p.Label)
					}
				default:
					values[p.Key] = fmt.Sprintf("[%s]", p.Label)
				}
			}
		default:
			if v, ok := knownDefaults[p.Key]; ok {
				values[p.Key] = v
			} else {
				values[p.Key] = fmt.Sprintf("[%s]", p.Key)
			}
		}
	}

	// Also add all known defaults directly — this ensures tokens that exist in
	// the DOCX but aren't mapped in placeholders still get replaced.
	for k, v := range knownDefaults {
		if _, exists := values[k]; !exists {
			values[k] = v
		}
	}

	return values
}

// formatTanggal formats a time.Time as "17 Juni 2026" (Indonesian date format).
func formatTanggal(t time.Time) string {
	bulan := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%d %s %d", t.Day(), bulan[t.Month()], t.Year())
}

