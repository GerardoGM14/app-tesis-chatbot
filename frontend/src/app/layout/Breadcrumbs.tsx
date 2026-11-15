import { ChevronRight } from 'lucide-react';
import { Fragment } from 'react';
import { Link, useMatches } from 'react-router';

export interface RouteHandle {
  crumb?: string;
}

function crumbOf(handle: unknown): string | undefined {
  if (typeof handle === 'object' && handle !== null && 'crumb' in handle) {
    const crumb = (handle as RouteHandle).crumb;
    return typeof crumb === 'string' ? crumb : undefined;
  }
  return undefined;
}

/** Breadcrumb generado desde el `handle.crumb` de cada ruta coincidente. */
export function Breadcrumbs() {
  const crumbs = useMatches()
    .map((match) => ({ label: crumbOf(match.handle), path: match.pathname }))
    .filter((c): c is { label: string; path: string } => c.label !== undefined);

  return (
    <nav aria-label="Ruta de navegación" className="min-w-0">
      <ol className="flex items-center gap-1.5 text-sm">
        {crumbs.map((crumb, index) => {
          const last = index === crumbs.length - 1;
          return (
            <Fragment key={crumb.path + crumb.label}>
              {index > 0 && <ChevronRight size={14} className="shrink-0 text-ink-subtle" aria-hidden />}
              <li className="truncate">
                {last ? (
                  <span aria-current="page" className="font-medium text-ink">
                    {crumb.label}
                  </span>
                ) : (
                  <Link to={crumb.path} className="text-ink-muted hover:text-ink">
                    {crumb.label}
                  </Link>
                )}
              </li>
            </Fragment>
          );
        })}
      </ol>
    </nav>
  );
}
