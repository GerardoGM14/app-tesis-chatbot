# Diagrama de estados — Conversación

Ciclo de vida de una conversación. El campo `resuelta_por` se fija al cerrar y alimenta el indicador de consultas atendidas automáticamente.

```mermaid
stateDiagram-v2
  [*] --> EN_BOT : primer mensaje del contacto

  EN_BOT --> EN_BOT : intención reconocida / responde
  EN_BOT --> DERIVADA : pide un asesor o no hay coincidencia dos veces
  EN_BOT --> ESPERANDO_ENCUESTA : el contacto confirma que se resolvió
  EN_BOT --> CERRADA : inactividad / sin_resolver

  DERIVADA --> EN_ATENCION : un agente la toma
  DERIVADA --> CERRADA : inactividad / sin_resolver

  EN_ATENCION --> ESPERANDO_ENCUESTA : el agente la marca como resuelta
  EN_ATENCION --> CERRADA : inactividad / sin_resolver

  ESPERANDO_ENCUESTA --> CERRADA : responde la encuesta o vence el plazo

  CERRADA --> [*]
```

| Estado | Quién actúa | Se mide |
|---|---|---|
| EN_BOT | Motor del bot | Tiempo de respuesta del bot |
| DERIVADA | Cola de la empresa | Tiempo de espera hasta que un agente la toma |
| EN_ATENCION | Agente | Tiempo de respuesta humana |
| ESPERANDO_ENCUESTA | Contacto | Resolución y satisfacción |
| CERRADA | — | `resuelta_por`: bot, humano o sin_resolver |
