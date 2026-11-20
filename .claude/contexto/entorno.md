# Entorno

## PC del trabajo (Windows)

- Hay Node y git. **No hay Docker, ni Java, ni Go instalados.**
- La política de seguridad corporativa **bloquea ejecutar binarios recién compilados**
  (`go test`, `go run`, incluso un hola mundo: "Acceso denegado"). No intentar saltarse el bloqueo.
- Se puede `go build`, `go vet` y `gofmt` con un Go portable, pero las pruebas de Go solo corren en CI.
- El frontend sí funciona completo (Vite, Vitest, ESLint). Sin backend, `npm run dev` da
  `ECONNREFUSED` en `/api` porque nada escucha en `localhost:8000`.
- Para probar el sistema completo: una PC con Docker (`docker compose up --build`) o el staging del VPS.

## Credenciales de prueba

- Superadmin: el de `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` del `.env` (pide cambiarla al entrar).
- Datos demo (`docker compose run --rm migrate seed-demo`): `admin@progreso.demo`,
  `agente@progreso.demo`, `admin@santarosa.demo`, `agente@santarosa.demo`, contraseña `Demo-Sertech-2026`.

## VPS

- Ubuntu 24.04, Docker, Caddy, SonarQube, runner self-hosted con la etiqueta `vps` y usuario `deploy`.
- Secretos en `/opt/plataforma/{staging,production}.env`. Guía completa en `deploy/README.md`.
- El repositorio debe ser privado (el runner del VPS tiene acceso a Docker).
