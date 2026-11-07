// Package domain contiene las entidades y reglas de autenticación.
package domain

import (
	"time"

	"github.com/google/uuid"

	"backend/internal/platform/apperr"
	"backend/internal/platform/authz"
)

const (
	// MaxFailedAttempts es la cantidad de intentos fallidos seguidos que bloquea la cuenta.
	MaxFailedAttempts = 5
	// LockDuration es el tiempo que la cuenta queda bloqueada.
	LockDuration = 15 * time.Minute
)

var (
	// ErrInvalidCredentials es el único mensaje de fallo de login: no revela si
	// el email existe, si la cuenta está bloqueada o si la empresa está suspendida.
	ErrInvalidCredentials = apperr.New(apperr.KindUnauthorized, "credenciales_invalidas", "Email o contraseña incorrectos")
	ErrInvalidSession     = apperr.New(apperr.KindUnauthorized, "sesion_invalida", "Tu sesión expiró, vuelve a iniciar sesión")
	ErrWrongPassword      = apperr.Validation(map[string]string{"actual": "La contraseña actual no es correcta"})
	ErrSamePassword       = apperr.Validation(map[string]string{"nueva": "La nueva contraseña debe ser distinta a la actual"})
)

// User es el usuario tal como lo necesita el módulo de autenticación.
type User struct {
	ID                  uuid.UUID
	TenantID            uuid.UUID
	TenantNombre        string
	TenantActivo        bool
	Nombre              string
	Email               string
	PasswordHash        string
	Rol                 authz.Role
	Activo              bool
	DebeCambiarPassword bool
	IntentosFallidos    int
	BloqueadoHasta      *time.Time
}

// IsLocked indica si la cuenta está bloqueada en el instante now.
func (u User) IsLocked(now time.Time) bool {
	return u.BloqueadoHasta != nil && now.Before(*u.BloqueadoHasta)
}

// CanLogin indica si el usuario y su empresa están habilitados.
func (u User) CanLogin() bool {
	return u.Activo && (u.Rol == authz.RoleSuperadmin || u.TenantActivo)
}

// RegisterFailure calcula el nuevo contador de intentos y, si llega al máximo,
// el bloqueo. Al bloquear, el contador vuelve a cero.
func (u User) RegisterFailure(now time.Time) (attempts int, lockedUntil *time.Time) {
	attempts = u.IntentosFallidos + 1
	if attempts >= MaxFailedAttempts {
		until := now.Add(LockDuration)
		return 0, &until
	}
	return attempts, nil
}

// Principal construye el usuario autenticado para el access token.
func (u User) Principal() authz.Principal {
	return authz.Principal{UserID: u.ID, TenantID: u.TenantID, Role: u.Rol, MustChangePassword: u.DebeCambiarPassword}
}

// View es la representación pública del usuario autenticado.
type View struct {
	ID                  uuid.UUID          `json:"id"`
	Nombre              string             `json:"nombre"`
	Email               string             `json:"email"`
	Rol                 authz.Role         `json:"rol"`
	Permisos            []authz.Permission `json:"permisos"`
	TenantID            *uuid.UUID         `json:"tenantId"`
	TenantNombre        string             `json:"tenantNombre,omitempty"`
	DebeCambiarPassword bool               `json:"debeCambiarPassword"`
}

// View devuelve la representación pública del usuario.
func (u User) View() View {
	v := View{
		ID: u.ID, Nombre: u.Nombre, Email: u.Email, Rol: u.Rol, Permisos: u.Rol.Permissions(),
		TenantNombre: u.TenantNombre, DebeCambiarPassword: u.DebeCambiarPassword,
	}
	if u.TenantID != uuid.Nil {
		id := u.TenantID
		v.TenantID = &id
	}
	return v
}

// RefreshToken es una sesión persistida. Solo se guarda el hash del token.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Family    uuid.UUID
	Hash      []byte
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// Session es el resultado de iniciar o renovar una sesión.
type Session struct {
	AccessToken      string
	AccessTTL        time.Duration
	RefreshToken     string
	RefreshExpiresAt time.Time
	User             View
}
