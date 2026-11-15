import { Building2, KeyRound, Plus, Power, PowerOff, Search } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { useSetTenantStatus, useTenants } from '@/application/hooks/useTenants';
import type { CreatedTenant, Tenant } from '@/domain/entities';
import { TableSkeleton } from '../components/skeletons/Skeletons';
import { Badge } from '../components/ui/Badge';
import { Button, IconButton } from '../components/ui/Button';
import { ConfirmDialog } from '../components/ui/ConfirmDialog';
import { Dialog } from '../components/ui/Dialog';
import { TextField } from '../components/ui/Field';
import { Card, EmptyState, ErrorState, PageHeader, Table, Td, Th } from '../components/ui/Layout';
import { SecretReveal } from '../components/ui/SecretReveal';
import { errorMessage } from '../forms';
import { formatDate, useDebouncedValue } from '../format';
import { AdminsDialog } from './tenants/AdminsDialog';
import { CreateTenantDialog } from './tenants/CreateTenantDialog';

export function TenantsPage() {
  const [search, setSearch] = useState('');
  const debounced = useDebouncedValue(search.trim());
  const tenants = useTenants(debounced);
  const setStatus = useSetTenantStatus();

  const [createOpen, setCreateOpen] = useState(false);
  const [created, setCreated] = useState<CreatedTenant | null>(null);
  const [statusTarget, setStatusTarget] = useState<Tenant | null>(null);
  const [adminsTarget, setAdminsTarget] = useState<Tenant | null>(null);

  const suspending = statusTarget?.estado === 'activo';
  const targetName = statusTarget?.nombre ?? '';

  const confirmStatus = () => {
    if (!statusTarget) return;
    setStatus.mutate(
      { id: statusTarget.id, suspend: suspending },
      {
        onSuccess: (tenant) => {
          toast.success(tenant.estado === 'suspendido' ? 'Empresa suspendida' : 'Empresa reactivada', {
            description: tenant.nombre,
          });
          setStatusTarget(null);
        },
        onError: (error) => toast.error(errorMessage(error)),
      },
    );
  };

  const newButton = (
    <Button
      icon={Plus}
      onClick={() => {
        setCreateOpen(true);
      }}
    >
      Nueva empresa
    </Button>
  );

  return (
    <>
      <PageHeader
        title="Empresas"
        description="Empresas que usan la plataforma y su estado de acceso."
        actions={newButton}
      />

      <Card>
        <div className="border-b border-line p-4">
          <div className="max-w-sm">
            <TextField
              label="Buscar"
              icon={Search}
              placeholder="Nombre o RUC"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
              }}
            />
          </div>
        </div>

        {tenants.isPending && <TableSkeleton columns={5} />}
        {tenants.isError && (
          <ErrorState message={errorMessage(tenants.error)} onRetry={() => void tenants.refetch()} />
        )}
        {tenants.data?.length === 0 && (
          <EmptyState
            icon={Building2}
            title={debounced ? 'Sin resultados' : 'Aún no hay empresas'}
            description={
              debounced
                ? 'Prueba con otro nombre o RUC.'
                : 'Registra la primera empresa para que empiece a usar el bot.'
            }
            action={debounced ? undefined : newButton}
          />
        )}
        {tenants.data && tenants.data.length > 0 && (
          <Table label="Empresas">
            <thead>
              <tr>
                <Th>Empresa</Th>
                <Th>RUC</Th>
                <Th className="text-right">Usuarios</Th>
                <Th>Estado</Th>
                <Th>Alta</Th>
                <Th className="w-24 text-right">
                  <span className="sr-only">Acciones</span>
                </Th>
              </tr>
            </thead>
            <tbody>
              {tenants.data.map((tenant) => (
                <tr key={tenant.id} className="hover:bg-canvas/60">
                  <Td>
                    <p className="font-medium">{tenant.nombre}</p>
                    <p className="text-xs text-ink-subtle">{tenant.slug}</p>
                  </Td>
                  <Td className="font-mono text-xs text-ink-muted">{tenant.ruc}</Td>
                  <Td className="text-right tabular-nums">{tenant.usuarios}</Td>
                  <Td>
                    {tenant.estado === 'activo' ? (
                      <Badge tone="success">Activa</Badge>
                    ) : (
                      <Badge tone="danger">Suspendida</Badge>
                    )}
                  </Td>
                  <Td className="text-ink-muted">{formatDate(tenant.creadoEn)}</Td>
                  <Td>
                    <div className="flex justify-end gap-1">
                      <IconButton
                        icon={KeyRound}
                        label="Acceso de administradores"
                        onClick={() => {
                          setAdminsTarget(tenant);
                        }}
                      />
                      <IconButton
                        icon={tenant.estado === 'activo' ? PowerOff : Power}
                        label={tenant.estado === 'activo' ? 'Suspender' : 'Reactivar'}
                        tone={tenant.estado === 'activo' ? 'danger' : 'default'}
                        onClick={() => {
                          setStatusTarget(tenant);
                        }}
                      />
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </Card>

      <CreateTenantDialog
        open={createOpen}
        onClose={() => {
          setCreateOpen(false);
        }}
        onCreated={(result) => {
          setCreateOpen(false);
          setCreated(result);
          toast.success('Empresa creada', { description: result.empresa.nombre });
        }}
      />

      <Dialog
        open={created !== null}
        onClose={() => {
          setCreated(null);
        }}
        title="Credenciales del administrador"
        description={created?.empresa.nombre}
        footer={
          <Button
            onClick={() => {
              setCreated(null);
            }}
          >
            Listo
          </Button>
        }
      >
        {created && (
          <SecretReveal
            label="Contraseña temporal"
            value={created.passwordTemporal}
            email={created.adminEmail}
          />
        )}
      </Dialog>

      <ConfirmDialog
        open={statusTarget !== null}
        title={suspending ? 'Suspender empresa' : 'Reactivar empresa'}
        description={
          suspending
            ? `${targetName} perderá el acceso de inmediato y se cerrarán las sesiones de sus usuarios.`
            : `${targetName} podrá volver a ingresar a la plataforma.`
        }
        confirmLabel={suspending ? 'Suspender' : 'Reactivar'}
        tone={suspending ? 'danger' : 'primary'}
        loading={setStatus.isPending}
        onConfirm={confirmStatus}
        onClose={() => {
          setStatusTarget(null);
        }}
      />

      <AdminsDialog
        tenant={adminsTarget}
        onClose={() => {
          setAdminsTarget(null);
        }}
      />
    </>
  );
}
