package config

import (
	"fmt"
	"os"
	"strconv"
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

	// Static SPA (admin dashboard built by Vite). When set, the server serves
	// the built dashboard from this directory with SPA fallback to index.html.
	StaticDir string // Directory of dashboard build output (e.g. web/dashboard/dist)

	// OCR AI Providers
	GeminiAPIKey   string // Google Gemini API key
	GeminiModel    string // Gemini model name (default: gemini-1.5-flash)
	MistralAPIKey  string // Mistral AI API key
	MistralModel   string // Mistral model name (default: pixtral-12b)
	GroqAPIKey     string // Groq Cloud API key
	GroqModel      string // Groq model name (default: llama-3.2-11b-vision-preview)
	OCRStrategy    string // failover | round_robin (default: failover)
	OCRMockEnabled bool   // enable offline mock fallback (default: true)
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
		StaticDir:          getEnv("STATIC_DIR", ""),
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GeminiModel:        getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		MistralAPIKey:      getEnv("MISTRAL_API_KEY", ""),
		MistralModel:       getEnv("MISTRAL_MODEL", "pixtral-12b"),
		GroqAPIKey:         getEnv("GROQ_API_KEY", ""),
		GroqModel:          getEnv("GROQ_MODEL", "llama-3.2-11b-vision-preview"),
		OCRStrategy:        getEnv("OCR_STRATEGY", "failover"),
		OCRMockEnabled:     getEnvBool("OCR_MOCK_ENABLED", true),
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

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}
