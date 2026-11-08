// Package service implementa la gestión de usuarios de una empresa.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"backend/internal/modules/users/domain"
	"backend/internal/platform/apperr"
	"backend/internal/platform/audit"
	"backend/internal/platform/authz"
	"backend/internal/platform/database"
	"backend/internal/platform/security"
)

// Repository es la persistencia de usuarios.
type Repository interface {
	List(ctx context.Context, q database.Querier) ([]domain.User, error)
	ListAdmins(ctx context.Context, q database.Querier, tenantID uuid.UUID) ([]domain.User, error)
	Get(ctx context.Context, q database.Querier, id uuid.UUID) (domain.User, error)
	Insert(ctx context.Context, q database.Querier, tenantID uuid.UUID, in domain.CreateInput, hash string) (domain.User, error)
	Update(ctx context.Context, q database.Querier, id uuid.UUID, in domain.UpdateInput) (domain.User, error)
	SetPassword(ctx context.Context, q database.Querier, id uuid.UUID, hash string, mustChange bool) error
}

// PasswordHasher genera hashes de contraseña.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// SessionRevoker cierra las sesiones de un usuario (lo implementa el módulo auth).
type SessionRevoker interface {
	RevokeAllForUser(ctx context.Context, q database.Querier, userID uuid.UUID) error
}

// Auditor registra acciones en la auditoría.
type Auditor interface {
	Record(ctx context.Context, q database.Querier, e audit.Entry) error
}

// Service implementa los casos de uso de usuarios.
type Service struct {
	tx      database.TxRunner
	repo    Repository
	hasher  PasswordHasher
	revoker SessionRevoker
	audit   Auditor
}

// New crea el servicio.
func New(tx database.TxRunner, repo Repository, hasher PasswordHasher, revoker SessionRevoker, auditor Auditor) *Service {
	return &Service{tx: tx, repo: repo, hasher: hasher, revoker: revoker, audit: auditor}
}

// List devuelve los usuarios de la empresa del administrador.
func (s *Service) List(ctx context.Context, p authz.Principal) ([]domain.User, error) {
	var users []domain.User
	err := s.tx.InTx(ctx, p.Scope(), func(q database.Querier) error {
		var err error
		users, err = s.repo.List(ctx, q)
		return err
	})
	return users, err
}

// ListAdmins devuelve los administradores de una empresa (uso del superadmin).
func (s *Service) ListAdmins(ctx context.Context, tenantID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		var err error
		users, err = s.repo.ListAdmins(ctx, q, tenantID)
		return err
	})
	return users, err
}

// Create crea un usuario en la empresa del administrador.
func (s *Service) Create(ctx context.Context, p authz.Principal, in domain.CreateInput) (domain.Created, error) {
	var created domain.Created
	err := s.tx.InTx(ctx, p.Scope(), func(q database.Querier) error {
		var err error
		created, err = s.CreateInTx(ctx, q, p.TenantID, in, p.UserID)
		return err
	})
	return created, err
}

// CreateInTx crea un usuario con contraseña temporal dentro de una transacción
// existente. El módulo de empresas lo usa para crear al primer administrador.
func (s *Service) CreateInTx(ctx context.Context, q database.Querier, tenantID uuid.UUID, in domain.CreateInput, actorID uuid.UUID) (domain.Created, error) {
	in = in.Normalize()
	if err := in.Validate(); err != nil {
		return domain.Created{}, err
	}
	temp, hash, err := s.temporaryPassword()
	if err != nil {
		return domain.Created{}, err
	}
	user, err := s.repo.Insert(ctx, q, tenantID, in, hash)
	if err != nil {
		return domain.Created{}, err
	}
	err = s.audit.Record(ctx, q, audit.Entry{
		TenantID: tenantID, UserID: actorID, Action: audit.ActionUserCreated, Entity: "usuario", EntityID: user.ID,
		Detail: map[string]any{"rol": user.Rol},
	})
	return domain.Created{Usuario: user, PasswordTemporal: temp}, err
}

// Update cambia nombre, rol o estado de un usuario de la empresa. Si cambia el
// acceso (rol o estado) se cierran sus sesiones.
func (s *Service) Update(ctx context.Context, p authz.Principal, id uuid.UUID, in domain.UpdateInput) (domain.User, error) {
	in = in.Normalize()
	if err := in.Validate(); err != nil {
		return domain.User{}, err
	}
	if id == p.UserID && (in.Rol != nil || in.Activo != nil) {
		return domain.User{}, domain.ErrSelfChange
	}

	var updated domain.User
	err := s.tx.InTx(ctx, p.Scope(), func(q database.Querier) error {
		current, err := s.repo.Get(ctx, q, id)
		if err != nil {
			return notFound(err)
		}
		if updated, err = s.repo.Update(ctx, q, id, in); err != nil {
			return err
		}
		if in.TouchesAccess(current) {
			if err := s.revoker.RevokeAllForUser(ctx, q, id); err != nil {
				return err
			}
		}
		return s.audit.Record(ctx, q, audit.Entry{
			TenantID: current.TenantID, UserID: p.UserID, Action: audit.ActionUserUpdated, Entity: "usuario", EntityID: id,
		})
	})
	return updated, err
}

// ResetPassword genera una contraseña temporal para el usuario, lo obliga a
// cambiarla en su próximo ingreso y cierra sus sesiones.
func (s *Service) ResetPassword(ctx context.Context, p authz.Principal, id uuid.UUID) (string, error) {
	var temp string
	err := s.tx.InTx(ctx, p.Scope(), func(q database.Querier) error {
		target, err := s.repo.Get(ctx, q, id)
		if err != nil {
			return notFound(err)
		}
		if !domain.CanReset(p, target) {
			return domain.ErrResetForbidden
		}
		var hash string
		if temp, hash, err = s.temporaryPassword(); err != nil {
			return err
		}
		if err := s.repo.SetPassword(ctx, q, id, hash, true); err != nil {
			return err
		}
		if err := s.revoker.RevokeAllForUser(ctx, q, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{
			TenantID: target.TenantID, UserID: p.UserID, Action: audit.ActionPasswordReset, Entity: "usuario", EntityID: id,
		})
	})
	return temp, err
}

func (s *Service) temporaryPassword() (plain, hash string, err error) {
	if plain, err = security.TemporaryPassword(); err != nil {
		return "", "", err
	}
	hash, err = s.hasher.Hash(plain)
	return plain, hash, err
}

func notFound(err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return apperr.ErrNotFound
	}
	return err
}
