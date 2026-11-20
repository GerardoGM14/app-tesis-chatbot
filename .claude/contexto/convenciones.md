# Convenciones

## Diseño de la interfaz

- Elegante y ordenado, **sin aspecto genérico de IA**, sin sobrecargar las pantallas.
- **Sin sombras externas ni degradados.** Jerarquía con bordes de 1px, espacio y tipografía.
- **Sin emojis**: solo iconos de lucide-react (trazo 1.75).
- Sidebar colapsable, breadcrumb, skeleton screens con la forma del contenido, toasts abajo a la derecha.
- Colores medidos del logo: azul `#1834BC` (brand-600) y naranja `#EB6B1D` (accent-500, solo para
  lo que exige atención). Tokens en `frontend/src/index.css`.
- Tipografía: **Sora** para títulos y KPIs, **Poppins** para el resto (subconjunto latino).
- Todo texto visible al usuario en español.

## Calidad (debe pasar SonarQube)

- Frontend: TypeScript estricto, ESLint con `strictTypeChecked` + `eslint-plugin-sonarjs`, Prettier.
  Props de componentes como `Readonly<...>`, sin ternarios anidados, sin `any`.
- Go: `gofmt`, golangci-lint v2 (`.golangci.yml`), complejidad cognitiva máxima 15.
- Cobertura mínima 80 % en código nuevo. Duplicación < 3 %.
- Quality gate y casos de prueba (AUT-xx, AIS-xx) en `docs/04-calidad/estrategia-de-pruebas.md`.

## Backend

- Errores de negocio con `internal/platform/apperr` (sin depender de HTTP); `httpx` los traduce.
- Toda consulta pasa por `database.InTx(ctx, scope, fn)`.
- Validación con `internal/platform/validate`; JSON con `DisallowUnknownFields` y límite de 1 MB.
- Acciones de seguridad se registran en la tabla `auditoria` dentro de la misma transacción.
- Las migraciones solo avanzan y deben ser compatibles con la versión anterior (rollback seguro).

## Documentación

- Diagramas en Mermaid dentro de `docs/**.md`. Exportar con `cd docs && npm run render:png`
  (genera `docs/exports/png/` con la misma estructura de carpetas).
- Al cambiar arquitectura o decisiones, actualizar los diagramas y el ADR correspondiente.

## Git

- Conventional Commits en español: `feat(backend): ...`, `fix(frontend): ...`, `docs: ...`, `ci: ...`.
- Sin línea `Co-Authored-By`.
