# Proceso de alta de una empresa (tenant)

Demuestra que una nueva empresa se incorpora sin crear una aplicación aparte: solo se crean filas nuevas en la misma base de datos.

```mermaid
flowchart TB
  subgraph SA["Superadministrador"]
    S1(["Registra empresa:<br/>nombre, RUC, admin"])
    S2["Revisa datos"]
  end

  subgraph SIS["Plataforma"]
    P1{"¿RUC o email<br/>ya existen?"}
    P2["Muestra error de validación"]
    P3["Crea tenant, configuración<br/>por defecto y horario base"]
    P4["Crea usuario admin_empresa<br/>con contraseña temporal"]
    P5["Registra en auditoría"]
    P6["Envía credenciales"]
  end

  subgraph AD["Admin de empresa"]
    A1["Inicia sesión y cambia contraseña"]
    A2["Vincula WhatsApp con QR"]
    A3["Carga intenciones y FAQs"]
    A4["Crea agentes"]
    A5(["Empresa operativa"])
  end

  S1 --> S2 --> P1
  P1 -- "Sí" --> P2 --> S2
  P1 -- "No" --> P3 --> P4 --> P5 --> P6 --> A1
  A1 --> A2 --> A3 --> A4 --> A5
```
