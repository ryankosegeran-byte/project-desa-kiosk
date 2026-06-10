package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/project-desa-kiosk/internal/models"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, nama, role, jabatan, desa_id, active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanRow(row)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, nama, role, jabatan, desa_id, active, last_login_at, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	row := r.db.QueryRowContext(ctx, query, username)
	return r.scanRow(row)
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (
			id, username, password_hash, nama, role, jabatan, desa_id, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	var desaID interface{}
	if u.DesaID != "" {
		desaID = u.DesaID
	}

	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Username, u.PasswordHash, u.Nama, u.Role, u.Jabatan, desaID, u.Active, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal create user server: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	query := `
		UPDATE users SET
			username = $1,
			nama = $2,
			role = $3,
			jabatan = $4,
			desa_id = $5,
			active = $6,
			updated_at = $7
		WHERE id = $8
	`
	var desaID interface{}
	if u.DesaID != "" {
		desaID = u.DesaID
	}

	u.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		u.Username, u.Nama, u.Role, u.Jabatan, desaID, u.Active, u.UpdatedAt, u.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update user server: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `
		UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("gagal update password user server: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE users SET last_login_at = $1 WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("gagal update last login user server: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, role string) ([]models.User, error) {
	query := `
		SELECT id, username, password_hash, nama, role, jabatan, desa_id, active, last_login_at, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR role = $1)
		ORDER BY username ASC
	`
	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("gagal list user server: %w", err)
	}
	defer rows.Close()

	var result []models.User
	for rows.Next() {
		u, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *u)
	}
	return result, nil
}

// LogActivity saves a log item of user actions
func (r *UserRepository) LogActivity(ctx context.Context, l *models.UserActivityLog) error {
	query := `
		INSERT INTO user_activity_log (
			user_id, desa_id, action, entity_type, entity_id, detail, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var desaID, entityID, detail interface{}
	if l.DesaID != "" {
		desaID = l.DesaID
	}
	if l.EntityID != "" {
		entityID = l.EntityID
	}
	if l.Detail != "" {
		detail = l.Detail
	}

	l.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		l.UserID, desaID, l.Action, l.EntityType, entityID, detail, l.IPAddress, l.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal log activity user server: %w", err)
	}
	return nil
}

// ListActivityLogs retrieves activity logs
func (r *UserRepository) ListActivityLogs(ctx context.Context, desaID string, limit, offset int) ([]models.UserActivityLog, error) {
	query := `
		SELECT id, user_id, desa_id, action, entity_type, entity_id, detail, ip_address, created_at
		FROM user_activity_log
		WHERE ($1 = '' OR desa_id = $1::uuid)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, desaID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("gagal list activity logs server: %w", err)
	}
	defer rows.Close()

	var result []models.UserActivityLog
	for rows.Next() {
		var l models.UserActivityLog
		var dID, eID, dt, ip sql.NullString
		err := rows.Scan(
			&l.ID, &l.UserID, &dID, &l.Action, &eID, &eID, &dt, &ip, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan activity logs row: %w", err)
		}
		l.DesaID = dID.String
		l.EntityID = eID.String
		l.Detail = dt.String
		l.IPAddress = ip.String
		result = append(result, l)
	}
	return result, nil
}

func (r *UserRepository) scanRow(row *sql.Row) (*models.User, error) {
	var u models.User
	var desaID, jabatan sql.NullString
	var lastLogin sql.NullTime

	err := row.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role, &jabatan, &desaID, &u.Active, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("gagal scan user row server: %w", err)
	}

	u.DesaID = desaID.String
	u.Jabatan = jabatan.String
	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}

	return &u, nil
}

func (r *UserRepository) scanRows(rows *sql.Rows) (*models.User, error) {
	var u models.User
	var desaID, jabatan sql.NullString
	var lastLogin sql.NullTime

	err := rows.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Nama, &u.Role, &jabatan, &desaID, &u.Active, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal scan user rows server: %w", err)
	}

	u.DesaID = desaID.String
	u.Jabatan = jabatan.String
	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}

	return &u, nil
}
