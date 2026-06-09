package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// handleGetWargaByRFID handles looking up a resident by RFID UID.
func (s *Server) handleGetWargaByRFID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := chi.URLParam(r, "uid")

	if uid == "" {
		sendError(w, http.StatusBadRequest, "RFID UID tidak boleh kosong")
		return
	}

	warga, err := s.wargaRepo.FindByRFID(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "KTP dengan RFID UID tersebut belum terdaftar di sistem")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data warga: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, warga)
}

// handleGetWargaByNIK handles looking up a resident by NIK (fallback).
func (s *Server) handleGetWargaByNIK(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nik := chi.URLParam(r, "nik")

	if len(nik) != 16 {
		sendError(w, http.StatusBadRequest, "NIK harus terdiri dari 16 digit")
		return
	}

	warga, err := s.wargaRepo.FindByNIK(ctx, nik)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendError(w, http.StatusNotFound, "NIK tidak ditemukan")
			return
		}
		sendError(w, http.StatusInternalServerError, "Gagal mengambil data warga: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, warga)
}

// handleSearchWarga handles searching residents by name or NIK.
func (s *Server) handleSearchWarga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")

	if query == "" {
		sendError(w, http.StatusBadRequest, "Query pencarian (q) tidak boleh kosong")
		return
	}

	results, err := s.wargaRepo.Search(ctx, query)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal mencari data warga: "+err.Error())
		return
	}

	if len(results) == 0 {
		sendJSON(w, http.StatusOK, []interface{}{})
		return
	}

	sendJSON(w, http.StatusOK, results)
}
