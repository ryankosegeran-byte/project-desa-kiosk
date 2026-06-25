package rfid

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ForwarderConfig configures the server relay forwarder.
type ForwarderConfig struct {
	ServerURL string // e.g. "https://server.example.com"
	APIKey    string // kiosk API key for authentication
}

// StartForwarder subscribes to the local broker and forwards every scanned UID
// to the online server's /api/rfid/relay endpoint so that admin-panel users can
// receive UIDs in real time. It is a best-effort, fire-and-forget relay: if the
// server is unreachable the UID is simply skipped (the primary SSE path to the
// local kiosk UI is unaffected).
func StartForwarder(ctx context.Context, broker *Broker, cfg ForwarderConfig) {
	if cfg.ServerURL == "" {
		log.Info().Msg("RFID forwarder dinonaktifkan (tidak ada KIOSK_SERVER_URL)")
		return
	}

	events := broker.Subscribe()

	go func() {
		defer broker.Unsubscribe(events)

		client := &http.Client{Timeout: 5 * time.Second}
		relayURL := cfg.ServerURL + "/api/rfid/relay"

		log.Info().Str("relay_url", relayURL).Msg("RFID forwarder ke server dimulai")

		for {
			select {
			case <-ctx.Done():
				return
			case uid, ok := <-events:
				if !ok {
					return
				}
				go relayUID(client, relayURL, cfg.APIKey, uid)
			}
		}
	}()
}

func relayUID(client *http.Client, url, apiKey, uid string) {
	body, _ := json.Marshal(map[string]string{"uid": uid})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Msg("Gagal relay UID ke server (server offline?)")
		return
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Debug().Str("uid", uid).Msg("UID berhasil di-relay ke server")
	} else {
		log.Warn().Int("status", resp.StatusCode).Str("uid", uid).Msg("Server menolak relay UID")
	}
}
