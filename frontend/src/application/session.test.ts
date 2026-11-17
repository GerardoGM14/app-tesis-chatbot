import { beforeEach, describe, expect, it } from 'vitest';
import { request } from '@/infrastructure/http/client';
import { apiError, makeUser, mockApi } from '@/test/utils';
import { bootstrapSession, installAuthBridge, logout, refreshAccessToken } from './session';
import { useSessionStore } from './stores/sessionStore';

const session = { accessToken: 'nuevo', expiresIn: 900, user: makeUser('agente') };

describe('sesión', () => {
  beforeEach(() => {
    useSessionStore.setState({ status: 'loading', accessToken: null, user: null });
    installAuthBridge();
  });

  it('restaura la sesión con la cookie de refresh', async () => {
    mockApi({ 'POST /auth/refresh': { body: session } });
    await bootstrapSession();
    expect(useSessionStore.getState()).toMatchObject({ status: 'authenticated', accessToken: 'nuevo' });
  });

  it('queda anónima si no hay sesión que restaurar', async () => {
    mockApi({ 'POST /auth/refresh': apiError(401, 'sesion_invalida', 'Expiró') });
    await bootstrapSession();
    expect(useSessionStore.getState()).toMatchObject({ status: 'anonymous', user: null });
  });

  it('comparte una sola renovación entre peticiones simultáneas', async () => {
    const calls = mockApi({ 'POST /auth/refresh': { body: session } });
    const tokens = await Promise.all([refreshAccessToken(), refreshAccessToken(), refreshAccessToken()]);
    expect(tokens).toEqual(['nuevo', 'nuevo', 'nuevo']);
    expect(calls.filter((c) => c.path === '/auth/refresh')).toHaveLength(1);
  });

  it('cierra la sesión local aunque falle el servidor', async () => {
    useSessionStore.getState().setSession(session);
    mockApi({ 'POST /auth/logout': apiError(500, 'error_interno', 'Falló') });
    await expect(logout()).rejects.toBeDefined();
    expect(useSessionStore.getState().status).toBe('anonymous');
  });

  it('limpia la sesión cuando una petición no puede renovarse', async () => {
    useSessionStore.getState().setSession(session);
    mockApi({
      'GET /usuarios': apiError(401, 'no_autenticado', 'Expiró'),
      'POST /auth/refresh': apiError(401, 'sesion_invalida', 'Expiró'),
    });
    await expect(request('/usuarios')).rejects.toBeDefined();
    expect(useSessionStore.getState().status).toBe('anonymous');
  });
});
