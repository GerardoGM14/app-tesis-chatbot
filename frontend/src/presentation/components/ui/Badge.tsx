import { clsx } from 'clsx';
import type { ReactNode } from 'react';

type Tone = 'neutral' | 'brand' | 'success' | 'danger' | 'accent';

const tones: Record<Tone, string> = {
  neutral: 'bg-canvas text-ink-muted border-line',
  brand: 'bg-brand-50 text-brand-700 border-brand-100',
  success: 'bg-success-50 text-success-600 border-success-600/20',
  danger: 'bg-danger-50 text-danger-600 border-danger-600/20',
  accent: 'bg-accent-50 text-accent-600 border-accent-500/25',
};

export function Badge({ tone = 'neutral', children }: Readonly<{ tone?: Tone; children: ReactNode }>) {
  return (
    <span
      className={clsx(
        'inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap',
        tones[tone],
      )}
    >
      {children}
    </span>
  );
}
