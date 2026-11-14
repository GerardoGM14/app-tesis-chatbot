import { Check, Copy, TriangleAlert } from 'lucide-react';
import { useState } from 'react';
import { IconButton } from './Button';

interface SecretRevealProps {
  label: string;
  value: string;
  email?: string;
}

/** Muestra una contraseña temporal que el servidor entrega una sola vez. */
export function SecretReveal({ label, value, email }: Readonly<SecretRevealProps>) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => {
        setCopied(false);
      }, 2000);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="flex flex-col gap-3">
      {email && (
        <p className="text-sm text-ink-muted">
          Usuario: <span className="font-medium text-ink">{email}</span>
        </p>
      )}
      <div className="flex flex-col gap-1.5">
        <span className="text-xs font-medium text-ink">{label}</span>
        <div className="flex items-center justify-between gap-2 rounded-md border border-line-strong bg-canvas py-1 pr-1 pl-3">
          <code className="font-mono text-sm tracking-wide text-ink select-all">{value}</code>
          <IconButton icon={copied ? Check : Copy} label={copied ? 'Copiada' : 'Copiar'} onClick={copy} />
        </div>
      </div>
      <p className="flex gap-2 rounded-md border border-accent-500/25 bg-accent-50 px-3 py-2 text-xs text-accent-600">
        <TriangleAlert size={14} strokeWidth={1.75} className="mt-px shrink-0" aria-hidden />
        Solo se muestra esta vez. Entrégala por un canal interno; se pedirá cambiarla en el primer ingreso.
      </p>
    </div>
  );
}
