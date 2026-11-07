// Package service implementa los casos de uso de autenticación: iniciar y
// cerrar sesión, renovar la sesión y cambiar la contraseña.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"backend/internal/modules/auth/domain"
	"backend/internal/platform/apperr"
	"backend/internal/platform/audit"
	"backend/internal/platform/authz"
	"backend/internal/platform/database"
	"backend/internal/platform/security"
)

// Repository es la persistencia que necesita el servicio.
type Repository interface {
	FindUserByEmail(ctx context.Context, q database.Querier, email string) (domain.User, error)
	FindUserByID(ctx context.Context, q database.Querier, id uuid.UUID) (domain.User, error)
	UpdateFailedAttempts(ctx context.Context, q database.Querier, id uuid.UUID, attempts int, lockedUntil *time.Time) error
	RegisterSuccessfulLogin(ctx context.Context, q database.Querier, id uuid.UUID) error
	UpdatePassword(ctx context.Context, q database.Querier, id uuid.UUID, hash string, mustChange bool) error
	InsertRefreshToken(ctx context.Context, q database.Querier, t domain.RefreshToken) error
	FindRefreshTokenForUpdate(ctx context.Context, q database.Querier, hash []byte) (domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, q database.Querier, id uuid.UUID) error
	RevokeFamily(ctx context.Context, q database.Querier, family uuid.UUID) error
	RevokeAllForUser(ctx context.Context, q database.Querier, userID uuid.UUID) error
}

// PasswordHasher genera y verifica hashes de contraseña.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(encoded, password string) bool
}

// TokenIssuer firma access tokens.
type TokenIssuer interface {
	Issue(p authz.Principal) (string, time.Duration, error)
}

// Auditor registra acciones en la auditoría.
type Auditor interface {
	Record(ctx context.Context, q database.Querier, e audit.Entry) error
}

// Service implementa los casos de uso de autenticación.
type Service struct {
	tx         database.TxRunner
	repo       Repository
	hasher     PasswordHasher
	tokens     TokenIssuer
	audit      Auditor
	refreshTTL time.Duration
	now        func() time.Time
	dummyHash  string
}

// New crea el servicio.
func New(tx database.TxRunner, repo Repository, hasher PasswordHasher, tokens TokenIssuer, auditor Auditor, refreshTTL time.Duration) (*Service, error) {
	// Hash de referencia para que un email inexistente tarde lo mismo que uno válido.
	dummy, err := hasher.Hash("hash-de-referencia-para-tiempo-constante")
	if err != nil {
		return nil, err
	}
	return &Service{
		tx: tx, repo: repo, hasher: hasher, tokens: tokens, audit: auditor,
		refreshTTL: refreshTTL, now: time.Now, dummyHash: dummy,
	}, nil
}

// outcome separa el error que se informa al cliente (la transacción igual se
// confirma, para guardar intentos fallidos y auditoría) de un error interno
// (que provoca rollback).
type outcome struct {
	session   domain.Session
	clientErr error
}

func rejected(err error) (outcome, error) {
	return outcome{clientErr: domain.ErrInvalidCredentials}, err
}

func invalidSession(err error) (outcome, error) {
	return outcome{clientErr: domain.ErrInvalidSession}, err
}

// runOutcome ejecuta fn en una transacción de plataforma y separa los errores.
func (s *Service) runOutcome(ctx context.Context, fn func(q database.Querier) (outcome, error)) (domain.Session, error) {
	var out outcome
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		var err error
		out, err = fn(q)
		return err
	})
	if err != nil {
		return domain.Session{}, err
	}
	return out.session, out.clientErr
}

// Login valida las credenciales y abre una sesión nueva.
func (s *Service) Login(ctx context.Context, email, password string) (domain.Session, error) {
	return s.runOutcome(ctx, func(q database.Querier) (outcome, error) {
		return s.login(ctx, q, email, password)
	})
}

func (s *Service) login(ctx context.Context, q database.Querier, email, password string) (outcome, error) {
	user, err := s.repo.FindUserByEmail(ctx, q, email)
	if errors.Is(err, database.ErrNotFound) {
		s.hasher.Verify(s.dummyHash, password)
		return rejected(s.audit.Record(ctx, q, audit.Entry{
			Action: audit.ActionLoginFailed, Detail: map[string]any{"motivo": "email_desconocido"},
		}))
	}
	if err != nil {
		return outcome{}, err
	}

	now := s.now()
	switch {
	case user.IsLocked(now):
		s.hasher.Verify(s.dummyHash, password)
		return rejected(s.audit.Record(ctx, q, failedEntry(user, "cuenta_bloqueada")))
	case !s.hasher.Verify(user.PasswordHash, password):
		return rejected(s.registerFailure(ctx, q, user, now))
	case !user.CanLogin():
		return rejected(s.audit.Record(ctx, q, failedEntry(user, "usuario_o_empresa_inactivos")))
	}

	if err := s.repo.RegisterSuccessfulLogin(ctx, q, user.ID); err != nil {
		return outcome{}, err
	}
	session, err := s.openSession(ctx, q, user, uuid.New())
	if err != nil {
		return outcome{}, err
	}
	return outcome{session: session}, s.audit.Record(ctx, q, audit.Entry{
		TenantID: user.TenantID, UserID: user.ID, Action: audit.ActionLoginOK,
	})
}

