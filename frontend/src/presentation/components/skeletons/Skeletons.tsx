import { clsx } from 'clsx';

/** Bloque base del skeleton screen: misma forma que el contenido que reemplaza. */
export function Skeleton({ className }: Readonly<{ className?: string }>) {
  return <div className={clsx('animate-shimmer rounded-md bg-line', className)} aria-hidden />;
}

export function TableSkeleton({ rows = 5, columns = 4 }: Readonly<{ rows?: number; columns?: number }>) {
  return (
    <div role="status" aria-label="Cargando datos">
      <div className="flex gap-4 border-b border-line bg-canvas px-4 py-3">
        {Array.from({ length: columns }, (_, c) => (
          <Skeleton key={c} className="h-3 flex-1" />
        ))}
      </div>
      {Array.from({ length: rows }, (_, r) => (
        <div key={r} className="flex items-center gap-4 border-b border-line px-4 py-4">
          {Array.from({ length: columns }, (_, c) => (
            <Skeleton key={c} className={clsx('h-3.5 flex-1', c === 0 && 'max-w-60')} />
          ))}
        </div>
      ))}
    </div>
  );
}

export function KpiSkeleton({ count = 4 }: Readonly<{ count?: number }>) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4" role="status" aria-label="Cargando indicadores">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="rounded-lg border border-line bg-surface p-5">
          <Skeleton className="h-3 w-28" />
          <Skeleton className="mt-4 h-8 w-20" />
          <Skeleton className="mt-3 h-3 w-36" />
        </div>
      ))}
    </div>
  );
}

/** Contenido de una página mientras se descarga su código. */
export function PageSkeleton() {
  return (
    <div role="status" aria-label="Cargando página">
      <Skeleton className="h-7 w-56" />
      <Skeleton className="mt-2 h-3 w-80" />
      <div className="mt-8 rounded-lg border border-line bg-surface">
        <TableSkeleton rows={4} />
      </div>
    </div>
  );
}

/** Pantalla mientras se restaura la sesión: reproduce el layout para evitar saltos. */
export function AppSkeleton() {
  return (
    <div className="flex min-h-dvh" role="status" aria-label="Cargando la aplicación">
      <div className="hidden w-64 shrink-0 border-r border-line bg-surface p-4 lg:block">
        <Skeleton className="h-9 w-36" />
        <div className="mt-8 flex flex-col gap-3">
          {Array.from({ length: 6 }, (_, i) => (
            <Skeleton key={i} className="h-8 w-full" />
          ))}
        </div>
      </div>
      <div className="flex-1">
        <div className="flex h-14 items-center border-b border-line bg-surface px-6">
          <Skeleton className="h-3 w-48" />
        </div>
        <div className="p-6">
          <Skeleton className="h-7 w-56" />
          <Skeleton className="mt-2 h-3 w-80" />
          <div className="mt-8">
            <KpiSkeleton />
          </div>
        </div>
      </div>
    </div>
  );
}
