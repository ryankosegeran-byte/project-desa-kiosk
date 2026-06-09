package sync

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type Detector struct {
	serverURL string
	isOnline  int32 // atomic bool (0 or 1)
	client    *http.Client
}

func NewDetector(serverURL string) *Detector {
	return &Detector{
		serverURL: serverURL,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Check pings the server to detect connection status.
func (d *Detector) Check(ctx context.Context) bool {
	if d.serverURL == "" {
		atomic.StoreInt32(&d.isOnline, 0)
		return false
	}

	url := d.serverURL + "/ping"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		atomic.StoreInt32(&d.isOnline, 0)
		return false
	}

	resp, err := d.client.Do(req)
	if err != nil {
		if atomic.LoadInt32(&d.isOnline) == 1 {
			log.Warn().Err(err).Msg("Kiosk offline: Gagal menghubungi server hub")
		}
		atomic.StoreInt32(&d.isOnline, 0)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if atomic.LoadInt32(&d.isOnline) == 0 {
			log.Info().Msg("Kiosk online: Terhubung ke server hub")
		}
		atomic.StoreInt32(&d.isOnline, 1)
		return true
	}

	atomic.StoreInt32(&d.isOnline, 0)
	return false
}

// IsOnline returns the last checked online status.
func (d *Detector) IsOnline() bool {
	return atomic.LoadInt32(&d.isOnline) == 1
}