func (s *Service) registerFailure(ctx context.Context, q database.Querier, user domain.User, now time.Time) error {
	attempts, lockedUntil := user.RegisterFailure(now)
	if err := s.repo.UpdateFailedAttempts(ctx, q, user.ID, attempts, lockedUntil); err != nil {
		return err
	}
	if lockedUntil != nil {
		return s.audit.Record(ctx, q, audit.Entry{TenantID: user.TenantID, UserID: user.ID, Action: audit.ActionAccountLocked})
	}
	return s.audit.Record(ctx, q, failedEntry(user, "password_incorrecto"))
}

func failedEntry(user domain.User, reason string) audit.Entry {
	return audit.Entry{TenantID: user.TenantID, UserID: user.ID, Action: audit.ActionLoginFailed, Detail: map[string]any{"motivo": reason}}
}

// Refresh rota el refresh token y emite un access token nuevo. Si se presenta
// un token ya revocado se asume robo y se revoca toda la familia.
func (s *Service) Refresh(ctx context.Context, rawToken string) (domain.Session, error) {
	return s.runOutcome(ctx, func(q database.Querier) (outcome, error) {
		return s.refresh(ctx, q, rawToken)
	})
}

func (s *Service) refresh(ctx context.Context, q database.Querier, rawToken string) (outcome, error) {
	current, err := s.repo.FindRefreshTokenForUpdate(ctx, q, security.HashToken(rawToken))
	if errors.Is(err, database.ErrNotFound) {
		return invalidSession(nil)
	}
	if err != nil {
		return outcome{}, err
	}

	if current.RevokedAt != nil {
		if err := s.repo.RevokeFamily(ctx, q, current.Family); err != nil {
			return outcome{}, err
		}
		return invalidSession(s.audit.Record(ctx, q, audit.Entry{UserID: current.UserID, Action: audit.ActionRefreshReuse}))
	}
	if !s.now().Before(current.ExpiresAt) {
		return invalidSession(nil)
	}

	user, err := s.repo.FindUserByID(ctx, q, current.UserID)
	if err != nil {
		return outcome{}, err
	}
	if !user.CanLogin() {
		return invalidSession(s.repo.RevokeFamily(ctx, q, current.Family))
	}

	if err := s.repo.RevokeRefreshToken(ctx, q, current.ID); err != nil {
		return outcome{}, err
	}
	session, err := s.openSession(ctx, q, user, current.Family)
	return outcome{session: session}, err
}

// Logout revoca la familia de sesiones del refresh token, si existe.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	return s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		current, err := s.repo.FindRefreshTokenForUpdate(ctx, q, security.HashToken(rawToken))
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.repo.RevokeFamily(ctx, q, current.Family); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{UserID: current.UserID, Action: audit.ActionLogout})
	})
}

// Me devuelve los datos del usuario autenticado.
func (s *Service) Me(ctx context.Context, p authz.Principal) (domain.View, error) {
	var view domain.View
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		user, err := s.repo.FindUserByID(ctx, q, p.UserID)
		if errors.Is(err, database.ErrNotFound) {
			return apperr.ErrUnauthorized
		}
		view = user.View()
		return err
	})
	return view, err
}

// ChangePassword cambia la contraseña del usuario autenticado, cierra sus
// otras sesiones y abre una nueva.
func (s *Service) ChangePassword(ctx context.Context, p authz.Principal, current, next string) (domain.Session, error) {
	var session domain.Session
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		user, err := s.repo.FindUserByID(ctx, q, p.UserID)
		if errors.Is(err, database.ErrNotFound) {
			return apperr.ErrUnauthorized
		}
		if err != nil {
			return err
		}
		if !s.hasher.Verify(user.PasswordHash, current) {
			return domain.ErrWrongPassword
		}
		if current == next {
			return domain.ErrSamePassword
		}
		if msg := security.CheckPasswordPolicy(next, user.Email); msg != "" {
			return apperr.Validation(map[string]string{"nueva": msg})
		}

		hash, err := s.hasher.Hash(next)
		if err != nil {
			return err
		}
		if err := s.repo.UpdatePassword(ctx, q, user.ID, hash, false); err != nil {
			return err
		}
		if err := s.repo.RevokeAllForUser(ctx, q, user.ID); err != nil {
			return err
		}
		user.DebeCambiarPassword = false
		if session, err = s.openSession(ctx, q, user, uuid.New()); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{TenantID: user.TenantID, UserID: user.ID, Action: audit.ActionPasswordChanged})
	})
	return session, err
}

func (s *Service) openSession(ctx context.Context, q database.Querier, user domain.User, family uuid.UUID) (domain.Session, error) {
	access, ttl, err := s.tokens.Issue(user.Principal())
	if err != nil {
		return domain.Session{}, err
	}
	raw, err := security.RandomToken()
	if err != nil {
		return domain.Session{}, err
	}
	expires := s.now().Add(s.refreshTTL)
	if err := s.repo.InsertRefreshToken(ctx, q, domain.RefreshToken{
		ID: uuid.New(), UserID: user.ID, Family: family, Hash: security.HashToken(raw), ExpiresAt: expires,
	}); err != nil {
		return domain.Session{}, err
	}
	return domain.Session{
		AccessToken: access, AccessTTL: ttl, RefreshToken: raw, RefreshExpiresAt: expires, User: user.View(),
	}, nil
}
