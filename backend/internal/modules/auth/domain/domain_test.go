package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/platform/authz"
)

func TestRegisterFailureLocksOnFifthAttempt(t *testing.T) {
	now := time.Now()
	u := User{IntentosFallidos: 3}
	attempts, lock := u.RegisterFailure(now)
	if attempts != 4 || lock != nil {
		t.Errorf("cuarto intento: %d, %v", attempts, lock)
	}

	u.IntentosFallidos = 4
	attempts, lock = u.RegisterFailure(now)
	if attempts != 0 || lock == nil || !lock.Equal(now.Add(LockDuration)) {
		t.Errorf("quinto intento debe bloquear: %d, %v", attempts, lock)
	}
}

func TestIsLocked(t *testing.T) {
	now := time.Now()
	future, past := now.Add(time.Minute), now.Add(-time.Minute)
	if (User{}).IsLocked(now) || (User{BloqueadoHasta: &past}).IsLocked(now) {
		t.Error("no debe estar bloqueada")
	}
	if !(User{BloqueadoHasta: &future}).IsLocked(now) {
		t.Error("debe estar bloqueada")
	}
}

func TestCanLogin(t *testing.T) {
	cases := []struct {
		user User
		want bool
	}{
		{User{Activo: true, Rol: authz.RoleSuperadmin}, true},
		{User{Activo: true, Rol: authz.RoleAgente, TenantActivo: true}, true},
		{User{Activo: true, Rol: authz.RoleAgente, TenantActivo: false}, false},
		{User{Activo: false, Rol: authz.RoleAdminEmpresa, TenantActivo: true}, false},
	}
	for i, c := range cases {
		if got := c.user.CanLogin(); got != c.want {
			t.Errorf("caso %d: %v, se esperaba %v", i, got, c.want)
		}
	}
}

func TestViewAndPrincipal(t *testing.T) {
	tenant := uuid.New()
	u := User{ID: uuid.New(), TenantID: tenant, Rol: authz.RoleAgente, DebeCambiarPassword: true, TenantNombre: "Acme"}
	v := u.View()
	if v.TenantID == nil || *v.TenantID != tenant || len(v.Permisos) == 0 {
		t.Errorf("vista inesperada: %+v", v)
	}
	p := u.Principal()
	if p.TenantID != tenant || !p.MustChangePassword || p.Role != authz.RoleAgente {
		t.Errorf("principal inesperado: %+v", p)
	}
	if (User{Rol: authz.RoleSuperadmin}).View().TenantID != nil {
		t.Error("el superadmin no debe exponer tenant")
	}
}
