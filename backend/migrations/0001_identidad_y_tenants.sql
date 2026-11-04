-- Empresas (tenants), usuarios, sesiones y auditoría.
-- El backend se conecta con el rol plataforma_app, que NO es dueño de las
-- tablas ni tiene BYPASSRLS: las políticas de Row-Level Security le aplican
-- siempre. Las migraciones corren con el rol dueño.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'plataforma_app') THEN
    CREATE ROLE plataforma_app NOLOGIN;
  END IF;
END
$$;

-- Alcance de la transacción, fijado por el backend con SET LOCAL.
CREATE FUNCTION app_tenant_id() RETURNS uuid
  LANGUAGE sql STABLE
  AS $$ SELECT NULLIF(current_setting('app.tenant_id', true), '')::uuid $$;

CREATE FUNCTION app_is_platform() RETURNS boolean
  LANGUAGE sql STABLE
  AS $$ SELECT coalesce(current_setting('app.is_platform', true), 'off') = 'on' $$;

CREATE TABLE tenants (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  nombre         varchar(120) NOT NULL,
  ruc            char(11)     NOT NULL,
  slug           varchar(80)  NOT NULL,
  estado         varchar(20)  NOT NULL DEFAULT 'activo',
  creado_en      timestamptz  NOT NULL DEFAULT now(),
  actualizado_en timestamptz  NOT NULL DEFAULT now(),
  CONSTRAINT tenants_ruc_key  UNIQUE (ruc),
  CONSTRAINT tenants_slug_key UNIQUE (slug),
  CONSTRAINT tenants_estado_check CHECK (estado IN ('activo', 'suspendido'))
);

CREATE TABLE usuarios (
  id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             uuid REFERENCES tenants (id) ON DELETE RESTRICT,
  nombre                varchar(120) NOT NULL,
  email                 varchar(254) NOT NULL,
  password_hash         text         NOT NULL,
  rol                   varchar(20)  NOT NULL,
  activo                boolean      NOT NULL DEFAULT true,
  debe_cambiar_password boolean      NOT NULL DEFAULT true,
  intentos_fallidos     integer      NOT NULL DEFAULT 0,
  bloqueado_hasta       timestamptz,
  ultimo_acceso         timestamptz,
  creado_en             timestamptz  NOT NULL DEFAULT now(),
  actualizado_en        timestamptz  NOT NULL DEFAULT now(),
  CONSTRAINT usuarios_rol_check CHECK (rol IN ('superadmin', 'admin_empresa', 'agente')),
  -- Solo el superadmin vive fuera de una empresa.
  CONSTRAINT usuarios_tenant_por_rol CHECK ((rol = 'superadmin') = (tenant_id IS NULL)),
  CONSTRAINT usuarios_email_minusculas CHECK (email = lower(email))
);
CREATE UNIQUE INDEX usuarios_email_key ON usuarios (email);
CREATE INDEX usuarios_tenant_idx ON usuarios (tenant_id);

CREATE TABLE refresh_tokens (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  usuario_id  uuid        NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
  familia     uuid        NOT NULL,
  token_hash  bytea       NOT NULL,
  expira_en   timestamptz NOT NULL,
  revocado_en timestamptz,
  creado_en   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT refresh_tokens_hash_key UNIQUE (token_hash)
);
CREATE INDEX refresh_tokens_usuario_idx ON refresh_tokens (usuario_id);
CREATE INDEX refresh_tokens_familia_idx ON refresh_tokens (familia);

CREATE TABLE auditoria (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  tenant_id  uuid REFERENCES tenants (id) ON DELETE SET NULL,
  usuario_id uuid REFERENCES usuarios (id) ON DELETE SET NULL,
  accion     varchar(60) NOT NULL,
  entidad    varchar(60),
  entidad_id uuid,
  ip         inet,
  detalle    jsonb,
  creado_en  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX auditoria_tenant_fecha_idx ON auditoria (tenant_id, creado_en DESC);

-- Row-Level Security --------------------------------------------------------

ALTER TABLE tenants        ENABLE ROW LEVEL SECURITY;
ALTER TABLE usuarios       ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE auditoria      ENABLE ROW LEVEL SECURITY;

-- Una empresa ve su propia fila; solo la plataforma crea o modifica empresas.
CREATE POLICY tenants_lectura ON tenants FOR SELECT
  USING (app_is_platform() OR id = app_tenant_id());
CREATE POLICY tenants_escritura ON tenants FOR ALL
  USING (app_is_platform()) WITH CHECK (app_is_platform());

CREATE POLICY usuarios_aislamiento ON usuarios
  USING (app_is_platform() OR tenant_id = app_tenant_id())
  WITH CHECK (app_is_platform() OR tenant_id = app_tenant_id());

-- Las sesiones solo se manejan desde el módulo de autenticación (plataforma).
CREATE POLICY refresh_tokens_plataforma ON refresh_tokens
  USING (app_is_platform()) WITH CHECK (app_is_platform());

CREATE POLICY auditoria_aislamiento ON auditoria
  USING (app_is_platform() OR tenant_id = app_tenant_id())
  WITH CHECK (app_is_platform() OR tenant_id = app_tenant_id());

-- Permisos del rol de la aplicación -----------------------------------------

GRANT USAGE ON SCHEMA public TO plataforma_app;
GRANT SELECT, INSERT, UPDATE ON tenants, usuarios, refresh_tokens TO plataforma_app;
-- La auditoría es de solo inserción para la aplicación.
GRANT SELECT, INSERT ON auditoria TO plataforma_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO plataforma_app;
