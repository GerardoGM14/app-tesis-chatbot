# Modelo entidad-relación

Modelo físico en PostgreSQL. Todas las tablas con `tenant_id` tienen Row-Level Security activada con la política `tenant_id = current_setting('app.tenant_id')::uuid`.

```mermaid
erDiagram
  tenants ||--o{ usuarios : "tiene"
  tenants ||--o| canales_whatsapp : "conecta"
  tenants ||--|| configuracion_bot : "configura"
  tenants ||--o{ horarios_atencion : "define"
  tenants ||--o{ intenciones : "define"
  tenants ||--o{ contactos : "atiende"
  contactos ||--o{ conversaciones : "inicia"
  conversaciones ||--o{ mensajes : "contiene"
  conversaciones ||--o| encuestas : "recibe"
  usuarios |o--o{ conversaciones : "atiende"
  intenciones |o--o{ mensajes : "detectada en"
  usuarios ||--o{ refresh_tokens : "posee"
  tenants ||--o{ auditoria : "registra"

  tenants {
    uuid id PK
    varchar nombre
    varchar ruc UK
    varchar slug UK
    varchar estado "activo | suspendido"
    timestamptz creado_en
  }
  usuarios {
    uuid id PK
    uuid tenant_id FK "null solo para superadmin"
    varchar nombre
    varchar email UK "minúsculas"
    text password_hash "argon2id"
    varchar rol "superadmin | admin_empresa | agente"
    boolean activo
    boolean debe_cambiar_password
    int intentos_fallidos
    timestamptz bloqueado_hasta
    timestamptz ultimo_acceso
  }
  refresh_tokens {
    uuid id PK
    uuid usuario_id FK
    uuid familia "detecta reutilización"
    text token_hash
    timestamptz expira_en
    timestamptz revocado_en
  }
  canales_whatsapp {
    uuid id PK
    uuid tenant_id FK
    varchar numero
    varchar jid
    varchar estado "desconectado | esperando_qr | conectado"
    timestamptz conectado_en
  }
  configuracion_bot {
    uuid tenant_id PK
    text mensaje_bienvenida
    text mensaje_fuera_horario
    text mensaje_derivacion
    numeric umbral_coincidencia
    int minutos_inactividad
  }
  horarios_atencion {
    uuid id PK
    uuid tenant_id FK
    smallint dia_semana
    time hora_inicio
    time hora_fin
  }
  intenciones {
    uuid id PK
    uuid tenant_id FK
    varchar nombre
    text_array palabras_clave
    text respuesta
    int prioridad
    boolean activa
  }
  contactos {
    uuid id PK
    uuid tenant_id FK
    varchar telefono
    varchar nombre
  }
  conversaciones {
    uuid id PK
    uuid tenant_id FK
    uuid contacto_id FK
    uuid agente_id FK
    varchar estado
    varchar resuelta_por "bot | humano | sin_resolver"
    timestamptz iniciada_en
    timestamptz cerrada_en
  }
  mensajes {
    uuid id PK
    uuid tenant_id FK
    uuid conversacion_id FK
    uuid intencion_id FK
    varchar direccion "entrante | saliente"
    varchar origen "contacto | bot | agente"
    text contenido
    varchar wa_message_id UK
    timestamptz recibido_en
    timestamptz enviado_en
  }
  encuestas {
    uuid conversacion_id PK
    uuid tenant_id FK
    boolean resuelta
    smallint calificacion "1 a 5"
    timestamptz respondida_en
  }
  auditoria {
    bigint id PK
    uuid tenant_id FK
    uuid usuario_id FK
    varchar accion
    varchar entidad
    uuid entidad_id
    inet ip
    timestamptz creado_en
  }
```

## Indicadores calculados desde el modelo

| Indicador | Consulta base |
|---|---|
| % atendidas automáticamente | `conversaciones` cerradas con `resuelta_por = 'bot'` / total cerradas |
| Tiempo promedio de respuesta | Promedio de `enviado_en` de la primera respuesta − `recibido_en` del mensaje entrante |
| % resueltas satisfactoriamente | `encuestas.resuelta = true` / encuestas respondidas |
| Satisfacción | Promedio de `encuestas.calificacion` expresado en % |
