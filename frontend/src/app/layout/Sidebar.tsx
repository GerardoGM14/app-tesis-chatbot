import { clsx } from 'clsx';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { NavLink } from 'react-router';
import { useCurrentUser } from '@/application/hooks/useAuth';
import { BrandMark } from '@/presentation/components/BrandMark';
import { visibleNavItems } from '../navigation';

interface SidebarProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
  onNavigate?: () => void;
}

export function Sidebar({ collapsed, onToggleCollapsed, onNavigate }: Readonly<SidebarProps>) {
  const user = useCurrentUser();
  const items = visibleNavItems(user?.permisos ?? []);

  return (
    <aside
      className={clsx(
        'flex h-full flex-col border-r border-line bg-surface transition-[width] duration-200',
        collapsed ? 'w-[68px]' : 'w-64',
      )}
    >
      <div
        className={clsx(
          'flex h-14 items-center gap-3 border-b border-line',
          collapsed ? 'justify-center' : 'px-4',
        )}
      >
        <BrandMark className="size-8" />
        {!collapsed && (
          <div className="min-w-0 leading-tight">
            <p className="font-display text-sm font-semibold tracking-tight">Sertech</p>
            <p className="truncate text-xs text-ink-subtle">Plataforma de atención</p>
          </div>
        )}
      </div>

      <nav aria-label="Navegación principal" className="flex-1 overflow-y-auto px-3 py-4">
        <ul className="flex flex-col gap-1">
          {items.map(({ to, label, icon: Icon }) => (
            <li key={to}>
              <NavLink
                to={to}
                onClick={onNavigate}
                title={collapsed ? label : undefined}
                className={({ isActive }) =>
                  clsx(
                    'flex h-9 items-center gap-3 rounded-md text-sm transition-colors',
                    collapsed ? 'justify-center' : 'px-3',
                    isActive
                      ? 'bg-brand-50 font-medium text-brand-700'
                      : 'text-ink-muted hover:bg-canvas hover:text-ink',
                  )
                }
              >
                <Icon size={18} strokeWidth={1.75} aria-hidden className="shrink-0" />
                {collapsed ? (
                  <span className="sr-only">{label}</span>
                ) : (
                  <span className="truncate">{label}</span>
                )}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>

      <div className="hidden border-t border-line p-3 lg:block">
        <button
          type="button"
          onClick={onToggleCollapsed}
          aria-label={collapsed ? 'Expandir menú' : 'Contraer menú'}
          title={collapsed ? 'Expandir menú' : 'Contraer menú'}
          className={clsx(
            'flex h-9 w-full items-center gap-3 rounded-md text-sm text-ink-muted hover:bg-canvas hover:text-ink',
            collapsed ? 'justify-center' : 'px-3',
          )}
        >
          {collapsed ? (
            <PanelLeftOpen size={18} strokeWidth={1.75} aria-hidden />
          ) : (
            <>
              <PanelLeftClose size={18} strokeWidth={1.75} aria-hidden />
              <span>Contraer</span>
            </>
          )}
        </button>
      </div>
    </aside>
  );
}
