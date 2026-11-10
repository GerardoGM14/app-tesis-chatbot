package app_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"backend/internal/app"
	"backend/internal/bootstrap"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/security"
	"backend/internal/platform/validate"
)

// Las pruebas de integración necesitan PostgreSQL. TEST_DATABASE_URL debe
// apuntar a una base desechable con un usuario dueño (se borra el esquema public).
const (
	appPassword        = "clave-del-rol-app-para-pruebas"
	superadminEmail    = "root@plataforma.test"
	superadminInitial  = "Inicial-Plataforma-2026"
	superadminPassword = "Definitiva-Plataforma-2026"
	userPassword       = "Clave-Usuario-Final-2026"
)

var (
	env        *testEnv
	fastArgon2 = security.Argon2Params{Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
)

type testEnv struct {
	server *httptest.Server
	db     *database.DB
	owner  *pgx.Conn
}

func TestMain(m *testing.M) {
	ownerURL := os.Getenv("TEST_DATABASE_URL")
	if ownerURL == "" {
		fmt.Println("TEST_DATABASE_URL no definida: se omiten las pruebas de integración")
		os.Exit(0)
	}
	e, err := setup(ownerURL)
	if err != nil {
		fmt.Println("no se pudo preparar el entorno:", err)
		os.Exit(1)
	}
	env = e
	code := m.Run()
	e.server.Close()
	e.db.Close()
	_ = e.owner.Close(context.Background())
	os.Exit(code)
}

func setup(ownerURL string) (*testEnv, error) {
	ctx := context.Background()
	owner, err := pgx.Connect(ctx, ownerURL)
	if err != nil {
		return nil, err
	}
	if _, err := owner.Exec(ctx, "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"); err != nil {
		return nil, err
	}
	hasher := security.NewPasswordHasher(fastArgon2)
	if _, err := bootstrap.Run(ctx, owner, appPassword, bootstrap.Superadmin{
		Nombre: "Root", Email: superadminEmail, Password: superadminInitial,
	}, hasher); err != nil {
		return nil, err
	}

	appURL, err := url.Parse(ownerURL)
	if err != nil {
		return nil, err
	}
	appURL.User = url.UserPassword(bootstrap.AppRole, appPassword)
	db, err := database.Connect(ctx, appURL.String())
	if err != nil {
		return nil, err
	}

	cfg := config.Config{
		Env: "development", JWTSecret: []byte("secreto-de-pruebas-de-32-bytes!!"),
		AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: time.Hour, CookieSecure: false,
	}
	handler, err := app.New(cfg, db, slog.New(slog.NewTextHandler(io.Discard, nil)), app.Options{
		Argon2: fastArgon2, TenantCacheTTL: time.Nanosecond, DisableRateLimit: true,
	})
	if err != nil {
		return nil, err
	}
	return &testEnv{server: httptest.NewServer(handler), db: db, owner: owner}, nil
}

// client simula un navegador: guarda cookies y el access token.
type client struct {
	t     *testing.T
	http  *http.Client
	token string
}

func newClient(t *testing.T) *client {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	return &client{t: t, http: &http.Client{Jar: jar}}
}

type apiError struct {
	Error struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
	} `json:"error"`
}

// call hace la petición y decodifica la respuesta en out si es 2xx.
// Devuelve el código de estado y el código de error, si lo hay.
func (c *client) call(method, path string, body, out any) (int, string) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, env.server.URL+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out != nil {
			if err := json.Unmarshal(raw, out); err != nil {
				c.t.Fatalf("respuesta inválida de %s %s: %v (%s)", method, path, err, raw)
			}
		}
		return resp.StatusCode, ""
	}
	var e apiError
	_ = json.Unmarshal(raw, &e)
	return resp.StatusCode, e.Error.Code
}

type sessionResp struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID                  string   `json:"id"`
		Rol                 string   `json:"rol"`
		TenantID            *string  `json:"tenantId"`
		Permisos            []string `json:"permisos"`
		DebeCambiarPassword bool     `json:"debeCambiarPassword"`
	} `json:"user"`
}

