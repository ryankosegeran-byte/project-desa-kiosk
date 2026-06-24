package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/server/middleware"
	rfidpkg "github.com/project-desa-kiosk/server/rfid"
)

// handleRFIDRelay receives a scanned UID from a kiosk and broadcasts it to any
// admin-panel browser currently listening for that desa.
// Auth: kiosk API-key (same as sync endpoints).
func (s *Server) handleRFIDRelay(w http.ResponseWriter, r *http.Request) {
	kiosk := middleware.GetKiosk(r.Context())
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi")
		return
	}

	var req struct {
		UID string `json:"uid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload tidak valid: "+err.Error())
		return
	}
	if req.UID == "" {
		sendError(w, http.StatusBadRequest, "uid harus diisi")
		return
	}

	s.rfidRelay.Publish(kiosk.DesaID, req.UID)

	log.Info().
		Str("desa_id", kiosk.DesaID).
		Str("uid", req.UID).
		Msg("RFID UID di-relay dari kiosk ke admin panel")

	sendJSON(w, http.StatusOK, map[string]string{"status": "relayed", "uid": req.UID})
}

// handleRFIDStream sends Server-Sent Events to an admin-panel browser,
// streaming RFID UIDs scanned at the desa's kiosk in real time.
// Auth: JWT (dashboard user). Scoped to the user's desa_id.
func (s *Server) handleRFIDStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming tidak didukung", http.StatusInternalServerError)
		return
	}

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	desaID := claims.DesaID
	if desaID == "" {
		// Superadmin can optionally specify desa_id via query param
		desaID = r.URL.Query().Get("desa_id")
		if desaID == "" {
			sendError(w, http.StatusBadRequest, "desa_id diperlukan (superadmin: query param, PIC: otomatis)")
			return
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events := s.rfidRelay.Subscribe(desaID)
	defer s.rfidRelay.Unsubscribe(desaID, events)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Debug().Str("desa_id", desaID).Msg("Admin panel terhubung ke RFID stream")

	for {
		select {
		case <-r.Context().Done():
			log.Debug().Str("desa_id", desaID).Msg("RFID stream admin panel terputus")
			return

		case uid, ok := <-events:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", uid); err != nil {
				return
			}
			flusher.Flush()

		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// resolveSessionDesaID returns the desa scope for a dashboard user: PIC uses
// their own desa; superadmin must pass ?desa_id=. Returns "" with false if missing.
func resolveSessionDesaID(r *http.Request) (string, bool) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		return "", false
	}
	if claims.DesaID != "" {
		return claims.DesaID, true
	}
	q := r.URL.Query().Get("desa_id")
	if q == "" {
		return "", false
	}
	return q, true
}

// handleRFIDSessionStart marks a registration session active for the operator”s
// desa, carrying the warga name being registered. Auth: JWT (dashboard).
func (s *Server) handleRFIDSessionStart(w http.ResponseWriter, r *http.Request) {
	desaID, ok := resolveSessionDesaID(r)
	if !ok {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan")
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	s.rfidRelay.StartSession(desaID, name)
	log.Info().Str("desa_id", desaID).Str("name", name).Msg("Sesi pendaftaran RFID dimulai")
	st := s.rfidRelay.State(desaID)
	sendJSON(w, http.StatusOK, st)
}

// handleRFIDSessionStop ends the registration session. Auth: JWT (dashboard).
func (s *Server) handleRFIDSessionStop(w http.ResponseWriter, r *http.Request) {
	desaID, ok := resolveSessionDesaID(r)
	if !ok {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan")
		return
	}
	s.rfidRelay.StopSession(desaID)
	log.Info().Str("desa_id", desaID).Msg("Sesi pendaftaran RFID dihentikan")
	sendJSON(w, http.StatusOK, s.rfidRelay.State(desaID))
}

// handleKioskBusy lets a kiosk report whether a resident is mid-flow (creating a
// letter). While busy, the registration session is "pending". Auth: kiosk key.
func (s *Server) handleKioskBusy(w http.ResponseWriter, r *http.Request) {
	kiosk := middleware.GetKiosk(r.Context())
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi")
		return
	}
	var req struct {
		Busy bool `json:"busy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload tidak valid")
		return
	}
	s.rfidRelay.SetKioskBusy(kiosk.DesaID, req.Busy)
	sendJSON(w, http.StatusOK, s.rfidRelay.State(kiosk.DesaID))
}

// handleRFIDSessionStream streams the full session state (JSON) for a kiosk”s
// desa over SSE. Auth: kiosk API-key.
func (s *Server) handleRFIDSessionStream(w http.ResponseWriter, r *http.Request) {
	kiosk := middleware.GetKiosk(r.Context())
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi")
		return
	}
	s.streamSessionState(w, r, kiosk.DesaID)
}

// handleRFIDSessionAdminStream streams session state to the admin panel so it
// can show "pending" while the kiosk is busy. Auth: JWT (dashboard).
func (s *Server) handleRFIDSessionAdminStream(w http.ResponseWriter, r *http.Request) {
	desaID, ok := resolveSessionDesaID(r)
	if !ok {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan")
		return
	}
	s.streamSessionState(w, r, desaID)
}

// streamSessionState is the shared SSE loop emitting JSON SessionState.
func (s *Server) streamSessionState(w http.ResponseWriter, r *http.Request, desaID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming tidak didukung", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	statusCh := s.rfidRelay.SubscribeStatus(desaID)
	defer s.rfidRelay.UnsubscribeStatus(desaID, statusCh)

	writeSessionState(w, flusher, s.rfidRelay.State(desaID))

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case st, ok := <-statusCh:
			if !ok {
				return
			}
			writeSessionState(w, flusher, st)
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSessionState(w http.ResponseWriter, flusher http.Flusher, st rfidpkg.SessionState) {
	payload, err := json.Marshal(st)
	if err != nil {
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return
	}
	flusher.Flush()
}
