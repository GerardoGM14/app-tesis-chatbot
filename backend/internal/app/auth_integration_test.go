package app_test

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// Casos AUT de docs/04-calidad/estrategia-de-pruebas.md.

func TestHealthEndpoints(t *testing.T) {
	c := newClient(t)
	expect(t, "healthz", statusOf(c.call(http.MethodGet, "/healthz", nil, nil)), "", http.StatusOK, "")
	expect(t, "readyz", statusOf(c.call(http.MethodGet, "/readyz", nil, nil)), "", http.StatusOK, "")
	status, code := c.call(http.MethodGet, "/no-existe", nil, nil)
	expect(t, "ruta inexistente", status, code, http.StatusNotFound, "no_encontrado")
	status, code = c.call(http.MethodPut, "/healthz", nil, nil)
	expect(t, "método no permitido", status, code, http.StatusUnprocessableEntity, "metodo_no_permitido")
}

func statusOf(status int, _ string) int { return status }

func TestLoginSuccessAndRoleRedirectData(t *testing.T) { // AUT-01
	root := superadmin(t)
	var me struct {
		Rol      string   `json:"rol"`
		TenantID *string  `json:"tenantId"`
		Permisos []string `json:"permisos"`
	}
	status, code := root.call(http.MethodGet, "/api/v1/auth/me", nil, &me)
	expect(t, "me", status, code, http.StatusOK, "")
	if me.Rol != "superadmin" || me.TenantID != nil || len(me.Permisos) == 0 {
		t.Errorf("datos de sesión inesperados: %+v", me)
	}
	if root.refreshCookie() == "" {
		t.Error("el login debe dejar la cookie de refresh")
	}
}

func TestLoginFailuresAreGeneric(t *testing.T) { // AUT-02, AUT-03
	co := newCompany(t, superadmin(t))
	c := newClient(t)

	status, wrongPwd := c.login(co.agentEmail, "Clave-Incorrecta-123")
	expect(t, "contraseña incorrecta", status, wrongPwd, http.StatusUnauthorized, "credenciales_invalidas")
	status, unknown := c.login("nadie@empresa.test", "Clave-Incorrecta-123")
	expect(t, "email inexistente", status, unknown, http.StatusUnauthorized, "credenciales_invalidas")

	status, code := c.call(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "no-es-email", "password": ""}, nil)
	expect(t, "formato inválido", status, code, http.StatusUnprocessableEntity, "validacion")

	status, code = c.call(http.MethodPost, "/api/v1/auth/login", map[string]any{"email": "a@b.pe", "password": "x", "extra": 1}, nil)
	expect(t, "campo desconocido", status, code, http.StatusUnprocessableEntity, "json_invalido")
}

func TestAccountLockout(t *testing.T) { // AUT-04
	co := newCompany(t, superadmin(t))
	c := newClient(t)
	for range 5 {
		status, code := c.login(co.agentEmail, "Clave-Incorrecta-123")
		expect(t, "intento fallido", status, code, http.StatusUnauthorized, "credenciales_invalidas")
	}
	status, code := c.login(co.agentEmail, userPassword)
	expect(t, "cuenta bloqueada con contraseña correcta", status, code, http.StatusUnauthorized, "credenciales_invalidas")

	var locked bool
	if err := env.owner.QueryRow(context.Background(),
		"SELECT bloqueado_hasta > now() FROM usuarios WHERE email = $1", co.agentEmail).Scan(&locked); err != nil || !locked {
		t.Errorf("la cuenta debería estar bloqueada: %v", err)
	}
	var audited int
	_ = env.owner.QueryRow(context.Background(),
		"SELECT count(*) FROM auditoria a JOIN usuarios u ON u.id = a.usuario_id WHERE u.email = $1 AND a.accion = 'cuenta_bloqueada'",
		co.agentEmail).Scan(&audited)
	if audited != 1 {
		t.Errorf("se esperaba un registro de bloqueo en auditoría, hay %d", audited)
	}
}

func TestProtectedRoutesRequireValidToken(t *testing.T) { // AUT-06, AUT-07
	c := newClient(t)
	status, code := c.call(http.MethodGet, "/api/v1/usuarios", nil, nil)
	expect(t, "sin token", status, code, http.StatusUnauthorized, "no_autenticado")

	root := superadmin(t)
	root.token = root.token[:len(root.token)-3] + "abc"
	status, code = root.call(http.MethodGet, "/api/v1/empresas", nil, nil)
	expect(t, "token alterado", status, code, http.StatusUnauthorized, "no_autenticado")
}

