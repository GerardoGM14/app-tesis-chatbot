import { z } from 'zod';
import { isValidRuc } from './ruc';

const email = z
  .string()
  .trim()
  .toLowerCase()
  .pipe(z.email({ message: 'Ingresa un email válido' }));
const personName = z
  .string()
  .trim()
  .min(2, 'Ingresa al menos 2 caracteres')
  .max(120, 'Máximo 120 caracteres');

export const loginSchema = z.object({
  email,
  password: z.string().min(1, 'Ingresa tu contraseña').max(128),
});
export type LoginValues = z.infer<typeof loginSchema>;

export const changePasswordSchema = z
  .object({
    actual: z.string().min(1, 'Ingresa tu contraseña actual'),
    nueva: z.string().min(10, 'Debe tener al menos 10 caracteres').max(128, 'Máximo 128 caracteres'),
    confirmar: z.string(),
  })
  .refine((v) => v.nueva === v.confirmar, { path: ['confirmar'], message: 'Las contraseñas no coinciden' })
  .refine((v) => v.nueva !== v.actual, { path: ['nueva'], message: 'Debe ser distinta a la actual' });
export type ChangePasswordValues = z.infer<typeof changePasswordSchema>;

export const tenantCreateSchema = z.object({
  nombre: personName,
  ruc: z.string().trim().refine(isValidRuc, 'Ingresa un RUC válido de 11 dígitos'),
  adminNombre: personName,
  adminEmail: email,
});
export type TenantCreateValues = z.infer<typeof tenantCreateSchema>;

const assignableRole = z.enum(['admin_empresa', 'agente'], { message: 'Elige un rol' });

export const userCreateSchema = z.object({
  nombre: personName,
  email,
  rol: assignableRole,
});
export type UserCreateValues = z.infer<typeof userCreateSchema>;

export const userEditSchema = z.object({
  nombre: personName,
  rol: assignableRole,
});
export type UserEditValues = z.infer<typeof userEditSchema>;
