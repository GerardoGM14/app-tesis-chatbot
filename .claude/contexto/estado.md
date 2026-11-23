# Estado del proyecto

## Sprint 1 · terminado

- Backend: autenticación completa, empresas, usuarios, RLS, auditoría, migración inicial, OpenAPI
  (`backend/api/openapi.yaml`), pruebas unitarias y de integración (AUT y AIS).
- Frontend: login, cambio obligatorio de contraseña, layout por rol, dashboard (KPIs vacíos),
  páginas de Empresas y Usuarios, módulos pendientes con estado vacío. 57 pruebas, ~87 % de cobertura.
- wa-service: solo esqueleto (health + rutas internas con HMAC que responden 501).
- Docker, compose local y de VPS, Caddy, SonarQube, CI (`ci.yml`) y CD (`cd.yml`) con OWASP ZAP.
- Corrección del 2026-11-23: un admin de empresa no cerraba las sesiones de sus trabajadores (RLS);
  se resolvió con la función `revocar_sesiones_usuario` (migración 0002). CI corregido localmente;
  falta confirmarlo en GitHub.

## Pendientes técnicos

- La firma HMAC de wa-service no cubre método ni ruta, y el tenant va en la ruta: corregir antes de
  integrar whatsmeow.
- `mensajes.wa_message_id` debe ser único por tenant, no global.
- Casos AIS de `docs/04-calidad`: alinear con las pruebas (faltan 02, 08, 09 y 10; 03, 04 y 05 difieren).

## Siguientes sprints

1. **Sprint 2-3:** whatsmeow en wa-service (vinculación QR, sesiones por tenant, cola de reintentos),
   tablas de contactos, conversaciones y mensajes con RLS, bandeja en vivo con SSE.
2. **Sprint 4:** motor del bot (intenciones, FAQs, menú, horario), derivación a humano, encuesta.
3. **Sprint 5:** métricas reales del dashboard, Playwright E2E y k6 en el pipeline.

## Pendientes del informe de tesis (PTI_CorregidoII.docx, no está en el repo)

Recordárselos al usuario cuando vuelva al informe:

1. Justificación tecnológica y marco teórico 2.7.6.5 dicen backend en **Next.js**; se implementó en Go.
2. Al final de 2.8.1 quedó texto de IA: "...consistente con el planteamiento de **tu** investigación". Borrar.
3. Población = 55 clientes (personas) vs. muestra = 20 empresas clientes: unidades distintas.
   Además aparece "Sertech E.I.R.L." y "Sertech Perú E.I.R.L.".
4. Antigravity está en el marco teórico (2.7.6.9) pero no en la metodología.
5. Agregar SonarQube, OWASP ZAP, Trivy, Semgrep, k6 y Playwright en la sección de pruebas.
6. El logo dice "SERTCH" y el banner "sertech": confirmar el nombre oficial.
