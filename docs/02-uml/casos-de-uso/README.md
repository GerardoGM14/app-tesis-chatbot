# Casos de uso

Los casos de uso se dividen en cuatro vistas para que cada diagrama se lea en una sola hoja.

| Vista | Actores | Casos de uso |
|---|---|---|
| [1. Canal de WhatsApp](01-canal-whatsapp.md) | Cliente final, por WhatsApp | CU01 a CU04 |
| [2. Autenticación](02-autenticacion.md) | Agente, Admin de empresa, Superadmin | CU05 a CU08 |
| [3. Panel de la empresa](03-panel-empresa.md) | Agente, Admin de empresa | CU09 a CU15 |
| [4. Administración de la plataforma](04-administracion-plataforma.md) | Superadmin, GitHub Actions | CU16 a CU18 |

## Actores

| Actor | ¿Ingresa a la plataforma? | Descripción |
|---|:---:|---|
| Cliente final | No | Persona que consulta a la empresa desde su WhatsApp normal |
| WhatsApp | — | Sistema externo que transporta los mensajes |
| Agente | Sí | Trabajador de la empresa que atiende conversaciones |
| Admin de empresa | Sí | Responsable de la empresa: configura y también puede atender |
| Superadmin | Sí | Equipo de Sertech que administra las empresas |
| GitHub Actions | — | Sistema externo que despliega nuevas versiones |

**Personal con cuenta** agrupa a Agente, Admin de empresa y Superadmin.

## Catálogo

| ID | Caso de uso | Actor principal |
|---|---|---|
| CU01 | Consultar por WhatsApp | Cliente final, por WhatsApp |
| CU02 | Recibir respuesta automática | Cliente final, por WhatsApp |
| CU03 | Solicitar atención humana | Cliente final, por WhatsApp |
| CU04 | Responder encuesta | Cliente final, por WhatsApp |
| CU05 | Iniciar sesión | Personal con cuenta |
| CU06 | Cerrar sesión | Personal con cuenta |
| CU07 | Cambiar contraseña | Personal con cuenta |
| CU08 | Restablecer contraseña de un usuario | Admin de empresa, Superadmin |
| CU09 | Atender conversación derivada | Agente, Admin de empresa |
| CU10 | Cerrar conversación | Agente, Admin de empresa |
| CU11 | Ver dashboard de indicadores | Agente, Admin de empresa |
| CU12 | Gestionar intenciones y FAQs | Admin de empresa |
| CU13 | Configurar horario y mensajes | Admin de empresa |
| CU14 | Vincular WhatsApp con QR | Admin de empresa |
| CU15 | Gestionar usuarios de la empresa | Admin de empresa |
| CU16 | Gestionar empresas | Superadmin |
| CU17 | Ver métricas globales | Superadmin |
| CU18 | Desplegar nueva versión | GitHub Actions |

Los permisos detallados de cada rol están en [Autenticación y roles](../../01-arquitectura/seguridad/autenticacion-y-roles.md).
