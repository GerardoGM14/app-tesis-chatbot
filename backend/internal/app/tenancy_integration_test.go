package app_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"backend/internal/platform/database"
)

// Casos AIS de docs/04-calidad/estrategia-de-pruebas.md: el tenant A nunca
// lee ni modifica datos del tenant B.

func TestIsolationThroughAPI(t *testing.T) {
	root := superadmin(t)
	a, b := newCompany(t, root), newCompany(t, root)

	// AIS-01: el listado solo trae usuarios de la propia empresa.
	var list struct {
		Items []userResp `json:"items"`
	}
	status, code := a.admin.call(http.MethodGet, "/api/v1/usuarios", nil, &list)
	expect(t, "listar", status, code, http.StatusOK, "")
	if len(list.Items) != 2 {
		t.Errorf("A debería ver 2 usuarios, ve %d", len(list.Items))
	}
	for _, u := range list.Items {
		if u.TenantID != a.tenant.ID {
			t.Errorf("AIS-01: A ve un usuario de otra empresa: %s", u.Email)
		}
	}

	// AIS-03/04: modificar o restablecer usuarios de B responde 404.
	status, code = a.admin.call(http.MethodPatch, "/api/v1/usuarios/"+b.agentID, map[string]string{"nombre": "Hackeado"}, nil)
	expect(t, "AIS-03 editar usuario de B", status, code, http.StatusNotFound, "no_encontrado")
	status, code = a.admin.call(http.MethodPost, "/api/v1/usuarios/"+b.agentID+"/restablecer", nil, nil)
	expect(t, "AIS-04 restablecer usuario de B", status, code, http.StatusNotFound, "no_encontrado")

	var name string
	_ = env.owner.QueryRow(context.Background(), "SELECT nombre FROM usuarios WHERE id = $1", b.agentID).Scan(&name)
	if name == "Hackeado" {
		t.Error("AIS-03: el usuario de B fue modificado")
	}

	// AIS-05: no se acepta tenant_id en el cuerpo.
	status, code = a.admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
		"nombre": "Intruso", "email": "intruso" + uniqueSuffix() + "@empresa.test", "rol": "agente", "tenantId": b.tenant.ID,
	}, nil)
	expect(t, "AIS-05 tenantId en el cuerpo", status, code, http.StatusUnprocessableEntity, "json_invalido")

	// IDs con formato inválido no revelan nada.
	status, code = a.admin.call(http.MethodPatch, "/api/v1/usuarios/no-es-uuid", map[string]string{"nombre": "X Y"}, nil)
	expect(t, "id inválido", status, code, http.StatusNotFound, "no_encontrado")
}

