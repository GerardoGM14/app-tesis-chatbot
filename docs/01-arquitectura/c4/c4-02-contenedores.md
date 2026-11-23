# C4 · Nivel 2 — Diagrama de contenedores

Cada caja es un contenedor Docker desplegado en el VPS. Solo Caddy se expone a internet; el resto vive en la red interna de Docker. Hay dos caminos de entrada: el personal con cuenta entra al panel por HTTPS, y los mensajes del cliente final llegan por la sesión de WhatsApp que mantiene wa-service.

```mermaid
---
title: "C4 nivel 2 · Contenedores"
---
flowchart TB
  personal(["<b>Personal con cuenta</b><br/><i>[Persona]</i><br/>Administrador de empresa,<br/>agente o superadministrador"])
  cliente(["<b>Cliente final</b><br/><i>[Persona sin cuenta]</i><br/>Escribe desde su<br/>WhatsApp normal"])
  whatsapp["<b>WhatsApp</b><br/><i>[Sistema externo]</i><br/>Servidores de Meta"]

  subgraph SYS["Plataforma de atención multitenant · [Sistema]"]
    proxy["<b>Reverse proxy</b><br/><i>[Contenedor: Caddy]</i><br/>TLS, enrutamiento<br/>y cabeceras de seguridad"]
    web["<b>Frontend</b><br/><i>[Contenedor: React, TailwindCSS, nginx]</i><br/>Panel: dashboard, conversaciones,<br/>bot, usuarios"]
    api["<b>Backend API</b><br/><i>[Contenedor: Go, chi, pgx]</i><br/>Monolito modular: auth, tenants,<br/>usuarios, bot, conversaciones, métricas"]
    wa["<b>wa-service</b><br/><i>[Contenedor: Go, whatsmeow]</i><br/>Una sesión de WhatsApp<br/>por empresa"]
    db[("<b>Base de datos</b><br/><i>[Contenedor: PostgreSQL 17]</i><br/>Datos con RLS por tenant<br/>y sesiones en schema aparte")]
  end

  personal -- "Inicia sesión y usa el panel<br/>[HTTPS]" --> proxy
  proxy -- "Sirve estáticos<br/>[HTTP]" --> web
  proxy -- "Reenvía /api<br/>[HTTP]" --> api
  web -- "Consume<br/>[REST JSON + SSE]" --> api
  cliente -- "Escribe y recibe<br/>respuestas" --> whatsapp
  whatsapp <-- "Sesión multidispositivo<br/>[WebSocket cifrado]" --> wa
  wa <-- "Eventos y comandos<br/>[HTTP interno + HMAC]" --> api
  api -- "Lee y escribe<br/>[SQL, rol con RLS]" --> db
  wa -- "Guarda sesiones<br/>[SQL]" --> db

  classDef person fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef client fill:#FFFFFF,stroke:#142A97,color:#142A97
  classDef container fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef external fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  class personal person
  class cliente client
  class proxy,web,api,wa,db container
  class whatsapp external
  style SYS fill:#F7F8FA,stroke:#1834BC,stroke-dasharray:6 4,color:#0F1219
```

## Decisiones clave

| Decisión | Motivo | ADR |
|---|---|---|
| Backend en Go como monolito modular | Una sola unidad de despliegue con módulos de bajo acoplamiento | [ADR-0001](../../adr/0001-backend-go-monolito-modular.md) |
| Aislamiento con RLS en PostgreSQL | El aislamiento no depende de que cada consulta lleve `WHERE tenant_id` | [ADR-0002](../../adr/0002-multitenancy-rls.md) |
| wa-service separado | Redesplegar el backend no corta las sesiones de WhatsApp | [ADR-0003](../../adr/0003-wa-service-separado.md) |
| El cliente final no tiene cuenta | Usa su WhatsApp de siempre; el panel es solo para el personal | [ADR-0004](../../adr/0004-cliente-final-solo-whatsapp.md) |
