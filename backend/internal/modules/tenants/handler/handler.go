// Package handler expone la administración de empresas por HTTP.
package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/modules/tenants/domain"
	"backend/internal/modules/tenants/service"
	"backend/internal/platform/authz"
	"backend/internal/platform/httpx"
	mw "backend/internal/platform/middleware"
)

// Handler atiende /api/v1/empresas.
type Handler struct {
	svc *service.Service
}

// New crea el handler.
func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

// Routes registra las rutas. Todas exigen el permiso empresas:gestionar.
func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(mw.RequirePermission(authz.PermEmpresasGestionar))
		r.Get("/empresas", h.list)
		r.Post("/empresas", h.create)
		r.Get("/empresas/{id}", h.get)
		r.Patch("/empresas/{id}", h.rename)
		r.Post("/empresas/{id}/suspender", h.suspend)
		r.Post("/empresas/{id}/reactivar", h.reactivate)
	})
}

type listResponse struct {
	Items []domain.Tenant `json:"items"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(search) > 120 {
		search = search[:120]
	}
	tenants, err := h.svc.List(r.Context(), search)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, listResponse{Items: tenants})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	tenant, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, tenant)
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

type renameRequest struct {
	Nombre string `json:"nombre"`
}

func (h *Handler) rename(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	var req renameRequest
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	tenant, err := h.svc.Rename(r.Context(), p, id, strings.Join(strings.Fields(req.Nombre), " "))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, tenant)
}

func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.svc.Suspend)
}

func (h *Handler) reactivate(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.svc.Reactivate)
}

type statusFn func(ctx context.Context, p authz.Principal, id uuid.UUID) (domain.Tenant, error)

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request, fn statusFn) {
	id, err := httpx.PathUUID(r, "id")
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	p, _ := authz.PrincipalFrom(r.Context())
	tenant, err := fn(r.Context(), p, id)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, tenant)
}
