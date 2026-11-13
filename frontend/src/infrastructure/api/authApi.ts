import type { Session, SessionUser } from '@/domain/entities';
import { request } from '../http/client';

export const authApi = {
  login: (email: string, password: string) =>
    request<Session>('/auth/login', { method: 'POST', body: { email, password }, auth: false }),

  refresh: () => request<Session>('/auth/refresh', { method: 'POST', auth: false }),

  logout: () => request<undefined>('/auth/logout', { method: 'POST', auth: false }),

  me: () => request<SessionUser>('/auth/me'),

  changePassword: (actual: string, nueva: string) =>
    request<Session>('/auth/password', { method: 'POST', body: { actual, nueva } }),
};
