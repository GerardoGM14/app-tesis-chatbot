# Sistema de diseño

Reglas visuales del panel, que usa solo el personal con cuenta. El objetivo es un diseño sobrio y ordenado que se reconozca como de Sertech.

## Principios

- Sin sombras externas ni degradados. La jerarquía se marca con bordes de 1px, espacio y peso tipográfico.
- Sin emojis. Toda señal visual usa iconos de lucide-react con trazo de 1.75px.
- Poca carga por pantalla: una acción primaria por vista y la información secundaria en paneles o pestañas.
- El naranja se reserva para lo que exige atención; no se usa como decoración.

## Paleta (colores medidos sobre los archivos del logo)

| Token | Claro | Uso |
|---|---|---|
| `brand-600` | `#1834BC` | Botón primario, ítem activo del sidebar, enlaces |
| `brand-700` | `#142A97` | Hover del primario |
| `brand-50` | `#EEF1FC` | Fondo del ítem activo, selección |
| `accent-500` | `#EB6B1D` | Conversaciones derivadas, alertas, insignias de pendiente |
| `neutral-950` | `#0F1219` | Texto principal |
| `neutral-600` | `#4B5263` | Texto secundario |
| `neutral-200` | `#E4E6EB` | Bordes |
| `neutral-50` | `#F7F8FA` | Fondo de la aplicación |
| `white` | `#FFFFFF` | Superficies (tarjetas, tablas, sidebar) |
| `success-600` | `#16803C` | Conectado, resuelta |
| `danger-600` | `#C8312B` | Errores, acciones destructivas |

El modo oscuro reutiliza los mismos tokens con valores propios; los componentes nunca usan colores directos.

## Tipografía

| Rol | Fuente | Peso | Tamaño |
|---|---|---|---|
| Títulos de página | Sora | 600 | 24 px |
| Números de KPI | Sora | 600 | 30 px |
| Títulos de sección y tarjetas | Sora | 600 | 16 / 18 px |
| Cuerpo e interfaz | Poppins | 400 / 500 | 14 px |
| Etiquetas y ayudas | Poppins | 500 | 12 px |

Ambas fuentes se sirven desde el propio frontend (paquetes `@fontsource`), no desde Google Fonts, para no depender de terceros ni romper la CSP.

## Pantalla de login

Es la única pantalla sin sidebar. Se divide en dos columnas en escritorio y en una sola en móvil.

```mermaid
---
title: "Pantalla de login"
---
flowchart LR
  subgraph IZQ["Panel de marca · fondo brand-600 plano"]
    direction TB
    L1["Logo Sertech en blanco"]
    L2["Título en Sora<br/>Atención automatizada<br/>para tu empresa"]
    L3["Las barras diagonales del logo<br/>como único recurso gráfico"]
    L1 --- L2 --- L3
  end

  subgraph DER["Formulario · fondo blanco"]
    direction TB
    D1["Iniciar sesión<br/>Sora 24 px"]
    D2["Email<br/>icono Mail"]
    D3["Contraseña<br/>icono Lock + mostrar u ocultar"]
    D4["Botón Ingresar<br/>brand-600, spinner al enviar"]
    D5["Error genérico en línea<br/>danger-600 con icono"]
    D1 --- D2 --- D3 --- D4 --- D5
  end

  IZQ ~~~ DER

  classDef brand fill:#1834BC,stroke:#1834BC,color:#FFFFFF
  classDef form fill:#FFFFFF,stroke:#C9CDD6,color:#0F1219
  class L1,L2,L3 brand
  class D1,D2,D3,D4,D5 form
  style IZQ fill:#142A97,stroke:#142A97,color:#FFFFFF
  style DER fill:#F7F8FA,stroke:#C9CDD6,color:#0F1219
```

- Sin "Recordarme" ni "Olvidé mi contraseña": el restablecimiento lo hace un administrador (ver [Autenticación y roles](../01-arquitectura/seguridad/autenticacion-y-roles.md)).
- Errores siempre genéricos: "Email o contraseña incorrectos".
- El botón se desactiva mientras se envía, para evitar doble envío.

## Componentes del layout

```mermaid
---
title: "Componentes del layout"
---
flowchart LR
  subgraph SHELL["AppShell"]
    direction LR
    SB["Sidebar<br/>logo, navegación por rol,<br/>colapsable a iconos"]
    subgraph MAIN["Área principal"]
      direction TB
      TB["Topbar<br/>breadcrumb, estado del canal,<br/>menú de usuario"]
      CT["Contenido de la página<br/>skeleton mientras carga"]
    end
    TS["Toaster<br/>abajo a la derecha"]
  end
  SB --- MAIN
  MAIN -.- TS
```

## Navegación del sidebar por rol

| Ítem | Icono | Superadmin | Admin empresa | Agente |
|---|---|---|---|---|
| Dashboard | `LayoutDashboard` | Sí | Sí | Sí |
| Conversaciones | `MessagesSquare` | No | Sí | Sí |
| Bot | `Bot` | No | Sí | No |
| Canal de WhatsApp | `QrCode` | No | Sí | No |
| Usuarios | `Users` | No | Sí | No |
| Empresas | `Building2` | Sí | No | No |
| Configuración | `Settings` | Sí | Sí | No |
