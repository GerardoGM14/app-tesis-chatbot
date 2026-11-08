// Package bootstrap prepara la base de datos después de migrar: activa el rol
// de la aplicación, crea el superadmin inicial y, en desarrollo, datos de demo.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"backend/internal/platform/database"
	"backend/internal/platform/security"
	"backend/migrations"
)

// AppRole es el rol con el que se conecta el backend.
const AppRole = "plataforma_app"

// Hasher genera hashes de contraseña.
type Hasher interface {
	Hash(password string) (string, error)
}

// Run migra la base y deja listo el rol de la aplicación y el superadmin.
// Se conecta con el rol dueño de las tablas (MIGRATION_DATABASE_URL).
func Run(ctx context.Context, conn *pgx.Conn, appPassword string, superadmin Superadmin, hasher Hasher) ([]string, error) {
	applied, err := database.Migrate(ctx, conn, migrations.FS)
	if err != nil {
		return applied, err
	}
	if err := SetAppRolePassword(ctx, conn, appPassword); err != nil {
		return applied, err
	}
	if superadmin.Email != "" {
		if err := EnsureSuperadmin(ctx, conn, superadmin, hasher); err != nil {
			return applied, err
		}
	}
	return applied, nil
}

// SetAppRolePassword habilita el login del rol de la aplicación con la contraseña dada.
func SetAppRolePassword(ctx context.Context, conn *pgx.Conn, password string) error {
	if password == "" {
		return errors.New("la contraseña del rol de la aplicación está vacía")
	}
	var stmt string
	if err := conn.QueryRow(ctx, "SELECT format('ALTER ROLE %I LOGIN PASSWORD %L', $1::text, $2::text)", AppRole, password).Scan(&stmt); err != nil {
		return err
	}
	if _, err := conn.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("no se pudo configurar el rol %s: %w", AppRole, err)
	}
	return nil
}

// Superadmin son los datos del superadministrador inicial.
type Superadmin struct {
	Nombre   string
	Email    string
	Password string
}

// EnsureSuperadmin crea el superadmin si todavía no existe ninguno. Se le exige
// cambiar la contraseña en el primer ingreso.
func EnsureSuperadmin(ctx context.Context, conn *pgx.Conn, s Superadmin, hasher Hasher) error {
	if msg := security.CheckPasswordPolicy(s.Password, s.Email); msg != "" {
		return fmt.Errorf("SUPERADMIN_PASSWORD no cumple la política: %s", msg)
	}
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM usuarios WHERE rol = 'superadmin')").Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	hash, err := hasher.Hash(s.Password)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `INSERT INTO usuarios (nombre, email, password_hash, rol, debe_cambiar_password)
		VALUES ($1, $2, $3, 'superadmin', true)`, s.Nombre, s.Email, hash)
	return err
}
