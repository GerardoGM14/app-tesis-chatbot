# Casos de uso · Autenticación

Todo usuario interno debe iniciar sesión para usar el panel. Agente, Admin de empresa y Superadmin son especializaciones de **Usuario interno** y heredan sus casos de uso.

```mermaid
---
title: "Casos de uso · Autenticación"
---
flowchart LR
  UI(("Usuario<br/>interno"))
  AG(("Agente"))
  AD(("Admin de<br/>empresa"))
  SA(("Super<br/>admin"))

  AG -- "«generaliza»" --> UI
  AD -- "«generaliza»" --> UI
  SA -- "«generaliza»" --> UI

  subgraph SYS["Plataforma de atención multitenant"]
    UC05(["CU05 Iniciar sesión"])
    UC06(["CU06 Cerrar sesión"])
    UC07(["CU07 Cambiar contraseña"])
    UC08(["CU08 Restablecer contraseña<br/>de un usuario"])
  end

  UI --- UC05
  UI --- UC06
  UI --- UC07
  UC07 -. "«extend»<br/>primer ingreso" .-> UC05
  AD --- UC08
  SA --- UC08

  classDef actor fill:#142A97,stroke:#142A97,color:#FFFFFF
  classDef abstract fill:#FFFFFF,stroke:#142A97,color:#142A97,stroke-dasharray:4 3
  classDef usecase fill:#EEF1FC,stroke:#1834BC,color:#0F1219
  class AG,AD,SA actor
  class UI abstract
  class UC05,UC06,UC07,UC08 usecase
  style SYS fill:#FFFFFF,stroke:#1834BC,color:#0F1219
```

## Especificación de CU05 · Iniciar sesión

| Campo | Detalle |
|---|---|
| Actor | Usuario interno (Agente, Admin de empresa, Superadmin) |
| Precondición | El usuario existe, está activo y su empresa no está suspendida |
| Flujo principal | 1. Ingresa email y contraseña. 2. El sistema valida las credenciales. 3. Emite la sesión. 4. Redirige a la pantalla inicial de su rol |
| Flujo alterno A | Credenciales inválidas: mensaje genérico, suma un intento fallido |
| Flujo alterno B | 5 intentos fallidos: cuenta bloqueada 15 minutos |
| Flujo alterno C | Usuario inactivo o empresa suspendida: mensaje genérico, sin revelar el motivo |
| Extensión | Primer ingreso o contraseña restablecida: se obliga a ejecutar CU07 antes de continuar |
| Postcondición | Sesión activa y registro en auditoría |
