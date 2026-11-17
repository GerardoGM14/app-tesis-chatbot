import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import type { Tenant, User } from '@/domain/entities';
import { apiError, makeUser, mockApi, renderApp, signIn } from '@/test/utils';

const tenant: Tenant = {
  id: 't1',
  nombre: 'Ferretería El Progreso',
  ruc: '20600000005',
  slug: 'ferreteria-el-progreso',
  estado: 'activo',
  usuarios: 2,
  creadoEn: '2026-11-19T10:00:00Z',
};

const agent: User = {
  id: 'u2',
  tenantId: 'tenant-1',
  nombre: 'Diego Salazar',
  email: 'agente@progreso.demo',
  rol: 'agente',
  activo: true,
  debeCambiarPassword: false,
  ultimoAcceso: null,
  creadoEn: '2026-11-19T10:00:00Z',
};

const me: User = { ...agent, id: 'id-admin_empresa', nombre: 'Lucía Ramos', rol: 'admin_empresa' };

describe('Empresas', () => {
  it('lista, busca y muestra el estado de las empresas', async () => {
    signIn(makeUser('superadmin'));
    const calls = mockApi({
      'GET /empresas': (_, url) => ({
        body: { items: url.searchParams.get('q') === 'nada' ? [] : [tenant] },
      }),
    });
    renderApp('/empresas');
    expect(await screen.findByText('Ferretería El Progreso')).toBeInTheDocument();
    expect(screen.getByText('Activa')).toBeInTheDocument();

    await userEvent.type(screen.getByLabelText('Buscar'), 'nada');
    expect(await screen.findByText('Sin resultados')).toBeInTheDocument();
    expect(calls.some((c) => c.path === '/empresas')).toBe(true);
  });

  it('crea una empresa y muestra la contraseña temporal una sola vez', async () => {
    signIn(makeUser('superadmin'));
    const calls = mockApi({
      'GET /empresas': { body: { items: [] } },
      'POST /empresas': {
        status: 201,
        body: { empresa: tenant, adminEmail: 'admin@progreso.demo', passwordTemporal: 'Abc123Xyz456Qw' },
      },
    });
    renderApp('/empresas');
    await userEvent.click((await screen.findAllByRole('button', { name: 'Nueva empresa' }))[0]!);

    const dialog = screen.getByRole('dialog', { name: 'Nueva empresa' });
    await userEvent.type(within(dialog).getByLabelText('Razón social o nombre comercial'), tenant.nombre);
    await userEvent.type(within(dialog).getByLabelText('RUC'), tenant.ruc);
    await userEvent.type(within(dialog).getByLabelText('Nombre completo'), 'Lucía Ramos');
    await userEvent.type(within(dialog).getByLabelText('Email'), 'admin@progreso.demo');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear empresa' }));

    const credentials = await screen.findByRole('dialog', { name: 'Credenciales del administrador' });
    expect(within(credentials).getByText('Abc123Xyz456Qw')).toBeInTheDocument();
    expect(calls.find((c) => c.method === 'POST')?.body).toMatchObject({ ruc: tenant.ruc });
  });

  it('muestra en el campo el RUC duplicado que informa el servidor', async () => {
    signIn(makeUser('superadmin'));
    mockApi({
      'GET /empresas': { body: { items: [] } },
      'POST /empresas': apiError(422, 'validacion', 'Revisa los datos', { ruc: 'Ya registrado' }),
    });
    renderApp('/empresas');
    await userEvent.click((await screen.findAllByRole('button', { name: 'Nueva empresa' }))[0]!);
    const dialog = screen.getByRole('dialog', { name: 'Nueva empresa' });
    await userEvent.type(within(dialog).getByLabelText('Razón social o nombre comercial'), 'Otra');
    await userEvent.type(within(dialog).getByLabelText('RUC'), tenant.ruc);
    await userEvent.type(within(dialog).getByLabelText('Nombre completo'), 'Ana Torres');
    await userEvent.type(within(dialog).getByLabelText('Email'), 'ana@otra.pe');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear empresa' }));
    expect(await within(dialog).findByText('Ya registrado')).toBeInTheDocument();
  });

  it('suspende una empresa tras confirmar', async () => {
    signIn(makeUser('superadmin'));
    const calls = mockApi({
      'GET /empresas': { body: { items: [tenant] } },
      'POST /empresas/t1/suspender': { body: { ...tenant, estado: 'suspendido' } },
    });
    renderApp('/empresas');
    await userEvent.click(await screen.findByRole('button', { name: 'Suspender' }));
    const confirm = screen.getByRole('dialog', { name: 'Suspender empresa' });
    await userEvent.click(within(confirm).getByRole('button', { name: 'Suspender' }));
    await waitFor(() => {
      expect(calls.some((c) => c.path === '/empresas/t1/suspender')).toBe(true);
    });
  });

  it('restablece el acceso de un administrador', async () => {
    signIn(makeUser('superadmin'));
    mockApi({
      'GET /empresas': { body: { items: [tenant] } },
      'GET /empresas/t1/administradores': { body: { items: [me] } },
      'POST /usuarios/id-admin_empresa/restablecer': { body: { passwordTemporal: 'Temporal987Abc' } },
    });
    renderApp('/empresas');
    await userEvent.click(await screen.findByRole('button', { name: 'Acceso de administradores' }));
    const dialog = screen.getByRole('dialog', { name: 'Acceso de administradores' });
    await userEvent.click(await within(dialog).findByRole('button', { name: 'Restablecer' }));
    expect(await within(dialog).findByText('Temporal987Abc')).toBeInTheDocument();
  });

  it('ofrece reintentar si falla la carga', async () => {
    signIn(makeUser('superadmin'));
    mockApi({ 'GET /empresas': apiError(500, 'error_interno', 'Ocurrió un error inesperado') });
    renderApp('/empresas');
    expect(await screen.findByRole('button', { name: 'Reintentar' })).toBeInTheDocument();
  });
});

