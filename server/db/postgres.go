package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// DB is a wrapper around sql.DB for PostgreSQL.
type DB struct {
	*sql.DB
}

// Open opens a connection to PostgreSQL and runs schema migrations.
func Open(databaseURL string) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	wrapper := &DB{DB: db}

	if err := wrapper.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Koneksi PostgreSQL berhasil dibuka dan migrasi selesai")
	return wrapper, nil
}

// runMigrations runs all migration scripts in migrations/ sequentially.
func (db *DB) runMigrations() error {
	files, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		log.Info().Str("file", file.Name()).Msg("Menjalankan migrasi database server...")
		content, err := migrationFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}
