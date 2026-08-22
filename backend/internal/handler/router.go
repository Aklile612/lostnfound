package handler

import (
	"net/http"
	"os"
	"strings"
)

func Router(h *Handler, origin, uploadDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/meta", h.Meta)
	mux.HandleFunc("POST /api/reports/lost", h.CreateLost)
	mux.HandleFunc("POST /api/reports/found", h.CreateFound)
	mux.HandleFunc("GET /api/reports/{id}", h.GetReport)
	mux.HandleFunc("POST /api/reports/{id}/matches", h.RefreshMatches)
	mux.HandleFunc("POST /api/reports/{id}/claim", h.Claim)
	fs := http.FileServer(http.Dir(uploadDir))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", fs))
	return withCORS(mux, origin)
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func EnsureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func NormalizeOrigin(origin string) string {
	return strings.TrimRight(origin, "/")
}
