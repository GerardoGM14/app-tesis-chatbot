# C4 · Nivel 3 — Capas del frontend

El frontend es el panel del personal con cuenta; el cliente final no lo usa. Sigue una arquitectura en capas. Cada capa solo depende de las que están debajo de ella.

```mermaid
flowchart TB
  subgraph APP["app"]
    direction LR
    A1["Router y rutas protegidas"]
    A2["Providers: QueryClient, Auth, Toaster"]
    A3["Layout: Sidebar, Breadcrumb, Topbar"]
  end

  subgraph PRES["presentation"]
    direction LR
    P1["pages/"]
    P2["components/ui"]
    P3["components/skeletons"]
  end

  subgraph APPL["application"]
    direction LR
    L1["hooks de casos de uso<br/>useConversations, useLogin"]
    L2["stores Zustand<br/>sesión, preferencias UI"]
  end

  subgraph DOM["domain"]
    direction LR
    D1["entidades y tipos"]
    D2["esquemas Zod"]
    D3["permisos por rol"]
  end

  subgraph INFRA["infrastructure"]
    direction LR
    I1["cliente HTTP<br/>refresh automático"]
    I2["endpoints por módulo"]
    I3["cliente SSE"]
  end

  APP --> PRES
  PRES --> APPL
  APPL --> DOM
  APPL --> INFRA
  INFRA --> DOM
  INFRA -- "REST + SSE" --> API[("Backend API")]
```

## Estructura de carpetas

```
frontend/src/
├── app/
│   ├── router.tsx
│   ├── guards.tsx           # RequireSession, RequirePermission
│   ├── layout/              # AppShell, Sidebar, Breadcrumb, Topbar
│   └── providers.tsx
├── presentation/
│   ├── pages/               # Dashboard, Conversaciones, Bot, Canal, Usuarios, Empresas, Login
│   └── components/
│       ├── ui/              # Button, Input, Table, Dialog, Badge, Tabs
│       └── skeletons/       # TableSkeleton, KpiSkeleton, ChatSkeleton
├── application/
│   ├── hooks/
│   └── stores/
├── domain/
│   ├── entities/
│   ├── schemas/
│   └── permissions.ts
└── infrastructure/
    ├── http/
    ├── api/
    └── realtime/
```
