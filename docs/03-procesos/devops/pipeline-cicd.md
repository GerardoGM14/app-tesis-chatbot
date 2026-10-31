# Pipeline CI/CD

Cada flecha es una dependencia entre jobs de GitHub Actions. Si una etapa falla, las siguientes no se ejecutan y el cambio no llega a producción. Los resultados de cada ejecución alimentan el indicador de despliegues automatizados exitosos.

```mermaid
---
title: "Pipeline CI/CD"
---
flowchart TB
  T(["push o pull request"])

  subgraph CI["1 · Integración continua · runners de GitHub"]
    direction LR
    Q["<b>Calidad</b><br/>lint + unitarias<br/>3 proyectos"]
    IT["<b>Integración</b><br/>API + PostgreSQL<br/>aislamiento tenant"]
    SA["<b>SAST</b><br/>gosec, Semgrep,<br/>npm audit, gitleaks"]
    SQ["<b>SonarQube</b><br/>cobertura +<br/>quality gate"]
    B["<b>Build</b><br/>imágenes +<br/>Trivy"]
    Q --> IT --> SQ --> B
    Q --> SA --> B
  end

  G{"¿Rama main?"}
  PR(["Resultado en el PR"])

  subgraph STG["2 · Staging: QA y OWASP · runner self-hosted"]
    direction LR
    PUSH["<b>Publica</b><br/>en GHCR"]
    DS["<b>Deploy staging</b><br/>compose up +<br/>healthcheck"]
    E2E["<b>E2E</b><br/>Playwright"]
    ZAP["<b>OWASP ZAP</b><br/>baseline +<br/>API scan"]
    K6["<b>k6</b><br/>p95 ≤ 5 s"]
    PUSH --> DS --> E2E --> ZAP --> K6
  end

  subgraph PRD["3 · Producción · runner self-hosted"]
    direction LR
    AP{"Aprobación<br/>manual"}
    DP["<b>Deploy<br/>producción</b>"]
    HC{"¿Healthcheck<br/>OK?"}
    OK(["Despliegue<br/>exitoso"])
    RB["<b>Rollback</b><br/>imagen anterior"]
    AP --> DP --> HC
    HC -- "Sí" --> OK
    HC -- "No" --> RB
  end

  T --> CI
  CI --> G
  G -- "No" --> PR
  G -- "Sí" --> STG
  STG --> PRD

  classDef step fill:#EEF1FC,stroke:#1834BC,color:#0F1219
  classDef decision fill:#FFF3EC,stroke:#EB6B1D,color:#0F1219
  classDef terminal fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  class Q,IT,SA,SQ,B,PUSH,DS,E2E,ZAP,K6,DP,RB step
  class G,AP,HC decision
  class T,PR,OK terminal
  style CI fill:#FFFFFF,stroke:#9AA0AC,color:#0F1219
  style STG fill:#FFFFFF,stroke:#9AA0AC,color:#0F1219
  style PRD fill:#FFFFFF,stroke:#9AA0AC,color:#0F1219
```

## Relación con ISO/IEC/IEEE 29119 y el enfoque Shift-Left

| Etapa del proceso de pruebas | Dónde ocurre |
|---|---|
| Estrategia y planificación | [estrategia-de-pruebas.md](../../04-calidad/estrategia-de-pruebas.md) |
| Diseño e implementación | Pruebas escritas junto con el código en cada historia |
| Preparación del entorno y datos | Postgres en contenedor + semillas de dos tenants |
| Ejecución | Jobs de calidad, integración, E2E, ZAP y k6 |
| Ejecución automática en el pipeline | Todo este diagrama |
| Registro de incidentes | Issues de GitHub creados desde los reportes |
| Monitoreo y control | Resumen de jobs + dashboard de SonarQube |
| Finalización y reporte | Artefactos del pipeline: cobertura, reporte ZAP, reporte k6 |
