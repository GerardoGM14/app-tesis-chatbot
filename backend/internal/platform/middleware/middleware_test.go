package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/platform/authz"
	"backend/internal/platform/reqmeta"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

type fakeParser struct {
	principal authz.Principal
	err       error
}

func (f fakeParser) Parse(string) (authz.Principal, error) { return f.principal, f.err }

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSecureHeaders(t *testing.T) {
	rec := serve(SecureHeaders(okHandler), httptest.NewRequest(http.MethodGet, "/", nil))
	for _, h := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Cache-Control"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("falta la cabecera %s", h)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	p := authz.Principal{UserID: uuid.New(), Role: authz.RoleAgente, TenantID: uuid.New()}
	var got authz.Principal
	capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = authz.PrincipalFrom(r.Context())
	})

	cases := []struct {
		name   string
		header string
		parser fakeParser
		status int
	}{
		{"sin cabecera", "", fakeParser{principal: p}, http.StatusUnauthorized},
		{"esquema incorrecto", "Basic abc", fakeParser{principal: p}, http.StatusUnauthorized},
		{"token vacío", "Bearer ", fakeParser{principal: p}, http.StatusUnauthorized},
		{"token inválido", "Bearer x", fakeParser{err: errors.New("x")}, http.StatusUnauthorized},
		{"token válido", "Bearer x", fakeParser{principal: p}, http.StatusOK},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.header != "" {
			req.Header.Set("Authorization", c.header)
		}
		rec := serve(Authenticate(c.parser)(capture), req)
		if rec.Code != c.status {
			t.Errorf("%s: estado %d, se esperaba %d", c.name, rec.Code, c.status)
		}
	}
	if got != p {
		t.Error("el principal no llegó al handler")
	}
}

func withPrincipal(p authz.Principal) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return req.WithContext(authz.WithPrincipal(req.Context(), p))
}

func TestRequirePermission(t *testing.T) {
	h := RequirePermission(authz.PermUsuariosGestionar)(okHandler)
	if rec := serve(h, httptest.NewRequest(http.MethodGet, "/", nil)); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin principal: %d", rec.Code)
	}
	if rec := serve(h, withPrincipal(authz.Principal{Role: authz.RoleAgente})); rec.Code != http.StatusForbidden {
		t.Errorf("agente: %d", rec.Code)
	}
	if rec := serve(h, withPrincipal(authz.Principal{Role: authz.RoleAdminEmpresa})); rec.Code != http.StatusOK {
		t.Errorf("admin: %d", rec.Code)
	}
}

func TestRequirePasswordChanged(t *testing.T) {
	h := RequirePasswordChanged(okHandler)
	if rec := serve(h, withPrincipal(authz.Principal{MustChangePassword: true})); rec.Code != http.StatusForbidden {
		t.Errorf("cambio pendiente: %d", rec.Code)
	}
	if rec := serve(h, withPrincipal(authz.Principal{})); rec.Code != http.StatusOK {
		t.Errorf("sin cambio pendiente: %d", rec.Code)
	}
}

type fakeChecker struct {
	active bool
	err    error
	calls  int
}

func (f *fakeChecker) IsActive(context.Context, uuid.UUID) (bool, error) {
	f.calls++
	return f.active, f.err
}

func TestTenantGuard(t *testing.T) {
	tenant := authz.Principal{Role: authz.RoleAgente, TenantID: uuid.New()}

	active := &fakeChecker{active: true}
	h := TenantGuard(active, time.Minute)(okHandler)
	for range 3 {
		if rec := serve(h, withPrincipal(tenant)); rec.Code != http.StatusOK {
			t.Fatalf("empresa activa: %d", rec.Code)
		}
	}
	if active.calls != 1 {
		t.Errorf("se esperaba una sola consulta gracias a la caché, hubo %d", active.calls)
	}

	suspended := &fakeChecker{active: false}
	if rec := serve(TenantGuard(suspended, time.Minute)(okHandler), withPrincipal(tenant)); rec.Code != http.StatusForbidden {
		t.Errorf("empresa suspendida: %d", rec.Code)
	}

	failing := &fakeChecker{err: errors.New("bd caída")}
	if rec := serve(TenantGuard(failing, time.Minute)(okHandler), withPrincipal(tenant)); rec.Code != http.StatusInternalServerError {
		t.Errorf("error del checker: %d", rec.Code)
	}

	platform := &fakeChecker{}
	if rec := serve(TenantGuard(platform, time.Minute)(okHandler), withPrincipal(authz.Principal{Role: authz.RoleSuperadmin})); rec.Code != http.StatusOK || platform.calls != 0 {
		t.Error("el superadmin no debe consultar estado de empresa")
	}
}

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(60, 2)
	clock := time.Now()
	l.now = func() time.Time { return clock }

	if !l.Allow("1.1.1.1") || !l.Allow("1.1.1.1") {
		t.Fatal("la ráfaga inicial debe permitirse")
	}
	if l.Allow("1.1.1.1") {
		t.Error("la tercera petición inmediata debe bloquearse")
	}
	if !l.Allow("2.2.2.2") {
		t.Error("otra IP no debe verse afectada")
	}

	clock = clock.Add(time.Second)
	if !l.Allow("1.1.1.1") {
		t.Error("tras un segundo debe recuperar un token")
	}

	clock = clock.Add(20 * time.Minute)
	l.Allow("3.3.3.3")
	if _, kept := l.visitors["2.2.2.2"]; kept {
		t.Error("las IPs inactivas deben limpiarse")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	l := NewRateLimiter(60, 1)
	h := reqmeta.Middleware(l.Middleware(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if rec := serve(h, req); rec.Code != http.StatusOK {
		t.Fatalf("primera petición: %d", rec.Code)
	}
	rec := serve(h, req)
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Errorf("segunda petición: %d", rec.Code)
	}
}

func TestRequestLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if rec := serve(RequestLogger(logger)(okHandler), httptest.NewRequest(http.MethodGet, "/", nil)); rec.Code != http.StatusOK {
		t.Errorf("estado %d", rec.Code)
	}
}
