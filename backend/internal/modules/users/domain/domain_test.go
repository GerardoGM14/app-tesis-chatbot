package domain

import (
	"testing"

	"github.com/google/uuid"

	"backend/internal/platform/authz"
)

func ptr[T any](v T) *T { return &v }

func TestCreateInputNormalizeAndValidate(t *testing.T) {
	in := CreateInput{Nombre: "  Ana Torres ", Email: " ANA@Empresa.PE ", Rol: authz.RoleAgente}.Normalize()
	if in.Nombre != "Ana Torres" || in.Email != "ana@empresa.pe" {
		t.Errorf("normalización inesperada: %+v", in)
	}
	if err := in.Validate(); err != nil {
		t.Errorf("entrada válida rechazada: %v", err)
	}

	invalid := []CreateInput{
		{Nombre: "A", Email: "ana@empresa.pe", Rol: authz.RoleAgente},
		{Nombre: "Ana", Email: "ana", Rol: authz.RoleAgente},
		{Nombre: "Ana", Email: "ana@empresa.pe", Rol: authz.RoleSuperadmin},
	}
	for _, c := range invalid {
		if c.Validate() == nil {
			t.Errorf("se esperaba rechazo: %+v", c)
		}
	}
}

func TestUpdateInputValidate(t *testing.T) {
	if (UpdateInput{}).Validate() == nil {
		t.Error("una actualización vacía debe rechazarse")
	}
	if (UpdateInput{Rol: ptr(authz.RoleSuperadmin)}).Validate() == nil {
		t.Error("no se puede asignar superadmin")
	}
	if (UpdateInput{Nombre: ptr(" x ")}).Normalize().Validate() == nil {
		t.Error("nombre demasiado corto")
	}
	if err := (UpdateInput{Activo: ptr(false)}).Validate(); err != nil {
		t.Errorf("desactivar es válido: %v", err)
	}
}

func TestTouchesAccess(t *testing.T) {
	current := User{Rol: authz.RoleAgente, Activo: true}
	if (UpdateInput{Nombre: ptr("Otro")}).TouchesAccess(current) {
		t.Error("cambiar el nombre no afecta el acceso")
	}
	if (UpdateInput{Rol: ptr(authz.RoleAgente)}).TouchesAccess(current) {
		t.Error("el mismo rol no afecta el acceso")
	}
	if !(UpdateInput{Rol: ptr(authz.RoleAdminEmpresa)}).TouchesAccess(current) {
		t.Error("cambiar el rol afecta el acceso")
	}
	if !(UpdateInput{Activo: ptr(false)}).TouchesAccess(current) {
		t.Error("desactivar afecta el acceso")
	}
}

func TestCanReset(t *testing.T) {
	tenant, other := uuid.New(), uuid.New()
	super := authz.Principal{UserID: uuid.New(), Role: authz.RoleSuperadmin}
	admin := authz.Principal{UserID: uuid.New(), Role: authz.RoleAdminEmpresa, TenantID: tenant}
	agentP := authz.Principal{UserID: uuid.New(), Role: authz.RoleAgente, TenantID: tenant}

	adminUser := User{ID: uuid.New(), Rol: authz.RoleAdminEmpresa, TenantID: tenant}
	agentUser := User{ID: uuid.New(), Rol: authz.RoleAgente, TenantID: tenant}
	foreignAgent := User{ID: uuid.New(), Rol: authz.RoleAgente, TenantID: other}

	cases := []struct {
		name   string
		actor  authz.Principal
		target User
		want   bool
	}{
		{"superadmin a admin", super, adminUser, true},
		{"superadmin a agente", super, agentUser, false},
		{"admin a agente propio", admin, agentUser, true},
		{"admin a otro admin", admin, adminUser, false},
		{"admin a agente ajeno", admin, foreignAgent, false},
		{"admin a sí mismo", admin, User{ID: admin.UserID, Rol: authz.RoleAgente, TenantID: tenant}, false},
		{"agente a agente", agentP, agentUser, false},
	}
	for _, c := range cases {
		if got := CanReset(c.actor, c.target); got != c.want {
			t.Errorf("%s: %v, se esperaba %v", c.name, got, c.want)
		}
	}
}
