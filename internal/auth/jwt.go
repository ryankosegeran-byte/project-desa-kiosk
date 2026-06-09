package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrInvalidRole  = errors.New("insufficient permissions")
)

// Claims represents the JWT claims used in the system.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	DesaID   string `json:"desa_id,omitempty"`
	Jabatan  string `json:"jabatan,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation.
type JWTManager struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

// NewJWTManager creates a new JWTManager with the given secret and expiry durations.
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:          []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		issuer:             "kiosk-desa",
	}
}

// GenerateAccessToken creates a new access token for the given user details.
func (m *JWTManager) GenerateAccessToken(userID, username, role, desaID, jabatan string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		DesaID:   desaID,
		Jabatan:  jabatan,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken creates a new refresh token for the given user ID.
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken parses and validates a JWT token string, returning the claims.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HasRole checks if the claims contain one of the allowed roles.
func HasRole(claims *Claims, allowedRoles ...string) bool {
	for _, r := range allowedRoles {
		if claims.Role == r {
			return true
		}
	}
	return false
}
