package sync

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/kiosk/config"
)

// Listener maintains a Server-Sent Events connection to the online hub and
// triggers an immediate incremental warga pull whenever the server signals a
// data change. This makes resident sync effectively real-time while the
// periodic polling in Engine remains as a fallback when the stream is down.
type Listener struct {
	cfg    *config.Config
	puller *Puller
	client *http.Client
}

// NewListener creates a sync event listener.
func NewListener(cfg *config.Config, puller *Puller) *Listener {
	return &Listener{
		cfg:    cfg,
		puller: puller,
		// No timeout: the SSE connection is long-lived.
		client: &http.Client{},
	}
}

// Start runs the listener loop in the background, reconnecting as needed.
func (l *Listener) Start(ctx context.Context) {
	if l.cfg.ServerURL == "" {
		log.Info().Msg("Sync Listener dinonaktifkan (tidak ada KIOSK_SERVER_URL)")
		return
	}
	go func() {
		log.Info().Msg("Sync Listener (real-time SSE) dijalankan")
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			l.connect(ctx)
			// Reconnect backoff; polling fallback keeps data fresh meanwhile.
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

func (l *Listener) connect(ctx context.Context) {
	url := l.cfg.ServerURL + "/api/sync/events"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-API-Key", l.cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := l.client.Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("Sync Listener gagal terhubung ke server (offline?)")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Debug().Int("status", resp.StatusCode).Msg("Sync Listener: server menolak koneksi events")
		return
	}

	log.Info().Msg("Sync Listener terhubung ke stream events server (real-time)")

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 4096), 1<<20)

	// Debounce bursts of events into a single pull.
	var lastPull time.Time
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue // ignore keepalive comments / blank lines
		}
		kind := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if kind == "" {
			continue
		}

		// Debounce: at most one pull per second regardless of event burst.
		if time.Since(lastPull) < time.Second {
			continue
		}
		lastPull = time.Now()

		if err := l.puller.PullWarga(ctx); err != nil {
			log.Debug().Err(err).Msg("Sync Listener: pull warga gagal (akan dicoba lagi via polling)")
		} else {
			log.Info().Str("kind", kind).Msg("Sync real-time: data warga diperbarui dari server")
		}
	}
}
