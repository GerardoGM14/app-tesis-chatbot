// Package repository implementa la persistencia del módulo de autenticación.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"backend/internal/modules/auth/domain"
	"backend/internal/platform/database"
)

// Postgres es el repositorio de autenticación sobre PostgreSQL.
type Postgres struct{}

const selectUser = `SELECT u.id, coalesce(u.tenant_id, '00000000-0000-0000-0000-000000000000'::uuid),
	coalesce(t.nombre, ''), coalesce(t.estado = 'activo', false),
	u.nombre, u.email, u.password_hash, u.rol, u.activo, u.debe_cambiar_password,
	u.intentos_fallidos, u.bloqueado_hasta
	FROM usuarios u LEFT JOIN tenants t ON t.id = u.tenant_id `

func scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.TenantID, &u.TenantNombre, &u.TenantActivo,
		&u.Nombre, &u.Email, &u.PasswordHash, &u.Rol, &u.Activo, &u.DebeCambiarPassword,
		&u.IntentosFallidos, &u.BloqueadoHasta)
	return u, database.NotFound(err)
}

// FindUserByEmail busca un usuario por email (ya normalizado en minúsculas).
func (Postgres) FindUserByEmail(ctx context.Context, q database.Querier, email string) (domain.User, error) {
	return scanUser(q.QueryRow(ctx, selectUser+"WHERE u.email = $1", email))
}

// FindUserByID busca un usuario por id.
func (Postgres) FindUserByID(ctx context.Context, q database.Querier, id uuid.UUID) (domain.User, error) {
	return scanUser(q.QueryRow(ctx, selectUser+"WHERE u.id = $1", id))
}

// UpdateFailedAttempts guarda el contador de intentos y el bloqueo.
func (Postgres) UpdateFailedAttempts(ctx context.Context, q database.Querier, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	_, err := q.Exec(ctx, `UPDATE usuarios SET intentos_fallidos = $2, bloqueado_hasta = $3 WHERE id = $1`, id, attempts, lockedUntil)
	return err
}

// RegisterSuccessfulLogin reinicia el contador y registra el último acceso.
func (Postgres) RegisterSuccessfulLogin(ctx context.Context, q database.Querier, id uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE usuarios SET intentos_fallidos = 0, bloqueado_hasta = NULL, ultimo_acceso = now() WHERE id = $1`, id)
	return err
}

// UpdatePassword guarda un nuevo hash y el indicador de cambio obligatorio.
func (Postgres) UpdatePassword(ctx context.Context, q database.Querier, id uuid.UUID, hash string, mustChange bool) error {
	_, err := q.Exec(ctx, `UPDATE usuarios SET password_hash = $2, debe_cambiar_password = $3, actualizado_en = now() WHERE id = $1`,
		id, hash, mustChange)
	return err
}

// InsertRefreshToken persiste una sesión.
func (Postgres) InsertRefreshToken(ctx context.Context, q database.Querier, t domain.RefreshToken) error {
	_, err := q.Exec(ctx, `INSERT INTO refresh_tokens (id, usuario_id, familia, token_hash, expira_en) VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.UserID, t.Family, t.Hash, t.ExpiresAt)
	return err
}

// FindRefreshTokenForUpdate busca la sesión por hash y la bloquea hasta el fin
// de la transacción, para que dos renovaciones simultáneas no la usen a la vez.
func (Postgres) FindRefreshTokenForUpdate(ctx context.Context, q database.Querier, hash []byte) (domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := q.QueryRow(ctx, `SELECT id, usuario_id, familia, token_hash, expira_en, revocado_en
		FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE`, hash).
		Scan(&t.ID, &t.UserID, &t.Family, &t.Hash, &t.ExpiresAt, &t.RevokedAt)
	return t, database.NotFound(err)
}

// RevokeRefreshToken revoca una sesión.
func (Postgres) RevokeRefreshToken(ctx context.Context, q database.Querier, id uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE refresh_tokens SET revocado_en = now() WHERE id = $1 AND revocado_en IS NULL`, id)
	return err
}

// RevokeFamily revoca todas las sesiones derivadas del mismo login.
func (Postgres) RevokeFamily(ctx context.Context, q database.Querier, family uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE refresh_tokens SET revocado_en = now() WHERE familia = $1 AND revocado_en IS NULL`, family)
	return err
}

// RevokeAllForUser cierra todas las sesiones de un usuario.
func (Postgres) RevokeAllForUser(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE refresh_tokens SET revocado_en = now() WHERE usuario_id = $1 AND revocado_en IS NULL`, userID)
	return err
}

// RevokeAllForTenant cierra todas las sesiones de los usuarios de una empresa.
func (Postgres) RevokeAllForTenant(ctx context.Context, q database.Querier, tenantID uuid.UUID) error {
	_, err := q.Exec(ctx, `UPDATE refresh_tokens SET revocado_en = now()
		WHERE revocado_en IS NULL AND usuario_id IN (SELECT id FROM usuarios WHERE tenant_id = $1)`, tenantID)
	return err
}
