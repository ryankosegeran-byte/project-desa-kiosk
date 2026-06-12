package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/kiosk/print"
)

// handleListJenisSurat returns the list of active jenis surat for this desa.
func (s *Server) handleListJenisSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.jenisSuratRepo.ListAktif(ctx)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil daftar jenis surat: "+err.Error())
		return
	}

	if len(list) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, list)
}

// handleGetJenisSuratSchema returns the schema for a specific jenis surat.
func (s *Server) handleGetJenisSuratSchema(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID jenis surat tidak boleh kosong")
		return
	}

	js, err := s.jenisSuratRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Jenis surat tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil schema: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, js.FieldsSchema)
}

// handleCreateSurat handles creating a new surat.
func (s *Server) handleCreateSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateSuratRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.JenisSuratID == "" {
		sendError(w, http.StatusBadRequest, "ID Jenis Surat harus diisi")
		return
	}
	if req.NIKPemohon == "" {
		sendError(w, http.StatusBadRequest, "NIK Pemohon harus diisi")
		return
	}
	if req.NamaPemohon == "" {
		sendError(w, http.StatusBadRequest, "Nama Pemohon harus diisi")
		return
	}

	// 1. Get JenisSurat detail to denormalize
	js, err := s.jenisSuratRepo.FindByID(ctx, req.JenisSuratID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusBadRequest, "Jenis surat tidak terdaftar")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal memverifikasi jenis surat: "+err.Error())
		return
	}

	// 2. Prepare Surat model
	suratID := uuid.New().String()
	surat := &models.Surat{
		ID:             suratID,
		JenisSuratID:   js.ID,
		JenisSuratKode: js.Kode,
		JenisSuratNama: js.Nama,
		WargaID:        req.WargaID,
		NIKPemohon:     req.NIKPemohon,
		NamaPemohon:    req.NamaPemohon,
		DataSurat:      req.DataSurat,
		Status:         models.SuratStatusDraft,
		DesaID:         s.cfg.DesaID,
		CreatedAt:      time.Now(),
		Synced:         false,
	}

	// 3. Save to database
	if err := s.suratRepo.Create(ctx, surat); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat surat: "+err.Error())
		return
	}

	log.Info().Str("surat_id", surat.ID).Str("jenis", js.Kode).Msg("Surat baru berhasil dibuat")
	sendJSON(w, http.StatusCreated, surat)
}

// handleGetSurat returns details of a single surat.
func (s *Server) handleGetSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID surat tidak boleh kosong")
		return
	}

	surat, err := s.suratRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Surat tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data surat: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, surat)
}

// handleListTodaySurat lists today's letters.
func (s *Server) handleListTodaySurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.suratRepo.ListToday(ctx, s.cfg.DesaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data surat hari ini: "+err.Error())
		return
	}

	if len(list) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, list)
}

// handlePrintSurat handles printing the generated letter.
func (s *Server) handlePrintSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID surat tidak boleh kosong")
		return
	}

	// 1. Verify that the surat exists
	surat, err := s.suratRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Surat tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data surat: "+err.Error())
		return
	}

	// [Offline-First] Generate and assign Nomor Surat
	nomorInt, err := s.nomorSuratRepo.GetNextNumber(ctx, surat.JenisSuratID)
	if err != nil {
		sendError(w, http.StatusPreconditionFailed, err.Error())
		return
	}

	// Format nomor surat menggunakan pattern dari config
	kodeDesa, _ := s.configRepo.Get(ctx, "kode_desa")
	nomorStr, err := s.nomorSuratRepo.FormatNomorSurat(ctx, surat.JenisSuratID, nomorInt, surat.JenisSuratKode, kodeDesa)
	if err != nil {
		nomorStr = fmt.Sprintf("%d", nomorInt) // fallback ke plain number
	}

	if err := s.suratRepo.UpdateNomorSurat(ctx, id, nomorStr); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal update nomor surat: "+err.Error())
		return
	}
	surat.NomorSurat = nomorStr

	// 2. Fetch HTML template for this jenis_surat
	tplObj, err := s.jenisSuratRepo.GetTemplate(ctx, surat.JenisSuratID, s.cfg.DesaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template surat tidak ditemukan untuk desa ini")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil template surat: "+err.Error())
		return
	}

	// 3. Get warga detail (fallback if not found)
	warga, err := s.wargaRepo.FindByNIK(ctx, surat.NIKPemohon)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Fallback: create partial Warga struct for template rendering if not registered
			warga = &models.Warga{
				NIK:  surat.NIKPemohon,
				Nama: surat.NamaPemohon,
			}
		} else {
			sendError(w, http.StatusInternalServerError, "Gagal mengambil data warga: "+err.Error())
			return
		}
	}

	// 4. Parse dynamic fields data_surat
	var dataSurat map[string]interface{}
	if len(surat.DataSurat) > 0 {
		if err := json.Unmarshal(surat.DataSurat, &dataSurat); err != nil {
			sendError(w, http.StatusInternalServerError, "Gagal unmarshal data_surat: "+err.Error())
			return
		}
	}

	// 5. Format current date in Indonesian format
	dateToday := print.FormatIndonesianDate(time.Now())

	// 6. Generate PDF via chromedp
	// Ambil info desa untuk template
	desaKepalaDesa, _ := s.configRepo.Get(ctx, "desa_kepala_desa")
	desaNIP, _ := s.configRepo.Get(ctx, "desa_nip")

	pdfPath, err := s.pdfGen.GeneratePDF(ctx, tplObj.TemplateHTML, warga, dataSurat, dateToday, surat.NomorSurat, desaKepalaDesa, desaNIP)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal generate PDF: "+err.Error())
		return
	}

	// 7. Silent print via SumatraPDF
	if err := s.printer.PrintPDF(pdfPath); err != nil {
		log.Error().Err(err).Str("pdf_path", pdfPath).Msg("Gagal melakukan printing fisik (SumatraPDF)")
		// Note: we still proceed to record it as printed to prevent stopping user workflow,
		// but we logged the printing command error.
	}

	// 8. Mark as printed in SQLite (also creates sync queue item)
	if err := s.suratRepo.MarkPrinted(ctx, id, pdfPath); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal meng-update status surat: "+err.Error())
		return
	}

	log.Info().Str("surat_id", id).Str("pdf_path", pdfPath).Msg("Surat berhasil diprint")

	// Re-retrieve to return the updated record
	updatedSurat, _ := s.suratRepo.FindByID(ctx, id)
	sendJSON(w, http.StatusOK, updatedSurat)
}

// handleGetTemplate returns the HTML template for a specific jenis_surat.
func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jenisSuratID := chi.URLParam(r, "jenis_surat_id")

	if jenisSuratID == "" {
		sendError(w, http.StatusBadRequest, "ID Jenis Surat tidak boleh kosong")
		return
	}

	template, err := s.jenisSuratRepo.GetTemplate(ctx, jenisSuratID, s.cfg.DesaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Template tidak ditemukan untuk desa ini")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil template: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, template)
}

// handleNomorSuratStatus returns all batch statuses.
func (s *Server) handleNomorSuratStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	batches, err := s.nomorSuratRepo.ListAllBatches(ctx)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil status nomor surat: "+err.Error())
		return
	}

	if len(batches) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, batches)
}