func TestIsolationAtDatabaseLevel(t *testing.T) {
	root := superadmin(t)
	a, b := newCompany(t, root), newCompany(t, root)
	ctx := context.Background()
	tenantA, tenantB := uuid.MustParse(a.tenant.ID), uuid.MustParse(b.tenant.ID)

	// AIS-06: una consulta sin WHERE con el rol de la app solo ve filas de A.
	err := env.db.InTx(ctx, database.TenantScope(tenantA), func(q database.Querier) error {
		var foreign, total int
		if err := q.QueryRow(ctx, "SELECT count(*) FILTER (WHERE tenant_id <> $1), count(*) FROM usuarios", tenantA).Scan(&foreign, &total); err != nil {
			return err
		}
		if foreign != 0 || total != 2 {
			t.Errorf("AIS-06: A ve %d filas ajenas de %d", foreign, total)
		}
		var tenants int
		if err := q.QueryRow(ctx, "SELECT count(*) FROM tenants").Scan(&tenants); err != nil {
			return err
		}
		if tenants != 1 {
			t.Errorf("AIS-06: A ve %d empresas", tenants)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// AIS-07: insertar con tenant_id de B desde el alcance de A lo rechaza RLS.
	err = env.db.InTx(ctx, database.TenantScope(tenantA), func(q database.Querier) error {
		_, err := q.Exec(ctx, `INSERT INTO usuarios (tenant_id, nombre, email, password_hash, rol)
			VALUES ($1, 'Intruso', $2, 'x', 'agente')`, tenantB, "rls"+uniqueSuffix()+"@empresa.test")
		return err
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Errorf("AIS-07: se esperaba violación de RLS (42501), se obtuvo %v", err)
	}

	// Sin alcance no se ve nada.
	err = env.db.InTx(ctx, database.Scope{}, func(q database.Querier) error {
		var n int
		if err := q.QueryRow(ctx, "SELECT count(*) FROM usuarios").Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("sin alcance se ven %d usuarios", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// La auditoría de A no incluye eventos de B.
	err = env.db.InTx(ctx, database.TenantScope(tenantA), func(q database.Querier) error {
		var foreign int
		if err := q.QueryRow(ctx, "SELECT count(*) FROM auditoria WHERE tenant_id IS DISTINCT FROM $1", tenantA).Scan(&foreign); err != nil {
			return err
		}
		if foreign != 0 {
			t.Errorf("A ve %d eventos de auditoría ajenos", foreign)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppRoleCannotDeleteOrBypass(t *testing.T) {
	ctx := context.Background()
	err := env.db.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		_, err := q.Exec(ctx, "DELETE FROM auditoria")
		return err
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Errorf("la app no debe poder borrar la auditoría: %v", err)
	}
}

func TestTenantLifecycle(t *testing.T) {
	root := superadmin(t)
	co := newCompany(t, root)

	// Validaciones y duplicados.
	status, code := root.call(http.MethodPost, "/api/v1/empresas", map[string]string{
		"nombre": "X", "ruc": "123", "adminNombre": "", "adminEmail": "x",
	}, nil)
	expect(t, "datos inválidos", status, code, http.StatusUnprocessableEntity, "validacion")

	var detail tenantResp
	status, code = root.call(http.MethodGet, "/api/v1/empresas/"+co.tenant.ID, nil, &detail)
	expect(t, "detalle", status, code, http.StatusOK, "")

	var ruc string
	_ = env.owner.QueryRow(context.Background(), "SELECT ruc FROM tenants WHERE id = $1", co.tenant.ID).Scan(&ruc)
	status, code = root.call(http.MethodPost, "/api/v1/empresas", map[string]string{
		"nombre": "Otra " + uniqueSuffix(), "ruc": ruc, "adminNombre": "Admin", "adminEmail": "x" + uniqueSuffix() + "@empresa.test",
	}, nil)
	expect(t, "RUC duplicado", status, code, http.StatusConflict, "ruc_registrado")

	dupName := "Duplicada " + uniqueSuffix()
	status, code = root.call(http.MethodPost, "/api/v1/empresas", map[string]string{
		"nombre": dupName, "ruc": randomRUC(), "adminNombre": "Admin", "adminEmail": co.adminEmail,
	}, nil)
	expect(t, "email de admin duplicado", status, code, http.StatusConflict, "email_registrado")
	var leaked int
	_ = env.owner.QueryRow(context.Background(), "SELECT count(*) FROM tenants WHERE nombre = $1", dupName).Scan(&leaked)
	if leaked != 0 {
		t.Error("si falla el admin, la empresa no debe quedar creada")
	}

	// Renombrar y buscar.
	newName := "Renombrada " + uniqueSuffix()
	status, code = root.call(http.MethodPatch, "/api/v1/empresas/"+co.tenant.ID, map[string]string{"nombre": newName}, nil)
	expect(t, "renombrar", status, code, http.StatusOK, "")
	var found struct {
		Items []tenantResp `json:"items"`
	}
	root.call(http.MethodGet, "/api/v1/empresas?q="+url.QueryEscape(newName[:14]), nil, &found)
	if len(found.Items) == 0 {
		t.Error("la búsqueda por nombre debe encontrar la empresa")
	}
	root.call(http.MethodGet, "/api/v1/empresas?q=%25", nil, &found)
	for _, f := range found.Items {
		if f.ID == co.tenant.ID {
			t.Error("el comodín % no debe coincidir con todo")
		}
	}

	// Suspender: corta acceso, sesiones y login.
	status, code = root.call(http.MethodPost, "/api/v1/empresas/"+co.tenant.ID+"/suspender", nil, &detail)
	expect(t, "suspender", status, code, http.StatusOK, "")
	if detail.Estado != "suspendido" {
		t.Errorf("estado inesperado: %s", detail.Estado)
	}
	status, code = co.admin.call(http.MethodGet, "/api/v1/usuarios", nil, nil)
	expect(t, "token vigente de empresa suspendida", status, code, http.StatusForbidden, "empresa_suspendida")
	status, code = co.admin.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "refresh de empresa suspendida", status, code, http.StatusUnauthorized, "sesion_invalida")
	status, code = newClient(t).login(co.agentEmail, userPassword)
	expect(t, "login de empresa suspendida", status, code, http.StatusUnauthorized, "credenciales_invalidas")

	status, code = root.call(http.MethodPost, "/api/v1/empresas/"+co.tenant.ID+"/reactivar", nil, &detail)
	expect(t, "reactivar", status, code, http.StatusOK, "")
	status, code = newClient(t).login(co.agentEmail, userPassword)
	expect(t, "login tras reactivar", status, code, http.StatusOK, "")

	status, code = root.call(http.MethodPost, "/api/v1/empresas/"+uuid.NewString()+"/suspender", nil, nil)
	expect(t, "empresa inexistente", status, code, http.StatusNotFound, "no_encontrado")
	status, code = root.call(http.MethodPatch, "/api/v1/empresas/"+uuid.NewString(), map[string]string{"nombre": "Nombre Valido"}, nil)
	expect(t, "renombrar inexistente", status, code, http.StatusNotFound, "no_encontrado")
	status, code = root.call(http.MethodGet, "/api/v1/empresas/"+uuid.NewString(), nil, nil)
	expect(t, "detalle inexistente", status, code, http.StatusNotFound, "no_encontrado")
}

func TestUserManagement(t *testing.T) {
	root := superadmin(t)
	co := newCompany(t, root)

	var updated userResp
	status, code := co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]string{"nombre": "Nombre Nuevo"}, &updated)
	expect(t, "renombrar agente", status, code, http.StatusOK, "")
	if updated.Nombre != "Nombre Nuevo" {
		t.Errorf("nombre no actualizado: %s", updated.Nombre)
	}

	status, code = co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.adminID, map[string]any{"activo": false}, nil)
	expect(t, "desactivarse a sí mismo", status, code, http.StatusForbidden, "cambio_propio")

	status, code = co.admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
		"nombre": "Repetido", "email": co.agentEmail, "rol": "agente",
	}, nil)
	expect(t, "email repetido", status, code, http.StatusConflict, "email_registrado")

	status, code = co.admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
		"nombre": "Escalada", "email": "esc" + uniqueSuffix() + "@empresa.test", "rol": "superadmin",
	}, nil)
	expect(t, "crear superadmin", status, code, http.StatusUnprocessableEntity, "validacion")

	// Desactivar al agente cierra sus sesiones.
	status, code = co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]any{"activo": false}, &updated)
	expect(t, "desactivar agente", status, code, http.StatusOK, "")
	status, code = co.agent.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "refresh de agente desactivado", status, code, http.StatusUnauthorized, "sesion_invalida")
	status, code = newClient(t).login(co.agentEmail, userPassword)
	expect(t, "login de agente desactivado", status, code, http.StatusUnauthorized, "credenciales_invalidas")

	// Reactivar y restablecer contraseña.
	co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]any{"activo": true}, nil)
	var reset struct {
		PasswordTemporal string `json:"passwordTemporal"`
	}
	status, code = co.admin.call(http.MethodPost, "/api/v1/usuarios/"+co.agentID+"/restablecer", nil, &reset)
	expect(t, "restablecer agente", status, code, http.StatusOK, "")
	status, code = newClient(t).login(co.agentEmail, userPassword)
	expect(t, "contraseña anterior ya no sirve", status, code, http.StatusUnauthorized, "credenciales_invalidas")
	s := newClient(t).mustLogin(co.agentEmail, reset.PasswordTemporal)
	if !s.User.DebeCambiarPassword {
		t.Error("tras restablecer debe exigirse cambio de contraseña")
	}

	status, code = co.admin.call(http.MethodPost, "/api/v1/usuarios/"+co.adminID+"/restablecer", nil, nil)
	expect(t, "restablecerse a sí mismo", status, code, http.StatusForbidden, "restablecimiento_no_permitido")
	status, code = co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+uuid.NewString(), map[string]string{"nombre": "Nadie Aqui"}, nil)
	expect(t, "editar inexistente", status, code, http.StatusNotFound, "no_encontrado")
	status, code = co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]any{}, nil)
	expect(t, "cambios vacíos", status, code, http.StatusUnprocessableEntity, "validacion")
}

