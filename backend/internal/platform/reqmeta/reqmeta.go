// Package reqmeta guarda metadatos de la petición (IP del cliente) en el
// contexto para que la auditoría los registre sin pasarlos por parámetro.
package reqmeta

import (
	"context"
	"net"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type ipKey struct{}

// Middleware usa la IP que dejó un middleware ClientIPFrom* de chi (cuando hay un
// proxy de confianza) y, si no hay ninguna, la de r.RemoteAddr.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := chimw.GetClientIP(r.Context())
		if ip == "" {
			ip = hostOnly(r.RemoteAddr)
		}
		next.ServeHTTP(w, r.WithContext(WithClientIP(r.Context(), ip)))
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
