// Entidades del dominio. Reflejan los contratos del API del backend.

export type Role = 'superadmin' | 'admin_empresa' | 'agente';

export type Permission =
  | 'dashboard:ver'
  | 'perfil:editar'
  | 'empresas:gestionar'
  | 'metricas:globales'
  | 'bot:gestionar'
  | 'canal:gestionar'
  | 'usuarios:gestionar'
  | 'usuarios:restablecer'
  | 'conversaciones:atender';

export interface SessionUser {
  id: string;
  nombre: string;
  email: string;
  rol: Role;
  permisos: Permission[];
  tenantId: string | null;
  tenantNombre?: string;
  debeCambiarPassword: boolean;
}

export interface Session {
  accessToken: string;
  expiresIn: number;
  user: SessionUser;
}

export type TenantStatus = 'activo' | 'suspendido';

export interface Tenant {
  id: string;
  nombre: string;
  ruc: string;
  slug: string;
  estado: TenantStatus;
  usuarios: number;
  creadoEn: string;
}

export interface CreatedTenant {
  empresa: Tenant;
  adminEmail: string;
  passwordTemporal: string;
}

export interface User {
  id: string;
  tenantId: string;
  nombre: string;
  email: string;
  rol: Role;
  activo: boolean;
  debeCambiarPassword: boolean;
  ultimoAcceso: string | null;
  creadoEn: string;
}

export interface CreatedUser {
  usuario: User;
  passwordTemporal: string;
}

export interface ListResponse<T> {
  items: T[];
}
