import { zodResolver } from '@hookform/resolvers/zod';
import { KeyRound, LogOut, ShieldCheck } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useChangePassword, useCurrentUser, useLogout } from '@/application/hooks/useAuth';
import { type ChangePasswordValues, changePasswordSchema } from '@/domain/schemas';
import { BrandMark } from '../components/BrandMark';
import { Button } from '../components/ui/Button';
import { TextField } from '../components/ui/Field';
import { Card, PageHeader } from '../components/ui/Layout';
import { applyApiError } from '../forms';

const FIELDS = ['actual', 'nueva', 'confirmar'] as const;

function ChangePasswordForm({ submitLabel }: Readonly<{ submitLabel: string }>) {
  const change = useChangePassword();
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<ChangePasswordValues>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { actual: '', nueva: '', confirmar: '' },
  });

  const onSubmit = handleSubmit((values) => {
    change.mutate(values, {
      onSuccess: () => {
        reset();
        toast.success('Contraseña actualizada', { description: 'Se cerraron tus otras sesiones.' });
      },
      onError: (error) => {
        applyApiError(error, setError, FIELDS);
      },
    });
  });

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-5">
      <TextField
        label="Contraseña actual"
        type="password"
        autoComplete="current-password"
        error={errors.actual?.message}
        {...register('actual')}
      />
      <TextField
        label="Nueva contraseña"
        type="password"
        autoComplete="new-password"
        hint="Mínimo 10 caracteres. Una frase es más fácil de recordar."
        error={errors.nueva?.message}
        {...register('nueva')}
      />
      <TextField
        label="Repite la nueva contraseña"
        type="password"
        autoComplete="new-password"
        error={errors.confirmar?.message}
        {...register('confirmar')}
      />
      <Button type="submit" icon={KeyRound} loading={change.isPending}>
        {submitLabel}
      </Button>
    </form>
  );
}

/** Cambio voluntario, dentro del panel. */
export function ChangePasswordPage() {
  return (
    <>
      <PageHeader
        title="Cambiar contraseña"
        description="Al cambiarla se cerrarán tus sesiones en otros dispositivos."
      />
      <Card className="max-w-lg p-6">
        <ChangePasswordForm submitLabel="Guardar contraseña" />
      </Card>
    </>
  );
}

/** Cambio obligatorio en el primer ingreso o tras un restablecimiento. */
export function ForcedChangePasswordPage() {
  const user = useCurrentUser();
  const logout = useLogout();

  return (
    <div className="flex min-h-dvh items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="mb-8 flex items-center gap-3">
          <BrandMark className="size-9" />
          <span className="font-display text-base font-semibold">Sertech</span>
        </div>
        <Card className="p-6 sm:p-8">
          <div className="mb-6 flex items-start gap-3">
            <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-brand-50 text-brand-600">
              <ShieldCheck size={18} strokeWidth={1.75} aria-hidden />
            </div>
            <div>
              <h1 className="text-xl font-semibold">Crea tu contraseña</h1>
              <p className="mt-1 text-sm text-ink-muted">
                Hola{user ? `, ${user.nombre.split(' ')[0] ?? ''}` : ''}. Por seguridad, reemplaza la
                contraseña temporal antes de continuar.
              </p>
            </div>
          </div>
          <ChangePasswordForm submitLabel="Guardar y continuar" />
        </Card>
        <Button
          variant="ghost"
          icon={LogOut}
          className="mt-4"
          onClick={() => {
            logout.mutate();
          }}
        >
          Cerrar sesión
        </Button>
      </div>
    </div>
  );
}
