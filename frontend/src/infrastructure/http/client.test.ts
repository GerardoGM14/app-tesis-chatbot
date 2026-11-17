import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiError, mockApi } from '@/test/utils';
import { ApiError, configureAuth, request } from './client';

describe('cliente HTTP', () => {
  const bridge = {
    getAccessToken: vi.fn<() => string | null>(),
    refreshAccessToken: vi.fn<() => Promise<string | null>>(),
    onSessionExpired: vi.fn(),
  };

  beforeEach(() => {
    bridge.getAccessToken.mockReset().mockReturnValue('token-1');
    bridge.refreshAccessToken.mockReset();
    bridge.onSessionExpired.mockReset();
    configureAuth(bridge);
  });

  it('envía el token y el cuerpo JSON', async () => {
    mockApi({ 'POST /eco': (body) => ({ body: { recibido: body } }) });
    const result = await request<{ recibido: unknown }>('/eco', { method: 'POST', body: { a: 1 } });
    expect(result.recibido).toEqual({ a: 1 });
    const [, init] = vi.mocked(fetch).mock.calls[0]!;
    expect(new Headers(init?.headers).get('Authorization')).toBe('Bearer token-1');
  });

  it('no envía token cuando auth es false', async () => {
    mockApi({ 'GET /publico': { body: {} } });
    await request('/publico', { auth: false });
    const [, init] = vi.mocked(fetch).mock.calls[0]!;
    expect(new Headers(init?.headers).has('Authorization')).toBe(false);
  });

  it('devuelve undefined en 204', async () => {
    mockApi({ 'POST /vacio': { status: 204 } });
    await expect(request('/vacio', { method: 'POST' })).resolves.toBeUndefined();
  });

  it('convierte el error del backend en ApiError', async () => {
    mockApi({ 'POST /x': apiError(422, 'validacion', 'Revisa los datos', { email: 'Inválido' }) });
    const error = await request('/x', { method: 'POST', body: {} }).catch((e: unknown) => e);
    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 422, code: 'validacion', fields: { email: 'Inválido' } });
  });

  it('usa un mensaje genérico si la respuesta no es JSON', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('<html>', { status: 502 }));
    await expect(request('/x')).rejects.toMatchObject({ status: 502, code: 'error_desconocido' });
  });

  it('informa la falta de conexión', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));
    await expect(request('/x')).rejects.toMatchObject({ status: 0, code: 'sin_conexion' });
  });

  it('renueva la sesión una vez ante un 401 y repite la petición', async () => {
    bridge.refreshAccessToken.mockResolvedValue('token-2');
    let attempts = 0;
    mockApi({
      'GET /datos': () => {
        attempts += 1;
        return attempts === 1 ? apiError(401, 'no_autenticado', 'Expiró') : { body: { ok: true } };
      },
    });
    await expect(request('/datos')).resolves.toEqual({ ok: true });
    const [, retry] = vi.mocked(fetch).mock.calls[1]!;
    expect(new Headers(retry?.headers).get('Authorization')).toBe('Bearer token-2');
  });

  it('avisa que la sesión expiró si no puede renovarla', async () => {
    bridge.refreshAccessToken.mockResolvedValue(null);
    mockApi({ 'GET /datos': apiError(401, 'no_autenticado', 'Expiró') });
    await expect(request('/datos')).rejects.toMatchObject({ status: 401 });
    expect(bridge.onSessionExpired).toHaveBeenCalledOnce();
  });

  it('no intenta renovar en peticiones sin autenticación', async () => {
    mockApi({
      'POST /auth/login': apiError(401, 'credenciales_invalidas', 'Email o contraseña incorrectos'),
    });
    await expect(request('/auth/login', { method: 'POST', auth: false })).rejects.toMatchObject({
      code: 'credenciales_invalidas',
    });
    expect(bridge.refreshAccessToken).not.toHaveBeenCalled();
  });
});
