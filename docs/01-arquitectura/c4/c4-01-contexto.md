# C4 · Nivel 1 — Diagrama de contexto

Muestra la plataforma como una sola caja, las personas que la usan y los sistemas externos con los que se comunica.

```mermaid
---
title: "C4 nivel 1 · Contexto del sistema"
---
flowchart TB
  cliente(["<b>Cliente final</b><br/><i>[Persona]</i><br/>Consulta a una empresa<br/>por WhatsApp"])
  agente(["<b>Agente de atención</b><br/><i>[Persona]</i><br/>Atiende las conversaciones<br/>derivadas"])
  admin(["<b>Administrador de empresa</b><br/><i>[Persona]</i><br/>Configura el bot, usuarios<br/>y canal de su empresa"])
  superadmin(["<b>Superadministrador</b><br/><i>[Persona]</i><br/>Equipo Sertech: alta y<br/>supervisión de empresas"])

  plataforma["<b>Plataforma de atención multitenant</b><br/><i>[Sistema de software]</i><br/>SaaS que atiende a varias empresas con un<br/>chatbot de WhatsApp y aísla los datos de cada una"]

  whatsapp["<b>WhatsApp</b><br/><i>[Sistema externo]</i><br/>Red de mensajería de Meta"]
  github["<b>GitHub</b><br/><i>[Sistema externo]</i><br/>Repositorio, Actions y GHCR"]
  sonar["<b>SonarQube</b><br/><i>[Sistema externo]</i><br/>Análisis estático y quality gate"]

  cliente -- "Envía consultas" --> whatsapp
  whatsapp <-- "Entrega y recibe mensajes<br/>[multidispositivo]" --> plataforma
  agente -- "Atiende conversaciones<br/>[HTTPS]" --> plataforma
  admin -- "Configura su empresa<br/>[HTTPS]" --> plataforma
  superadmin -- "Administra empresas<br/>[HTTPS]" --> plataforma
  github -- "Prueba, construye y despliega<br/>[runner self-hosted]" --> plataforma
  github -- "Envía análisis" --> sonar

  classDef person fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef system fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef external fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  class cliente,agente,admin,superadmin person
  class plataforma system
  class whatsapp,github,sonar external
```

**Leyenda:** azul oscuro = persona · azul = sistema que construimos · gris = sistema externo.
