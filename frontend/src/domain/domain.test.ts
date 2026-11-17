import { describe, expect, it } from 'vitest';
import { makeUser } from '@/test/utils';
import { can, homePathFor } from './permissions';
import { isValidRuc } from './ruc';
import { changePasswordSchema, loginSchema, tenantCreateSchema, userCreateSchema } from './schemas';

describe('isValidRuc', () => {
  it.each([
    ['20131312955', true],
    ['20600000005', true],
    ['10467793549', true],
    ['20131312954', false],
    ['30131312955', false],
    ['2013131295', false],
    ['2013131295A', false],
  ])('%s → %s', (ruc, expected) => {
    expect(isValidRuc(ruc)).toBe(expected);
  });
});

describe('permisos', () => {
  it('lleva a cada rol a su pantalla inicial', () => {
    expect(homePathFor('superadmin')).toBe('/empresas');
    expect(homePathFor('admin_empresa')).toBe('/dashboard');
    expect(homePathFor('agente')).toBe('/conversaciones');
  });

  it('consulta los permisos del usuario', () => {
    expect(can(makeUser('agente'), 'conversaciones:atender')).toBe(true);
    expect(can(makeUser('agente'), 'usuarios:gestionar')).toBe(false);
    expect(can(makeUser('superadmin'), 'conversaciones:atender')).toBe(false);
    expect(can(null, 'dashboard:ver')).toBe(false);
  });
});

describe('esquemas', () => {
  it('normaliza el email del login', () => {
    expect(loginSchema.parse({ email: '  ANA@Empresa.PE ', password: 'x' }).email).toBe('ana@empresa.pe');
    expect(loginSchema.safeParse({ email: 'ana', password: '' }).success).toBe(false);
  });

  it('exige confirmar y cambiar la contraseña', () => {
    const base = { actual: 'Temporal-123', nueva: 'Nueva-Clave-2026', confirmar: 'Nueva-Clave-2026' };
    expect(changePasswordSchema.safeParse(base).success).toBe(true);
    expect(changePasswordSchema.safeParse({ ...base, confirmar: 'otra' }).success).toBe(false);
    expect(changePasswordSchema.safeParse({ ...base, nueva: 'corta', confirmar: 'corta' }).success).toBe(
      false,
    );
    expect(
      changePasswordSchema.safeParse({
        actual: 'Igual-Clave-2026',
        nueva: 'Igual-Clave-2026',
        confirmar: 'Igual-Clave-2026',
      }).success,
    ).toBe(false);
  });

  it('valida el alta de empresa con RUC', () => {
    const valid = { nombre: 'Mi Empresa', ruc: '20600000005', adminNombre: 'Ana', adminEmail: 'ana@x.pe' };
    expect(tenantCreateSchema.safeParse(valid).success).toBe(true);
    expect(tenantCreateSchema.safeParse({ ...valid, ruc: '12345678901' }).success).toBe(false);
  });

  it('no permite crear superadmins desde el panel', () => {
    expect(userCreateSchema.safeParse({ nombre: 'Ana', email: 'a@b.pe', rol: 'superadmin' }).success).toBe(
      false,
    );
    expect(userCreateSchema.safeParse({ nombre: 'Ana', email: 'a@b.pe', rol: 'agente' }).success).toBe(true);
  });
});
