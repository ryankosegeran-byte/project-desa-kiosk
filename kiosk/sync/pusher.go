package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/kiosk/config"
	"github.com/project-desa-kiosk/kiosk/db"
)

type Pusher struct {
	cfg        *config.Config
	syncRepo   *db.SyncRepository
	suratRepo  *db.SuratRepository
	configRepo *db.ConfigRepository
	client     *http.Client
}

func NewPusher(
	cfg *config.Config,
	syncRepo *db.SyncRepository,
	suratRepo *db.SuratRepository,
	configRepo *db.ConfigRepository,
) *Pusher {
	return &Pusher{
		cfg:        cfg,
		syncRepo:   syncRepo,
		suratRepo:  suratRepo,
		configRepo: configRepo,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *Pusher) Push(ctx context.Context) error {
	// 1. Fetch pending sync items from local SQLite queue
	items, err := p.syncRepo.ListPendingSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch pending sync items: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	log.Info().Int("count", len(items)).Msg("Mengirim data transaksi lokal ke server hub...")

	// 2. Resolve Kiosk ID (generate one if not already saved in SQLite)
	kioskID, err := p.configRepo.Get(ctx, "kiosk_id")
	if err != nil || kioskID == "" {
		kioskID = uuid.New().String()
		_ = p.configRepo.Set(ctx, "kiosk_id", kioskID)
	}

	// 3. Construct push items
	var pushItems []models.SyncPushItem
	for _, it := range items {
		pushItems = append(pushItems, models.SyncPushItem{
			EntityType: it.EntityType,
			EntityID:   it.EntityID,
			Operation:  it.Operation,
			Payload:    it.Payload,
		})
	}

	payload := models.SyncPushPayload{
		DesaID:  p.cfg.DesaID,
		KioskID: kioskID,
		Items:   pushItems,
	}

	// 4. Send request to server
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal push payload: %w", err)
	}

	url := p.cfg.ServerURL + "/api/sync/push"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.cfg.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("push request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		errMsg := fmt.Sprintf("server returned status %d: %v", resp.StatusCode, errData)
		for _, it := range items {
			_ = p.syncRepo.IncrementAttempts(ctx, it.ID, errMsg)
		}
		return errors.New(errMsg)
	}

	// 5. Parse response
	var response models.SyncPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode push response: %w", err)
	}

	// 6. Process results
	// For simplicity, we assume if StatusOK is returned and no errors are specified, all accepted.
	// In case of individual item failures in server, it reports them.
	// But if success, mark all processed:
	for _, it := range items {
		// Mark processed in sync queue
		_ = p.syncRepo.MarkProcessed(ctx, it.ID)
		
		// Mark synced in local surat transaction table
		if it.EntityType == "surat" {
			_ = p.suratRepo.MarkSynced(ctx, it.EntityID)
		}
	}

	log.Info().Int("accepted", response.Accepted).Msg("Sync push transaksi berhasil diselesaikan")
	return nil
}
