import { clsx } from 'clsx';
import { LoaderCircle, type LucideIcon } from 'lucide-react';
import type { ButtonHTMLAttributes } from 'react';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md';

const variants: Record<Variant, string> = {
  primary: 'bg-brand-600 text-white hover:bg-brand-700 disabled:bg-brand-600/50',
  secondary: 'border border-line-strong bg-surface text-ink hover:bg-canvas disabled:text-ink-subtle',
  ghost: 'text-ink-muted hover:bg-canvas hover:text-ink disabled:text-ink-subtle',
  danger: 'bg-danger-600 text-white hover:bg-danger-600/90 disabled:bg-danger-600/50',
};

const sizes: Record<Size, string> = {
  sm: 'h-8 gap-1.5 px-3 text-xs',
  md: 'h-10 gap-2 px-4 text-sm',
};

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  icon?: LucideIcon;
  loading?: boolean;
}

export function Button({
  variant = 'primary',
  size = 'md',
  icon: Icon,
  loading = false,
  disabled,
  className,
  children,
  type = 'button',
  ...props
}: Readonly<ButtonProps>) {
  const iconSize = size === 'sm' ? 14 : 16;
  return (
    <button
      type={type}
      disabled={disabled ?? loading}
      aria-busy={loading || undefined}
      className={clsx(
        'inline-flex shrink-0 items-center justify-center rounded-md font-medium transition-colors disabled:cursor-not-allowed',
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    >
      {loading ? (
        <LoaderCircle size={iconSize} className="animate-spin" aria-hidden />
      ) : (
        Icon && <Icon size={iconSize} strokeWidth={1.75} aria-hidden />
      )}
      {children}
    </button>
  );
}

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: LucideIcon;
  label: string;
  tone?: 'default' | 'danger';
}

/** Botón solo con icono. label se usa como texto accesible y tooltip. */
export function IconButton({
  icon: Icon,
  label,
  tone = 'default',
  className,
  ...props
}: Readonly<IconButtonProps>) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={clsx(
        'inline-flex size-8 items-center justify-center rounded-md transition-colors disabled:cursor-not-allowed disabled:opacity-40',
        tone === 'danger'
          ? 'text-ink-muted hover:bg-danger-50 hover:text-danger-600'
          : 'text-ink-muted hover:bg-canvas hover:text-ink',
        className,
      )}
      {...props}
    >
      <Icon size={16} strokeWidth={1.75} aria-hidden />
    </button>
  );
}
