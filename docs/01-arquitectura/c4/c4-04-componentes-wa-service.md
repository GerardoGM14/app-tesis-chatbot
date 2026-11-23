# C4 · Nivel 3 — Componentes de wa-service

wa-service es un gateway delgado: no contiene reglas de negocio del bot, solo mantiene las sesiones y traduce eventos de WhatsApp a llamadas HTTP internas. Es el único punto por el que llegan los mensajes del cliente final.

```mermaid
---
title: "C4 nivel 3 · Componentes de wa-service"
---
flowchart LR
  api["<b>Backend API</b><br/><i>[Contenedor: Go]</i>"]
  whatsapp["<b>WhatsApp</b><br/><i>[Sistema externo]</i>"]

  subgraph WA["wa-service · [Contenedor: Go, whatsmeow]"]
    http["<b>API interna</b><br/><i>[Componente: net/http]</i><br/>connect, disconnect,<br/>status, send · HMAC"]
    manager["<b>Gestor de sesiones</b><br/><i>[Componente]</i><br/>Una conexión por tenant,<br/>reconexión automática"]
    handler["<b>Manejador de eventos</b><br/><i>[Componente]</i><br/>Message, Connected,<br/>LoggedOut, QR"]
    outbox["<b>Cola de reintentos</b><br/><i>[Componente: tabla outbox]</i><br/>Backoff si el backend<br/>no responde"]
    store["<b>Sqlstore</b><br/><i>[Componente: whatsmeow]</i><br/>Llaves y dispositivos"]
  end

  db[("<b>PostgreSQL</b><br/><i>[Contenedor]</i><br/>schema whatsmeow")]

  api -- "Comandos<br/>[HTTP + HMAC]" --> http
  http --> manager
  manager <-- "Sesión<br/>[WebSocket cifrado]" --> whatsapp
  whatsapp -- "Eventos" --> handler
  handler -- "Encola" --> outbox
  outbox -- "Entrega evento<br/>[HTTP + HMAC]" --> api
  manager --> store
  store --> db
  outbox --> db

  classDef container fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  classDef component fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  class api,whatsapp,db container
  class http,manager,handler,outbox,store component
  style WA fill:#F7F8FA,stroke:#1834BC,stroke-dasharray:6 4,color:#0F1219
```
