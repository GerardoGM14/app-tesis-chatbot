import { Bot, type LucideIcon, MessagesSquare, QrCode } from 'lucide-react';
import { Card, EmptyState, PageHeader } from '../components/ui/Layout';

type Module = 'conversaciones' | 'bot' | 'canal';

const MODULES: Record<Module, { title: string; description: string; icon: LucideIcon; pending: string }> = {
  conversaciones: {
    title: 'Conversaciones',
    description: 'Bandeja de conversaciones de WhatsApp y atención de derivaciones.',
    icon: MessagesSquare,
    pending: 'La bandeja se habilita cuando la empresa vincule su número de WhatsApp.',
  },
  bot: {
    title: 'Bot',
    description: 'Intenciones, preguntas frecuentes, menú y horario de atención.',
    icon: Bot,
    pending: 'La configuración del bot se habilita junto con el canal de WhatsApp.',
  },
  canal: {
    title: 'Canal de WhatsApp',
    description: 'Vinculación del número de la empresa mediante código QR.',
    icon: QrCode,
    pending: 'La vinculación por QR estará disponible en la próxima versión.',
  },
};

/** Módulos planificados para los siguientes sprints. */
export function ModulePendingPage({ module }: Readonly<{ module: Module }>) {
  const info = MODULES[module];
  return (
    <>
      <PageHeader title={info.title} description={info.description} />
      <Card>
        <EmptyState icon={info.icon} title="Módulo en desarrollo" description={info.pending} />
      </Card>
    </>
  );
}
