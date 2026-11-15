import { KeyRound } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { useTenantAdmins } from '@/application/hooks/useTenants';
import { useResetPassword } from '@/application/hooks/useUsers';
import type { Tenant, User } from '@/domain/entities';
import { Skeleton } from '../../components/skeletons/Skeletons';
import { Button } from '../../components/ui/Button';
import { Dialog } from '../../components/ui/Dialog';
import { ErrorState } from '../../components/ui/Layout';
import { SecretReveal } from '../../components/ui/SecretReveal';
import { errorMessage } from '../../forms';

interface AdminsDialogProps {
  tenant: Tenant | null;
  onClose: () => void;
}

/** El superadmin restablece el acceso de los administradores de una empresa. */
export function AdminsDialog({ tenant, onClose }: Readonly<AdminsDialogProps>) {
  const admins = useTenantAdmins(tenant?.id ?? null);
  const reset = useResetPassword();
  const [result, setResult] = useState<{ email: string; password: string } | null>(null);

  const close = () => {
    setResult(null);
    reset.reset();
    onClose();
  };

  const resetAdmin = (admin: User) => {
    reset.mutate(admin.id, {
      onSuccess: ({ passwordTemporal }) => {
        setResult({ email: admin.email, password: passwordTemporal });
      },
      onError: (error) => toast.error(errorMessage(error)),
    });
  };

  return (
    <Dialog
      open={tenant !== null}
      onClose={close}
      title="Acceso de administradores"
      description={tenant?.nombre}
      footer={
        <Button variant="secondary" onClick={close}>
          {result ? 'Listo' : 'Cerrar'}
        </Button>
      }
    >
      {result && <SecretReveal label="Contraseña temporal" value={result.password} email={result.email} />}
      {!result && admins.isPending && (
        <div className="flex flex-col gap-3" role="status" aria-label="Cargando administradores">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-12 w-full" />
        </div>
      )}
      {!result && admins.isError && (
        <ErrorState message={errorMessage(admins.error)} onRetry={() => void admins.refetch()} />
      )}
      {!result && admins.data && (
        <ul className="flex flex-col divide-y divide-line rounded-md border border-line">
          {admins.data.map((admin) => (
            <li key={admin.id} className="flex items-center justify-between gap-3 px-3 py-2.5">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{admin.nombre}</p>
                <p className="truncate text-xs text-ink-subtle">{admin.email}</p>
              </div>
              <Button
                size="sm"
                variant="secondary"
                icon={KeyRound}
                loading={reset.isPending && reset.variables === admin.id}
                onClick={() => {
                  resetAdmin(admin);
                }}
              >
                Restablecer
              </Button>
            </li>
          ))}
        </ul>
      )}
    </Dialog>
  );
}
