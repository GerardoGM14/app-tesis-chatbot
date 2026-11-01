# ADR-0003 · Servicio de WhatsApp separado del backend

- **Estado:** aceptada
- **Fecha:** 2026-11-19

## Contexto

Las sesiones de whatsmeow mantienen una conexión permanente con WhatsApp. Si viven dentro del backend, cada despliegue las corta y los mensajes que lleguen en ese momento se pierden.

## Decisión

Las sesiones viven en `wa-service`, un contenedor aparte que solo hace de gateway: recibe eventos, los reenvía al backend y envía los mensajes que el backend le ordena. La lógica del bot queda en el backend. La comunicación es HTTP interno firmado con HMAC y una cola de reintentos guarda los eventos si el backend no responde.

## Consecuencias

- El backend se puede desplegar muchas veces al día sin desconectar WhatsApp, lo que favorece el indicador de disponibilidad.
- wa-service cambia poco, así que se redespliega rara vez.
- Hay un contrato interno que mantener y probar entre ambos servicios.
