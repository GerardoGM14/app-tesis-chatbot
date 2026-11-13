import type { CreatedTenant, ListResponse, Tenant, User } from '@/domain/entities';
import type { TenantCreateValues } from '@/domain/schemas';
import { request } from '../http/client';

export const tenantsApi = {
  list: (search = '') => {
    const query = search ? '?q=' + encodeURIComponent(search) : '';
    return request<ListResponse<Tenant>>('/empresas' + query);
  },

  create: (values: TenantCreateValues) =>
    request<CreatedTenant>('/empresas', { method: 'POST', body: values }),

  rename: (id: string, nombre: string) =>
    request<Tenant>(`/empresas/${id}`, { method: 'PATCH', body: { nombre } }),

  suspend: (id: string) => request<Tenant>(`/empresas/${id}/suspender`, { method: 'POST' }),

  reactivate: (id: string) => request<Tenant>(`/empresas/${id}/reactivar`, { method: 'POST' }),

  admins: (id: string) => request<ListResponse<User>>(`/empresas/${id}/administradores`),
};
