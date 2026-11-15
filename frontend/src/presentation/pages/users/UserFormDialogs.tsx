import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useCreateUser, useUpdateUser } from '@/application/hooks/useUsers';
import type { CreatedUser, User } from '@/domain/entities';
import { ROLE_LABEL } from '@/domain/permissions';
import {
  type UserCreateValues,
  type UserEditValues,
  userCreateSchema,
  userEditSchema,
} from '@/domain/schemas';
import { Button } from '../../components/ui/Button';
import { Dialog } from '../../components/ui/Dialog';
import { SelectField, TextField } from '../../components/ui/Field';
import { applyApiError } from '../../forms';

const ROLE_OPTIONS = [
  { value: 'agente', label: ROLE_LABEL.agente },
  { value: 'admin_empresa', label: ROLE_LABEL.admin_empresa },
];

const CREATE_FORM_ID = 'form-nuevo-usuario';
const EDIT_FORM_ID = 'form-editar-usuario';

export function CreateUserDialog({
  open,
  onClose,
  onCreated,
}: Readonly<{
  open: boolean;
  onClose: () => void;
  onCreated: (created: CreatedUser) => void;
}>) {
  const create = useCreateUser();
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<UserCreateValues>({
    resolver: zodResolver(userCreateSchema),
    defaultValues: { nombre: '', email: '', rol: 'agente' },
  });

  const close = () => {
    reset();
    onClose();
  };

  const onSubmit = handleSubmit((values) => {
    create.mutate(values, {
      onSuccess: (created) => {
        reset();
        onCreated(created);
      },
      onError: (error) => {
        applyApiError(error, setError, ['nombre', 'email', 'rol']);
      },
    });
  });

  return (
    <Dialog
      open={open}
      onClose={close}
      title="Nuevo usuario"
      description="Recibirá una contraseña temporal que deberá cambiar al ingresar."
      footer={
        <>
          <Button variant="secondary" onClick={close} disabled={create.isPending}>
            Cancelar
          </Button>
          <Button type="submit" form={CREATE_FORM_ID} loading={create.isPending}>
            Crear usuario
          </Button>
        </>
      }
    >
      <form id={CREATE_FORM_ID} onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        <TextField label="Nombre completo" error={errors.nombre?.message} {...register('nombre')} />
        <TextField
          label="Email"
          type="email"
          autoComplete="off"
          error={errors.email?.message}
          {...register('email')}
        />
        <SelectField label="Rol" options={ROLE_OPTIONS} error={errors.rol?.message} {...register('rol')} />
      </form>
    </Dialog>
  );
}

export function EditUserDialog({ user, onClose }: Readonly<{ user: User | null; onClose: () => void }>) {
  const update = useUpdateUser();
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isDirty },
  } = useForm<UserEditValues>({ resolver: zodResolver(userEditSchema) });

  useEffect(() => {
    if (user && user.rol !== 'superadmin') reset({ nombre: user.nombre, rol: user.rol });
  }, [user, reset]);

  const onSubmit = handleSubmit((values) => {
    if (!user) return;
    update.mutate(
      { id: user.id, patch: values },
      {
        onSuccess: onClose,
        onError: (error) => {
          applyApiError(error, setError, ['nombre', 'rol']);
        },
      },
    );
  });

  return (
    <Dialog
      open={user !== null}
      onClose={onClose}
      title="Editar usuario"
      description={user?.email}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={update.isPending}>
            Cancelar
          </Button>
          <Button type="submit" form={EDIT_FORM_ID} loading={update.isPending} disabled={!isDirty}>
            Guardar cambios
          </Button>
        </>
      }
    >
      <form id={EDIT_FORM_ID} onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        <TextField label="Nombre completo" error={errors.nombre?.message} {...register('nombre')} />
        <SelectField label="Rol" options={ROLE_OPTIONS} error={errors.rol?.message} {...register('rol')} />
        <p className="text-xs text-ink-subtle">
          Si cambias el rol, se cerrarán las sesiones abiertas del usuario.
        </p>
      </form>
    </Dialog>
  );
}
