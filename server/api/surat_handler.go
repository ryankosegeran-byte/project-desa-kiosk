package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListSurat returns all printed/synced letters from kiosks.
func (s *Server) handleListSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := r.URL.Query().Get("desa_id")
	// Enforce tenant isolation for PIC Desa
	if claims.Role == models.RolePICDesa {
		desaID = claims.DesaID
	}

	// Pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	suratList, err := s.suratRepo.List(ctx, desaID, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data surat: "+err.Error())
		return
	}

	if len(suratList) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, suratList)
}

// handleGetSurat returns details of a single letter.
func (s *Server) handleGetSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

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
		sendError(w, http.StatusInternalServerError, "Gagal mengambil detail surat: "+err.Error())
		return
	}

	// Enforce PIC isolation
	if claims.Role == models.RolePICDesa && surat.DesaID != claims.DesaID {
		sendError(w, http.StatusForbidden, "Akses ditolak: desa tidak sesuai")
		return
	}

	sendJSON(w, http.StatusOK, surat)
}