func TestRefreshRotationAndReuseDetection(t *testing.T) { // AUT-08
	co := newCompany(t, superadmin(t))
	c := newClient(t)
	c.mustLogin(co.agentEmail, userPassword)
	first := c.refreshCookie()

	var s sessionResp
	status, code := c.call(http.MethodPost, "/api/v1/auth/refresh", nil, &s)
	expect(t, "refresh", status, code, http.StatusOK, "")
	second := c.refreshCookie()
	if second == "" || second == first || s.AccessToken == "" {
		t.Fatal("el refresh debe rotar la cookie y emitir un access token")
	}

	attacker := newClient(t)
	attacker.setRefreshCookie(first)
	status, code = attacker.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "reutilizar token rotado", status, code, http.StatusUnauthorized, "sesion_invalida")

	status, code = c.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "la familia queda revocada", status, code, http.StatusUnauthorized, "sesion_invalida")
}

func TestRefreshWithoutCookieAndLogout(t *testing.T) {
	c := newClient(t)
	status, code := c.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "refresh sin cookie", status, code, http.StatusUnauthorized, "sesion_invalida")

	co := newCompany(t, superadmin(t))
	c.mustLogin(co.agentEmail, userPassword)
	status, _ = c.call(http.MethodPost, "/api/v1/auth/logout", nil, nil)
	expect(t, "logout", status, "", http.StatusNoContent, "")
	if c.refreshCookie() != "" {
		t.Error("el logout debe borrar la cookie")
	}

	stolen := newClient(t)
	stolen.mustLogin(co.agentEmail, userPassword)
	cookie := stolen.refreshCookie()
	_, _ = stolen.call(http.MethodPost, "/api/v1/auth/logout", nil, nil)
	stolen.setRefreshCookie(cookie)
	status, code = stolen.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "refresh tras logout", status, code, http.StatusUnauthorized, "sesion_invalida")
}

func TestExpiredRefreshToken(t *testing.T) {
	co := newCompany(t, superadmin(t))
	c := newClient(t)
	c.mustLogin(co.agentEmail, userPassword)
	if _, err := env.owner.Exec(context.Background(),
		"UPDATE refresh_tokens SET expira_en = $1 WHERE usuario_id = $2", time.Now().Add(-time.Minute), co.agentID); err != nil {
		t.Fatal(err)
	}
	status, code := c.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "refresh expirado", status, code, http.StatusUnauthorized, "sesion_invalida")
}

func TestRoleBasedAccess(t *testing.T) { // AUT-09, AUT-10, AUT-11
	root := superadmin(t)
	co := newCompany(t, root)

	cases := []struct {
		label  string
		c      *client
		method string
		path   string
		status int
	}{
		{"agente a usuarios", co.agent, http.MethodGet, "/api/v1/usuarios", http.StatusForbidden},
		{"agente a empresas", co.agent, http.MethodGet, "/api/v1/empresas", http.StatusForbidden},
		{"admin a empresas", co.admin, http.MethodGet, "/api/v1/empresas", http.StatusForbidden},
		{"superadmin a usuarios de empresa", root, http.MethodGet, "/api/v1/usuarios", http.StatusForbidden},
		{"admin a usuarios", co.admin, http.MethodGet, "/api/v1/usuarios", http.StatusOK},
		{"superadmin a empresas", root, http.MethodGet, "/api/v1/empresas", http.StatusOK},
	}
	for _, c := range cases {
		status, _ := c.c.call(c.method, c.path, nil, nil)
		if status != c.status {
			t.Errorf("%s: %d, se esperaba %d", c.label, status, c.status)
		}
	}
}

