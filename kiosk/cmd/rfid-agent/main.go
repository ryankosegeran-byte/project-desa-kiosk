// Command rfid-agent is a lightweight local bridge that lets a PIC operator's
// laptop expose a PC/SC RFID reader (e.g. ACR122U) to the online admin panel
// running in the browser.
//
// The admin dashboard is served from a remote server, so the browser cannot
// talk to the USB reader directly. This agent runs locally on the PIC laptop,
// reads card UIDs from the reader and streams them over Server-Sent Events at
//
//	GET http://localhost:8088/api/rfid/events
//
// which the dashboard subscribes to (with CORS allowed for any origin).
//
// It reuses the exact same rfid package as the kiosk backend, but without any
// database or desa configuration, so it is tiny and easy to run.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/project-desa-kiosk/kiosk/rfid"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	listenAddr := getEnv("RFID_AGENT_LISTEN", ":8088")
	uidFormat := getEnv("RFID_AGENT_UID_FORMAT", "hex")
	readerFilter := getEnv("RFID_AGENT_READER", "")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().
		Str("listen", listenAddr).
		Str("uid_format", uidFormat).
		Str("reader_filter", readerFilter).
		Msg("Memulai RFID Agent (jembatan ACR122U untuk admin panel)")

	broker := rfid.NewBroker()
	broker.Start(ctx)

	rfid.StartReader(ctx, broker, rfid.ReaderConfig{
		Enabled:          true,
		ReaderNameFilter: readerFilter,
		UIDFormat:        uidFormat,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/rfid/events", rfid.ServeEvents(broker))
	mux.HandleFunc("/api/rfid/mock", rfid.HandleMockScan(broker))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"status":"ok","service":"rfid-agent","uid_format":%q}`, uidFormat)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprint(w, strings.TrimSpace(`
RFID Agent berjalan.
- SSE  : GET  /api/rfid/events
- Mock : POST /api/rfid/mock  {"uid":"04A1B2C3"}
- Health: GET /healthz
`)+"\n")
	})

	srv := &http.Server{
		Addr:        listenAddr,
		Handler:     corsMiddleware(mux),
		ReadTimeout: 0, // SSE is long-lived
		IdleTimeout: 60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", listenAddr).Msg("RFID Agent HTTP server berjalan")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Mematikan RFID Agent...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info().Msg("RFID Agent berhenti.")
}
