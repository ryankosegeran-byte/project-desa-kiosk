package api

import "net/http"

// handleOCRStatus returns the current OCR provider configuration and failover strategy.
func (s *Server) handleOCRStatus(w http.ResponseWriter, r *http.Request) {
	strategy, providers := s.ocrService.GetStatus()

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"strategy":  strategy,
		"providers": providers,
	})
}
