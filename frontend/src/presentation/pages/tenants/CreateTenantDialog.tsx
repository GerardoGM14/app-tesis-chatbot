import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { useCreateTenant } from '@/application/hooks/useTenants';
import type { CreatedTenant } from '@/domain/entities';
import { type TenantCreateValues, tenantCreateSchema } from '@/domain/schemas';
import { Button } from '../../components/ui/Button';
import { Dialog } from '../../components/ui/Dialog';
import { TextField } from '../../components/ui/Field';
import { applyApiError } from '../../forms';

const FIELDS = ['nombre', 'ruc', 'adminNombre', 'adminEmail'] as const;
const FORM_ID = 'form-nueva-empresa';

interface CreateTenantDialogProps {
  open: boolean;
  onClose: () => void;
  onCreated: (created: CreatedTenant) => void;
}

export function CreateTenantDialog({ open, onClose, onCreated }: Readonly<CreateTenantDialogProps>) {
  const create = useCreateTenant();
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<TenantCreateValues>({
    resolver: zodResolver(tenantCreateSchema),
    defaultValues: { nombre: '', ruc: '', adminNombre: '', adminEmail: '' },
  });

  const close = () => {
    reset();
    create.reset();
    onClose();
  };

  const onSubmit = handleSubmit((values) => {
    create.mutate(values, {
      onSuccess: (created) => {
        reset();
        onCreated(created);
      },
      onError: (error) => {
        applyApiError(error, setError, FIELDS);
      },
    });
  });

  return (
    <Dialog
      open={open}
      onClose={close}
      title="Nueva empresa"
      description="Se creará la empresa con su primer administrador."
      footer={
        <>
          <Button variant="secondary" onClick={close} disabled={create.isPending}>
            Cancelar
          </Button>
          <Button type="submit" form={FORM_ID} loading={create.isPending}>
            Crear empresa
          </Button>
        </>
      }
    >
      <form id={FORM_ID} onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
        <TextField
          label="Razón social o nombre comercial"
          error={errors.nombre?.message}
          {...register('nombre')}
        />
        <TextField
          label="RUC"
          inputMode="numeric"
          maxLength={11}
          error={errors.ruc?.message}
          {...register('ruc')}
        />
        <div className="mt-1 border-t border-line pt-4">
          <p className="mb-3 text-xs font-medium tracking-wide text-ink-muted uppercase">Administrador</p>
          <div className="flex flex-col gap-4">
            <TextField
              label="Nombre completo"
              error={errors.adminNombre?.message}
              {...register('adminNombre')}
            />
            <TextField
              label="Email"
              type="email"
              autoComplete="off"
              error={errors.adminEmail?.message}
              {...register('adminEmail')}
            />
          </div>
        </div>
      </form>
    </Dialog>
  );
}
