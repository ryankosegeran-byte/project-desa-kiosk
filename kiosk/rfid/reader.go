package rfid

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ReaderConfig configures the physical PC/SC RFID reader (e.g. ACR122U).
type ReaderConfig struct {
	// Enabled toggles the physical reader polling loop.
	Enabled bool
	// ReaderNameFilter, when non-empty, only attaches to readers whose name
	// contains this substring (case-insensitive). Empty means: use the first
	// available reader.
	ReaderNameFilter string
	// UIDFormat controls how the card UID is rendered before publishing.
	// One of: "hex" (upper, no separator, default), "hex-colon", "decimal".
	UIDFormat string
	// PollInterval is how often readers are (re)scanned when none is present.
	PollInterval time.Duration
}

// StartReader launches the physical PC/SC reader loop in a background goroutine.
// Safe to call on any platform: on unsupported builds it is a no-op.
// Detected card UIDs are published to the broker, reusing the same SSE pipeline
// as mock scans and keyboard-wedge readers.
func StartReader(ctx context.Context, broker *Broker, cfg ReaderConfig) {
	if !cfg.Enabled {
		log.Info().Msg("RFID PC/SC reader dinonaktifkan (KIOSK_RFID_PCSC_ENABLED=false)")
		return
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}

	go func() {
		log.Info().
			Str("uid_format", cfg.UIDFormat).
			Str("reader_filter", cfg.ReaderNameFilter).
			Msg("Memulai loop pembaca RFID PC/SC (ACR122U)")

		if err := runReaderLoop(ctx, broker, cfg); err != nil {
			log.Error().Err(err).Msg("Loop pembaca RFID PC/SC berhenti")
		}
	}()
}

// publishUID formats and publishes a raw UID, debouncing duplicates that arrive
// within a short window (same card resting on the reader).
func publishUID(broker *Broker, raw []byte, format string, lastUID *string, lastAt *time.Time) {
	if len(raw) == 0 {
		return
	}
	uid := formatUID(raw, format)
	if uid == "" {
		return
	}

	now := time.Now()
	if uid == *lastUID && now.Sub(*lastAt) < 1500*time.Millisecond {
		*lastAt = now
		return
	}
	*lastUID = uid
	*lastAt = now

	log.Info().Str("uid", uid).Msg("Kartu RFID terbaca dari pembaca fisik (PC/SC)")
	broker.Publish(uid)
}

// formatUID renders a UID byte slice according to the configured format.
func formatUID(raw []byte, format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "hex-colon":
		return hexColon(raw)
	case "decimal":
		return decimalLE(raw)
	default: // "hex" / empty
		return hexUpper(raw)
	}
}

const hexDigits = "0123456789ABCDEF"

func hexUpper(raw []byte) string {
	var b strings.Builder
	b.Grow(len(raw) * 2)
	for _, c := range raw {
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&0x0F])
	}
	return b.String()
}

func hexColon(raw []byte) string {
	var b strings.Builder
	b.Grow(len(raw) * 3)
	for i, c := range raw {
		if i > 0 {
			b.WriteByte(':')
		}
		b.WriteByte(hexDigits[c>>4])
		b.WriteByte(hexDigits[c&0x0F])
	}
	return b.String()
}

// decimalLE interprets the UID as a little-endian unsigned integer and returns
// its decimal string (common on Wiegand/HID-style mappings).
func decimalLE(raw []byte) string {
	var n uint64
	for i := len(raw) - 1; i >= 0; i-- {
		n = (n << 8) | uint64(raw[i])
	}
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