func TestFirstLoginRequiresPasswordChange(t *testing.T) { // AUT-12, AUT-13
	root := superadmin(t)
	co := newCompany(t, root)

	var created createdUser
	status, code := co.admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
		"nombre": "Nuevo Agente", "email": "nuevo" + uniqueSuffix() + "@empresa.test", "rol": "agente",
	}, &created)
	expect(t, "alta", status, code, http.StatusCreated, "")

	c := newClient(t)
	s := c.mustLogin(created.Usuario.Email, created.PasswordTemporal)
	if !s.User.DebeCambiarPassword {
		t.Error("el primer ingreso debe exigir cambio de contraseña")
	}
	status, code = c.call(http.MethodGet, "/api/v1/auth/me", nil, nil)
	expect(t, "me permitido", status, code, http.StatusOK, "")

	status, code = c.call(http.MethodPost, "/api/v1/auth/password", map[string]string{"actual": created.PasswordTemporal, "nueva": "corta"}, nil)
	expect(t, "contraseña débil", status, code, http.StatusUnprocessableEntity, "validacion")
	status, code = c.call(http.MethodPost, "/api/v1/auth/password", map[string]string{"actual": "incorrecta", "nueva": userPassword}, nil)
	expect(t, "contraseña actual incorrecta", status, code, http.StatusUnprocessableEntity, "validacion")
	status, code = c.call(http.MethodPost, "/api/v1/auth/password", map[string]string{"actual": created.PasswordTemporal, "nueva": created.PasswordTemporal}, nil)
	expect(t, "misma contraseña", status, code, http.StatusUnprocessableEntity, "validacion")

	other := newClient(t)
	other.mustLogin(created.Usuario.Email, created.PasswordTemporal)
	c.changePassword(created.PasswordTemporal, userPassword)

	status, code = other.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	expect(t, "las otras sesiones se cierran", status, code, http.StatusUnauthorized, "sesion_invalida")
}

func TestPendingPasswordChangeBlocksBusinessRoutes(t *testing.T) { // AUT-12
	root := superadmin(t)
	suffix := uniqueSuffix()
	var created createdTenant
	status, code := root.call(http.MethodPost, "/api/v1/empresas", map[string]string{
		"nombre": "Pendiente " + suffix, "ruc": randomRUC(), "adminNombre": "Admin", "adminEmail": "p" + suffix + "@empresa.test",
	}, &created)
	expect(t, "alta", status, code, http.StatusCreated, "")

	admin := newClient(t)
	admin.mustLogin(created.AdminEmail, created.PasswordTemporal)
	status, code = admin.call(http.MethodGet, "/api/v1/usuarios", nil, nil)
	expect(t, "ruta de negocio con cambio pendiente", status, code, http.StatusForbidden, "cambio_password_requerido")
}

func TestCompanyAdminActionsCloseSessions(t *testing.T) { // AUT-14
	root := superadmin(t)
	cases := []struct {
		label string
		act   func(t *testing.T, co company)
	}{
		{"restablecer contraseña", func(t *testing.T, co company) {
			status, code := co.admin.call(http.MethodPost, "/api/v1/usuarios/"+co.agentID+"/restablecer", nil, nil)
			expect(t, "restablecer", status, code, http.StatusOK, "")
		}},
		// Se reactiva antes de probar el refresh: así el rechazo se debe a la
		// sesión revocada y no a que el usuario esté inactivo.
		{"desactivar y reactivar", func(t *testing.T, co company) {
			for _, active := range []bool{false, true} {
				status, code := co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]any{"activo": active}, nil)
				expect(t, "cambiar estado", status, code, http.StatusOK, "")
			}
		}},
		{"cambiar rol", func(t *testing.T, co company) {
			status, code := co.admin.call(http.MethodPatch, "/api/v1/usuarios/"+co.agentID, map[string]any{"rol": "admin_empresa"}, nil)
			expect(t, "cambiar rol", status, code, http.StatusOK, "")
		}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			co := newCompany(t, root)
			c.act(t, co)
			status, code := co.agent.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
			expect(t, "refresh de la sesión anterior", status, code, http.StatusUnauthorized, "sesion_invalida")
			status, code = co.admin.call(http.MethodPost, "/api/v1/auth/refresh", nil, nil)
			expect(t, "la sesión del administrador sigue", status, code, http.StatusOK, "")
		})
	}
}

func TestFinalCustomerHasNoAccess(t *testing.T) { // AUT-15
	co := newCompany(t, superadmin(t))

	// No existe un rol para el cliente final ni forma de asignarlo.
	for _, role := range []string{"cliente", "contacto", ""} {
		status, code := co.admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
			"nombre": "Cliente Final", "email": "cliente" + uniqueSuffix() + "@correo.test", "rol": role,
		}, nil)
		expect(t, "rol "+role, status, code, http.StatusUnprocessableEntity, "validacion")
	}

	// No hay registro público.
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/registro"} {
		status, code := newClient(t).call(http.MethodPost, path, map[string]string{"email": "cliente@correo.test"}, nil)
		expect(t, path, status, code, http.StatusNotFound, "no_encontrado")
	}
}
