import { clsx } from 'clsx';
import { Building2, Menu } from 'lucide-react';
import { Suspense, useState } from 'react';
import { Outlet } from 'react-router';
import { useCurrentUser } from '@/application/hooks/useAuth';
import { PageSkeleton } from '@/presentation/components/skeletons/Skeletons';
import { Breadcrumbs } from './Breadcrumbs';
import { Sidebar } from './Sidebar';
import { UserMenu } from './UserMenu';

const COLLAPSED_KEY = 'sidebar-contraido';

/** Preferencia por usuario y navegador; si el almacenamiento falla se ignora. */
function readCollapsed(): boolean {
  try {
    return window.localStorage.getItem(COLLAPSED_KEY) === '1';
  } catch {
    return false;
  }
}

function writeCollapsed(value: boolean): void {
  try {
    window.localStorage.setItem(COLLAPSED_KEY, value ? '1' : '0');
  } catch {
    // Almacenamiento no disponible (modo privado): la preferencia no se guarda.
  }
}

export function AppShell() {
  const user = useCurrentUser();
  const [collapsed, setCollapsed] = useState(readCollapsed);
  const [mobileOpen, setMobileOpen] = useState(false);

  const toggleCollapsed = () => {
    setCollapsed((value) => {
      writeCollapsed(!value);
      return !value;
    });
  };

  return (
    <div className="flex h-dvh overflow-hidden">
      <div className="hidden lg:block">
        <Sidebar collapsed={collapsed} onToggleCollapsed={toggleCollapsed} />
      </div>

      {mobileOpen && (
        <div className="fixed inset-0 z-40 lg:hidden">
          <button
            type="button"
            aria-label="Cerrar menú"
            className="absolute inset-0 bg-ink/40"
            onClick={() => {
              setMobileOpen(false);
            }}
          />
          <div className="relative h-full w-64">
            <Sidebar
              collapsed={false}
              onToggleCollapsed={toggleCollapsed}
              onNavigate={() => {
                setMobileOpen(false);
              }}
            />
          </div>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center justify-between gap-4 border-b border-line bg-surface px-4 lg:px-6">
          <div className="flex min-w-0 items-center gap-3">
            <button
              type="button"
              aria-label="Abrir menú"
              className="flex size-8 items-center justify-center rounded-md text-ink-muted hover:bg-canvas lg:hidden"
              onClick={() => {
                setMobileOpen(true);
              }}
            >
              <Menu size={18} strokeWidth={1.75} aria-hidden />
            </button>
            <Breadcrumbs />
          </div>
          <div className="flex items-center gap-3">
            {user?.tenantNombre && (
              <span
                className={clsx(
                  'hidden items-center gap-1.5 rounded-md border border-line px-2.5 py-1 text-xs text-ink-muted md:flex',
                )}
              >
                <Building2 size={14} strokeWidth={1.75} aria-hidden />
                <span className="max-w-48 truncate">{user.tenantNombre}</span>
              </span>
            )}
            <UserMenu />
          </div>
        </header>

        <main className="flex-1 overflow-y-auto">
          <div className="mx-auto w-full max-w-7xl px-4 py-6 lg:px-8 lg:py-8">
            <Suspense fallback={<PageSkeleton />}>
              <Outlet />
            </Suspense>
          </div>
        </main>
      </div>
    </div>
  );
}
