# Plataforma de atención multitenant · Sertech

Plataforma SaaS multitenant con chatbot de WhatsApp e integración CI/CD para la atención al cliente.
Proyecto de tesis — Escuela de Ingeniería de Sistemas e Informática, Universidad Nacional del Santa.

## Estructura

```
├── frontend/      React + TailwindCSS (arquitectura en capas)
├── backend/       Go · monolito modular (auth, empresas, usuarios…)
├── wa-service/    Go · gateway de WhatsApp (whatsmeow, sprint 3)
├── deploy/        Compose de producción, Caddy, SonarQube y scripts del VPS
├── docs/          Arquitectura, C4, UML, procesos, calidad y diseño
└── .github/       Pipelines CI/CD
```

La arquitectura está documentada en [docs/](docs/README.md).

## Requisitos

| Herramienta | Versión |
|---|---|
| Go | 1.27 |
| Node.js | 24 LTS (mínimo 22) |
| Docker + Compose | 27 o superior |
| PostgreSQL | 17 (vía Docker) |

## Desarrollo local

### Todo con Docker

```bash
cp .env.example .env              # y cambia los secretos
docker compose up --build
docker compose run --rm migrate seed-demo   # opcional: datos de demostración
```

Panel en <http://localhost:8080>. Con `seed-demo` puedes ingresar con:

| Rol | Email | Contraseña |
|---|---|---|
| Administrador | `admin@progreso.demo` | `Demo-Sertech-2026` |
| Agente | `agente@progreso.demo` | `Demo-Sertech-2026` |
| Superadmin | el de `SUPERADMIN_EMAIL` en `.env` | se pide cambiarla al ingresar |

### Frontend con recarga en caliente

Con el backend corriendo (Docker o `go run`):

```bash
cd frontend
npm ci
npm run dev          # http://localhost:5173, /api se redirige a localhost:8000
```

### Backend sin Docker

```bash
cd backend
export MIGRATION_DATABASE_URL="postgres://plataforma_owner:...@localhost:5432/plataforma?sslmode=disable"
export APP_DB_PASSWORD=... SUPERADMIN_EMAIL=... SUPERADMIN_PASSWORD=...
go run ./cmd/api migrate

export APP_ENV=development COOKIE_SECURE=false JWT_SECRET=...
export DATABASE_URL="postgres://plataforma_app:${APP_DB_PASSWORD}@localhost:5432/plataforma?sslmode=disable"
go run ./cmd/api serve
```

## Calidad

| Comando | Qué hace |
|---|---|
| `cd backend && go test ./...` | Pruebas unitarias |
| `cd backend && TEST_DATABASE_URL=postgres://... go test ./...` | Además, integración, aislamiento entre empresas y autenticación (borra el esquema `public` de esa base) |
| `cd frontend && npm run lint` | ESLint con reglas de SonarJS |
| `cd frontend && npm run test:coverage` | Vitest con cobertura |
| `golangci-lint run` | Lint de Go (configuración en `.golangci.yml`) |

La estrategia de pruebas y el quality gate están en [docs/04-calidad](docs/04-calidad/estrategia-de-pruebas.md).

## CI/CD

```
push / PR ──► CI (GitHub) ── lint · pruebas · SonarQube · SAST · Trivy ──► imágenes en GHCR
                                                                              │
main ──► CD (runner en el VPS) ── staging ──► OWASP ZAP ──► aprobación ──► producción
```

| Workflow | Runner | Etapas |
|---|---|---|
| [ci.yml](.github/workflows/ci.yml) | GitHub | gofmt + golangci-lint + pruebas con PostgreSQL · ESLint + Vitest · SonarQube · gitleaks + govulncheck + npm audit + Semgrep · build + Trivy |
| [cd.yml](.github/workflows/cd.yml) | self-hosted (VPS) | deploy staging · ZAP baseline + API scan · deploy producción con aprobación |

### Configuración en GitHub

| Tipo | Nombre | Valor |
|---|---|---|
| Secret | `SONAR_TOKEN` | Token de análisis de SonarQube |
| Variable | `SONAR_HOST_URL` | `https://sonar.tu-dominio` |
| Variable | `STAGING_URL` | `https://staging.tu-dominio` |
| Variable | `PRODUCTION_URL` | `https://app.tu-dominio` |
| Environment | `staging` | Sin restricciones |
| Environment | `production` | Con revisores obligatorios |

Preparación del VPS, runner y SonarQube: [deploy/README.md](deploy/README.md).
