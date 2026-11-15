import { createBrowserRouter, type RouteObject } from 'react-router';
import { ChangePasswordPage, ForcedChangePasswordPage } from '@/presentation/pages/ChangePasswordPage';
import { LoginPage } from '@/presentation/pages/LoginPage';
import { NotFoundPage } from '@/presentation/pages/StatusPages';
import { CHANGE_PASSWORD_PATH, GuestOnly, HomeRedirect, RequirePermission, RequireSession } from './guards';
import { DashboardPage, ModulePendingPage, TenantsPage, UsersPage } from './lazyPages';
import { AppShell } from './layout/AppShell';
import type { RouteHandle } from './layout/Breadcrumbs';

const crumb = (label: string): RouteHandle => ({ crumb: label });

export const routes: RouteObject[] = [
  {
    path: '/login',
    element: (
      <GuestOnly>
        <LoginPage />
      </GuestOnly>
    ),
  },
  {
    path: CHANGE_PASSWORD_PATH,
    element: (
      <RequireSession allowPendingPassword>
        <ForcedChangePasswordPage />
      </RequireSession>
    ),
  },
  {
    path: '/',
    element: (
      <RequireSession>
        <AppShell />
      </RequireSession>
    ),
    handle: crumb('Inicio'),
    children: [
      { index: true, element: <HomeRedirect /> },
      {
        path: 'dashboard',
        handle: crumb('Dashboard'),
        element: (
          <RequirePermission permission="dashboard:ver">
            <DashboardPage />
          </RequirePermission>
        ),
      },
      {
        path: 'conversaciones',
        handle: crumb('Conversaciones'),
        element: (
          <RequirePermission permission="conversaciones:atender">
            <ModulePendingPage module="conversaciones" />
          </RequirePermission>
        ),
      },
      {
        path: 'bot',
        handle: crumb('Bot'),
        element: (
          <RequirePermission permission="bot:gestionar">
            <ModulePendingPage module="bot" />
          </RequirePermission>
        ),
      },
      {
        path: 'canal',
        handle: crumb('Canal de WhatsApp'),
        element: (
          <RequirePermission permission="canal:gestionar">
            <ModulePendingPage module="canal" />
          </RequirePermission>
        ),
      },
      {
        path: 'usuarios',
        handle: crumb('Usuarios'),
        element: (
          <RequirePermission permission="usuarios:gestionar">
            <UsersPage />
          </RequirePermission>
        ),
      },
      {
        path: 'empresas',
        handle: crumb('Empresas'),
        element: (
          <RequirePermission permission="empresas:gestionar">
            <TenantsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'cuenta',
        handle: crumb('Mi cuenta'),
        children: [
          { index: true, element: <HomeRedirect /> },
          { path: 'password', handle: crumb('Cambiar contraseña'), element: <ChangePasswordPage /> },
        ],
      },
      { path: '*', handle: crumb('No encontrada'), element: <NotFoundPage /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
