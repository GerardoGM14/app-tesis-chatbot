import { clsx } from 'clsx';
import type { LucideIcon } from 'lucide-react';
import type { ReactNode } from 'react';

export function PageHeader({
  title,
  description,
  actions,
}: Readonly<{
  title: string;
  description?: string;
  actions?: ReactNode;
}>) {
  return (
    <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {description && <p className="mt-1 text-sm text-ink-muted">{description}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}

export function Card({ children, className }: Readonly<{ children: ReactNode; className?: string }>) {
  return (
    <section className={clsx('rounded-lg border border-line bg-surface', className)}>{children}</section>
  );
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
}: Readonly<{
  icon: LucideIcon;
  title: string;
  description: string;
  action?: ReactNode;
}>) {
  return (
    <div className="flex flex-col items-center px-6 py-14 text-center">
      <div className="mb-4 flex size-11 items-center justify-center rounded-lg border border-line bg-canvas text-brand-600">
        <Icon size={20} strokeWidth={1.75} aria-hidden />
      </div>
      <h2 className="text-base font-semibold">{title}</h2>
      <p className="mt-1 max-w-sm text-sm text-ink-muted">{description}</p>
      {action && <div className="mt-5">{action}</div>}
    </div>
  );
}

export function Table({ children, label }: Readonly<{ children: ReactNode; label: string }>) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[640px] border-collapse text-left text-sm" aria-label={label}>
        {children}
      </table>
    </div>
  );
}

export function Th({ children, className }: Readonly<{ children?: ReactNode; className?: string }>) {
  return (
    <th
      scope="col"
      className={clsx(
        'border-b border-line bg-canvas px-4 py-2.5 text-xs font-medium tracking-wide text-ink-muted uppercase',
        className,
      )}
    >
      {children}
    </th>
  );
}

export function Td({ children, className }: Readonly<{ children?: ReactNode; className?: string }>) {
  return <td className={clsx('border-b border-line px-4 py-3 align-middle', className)}>{children}</td>;
}

export function ErrorState({ message, onRetry }: Readonly<{ message: string; onRetry?: () => void }>) {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-12 text-center" role="alert">
      <p className="text-sm text-ink-muted">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="text-sm font-medium text-brand-600 hover:underline"
        >
          Reintentar
        </button>
      )}
    </div>
  );
}
