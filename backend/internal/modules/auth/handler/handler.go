// Package handler expone los casos de uso de autenticación por HTTP.
package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"backend/internal/modules/auth/domain"
	"backend/internal/modules/auth/service"
	"backend/internal/platform/apperr"
	"backend/internal/platform/authz"
	"backend/internal/platform/httpx"
	"backend/internal/platform/validate"
)

const (
	refreshCookie = "rt"
	cookiePath    = "/api/v1/auth"
)

// Handler atiende /api/v1/auth.
type Handler struct {
	svc          *service.Service
	cookieSecure bool
}

// New crea el handler.
func New(svc *service.Service, cookieSecure bool) *Handler {
	return &Handler{svc: svc, cookieSecure: cookieSecure}
}

// Routes registra las rutas. limiter protege login y refresh contra fuerza bruta;
// authenticate exige access token en las rutas de la sesión actual.
func (h *Handler) Routes(r chi.Router, limiter, authenticate func(http.Handler) http.Handler) {
	r.With(limiter).Post("/login", h.login)
	r.With(limiter).Post("/refresh", h.refresh)
	r.Post("/logout", h.logout)
	r.Group(func(r chi.Router) {
		r.Use(authenticate)
		r.Get("/me", h.me)
		r.Post("/password", h.changePassword)
	})
}

type sessionResponse struct {
	AccessToken string      `json:"accessToken"`
	ExpiresIn   int         `json:"expiresIn"`
	User        domain.View `json:"user"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	v := validate.New()
	v.Check(validate.Email(email), "email", "Ingresa un email válido")
	v.Check(req.Password != "" && len(req.Password) <= 128, "password", "Ingresa tu contraseña")
	if err := v.Err(); err != nil {
		httpx.Error(w, r, err)
		return
	}

	session, err := h.svc.Login(r.Context(), email, req.Password)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	h.writeSession(w, session)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookie)
	if err != nil || cookie.Value == "" {
		httpx.Error(w, r, domain.ErrInvalidSession)
		return
	}
	session, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearCookie(w)
		httpx.Error(w, r, err)
		return
	}
	h.writeSession(w, session)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookie); err == nil && cookie.Value != "" {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			httpx.Error(w, r, err)
			return
		}
	}
	h.clearCookie(w)
	httpx.NoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	p, _ := authz.PrincipalFrom(r.Context())
	view, err := h.svc.Me(r.Context(), p)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

type changePasswordRequest struct {
	Actual string `json:"actual"`
	Nueva  string `json:"nueva"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	if req.Actual == "" || req.Nueva == "" {
		httpx.Error(w, r, apperr.Validation(map[string]string{"nueva": "Completa ambas contraseñas"}))
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	session, err := h.svc.ChangePassword(r.Context(), p, req.Actual, req.Nueva)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	h.writeSession(w, session)
}

func (h *Handler) writeSession(w http.ResponseWriter, s domain.Session) {
	http.SetCookie(w, &http.Cookie{
		Name: refreshCookie, Value: s.RefreshToken, Path: cookiePath,
		Expires: s.RefreshExpiresAt, MaxAge: int(time.Until(s.RefreshExpiresAt).Seconds()),
		HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteStrictMode,
	})
	httpx.JSON(w, http.StatusOK, sessionResponse{
		AccessToken: s.AccessToken, ExpiresIn: int(s.AccessTTL.Seconds()), User: s.User,
	})
}

func (h *Handler) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: refreshCookie, Value: "", Path: cookiePath, MaxAge: -1,
		HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteStrictMode,
	})
}
