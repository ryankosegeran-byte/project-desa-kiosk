package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/project-desa-kiosk/internal/models"
	"github.com/project-desa-kiosk/server/middleware"
)

// resolveDesaID memilih desa yang dikelola: PIC selalu desanya sendiri;
// superadmin boleh menargetkan desa lain lewat parameter (fallback ke desa sendiri).
func resolveDesaID(role, claimDesa, requested string) string {
	if role == models.RoleSuperAdmin && requested != "" {
		return requested
	}
	return claimDesa
}

type NomorSuratConfig struct {
	ID             string    `json:"id"`
	DesaID         string    `json:"desa_id"`
	JenisSuratID   string    `json:"jenis_surat_id"`
	NomorMulai     int       `json:"nomor_mulai"`
	BatasAtas      int       `json:"batas_atas"`
	NomorTerakhir  int       `json:"nomor_terakhir"`
	FormatNomor    string    `json:"format_nomor"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// handleListNomorSuratConfig lists all nomor surat config for the user's desa.
func (s *Server) handleListNomorSuratConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "User tidak teridentifikasi")
		return
	}

	desaID := resolveDesaID(claims.Role, claims.DesaID, r.URL.Query().Get("desa_id"))
	if desaID == "" {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan")
		return
	}

	query := `SELECT id, desa_id, jenis_surat_id, nomor_mulai, batas_atas, nomor_terakhir, COALESCE(format_nomor, ''), updated_at
		FROM nomor_surat_config WHERE desa_id = $1 ORDER BY updated_at DESC`

	rows, err := s.db.QueryContext(ctx, query, desaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal query: "+err.Error())
		return
	}
	defer rows.Close()

	var configs []NomorSuratConfig
	for rows.Next() {
		var c NomorSuratConfig
		if err := rows.Scan(&c.ID, &c.DesaID, &c.JenisSuratID, &c.NomorMulai, &c.BatasAtas, &c.NomorTerakhir, &c.FormatNomor, &c.UpdatedAt); err != nil {
			sendError(w, http.StatusInternalServerError, "Gagal scan: "+err.Error())
			return
		}
		configs = append(configs, c)
	}

	if configs == nil {
		configs = []NomorSuratConfig{}
	}
	sendJSON(w, http.StatusOK, configs)
}

// handleUpdateNomorSuratConfig updates range for a specific jenis surat.
func (s *Server) handleUpdateNomorSuratConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := middleware.GetClaims(ctx)
	if claims == nil {
		sendError(w, http.StatusUnauthorized, "User tidak teridentifikasi")
		return
	}
	jenisSuratID := chi.URLParam(r, "jenis_surat_id")

	var req struct {
		DesaID      string `json:"desa_id"`
		NomorMulai  int    `json:"nomor_mulai"`
		BatasAtas   int    `json:"batas_atas"`
		FormatNomor string `json:"format_nomor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Payload tidak valid: "+err.Error())
		return
	}

	desaID := resolveDesaID(claims.Role, claims.DesaID, req.DesaID)
	if desaID == "" {
		sendError(w, http.StatusBadRequest, "desa_id diperlukan")
		return
	}

	if req.BatasAtas < req.NomorMulai {
		sendError(w, http.StatusBadRequest, "Batas atas harus >= nomor mulai")
		return
	}

	// nomor_terakhir di-set ke (nomor_mulai - 1) saat config baru agar surat pertama
	// memakai nomor_mulai. Untuk update, biarkan nomor_terakhir apa adanya.
	query := `
		INSERT INTO nomor_surat_config (desa_id, jenis_surat_id, nomor_mulai, batas_atas, nomor_terakhir, format_nomor, updated_at)
		VALUES ($1, $2, $3, $4, GREATEST($3 - 1, 0), $5, NOW())
		ON CONFLICT(desa_id, jenis_surat_id) DO UPDATE SET
		nomor_mulai = EXCLUDED.nomor_mulai,
		batas_atas = EXCLUDED.batas_atas,
		format_nomor = EXCLUDED.format_nomor,
		updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, query, desaID, jenisSuratID, req.NomorMulai, req.BatasAtas, req.FormatNomor)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal update: "+err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleSyncPullNomorSurat returns batch config for kiosk sync.
func (s *Server) handleSyncPullNomorSurat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	kiosk := middleware.GetKiosk(ctx)
	if kiosk == nil {
		sendError(w, http.StatusUnauthorized, "Kiosk tidak teridentifikasi")
		return
	}

	query := `SELECT jenis_surat_id, nomor_terakhir, batas_atas, COALESCE(format_nomor, '')
		FROM nomor_surat_config WHERE desa_id = $1`

	rows, err := s.db.QueryContext(ctx, query, kiosk.DesaID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Gagal query: "+err.Error())
		return
	}
	defer rows.Close()

	type BatchItem struct {
		JenisSuratID  string `json:"jenis_surat_id"`
		NomorTerakhir int    `json:"nomor_terakhir"`
		BatasAtas     int    `json:"batas_atas"`
		FormatNomor   string `json:"format_nomor"`
	}
	var batches []BatchItem
	for rows.Next() {
		var b BatchItem
		if err := rows.Scan(&b.JenisSuratID, &b.NomorTerakhir, &b.BatasAtas, &b.FormatNomor); err != nil {
			continue
		}
		batches = append(batches, b)
	}

	if batches == nil {
		batches = []BatchItem{}
	}
	sendJSON(w, http.StatusOK, batches)
}
