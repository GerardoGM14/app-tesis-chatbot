import { Compass, ShieldOff } from 'lucide-react';
import { Link } from 'react-router';
import { Card, EmptyState } from '../components/ui/Layout';

const backLink = (
  <Link to="/" className="text-sm font-medium text-brand-600 hover:underline">
    Ir al inicio
  </Link>
);

export function ForbiddenPage() {
  return (
    <Card>
      <EmptyState
        icon={ShieldOff}
        title="No tienes acceso a esta sección"
        description="Tu rol no incluye este módulo. Si crees que es un error, consulta con el administrador de tu empresa."
        action={backLink}
      />
    </Card>
  );
}

export function NotFoundPage() {
  return (
    <Card>
      <EmptyState
        icon={Compass}
        title="Página no encontrada"
        description="La dirección no existe o fue movida."
        action={backLink}
      />
    </Card>
  );
}
