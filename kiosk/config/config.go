package config

import (
	"fmt"
	"os"
)

// Config holds the kiosk local configuration.
type Config struct {
	// Kiosk identity
	DesaID    string // UUID of the desa this kiosk belongs to
	KioskName string // Display name of this kiosk

	// Network
	ListenAddr string // Local HTTP listen address (default: :8080)
	ServerURL  string // Online backend URL for sync
	APIKey     string // API key for authenticating with the server

	// Database
	DBPath string // Path to SQLite database file

	// Print
	PrintCommand string // Command to use for printing (e.g., SumatraPDF.exe path)

	// Static files
	StaticDir string // Path to kiosk UI static files

	// Sync
	SyncInterval int // Sync interval in seconds (default: 30)
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		DesaID:       getEnv("KIOSK_DESA_ID", ""),
		KioskName:    getEnv("KIOSK_NAME", "Kiosk Desa"),
		ListenAddr:   getEnv("KIOSK_LISTEN_ADDR", ":8080"),
		ServerURL:    getEnv("KIOSK_SERVER_URL", ""),
		APIKey:       getEnv("KIOSK_API_KEY", ""),
		DBPath:       getEnv("KIOSK_DB_PATH", "data/kiosk.db"),
		PrintCommand: getEnv("KIOSK_PRINT_CMD", "SumatraPDF.exe"),
		StaticDir:    getEnv("KIOSK_STATIC_DIR", "web/kiosk-ui/dist"),
		SyncInterval: getEnvInt("KIOSK_SYNC_INTERVAL", 30),
	}

	if cfg.DesaID == "" {
		return nil, fmt.Errorf("KIOSK_DESA_ID harus diisi")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	var result int
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		return fallback
	}
	return result
}
