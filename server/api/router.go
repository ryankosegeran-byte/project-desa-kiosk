package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/internal/auth"
	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/config"
	"github.com/project-desa-kiosk/server/db"
	serverMiddleware "github.com/project-desa-kiosk/server/middleware"
	"github.com/project-desa-kiosk/server/ocr"
)

// Server represents the online hub API server.
type Server struct {
	cfg            *config.Config
	db             *db.DB
	wargaRepo      *db.WargaRepository
	suratRepo      *db.SuratRepository
	userRepo       *db.UserRepository
	jenisSuratRepo *db.JenisSuratRepository
	templateRepo   *db.TemplateRepository
	desaRepo       *db.DesaRepository
	jwtManager     *auth.JWTManager
	ocrService     *ocr.Service
}

// NewServer creates a new Server instance.
func NewServer(
	cfg *config.Config,
	database *db.DB,
	wargaRepo *db.WargaRepository,
	suratRepo *db.SuratRepository,
	userRepo *db.UserRepository,
	jenisSuratRepo *db.JenisSuratRepository,
	templateRepo *db.TemplateRepository,
	desaRepo *db.DesaRepository,
	jwtManager *auth.JWTManager,
	ocrService *ocr.Service,
) *Server {
	return &Server{
		cfg:            cfg,
		db:             database,
		wargaRepo:      wargaRepo,
		suratRepo:      suratRepo,
		userRepo:       userRepo,
		jenisSuratRepo: jenisSuratRepo,
		templateRepo:   templateRepo,
		desaRepo:       desaRepo,
		jwtManager:     jwtManager,
		ocrService:     ocrService,
	}
}

// Handler returns the HTTP handler with all routing configurations.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LoggerMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(serverMiddleware.CORS(s.cfg.AllowedOrigins))

	// Public routes
	r.Post("/api/auth/login", s.handleLogin)
	r.Post("/api/auth/refresh", s.handleRefreshToken)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","version":"0.1.0"}`))
	})

	// Kiosk Sync endpoints (Authenticated via API Key)
	r.Group(func(r chi.Router) {
		r.Use(serverMiddleware.KioskKeyMiddleware(s.desaRepo))
		r.Post("/api/sync/push", s.handleSyncPush)
		r.Get("/api/sync/pull/warga", s.handleSyncPullWarga)
		r.Get("/api/sync/pull/config", s.handleSyncPullConfig)
	})

	// Dashboard User endpoints (Authenticated via JWT)
	r.Group(func(r chi.Router) {
		r.Use(serverMiddleware.AuthMiddleware(s.jwtManager))

		// Warga actions
		r.Route("/api/warga", func(r chi.Router) {
			r.Get("/", s.handleListWarga)
			r.Post("/", s.handleCreateWarga)
			r.Put("/{id}", s.handleUpdateWarga)
			r.Put("/{id}/rfid", s.handleLinkRFID)

			// Draft endpoints
			r.Post("/draft", s.handleCreateDraft)
			r.Get("/draft/{token}", s.handleGetDraft)
			r.Put("/draft/{token}/complete", s.handleCompleteDraft)
		})

		// OCR triggers
		r.Post("/api/ocr/ktp", s.handleOCRExtract)
		r.Get("/api/ocr/status", s.handleOCRStatus)
		r.Post("/api/ocr/test", s.handleOCRTest)

		// Synced letters logs
		r.Route("/api/surat", func(r chi.Router) {
			r.Get("/", s.handleListSurat)
			r.Get("/{id}", s.handleGetSurat)
		})

		// Custom HTML templates
		r.Route("/api/templates", func(r chi.Router) {
			r.Get("/", s.handleListTemplates)
			r.Post("/", s.handleUpsertTemplate)
		})

		// Kiosks monitoring and stats
		r.Get("/api/kiosks", s.handleListKiosks)
		r.Post("/api/kiosks", s.handleRegisterKiosk)
		r.Get("/api/dashboard/stats", s.handleGetStats)
		r.Get("/api/activity-log/desa/my", s.handleListMyActivityLogs)

		// Superadmin actions
		r.Group(func(r chi.Router) {
			r.Use(serverMiddleware.RoleMiddleware(models.RoleSuperAdmin))

			// Villages profiles and matrixes
			r.Route("/api/desa", func(r chi.Router) {
				r.Get("/", s.handleListDesa)
				r.Post("/", s.handleCreateDesa)
				r.Put("/{id}/jenis-surat", s.handleToggleDesaJenisSurat)
			})

			// User profiles
			r.Route("/api/users", func(r chi.Router) {
				r.Get("/", s.handleListUsers)
				r.Post("/", s.handleCreateUser)
				r.Put("/{id}", s.handleUpdateUser)
				r.Put("/{id}/reset-password", s.handleResetPassword)
			})

			// Global letter schemas
			r.Route("/api/jenis-surat", func(r chi.Router) {
				r.Get("/", s.handleListJenisSurat)
				r.Post("/", s.handleCreateJenisSurat)
				r.Put("/{id}", s.handleUpdateJenisSurat)
			})

			// Global audit trails
			r.Get("/api/activity-log", s.handleListActivityLogs)
		})
	})

	return r
}

// LoggerMiddleware is a logging middleware utilizing structured logging.
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
			Msg("Server HTTP Request")
	})
}
