package api

import (
	"database/sql"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/project-desa-kiosk/internal/models"
)

// handleLogin logs in a user and returns JWT access and refresh tokens.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.LoginRequest
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		sendError(w, http.StatusBadRequest, "Username dan Password harus diisi")
		return
	}

	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusUnauthorized, "Username atau Password salah")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal memproses login: "+err.Error())
		return
	}

	if !user.Active {
		sendError(w, http.StatusForbidden, "Akun Anda telah dinonaktifkan")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Username atau Password salah")
		return
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate access and refresh tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role, user.DesaID, user.Jabatan)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat token akses: "+err.Error())
		return
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat token refresh: "+err.Error())
		return
	}

	// Log activity
	_ = s.userRepo.LogActivity(ctx, &models.UserActivityLog{
		UserID:    user.ID,
		DesaID:    user.DesaID,
		Action:    "LOGIN",
		IPAddress: r.RemoteAddr,
	})

	sendJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	})
}

// handleRefreshToken refreshes the access token using a valid refresh token.
func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.RefreshToken == "" {
		sendError(w, http.StatusBadRequest, "Refresh token harus diisi")
		return
	}

	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Refresh token tidak valid atau kedaluwarsa")
		return
	}

	// Find user to verify they still exist and are active
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusUnauthorized, "User tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data user: "+err.Error())
		return
	}

	if !user.Active {
		sendError(w, http.StatusForbidden, "Akun Anda telah dinonaktifkan")
		return
	}

	// Generate a new access token
	newAccessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role, user.DesaID, user.Jabatan)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal membuat token akses baru: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}
