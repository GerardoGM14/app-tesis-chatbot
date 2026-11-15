import { Clock, Gauge, type LucideIcon, MessageCircleCheck, Smile, Users } from 'lucide-react';
import { useCurrentUser } from '@/application/hooks/useAuth';
import { useUsers } from '@/application/hooks/useUsers';
import { can } from '@/domain/permissions';
import { KpiSkeleton } from '../components/skeletons/Skeletons';
import { Card, PageHeader } from '../components/ui/Layout';

interface Kpi {
  label: string;
  value: string;
  caption: string;
  icon: LucideIcon;
}

function KpiCard({ label, value, caption, icon: Icon }: Readonly<Kpi>) {
  return (
    <Card className="p-5">
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium text-ink-muted">{label}</p>
        <Icon size={16} strokeWidth={1.75} className="text-ink-subtle" aria-hidden />
      </div>
      <p className="mt-3 font-display text-3xl font-semibold tracking-tight tabular-nums">{value}</p>
      <p className="mt-1.5 text-xs text-ink-subtle">{caption}</p>
    </Card>
  );
}

/** Indicadores de la tesis; se alimentan cuando el canal de WhatsApp esté activo. */
const THESIS_KPIS: Kpi[] = [
  { label: 'Atendidas por el bot', value: '—', caption: 'Meta: 80 % o más', icon: MessageCircleCheck },
  { label: 'Tiempo promedio de respuesta', value: '—', caption: 'Meta: 5 s o menos', icon: Clock },
  { label: 'Disponibilidad del canal', value: '—', caption: 'Meta: 99 % o más', icon: Gauge },
  { label: 'Satisfacción', value: '—', caption: 'Meta: 80 % o más', icon: Smile },
];

export function DashboardPage() {
  const user = useCurrentUser();
  const canSeeUsers = can(user, 'usuarios:gestionar');
  const users = useUsers(canSeeUsers);

  const firstName = user?.nombre.split(' ')[0] ?? '';
  const activeUsers = users.data?.filter((u) => u.activo).length ?? 0;

  return (
    <>
      <PageHeader
        title={`Hola, ${firstName}`}
        description={user?.tenantNombre ? `Resumen de ${user.tenantNombre}` : 'Resumen de la plataforma'}
      />

      <h2 className="mb-3 text-sm font-semibold">Indicadores de atención</h2>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {THESIS_KPIS.map((kpi) => (
          <KpiCard key={kpi.label} {...kpi} />
        ))}
      </div>
      <p className="mt-3 text-xs text-ink-subtle">
        Los indicadores se calculan a partir de las conversaciones de WhatsApp y empiezan a mostrarse al
        vincular el canal.
      </p>

      {canSeeUsers && (
        <>
          <h2 className="mt-10 mb-3 text-sm font-semibold">Equipo</h2>
          {users.isPending ? (
            <KpiSkeleton count={2} />
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
              <KpiCard
                label="Usuarios activos"
                value={String(activeUsers)}
                caption={`${users.data?.length ?? 0} en total`}
                icon={Users}
              />
            </div>
          )}
        </>
      )}
    </>
  );
}
