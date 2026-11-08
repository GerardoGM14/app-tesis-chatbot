// Package service implementa el alta y administración de empresas.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"backend/internal/modules/tenants/domain"
	usersdomain "backend/internal/modules/users/domain"
	"backend/internal/platform/apperr"
	"backend/internal/platform/audit"
	"backend/internal/platform/authz"
	"backend/internal/platform/database"
)

// Repository es la persistencia de empresas.
type Repository interface {
	List(ctx context.Context, q database.Querier, search string) ([]domain.Tenant, error)
	Get(ctx context.Context, q database.Querier, id uuid.UUID) (domain.Tenant, error)
	Insert(ctx context.Context, q database.Querier, in domain.CreateInput) (domain.Tenant, error)
	UpdateName(ctx context.Context, q database.Querier, id uuid.UUID, nombre string) error
	SetStatus(ctx context.Context, q database.Querier, id uuid.UUID, estado string) error
	IsActive(ctx context.Context, q database.Querier, id uuid.UUID) (bool, error)
}

// AdminProvisioner crea el primer administrador (lo implementa el módulo de usuarios).
type AdminProvisioner interface {
	CreateInTx(ctx context.Context, q database.Querier, tenantID uuid.UUID, in usersdomain.CreateInput, actorID uuid.UUID) (usersdomain.Created, error)
}

// SessionRevoker cierra las sesiones de una empresa (lo implementa el módulo auth).
type SessionRevoker interface {
	RevokeAllForTenant(ctx context.Context, q database.Querier, tenantID uuid.UUID) error
}

// Auditor registra acciones en la auditoría.
type Auditor interface {
	Record(ctx context.Context, q database.Querier, e audit.Entry) error
}

// Service implementa los casos de uso de empresas. Todas las operaciones
// corren con alcance de plataforma: solo el superadmin llega aquí.
type Service struct {
	tx      database.TxRunner
	repo    Repository
	admins  AdminProvisioner
	revoker SessionRevoker
	audit   Auditor
}

// New crea el servicio.
func New(tx database.TxRunner, repo Repository, admins AdminProvisioner, revoker SessionRevoker, auditor Auditor) *Service {
	return &Service{tx: tx, repo: repo, admins: admins, revoker: revoker, audit: auditor}
}

// List devuelve las empresas.
func (s *Service) List(ctx context.Context, search string) ([]domain.Tenant, error) {
	var tenants []domain.Tenant
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		var err error
		tenants, err = s.repo.List(ctx, q, search)
		return err
	})
	return tenants, err
}

// Get devuelve una empresa.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		var err error
		tenant, err = s.repo.Get(ctx, q, id)
		return notFound(err)
	})
	return tenant, err
}

// Create da de alta la empresa y su primer administrador en una sola transacción.
func (s *Service) Create(ctx context.Context, p authz.Principal, in domain.CreateInput) (domain.Created, error) {
	in = in.Normalize()
	if err := in.Validate(); err != nil {
		return domain.Created{}, err
	}

	var created domain.Created
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		tenant, err := s.repo.Insert(ctx, q, in)
		if err != nil {
			return err
		}
		admin, err := s.admins.CreateInTx(ctx, q, tenant.ID, usersdomain.CreateInput{
			Nombre: in.AdminNombre, Email: in.AdminEmail, Rol: authz.RoleAdminEmpresa,
		}, p.UserID)
		if err != nil {
			return prefixAdminFields(err)
		}
		tenant.Usuarios = 1
		created = domain.Created{Empresa: tenant, AdminEmail: admin.Usuario.Email, PasswordTemporal: admin.PasswordTemporal}
		return s.audit.Record(ctx, q, audit.Entry{
			TenantID: tenant.ID, UserID: p.UserID, Action: audit.ActionTenantCreated, Entity: "empresa", EntityID: tenant.ID,
		})
	})
	return created, err
}

// Rename cambia el nombre de la empresa.
func (s *Service) Rename(ctx context.Context, p authz.Principal, id uuid.UUID, nombre string) (domain.Tenant, error) {
	if err := domain.ValidateName(nombre); err != nil {
		return domain.Tenant{}, err
	}
	var tenant domain.Tenant
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		if err := s.repo.UpdateName(ctx, q, id, nombre); err != nil {
			return notFound(err)
		}
		var err error
		if tenant, err = s.repo.Get(ctx, q, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{
			TenantID: id, UserID: p.UserID, Action: audit.ActionTenantUpdated, Entity: "empresa", EntityID: id,
		})
	})
	return tenant, err
}

// Suspend bloquea el acceso de la empresa y cierra las sesiones de sus usuarios.
func (s *Service) Suspend(ctx context.Context, p authz.Principal, id uuid.UUID) (domain.Tenant, error) {
	return s.setStatus(ctx, p, id, domain.EstadoSuspendido, audit.ActionTenantSuspended)
}

// Reactivate vuelve a habilitar la empresa.
func (s *Service) Reactivate(ctx context.Context, p authz.Principal, id uuid.UUID) (domain.Tenant, error) {
	return s.setStatus(ctx, p, id, domain.EstadoActivo, audit.ActionTenantReactivated)
}

func (s *Service) setStatus(ctx context.Context, p authz.Principal, id uuid.UUID, estado, action string) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		if err := s.repo.SetStatus(ctx, q, id, estado); err != nil {
			return notFound(err)
		}
		if estado == domain.EstadoSuspendido {
			if err := s.revoker.RevokeAllForTenant(ctx, q, id); err != nil {
				return err
			}
		}
		var err error
		if tenant, err = s.repo.Get(ctx, q, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{TenantID: id, UserID: p.UserID, Action: action, Entity: "empresa", EntityID: id})
	})
	return tenant, err
}

// IsActive indica si la empresa está activa. Lo usa el middleware TenantGuard.
func (s *Service) IsActive(ctx context.Context, id uuid.UUID) (bool, error) {
	var active bool
	err := s.tx.InTx(ctx, database.PlatformScope(), func(q database.Querier) error {
		var err error
		active, err = s.repo.IsActive(ctx, q, id)
		return err
	})
	return active, err
}

// prefixAdminFields renombra los campos de validación del administrador para
// que coincidan con el formulario de alta (adminNombre, adminEmail).
func prefixAdminFields(err error) error {
	appErr, ok := apperr.As(err)
	if !ok || appErr.Fields == nil {
		return err
	}
	fields := make(map[string]string, len(appErr.Fields))
	for k, v := range appErr.Fields {
		switch k {
		case "nombre":
			fields["adminNombre"] = v
		case "email":
			fields["adminEmail"] = v
		default:
			fields[k] = v
		}
	}
	return apperr.Validation(fields)
}

func notFound(err error) error {
	if errors.Is(err, database.ErrNotFound) {
		return apperr.ErrNotFound
	}
	return err
}
