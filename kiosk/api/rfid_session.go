package api

import (
	"encoding/json"
	"net/http"
)

// handleRFIDSession reports whether a registration session is active for this
// kiosk”s desa, plus the name of the warga being registered. The kiosk UI polls
// this to switch into registration-scan mode and show the operator”s context.
func (s *Server) handleRFIDSession(w http.ResponseWriter, r *http.Request) {
	active := false
	name := ""
	if s.sessionWatcher != nil {
		active = s.sessionWatcher.IsActive()
		name = s.sessionWatcher.Name()
	}
	sendJSON(w, http.StatusOK, map[string]any{
		"active": active,
		"name":   name,
	})
}

// handleKioskBusy lets the kiosk UI report that a resident is mid-flow (creating
// a letter). This is forwarded to the server so the admin panel can show the
// registration request as "pending" until the resident finishes.
func (s *Server) handleKioskBusy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Busy bool `json:"busy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload tidak valid")
		return
	}
	if s.sessionWatcher != nil {
		s.sessionWatcher.SetBusy(r.Context(), req.Busy)
	}
	sendJSON(w, http.StatusOK, map[string]bool{"busy": req.Busy})
}
