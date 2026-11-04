// Package database administra el pool de conexiones y las transacciones con
// contexto de tenant. Toda consulta de negocio pasa por InTx para que las
// políticas de Row-Level Security de PostgreSQL apliquen siempre.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound indica que la consulta no devolvió filas.
var ErrNotFound = errors.New("registro no encontrado")

// Querier es el subconjunto de pgx que usan los repositorios.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Scope define qué filas puede ver una transacción.
type Scope struct {
	// TenantID limita la transacción a una empresa. uuid.Nil si no aplica.
	TenantID uuid.UUID
	// Platform habilita el alcance de plataforma (superadmin y procesos de sistema).
	Platform bool
}

// TenantScope crea un alcance limitado a una empresa.
func TenantScope(id uuid.UUID) Scope { return Scope{TenantID: id} }

// PlatformScope crea el alcance de plataforma.
func PlatformScope() Scope { return Scope{Platform: true} }

// TxRunner ejecuta una función dentro de una transacción con alcance.
type TxRunner interface {
	InTx(ctx context.Context, scope Scope, fn func(q Querier) error) error
}

// DB envuelve el pool de conexiones.
type DB struct {
	pool *pgxpool.Pool
}

// Connect abre el pool y verifica la conexión.
func Connect(ctx context.Context, url string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("configuración de base de datos inválida: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close cierra el pool.
func (d *DB) Close() { d.pool.Close() }

// Ping verifica que la base de datos responda.
func (d *DB) Ping(ctx context.Context) error { return d.pool.Ping(ctx) }

// InTx abre una transacción, fija el alcance con SET LOCAL y ejecuta fn.
// Si fn devuelve error se hace rollback; si no, commit.
func (d *DB) InTx(ctx context.Context, scope Scope, fn func(q Querier) error) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar la transacción: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tenant := ""
	if scope.TenantID != uuid.Nil {
		tenant = scope.TenantID.String()
	}
	platform := "off"
	if scope.Platform {
		platform = "on"
	}
	if _, err := tx.Exec(ctx,
		"SELECT set_config('app.tenant_id', $1, true), set_config('app.is_platform', $2, true)",
		tenant, platform,
	); err != nil {
		return fmt.Errorf("no se pudo fijar el alcance: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("no se pudo confirmar la transacción: %w", err)
	}
	return nil
}

// NotFound convierte pgx.ErrNoRows en ErrNotFound.
func NotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// UniqueViolation indica si el error es una violación de unicidad y devuelve
// el nombre de la restricción.
func UniqueViolation(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}
