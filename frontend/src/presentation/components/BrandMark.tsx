import { clsx } from 'clsx';
import logo from '@/assets/brand/sertech-logo.png';

/**
 * Isotipo de Sertech recortado del logo cuadrado. El fondo del archivo es el
 * mismo azul de marca, así que el recorte se ve como una tesela plana.
 */
export function BrandMark({ className }: Readonly<{ className?: string }>) {
  return (
    <span
      role="img"
      aria-label="Sertech"
      className={clsx('block shrink-0 rounded-md bg-brand-600 bg-no-repeat', className)}
      style={{ backgroundImage: `url(${logo})`, backgroundSize: '210%', backgroundPosition: '50% 30%' }}
    />
  );
}

/** Logo completo (isotipo + nombre) sobre fondo azul de marca. */
export function BrandLogo({ className }: Readonly<{ className?: string }>) {
  return <img src={logo} alt="Sertech" className={clsx('block', className)} />;
}
