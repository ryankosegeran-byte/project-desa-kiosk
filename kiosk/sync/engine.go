package sync

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/kiosk/config"
)

// Engine orchestrates background synchronization cycles.
type Engine struct {
	cfg      *config.Config
	detector *Detector
	pusher   *Pusher
	puller   *Puller
}

// NewEngine creates a new Engine instance.
func NewEngine(cfg *config.Config, detector *Detector, pusher *Pusher, puller *Puller) *Engine {
	return &Engine{
		cfg:      cfg,
		detector: detector,
		pusher:   pusher,
		puller:   puller,
	}
}

// Start launches the background synchronization loops.
func (e *Engine) Start(ctx context.Context) {
	if e.cfg.ServerURL == "" {
		log.Info().Msg("Sync Engine dinonaktifkan: KIOSK_SERVER_URL tidak diatur")
		return
	}

	interval := time.Duration(e.cfg.SyncInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		log.Info().Dur("interval", interval).Msg("Sync Engine background worker dimulai")
		
		// Run initial check immediately
		e.runSync(ctx)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Info().Msg("Sync Engine background worker dihentikan")
				return
			case <-ticker.C:
				e.runSync(ctx)
			}
		}
	}()
}

func (e *Engine) runSync(ctx context.Context) {
	// 1. Connection check
	if !e.detector.Check(ctx) {
		return
	}

	// 2. Push printed records
	if err := e.pusher.Push(ctx); err != nil {
		log.Error().Err(err).Msg("Sync push transaksi gagal")
	}

	// 3. Pull resident updates
	if err := e.puller.PullWarga(ctx); err != nil {
		log.Error().Err(err).Msg("Sync pull warga gagal")
	}

	// 4. Pull dynamic configurations
	if err := e.puller.PullConfig(ctx); err != nil {
		log.Error().Err(err).Msg("Sync pull konfigurasi surat gagal")
	}

	// 5. Pull nomor surat range
	if err := e.puller.PullNomorSurat(ctx); err != nil {
		log.Warn().Err(err).Msg("Gagal pull nomor surat batch (non-fatal)")
	}
}
