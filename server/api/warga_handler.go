package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListWarga lists all residents, enforcing tenant isolation.
func (s *Server) handleListWarga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := r.URL.Query().Get("desa_id")
	// Enforce PIC Desa isolation
	if claims.Role == models.RolePICDesa {
		desaID = claims.DesaID
	}

	wargaList, err := s.wargaRepo.List(ctx, desaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data warga: "+err.Error())
		return
	}

	if len(wargaList) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, wargaList)
}

// handleCreateWarga creates a new resident.
func (s *Server) handleCreateWarga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	var req models.Warga
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.NIK == "" || req.Nama == "" {
		sendError(w, http.StatusBadRequest, "NIK dan Nama harus diisi")
		return
	}

	// Enforce village mapping
	if claims.Role == models.RolePICDesa {
		req.DesaID = claims.DesaID
	} else if req.DesaID == "" {
		sendError(w, http.StatusBadRequest, "Desa ID harus diisi")
		return
	}

	// Generate UUID
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// Check if already exists
	existing, err := s.wargaRepo.FindByNIK(ctx, req.NIK)
	if err == nil && existing != nil {
		sendError(w, http.StatusConflict, "NIK sudah terdaftar")
		return
	}

	if err := s.wargaRepo.Create(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat data warga: "+err.Error())
		return
	}

	// Audit activity logging
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "REGISTER_WARGA",
		EntityType: "warga",
		EntityID:   req.ID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusCreated, req)
}

// handleUpdateWarga updates a resident's details.
func (s *Server) handleUpdateWarga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "ID warga tidak boleh kosong")
		return
	}

	// Find existing record
	warga, err := s.wargaRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Warga tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data warga: "+err.Error())
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && warga.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	var req models.Warga
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	// Copy updateable fields
	warga.NIK = req.NIK
	warga.Nama = req.Nama
	warga.TempatLahir = req.TempatLahir
	warga.TanggalLahir = req.TanggalLahir
	warga.JenisKelamin = req.JenisKelamin
	warga.Alamat = req.Alamat
	warga.RT = req.RT
	warga.RW = req.RW
	warga.Kelurahan = req.Kelurahan
	warga.Kecamatan = req.Kecamatan
	warga.Kabupaten = req.Kabupaten
	warga.Provinsi = req.Provinsi
	warga.Agama = req.Agama
	warga.StatusKawin = req.StatusKawin
	warga.Pekerjaan = req.Pekerjaan
	warga.Kewarganegaraan = req.Kewarganegaraan

	// If rfid_uid is updated here, preserve it unless explicitly requested otherwise
	if req.RFIDUID != "" {
		warga.RFIDUID = req.RFIDUID
	}

	if err := s.wargaRepo.Update(ctx, warga); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memperbarui data warga: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, warga)
}

// handleLinkRFID associates an RFID UID with a resident.
func (s *Server) handleLinkRFID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "ID warga tidak boleh kosong")
		return
	}

	var req struct {
		RFIDUID string `json:"rfid_uid"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.RFIDUID == "" {
		sendError(w, http.StatusBadRequest, "rfid_uid harus diisi")
		return
	}

	// Find warga
	warga, err := s.wargaRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Warga tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal memproses data warga: "+err.Error())
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && warga.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	// Check if this RFID UID is already linked to another person
	existing, err := s.wargaRepo.FindByRFID(ctx, req.RFIDUID)
	if err == nil && existing != nil && existing.ID != warga.ID {
		sendError(w, http.StatusConflict, "Kartu RFID/KTP ini sudah ditautkan ke warga lain: "+existing.Nama)
		return
	}

	if err := s.wargaRepo.LinkRFID(ctx, warga.ID, req.RFIDUID); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menautkan kartu RFID: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "LINK_RFID",
		EntityType: "warga",
		EntityID:   warga.ID,
		Detail:     `{"rfid_uid": "` + req.RFIDUID + `"}`,
		IPAddress:  r.RemoteAddr,
	})

	warga.RFIDUID = req.RFIDUID
	sendJSON(w, http.StatusOK, warga)
}
