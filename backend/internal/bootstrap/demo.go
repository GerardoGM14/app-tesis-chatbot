package bootstrap

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// DemoPassword es la contraseña de todos los usuarios de demostración.
// Solo se usa en APP_ENV=development.
const DemoPassword = "Demo-Sertech-2026" //nolint:gosec // credencial pública de demostración, el comando se niega fuera de development

type demoTenant struct {
	nombre, ruc, slug string
	users             []demoUser
}

type demoUser struct {
	nombre, email, rol string
}

var demoTenants = []demoTenant{
	{
		nombre: "Ferretería El Progreso", ruc: "20600000005", slug: "ferreteria-el-progreso",
		users: []demoUser{
			{"Lucía Ramos", "admin@progreso.demo", "admin_empresa"},
			{"Diego Salazar", "agente@progreso.demo", "agente"},
		},
	},
	{
		nombre: "Clínica Santa Rosa", ruc: "20611111119", slug: "clinica-santa-rosa",
		users: []demoUser{
			{"María Paredes", "admin@santarosa.demo", "admin_empresa"},
			{"Jorge Huamán", "agente@santarosa.demo", "agente"},
		},
	},
}

// SeedDemo crea dos empresas con un administrador y un agente cada una, si
// todavía no existen. Los usuarios no necesitan cambiar la contraseña.
func SeedDemo(ctx context.Context, conn *pgx.Conn, hasher Hasher) error {
	hash, err := hasher.Hash(DemoPassword)
	if err != nil {
		return err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, t := range demoTenants {
		var tenantID string
		err := tx.QueryRow(ctx, `INSERT INTO tenants (nombre, ruc, slug) VALUES ($1, $2, $3)
			ON CONFLICT (ruc) DO UPDATE SET nombre = tenants.nombre RETURNING id`, t.nombre, t.ruc, t.slug).Scan(&tenantID)
		if err != nil {
			return err
		}
		for _, u := range t.users {
			if _, err := tx.Exec(ctx, `INSERT INTO usuarios (tenant_id, nombre, email, password_hash, rol, debe_cambiar_password)
				VALUES ($1, $2, $3, $4, $5, false) ON CONFLICT (email) DO NOTHING`,
				tenantID, u.nombre, u.email, hash, u.rol); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}
