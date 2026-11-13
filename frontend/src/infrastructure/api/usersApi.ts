import type { CreatedUser, ListResponse, User } from '@/domain/entities';
import type { UserCreateValues } from '@/domain/schemas';
import { request } from '../http/client';

export interface UserPatch {
  nombre?: string;
  rol?: User['rol'];
  activo?: boolean;
}

export const usersApi = {
  list: () => request<ListResponse<User>>('/usuarios'),

  create: (values: UserCreateValues) => request<CreatedUser>('/usuarios', { method: 'POST', body: values }),

  update: (id: string, patch: UserPatch) =>
    request<User>(`/usuarios/${id}`, { method: 'PATCH', body: patch }),

  resetPassword: (id: string) =>
    request<{ passwordTemporal: string }>(`/usuarios/${id}/restablecer`, { method: 'POST' }),
};
