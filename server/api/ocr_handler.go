package api

import (
	"io"
	"net/http"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleOCRExtract handles uploading a KTP photo and extracting its information via OCR.
func (s *Server) handleOCRExtract(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	// Max 5 MB files
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Ukuran file terlalu besar (maksimal 5MB)")
		return
	}

	file, _, err := r.FormFile("foto_ktp")
	if err != nil {
		sendError(w, http.StatusBadRequest, "File foto_ktp diperlukan")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membaca file: "+err.Error())
		return
	}

	// Call AI OCR Service orchestrator
	extracted, err := s.ocrService.ExtractKTP(ctx, fileBytes)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "AI OCR Gagal mengekstraksi KTP: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:     claims.UserID,
		DesaID:     claims.DesaID,
		Action:     "OCR_KTP",
		EntityType: "warga",
		Detail:     `{"confidence": ` + string(rune(extracted.Confidence)) + `}`,
		IPAddress:  r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, extracted)
}
