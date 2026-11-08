// Package handler expone la gestión de usuarios por HTTP.
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/modules/users/domain"
	"backend/internal/modules/users/service"
	"backend/internal/platform/authz"
	"backend/internal/platform/httpx"
	mw "backend/internal/platform/middleware"
)

// Handler atiende /api/v1/usuarios.
type Handler struct {
	svc *service.Service
}

// New crea el handler.
func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

// Routes registra las rutas con su permiso.
func (h *Handler) Routes(r chi.Router) {
	r.With(mw.RequirePermission(authz.PermUsuariosGestionar)).Get("/usuarios", h.list)
	r.With(mw.RequirePermission(authz.PermUsuariosGestionar)).Post("/usuarios", h.create)
	r.With(mw.RequirePermission(authz.PermUsuariosGestionar)).Patch("/usuarios/{id}", h.update)
	r.With(mw.RequirePermission(authz.PermUsuariosRestablecer)).Post("/usuarios/{id}/restablecer", h.reset)
	r.With(mw.RequirePermission(authz.PermEmpresasGestionar)).Get("/empresas/{id}/administradores", h.listAdmins)
}

type listResponse struct {
	Items []domain.User `json:"items"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	p, _ := authz.PrincipalFrom(r.Context())
	users, err := h.svc.List(r.Context(), p)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, listResponse{Items: users})
}

func (h *Handler) listAdmins(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	users, err := h.svc.ListAdmins(r.Context(), tenantID)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, listResponse{Items: users})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, r, err)
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	created, err := h.svc.Create(r.Context(), p, in)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var in domain.UpdateInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, r, err)
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	user, err := h.svc.Update(r.Context(), p, id, in)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

type resetResponse struct {
	PasswordTemporal string `json:"passwordTemporal"`
}

func (h *Handler) reset(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	temp, err := h.svc.ResetPassword(r.Context(), p, id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, resetResponse{PasswordTemporal: temp})
}
