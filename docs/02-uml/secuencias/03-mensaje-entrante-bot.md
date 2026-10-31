# Secuencia — Mensaje entrante respondido por el bot

Es el flujo principal del sistema. Los instantes `recibido_en` y `enviado_en` alimentan el indicador de tiempo promedio de respuesta (meta: 5 segundos o menos).

```mermaid
sequenceDiagram
  autonumber
  actor C as Cliente
  participant WA as WhatsApp
  participant G as wa-service
  participant A as Backend API
  participant B as Módulo bot
  participant D as PostgreSQL
  participant P as Panel web

  C->>WA: Escribe consulta
  WA->>G: Evento Message
  G->>G: Identifica el tenant por la sesión
  G->>A: POST /internal/wa/events firmado con HMAC
  A->>A: Verifica firma y timestamp
  A->>D: Upsert contacto y conversación
  A->>D: Guarda mensaje entrante con recibido_en
  A->>B: Responder texto en contexto del tenant
  B->>D: Lee configuración, horario e intenciones
  B->>B: Normaliza texto y calcula puntaje por intención
  alt Coincidencia sobre el umbral
    B-->>A: Respuesta de la intención
  else Sin coincidencia
    B-->>A: Menú de opciones o derivación
  end
  A->>G: POST /internal/sessions/tenant/send
  G->>WA: Envía mensaje
  WA-->>C: Respuesta
  G-->>A: 200 con wa_message_id
  A->>D: Guarda mensaje saliente con enviado_en
  A-->>P: Evento SSE conversación actualizada
```
