# Documentación técnica

Plataforma de atención multitenant con chatbot de WhatsApp e integración CI/CD — Sertech.

Los diagramas están escritos en [Mermaid](https://mermaid.js.org) dentro de cada `.md`, así que GitHub y VS Code los muestran sin instalar nada. Para el informe se exportan a imagen (ver [Exportar diagramas](#exportar-diagramas)).

## Estructura

```
docs/
├── 01-arquitectura/
│   ├── c4/                 # Contexto, contenedores, componentes
│   ├── infraestructura/    # Despliegue en el VPS
│   └── seguridad/          # Seguridad por capas, autenticación y roles
├── 02-uml/
│   ├── casos-de-uso/       # 4 vistas + catálogo
│   ├── estructura/         # Clases y modelo entidad-relación
│   ├── estados/            # Ciclo de vida de la conversación
│   └── secuencias/         # 6 diagramas de secuencia
├── 03-procesos/
│   ├── negocio/            # AS-IS, TO-BE, alta de empresa
│   ├── sistema/            # Inicio de sesión, motor del bot
│   └── devops/             # Pipeline CI/CD
├── 04-calidad/             # Estrategia de pruebas y quality gate
├── 05-diseno/              # Sistema de diseño y pantalla de login
├── adr/                    # Decisiones de arquitectura
└── exports/
    ├── png/                # Misma estructura que arriba, para Word
    └── svg/                # Misma estructura que arriba, vectorial
```

## 1. Arquitectura

| Documento | Tipo |
|---|---|
| [Contexto](01-arquitectura/c4/c4-01-contexto.md) | C4 nivel 1 |
| [Contenedores](01-arquitectura/c4/c4-02-contenedores.md) | C4 nivel 2 |
| [Componentes del backend](01-arquitectura/c4/c4-03-componentes-backend.md) | C4 nivel 3 |
| [Componentes de wa-service](01-arquitectura/c4/c4-04-componentes-wa-service.md) | C4 nivel 3 |
| [Capas del frontend](01-arquitectura/c4/c4-05-capas-frontend.md) | C4 nivel 3 |
| [Despliegue en el VPS](01-arquitectura/infraestructura/despliegue-vps.md) | C4 despliegue |
| [Seguridad por capas y OWASP Top 10](01-arquitectura/seguridad/seguridad-capas.md) | Seguridad |
| [Autenticación y roles](01-arquitectura/seguridad/autenticacion-y-roles.md) | RBAC, matriz de permisos |

## 2. UML

| Documento | Tipo |
|---|---|
| [Casos de uso: catálogo](02-uml/casos-de-uso/README.md) | Índice |
| [Canal de WhatsApp](02-uml/casos-de-uso/01-canal-whatsapp.md) | Casos de uso |
| [Autenticación](02-uml/casos-de-uso/02-autenticacion.md) | Casos de uso |
| [Panel de la empresa](02-uml/casos-de-uso/03-panel-empresa.md) | Casos de uso |
| [Administración de la plataforma](02-uml/casos-de-uso/04-administracion-plataforma.md) | Casos de uso |
| [Clases del dominio](02-uml/estructura/clases-dominio.md) | Clases |
| [Modelo entidad-relación](02-uml/estructura/modelo-entidad-relacion.md) | ER |
| [Estados de la conversación](02-uml/estados/estados-conversacion.md) | Estados |
| [Login y renovación de token](02-uml/secuencias/01-login-refresh.md) | Secuencia |
| [Petición protegida con RLS](02-uml/secuencias/02-peticion-protegida-rls.md) | Secuencia |
| [Mensaje entrante respondido por el bot](02-uml/secuencias/03-mensaje-entrante-bot.md) | Secuencia |
| [Derivación a un agente y encuesta](02-uml/secuencias/04-derivacion-humano.md) | Secuencia |
| [Vinculación de WhatsApp con QR](02-uml/secuencias/05-vinculacion-qr.md) | Secuencia |
| [Restablecer contraseña](02-uml/secuencias/06-restablecer-contrasena.md) | Secuencia |

## 3. Procesos

| Documento | Tipo |
|---|---|
| [Atención actual (AS-IS)](03-procesos/negocio/proceso-atencion-as-is.md) | Negocio |
| [Atención propuesta (TO-BE)](03-procesos/negocio/proceso-atencion-to-be.md) | Negocio |
| [Alta de una empresa](03-procesos/negocio/alta-de-empresa.md) | Negocio |
| [Flujo de inicio de sesión](03-procesos/sistema/flujo-inicio-sesion.md) | Actividad |
| [Algoritmo del motor del bot](03-procesos/sistema/motor-del-bot.md) | Actividad |
| [Pipeline CI/CD](03-procesos/devops/pipeline-cicd.md) | DevOps |

## 4. Calidad

- [Estrategia de pruebas, suites de aislamiento y autenticación, quality gate](04-calidad/estrategia-de-pruebas.md)

## 5. Diseño

- [Sistema de diseño: paleta, tipografía (Sora + Poppins), layout y login](05-diseno/sistema-de-diseno.md)

## Decisiones de arquitectura (ADR)

- [ADR-0001 · Backend en Go como monolito modular](adr/0001-backend-go-monolito-modular.md)
- [ADR-0002 · Multitenancy con modelo Pool y RLS](adr/0002-multitenancy-rls.md)
- [ADR-0003 · Servicio de WhatsApp separado](adr/0003-wa-service-separado.md)

## Exportar diagramas

Requiere Node 20 o superior. Si Puppeteer no logra descargar su Chromium, el script usa Edge o Chrome instalados (o la ruta en `PUPPETEER_EXECUTABLE_PATH`).

```bash
cd docs
npm install
npm run render       # SVG en exports/svg/
npm run render:png   # PNG a 3x en exports/png/, para pegar en Word
```

Cada exportación borra y regenera su carpeta, así que nunca quedan imágenes de diagramas que ya no existen.
