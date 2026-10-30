# C4 · Nivel 2 — Diagrama de contenedores

Cada caja es un contenedor Docker desplegado en el VPS. Solo Caddy se expone a internet; el resto vive en la red interna de Docker.

```mermaid
---
title: "C4 nivel 2 · Contenedores"
---
flowchart TB
  usuario(["<b>Usuario interno</b><br/><i>[Persona]</i><br/>Agente, administrador<br/>o superadministrador"])
  cliente(["<b>Cliente final</b><br/><i>[Persona]</i><br/>Escribe por WhatsApp"])
  whatsapp["<b>WhatsApp</b><br/><i>[Sistema externo]</i><br/>Servidores de Meta"]

  subgraph SYS["Plataforma de atención multitenant · [Sistema]"]
    proxy["<b>Reverse proxy</b><br/><i>[Contenedor: Caddy]</i><br/>TLS, enrutamiento<br/>y cabeceras de seguridad"]
    web["<b>Frontend</b><br/><i>[Contenedor: React, TailwindCSS, nginx]</i><br/>Panel: dashboard, conversaciones,<br/>bot, usuarios"]
    api["<b>Backend API</b><br/><i>[Contenedor: Go, chi, pgx]</i><br/>Monolito modular: auth, tenants,<br/>usuarios, bot, conversaciones, métricas"]
    wa["<b>wa-service</b><br/><i>[Contenedor: Go, whatsmeow]</i><br/>Una sesión de WhatsApp<br/>por empresa"]
    db[("<b>Base de datos</b><br/><i>[Contenedor: PostgreSQL 17]</i><br/>Datos con RLS por tenant<br/>y sesiones en schema aparte")]
  end

  usuario -- "Usa<br/>[HTTPS]" --> proxy
  proxy -- "Sirve estáticos<br/>[HTTP]" --> web
  proxy -- "Reenvía /api<br/>[HTTP]" --> api
  web -- "Consume<br/>[REST JSON + SSE]" --> api
  cliente -- "Envía mensajes" --> whatsapp
  whatsapp <-- "Sesión multidispositivo<br/>[WebSocket cifrado]" --> wa
  wa <-- "Eventos y comandos<br/>[HTTP interno + HMAC]" --> api
  api -- "Lee y escribe<br/>[SQL, rol con RLS]" --> db
  wa -- "Guarda sesiones<br/>[SQL]" --> db

  classDef person fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef container fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef external fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  class usuario,cliente person
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
