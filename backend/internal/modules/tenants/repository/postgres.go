// Package repository implementa la persistencia de empresas.
package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"backend/internal/modules/tenants/domain"
	"backend/internal/platform/database"
)

// Postgres es el repositorio de empresas.
type Postgres struct{}

// likeEscaper evita que %, _ y \ del texto de búsqueda actúen como comodines.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

const selectTenant = `SELECT t.id, t.nombre, t.ruc, t.slug, t.estado, t.creado_en,
	(SELECT count(*) FROM usuarios u WHERE u.tenant_id = t.id)::int
	FROM tenants t `

func scan(row pgx.Row) (domain.Tenant, error) {
	var t domain.Tenant
	err := row.Scan(&t.ID, &t.Nombre, &t.RUC, &t.Slug, &t.Estado, &t.CreadoEn, &t.Usuarios)
	return t, database.NotFound(err)
}

// List devuelve las empresas, filtradas por nombre o RUC si search no está vacío.
func (Postgres) List(ctx context.Context, q database.Querier, search string) ([]domain.Tenant, error) {
	// La barra invertida es el escape por defecto de LIKE en PostgreSQL.
	pattern := likeEscaper.Replace(search)
	rows, err := q.Query(ctx, selectTenant+`
		WHERE $1 = '' OR t.nombre ILIKE '%' || $1 || '%' OR t.ruc LIKE $1 || '%'
		ORDER BY t.nombre`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tenants := []domain.Tenant{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// Get busca una empresa por id.
func (Postgres) Get(ctx context.Context, q database.Querier, id uuid.UUID) (domain.Tenant, error) {
	return scan(q.QueryRow(ctx, selectTenant+`WHERE t.id = $1`, id))
}

// Insert crea la empresa.
func (Postgres) Insert(ctx context.Context, q database.Querier, in domain.CreateInput) (domain.Tenant, error) {
	var t domain.Tenant
	err := q.QueryRow(ctx, `INSERT INTO tenants (nombre, ruc, slug) VALUES ($1, $2, $3)
		RETURNING id, nombre, ruc, slug, estado, creado_en`,
		in.Nombre, in.RUC, domain.Slugify(in.Nombre)).
		Scan(&t.ID, &t.Nombre, &t.RUC, &t.Slug, &t.Estado, &t.CreadoEn)
	return t, mapUnique(err)
}

// UpdateName cambia el nombre. El slug no cambia para no romper referencias.
func (Postgres) UpdateName(ctx context.Context, q database.Querier, id uuid.UUID, nombre string) error {
	tag, err := q.Exec(ctx, `UPDATE tenants SET nombre = $2, actualizado_en = now() WHERE id = $1`, id, nombre)
	if err == nil && tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return mapUnique(err)
}

// SetStatus cambia el estado de la empresa.
func (Postgres) SetStatus(ctx context.Context, q database.Querier, id uuid.UUID, estado string) error {
	tag, err := q.Exec(ctx, `UPDATE tenants SET estado = $2, actualizado_en = now() WHERE id = $1`, id, estado)
	if err == nil && tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return err
}

// IsActive indica si la empresa existe y está activa.
func (Postgres) IsActive(ctx context.Context, q database.Querier, id uuid.UUID) (bool, error) {
	var active bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM tenants WHERE id = $1 AND estado = 'activo')`, id).Scan(&active)
	return active, err
}

func mapUnique(err error) error {
	switch constraint, dup := database.UniqueViolation(err); {
	case !dup:
		return err
	case constraint == "tenants_ruc_key":
		return domain.ErrRUCTaken
	default:
		return domain.ErrNameTaken
	}
}
