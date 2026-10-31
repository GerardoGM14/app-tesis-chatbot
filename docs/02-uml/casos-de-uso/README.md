# Casos de uso

Los casos de uso se dividen en cuatro vistas para que cada diagrama se lea en una sola hoja.

| Vista | Actores | Casos de uso |
|---|---|---|
| [1. Canal de WhatsApp](01-canal-whatsapp.md) | Cliente final | CU01 a CU04 |
| [2. Autenticación](02-autenticacion.md) | Agente, Admin de empresa, Superadmin | CU05 a CU08 |
| [3. Panel de la empresa](03-panel-empresa.md) | Agente, Admin de empresa | CU09 a CU15 |
| [4. Administración de la plataforma](04-administracion-plataforma.md) | Superadmin, GitHub Actions | CU16 a CU18 |

## Catálogo

| ID | Caso de uso | Actor principal |
|---|---|---|
| CU01 | Consultar por WhatsApp | Cliente final |
| CU02 | Recibir respuesta automática | Cliente final |
| CU03 | Solicitar atención humana | Cliente final |
| CU04 | Responder encuesta | Cliente final |
| CU05 | Iniciar sesión | Usuario interno |
| CU06 | Cerrar sesión | Usuario interno |
| CU07 | Cambiar contraseña | Usuario interno |
| CU08 | Restablecer contraseña de un usuario | Admin de empresa, Superadmin |
| CU09 | Atender conversación derivada | Agente |
| CU10 | Cerrar conversación | Agente |
| CU11 | Ver dashboard de indicadores | Agente, Admin de empresa |
| CU12 | Gestionar intenciones y FAQs | Admin de empresa |
| CU13 | Configurar horario y mensajes | Admin de empresa |
| CU14 | Vincular WhatsApp con QR | Admin de empresa |
| CU15 | Gestionar usuarios de la empresa | Admin de empresa |
| CU16 | Gestionar empresas | Superadmin |
| CU17 | Ver métricas globales | Superadmin |
| CU18 | Desplegar nueva versión | GitHub Actions |

Los permisos detallados de cada rol están en [Autenticación y roles](../../01-arquitectura/seguridad/autenticacion-y-roles.md).
