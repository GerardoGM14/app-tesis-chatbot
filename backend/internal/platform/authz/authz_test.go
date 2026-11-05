package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestPermissionMatrix(t *testing.T) {
	cases := []struct {
		role Role
		perm Permission
		want bool
	}{
		{RoleSuperadmin, PermEmpresasGestionar, true},
		{RoleSuperadmin, PermConversacionesAtender, false},
		{RoleSuperadmin, PermUsuariosGestionar, false},
		{RoleAdminEmpresa, PermUsuariosGestionar, true},
		{RoleAdminEmpresa, PermEmpresasGestionar, false},
		{RoleAgente, PermConversacionesAtender, true},
		{RoleAgente, PermBotGestionar, false},
		{RoleAgente, PermUsuariosRestablecer, false},
		{Role("desconocido"), PermDashboardVer, false},
	}
	for _, c := range cases {
		if got := c.role.Can(c.perm); got != c.want {
			t.Errorf("%s.Can(%s) = %v, se esperaba %v", c.role, c.perm, got, c.want)
		}
	}
}

func TestRoleValid(t *testing.T) {
	if !RoleAgente.Valid() || Role("root").Valid() {
		t.Error("validación de roles incorrecta")
	}
}

func TestPermissionsReturnsCopy(t *testing.T) {
	perms := RoleAgente.Permissions()
	perms[0] = PermEmpresasGestionar
	if RoleAgente.Can(PermEmpresasGestionar) {
		t.Error("modificar la copia no debe alterar la matriz")
	}
}

func TestPrincipalScope(t *testing.T) {
	tenant := uuid.New()
	if s := (Principal{Role: RoleSuperadmin}).Scope(); !s.Platform {
		t.Error("el superadmin debe tener alcance de plataforma")
	}
	s := (Principal{Role: RoleAgente, TenantID: tenant}).Scope()
	if s.Platform || s.TenantID != tenant {
		t.Error("un agente debe tener alcance de su empresa")
	}
}

func TestPrincipalContext(t *testing.T) {
	if _, ok := PrincipalFrom(context.Background()); ok {
		t.Error("un contexto vacío no debe tener principal")
	}
	p := Principal{UserID: uuid.New(), Role: RoleAgente}
	got, ok := PrincipalFrom(WithPrincipal(context.Background(), p))
	if !ok || got != p {
		t.Error("el principal no se recuperó del contexto")
	}
}
