package models

import "time"

// User represents a system user (superadmin, pic_desa, or kiosk).
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"` // never expose in JSON
	Nama         string     `json:"nama"`
	Role         string     `json:"role"`              // superadmin, pic_desa, kiosk
	Jabatan      string     `json:"jabatan,omitempty"` // custom title: Sekretaris Desa, Operator, etc.
	DesaID       string     `json:"desa_id,omitempty"` // NULL for superadmin
	Active       bool       `json:"active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// User role constants.
const (
	RoleSuperAdmin = "superadmin"
	RolePICDesa    = "pic_desa"
	RoleKiosk      = "kiosk"
)

// UserActivityLog records actions performed by users for monitoring.
type UserActivityLog struct {
	ID         int64     `json:"id"`
	UserID     string    `json:"user_id"`
	DesaID     string    `json:"desa_id,omitempty"`
	Action     string    `json:"action"`      // LOGIN, REGISTER_WARGA, OCR_KTP, LINK_RFID, VIEW_SURAT, etc.
	EntityType string    `json:"entity_type,omitempty"` // warga, surat, etc.
	EntityID   string    `json:"entity_id,omitempty"`
	Detail     string    `json:"detail,omitempty"` // JSON string
	IPAddress  string    `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse contains the JWT tokens.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// CreateUserRequest is the payload for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nama     string `json:"nama"`
	Role     string `json:"role"`
	Jabatan  string `json:"jabatan,omitempty"`
	DesaID   string `json:"desa_id,omitempty"`
}
