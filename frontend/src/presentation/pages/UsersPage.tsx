import { KeyRound, Pencil, Plus, Power, PowerOff, Users } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { useCurrentUser } from '@/application/hooks/useAuth';
import { useResetPassword, useUpdateUser, useUsers } from '@/application/hooks/useUsers';
import type { User } from '@/domain/entities';
import { ROLE_LABEL } from '@/domain/permissions';
import { TableSkeleton } from '../components/skeletons/Skeletons';
import { Badge } from '../components/ui/Badge';
import { Button, IconButton } from '../components/ui/Button';
import { ConfirmDialog } from '../components/ui/ConfirmDialog';
import { Dialog } from '../components/ui/Dialog';
import { Card, EmptyState, ErrorState, PageHeader, Table, Td, Th } from '../components/ui/Layout';
import { SecretReveal } from '../components/ui/SecretReveal';
import { errorMessage } from '../forms';
import { formatDateTime } from '../format';
import { CreateUserDialog, EditUserDialog } from './users/UserFormDialogs';

type PendingAction = { kind: 'toggle'; user: User } | { kind: 'reset'; user: User };

interface ConfirmCopy {
  title: string;
  description: string;
  confirmLabel: string;
  tone: 'primary' | 'danger';
}

function confirmCopy(action: PendingAction | null): ConfirmCopy {
  if (!action) return { title: '', description: '', confirmLabel: '', tone: 'primary' };
  const { user } = action;
  if (action.kind === 'reset') {
    return {
      title: 'Restablecer contraseña',
      description: `Se generará una contraseña temporal para ${user.nombre} y se cerrarán sus sesiones.`,
      confirmLabel: 'Restablecer',
      tone: 'primary',
    };
  }
  if (user.activo) {
    return {
      title: 'Desactivar usuario',
      description: `${user.nombre} no podrá ingresar y se cerrarán sus sesiones.`,
      confirmLabel: 'Desactivar',
      tone: 'danger',
    };
  }
  return {
    title: 'Activar usuario',
    description: `${user.nombre} podrá volver a ingresar.`,
    confirmLabel: 'Activar',
    tone: 'primary',
  };
}

function StatusBadges({ user }: Readonly<{ user: User }>) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {user.activo ? <Badge tone="success">Activo</Badge> : <Badge>Inactivo</Badge>}
      {user.activo && user.debeCambiarPassword && <Badge tone="accent">Contraseña temporal</Badge>}
    </div>
  );
}

export function UsersPage() {
  const me = useCurrentUser();
  const users = useUsers();
  const update = useUpdateUser();
  const resetPassword = useResetPassword();

  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<User | null>(null);
  const [pending, setPending] = useState<PendingAction | null>(null);
  const [credentials, setCredentials] = useState<{ email: string; password: string } | null>(null);

  const confirmPending = () => {
    if (!pending) return;
    const { user } = pending;
    if (pending.kind === 'toggle') {
      update.mutate(
        { id: user.id, patch: { activo: !user.activo } },
        {
          onSuccess: (u) => {
            toast.success(u.activo ? 'Usuario activado' : 'Usuario desactivado', { description: u.nombre });
            setPending(null);
          },
          onError: (error) => toast.error(errorMessage(error)),
        },
      );
      return;
    }
    resetPassword.mutate(user.id, {
      onSuccess: ({ passwordTemporal }) => {
        setPending(null);
        setCredentials({ email: user.email, password: passwordTemporal });
      },
      onError: (error) => toast.error(errorMessage(error)),
    });
  };

  const newButton = (
    <Button
      icon={Plus}
      onClick={() => {
        setCreateOpen(true);
      }}
    >
      Nuevo usuario
    </Button>
  );

  return (
    <>
      <PageHeader
        title="Usuarios"
        description="Personas de tu empresa con acceso al panel."
        actions={newButton}
      />

      <Card>
        {users.isPending && <TableSkeleton columns={5} />}
        {users.isError && (
          <ErrorState message={errorMessage(users.error)} onRetry={() => void users.refetch()} />
        )}
        {users.data?.length === 0 && (
          <EmptyState
            icon={Users}
            title="Aún no hay usuarios"
            description="Agrega a los agentes que atenderán las conversaciones."
            action={newButton}
          />
        )}
        {users.data && users.data.length > 0 && (
          <Table label="Usuarios">
            <thead>
              <tr>
                <Th>Usuario</Th>
                <Th>Rol</Th>
                <Th>Estado</Th>
                <Th>Último acceso</Th>
                <Th className="w-28 text-right">
                  <span className="sr-only">Acciones</span>
                </Th>
              </tr>
            </thead>
            <tbody>
              {users.data.map((user) => {
                const isMe = user.id === me?.id;
                return (
                  <tr key={user.id} className="hover:bg-canvas/60">
                    <Td>
                      <p className="font-medium">
                        {user.nombre}
                        {isMe && <span className="ml-1.5 text-xs font-normal text-ink-subtle">(tú)</span>}
                      </p>
                      <p className="text-xs text-ink-subtle">{user.email}</p>
                    </Td>
                    <Td>
                      <Badge tone={user.rol === 'admin_empresa' ? 'brand' : 'neutral'}>
                        {ROLE_LABEL[user.rol]}
                      </Badge>
                    </Td>
                    <Td>
                      <StatusBadges user={user} />
                    </Td>
                    <Td className="text-ink-muted">{formatDateTime(user.ultimoAcceso)}</Td>
                    <Td>
                      <div className="flex justify-end gap-1">
                        <IconButton
                          icon={Pencil}
                          label="Editar"
                          onClick={() => {
                            setEditing(user);
                          }}
                        />
                        <IconButton
                          icon={KeyRound}
                          label="Restablecer contraseña"
                          disabled={isMe || user.rol !== 'agente'}
                          onClick={() => {
                            setPending({ kind: 'reset', user });
                          }}
                        />
                        <IconButton
                          icon={user.activo ? PowerOff : Power}
                          label={user.activo ? 'Desactivar' : 'Activar'}
                          tone={user.activo ? 'danger' : 'default'}
                          disabled={isMe}
                          onClick={() => {
                            setPending({ kind: 'toggle', user });
                          }}
                        />
                      </div>
                    </Td>
                  </tr>
                );
              })}
            </tbody>
          </Table>
        )}
      </Card>

      <CreateUserDialog
        open={createOpen}
        onClose={() => {
          setCreateOpen(false);
        }}
        onCreated={(created) => {
          setCreateOpen(false);
          setCredentials({ email: created.usuario.email, password: created.passwordTemporal });
          toast.success('Usuario creado', { description: created.usuario.nombre });
        }}
      />

      <EditUserDialog
        user={editing}
        onClose={() => {
          setEditing(null);
        }}
      />

      <ConfirmDialog
        open={pending !== null}
        {...confirmCopy(pending)}
        loading={update.isPending || resetPassword.isPending}
        onConfirm={confirmPending}
        onClose={() => {
          setPending(null);
        }}
      />

      <Dialog
        open={credentials !== null}
        onClose={() => {
          setCredentials(null);
        }}
        title="Contraseña temporal"
        footer={
          <Button
            onClick={() => {
              setCredentials(null);
            }}
          >
            Listo
          </Button>
        }
      >
        {credentials && (
          <SecretReveal label="Contraseña temporal" value={credentials.password} email={credentials.email} />
        )}
      </Dialog>
    </>
  );
}
