package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/project-desa-kiosk/internal/auth"
	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/db"
)

type contextKey string

const (
	ClaimsKey contextKey = "claims"
	KioskKey  contextKey = "kiosk"
)

// AuthMiddleware extracts the Bearer token and validates it using JWTManager.
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Authorization header diperlukan"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"Format token Authorization tidak valid, gunakan Bearer <token>"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"Token tidak valid atau kedaluwarsa"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleMiddleware checks if the claims in context contain one of the allowed roles.
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
			if !ok || claims == nil {
				http.Error(w, `{"error":"Unauthenticated"}`, http.StatusUnauthorized)
				return
			}

			if !auth.HasRole(claims, allowedRoles...) {
				http.Error(w, `{"error":"Akses ditolak: hak akses tidak mencukupi"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// KioskKeyMiddleware validates the kiosk by checking the X-API-Key header.
func KioskKeyMiddleware(desaRepo *db.DesaRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, `{"error":"Header X-API-Key diperlukan"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			kiosk, err := desaRepo.FindKioskByKey(ctx, apiKey)
			if err != nil {
				http.Error(w, `{"error":"API Key Kiosk tidak valid"}`, http.StatusUnauthorized)
				return
			}

			if kiosk.Status != "active" {
				http.Error(w, `{"error":"Kiosk dinonaktifkan"}`, http.StatusForbidden)
				return
			}

			// Save kiosk info in context
			ctx = context.WithValue(ctx, KioskKey, kiosk)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves claims from context.
func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims
}

// GetKiosk retrieves kiosk details from context.
func GetKiosk(ctx context.Context) *models.Kiosk {
	kiosk, _ := ctx.Value(KioskKey).(*models.Kiosk)
	return kiosk
}
