// Package server expone el API interno de wa-service. Solo es accesible desde
// la red interna de Docker y toda ruta /internal exige firma HMAC.
package server

import (
	"encoding/json"
	"net/http"
	"time"

	"wa-service/internal/hmacauth"
)

// New arma el router HTTP.
func New(secret []byte) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	internal := http.NewServeMux()
	// Rutas planificadas para el sprint 3 (integración con whatsmeow).
	for _, pattern := range []string{
		"POST /internal/sessions/{tenant}/connect",
		"POST /internal/sessions/{tenant}/disconnect",
		"GET /internal/sessions/{tenant}/status",
		"POST /internal/sessions/{tenant}/send",
	} {
		internal.HandleFunc(pattern, notImplemented)
	}
	mux.Handle("/internal/", hmacauth.Middleware(secret, time.Now)(internal))
	return mux
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{
		"error": map[string]string{"code": "no_implementado", "message": "Disponible en el sprint 3"},
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
