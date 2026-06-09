package config

import (
	"fmt"
	"os"
)

// Config holds the online server configuration.
type Config struct {
	// Network
	ListenAddr string // HTTP listen address (default: :3000)

	// Database
	DatabaseURL string // PostgreSQL connection string

	// JWT
	JWTSecret          string // Secret key for JWT signing
	AccessTokenExpiry  int    // Access token expiry in minutes (default: 60)
	RefreshTokenExpiry int    // Refresh token expiry in hours (default: 168 = 7 days)

	// CORS
	AllowedOrigins string // Comma-separated list of allowed CORS origins

	// File storage
	UploadDir string // Directory for uploaded files (KTP photos, logos)
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:         getEnv("SERVER_LISTEN_ADDR", ":3000"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		AccessTokenExpiry:  getEnvInt("JWT_ACCESS_EXPIRY_MIN", 60),
		RefreshTokenExpiry: getEnvInt("JWT_REFRESH_EXPIRY_HR", 168),
		AllowedOrigins:     getEnv("CORS_ORIGINS", "http://localhost:4321"),
		UploadDir:          getEnv("UPLOAD_DIR", "uploads"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL harus diisi")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET harus diisi")
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
