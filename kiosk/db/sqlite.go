package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB wrapper around sql.DB
type DB struct {
	*sql.DB
}

// Open opens a connection to the SQLite database and runs migrations.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori database: %w", err)
	}

	// Open database connection using modernc.org/sqlite driver
	// Disable locking/journal synchronization features for performance if needed,
	// but for an offline-first kiosk, WAL mode is recommended.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka sqlite: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("gagal ping sqlite: %w", err)
	}

	wrapper := &DB{DB: db}

	// Run migrations
	if err := wrapper.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("gagal menjalankan migrasi: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("Koneksi SQLite berhasil dibuka dan migrasi selesai")
	return wrapper, nil
}

// runMigrations executes the SQL scripts inside migrations/ directory.
func (db *DB) runMigrations() error {
	// Read migration files
	files, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("gagal membaca folder migrasi: %w", err)
	}

	// For simple single file migrator, we just execute files sequentially.
	// Since we use IF NOT EXISTS, running them every time is safe.
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		log.Info().Str("file", file.Name()).Msg("Menjalankan migrasi database...")
		content, err := migrationFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("gagal membaca file migrasi %s: %w", file.Name(), err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			// SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN.
			// Skip "duplicate column" errors so re-running migrations is safe.
			if isDuplicateColumnErr(err) {
				log.Warn().Str("file", file.Name()).Msg("Migrasi dilewati: kolom sudah ada")
				continue
			}
			return fmt.Errorf("gagal eksekusi migrasi %s: %w", file.Name(), err)
		}
	}

	return nil
}

// isDuplicateColumnErr checks if the error is a SQLite "duplicate column name" error.
func isDuplicateColumnErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}

// ResetLocalData menghapus seluruh data lokal dan mereset timestamp sync.
// Dipanggil ketika desa_id berubah agar kiosk mulai dari nol untuk desa baru.
func (db *DB) ResetLocalData(ctx context.Context) error {
	tables := []string{"warga", "jenis_surat", "surat_template", "surat", "sync_queue", "activity_log"}
	for _, t := range tables {
		if _, err := db.ExecContext(ctx, "DELETE FROM "+t); err != nil {
			return fmt.Errorf("gagal hapus tabel %s: %w", t, err)
		}
	}
	// Hapus timestamp sync dari kiosk_config
	syncKeys := []string{"last_sync_at_warga", "last_sync_at_config"}
	for _, key := range syncKeys {
		db.ExecContext(ctx, "DELETE FROM kiosk_config WHERE key = ?", key)
	}
	return nil
}
