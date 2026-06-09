package rfid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// MockScanRequest represents the JSON payload to trigger a mock RFID scan.
type MockScanRequest struct {
	UID string `json:"uid"`
}

// ServeEvents returns an HTTP handler that streams RFID scan events using Server-Sent Events (SSE).
func ServeEvents(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming tidak didukung", http.StatusInternalServerError)
			return
		}

		// Set headers for SSE stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Subscribe to broker events
		events := broker.Subscribe()
		defer broker.Unsubscribe(events)

		// Create a ticker for keeping connection alive
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		log.Debug().Msg("Memulai stream SSE RFID untuk client baru")

		for {
			select {
			case <-r.Context().Done():
				log.Debug().Msg("Stream SSE RFID dihentikan oleh client")
				return

			case uid, ok := <-events:
				if !ok {
					return
				}
				// Format message as SSE data event
				_, err := fmt.Fprintf(w, "data: %s\n\n", uid)
				if err != nil {
					log.Error().Err(err).Msg("Gagal menulis event SSE RFID ke client")
					return
				}
				flusher.Flush()

			case <-ticker.C:
				// Send keep-alive comment
				_, err := fmt.Fprintf(w, ": keepalive\n\n")
				if err != nil {
					log.Debug().Msg("Gagal menulis keepalive SSE RFID ke client (kemungkinan terputus)")
					return
				}
				flusher.Flush()
			}
		}
	}
}

// HandleMockScan returns an HTTP handler that allows triggering mock RFID scan events via POST.
func HandleMockScan(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method tidak didukung", http.StatusMethodNotAllowed)
			return
		}

		var req MockScanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Payload JSON tidak valid"})
			return
		}

		if req.UID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "UID tidak boleh kosong"})
			return
		}

		// Broadcast to all listening SSE clients
		broker.Publish(req.UID)

		log.Info().Str("uid", req.UID).Msg("Mock scan RFID diterima dan dipublikasikan")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "uid": req.UID})
	}
}
