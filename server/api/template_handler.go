package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListTemplates lists all letter templates for a village.
func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
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

	list, err := s.templateRepo.ListTemplatesForDesa(ctx, desaID)
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

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa {
		req.DesaID = claims.DesaID
	} else if req.DesaID == "" {
		sendError(w, http.StatusBadRequest, "DesaID harus diisi")
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

	// Check if this is a new template or overwrite
	existing, err := s.templateRepo.GetTemplate(ctx, req.JenisSuratID, req.DesaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			req.ID = uuid.New().String()
			req.Version = 1
		} else {
			sendError(w, http.StatusInternalServerError, "Gagal memeriksa template: "+err.Error())
			return
		}
	} else {
		req.ID = existing.ID
		req.Version = existing.Version
	}

	req.CreatedBy = claims.UserID

	if err := s.templateRepo.UpsertTemplate(ctx, &req); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan template: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "UPSERT_TEMPLATE",
		EntityType: "template",
		EntityID:   req.ID,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, req)
}
