// Package authz define los roles, sus permisos y el usuario autenticado
// (Principal) que viaja en el contexto de cada petición.
package authz

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"backend/internal/platform/database"
)

// Role es el rol de un usuario.
type Role string

const (
	RoleSuperadmin   Role = "superadmin"
	RoleAdminEmpresa Role = "admin_empresa"
	RoleAgente       Role = "agente"
)

// Permission es una acción que un rol puede realizar.
type Permission string

const (
	PermDashboardVer          Permission = "dashboard:ver"
	PermPerfilEditar          Permission = "perfil:editar"
	PermEmpresasGestionar     Permission = "empresas:gestionar"
	PermMetricasGlobales      Permission = "metricas:globales"
	PermBotGestionar          Permission = "bot:gestionar"
	PermCanalGestionar        Permission = "canal:gestionar"
	PermUsuariosGestionar     Permission = "usuarios:gestionar"
	PermUsuariosRestablecer   Permission = "usuarios:restablecer"
	PermConversacionesAtender Permission = "conversaciones:atender"
)

// rolePermissions es la matriz de permisos documentada en
// docs/01-arquitectura/seguridad/autenticacion-y-roles.md.
var rolePermissions = map[Role][]Permission{
	RoleSuperadmin: {
		PermDashboardVer, PermPerfilEditar,
		PermEmpresasGestionar, PermMetricasGlobales,
		PermUsuariosRestablecer,
	},
	RoleAdminEmpresa: {
		PermDashboardVer, PermPerfilEditar,
		PermBotGestionar, PermCanalGestionar, PermUsuariosGestionar,
		PermUsuariosRestablecer, PermConversacionesAtender,
	},
	RoleAgente: {
		PermDashboardVer, PermPerfilEditar,
		PermConversacionesAtender,
	},
}

// Valid indica si el rol existe.
func (r Role) Valid() bool {
	_, ok := rolePermissions[r]
	return ok
}

// Can indica si el rol tiene el permiso.
func (r Role) Can(p Permission) bool {
	return slices.Contains(rolePermissions[r], p)
}

// Permissions devuelve una copia de los permisos del rol.
func (r Role) Permissions() []Permission {
	return slices.Clone(rolePermissions[r])
}

// Principal es el usuario autenticado de la petición.
type Principal struct {
	UserID             uuid.UUID
	TenantID           uuid.UUID
	Role               Role
	MustChangePassword bool
}

// Scope devuelve el alcance de base de datos que corresponde al usuario.
func (p Principal) Scope() database.Scope {
	if p.Role == RoleSuperadmin {
		return database.PlatformScope()
	}
	return database.TenantScope(p.TenantID)
}

type principalKey struct{}

// WithPrincipal guarda el usuario autenticado en el contexto.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom obtiene el usuario autenticado del contexto.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
