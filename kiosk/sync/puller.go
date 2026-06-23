package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	nomorSuratRepo *db.NomorSuratRepository
	client         *http.Client
}

func NewPuller(
	cfg *config.Config,
	wargaRepo *db.WargaRepository,
	jenisSuratRepo *db.JenisSuratRepository,
	configRepo *db.ConfigRepository,
	nomorSuratRepo *db.NomorSuratRepository,
) *Puller {
	return &Puller{
		cfg:            cfg,
		wargaRepo:      wargaRepo,
		jenisSuratRepo: jenisSuratRepo,
		configRepo:     configRepo,
		nomorSuratRepo: nomorSuratRepo,
		client:         &http.Client{Timeout: 15 * time.Second},
	}
}

// PullWarga fetches new/updated residents from the server hub.
func (p *Puller) PullWarga(ctx context.Context) error {
	// 1. Get last sync timestamp (separate key per entity type)
	lastSyncStr, err := p.configRepo.Get(ctx, "last_sync_at_warga")
	if err != nil {
		lastSyncStr = ""
	}

	// URL-encode the last_sync parameter to handle + in timezone offsets
	syncURL := fmt.Sprintf("%s/api/sync/pull/warga?last_sync=%s", p.cfg.ServerURL, url.QueryEscape(lastSyncStr))
	req, err := http.NewRequestWithContext(ctx, "GET", syncURL, nil)
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

	// 3. Save last sync time for warga
	syncedAtStr := response.SyncedAt.Format(time.RFC3339)
	_ = p.configRepo.Set(ctx, "last_sync_at_warga", syncedAtStr)

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

	// 1b. Deactivate local types that the server no longer considers active for
	// this desa. The server only returns active types, so anything not in the
	// list should be hidden on the kiosk.
	activeKodes := make([]string, 0, len(response.JenisSurat))
	for _, js := range response.JenisSurat {
		activeKodes = append(activeKodes, js.Kode)
	}
	if err := p.jenisSuratRepo.DeactivateExcept(ctx, activeKodes); err != nil {
		log.Error().Err(err).Msg("Gagal deaktivasi jenis_surat lokal yang tidak aktif di server")
	}

	// 2. Sync template documents
	// The server only returns templates relevant to this kiosk (per-desa or
	// general). Align the desa_id to the kiosk's own desa_id so that
	// GetTemplate (per-desa lookup) always finds the server template. This
	// also ensures an older locally-seeded HTML template sharing the same
	// (jenis_surat_id, desa_id) pair is replaced by the server DOCX version.
	if len(response.Templates) > 0 {
		log.Info().Int("count", len(response.Templates)).Msg("Sinkronisasi template cetak dokumen...")
		for _, t := range response.Templates {
			t.DesaID = p.cfg.DesaID
			if err := p.jenisSuratRepo.UpsertTemplate(ctx, &t); err != nil {
				log.Error().Err(err).Str("template_id", t.ID).Msg("Gagal upsert template secara lokal")
			}
		}
	}

	// 3. Save last sync time for config
	syncedAtStr := time.Now().Format(time.RFC3339)
	_ = p.configRepo.Set(ctx, "last_sync_at_config", syncedAtStr)

	log.Info().Msg("Sync pull config selesai")
	return nil
}

// PullNomorSurat fetches nomor surat batch config from server.
func (p *Puller) PullNomorSurat(ctx context.Context) error {
	syncURL := fmt.Sprintf("%s/api/sync/pull/nomor-surat", p.cfg.ServerURL)
	req, err := http.NewRequestWithContext(ctx, "GET", syncURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create pull nomor-surat request: %w", err)
	}
	req.Header.Set("X-API-Key", p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pull nomor-surat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	var batches []models.NomorSuratBatch
	if err := json.NewDecoder(resp.Body).Decode(&batches); err != nil {
		return fmt.Errorf("failed to parse nomor-surat response: %w", err)
	}

	for _, batch := range batches {
		// Cek apakah batch lokal sudah ada, jika sudah, jangan turunkan nomor_terakhir
		existing, err := p.nomorSuratRepo.GetBatch(ctx, batch.JenisSuratID)
		if err == nil && existing != nil {
			// Preserve local nomor_terakhir jika lebih besar (sudah terpakai)
			if existing.NomorTerakhir > batch.NomorTerakhir {
				batch.NomorTerakhir = existing.NomorTerakhir
			}
		}
		if err := p.nomorSuratRepo.UpdateBatch(ctx, batch); err != nil {
			log.Error().Err(err).Str("jenis_surat_id", batch.JenisSuratID).Msg("Gagal update batch nomor surat")
		}
	}

	log.Info().Int("count", len(batches)).Msg("Sync pull nomor surat selesai")
	return nil
}
