package rfid

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// SessionWatcher connects to the online server”s registration-session SSE stream
// and tracks whether a registration session is active for this kiosk”s desa,
// along with the name of the warga being registered. It also reports kiosk-busy
// state back to the server so the admin panel can show "pending".
type SessionWatcher struct {
	serverURL string
	apiKey    string

	active int32 // atomic bool
	busy   int32 // atomic bool (local resident mid-flow)

	mu   sync.RWMutex
	name string

	client *http.Client
}

// NewSessionWatcher creates a watcher. ServerURL/apiKey from kiosk config.
func NewSessionWatcher(serverURL, apiKey string) *SessionWatcher {
	return &SessionWatcher{
		serverURL: serverURL,
		apiKey:    apiKey,
		client:    &http.Client{},
	}
}

// IsActive reports whether a registration session is active.
func (sw *SessionWatcher) IsActive() bool {
	return atomic.LoadInt32(&sw.active) == 1
}

// Name returns the name of the warga currently being registered (may be empty).
func (sw *SessionWatcher) Name() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.name
}

// SetBusy reports to the server whether a resident is mid-flow on the kiosk.
// It debounces redundant reports.
func (sw *SessionWatcher) SetBusy(ctx context.Context, busy bool) {
	var b int32
	if busy {
		b = 1
	}
	if atomic.SwapInt32(&sw.busy, b) == b {
		return // no change
	}
	if sw.serverURL == "" {
		return
	}
	go sw.reportBusy(ctx, busy)
}

func (sw *SessionWatcher) reportBusy(ctx context.Context, busy bool) {
	body, _ := json.Marshal(map[string]bool{"busy": busy})
	req, err := http.NewRequestWithContext(ctx, "POST", sw.serverURL+"/api/kiosk/busy", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", sw.apiKey)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("Gagal lapor status sibuk kiosk ke server")
		return
	}
	resp.Body.Close()
}

// Start runs the watcher loop in the background, reconnecting as needed.
func (sw *SessionWatcher) Start(ctx context.Context) {
	if sw.serverURL == "" {
		log.Info().Msg("RFID SessionWatcher dinonaktifkan (tidak ada KIOSK_SERVER_URL)")
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			sw.connect(ctx)
			atomic.StoreInt32(&sw.active, 0)
			sw.mu.Lock()
			sw.name = ""
			sw.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
	log.Info().Msg("RFID SessionWatcher dijalankan")
}

type sessionStatePayload struct {
	Active    bool   `json:"active"`
	Name      string `json:"name"`
	KioskBusy bool   `json:"kiosk_busy"`
	Pending   bool   `json:"pending"`
}

func (sw *SessionWatcher) connect(ctx context.Context) {
	url := sw.serverURL + "/api/rfid/session/stream"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-API-Key", sw.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := sw.client.Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("SessionWatcher gagal terhubung ke server (offline?)")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Debug().Int("status", resp.StatusCode).Msg("SessionWatcher: server menolak koneksi sesi")
		return
	}

	log.Info().Msg("SessionWatcher terhubung ke stream sesi pendaftaran server")

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 4096), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		raw := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		var st sessionStatePayload
		if err := json.Unmarshal([]byte(raw), &st); err != nil {
			continue
		}
		sw.mu.Lock()
		sw.name = st.Name
		sw.mu.Unlock()
		if st.Active {
			if atomic.SwapInt32(&sw.active, 1) == 0 {
				log.Info().Str("nama", st.Name).Msg("Mode pendaftaran AKTIF (dari admin panel)")
			}
		} else {
			if atomic.SwapInt32(&sw.active, 0) == 1 {
				log.Info().Msg("Mode pendaftaran NONAKTIF")
			}
		}
	}
}
