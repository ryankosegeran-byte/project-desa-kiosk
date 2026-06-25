package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/server/middleware"
)

// handleSyncEventsStream streams Server-Sent Events to a kiosk, signalling when
// resident data for the kiosk's desa has changed. On each event the kiosk runs
// an immediate incremental sync pull instead of waiting for its polling tick.
//
// Auth: X-API-Key (KioskKeyMiddleware). Scoped to the kiosk's desa_id.
func (s *Server) handleSyncEventsStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming tidak didukung", http.StatusInternalServerError)
		return
	}

	kiosk := middleware.GetKiosk(r.Context())
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi di context")
		return
	}
	desaID := kiosk.DesaID

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events := s.rfidRelay.SubscribeSync(desaID)
	defer s.rfidRelay.UnsubscribeSync(desaID, events)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Debug().Str("desa_id", desaID).Msg("Kiosk terhubung ke sync events stream")

	// Send an initial hello so the kiosk knows the stream is live and can do a
	// reconciliation pull right away.
	if _, err := fmt.Fprintf(w, "data: warga\n\n"); err != nil {
		return
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			log.Debug().Str("desa_id", desaID).Msg("Sync events stream kiosk terputus")
			return

		case kind, ok := <-events:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", kind); err != nil {
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
