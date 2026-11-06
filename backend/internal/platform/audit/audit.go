// Package audit registra acciones relevantes para la seguridad en la tabla
// auditoria, dentro de la misma transacción que la acción auditada.
package audit

import (
	"context"
	"encoding/json"
	"net"

	"github.com/google/uuid"

	"backend/internal/platform/database"
	"backend/internal/platform/reqmeta"
)

// Acciones auditadas.
const (
	ActionLoginOK           = "login_exitoso"
	ActionLoginFailed       = "login_fallido"
	ActionAccountLocked     = "cuenta_bloqueada"
	ActionLogout            = "logout"
	ActionRefreshReuse      = "refresh_reutilizado"
	ActionPasswordChanged   = "password_cambiado"
	ActionPasswordReset     = "password_restablecido"
	ActionTenantCreated     = "empresa_creada"
	ActionTenantUpdated     = "empresa_actualizada"
	ActionTenantSuspended   = "empresa_suspendida"
	ActionTenantReactivated = "empresa_reactivada"
	ActionUserCreated       = "usuario_creado"
	ActionUserUpdated       = "usuario_actualizado"
)

// Entry es un registro de auditoría.
type Entry struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Action   string
	Entity   string
	EntityID uuid.UUID
	Detail   map[string]any
}

// Recorder escribe registros de auditoría.
type Recorder struct{}

// Record inserta el registro. La IP se toma del contexto de la petición.
func (Recorder) Record(ctx context.Context, q database.Querier, e Entry) error {
	var detail []byte
	if len(e.Detail) > 0 {
		var err error
		if detail, err = json.Marshal(e.Detail); err != nil {
			return err
		}
	}
	_, err := q.Exec(ctx, `INSERT INTO auditoria (tenant_id, usuario_id, accion, entidad, entidad_id, ip, detalle)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)`,
		nullable(e.TenantID), nullable(e.UserID), e.Action, e.Entity, nullable(e.EntityID),
		validIP(reqmeta.ClientIP(ctx)), detail)
	return err
}

func nullable(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func validIP(ip string) *string {
	if net.ParseIP(ip) == nil {
		return nil
	}
	return &ip
}
