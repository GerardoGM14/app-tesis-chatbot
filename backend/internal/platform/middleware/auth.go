package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"backend/internal/platform/apperr"
	"backend/internal/platform/authz"
	"backend/internal/platform/httpx"
)

// TokenParser valida un access token y devuelve el usuario.
type TokenParser interface {
	Parse(token string) (authz.Principal, error)
}

// Authenticate exige un Bearer token válido y guarda el usuario en el contexto.
func Authenticate(tokens TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				httpx.Error(w, r, apperr.ErrUnauthorized)
				return
			}
			p, err := tokens.Parse(raw)
			if err != nil {
				httpx.Error(w, r, apperr.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(authz.WithPrincipal(r.Context(), p)))
		})
	}
}

var errPasswordChangeRequired = apperr.New(apperr.KindForbidden, "cambio_password_requerido",
	"Debes cambiar tu contraseña antes de continuar")

// RequirePasswordChanged bloquea a quien tiene pendiente el cambio obligatorio
// de contraseña (primer ingreso o contraseña restablecida).
func RequirePasswordChanged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := authz.PrincipalFrom(r.Context()); ok && p.MustChangePassword {
			httpx.Error(w, r, errPasswordChangeRequired)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission exige que el rol del usuario tenga el permiso.
func RequirePermission(perm authz.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := authz.PrincipalFrom(r.Context())
			if !ok {
				httpx.Error(w, r, apperr.ErrUnauthorized)
				return
			}
			if !p.Role.Can(perm) {
				httpx.Error(w, r, apperr.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TenantStatusChecker consulta si una empresa está activa.
type TenantStatusChecker interface {
	IsActive(ctx context.Context, id uuid.UUID) (bool, error)
}

var errTenantSuspended = apperr.New(apperr.KindForbidden, "empresa_suspendida",
	"Tu empresa está suspendida. Contacta al administrador de la plataforma")

type cachedStatus struct {
	active  bool
	expires time.Time
}

// TenantGuard rechaza peticiones de usuarios cuya empresa fue suspendida.
// El resultado se guarda en memoria por ttl para no consultar la base en cada petición.
func TenantGuard(checker TenantStatusChecker, ttl time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	cache := map[uuid.UUID]cachedStatus{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := authz.PrincipalFrom(r.Context())
			if !ok || p.TenantID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			mu.Lock()
			entry, hit := cache[p.TenantID]
			mu.Unlock()

			if !hit || time.Now().After(entry.expires) {
				active, err := checker.IsActive(r.Context(), p.TenantID)
				if err != nil {
					httpx.Error(w, r, err)
					return
				}
				entry = cachedStatus{active: active, expires: time.Now().Add(ttl)}
				mu.Lock()
				cache[p.TenantID] = entry
				mu.Unlock()
			}

			if !entry.active {
				httpx.Error(w, r, errTenantSuspended)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
