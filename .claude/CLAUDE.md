# Contexto del proyecto para Claude Code

Plataforma SaaS multitenant con chatbot de WhatsApp e integración CI/CD para la atención al cliente
de empresas de Nuevo Chimbote (cliente piloto: Sertech). Es un proyecto de tesis de Ingeniería de
Sistemas (Universidad Nacional del Santa). Toda la comunicación con el usuario es en **español**.

Detalle por tema:

- @.claude/contexto/decisiones.md — stack, arquitectura y decisiones ya tomadas (no volver a discutirlas)
- @.claude/contexto/convenciones.md — reglas de código, diseño y calidad que el código debe cumplir
- @.claude/contexto/estado.md — qué está hecho, qué sigue y pendientes del informe de tesis
- @.claude/contexto/entorno.md — particularidades de las máquinas y del VPS

## Resumen rápido

| Carpeta | Qué es |
|---|---|
| `frontend/` | React 19 + TailwindCSS 4 + Vite, arquitectura en capas (no atómica) |
| `backend/` | Go 1.27, monolito modular con RLS de PostgreSQL por tenant |
| `wa-service/` | Go, gateway de WhatsApp con whatsmeow (librería Go, se importa con `go get`) |
| `deploy/` | Compose del VPS, Caddy, SonarQube y scripts de despliegue |
| `docs/` | Arquitectura C4, UML, procesos, calidad, diseño y ADRs (diagramas Mermaid) |
| `.github/workflows/` | CI (GitHub) y CD (runner self-hosted en el VPS) |

## Cómo trabajar con el usuario

- No asumir cosas que no dijo (por ejemplo, no llamar "SPA" al frontend sin que lo pida).
- Confirmar decisiones de stack antes de implementarlas; recomendar una opción en lugar de listar todas.
- Commits con Conventional Commits en español y **sin** línea de coautoría.
- Al terminar trabajo de código, verificar (tipos, lint, pruebas) y decir con honestidad qué no se pudo ejecutar.
