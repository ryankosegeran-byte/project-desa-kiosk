package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// sendJSON sends a JSON response with the given status code.
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Gagal encode JSON response")
	}
}

// sendError sends a JSON error response.
func sendError(w http.ResponseWriter, status int, msg string) {
	sendJSON(w, status, map[string]string{"error": msg})
}

// parseJSON decodes the request body into target struct.
func parseJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
