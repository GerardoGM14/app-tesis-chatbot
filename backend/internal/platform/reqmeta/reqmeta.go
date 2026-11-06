// Package reqmeta guarda metadatos de la petición (IP del cliente) en el
// contexto para que la auditoría los registre sin pasarlos por parámetro.
package reqmeta

import (
	"context"
	"net"
	"net/http"
)

type ipKey struct{}

// Middleware toma la IP de r.RemoteAddr. Si el servidor está detrás de un
// proxy de confianza, chi/middleware.RealIP debe ir antes.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(WithClientIP(r.Context(), hostOnly(r.RemoteAddr))))
	})
}

// WithClientIP guarda la IP del cliente en el contexto.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey{}, ip)
}

// ClientIP devuelve la IP guardada o cadena vacía.
func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(ipKey{}).(string)
	return ip
}

func hostOnly(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}
