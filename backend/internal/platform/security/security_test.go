package security

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"backend/internal/platform/authz"
)

// testArgon2 usa parámetros mínimos para que las pruebas sean rápidas.
var testArgon2 = Argon2Params{Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

var testSecret = []byte("0123456789abcdef0123456789abcdef")

func TestPasswordHashAndVerify(t *testing.T) {
	h := NewPasswordHasher(testArgon2)
	hash, err := h.Hash("ClaveSegura2026")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("formato inesperado: %s", hash)
	}
	if !h.Verify(hash, "ClaveSegura2026") {
		t.Error("la contraseña correcta debe verificarse")
	}
	if h.Verify(hash, "ClaveSegura2027") {
		t.Error("una contraseña incorrecta no debe verificarse")
	}
	other, _ := h.Hash("ClaveSegura2026")
	if other == hash {
		t.Error("dos hashes de la misma contraseña deben diferir por la sal")
	}
}

func TestVerifyRejectsMalformedHashes(t *testing.T) {
	h := NewPasswordHasher(testArgon2)
	malformed := []string{
		"",
		"texto-plano",
		"$bcrypt$v=19$m=64,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=18$m=64,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=x,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=64,t=1,p=1$***$aGFzaA",
		"$argon2id$v=19$m=64,t=1,p=1$c2FsdA$***",
		"$argon2id$v=19$m=64,t=1,p=1$c2FsdA$",
	}
	for _, m := range malformed {
		if h.Verify(m, "x") {
			t.Errorf("Verify(%q) debería fallar", m)
		}
	}
}

func TestTemporaryPassword(t *testing.T) {
	seen := map[string]bool{}
	for range 50 {
		p, err := TemporaryPassword()
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != 14 || strings.ContainsAny(p, "0O1lI") {
			t.Errorf("contraseña temporal inválida: %q", p)
		}
		if CheckPasswordPolicy(p, "a@b.pe") != "" {
			t.Errorf("la contraseña temporal debe cumplir la política: %q", p)
		}
		seen[p] = true
	}
	if len(seen) < 50 {
		t.Error("se generaron contraseñas repetidas")
	}
}

func TestPasswordPolicy(t *testing.T) {
	cases := map[string]bool{
		"corta":                    false,
		strings.Repeat("a", 129):   false,
		"Password123":              false,
		"aaaaaaaaaaaa":             false,
		"carlos-clave-2026":        false, // contiene el usuario del email
		"Monte-Chimbo-Azul-77":     true,
		"una frase larga y segura": true,
	}
	for pwd, ok := range cases {
		got := CheckPasswordPolicy(pwd, "carlos@empresa.pe") == ""
		if got != ok {
			t.Errorf("CheckPasswordPolicy(%q) aceptada=%v, se esperaba %v", pwd, got, ok)
		}
	}
}

func TestTokenRoundTrip(t *testing.T) {
	m := NewTokenManager(testSecret, 15*time.Minute)
	want := authz.Principal{UserID: uuid.New(), TenantID: uuid.New(), Role: authz.RoleAgente, MustChangePassword: true}

	token, ttl, err := m.Issue(want)
	if err != nil || ttl != 15*time.Minute {
		t.Fatalf("Issue falló: %v", err)
	}
	got, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse falló: %v", err)
	}
	if got != want {
		t.Errorf("principal distinto: %+v vs %+v", got, want)
	}
}

func TestSuperadminTokenHasNoTenant(t *testing.T) {
	m := NewTokenManager(testSecret, time.Minute)
	token, _, _ := m.Issue(authz.Principal{UserID: uuid.New(), Role: authz.RoleSuperadmin})
	p, err := m.Parse(token)
	if err != nil || p.TenantID != uuid.Nil {
		t.Errorf("el superadmin no debe tener tenant: %+v, %v", p, err)
	}
}

func TestParseRejectsInvalidTokens(t *testing.T) {
	m := NewTokenManager(testSecret, time.Minute)
	valid, _, _ := m.Issue(authz.Principal{UserID: uuid.New(), TenantID: uuid.New(), Role: authz.RoleAgente})

	expired := NewTokenManager(testSecret, time.Minute)
	expired.now = func() time.Time { return time.Now().Add(-time.Hour) }
	expiredToken, _, _ := expired.Issue(authz.Principal{UserID: uuid.New(), TenantID: uuid.New(), Role: authz.RoleAgente})

	otherKey := NewTokenManager([]byte("otra-clave-distinta-de-32-bytes!!"), time.Minute)
	foreign, _, _ := otherKey.Issue(authz.Principal{UserID: uuid.New(), TenantID: uuid.New(), Role: authz.RoleAgente})

	none := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": uuid.NewString(), "rol": "superadmin"})
	noneToken, _ := none.SignedString(jwt.UnsafeAllowNoneSignatureType)

	badRole := signRaw(t, jwt.MapClaims{"sub": uuid.NewString(), "rol": "root", "iss": jwtIssuer, "exp": time.Now().Add(time.Minute).Unix()})
	badSubject := signRaw(t, jwt.MapClaims{"sub": "x", "rol": "agente", "iss": jwtIssuer, "exp": time.Now().Add(time.Minute).Unix()})
	missingTenant := signRaw(t, jwt.MapClaims{"sub": uuid.NewString(), "rol": "agente", "iss": jwtIssuer, "exp": time.Now().Add(time.Minute).Unix()})

	invalid := map[string]string{
		"vacío":             "",
		"alterado":          valid[:len(valid)-2] + "xx",
		"expirado":          expiredToken,
		"otra clave":        foreign,
		"algoritmo none":    noneToken,
		"rol inválido":      badRole,
		"sujeto inválido":   badSubject,
		"agente sin tenant": missingTenant,
	}
	for name, token := range invalid {
		if _, err := m.Parse(token); err == nil {
			t.Errorf("token %s debería rechazarse", name)
		}
	}
}

func signRaw(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRandomTokenAndHash(t *testing.T) {
	a, _ := RandomToken()
	b, _ := RandomToken()
	if a == b || len(a) != 43 {
		t.Errorf("tokens inesperados: %q %q", a, b)
	}
	if !bytes.Equal(HashToken(a), HashToken(a)) || bytes.Equal(HashToken(a), HashToken(b)) {
		t.Error("el hash debe ser determinista y distinto por token")
	}
}
