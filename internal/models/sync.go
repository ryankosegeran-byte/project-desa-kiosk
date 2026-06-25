package models

import (
	"encoding/json"
	"time"
)

// SyncPushPayload is sent from kiosk to server — contains new surat records.
type SyncPushPayload struct {
	DesaID  string          `json:"desa_id"`
	KioskID string          `json:"kiosk_id"`
	Items   []SyncPushItem  `json:"items"`
}

// SyncPushItem represents a single entity to push to the server.
type SyncPushItem struct {
	EntityType string          `json:"entity_type"` // "surat"
	EntityID   string          `json:"entity_id"`
	Operation  string          `json:"operation"` // CREATE, UPDATE
	Payload    json.RawMessage `json:"payload"`
}

// SyncPushResponse is the server's response to a push request.
type SyncPushResponse struct {
	Accepted int      `json:"accepted"`
	Rejected int      `json:"rejected"`
	Errors   []string `json:"errors,omitempty"`
}

// SyncPullWargaResponse is the server's response with updated warga data.
type SyncPullWargaResponse struct {
	Warga     []Warga   `json:"warga"`
	SyncedAt  time.Time `json:"synced_at"`
	HasMore   bool      `json:"has_more"`
}

// SyncPullConfigResponse is the server's response with config data.
type SyncPullConfigResponse struct {
	JenisSurat []JenisSurat    `json:"jenis_surat"`
	Templates  []SuratTemplate `json:"templates"`
	Theme      string          `json:"theme"`
	SyncedAt   time.Time       `json:"synced_at"`
}

// SyncQueueItem represents a pending sync item in the kiosk's local queue.
type SyncQueueItem struct {
	ID          int64           `json:"id"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Operation   string          `json:"operation"`
	Payload     json.RawMessage `json:"payload"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	LastError   string          `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
}

// KioskStatus represents the current status of a kiosk.
type KioskStatus struct {
	DesaID       string     `json:"desa_id"`
	KioskName    string     `json:"kiosk_name"`
	IsOnline     bool       `json:"is_online"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
	PendingSync  int        `json:"pending_sync"`  // items in sync_queue
	PrinterReady bool       `json:"printer_ready"`
}

// Kiosk represents a registered kiosk device in the server.
type Kiosk struct {
	ID         string     `json:"id"`
	DesaID     string     `json:"desa_id"`
	Nama       string     `json:"nama,omitempty"`
	APIKey     string     `json:"api_key,omitempty"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	Status     string     `json:"status"`
	IPAddress  string     `json:"ip_address,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// OCRProvider represents an AI OCR provider configuration.
type OCRProvider struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`     // mistral, gemini, groq
	DisplayName string    `json:"display_name"`
	APIKeyEnc   string    `json:"-"`            // encrypted, never expose
	ModelName   string    `json:"model_name,omitempty"`
	Priority    int       `json:"priority"`
	Aktif       bool      `json:"aktif"`
	Strategy    string    `json:"strategy"` // failover, round_robin
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
