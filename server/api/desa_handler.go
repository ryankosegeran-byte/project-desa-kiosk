package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/db"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListDesa returns all registered villages (Superadmin only).
func (s *Server) handleListDesa(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	list, err := s.desaRepo.List(ctx)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil daftar desa: "+err.Error())
		return
	}

	if len(list) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, list)
}

// handleCreateDesa creates a new village profile (Superadmin only).
func (s *Server) handleCreateDesa(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.Desa
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.Nama == "" || req.KodeDesa == "" {
		sendError(w, http.StatusBadRequest, "Nama Desa dan Kode Desa harus diisi")
		return
	}

	// Check duplicates
	existing, err := s.desaRepo.FindByKode(ctx, req.KodeDesa)
	if err == nil && existing != nil {
		sendError(w, http.StatusConflict, "Kode desa sudah digunakan")
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	if req.Theme == "" {
		req.Theme = "merah-putih"
	}

	if err := s.desaRepo.Create(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat desa: "+err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, req)
}

// handleListKiosks returns kiosks registered in the system (tenant-isolated).
func (s *Server) handleListKiosks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := r.URL.Query().Get("desa_id")
	if claims.Role == models.RolePICDesa {
		desaID = claims.DesaID
	}

	kiosks, err := s.desaRepo.ListKiosks(ctx, desaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil daftar kiosk: "+err.Error())
		return
	}

	if len(kiosks) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, kiosks)
}

// handleRegisterKiosk registers a new Kiosk terminal and returns the API key.
func (s *Server) handleRegisterKiosk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	var req models.Kiosk
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.Nama == "" {
		sendError(w, http.StatusBadRequest, "Nama Kiosk harus diisi")
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa {
		req.DesaID = claims.DesaID
	} else if req.DesaID == "" {
		sendError(w, http.StatusBadRequest, "Desa ID harus diisi")
		return
	}

	// Verify desa exists
	_, err := s.desaRepo.FindByID(ctx, req.DesaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusBadRequest, "Desa tidak terdaftar")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal verifikasi desa: "+err.Error())
		return
	}

	// Generate UUID & Secret API Key
	req.ID = uuid.New().String()
	req.APIKey = "kiosk_" + uuid.New().String()
	req.Status = "active"

	if err := s.desaRepo.RegisterKiosk(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mendaftarkan kiosk: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "REGISTER_KIOSK",
		EntityType: "kiosk",
		EntityID:   req.ID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusCreated, req)
}

// handleUpdateDesaTheme updates the kiosk theme for a village (Superadmin only).
func (s *Server) handleUpdateDesaTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "ID desa harus diisi")
		return
	}

	var req struct {
		Theme string `json:"theme"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}
	if !db.AllowedThemes[req.Theme] {
		sendError(w, http.StatusBadRequest, "Tema tidak dikenal")
		return
	}

	if err := s.desaRepo.UpdateTheme(ctx, id, req.Theme); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memperbarui tema: "+err.Error())
		return
	}

	// Push a real-time config-change signal so kiosks update their theme
	// immediately instead of waiting for the next polling tick.
	s.rfidRelay.NotifySync(id, "config")

	sendJSON(w, http.StatusOK, map[string]string{"id": id, "theme": req.Theme})
}