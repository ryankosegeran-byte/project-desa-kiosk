package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// handleListUsers lists all users (Superadmin only).
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := r.URL.Query().Get("role")

	list, err := s.userRepo.List(ctx, role)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data user: "+err.Error())
		return
	}

	if len(list) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, list)
}

// handleCreateUser registers a new user (Superadmin only).
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateUserRequest
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if req.Username == "" || req.Password == "" || req.Nama == "" || req.Role == "" {
		sendError(w, http.StatusBadRequest, "Username, Password, Nama, dan Role harus diisi")
		return
	}

	// Verify duplicate username
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		sendError(w, http.StatusConflict, "Username sudah digunakan")
		return
	}

	// Hash password
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal hash password: "+err.Error())
		return
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: string(pwdHash),
		Nama:         req.Nama,
		Role:         req.Role,
		Jabatan:      req.Jabatan,
		DesaID:       req.DesaID,
		Active:       true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal menyimpan user: "+err.Error())
		return
	}

	sendJSON(w, http.StatusCreated, user)
}

// handleUpdateUser edits user details or toggles active status (Superadmin only).
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID tidak boleh kosong")
		return
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "User tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mencari user: "+err.Error())
		return
	}

	var req models.User
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	user.Username = req.Username
	user.Nama = req.Nama
	user.Role = req.Role
	user.Jabatan = req.Jabatan
	user.DesaID = req.DesaID
	user.Active = req.Active

	if err := s.userRepo.Update(ctx, user); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memperbarui user: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, user)
}

// handleResetPassword changes a user's password (Superadmin only).
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		sendError(w, http.StatusBadRequest, "ID tidak boleh kosong")
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := parseJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
		return
	}

	if len(req.NewPassword) < 6 {
		sendError(w, http.StatusBadRequest, "Password baru minimal 6 karakter")
		return
	}

	// Hash password
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal hash password: "+err.Error())
		return
	}

	if err := s.userRepo.UpdatePassword(ctx, id, string(pwdHash)); err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal memperbarui password: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleListActivityLogs lists the global activity logs (Superadmin only).
func (s *Server) handleListActivityLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	desaID := r.URL.Query().Get("desa_id")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	logs, err := s.userRepo.ListActivityLogs(ctx, desaID, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil activity log: "+err.Error())
		return
	}

	if len(logs) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, logs)
}

// handleListMyActivityLogs lists the activity logs of the PIC's village.
func (s *Server) handleListMyActivityLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "Token otorisasi diperlukan")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	logs, err := s.userRepo.ListActivityLogs(ctx, claims.DesaID, limit, offset)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data log aktivitas: "+err.Error())
		return
	}

	if len(logs) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, logs)
}
