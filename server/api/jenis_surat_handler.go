package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListJenisSurat lists all global types of letters.
func (s *Server) handleListJenisSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.jenisSuratRepo.List(ctx)
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

// handleCreateJenisSurat creates a new letter schema (Superadmin only).
func (s *Server) handleCreateJenisSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.JenisSurat
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.Kode == "" || req.Nama == "" || len(req.FieldsSchema) == 0 {
		sendError(w, http.StatusBadRequest, "Kode, Nama, dan Fields Schema harus diisi")
		return
	}

	// Check duplicates
	existing, err := s.jenisSuratRepo.FindByKode(ctx, req.Kode)
	if err == nil && existing != nil {
		sendError(w, http.StatusConflict, "Kode jenis surat sudah digunakan")
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if err := s.jenisSuratRepo.Create(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan jenis surat: "+err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, req)
}

// handleUpdateJenisSurat edits an existing letter schema (Superadmin only).
func (s *Server) handleUpdateJenisSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID tidak boleh kosong")
		return
	}

	js, err := s.jenisSuratRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Jenis surat tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mencari jenis surat: "+err.Error())
		return
	}

	var req models.JenisSurat
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	js.Kode = req.Kode
	js.Nama = req.Nama
	js.Deskripsi = req.Deskripsi
	js.FieldsSchema = req.FieldsSchema
	js.Aktif = req.Aktif
	js.Urutan = req.Urutan

	if err := s.jenisSuratRepo.Update(ctx, js); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memperbarui jenis surat: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, js)
}

// handleToggleDesaJenisSurat enables/disables or orders a letter schema for a specific village.
func (s *Server) handleToggleDesaJenisSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := chi.URLParam(r, "id")
	if desaID == "" {
		sendError(w, http.StatusBadRequest, "Desa ID tidak boleh kosong")
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && claims.DesaID != desaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: Anda hanya dapat mengatur desa Anda sendiri")
		return
	}

	var req struct {
		JenisSuratID string `json:"jenis_surat_id"`
		Aktif        bool   `json:"aktif"`
		Urutan       int    `json:"urutan"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.JenisSuratID == "" {
		sendError(w, http.StatusBadRequest, "jenis_surat_id harus diisi")
		return
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

	err = s.jenisSuratRepo.ToggleForDesa(ctx, desaID, req.JenisSuratID, req.Aktif, req.Urutan)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengkonfigurasi jenis surat desa: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "TOGGLE_JENIS_SURAT",
		EntityType: "jenis_surat",
		EntityID:   req.JenisSuratID,
		Detail:     `{"aktif": ` + strconv.FormatBool(req.Aktif) + `, "urutan": ` + strconv.Itoa(req.Urutan) + `}`,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
