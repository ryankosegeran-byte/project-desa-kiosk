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

	deleted := r.URL.Query().Get("deleted") == "true"

	var wargaList []models.Warga
	var err error
	if deleted {
		wargaList, err = s.wargaRepo.ListDeleted(ctx, desaID)
	} else {
		wargaList, err = s.wargaRepo.List(ctx, desaID)
	}
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

	// Notify kiosks of the change so they sync immediately.
	s.rfidRelay.NotifySync(req.DesaID, "warga")

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

	s.rfidRelay.NotifySync(warga.DesaID, "warga")

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
	s.rfidRelay.NotifySync(warga.DesaID, "warga")
	sendJSON(w, http.StatusOK, warga)
}

// handleCreateDraft creates a draft warga record (partial data, no NIK/nama required).
func (s *Server) handleCreateDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	var req struct {
		DesaID      string `json:"desa_id"`
		FotoKTPPath string `json:"foto_ktp_path,omitempty"`
		NIK         string `json:"nik,omitempty"`
		Nama        string `json:"nama,omitempty"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	// Enforce village mapping
	if claims.Role == models.RolePICDesa {
		req.DesaID = claims.DesaID
	} else if req.DesaID == "" {
		sendError(w, http.StatusBadRequest, "Desa ID harus diisi")
		return
	}

	draftToken := uuid.New().String()
	draft := &models.Warga{
		ID:          uuid.New().String(),
		DesaID:      req.DesaID,
		NIK:         req.NIK,
		Nama:        req.Nama,
		FotoKTPPath: req.FotoKTPPath,
		Status:      "draft",
		DraftToken:  draftToken,
	}

	if err := s.wargaRepo.Create(ctx, draft); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat draft warga: "+err.Error())
		return
	}

	// Audit activity logging
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "CREATE_DRAFT_WARGA",
		EntityType: "warga",
		EntityID:   draft.ID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusCreated, map[string]string{
		"id":          draft.ID,
		"draft_token": draftToken,
		"url":         "/warga/draft?token=" + draftToken,
	})
}

// handleGetDraft returns a draft warga record by its token.
func (s *Server) handleGetDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		sendError(w, http.StatusBadRequest, "Token draft tidak boleh kosong")
		return
	}

	draft, err := s.wargaRepo.FindByDraftToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Draft tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data draft: "+err.Error())
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && draft.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	sendJSON(w, http.StatusOK, draft)
}

// handleCompleteDraft transitions a draft warga record to complete status.
func (s *Server) handleCompleteDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		sendError(w, http.StatusBadRequest, "Token draft tidak boleh kosong")
		return
	}

	// Find existing draft
	draft, err := s.wargaRepo.FindByDraftToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Draft tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data draft: "+err.Error())
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && draft.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	var req models.Warga
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.NIK == "" || req.Nama == "" {
		sendError(w, http.StatusBadRequest, "NIK dan Nama harus diisi untuk menyelesaikan registrasi")
		return
	}

	// Check if NIK already exists on another complete record
	existing, err := s.wargaRepo.FindByNIK(ctx, req.NIK)
	if err == nil && existing != nil && existing.ID != draft.ID {
		sendError(w, http.StatusConflict, "NIK sudah terdaftar pada warga lain")
		return
	}

	// Copy full data into draft record
	draft.NIK = req.NIK
	draft.Nama = req.Nama
	draft.TempatLahir = req.TempatLahir
	draft.TanggalLahir = req.TanggalLahir
	draft.JenisKelamin = req.JenisKelamin
	draft.Alamat = req.Alamat
	draft.RT = req.RT
	draft.RW = req.RW
	draft.Kelurahan = req.Kelurahan
	draft.Kecamatan = req.Kecamatan
	draft.Kabupaten = req.Kabupaten
	draft.Provinsi = req.Provinsi
	draft.Agama = req.Agama
	draft.StatusKawin = req.StatusKawin
	draft.Pekerjaan = req.Pekerjaan
	draft.Kewarganegaraan = req.Kewarganegaraan
	draft.FotoKTPPath = req.FotoKTPPath
	if req.RFIDUID != "" {
		draft.RFIDUID = req.RFIDUID
	}

	if err := s.wargaRepo.UpdateToComplete(ctx, draft); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyelesaikan registrasi: "+err.Error())
		return
	}

	// Audit activity logging
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "COMPLETE_DRAFT_WARGA",
		EntityType: "warga",
		EntityID:   draft.ID,
		IPAddress:  r.RemoteAddr,
	})

	draft.Status = "complete"
	draft.DraftToken = ""
	s.rfidRelay.NotifySync(draft.DesaID, "warga")
	sendJSON(w, http.StatusOK, draft)
}

// handleDeleteWarga deletes a warga record by ID.
func (s *Server) handleDeleteWarga(w http.ResponseWriter, r *http.Request) {
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

	// Check warga exists and enforce PIC isolation
	warga, err := s.wargaRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Warga tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal memproses data: "+err.Error())
		return
	}

	if claims.Role == models.RolePICDesa && warga.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	if err := s.wargaRepo.Delete(ctx, id); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menghapus warga: "+err.Error())
		return
	}

	// Audit
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "DELETE_WARGA",
		EntityType: "warga",
		EntityID:   id,
		IPAddress:  r.RemoteAddr,
	})

	// Soft-delete propagates to kiosks (record carries deleted_at on the next
	// pull). Notify so they remove the local row immediately.
	s.rfidRelay.NotifySync(warga.DesaID, "warga")

	sendJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

// handleHardDeleteWarga permanently removes a soft-deleted warga record.
func (s *Server) handleHardDeleteWarga(w http.ResponseWriter, r *http.Request) {
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

	// Enforce PIC isolation via lookup of the (possibly soft-deleted) record.
	warga, err := s.wargaRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "Warga tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal memproses data: "+err.Error())
		return
	}
	if claims.Role == models.RolePICDesa && warga.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	if err := s.wargaRepo.HardDelete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusBadRequest, "Hanya data yang sudah dihapus yang bisa dihapus permanen")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal menghapus permanen: "+err.Error())
		return
	}

	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "HARD_DELETE_WARGA",
		EntityType: "warga",
		EntityID:   id,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, map[string]string{"status": "permanently_deleted", "id": id})
}
