package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleSyncPush processes the push sync queue from local kiosk (new/printed letters).
func (s *Server) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	kiosk := middleware.GetKiosk(ctx)
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi di context")
		return
	}

	var payload models.SyncPushPayload
	if err := parseJSON(r, &payload); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	// Verify desa matching
	if payload.DesaID != kiosk.DesaID {
		sendError(w, http.StatusBadRequest, "Desa ID payload tidak cocok dengan key Kiosk")
		return
	}

	accepted := 0
	rejected := 0
	var errorsList []string

	for _, item := range payload.Items {
		if item.EntityType == "surat" {
			var surat models.Surat
			if err := json.Unmarshal(item.Payload, &surat); err != nil {
				rejected++
				errorsList = append(errorsList, "Gagal unmarshal surat ID "+item.EntityID+": "+err.Error())
				continue
			}

			// Force server matching rules
			surat.DesaID = kiosk.DesaID

			if err := s.suratRepo.Upsert(ctx, &surat, kiosk.ID); err != nil {
				rejected++
				errorsList = append(errorsList, "Gagal simpan surat ID "+item.EntityID+": "+err.Error())
				continue
			}
			accepted++
		} else {
			rejected++
			errorsList = append(errorsList, "Entity type tidak disupport: "+item.EntityType)
		}
	}

	// Update Kiosk sync timestamps
	_ = s.desaRepo.UpdateKioskSyncTime(ctx, kiosk.ID)
	// Update Kiosk last seen IP Address
	_ = s.desaRepo.UpdateKioskStatus(ctx, kiosk.ID, r.RemoteAddr)

	sendJSON(w, http.StatusOK, models.SyncPushResponse{
		Accepted: accepted,
		Rejected: rejected,
		Errors:   errorsList,
	})
}

// handleSyncPullWarga sends warga updates for this desa since the last sync time.
func (s *Server) handleSyncPullWarga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	kiosk := middleware.GetKiosk(ctx)
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi di context")
		return
	}

	lastSyncStr := r.URL.Query().Get("last_sync")
	var lastSync time.Time
	var err error

	if lastSyncStr != "" {
		lastSync, err = time.Parse(time.RFC3339, lastSyncStr)
		if err != nil {
			// fallback to unix timestamp representation if RFC3339 parse fails
			sendError(w, http.StatusBadRequest, "Parameter last_sync tidak valid, gunakan format RFC3339")
			return
		}
	} else {
		lastSync = time.Time{} // zero time: pulls all warga
	}

	wargaList, err := s.wargaRepo.ListUpdatedSince(ctx, kiosk.DesaID, lastSync)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memproses data warga sync: "+err.Error())
		return
	}

	// Update Kiosk sync timestamps
	_ = s.desaRepo.UpdateKioskSyncTime(ctx, kiosk.ID)
	// Update Kiosk heartbeat status
	_ = s.desaRepo.UpdateKioskStatus(ctx, kiosk.ID, r.RemoteAddr)

	sendJSON(w, http.StatusOK, models.SyncPullWargaResponse{
		Warga:    wargaList,
		SyncedAt: time.Now(),
		HasMore:  false, // simplicity for now: no pagination needed for small village datasets
	})
}

// handleSyncPullConfig pulls available letters and templates config for this desa.
// Returns both per-desa templates and general templates.
func (s *Server) handleSyncPullConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	kiosk := middleware.GetKiosk(ctx)
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi di context")
		return
	}

	// Fetch active letter types for this village
	jenisSuratList, err := s.jenisSuratRepo.ListActiveForDesa(ctx, kiosk.DesaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil daftar jenis surat sync: "+err.Error())
		return
	}

	// Fetch ALL templates for this village including general templates, WITH the
	// DOCX bytes + placeholder mapping so the kiosk can render offline (Strategi B).
	// The kiosk handles fallback logic (per-desa > general).
	templatesList, err := s.templateRepo.ListTemplatesForDesaFull(ctx, kiosk.DesaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil templates sync: "+err.Error())
		return
	}

	// Resolve the kiosk theme from the village profile.
	theme := "merah-putih"
	if desa, derr := s.desaRepo.FindByID(ctx, kiosk.DesaID); derr == nil && desa != nil && desa.Theme != "" {
		theme = desa.Theme
	}

	// Update Kiosk heartbeat status
	_ = s.desaRepo.UpdateKioskStatus(ctx, kiosk.ID, r.RemoteAddr)

	sendJSON(w, http.StatusOK, models.SyncPullConfigResponse{
		JenisSurat: jenisSuratList,
		Templates:  templatesList,
		Theme:      theme,
		SyncedAt:   time.Now(),
	})
}
