import { lazy } from 'react';

// Las páginas del panel se cargan bajo demanda para que el login abra rápido.

export const DashboardPage = lazy(() =>
  import('@/presentation/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);

export const TenantsPage = lazy(() =>
  import('@/presentation/pages/TenantsPage').then((m) => ({ default: m.TenantsPage })),
);

export const UsersPage = lazy(() =>
  import('@/presentation/pages/UsersPage').then((m) => ({ default: m.UsersPage })),
);

export const ModulePendingPage = lazy(() =>
  import('@/presentation/pages/ModulePendingPage').then((m) => ({ default: m.ModulePendingPage })),
);
