import { ChevronDown, KeyRound, LogOut } from 'lucide-react';
import { useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { useCurrentUser, useLogout } from '@/application/hooks/useAuth';
import { ROLE_LABEL } from '@/domain/permissions';

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join('');
}

export function UserMenu() {
  const user = useCurrentUser();
  const logout = useLogout();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useEffect(() => {
    if (!open) return;
    const onPointer = (event: PointerEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('pointerdown', onPointer);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('pointerdown', onPointer);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  if (!user) return null;

  const itemClass =
    'flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-left text-sm text-ink hover:bg-canvas';

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        onClick={() => {
          setOpen((v) => !v);
        }}
        className="flex items-center gap-2.5 rounded-md py-1 pr-2 pl-1 hover:bg-canvas"
      >
        <span className="flex size-8 items-center justify-center rounded-md bg-brand-50 text-xs font-semibold text-brand-700">
          {initials(user.nombre)}
        </span>
        <span className="hidden text-left leading-tight sm:block">
          <span className="block max-w-40 truncate text-sm font-medium">{user.nombre}</span>
          <span className="block text-xs text-ink-subtle">{ROLE_LABEL[user.rol]}</span>
        </span>
        <ChevronDown size={14} className="text-ink-subtle" aria-hidden />
      </button>

      {open && (
        <div
          id={menuId}
          role="menu"
          className="absolute top-full right-0 z-30 mt-1 w-60 rounded-lg border border-line bg-surface p-1.5"
        >
          <div className="border-b border-line px-2.5 pt-1.5 pb-2.5">
            <p className="truncate text-sm font-medium">{user.nombre}</p>
            <p className="truncate text-xs text-ink-subtle">{user.email}</p>
          </div>
          <div className="pt-1.5">
            <button
              type="button"
              role="menuitem"
              className={itemClass}
              onClick={() => {
                setOpen(false);
                void navigate('/cuenta/password');
              }}
            >
              <KeyRound size={16} strokeWidth={1.75} className="text-ink-muted" aria-hidden />
              Cambiar contraseña
            </button>
            <button
              type="button"
              role="menuitem"
              className={itemClass}
              onClick={() => {
                setOpen(false);
                logout.mutate();
              }}
            >
              <LogOut size={16} strokeWidth={1.75} className="text-ink-muted" aria-hidden />
              Cerrar sesión
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
