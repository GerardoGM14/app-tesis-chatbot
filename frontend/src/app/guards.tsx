import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router';
import { useSessionStore } from '@/application/stores/sessionStore';
import type { Permission } from '@/domain/entities';
import { can, homePathFor } from '@/domain/permissions';
import { AppSkeleton } from '@/presentation/components/skeletons/Skeletons';
import { ForbiddenPage } from '@/presentation/pages/StatusPages';

export const CHANGE_PASSWORD_PATH = '/cambiar-password';

/**
 * Exige sesión. Si el usuario tiene pendiente el cambio obligatorio de
 * contraseña, solo puede estar en la pantalla de cambio.
 */
export function RequireSession({
  children,
  allowPendingPassword = false,
}: Readonly<{
  children: ReactNode;
  allowPendingPassword?: boolean;
}>) {
  const status = useSessionStore((s) => s.status);
  const user = useSessionStore((s) => s.user);
  const location = useLocation();

  if (status === 'loading') return <AppSkeleton />;
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  if (user.debeCambiarPassword && !allowPendingPassword)
    return <Navigate to={CHANGE_PASSWORD_PATH} replace />;
  if (!user.debeCambiarPassword && allowPendingPassword)
    return <Navigate to={homePathFor(user.rol)} replace />;
  return <>{children}</>;
}

/** Solo para visitantes sin sesión (login). */
export function GuestOnly({ children }: Readonly<{ children: ReactNode }>) {
  const status = useSessionStore((s) => s.status);
  const user = useSessionStore((s) => s.user);

  if (status === 'loading') return <AppSkeleton />;
  if (user)
    return <Navigate to={user.debeCambiarPassword ? CHANGE_PASSWORD_PATH : homePathFor(user.rol)} replace />;
  return <>{children}</>;
}

/** Muestra la página solo si el rol tiene el permiso. El backend vuelve a verificarlo. */
export function RequirePermission({
  permission,
  children,
}: Readonly<{ permission: Permission; children: ReactNode }>) {
  const user = useSessionStore((s) => s.user);
  return can(user, permission) ? <>{children}</> : <ForbiddenPage />;
}

/** Redirige la raíz a la pantalla inicial del rol. */
export function HomeRedirect() {
  const user = useSessionStore((s) => s.user);
  return <Navigate to={user ? homePathFor(user.rol) : '/login'} replace />;
}
