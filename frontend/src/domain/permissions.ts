import type { Permission, Role, SessionUser } from './entities';

export const ROLE_LABEL: Record<Role, string> = {
  superadmin: 'Superadministrador',
  admin_empresa: 'Administrador',
  agente: 'Agente',
};

/** Pantalla inicial de cada rol, según docs/01-arquitectura/seguridad/autenticacion-y-roles.md. */
const HOME_BY_ROLE: Record<Role, string> = {
  superadmin: '/empresas',
  admin_empresa: '/dashboard',
  agente: '/conversaciones',
};

export function homePathFor(role: Role): string {
  return HOME_BY_ROLE[role];
}

export function can(user: SessionUser | null | undefined, permission: Permission): boolean {
  return user?.permisos.includes(permission) ?? false;
}
