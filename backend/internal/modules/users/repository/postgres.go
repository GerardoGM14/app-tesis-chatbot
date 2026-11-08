// Package repository implementa la persistencia de usuarios.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"backend/internal/modules/users/domain"
	"backend/internal/platform/authz"
	"backend/internal/platform/database"
)

// Postgres es el repositorio de usuarios. Las políticas RLS limitan las filas
// visibles a la empresa del alcance de la transacción.
type Postgres struct{}

const columns = `id, tenant_id, nombre, email, rol, activo, debe_cambiar_password, ultimo_acceso, creado_en`

func scan(row pgx.Row) (domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.TenantID, &u.Nombre, &u.Email, &u.Rol, &u.Activo, &u.DebeCambiarPassword, &u.UltimoAcceso, &u.CreadoEn)
	return u, database.NotFound(err)
}

// List devuelve los usuarios visibles en el alcance, excepto el superadmin.
func (Postgres) List(ctx context.Context, q database.Querier) ([]domain.User, error) {
	return collect(ctx, q, `SELECT `+columns+` FROM usuarios WHERE tenant_id IS NOT NULL ORDER BY nombre`)
}

// ListAdmins devuelve los administradores de una empresa.
func (Postgres) ListAdmins(ctx context.Context, q database.Querier, tenantID uuid.UUID) ([]domain.User, error) {
	return collect(ctx, q, `SELECT `+columns+` FROM usuarios WHERE tenant_id = $1 AND rol = $2 ORDER BY nombre`,
		tenantID, authz.RoleAdminEmpresa)
}

func collect(ctx context.Context, q database.Querier, sql string, args ...any) ([]domain.User, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		u, err := scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Get busca un usuario por id dentro del alcance.
func (Postgres) Get(ctx context.Context, q database.Querier, id uuid.UUID) (domain.User, error) {
	return scan(q.QueryRow(ctx, `SELECT `+columns+` FROM usuarios WHERE id = $1 AND tenant_id IS NOT NULL`, id))
}

// Insert crea un usuario que debe cambiar su contraseña en el primer ingreso.
func (Postgres) Insert(ctx context.Context, q database.Querier, tenantID uuid.UUID, in domain.CreateInput, hash string) (domain.User, error) {
	u, err := scan(q.QueryRow(ctx, `INSERT INTO usuarios (tenant_id, nombre, email, password_hash, rol, debe_cambiar_password)
		VALUES ($1, $2, $3, $4, $5, true) RETURNING `+columns,
		tenantID, in.Nombre, in.Email, hash, in.Rol))
	if _, dup := database.UniqueViolation(err); dup {
		return domain.User{}, domain.ErrEmailTaken
	}
	return u, err
}

// Update aplica los cambios no nulos.
func (Postgres) Update(ctx context.Context, q database.Querier, id uuid.UUID, in domain.UpdateInput) (domain.User, error) {
	return scan(q.QueryRow(ctx, `UPDATE usuarios SET
		nombre = coalesce($2, nombre),
		rol = coalesce($3, rol),
		activo = coalesce($4, activo),
		actualizado_en = now()
		WHERE id = $1 AND tenant_id IS NOT NULL
		RETURNING `+columns, id, in.Nombre, in.Rol, in.Activo))
}

// SetPassword guarda un hash y el indicador de cambio obligatorio.
func (Postgres) SetPassword(ctx context.Context, q database.Querier, id uuid.UUID, hash string, mustChange bool) error {
	_, err := q.Exec(ctx, `UPDATE usuarios SET password_hash = $2, debe_cambiar_password = $3,
		intentos_fallidos = 0, bloqueado_hasta = NULL, actualizado_en = now() WHERE id = $1`, id, hash, mustChange)
	return err
}
