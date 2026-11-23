# C4 · Nivel 1 — Diagrama de contexto

Muestra la plataforma como una sola caja, las personas que se relacionan con ella y los sistemas externos con los que se comunica. El **cliente final no ingresa a la plataforma**: escribe desde su WhatsApp normal. Solo el personal de cada empresa y el equipo de Sertech inician sesión en el panel ([ADR-0004](../../adr/0004-cliente-final-solo-whatsapp.md)).

```mermaid
---
title: "C4 nivel 1 · Contexto del sistema"
---
flowchart TB
  subgraph SOLOWA["Solo usa su WhatsApp · sin cuenta"]
    cliente(["<b>Cliente final</b><br/><i>[Persona]</i><br/>Consulta a una empresa<br/>desde su WhatsApp normal"])
  end

  subgraph PANEL["Personal con cuenta · inicia sesión en el panel"]
    admin(["<b>Administrador de empresa</b><br/><i>[Persona]</i><br/>Configura el bot, las cuentas<br/>y el canal de su empresa"])
    agente(["<b>Agente</b><br/><i>[Persona: trabajador de la empresa]</i><br/>Atiende las conversaciones<br/>derivadas"])
    superadmin(["<b>Superadministrador</b><br/><i>[Persona: equipo Sertech]</i><br/>Alta y supervisión<br/>de empresas"])
  end

  whatsapp["<b>WhatsApp</b><br/><i>[Sistema externo]</i><br/>Red de mensajería de Meta"]

  plataforma["<b>Plataforma de atención multitenant</b><br/><i>[Sistema de software]</i><br/>SaaS que atiende a varias empresas con un<br/>chatbot de WhatsApp y aísla los datos de cada una"]

  github["<b>GitHub</b><br/><i>[Sistema externo]</i><br/>Repositorio, Actions y GHCR"]
  sonar["<b>SonarQube</b><br/><i>[Sistema externo]</i><br/>Análisis estático y quality gate"]

  cliente -- "Escribe y recibe<br/>respuestas" --> whatsapp
  whatsapp <-- "Entrega y recibe mensajes<br/>[multidispositivo]" --> plataforma
  admin -- "Configura su empresa<br/>[HTTPS]" --> plataforma
  agente -- "Atiende conversaciones<br/>[HTTPS]" --> plataforma
  superadmin -- "Administra empresas<br/>[HTTPS]" --> plataforma
  github -- "Prueba, construye y despliega<br/>[runner self-hosted]" --> plataforma
  github -- "Envía análisis" --> sonar

  classDef person fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef client fill:#FFFFFF,stroke:#142A97,color:#142A97
  classDef system fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef external fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  class agente,admin,superadmin person
  class cliente client
  class plataforma system
  class whatsapp,github,sonar external
  style SOLOWA fill:#F7F8FA,stroke:#9AA0AC,stroke-dasharray:6 4,color:#0F1219
  style PANEL fill:#EEF1FC,stroke:#1834BC,color:#0F1219
```

**Leyenda:** azul oscuro = personal con cuenta · contorno azul = cliente final, sin cuenta · azul = sistema que construimos · gris = sistema externo.
