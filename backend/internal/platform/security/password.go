// Package security implementa el hash de contraseñas, los tokens de sesión y
// la política de contraseñas.
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params son los parámetros de Argon2id.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2 sigue la recomendación mínima de OWASP: 19 MiB, 2 iteraciones.
var DefaultArgon2 = Argon2Params{Memory: 19 * 1024, Iterations: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32}

// PasswordHasher genera y verifica hashes Argon2id en formato PHC.
type PasswordHasher struct {
	params Argon2Params
}

// NewPasswordHasher crea un hasher con los parámetros indicados.
func NewPasswordHasher(params Argon2Params) PasswordHasher {
	return PasswordHasher{params: params}
}

var b64 = base64.RawStdEncoding

// maxKeyLength acota la longitud de clave aceptada al verificar un hash.
const maxKeyLength = 128

// Hash devuelve el hash de la contraseña con una sal aleatoria.
func (h PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, h.params.Iterations, h.params.Memory, h.params.Parallelism, h.params.KeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.Memory, h.params.Iterations, h.params.Parallelism,
		b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// Verify compara la contraseña con el hash en tiempo constante.
func (h PasswordHasher) Verify(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := b64.DecodeString(parts[5])
	if err != nil || len(want) == 0 || len(want) > maxKeyLength {
		return false
	}
	keyLength := uint32(len(want)) //nolint:gosec // acotado por maxKeyLength
	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(got, want) == 1
}

const tempPasswordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

// TemporaryPassword genera una contraseña temporal de 14 caracteres sin
// caracteres ambiguos (0/O, 1/l/I).
func TemporaryPassword() (string, error) {
	const length = 14
	limit := big.NewInt(int64(len(tempPasswordAlphabet)))
	var b strings.Builder
	b.Grow(length)
	for range length {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		b.WriteByte(tempPasswordAlphabet[n.Int64()])
	}
	return b.String(), nil
}
