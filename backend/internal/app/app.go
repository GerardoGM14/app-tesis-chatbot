// Package app es la raíz de composición: crea los módulos, conecta sus
// dependencias y arma el router HTTP.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	authhandler "backend/internal/modules/auth/handler"
	authrepo "backend/internal/modules/auth/repository"
	authservice "backend/internal/modules/auth/service"
	tenanthandler "backend/internal/modules/tenants/handler"
	tenantrepo "backend/internal/modules/tenants/repository"
	tenantservice "backend/internal/modules/tenants/service"
	userhandler "backend/internal/modules/users/handler"
	userrepo "backend/internal/modules/users/repository"
	userservice "backend/internal/modules/users/service"
	"backend/internal/platform/apperr"
	"backend/internal/platform/audit"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/httpx"
	mw "backend/internal/platform/middleware"
	"backend/internal/platform/reqmeta"
	"backend/internal/platform/security"
)

// Pinger verifica la disponibilidad de la base de datos.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB es lo que la aplicación necesita de la base de datos.
type DB interface {
	database.TxRunner
	Pinger
}

// Options permite ajustar dependencias en pruebas.
type Options struct {
	Argon2 security.Argon2Params
	// TenantCacheTTL es cuánto se recuerda el estado de una empresa. Por defecto 30s.
	TenantCacheTTL time.Duration
	// DisableRateLimit desactiva el límite global (las pruebas hacen muchas peticiones).
	DisableRateLimit bool
}

// New arma el router con todos los módulos.
func New(cfg config.Config, db DB, logger *slog.Logger, opts Options) (http.Handler, error) {
	if opts.Argon2 == (security.Argon2Params{}) {
		opts.Argon2 = security.DefaultArgon2
	}
	if opts.TenantCacheTTL == 0 {
		opts.TenantCacheTTL = 30 * time.Second
	}
	hasher := security.NewPasswordHasher(opts.Argon2)
	tokens := security.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	auditor := audit.Recorder{}

	authRepository := authrepo.Postgres{}
	authSvc, err := authservice.New(db, authRepository, hasher, tokens, auditor, cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	userSvc := userservice.New(db, userrepo.Postgres{}, hasher, authRepository, auditor)
	tenantSvc := tenantservice.New(db, tenantrepo.Postgres{}, userSvc, authRepository, auditor)

	authenticate := mw.Authenticate(tokens)
	loginLimiter := mw.NewRateLimiter(10, 10)
	globalLimiter := mw.NewRateLimiter(300, 60)

	r := chi.NewRouter()
	if cfg.TrustProxy {
		r.Use(chimw.RealIP)
	}
	r.Use(chimw.RequestID, reqmeta.Middleware, mw.RequestLogger(logger), chimw.Recoverer, mw.SecureHeaders)
	if !opts.DisableRateLimit {
		r.Use(globalLimiter.Middleware)
	}

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "sin_base_de_datos"})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			authhandler.New(authSvc, cfg.CookieSecure).Routes(r, loginLimiter.Middleware, authenticate)
		})
		r.Group(func(r chi.Router) {
			r.Use(authenticate, mw.TenantGuard(tenantSvc, opts.TenantCacheTTL), mw.RequirePasswordChanged)
			tenanthandler.New(tenantSvc).Routes(r)
			userhandler.New(userSvc).Routes(r)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) { httpx.Error(w, r, apperr.ErrNotFound) })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, r, apperr.New(apperr.KindInvalid, "metodo_no_permitido", "Método no permitido"))
	})
	return r, nil
}
