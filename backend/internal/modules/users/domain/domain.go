// Package domain contiene las entidades y reglas de los usuarios de una empresa.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/platform/apperr"
	"backend/internal/platform/authz"
	"backend/internal/platform/validate"
)

var (
	ErrEmailTaken     = apperr.New(apperr.KindConflict, "email_registrado", "Ya existe un usuario con ese email")
	ErrSelfChange     = apperr.New(apperr.KindForbidden, "cambio_propio", "No puedes cambiar tu propio rol ni desactivarte")
	ErrResetForbidden = apperr.New(apperr.KindForbidden, "restablecimiento_no_permitido", "No puedes restablecer la contraseña de este usuario")
)

// User es un usuario de una empresa.
type User struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenantId"`
	Nombre              string     `json:"nombre"`
	Email               string     `json:"email"`
	Rol                 authz.Role `json:"rol"`
	Activo              bool       `json:"activo"`
	DebeCambiarPassword bool       `json:"debeCambiarPassword"`
	UltimoAcceso        *time.Time `json:"ultimoAcceso"`
	CreadoEn            time.Time  `json:"creadoEn"`
}

// Created es un usuario recién creado con su contraseña temporal, que se
// muestra una sola vez.
type Created struct {
	Usuario          User   `json:"usuario"`
	PasswordTemporal string `json:"passwordTemporal"`
}

// AssignableRole indica si un administrador de empresa puede asignar el rol.
func AssignableRole(r authz.Role) bool {
	return r == authz.RoleAdminEmpresa || r == authz.RoleAgente
}

// CreateInput son los datos para crear un usuario.
type CreateInput struct {
	Nombre string     `json:"nombre"`
	Email  string     `json:"email"`
	Rol    authz.Role `json:"rol"`
}

// Normalize limpia espacios y pasa el email a minúsculas.
func (in CreateInput) Normalize() CreateInput {
	in.Nombre = strings.TrimSpace(in.Nombre)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	return in
}

// Validate verifica los datos de creación.
func (in CreateInput) Validate() error {
	v := validate.New()
	v.Check(validate.Length(in.Nombre, 2, 120), "nombre", "Ingresa un nombre de 2 a 120 caracteres")
	v.Check(validate.Email(in.Email), "email", "Ingresa un email válido")
	v.Check(AssignableRole(in.Rol), "rol", "Elige un rol válido")
	return v.Err()
}

// UpdateInput son los cambios permitidos sobre un usuario. Los campos nil no cambian.
type UpdateInput struct {
	Nombre *string     `json:"nombre"`
	Rol    *authz.Role `json:"rol"`
	Activo *bool       `json:"activo"`
}

// Normalize limpia espacios del nombre.
func (in UpdateInput) Normalize() UpdateInput {
	if in.Nombre != nil {
		n := strings.TrimSpace(*in.Nombre)
		in.Nombre = &n
	}
	return in
}

// Validate verifica los cambios.
func (in UpdateInput) Validate() error {
	v := validate.New()
	v.Check(in.Nombre != nil || in.Rol != nil || in.Activo != nil, "nombre", "No hay cambios para guardar")
	if in.Nombre != nil {
		v.Check(validate.Length(*in.Nombre, 2, 120), "nombre", "Ingresa un nombre de 2 a 120 caracteres")
	}
	if in.Rol != nil {
		v.Check(AssignableRole(*in.Rol), "rol", "Elige un rol válido")
	}
	return v.Err()
}

// TouchesAccess indica si el cambio afecta el acceso (rol o estado), lo que
// obliga a cerrar las sesiones del usuario.
func (in UpdateInput) TouchesAccess(current User) bool {
	return (in.Rol != nil && *in.Rol != current.Rol) || (in.Activo != nil && *in.Activo != current.Activo)
}

// CanReset indica si actor puede restablecer la contraseña de target:
// el superadmin a los administradores de empresa, y el administrador a los
// agentes de su propia empresa.
func CanReset(actor authz.Principal, target User) bool {
	if actor.UserID == target.ID {
		return false
	}
	switch actor.Role {
	case authz.RoleSuperadmin:
		return target.Rol == authz.RoleAdminEmpresa
	case authz.RoleAdminEmpresa:
		return target.Rol == authz.RoleAgente && target.TenantID == actor.TenantID
	default:
		return false
	}
}
