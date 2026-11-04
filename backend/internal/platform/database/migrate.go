package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// migrationLockID identifica el advisory lock que evita migraciones concurrentes.
const migrationLockID = 727_301

// Migrate aplica en orden los archivos .sql de fsys que aún no se aplicaron.
// Cada archivo corre en su propia transacción.
func Migrate(ctx context.Context, conn *pgx.Conn, fsys fs.FS) ([]string, error) {
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version text PRIMARY KEY,
		aplicada_en timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return nil, fmt.Errorf("no se pudo crear schema_migrations: %w", err)
	}

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockID); err != nil {
		return nil, fmt.Errorf("no se pudo obtener el lock de migración: %w", err)
	}
	defer func() { _, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLockID) }()

	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	applied := make([]string, 0, len(files))
	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		done, err := isApplied(ctx, conn, version)
		if err != nil {
			return applied, err
		}
		if done {
			continue
		}
		if err := applyFile(ctx, conn, fsys, file, version); err != nil {
			return applied, err
		}
		applied = append(applied, version)
	}
	return applied, nil
}

func isApplied(ctx context.Context, conn *pgx.Conn, version string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("no se pudo consultar la migración %s: %w", version, err)
	}
	return exists, nil
}

func applyFile(ctx context.Context, conn *pgx.Conn, fsys fs.FS, file, version string) error {
	content, err := fs.ReadFile(fsys, file)
	if err != nil {
		return err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Sin argumentos pgx usa el protocolo simple, que admite varias sentencias.
	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("falló la migración %s: %w", version, err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
