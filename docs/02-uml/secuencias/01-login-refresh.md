# Secuencia — Inicio de sesión y renovación de token

El access token vive en memoria del frontend; el refresh token solo viaja en una cookie `httpOnly`. Si un refresh token ya usado se presenta de nuevo, se revoca toda la familia.

```mermaid
sequenceDiagram
  autonumber
  actor U as Usuario
  participant W as Frontend
  participant A as Backend API
  participant D as PostgreSQL

  U->>W: Ingresa email y contraseña
  W->>A: POST /api/v1/auth/login
  A->>A: Rate limit por IP y email
  A->>D: Buscar usuario por email
  D-->>A: Usuario con password_hash
  A->>A: Verificar con Argon2id
  alt Credenciales inválidas
    A->>D: Incrementar intentos_fallidos
    A-->>W: 401 mensaje genérico
    W-->>U: Toast de error
  else Credenciales válidas
    A->>D: Guardar hash del refresh token y familia
    A-->>W: 200 access token + cookie refresh httpOnly
    W-->>U: Redirige al dashboard
  end

  Note over W,A: 15 minutos después el access token expira

  W->>A: GET /api/v1/conversaciones
  A-->>W: 401 token expirado
  W->>A: POST /api/v1/auth/refresh con cookie
  A->>D: Buscar refresh token por hash
  alt Token ya revocado
    A->>D: Revocar toda la familia
    A-->>W: 401 y borra cookie
    W-->>U: Vuelve al login
  else Token vigente
    A->>D: Revocar el usado y guardar uno nuevo
    A-->>W: 200 nuevo access token + nueva cookie
    W->>A: Reintenta GET /api/v1/conversaciones
    A-->>W: 200 datos
  end
```
