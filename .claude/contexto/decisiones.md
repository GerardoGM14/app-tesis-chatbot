# Decisiones tomadas

Estas decisiones ya están acordadas con el usuario. Se registran como ADR en `docs/adr/`.

## Stack

- **Frontend:** React 19, TailwindCSS 4, Vite 8, TypeScript, React Router 8, TanStack Query, Zustand,
  react-hook-form + Zod, lucide-react (iconos), sonner (toasts). Arquitectura **en capas**:
  `app` → `presentation` → `application` → `domain` / `infrastructure`.
- **Backend:** Go 1.27, **no Next.js** (el informe de tesis dice Next.js y hay que corregirlo).
  Monolito modular: cada módulo en `internal/modules/<modulo>/{domain,repository,service,handler}`.
  chi como router, pgx con SQL propio (sin ORM, sin sqlc), migraciones SQL embebidas con un runner propio.
- **wa-service:** Go con whatsmeow, servicio separado para que redesplegar el backend no corte las
  sesiones de WhatsApp. Se comunica con el backend por HTTP interno firmado con HMAC.
- **Base de datos:** PostgreSQL 17.
- **Infraestructura:** Docker en un VPS Linux propio, Caddy como proxy con TLS, GitHub Actions con
  runner self-hosted en el VPS para el despliegue. **SonarQube corre en el mismo VPS.**

## Multitenancy

- Modelo **Pool**: todas las tablas de negocio llevan `tenant_id` y tienen Row-Level Security.
- El backend se conecta con el rol `plataforma_app` (sin BYPASSRLS, no es dueño de las tablas).
- Cada transacción fija el alcance con `set_config('app.tenant_id', ..., true)` y `app.is_platform`.
- El tenant sale siempre del JWT, nunca del cuerpo ni de la URL.

## Quién ingresa a la plataforma (ADR-0004)

- El **cliente final no ingresa ni tiene cuenta**: usa su WhatsApp normal. Se guarda como contacto.
- Ingresan al panel: la empresa (`admin_empresa`), sus trabajadores (`agente`) y Sertech (`superadmin`).
- Sin registro público. En diagramas y textos: "cliente final" vs. "personal con cuenta"; no decir
  "usuario" para referirse al cliente.

## Autenticación y roles

- Roles: `superadmin` (plataforma, sin tenant), `admin_empresa`, `agente`. Matriz de permisos en
  `backend/internal/platform/authz/authz.go` y `docs/01-arquitectura/seguridad/autenticacion-y-roles.md`.
- El superadmin **no** puede leer conversaciones de las empresas.
- Access token JWT de 15 min en memoria del frontend; refresh token en cookie `httpOnly`
  `SameSite=Strict` con rotación y detección de reutilización (revoca la familia).
- Argon2id, bloqueo tras 5 intentos durante 15 min, mensajes de error siempre genéricos.
- Contraseña temporal + cambio obligatorio en el primer ingreso. Sin recuperación por correo:
  el admin restablece a sus agentes y el superadmin a los admins.
- Pantalla inicial: superadmin → `/empresas`, admin → `/dashboard`, agente → `/conversaciones`.

## Indicadores de la tesis que el sistema debe producir

Tiempo promedio de respuesta (≤ 5 s), % atendidas por el bot (≥ 80 %), disponibilidad (≥ 99 %),
% resueltas y satisfacción (encuesta por WhatsApp), % de aislamiento entre tenants (suite AIS),
% de despliegues exitosos (GitHub Actions).
