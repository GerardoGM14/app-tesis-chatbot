/** Error del API con el formato { error: { code, message, fields } } del backend. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
    readonly fields: Record<string, string> = {},
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/** Puente con la capa de aplicación, que es dueña de la sesión. */
export interface AuthBridge {
  getAccessToken: () => string | null;
  refreshAccessToken: () => Promise<string | null>;
  onSessionExpired: () => void;
}

let bridge: AuthBridge = {
  getAccessToken: () => null,
  refreshAccessToken: () => Promise.resolve(null),
  onSessionExpired: () => undefined,
};

export function configureAuth(next: AuthBridge): void {
  bridge = next;
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
  body?: unknown;
  /** Si es false no se envía el access token ni se intenta renovar la sesión. */
  auth?: boolean;
  signal?: AbortSignal;
}

const API_BASE = '/api/v1';

interface ErrorBody {
  error?: { code?: string; message?: string; fields?: Record<string, string> };
}

async function toApiError(response: Response): Promise<ApiError> {
  let body: ErrorBody = {};
  try {
    body = (await response.json()) as ErrorBody;
  } catch {
    // Respuesta sin JSON (por ejemplo, el proxy devolvió HTML).
  }
  return new ApiError(
    response.status,
    body.error?.code ?? 'error_desconocido',
    body.error?.message ?? 'No se pudo completar la operación',
    body.error?.fields ?? {},
  );
}

async function send(path: string, options: RequestOptions, token: string | null): Promise<Response> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (options.body !== undefined) headers['Content-Type'] = 'application/json';
  if (token) headers.Authorization = `Bearer ${token}`;

  try {
    return await fetch(API_BASE + path, {
      method: options.method ?? 'GET',
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      credentials: 'same-origin',
      signal: options.signal,
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') throw error;
    throw new ApiError(0, 'sin_conexion', 'No hay conexión con el servidor');
  }
}

/**
 * Hace una petición al API. Si el access token expiró (401) intenta renovar la
 * sesión una vez y repite la petición; si no puede, avisa que la sesión expiró.
 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const useAuth = options.auth ?? true;
  let response = await send(path, options, useAuth ? bridge.getAccessToken() : null);

  if (response.status === 401 && useAuth) {
    const token = await bridge.refreshAccessToken();
    if (!token) {
      bridge.onSessionExpired();
      throw await toApiError(response);
    }
    response = await send(path, options, token);
  }

  if (!response.ok) throw await toApiError(response);
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}