func (c *client) login(email, password string) (int, string) {
	c.t.Helper()
	var s sessionResp
	status, code := c.call(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": email, "password": password}, &s)
	if status == http.StatusOK {
		c.token = s.AccessToken
	}
	return status, code
}

func (c *client) mustLogin(email, password string) sessionResp {
	c.t.Helper()
	var s sessionResp
	status, code := c.call(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": email, "password": password}, &s)
	if status != http.StatusOK {
		c.t.Fatalf("login de %s falló: %d %s", email, status, code)
	}
	c.token = s.AccessToken
	return s
}

func (c *client) changePassword(current, next string) {
	c.t.Helper()
	var s sessionResp
	status, code := c.call(http.MethodPost, "/api/v1/auth/password", map[string]string{"actual": current, "nueva": next}, &s)
	if status != http.StatusOK {
		c.t.Fatalf("cambio de contraseña falló: %d %s", status, code)
	}
	c.token = s.AccessToken
}

// refreshCookie devuelve el valor actual de la cookie de sesión.
func (c *client) refreshCookie() string {
	u, _ := url.Parse(env.server.URL + "/api/v1/auth")
	for _, ck := range c.http.Jar.Cookies(u) {
		if ck.Name == "rt" {
			return ck.Value
		}
	}
	return ""
}

func (c *client) setRefreshCookie(value string) {
	u, _ := url.Parse(env.server.URL + "/api/v1/auth")
	c.http.Jar.SetCookies(u, []*http.Cookie{{Name: "rt", Value: value, Path: "/api/v1/auth"}})
}

// superadmin devuelve un cliente con sesión de superadmin y contraseña definitiva.
func superadmin(t *testing.T) *client {
	t.Helper()
	c := newClient(t)
	if status, _ := c.login(superadminEmail, superadminPassword); status == http.StatusOK {
		return c
	}
	c.mustLogin(superadminEmail, superadminInitial)
	c.changePassword(superadminInitial, superadminPassword)
	return c
}

type tenantResp struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado"`
}

type createdTenant struct {
	Empresa          tenantResp `json:"empresa"`
	AdminEmail       string     `json:"adminEmail"`
	PasswordTemporal string     `json:"passwordTemporal"`
}

type userResp struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Email    string `json:"email"`
	Rol      string `json:"rol"`
	Activo   bool   `json:"activo"`
	Nombre   string `json:"nombre"`
}

type createdUser struct {
	Usuario          userResp `json:"usuario"`
	PasswordTemporal string   `json:"passwordTemporal"`
}

// company es una empresa de prueba con un administrador y un agente listos.
type company struct {
	tenant     tenantResp
	admin      *client
	adminEmail string
	adminID    string
	agent      *client
	agentEmail string
	agentID    string
}

func uniqueSuffix() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	return fmt.Sprintf("%09d", n.Int64())
}

func randomRUC() string {
	first10 := "20" + uniqueSuffix()[:8]
	check, _ := validate.RUCCheckDigit(first10)
	return first10 + string(check)
}

// newCompany da de alta una empresa con su admin y un agente, ambos con
// contraseña definitiva.
func newCompany(t *testing.T, root *client) company {
	t.Helper()
	suffix := uniqueSuffix()
	var created createdTenant
	status, code := root.call(http.MethodPost, "/api/v1/empresas", map[string]string{
		"nombre": "Empresa " + suffix, "ruc": randomRUC(),
		"adminNombre": "Admin " + suffix, "adminEmail": "admin" + suffix + "@empresa.test",
	}, &created)
	if status != http.StatusCreated {
		t.Fatalf("alta de empresa falló: %d %s", status, code)
	}

	admin := newClient(t)
	adminSession := admin.mustLogin(created.AdminEmail, created.PasswordTemporal)
	admin.changePassword(created.PasswordTemporal, userPassword)

	agentEmail := "agente" + suffix + "@empresa.test"
	var agent createdUser
	if status, code := admin.call(http.MethodPost, "/api/v1/usuarios", map[string]string{
		"nombre": "Agente " + suffix, "email": agentEmail, "rol": "agente",
	}, &agent); status != http.StatusCreated {
		t.Fatalf("alta de agente falló: %d %s", status, code)
	}
	agentClient := newClient(t)
	agentClient.mustLogin(agentEmail, agent.PasswordTemporal)
	agentClient.changePassword(agent.PasswordTemporal, userPassword)

	return company{
		tenant: created.Empresa, admin: admin, adminEmail: created.AdminEmail, adminID: adminSession.User.ID,
		agent: agentClient, agentEmail: agentEmail, agentID: agent.Usuario.ID,
	}
}

func expect(t *testing.T, label string, gotStatus int, gotCode string, wantStatus int, wantCode string) {
	t.Helper()
	if gotStatus != wantStatus || (wantCode != "" && gotCode != wantCode) {
		t.Errorf("%s: %d %q, se esperaba %d %q", label, gotStatus, gotCode, wantStatus, wantCode)
	}
}
