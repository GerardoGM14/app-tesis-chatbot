# Diagrama de despliegue — VPS Linux

Staging y producción corren en el mismo VPS como dos proyectos de Docker Compose separados, cada uno con su red y sus volúmenes. OWASP ZAP solo ataca staging.

```mermaid
---
title: "C4 · Diagrama de despliegue"
---
flowchart TB
  internet(["<b>Usuarios</b><br/><i>[Internet]</i>"])

  subgraph GH["GitHub · [Nodo: SaaS]"]
    actions["<b>GitHub Actions</b><br/><i>[Runners hospedados]</i><br/>Lint, pruebas, SAST,<br/>build, Trivy"]
    ghcr[("<b>GHCR</b><br/><i>[Registro de imágenes]</i><br/>frontend, backend, wa-service")]
  end

  subgraph VPS["VPS · [Nodo: Ubuntu 24.04 LTS, Docker Engine]"]
    runner["<b>Runner self-hosted</b><br/><i>[systemd, usuario sin root]</i><br/>Despliega y ejecuta<br/>QA y OWASP"]
    caddy["<b>Caddy</b><br/><i>[Puertos 80 y 443]</i><br/>TLS automático,<br/>enruta por dominio"]

    subgraph STG["Compose: staging · [Red stg_net]"]
      stg_app["<b>frontend + backend<br/>+ wa-service</b><br/>Versión candidata"]
      stg_db[("<b>PostgreSQL</b><br/>Volumen stg_pgdata")]
    end

    subgraph PROD["Compose: producción · [Red prod_net]"]
      prod_app["<b>frontend + backend<br/>+ wa-service</b><br/>Versión aprobada"]
      prod_db[("<b>PostgreSQL</b><br/>Volumen prod_pgdata<br/>backup diario")]
    end
  end

  actions -- "Publica imágenes" --> ghcr
  ghcr -- "Descarga imágenes" --> runner
  runner -- "compose up" --> stg_app
  runner -- "compose up<br/>tras aprobación" --> prod_app
  internet -- "HTTPS" --> caddy
  caddy -- "staging.dominio" --> stg_app
  caddy -- "app.dominio" --> prod_app
  stg_app --> stg_db
  prod_app --> prod_db

  classDef infra fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef external fill:#E4E6EB,stroke:#9AA0AC,color:#0F1219
  classDef person fill:#142A97,stroke:#142A97,color:#FFFFFF
  class runner,caddy,stg_app,stg_db,prod_app,prod_db infra
  class actions,ghcr external
  class internet person
  style GH fill:#FFFFFF,stroke:#9AA0AC,color:#0F1219
  style VPS fill:#F7F8FA,stroke:#1834BC,color:#0F1219
  style STG fill:#FFF3EC,stroke:#EB6B1D,stroke-dasharray:6 4,color:#0F1219
  style PROD fill:#EEF1FC,stroke:#1834BC,stroke-dasharray:6 4,color:#0F1219
```

## Endurecimiento del servidor

| Elemento | Configuración |
|---|---|
| Firewall | `ufw` con solo 22, 80 y 443 abiertos |
| SSH | Solo llave, sin login de root, fail2ban |
| Runner | Servicio systemd con usuario dedicado; acepta solo jobs de `main` |
| Puertos de contenedores | Ninguno publicado salvo Caddy |
| Secretos | GitHub Environments + archivo `.env` con permisos `600` |
| Backups | `pg_dump` diario por cron, rotación de 7 días |
