# Proceso de atención actual (AS-IS)

Situación antes de la implementación (pretest, O₁). Toda consulta depende de que una persona esté disponible.

```mermaid
flowchart LR
  subgraph CLIENTE["Cliente"]
    C1(["Envía consulta por WhatsApp"])
    C2["Espera respuesta"]
    C3(["Recibe respuesta"])
  end

  subgraph PERSONAL["Personal de atención"]
    P1{"¿Está en horario<br/>laboral?"}
    P2["Mensaje queda sin leer<br/>hasta el día siguiente"]
    P3{"¿Está disponible?"}
    P4["Conversación en cola<br/>con otras pendientes"]
    P5["Lee la consulta"]
    P6["Busca la información<br/>manualmente"]
    P7["Redacta y envía respuesta"]
  end

  C1 --> C2 --> P1
  P1 -- "No" --> P2 --> P5
  P1 -- "Sí" --> P3
  P3 -- "No" --> P4 --> P5
  P3 -- "Sí" --> P5
  P5 --> P6 --> P7 --> C3
```

**Problemas observados:** dependencia total del personal, demoras fuera de horario, respuestas repetitivas a las mismas preguntas y ningún registro medible del tiempo de atención.
