# ADR-0002 · Multitenancy con modelo Pool y Row-Level Security

- **Estado:** aceptada
- **Fecha:** 2026-11-19

## Contexto

Varias empresas comparten la plataforma y sus datos deben permanecer aislados. Los modelos posibles son Silo (una BD por empresa), Bridge (un schema por empresa) y Pool (tablas compartidas con `tenant_id`).

## Decisión

Se usa el modelo **Pool**: todas las tablas de negocio llevan `tenant_id` y tienen RLS activada en PostgreSQL. El backend se conecta con un rol sin `BYPASSRLS` y, en cada transacción, ejecuta `SET LOCAL app.tenant_id` con el valor tomado del JWT.

## Consecuencias

- Agregar una empresa no requiere crear bases ni schemas.
- Si una consulta olvida filtrar por tenant, la BD lo hace de todos modos: defensa en profundidad.
- Las migraciones corren con un rol distinto (propietario de las tablas).
- El aislamiento se puede medir con una suite de pruebas automatizada (objetivo específico 1).
