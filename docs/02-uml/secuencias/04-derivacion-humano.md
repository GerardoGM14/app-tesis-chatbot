# Secuencia — Derivación a un agente y encuesta

El cliente final conversa todo el tiempo en su WhatsApp y no ingresa a la plataforma. El agente, que es un trabajador de la empresa, responde desde el panel.

```mermaid
sequenceDiagram
  autonumber
  actor C as Cliente final
  participant WA as WhatsApp
  participant G as wa-service
  participant A as Backend API
  participant D as PostgreSQL
  participant P as Panel del personal
  actor AG as Agente (trabajador)

  C->>WA: Quiero hablar con un asesor
  WA->>G: Evento Message
  G->>A: Evento mensaje entrante
  A->>D: Estado de la conversación a DERIVADA
  A->>G: Envía mensaje de derivación
  G->>WA: Envía mensaje
  WA-->>C: En breve un asesor te atenderá
  A-->>P: SSE nueva conversación en cola
  P-->>AG: Aparece en la bandeja con insignia
  AG->>P: Toma la conversación
  P->>A: POST /api/v1/conversaciones/id/asignar
  A->>D: Estado EN_ATENCION y agente_id
  loop Diálogo
    AG->>P: Escribe respuesta
    P->>A: POST /api/v1/conversaciones/id/mensajes
    A->>G: Enviar
    G->>WA: Envía mensaje
    WA-->>C: Mensaje del agente
    C->>WA: Responde
    WA->>G: Evento Message
    G->>A: Evento mensaje entrante
    A-->>P: SSE nuevo mensaje
  end
  AG->>P: Marcar como resuelta
  P->>A: POST /api/v1/conversaciones/id/resolver
  A->>D: Estado ESPERANDO_ENCUESTA
  A->>G: Enviar encuesta
  G->>WA: Envía encuesta
  WA-->>C: ¿Se resolvió tu consulta? 1 Sí 2 No
  C->>WA: 1
  WA->>G: Evento Message
  G->>A: Evento con la respuesta
  A->>G: Pide la calificación
  G->>WA: Envía mensaje
  WA-->>C: Califica la atención del 1 al 5
  C->>WA: 5
  WA->>G: Evento Message
  G->>A: Evento con la respuesta
  A->>D: Guarda encuesta y cierra con resuelta_por HUMANO
```
