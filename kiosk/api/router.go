package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/kiosk/config"
	"github.com/project-desa-kiosk/kiosk/db"
	"github.com/project-desa-kiosk/kiosk/print"
	"github.com/project-desa-kiosk/kiosk/rfid"
)

// Server represents the HTTP server for the kiosk API.
type Server struct {
	cfg            *config.Config
	wargaRepo      *db.WargaRepository
	suratRepo      *db.SuratRepository
	jenisSuratRepo *db.JenisSuratRepository
	configRepo     *db.ConfigRepository
	nomorSuratRepo *db.NomorSuratRepository // Add this
	rfidBroker     *rfid.Broker
	pdfGen         *print.PDFGenerator
	printer        *print.Printer
}

// NewServer creates a new API server instance.
func NewServer(
	cfg *config.Config,
	wargaRepo *db.WargaRepository,
	suratRepo *db.SuratRepository,
	jenisSuratRepo *db.JenisSuratRepository,
	configRepo *db.ConfigRepository,
	nomorSuratRepo *db.NomorSuratRepository, // Add this
	rfidBroker *rfid.Broker,
	pdfGen *print.PDFGenerator,
	printer *print.Printer,
) *Server {
	return &Server{
		cfg:            cfg,
		wargaRepo:      wargaRepo,
		suratRepo:      suratRepo,
		jenisSuratRepo: jenisSuratRepo,
		configRepo:     configRepo,
		nomorSuratRepo: nomorSuratRepo, // Add this
		rfidBroker:     rfidBroker,
		pdfGen:         pdfGen,
		printer:        printer,
	}
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	// Base Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LoggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(CORS)

	// API Routes
	r.Route("/api", func(r chi.Router) {
		// Health / Status
		r.Get("/status", s.handleStatus)

		// RFID Routes
		r.Get("/rfid/events", rfid.ServeEvents(s.rfidBroker))
		r.Post("/rfid/mock", rfid.HandleMockScan(s.rfidBroker))

		// Warga Routes
		r.Route("/warga", func(r chi.Router) {
			r.Get("/rfid/{uid}", s.handleGetWargaByRFID)
			r.Get("/nik/{nik}", s.handleGetWargaByNIK)
			r.Get("/search", s.handleSearchWarga)
		})

		// Jenis Surat Routes
		r.Route("/jenis-surat", func(r chi.Router) {
			r.Get("/", s.handleListJenisSurat)
			r.Get("/{id}/schema", s.handleGetJenisSuratSchema)
		})

		// Surat Routes
		r.Route("/surat", func(r chi.Router) {
			r.Post("/", s.handleCreateSurat)
			r.Get("/", s.handleListTodaySurat)
			r.Get("/{id}", s.handleGetSurat)
			r.Post("/{id}/print", s.handlePrintSurat)
		})

		// Template Routes
		r.Get("/template/{jenis_surat_id}", s.handleGetTemplate)

		// Nomor Surat Routes
		r.Get("/nomor-surat/status", s.handleNomorSuratStatus)
	})

	// Serve Static Files for Kiosk UI
	setupStaticFileServer(r, s.cfg.StaticDir)

	return r
}

// LoggerMiddleware is a simple structured logger using zerolog.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Int("size", ww.BytesWritten()).
			Dur("duration", time.Since(start)).
			Msg("HTTP Request")
	})
}

// CORS middleware configures cross-origin resource sharing.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupStaticFileServer serves the compiled React Kiosk UI files.
func setupStaticFileServer(r chi.Router, staticDir string) {
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, staticDir)

	// Check if static folder exists, if not warn log
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		log.Warn().Str("dir", filesDir).Msg("Static directory for Kiosk UI does not exist yet. Handlers will return 404 for front-end assets.")
	}

	fs := http.FileServer(http.Dir(filesDir))
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If requesting file that doesn't exist, serve index.html (for React Router SPA)
		path := filepath.Join(filesDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(filesDir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))
}

// Helper: send JSON response
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Gagal encode JSON response")
	}
}

// Helper: send JSON error response
func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, map[string]string{"error": msg})
}

// handleStatus returns the current status of the kiosk.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get last sync time from SQLite
	lastSync, err := s.configRepo.Get(ctx, "last_sync_at")
	if err != nil {
		lastSync = "Never"
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "online", // Kiosk backend is running. Online status to server is computed in UI
		"desa_id":    s.cfg.DesaID,
		"kiosk_name": s.cfg.KioskName,
		"last_sync":  lastSync,
		"version":    "0.1.0",
		"time":       time.Now().Format(time.RFC3339),
	})
}
