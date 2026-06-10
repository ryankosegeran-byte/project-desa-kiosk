package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/kiosk/config"
	"github.com/project-desa-kiosk/kiosk/db"
)

type Puller struct {
	cfg            *config.Config
	wargaRepo      *db.WargaRepository
	jenisSuratRepo *db.JenisSuratRepository
	configRepo     *db.ConfigRepository
	client         *http.Client
}

func NewPuller(
	cfg *config.Config,
	wargaRepo *db.WargaRepository,
	jenisSuratRepo *db.JenisSuratRepository,
	configRepo *db.ConfigRepository,
) *Puller {
	return &Puller{
		cfg:            cfg,
		wargaRepo:      wargaRepo,
		jenisSuratRepo: jenisSuratRepo,
		configRepo:     configRepo,
		client:         &http.Client{Timeout: 15 * time.Second},
	}
}

// PullWarga fetches new/updated residents from the server hub.
func (p *Puller) PullWarga(ctx context.Context) error {
	// 1. Get last sync timestamp
	lastSyncStr, err := p.configRepo.Get(ctx, "last_sync_at")
	if err != nil {
		lastSyncStr = ""
	}

	url := fmt.Sprintf("%s/api/sync/pull/warga?last_sync=%s", p.cfg.ServerURL, lastSyncStr)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create pull warga request: %w", err)
	}
	req.Header.Set("X-API-Key", p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pull warga request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	var response models.SyncPullWargaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse sync warga response: %w", err)
	}

	if len(response.Warga) == 0 {
		return nil
	}

	log.Info().Int("count", len(response.Warga)).Msg("Mengunduh data warga baru dari server hub...")

	// 2. Upsert each warga locally
	for _, w := range response.Warga {
		if err := p.wargaRepo.Upsert(ctx, &w); err != nil {
			log.Error().Err(err).Str("nik", w.NIK).Msg("Gagal menyimpan warga secara lokal")
		}
	}

	// 3. Save last sync time in SQLite config
	syncedAtStr := response.SyncedAt.Format(time.RFC3339)
	_ = p.configRepo.Set(ctx, "last_sync_at", syncedAtStr)

	log.Info().Msg("Sync pull warga selesai")
	return nil
}

// PullConfig fetches active types of letters and custom HTML templates.
func (p *Puller) PullConfig(ctx context.Context) error {
	url := p.cfg.ServerURL + "/api/sync/pull/config"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create pull config request: %w", err)
	}
	req.Header.Set("X-API-Key", p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pull config request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	var response models.SyncPullConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse pull config response: %w", err)
	}

	// 1. Sync active types
	if len(response.JenisSurat) > 0 {
		log.Info().Int("count", len(response.JenisSurat)).Msg("Sinkronisasi jenis surat aktif...")
		for _, js := range response.JenisSurat {
			if err := p.jenisSuratRepo.Upsert(ctx, &js); err != nil {
				log.Error().Err(err).Str("kode", js.Kode).Msg("Gagal upsert jenis_surat secara lokal")
			}
		}
	}

	// 2. Sync template documents
	if len(response.Templates) > 0 {
		log.Info().Int("count", len(response.Templates)).Msg("Sinkronisasi template cetak dokumen...")
		for _, t := range response.Templates {
			if err := p.jenisSuratRepo.UpsertTemplate(ctx, &t); err != nil {
				log.Error().Err(err).Str("template_id", t.ID).Msg("Gagal upsert template secara lokal")
			}
		}
	}

	// 3. Save last sync time in SQLite config
	syncedAtStr := time.Now().Format(time.RFC3339)
	_ = p.configRepo.Set(ctx, "last_sync_at", syncedAtStr)

	log.Info().Msg("Sync pull config selesai")
	return nil
}
