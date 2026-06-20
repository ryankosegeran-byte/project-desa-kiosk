package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// spaHandler serves the Vite-built admin dashboard (SPA) from a directory.
// Real files (JS/CSS/assets) are served directly; any other path falls back to
// index.html so client-side routing (React Router) works on deep links / refresh.
func spaHandler(staticDir string) http.HandlerFunc {
	indexPath := filepath.Join(staticDir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		// Never let the SPA fallback swallow API or health routes.
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ping" {
			http.NotFound(w, r)
			return
		}

		// Clean the request path and resolve it inside staticDir.
		clean := filepath.Clean(r.URL.Path)
		reqPath := filepath.Join(staticDir, filepath.FromSlash(clean))

		// Guard against path traversal outside staticDir.
		if !strings.HasPrefix(reqPath, filepath.Clean(staticDir)) {
			http.NotFound(w, r)
			return
		}

		// Serve the static file if it exists and is not a directory.
		if info, err := os.Stat(reqPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, reqPath)
			return
		}

		// Fallback: serve index.html for client-side routes.
		http.ServeFile(w, r, indexPath)
	}
}
