# ADR-0004 · El cliente final solo usa WhatsApp

- **Estado:** aceptada
- **Fecha:** 2026-11-23

## Contexto

En el sistema participan dos grupos: las personas que consultan a una empresa y el personal que las atiende. Los primeros diagramas usaban términos como "usuario", que no dejaban claro quién ingresa a la plataforma.

## Decisión

El cliente final no ingresa a la plataforma ni tiene cuenta: escribe desde su WhatsApp normal al número de la empresa. Solo inician sesión en el panel la empresa (`admin_empresa`), sus trabajadores (`agente`) y el equipo de Sertech (`superadmin`). No hay registro público: cada cuenta la crea un administrador.

El cliente final se guarda como **contacto**, identificado por su teléfono dentro de cada empresa, sin credenciales y sin relación con la tabla `usuarios`.

## Consecuencias

- El cliente no instala nada ni aprende un sistema nuevo.
- El panel y el API público solo atienden al personal; los mensajes del cliente llegan por wa-service.
- La encuesta de satisfacción se responde dentro del chat de WhatsApp.
- El acceso queda verificado por la prueba AUT-15.
- En la documentación, "usuario" significa una persona con cuenta; para quien consulta se usa "cliente final".
