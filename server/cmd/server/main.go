package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/auth"
	"github.com/project-desa-kiosk/server/api"
	"github.com/project-desa-kiosk/server/config"
	"github.com/project-desa-kiosk/server/db"
	"github.com/project-desa-kiosk/server/ocr"
)

func main() {
	// Setup structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	// Load .env file if exists
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Gagal memuat konfigurasi")
	}

	log.Info().
		Str("listen", cfg.ListenAddr).
		Msg("Memulai Server Online Kiosk Desa...")

	// 1. Initialize PostgreSQL database
	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Gagal membuka database PostgreSQL")
	}

	// 2. Instantiate Repositories
	wargaRepo := db.NewWargaRepository(database)
	suratRepo := db.NewSuratRepository(database)
	userRepo := db.NewUserRepository(database)
	jenisSuratRepo := db.NewJenisSuratRepository(database)
	templateRepo := db.NewTemplateRepository(database)
	desaRepo := db.NewDesaRepository(database)

	// 3. Seed Database with default records if empty
	if err := db.SeedServerData(database); err != nil {
		log.Fatal().Err(err).Msg("Gagal melakukan seeding database PostgreSQL")
	}

	// 4. Initialize JWT Manager
	accessExpiry := time.Duration(cfg.AccessTokenExpiry) * time.Minute
	refreshExpiry := time.Duration(cfg.RefreshTokenExpiry) * time.Hour
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, accessExpiry, refreshExpiry)

	// 5. Initialize AI OCR Service
	var ocrProviders []ocr.OCRProvider
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		ocrProviders = append(ocrProviders, ocr.NewGeminiProvider(key, os.Getenv("GEMINI_MODEL")))
	}
	if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
		ocrProviders = append(ocrProviders, ocr.NewMistralProvider(key, os.Getenv("MISTRAL_MODEL")))
	}
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		ocrProviders = append(ocrProviders, ocr.NewGroqProvider(key, os.Getenv("GROQ_MODEL")))
	}
	ocrService := ocr.NewService(ocrProviders, "failover")

	// 6. Initialize API server
	apiServer := api.NewServer(
		cfg,
		database,
		wargaRepo,
		suratRepo,
		userRepo,
		jenisSuratRepo,
		templateRepo,
		desaRepo,
		jwtManager,
		ocrService,
	)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      apiServer.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Str("addr", cfg.ListenAddr).Msg("Online server berjalan")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Mematikan server...")

	// Shutdown HTTP server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error saat shutdown server")
	}

	// Close database connection
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("Error saat menutup database PostgreSQL")
	} else {
		log.Info().Msg("Koneksi PostgreSQL berhasil ditutup.")
	}

	log.Info().Msg("Server berhenti.")
}