func TestSuperadminResetsCompanyAdmin(t *testing.T) {
	root := superadmin(t)
	co := newCompany(t, root)

	var admins struct {
		Items []userResp `json:"items"`
	}
	status, code := root.call(http.MethodGet, "/api/v1/empresas/"+co.tenant.ID+"/administradores", nil, &admins)
	expect(t, "listar administradores", status, code, http.StatusOK, "")
	if len(admins.Items) != 1 || admins.Items[0].ID != co.adminID {
		t.Errorf("administradores inesperados: %+v", admins.Items)
	}

	var reset struct {
		PasswordTemporal string `json:"passwordTemporal"`
	}
	status, code = root.call(http.MethodPost, "/api/v1/usuarios/"+co.adminID+"/restablecer", nil, &reset)
	expect(t, "superadmin restablece admin", status, code, http.StatusOK, "")
	if reset.PasswordTemporal == "" {
		t.Error("se esperaba una contraseña temporal")
	}

	status, code = root.call(http.MethodPost, "/api/v1/usuarios/"+co.agentID+"/restablecer", nil, nil)
	expect(t, "superadmin no restablece agentes", status, code, http.StatusForbidden, "restablecimiento_no_permitido")

	status, code = co.agent.call(http.MethodPost, "/api/v1/usuarios/"+co.adminID+"/restablecer", nil, nil)
	expect(t, "agente no restablece", status, code, http.StatusForbidden, "prohibido")
}

func TestSessionRevocationRespectsScope(t *testing.T) {
	root := superadmin(t)
	a, b := newCompany(t, root), newCompany(t, root)
	ctx := context.Background()
	scope := database.TenantScope(uuid.MustParse(a.tenant.ID))

	// Desde el alcance de A no se pueden revocar sesiones de B.
	err := env.db.InTx(ctx, scope, func(q database.Querier) error {
		_, err := q.Exec(ctx, "SELECT revocar_sesiones_usuario($1)", b.agentID)
		return err
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
		t.Errorf("se esperaba permiso denegado (42501), se obtuvo %v", err)
	}
	status, code := b.agent.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "la sesión de B sigue vigente", status, code, http.StatusOK, "")

	// La empresa no puede leer los tokens, ni siquiera los propios.
	err = env.db.InTx(ctx, scope, func(q database.Querier) error {
		var n int
		if err := q.QueryRow(ctx, "SELECT count(*) FROM refresh_tokens").Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("el alcance de empresa ve %d refresh tokens", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