describe('Usuarios', () => {
  it('lista los usuarios y protege las acciones sobre uno mismo', async () => {
    signIn(makeUser('admin_empresa'));
    mockApi({ 'GET /usuarios': { body: { items: [me, { ...agent, debeCambiarPassword: true }] } } });
    renderApp('/usuarios');
    expect(await screen.findByText('Diego Salazar')).toBeInTheDocument();
    expect(screen.getByText('Contraseña temporal')).toBeInTheDocument();

    const myRow = screen.getByText('(tú)').closest('tr')!;
    expect(within(myRow).getByRole('button', { name: 'Desactivar' })).toBeDisabled();
    expect(within(myRow).getByRole('button', { name: 'Restablecer contraseña' })).toBeDisabled();
  });

  it('crea un usuario y muestra su contraseña temporal', async () => {
    signIn(makeUser('admin_empresa'));
    mockApi({
      'GET /usuarios': { body: { items: [me] } },
      'POST /usuarios': { status: 201, body: { usuario: agent, passwordTemporal: 'Nuevo123Abc456' } },
    });
    renderApp('/usuarios');
    await userEvent.click(await screen.findByRole('button', { name: 'Nuevo usuario' }));
    const dialog = screen.getByRole('dialog', { name: 'Nuevo usuario' });
    await userEvent.type(within(dialog).getByLabelText('Nombre completo'), agent.nombre);
    await userEvent.type(within(dialog).getByLabelText('Email'), agent.email);
    await userEvent.click(within(dialog).getByRole('button', { name: 'Crear usuario' }));
    expect(await screen.findByText('Nuevo123Abc456')).toBeInTheDocument();
  });

  it('edita el nombre y el rol de un usuario', async () => {
    signIn(makeUser('admin_empresa'));
    const calls = mockApi({
      'GET /usuarios': { body: { items: [me, agent] } },
      'PATCH /usuarios/u2': (body) => ({ body: { ...agent, ...(body as object) } }),
    });
    renderApp('/usuarios');
    const row = await screen.findByRole('row', { name: /Diego Salazar/ });
    await userEvent.click(within(row).getByRole('button', { name: 'Editar' }));
    const dialog = screen.getByRole('dialog', { name: 'Editar usuario' });
    await userEvent.selectOptions(within(dialog).getByLabelText('Rol'), 'admin_empresa');
    await userEvent.click(within(dialog).getByRole('button', { name: 'Guardar cambios' }));
    await waitFor(() => {
      expect(calls.find((c) => c.method === 'PATCH')?.body).toMatchObject({ rol: 'admin_empresa' });
    });
  });

  it('desactiva y restablece la contraseña de un agente', async () => {
    signIn(makeUser('admin_empresa'));
    const calls = mockApi({
      'GET /usuarios': { body: { items: [me, agent] } },
      'PATCH /usuarios/u2': { body: { ...agent, activo: false } },
      'POST /usuarios/u2/restablecer': { body: { passwordTemporal: 'Reset123Abc456' } },
    });
    renderApp('/usuarios');
    const row = await screen.findByRole('row', { name: /Diego Salazar/ });

    await userEvent.click(within(row).getByRole('button', { name: 'Desactivar' }));
    await userEvent.click(
      within(screen.getByRole('dialog', { name: 'Desactivar usuario' })).getByRole('button', {
        name: 'Desactivar',
      }),
    );
    await waitFor(() => {
      expect(calls.find((c) => c.method === 'PATCH')?.body).toEqual({ activo: false });
    });

    await userEvent.click(within(row).getByRole('button', { name: 'Restablecer contraseña' }));
    await userEvent.click(
      within(screen.getByRole('dialog', { name: 'Restablecer contraseña' })).getByRole('button', {
        name: 'Restablecer',
      }),
    );
    expect(await screen.findByText('Reset123Abc456')).toBeInTheDocument();
  });
});

describe('Dashboard', () => {
  it('muestra los indicadores de la tesis y el equipo del administrador', async () => {
    signIn(makeUser('admin_empresa'));
    mockApi({ 'GET /usuarios': { body: { items: [me, agent, { ...agent, id: 'u3', activo: false }] } } });
    renderApp('/dashboard');
    expect(await screen.findByText('Atendidas por el bot')).toBeInTheDocument();
    expect(await screen.findByText('3 en total')).toBeInTheDocument();
  });

  it('muestra los módulos pendientes con su estado vacío', async () => {
    signIn(makeUser('admin_empresa'));
    renderApp('/canal');
    expect(await screen.findByText('Módulo en desarrollo')).toBeInTheDocument();
  });
});
