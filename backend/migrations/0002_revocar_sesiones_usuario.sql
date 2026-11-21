-- Cierre de las sesiones de un usuario desde el alcance de su empresa.
--
-- La política de refresh_tokens solo admite el alcance de plataforma, así que un
-- UPDATE hecho por un administrador de empresa afectaba cero filas sin dar error.
-- La función corre con el rol dueño (SECURITY DEFINER) y verifica el alcance: la
-- empresa puede revocar las sesiones de sus usuarios, pero no leer la tabla.

CREATE FUNCTION revocar_sesiones_usuario(p_usuario_id uuid) RETURNS void
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
  AS $$
BEGIN
  IF NOT public.app_is_platform() AND NOT EXISTS (
    SELECT 1 FROM public.usuarios
    WHERE id = p_usuario_id AND tenant_id = public.app_tenant_id()
  ) THEN
    RAISE EXCEPTION 'el usuario no pertenece al alcance de la transacción'
      USING ERRCODE = 'insufficient_privilege';
  END IF;

  UPDATE public.refresh_tokens SET revocado_en = now()
  WHERE usuario_id = p_usuario_id AND revocado_en IS NULL;
END
$$;

REVOKE ALL ON FUNCTION revocar_sesiones_usuario(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION revocar_sesiones_usuario(uuid) TO plataforma_app;
