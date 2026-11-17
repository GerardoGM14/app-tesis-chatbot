import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import { createMemoryRouter, RouterProvider } from 'react-router';
import { Toaster } from 'sonner';
import { vi } from 'vitest';
import { routes } from '@/app/router';
import { useSessionStore } from '@/application/stores/sessionStore';
import type { Permission, Role, SessionUser } from '@/domain/entities';

const PERMISSIONS: Record<Role, Permission[]> = {
  superadmin: [
    'dashboard:ver',
    'perfil:editar',
    'empresas:gestionar',
    'metricas:globales',
    'usuarios:restablecer',
  ],
  admin_empresa: [
    'dashboard:ver',
    'perfil:editar',
    'bot:gestionar',
    'canal:gestionar',
    'usuarios:gestionar',
    'usuarios:restablecer',
    'conversaciones:atender',
  ],
  agente: ['dashboard:ver', 'perfil:editar', 'conversaciones:atender'],
};

export function makeUser(rol: Role, overrides: Partial<SessionUser> = {}): SessionUser {
  return {
    id: `id-${rol}`,
    nombre: rol === 'superadmin' ? 'Root Plataforma' : 'Lucía Ramos',
    email: `${rol}@empresa.test`,
    rol,
    permisos: PERMISSIONS[rol],
    tenantId: rol === 'superadmin' ? null : 'tenant-1',
    tenantNombre: rol === 'superadmin' ? undefined : 'Ferretería El Progreso',
    debeCambiarPassword: false,
    ...overrides,
  };
}

export function signIn(user: SessionUser): void {
  useSessionStore.getState().setSession({ accessToken: 'token-de-prueba', expiresIn: 900, user });
}

export function signOut(): void {
  useSessionStore.getState().clear();
}

interface MockResponse {
  status?: number;
  body?: unknown;
}

type Handler = (body: unknown, url: URL) => MockResponse;

export interface RecordedCall {
  method: string;
  path: string;
  body: unknown;
}

function requestUrl(input: string | URL | Request): string {
  if (typeof input === 'string') return input;
  if (input instanceof URL) return input.href;
  return input.url;
}

/**
 * Simula el API. Las claves son "MÉTODO /ruta" sin el prefijo /api/v1 ni la
 * query string. Devuelve las llamadas registradas para hacer aserciones.
 */
export function mockApi(handlers: Record<string, Handler | MockResponse>): RecordedCall[] {
  const calls: RecordedCall[] = [];
  vi.spyOn(globalThis, 'fetch').mockImplementation((input, init) => {
    const url = new URL(requestUrl(input), 'https://plataforma.test');
    const method = init?.method ?? 'GET';
    const path = url.pathname.replace(/^\/api\/v1/, '');
    const body: unknown = typeof init?.body === 'string' ? JSON.parse(init.body) : undefined;
    calls.push({ method, path, body });

    const handler = handlers[`${method} ${path}`];
    const result =
      typeof handler === 'function' ? handler(body, url) : (handler ?? { status: 404, body: notFound });
    const status = result.status ?? 200;
    return Promise.resolve(
      new Response(status === 204 ? null : JSON.stringify(result.body ?? {}), {
        status,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
  });
  return calls;
}

const notFound = { error: { code: 'no_encontrado', message: 'El recurso no existe' } };

export function apiError(status: number, code: string, message: string, fields?: Record<string, string>) {
  return { status, body: { error: { code, message, fields } } };
}

export function renderApp(path: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  const view = render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster />
    </QueryClientProvider>,
  );
  return { ...view, router };
}
