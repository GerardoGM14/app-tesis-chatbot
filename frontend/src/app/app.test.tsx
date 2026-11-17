import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it } from 'vitest';
import { useSessionStore } from '@/application/stores/sessionStore';
import { apiError, makeUser, mockApi, renderApp, signIn, signOut } from '@/test/utils';

describe('autenticación y rutas', () => {
  beforeEach(() => {
    signOut();
  });

  it('muestra el skeleton mientras se restaura la sesión', () => {
    useSessionStore.setState({ status: 'loading' });
    renderApp('/dashboard');
    expect(screen.getByRole('status', { name: 'Cargando la aplicación' })).toBeInTheDocument();
  });

  it('envía al login a quien no tiene sesión', async () => {
    const { router } = renderApp('/usuarios');
    expect(await screen.findByRole('heading', { name: 'Iniciar sesión' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/login');
  });

  it('valida el formulario antes de enviar', async () => {
    const calls = mockApi({});
    renderApp('/login');
    await userEvent.click(await screen.findByRole('button', { name: 'Ingresar' }));
    expect(await screen.findByText('Ingresa un email válido')).toBeInTheDocument();
    expect(screen.getByText('Ingresa tu contraseña')).toBeInTheDocument();
    expect(calls).toHaveLength(0);
  });

  it('muestra un error genérico con credenciales inválidas', async () => {
    mockApi({
      'POST /auth/login': apiError(401, 'credenciales_invalidas', 'Email o contraseña incorrectos'),
    });
    renderApp('/login');
    await userEvent.type(await screen.findByLabelText('Email'), 'ana@empresa.pe');
    await userEvent.type(screen.getByLabelText('Contraseña'), 'incorrecta');
    await userEvent.click(screen.getByRole('button', { name: 'Ingresar' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Email o contraseña incorrectos');
  });

  it('muestra u oculta la contraseña', async () => {
    renderApp('/login');
    const input = await screen.findByLabelText('Contraseña');
    expect(input).toHaveAttribute('type', 'password');
    await userEvent.click(screen.getByRole('button', { name: 'Mostrar contraseña' }));
    expect(input).toHaveAttribute('type', 'text');
  });

  it('lleva a cada rol a su pantalla inicial tras el login', async () => {
    const superadmin = makeUser('superadmin');
    mockApi({
      'POST /auth/login': { body: { accessToken: 't', expiresIn: 900, user: superadmin } },
      'GET /empresas': { body: { items: [] } },
    });
    const { router } = renderApp('/login');
    await userEvent.type(await screen.findByLabelText('Email'), superadmin.email);
    await userEvent.type(screen.getByLabelText('Contraseña'), 'Clave-Correcta-2026');
    await userEvent.click(screen.getByRole('button', { name: 'Ingresar' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/empresas');
    });
  });

  it('obliga a cambiar la contraseña temporal antes de continuar', async () => {
    signIn(makeUser('agente', { debeCambiarPassword: true }));
    mockApi({
      'POST /auth/password': {
        body: { accessToken: 't2', expiresIn: 900, user: makeUser('agente') },
      },
    });
    const { router } = renderApp('/dashboard');
    expect(await screen.findByRole('heading', { name: 'Crea tu contraseña' })).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText('Contraseña actual'), 'Temporal-123');
    await userEvent.type(screen.getByLabelText('Nueva contraseña'), 'Nueva-Clave-2026');
    await userEvent.type(screen.getByLabelText('Repite la nueva contraseña'), 'Nueva-Clave-2026');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar y continuar' }));

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/conversaciones');
    });
  });

  it('muestra los errores de validación del servidor en el campo', async () => {
    signIn(makeUser('agente', { debeCambiarPassword: true }));
    mockApi({
      'POST /auth/password': apiError(422, 'validacion', 'Revisa los datos', {
        actual: 'La contraseña actual no es correcta',
      }),
    });
    renderApp('/cambiar-password');
    await userEvent.type(await screen.findByLabelText('Contraseña actual'), 'Mala-123');
    await userEvent.type(screen.getByLabelText('Nueva contraseña'), 'Nueva-Clave-2026');
    await userEvent.type(screen.getByLabelText('Repite la nueva contraseña'), 'Nueva-Clave-2026');
    await userEvent.click(screen.getByRole('button', { name: 'Guardar y continuar' }));
    expect(await screen.findByText('La contraseña actual no es correcta')).toBeInTheDocument();
  });

  it('impide volver al login con sesión iniciada', async () => {
    signIn(makeUser('admin_empresa'));
    mockApi({ 'GET /usuarios': { body: { items: [] } } });
    const { router } = renderApp('/login');
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/dashboard');
    });
  });
});

describe('layout por rol', () => {
  it('el sidebar del agente solo muestra sus módulos', async () => {
    signIn(makeUser('agente'));
    renderApp('/conversaciones');
    const nav = await screen.findByRole('navigation', { name: 'Navegación principal' });
    const links = within(nav)
      .getAllByRole('link')
      .map((a) => a.textContent);
    expect(links).toEqual(['Dashboard', 'Conversaciones']);
  });

  it('el superadmin ve Empresas y no Conversaciones', async () => {
    signIn(makeUser('superadmin'));
    mockApi({ 'GET /empresas': { body: { items: [] } } });
    renderApp('/empresas');
    const nav = await screen.findByRole('navigation', { name: 'Navegación principal' });
    expect(within(nav).getByRole('link', { name: 'Empresas' })).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: 'Conversaciones' })).not.toBeInTheDocument();
  });

  it('muestra la página de acceso denegado si el rol no tiene permiso', async () => {
    signIn(makeUser('agente'));
    renderApp('/usuarios');
    expect(await screen.findByText('No tienes acceso a esta sección')).toBeInTheDocument();
  });

  it('arma el breadcrumb desde las rutas', async () => {
    signIn(makeUser('admin_empresa'));
    renderApp('/cuenta/password');
    const crumbs = await screen.findByRole('navigation', { name: 'Ruta de navegación' });
    expect(within(crumbs).getByRole('link', { name: 'Inicio' })).toBeInTheDocument();
    expect(within(crumbs).getByRole('link', { name: 'Mi cuenta' })).toBeInTheDocument();
    expect(within(crumbs).getByText('Cambiar contraseña')).toHaveAttribute('aria-current', 'page');
  });

  it('muestra 404 en rutas inexistentes', async () => {
    signIn(makeUser('agente'));
    renderApp('/no-existe');
    expect(await screen.findByText('Página no encontrada')).toBeInTheDocument();
  });

  it('contrae el sidebar y recuerda la preferencia', async () => {
    signIn(makeUser('agente'));
    renderApp('/conversaciones');
    await userEvent.click(await screen.findByRole('button', { name: 'Contraer menú' }));
    expect(screen.getByRole('button', { name: 'Expandir menú' })).toBeInTheDocument();
    expect(window.localStorage.getItem('sidebar-contraido')).toBe('1');
    await userEvent.click(screen.getByRole('button', { name: 'Expandir menú' }));
    expect(window.localStorage.getItem('sidebar-contraido')).toBe('0');
  });

  it('abre el menú lateral en móvil', async () => {
    signIn(makeUser('agente'));
    renderApp('/conversaciones');
    await userEvent.click(await screen.findByRole('button', { name: 'Abrir menú' }));
    expect(screen.getAllByRole('navigation', { name: 'Navegación principal' })).toHaveLength(2);
    await userEvent.click(screen.getByRole('button', { name: 'Cerrar menú' }));
    expect(screen.getAllByRole('navigation', { name: 'Navegación principal' })).toHaveLength(1);
  });

  it('cierra la sesión desde el menú de usuario', async () => {
    signIn(makeUser('agente'));
    mockApi({ 'POST /auth/logout': { status: 204 } });
    const { router } = renderApp('/conversaciones');
    await userEvent.click(await screen.findByRole('button', { name: /Lucía Ramos/ }));
    await userEvent.click(screen.getByRole('menuitem', { name: 'Cerrar sesión' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/login');
    });
  });

  it('navega a cambiar contraseña desde el menú de usuario', async () => {
    signIn(makeUser('agente'));
    const { router } = renderApp('/conversaciones');
    await userEvent.click(await screen.findByRole('button', { name: /Lucía Ramos/ }));
    await userEvent.click(screen.getByRole('menuitem', { name: 'Cambiar contraseña' }));
    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/cuenta/password');
    });
    await userEvent.keyboard('{Escape}');
  });
});
