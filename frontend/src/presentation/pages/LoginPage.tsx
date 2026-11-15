import { zodResolver } from '@hookform/resolvers/zod';
import { CircleAlert, Eye, EyeOff, Lock, Mail } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useLogin } from '@/application/hooks/useAuth';
import { type LoginValues, loginSchema } from '@/domain/schemas';
import { ApiError } from '@/infrastructure/http/client';
import { BrandLogo } from '../components/BrandMark';
import { Button, IconButton } from '../components/ui/Button';
import { TextField } from '../components/ui/Field';

function loginErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 0 || error.status >= 500)
      return 'No pudimos conectar con el servidor. Inténtalo en un momento.';
    return error.message;
  }
  return 'Ocurrió un error inesperado';
}

/** Barras diagonales del isotipo como único recurso gráfico del panel de marca. */
function BrandStripes() {
  return (
    <svg
      viewBox="0 0 400 400"
      className="pointer-events-none absolute -right-24 -bottom-24 size-[420px]"
      aria-hidden
    >
      <g fill="none" stroke="currentColor" strokeWidth="28" className="text-white/10">
        <line x1="40" y1="400" x2="400" y2="40" />
        <line x1="120" y1="400" x2="400" y2="120" />
        <line x1="200" y1="400" x2="400" y2="200" />
      </g>
    </svg>
  );
}

export function LoginPage() {
  const login = useLogin();
  const [showPassword, setShowPassword] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const onSubmit = handleSubmit((values) => {
    login.mutate(values);
  });

  return (
    <div className="grid min-h-dvh lg:grid-cols-[minmax(0,5fr)_minmax(0,6fr)]">
      <section className="relative hidden overflow-hidden bg-brand-600 text-white lg:flex lg:flex-col lg:justify-between lg:p-12">
        <BrandLogo className="w-36" />
        <div className="relative max-w-md">
          <h1 className="font-display text-4xl leading-tight font-semibold text-white">
            Atención automatizada para tu empresa
          </h1>
          <p className="mt-4 text-base text-white/75">
            Responde por WhatsApp a cualquier hora y deriva a tu equipo solo lo que necesita una persona.
          </p>
        </div>
        <p className="relative text-xs text-white/60">Sertech · Nuevo Chimbote</p>
        <BrandStripes />
      </section>

      <section className="flex items-center justify-center px-4 py-12 sm:px-8">
        <div className="w-full max-w-sm">
          <BrandLogo className="mb-10 w-20 rounded-md lg:hidden" />
          <h2 className="text-2xl font-semibold tracking-tight">Iniciar sesión</h2>
          <p className="mt-1.5 text-sm text-ink-muted">Ingresa con la cuenta que te asignó tu empresa.</p>

          <form onSubmit={onSubmit} noValidate className="mt-8 flex flex-col gap-5">
            {login.isError && (
              <div
                role="alert"
                className="flex items-start gap-2.5 rounded-md border border-danger-600/25 bg-danger-50 px-3 py-2.5 text-sm text-danger-600"
              >
                <CircleAlert size={16} strokeWidth={1.75} className="mt-0.5 shrink-0" aria-hidden />
                {loginErrorMessage(login.error)}
              </div>
            )}

            <TextField
              label="Email"
              type="email"
              autoComplete="username"
              icon={Mail}
              placeholder="nombre@empresa.pe"
              error={errors.email?.message}
              {...register('email')}
            />
            <TextField
              label="Contraseña"
              type={showPassword ? 'text' : 'password'}
              autoComplete="current-password"
              icon={Lock}
              error={errors.password?.message}
              trailing={
                <IconButton
                  icon={showPassword ? EyeOff : Eye}
                  label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                  onClick={() => {
                    setShowPassword((v) => !v);
                  }}
                />
              }
              {...register('password')}
            />

            <Button type="submit" loading={login.isPending} className="mt-1 w-full">
              Ingresar
            </Button>
          </form>

          <p className="mt-8 text-xs text-ink-subtle">
            ¿Olvidaste tu contraseña? Pide al administrador de tu empresa que la restablezca.
          </p>
        </div>
      </section>
    </div>
  );
}
