package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"backend/internal/platform/authz"
)

const tokenIssuer = "plataforma-atencion"

// ErrInvalidToken indica un access token inválido o expirado.
var ErrInvalidToken = errors.New("token inválido")

type accessClaims struct {
	jwt.RegisteredClaims
	TenantID   string `json:"tid,omitempty"`
	Role       string `json:"rol"`
	MustChange bool   `json:"pwc,omitempty"`
}

// TokenManager emite y valida access tokens JWT firmados con HS256.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewTokenManager crea el gestor de tokens. El secreto debe tener al menos 32 bytes.
func NewTokenManager(secret []byte, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: secret, ttl: ttl, now: time.Now}
}

// Issue firma un access token para el usuario.
func (m *TokenManager) Issue(p authz.Principal) (string, time.Duration, error) {
	now := m.now()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   p.UserID.String(),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			ID:        uuid.NewString(),
		},
		Role:       string(p.Role),
		MustChange: p.MustChangePassword,
	}
	if p.TenantID != uuid.Nil {
		claims.TenantID = p.TenantID.String()
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return signed, m.ttl, err
}

// Parse valida la firma, el emisor y la expiración del token.
func (m *TokenManager) Parse(token string) (authz.Principal, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	var claims accessClaims
	if _, err := parser.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) { return m.secret, nil }); err != nil {
		return authz.Principal{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return authz.Principal{}, ErrInvalidToken
	}
	role := authz.Role(claims.Role)
	if !role.Valid() {
		return authz.Principal{}, ErrInvalidToken
	}
	p := authz.Principal{UserID: userID, Role: role, MustChangePassword: claims.MustChange}
	if role != authz.RoleSuperadmin {
		if p.TenantID, err = uuid.Parse(claims.TenantID); err != nil {
			return authz.Principal{}, ErrInvalidToken
		}
	}
	return p, nil
}

// RandomToken genera un token opaco de 256 bits en base64url.
func RandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken devuelve el SHA-256 del token; en la base solo se guarda el hash.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
