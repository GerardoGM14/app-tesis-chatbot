# Secuencia — Restablecer la contraseña de un usuario

Aplica solo al personal con cuenta. No hay recuperación por correo en esta versión: el administrador de la empresa restablece la contraseña de sus agentes, y el superadmin la de los administradores. La contraseña temporal se muestra una sola vez.

```mermaid
sequenceDiagram
  autonumber
  actor AD as Admin de empresa
  participant W as Frontend
  participant A as Backend API
  participant D as PostgreSQL
  actor AG as Agente

  AD->>W: Usuarios, elige al agente y pulsa Restablecer
  W->>A: POST /api/v1/usuarios/id/restablecer
  A->>A: RBAC usuarios:restablecer
  A->>D: SET LOCAL app.tenant_id y busca al usuario
  alt El usuario no es de su empresa
    D-->>A: Sin filas por RLS
    A-->>W: 404
  else Es de su empresa pero no es agente
    A-->>W: 403
  else Usuario válido
    A->>A: Genera contraseña temporal aleatoria
    A->>D: Guarda hash Argon2id y debe_cambiar_password
    A->>D: Revoca todos sus refresh tokens
    A->>D: Registra en auditoría
    A-->>W: 200 con contraseña temporal
    W-->>AD: La muestra una sola vez con botón copiar
  end
  AD-->>AG: Entrega la contraseña por un canal interno
  AG->>W: Inicia sesión con la temporal
  W-->>AG: Pantalla de cambio obligatorio
```
