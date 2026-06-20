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

	"github.com/project-desa-kiosk/kiosk/api"
	"github.com/project-desa-kiosk/kiosk/config"
	"github.com/project-desa-kiosk/kiosk/db"
	"github.com/project-desa-kiosk/kiosk/print"
	"github.com/project-desa-kiosk/kiosk/rfid"
	"github.com/project-desa-kiosk/kiosk/sync"
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

	// Graceful shutdown context (created early so it can be passed to background services)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().
		Str("desa_id", cfg.DesaID).
		Str("kiosk_name", cfg.KioskName).
		Str("listen", cfg.ListenAddr).
		Str("db_path", cfg.DBPath).
		Msg("Memulai Kiosk Desa...")

	// 1. Initialize SQLite database
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Gagal membuka database SQLite")
	}

	// 2. Instantiate Repositories
	wargaRepo := db.NewWargaRepository(database)
	suratRepo := db.NewSuratRepository(database)
	jenisSuratRepo := db.NewJenisSuratRepository(database)
	configRepo := db.NewConfigRepository(database)
	nomorSuratRepo := db.NewNomorSuratRepository(database)

	// 2.5 Deteksi perubahan desa_id — auto-cleanup data lokal
	ctxSeed := context.Background()
	prevDesaID, _ := configRepo.Get(ctxSeed, "desa_id")
	if prevDesaID != "" && prevDesaID != cfg.DesaID {
		log.Warn().
			Str("desa_lama", prevDesaID).
			Str("desa_baru", cfg.DesaID).
			Msg("Terdeteksi perubahan desa_id — menghapus semua data lokal...")

		if err := database.ResetLocalData(ctxSeed); err != nil {
			log.Fatal().Err(err).Msg("Gagal mereset data lokal")
		}
		log.Info().Msg("Data lokal berhasil dihapus. Sync akan dimulai dari nol.")
	}

	// 3. Seed Database with initial offline demo data if empty
	if err := db.SeedLocalData(database, cfg.DesaID); err != nil {
		log.Fatal().Err(err).Msg("Gagal melakukan seeding database lokal")
	}

	// 4. Save metadata settings to SQLite kiosk_config
	_ = configRepo.Set(ctxSeed, "desa_id", cfg.DesaID)
	_ = configRepo.Set(ctxSeed, "kiosk_name", cfg.KioskName)
	if cfg.ServerURL != "" {
		_ = configRepo.Set(ctxSeed, "server_url", cfg.ServerURL)
	}

	// 5. Initialize RFID Broker
	rfidBroker := rfid.NewBroker()
	rfidBroker.Start(ctx)

	// 7. Initialize PDF Generator & Printer Services
	pdfGen := print.NewPDFGenerator("data/printed")
	printer := print.NewPrinter(cfg.PrintCommand)
	docxRenderer := print.NewDocxRenderer("data/printed")
	defer docxRenderer.Close()

	// 8. Initialize Sync Engine
	syncRepo := db.NewSyncRepository(database)
	detector := sync.NewDetector(cfg.ServerURL)
	pusher := sync.NewPusher(cfg, syncRepo, suratRepo, configRepo)
	puller := sync.NewPuller(cfg, wargaRepo, jenisSuratRepo, configRepo, nomorSuratRepo)
	syncEngine := sync.NewEngine(cfg, detector, pusher, puller)
	syncEngine.Start(ctx)

	// 9. Initialize API server
	apiServer := api.NewServer(cfg, wargaRepo, suratRepo, jenisSuratRepo, configRepo, nomorSuratRepo, rfidBroker, pdfGen, printer, docxRenderer)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      apiServer.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", cfg.ListenAddr).Msg("Kiosk API server berjalan")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Info().Msg("Mematikan Kiosk server...")

	// Shutdown HTTP server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error saat shutdown HTTP server")
	}

	// Close database connection
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("Error saat menutup database SQLite")
	} else {
		log.Info().Msg("Koneksi SQLite berhasil ditutup.")
	}

	log.Info().Msg("Server Kiosk berhenti.")
}
