# ADR-0001 · Backend en Go como monolito modular

- **Estado:** aceptada
- **Fecha:** 2026-11-19

## Contexto

El PTI plantea una arquitectura SaaS multitenant basada en un monolito modular. El chatbot usa whatsmeow, que está escrito en Go. El sistema corre en un solo VPS, así que importan el consumo de memoria y el tamaño de las imágenes.

## Decisión

El backend se escribe en Go como un único binario dividido en módulos (`auth`, `tenants`, `users`, `bot`, `conversations`, `metrics`). Cada módulo tiene capas `handler`, `service`, `repository` y `domain`, y solo expone su `service` a los demás.

## Consecuencias

- Un solo lenguaje para el backend y wa-service.
- Imágenes de unos 20 MB y bajo consumo de RAM en el VPS.
- Un único despliegue, más simple que microservicios.
- El informe debe actualizarse: hoy menciona Next.js como backend.
