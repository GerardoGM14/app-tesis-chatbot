# Casos de uso · Canal de WhatsApp

El cliente final **no ingresa a la plataforma ni tiene cuenta**: chatea desde su WhatsApp normal con el número de la empresa. WhatsApp es el actor externo que conecta sus mensajes con el sistema.

```mermaid
---
title: "Casos de uso · Canal de WhatsApp"
---
flowchart LR
  CL(("Cliente<br/>final"))
  WA(("WhatsApp"))

  subgraph SYS["Plataforma de atención multitenant"]
    UC01(["CU01 Consultar por WhatsApp"])
    UC02(["CU02 Recibir respuesta automática"])
    UC03(["CU03 Solicitar atención humana"])
    UC04(["CU04 Responder encuesta"])
  end

  CL -- "chatea desde<br/>su teléfono" --- WA
  WA --- UC01
  WA --- UC03
  WA --- UC04
  UC01 -. "«include»" .-> UC02

  classDef actor fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef system fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  classDef usecase fill:#EEF1FC,stroke:#1834BC,color:#0F1219
  class CL actor
  class WA system
  class UC01,UC02,UC03,UC04 usecase
  style SYS fill:#FFFFFF,stroke:#1834BC,color:#0F1219
```
