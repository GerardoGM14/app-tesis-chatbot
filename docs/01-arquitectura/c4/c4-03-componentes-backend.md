# C4 · Nivel 3 — Componentes del backend

El backend es un monolito modular. Cada módulo sigue las capas `handler → service → repository → domain` y solo se comunica con otros módulos a través de su `service`.

```mermaid
---
title: "C4 nivel 3 · Componentes del Backend API"
---
flowchart TB
  web["<b>Frontend</b><br/><i>[Contenedor: React]</i>"]
  wa["<b>wa-service</b><br/><i>[Contenedor: Go, whatsmeow]</i>"]

  subgraph API["Backend API · [Contenedor: Go]"]
    router["<b>Router y middlewares</b><br/><i>[Componente: chi]</i><br/>Rate limit, JWT, tenant,<br/>RBAC, cabeceras, logging"]
    internal["<b>Endpoints internos</b><br/><i>[Componente]</i><br/>Eventos de wa-service,<br/>verificación HMAC"]

    subgraph MODS["Módulos"]
      auth["<b>auth</b><br/>Login, refresh,<br/>Argon2id"]
      tenants["<b>tenants</b><br/>Alta y suspensión<br/>de empresas"]
      users["<b>users</b><br/>Usuarios y roles"]
      bot["<b>bot</b><br/>Intenciones, FAQs,<br/>horarios, motor"]
      conv["<b>conversations</b><br/>Mensajes, derivación,<br/>encuestas"]
      metrics["<b>metrics</b><br/>Indicadores<br/>de la tesis"]
    end

    sse["<b>Hub SSE</b><br/><i>[Componente]</i><br/>Eventos en vivo por tenant"]
    dbtx["<b>Capa de datos</b><br/><i>[Componente: pgx]</i><br/>Transacción con<br/>SET LOCAL app.tenant_id"]
  end

  db[("<b>PostgreSQL</b><br/><i>[Contenedor]</i><br/>RLS por tenant")]

  web -- "REST JSON<br/>[HTTPS]" --> router
  router --> MODS
  wa -- "Eventos<br/>[HTTP + HMAC]" --> internal
  internal --> conv
  conv -- "Pide respuesta" --> bot
  conv -- "Envía respuesta<br/>[HTTP + HMAC]" --> wa
  conv -- "Publica cambios" --> sse
  sse -- "Eventos en vivo<br/>[SSE]" --> web
  MODS --> dbtx
  dbtx -- "SQL" --> db

  classDef container fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  classDef component fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef module fill:#EEF1FC,stroke:#1834BC,color:#0F1219
  class web,wa,db container
  class router,internal,sse,dbtx component
  class auth,tenants,users,bot,conv,metrics module
  style API fill:#F7F8FA,stroke:#1834BC,stroke-dasharray:6 4,color:#0F1219
  style MODS fill:#FFFFFF,stroke:#C9CDD6,color:#0F1219
```

## Estructura de carpetas

```
backend/
├── cmd/api/main.go
├── internal/
│   ├── platform/              # transversal
│   │   ├── config/            # variables de entorno validadas
│   │   ├── database/          # pool pgx, helper WithTenantTx
│   │   ├── httpx/             # servidor, errores, respuestas JSON
│   │   ├── middleware/        # ratelimit, auth, tenant, rbac, secure headers
│   │   └── realtime/          # hub SSE
│   └── modules/
│       └── <modulo>/
│           ├── domain/        # entidades y reglas, sin dependencias externas
│           ├── repository/    # SQL parametrizado con pgx
│           ├── service/       # casos de uso
│           └── handler/       # HTTP: decodifica, valida, llama al service
├── migrations/
└── api/openapi.yaml
```

Regla de dependencias: `handler → service → domain ← repository`. Un módulo no importa el `repository` de otro.
