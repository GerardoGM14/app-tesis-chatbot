import { authApi } from '@/infrastructure/api/authApi';
import { configureAuth } from '@/infrastructure/http/client';
import { useSessionStore } from './stores/sessionStore';

let inFlight: Promise<string | null> | null = null;

/**
 * Renueva el access token con la cookie de refresh. Si varias peticiones
 * reciben 401 a la vez, comparten una sola renovación: el backend rota el
 * refresh token y un segundo uso cerraría la sesión por seguridad.
 */
export function refreshAccessToken(): Promise<string | null> {
  inFlight ??= authApi
    .refresh()
    .then((session) => {
      useSessionStore.getState().setSession(session);
      return session.accessToken;
    })
    .catch(() => null)
    .finally(() => {
      inFlight = null;
    });
  return inFlight;
}

/** Conecta el cliente HTTP con el store de sesión. Se llama una vez al iniciar. */
export function installAuthBridge(): void {
  configureAuth({
    getAccessToken: () => useSessionStore.getState().accessToken,
    refreshAccessToken,
    onSessionExpired: () => {
      useSessionStore.getState().clear();
    },
  });
}

/** Restaura la sesión al abrir la aplicación usando la cookie de refresh. */
export async function bootstrapSession(): Promise<void> {
  const token = await refreshAccessToken();
  if (!token) useSessionStore.getState().clear();
}

/** Cierra la sesión en el servidor y limpia el estado local aunque falle la red. */
export async function logout(): Promise<void> {
  try {
    await authApi.logout();
  } finally {
    useSessionStore.getState().clear();
  }
}
