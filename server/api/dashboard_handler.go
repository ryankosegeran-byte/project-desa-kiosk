package api

import (
	"net/http"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleGetStats returns general statistics for the dashboard.
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
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

	// 1. Count Warga
	var totalWarga int
	wargaQuery := "SELECT COUNT(*) FROM warga WHERE ($1 = '' OR desa_id = $1::uuid)"
	err := s.db.QueryRowContext(ctx, wargaQuery, desaID).Scan(&totalWarga)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menghitung data warga: "+err.Error())
		return
	}

	// 2. Count Surat
	var totalSurat int
	suratQuery := "SELECT COUNT(*) FROM surat WHERE ($1 = '' OR desa_id = $1::uuid)"
	err = s.db.QueryRowContext(ctx, suratQuery, desaID).Scan(&totalSurat)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menghitung data surat: "+err.Error())
		return
	}

	// 3. Count Kiosks
	var totalKiosks int
	kioskQuery := "SELECT COUNT(*) FROM kiosks WHERE ($1 = '' OR desa_id = $1::uuid)"
	err = s.db.QueryRowContext(ctx, kioskQuery, desaID).Scan(&totalKiosks)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menghitung data kiosks: "+err.Error())
		return
	}

	// 4. Count active jenis surat
	var activeTypes int
	var typesQuery string
	var args []interface{}
	if desaID != "" {
		typesQuery = "SELECT COUNT(*) FROM desa_jenis_surat WHERE desa_id = $1 AND aktif = true"
		args = append(args, desaID)
	} else {
		typesQuery = "SELECT COUNT(*) FROM jenis_surat WHERE aktif_global = true"
	}
	err = s.db.QueryRowContext(ctx, typesQuery, args...).Scan(&activeTypes)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menghitung jenis surat: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"total_warga":        totalWarga,
		"total_surat":        totalSurat,
		"total_kiosks":       totalKiosks,
		"active_jenis_surat": activeTypes,
	})
}
