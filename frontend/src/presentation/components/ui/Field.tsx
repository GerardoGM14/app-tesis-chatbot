import { clsx } from 'clsx';
import type { LucideIcon } from 'lucide-react';
import { type InputHTMLAttributes, type ReactNode, type Ref, type SelectHTMLAttributes, useId } from 'react';

interface FieldShellProps {
  id: string;
  label: string;
  error?: string;
  hint?: string;
  children: ReactNode;
}

function FieldShell({ id, label, error, hint, children }: Readonly<FieldShellProps>) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-xs font-medium text-ink">
        {label}
      </label>
      {children}
      {error ? (
        <p id={`${id}-error`} className="text-xs text-danger-600" role="alert">
          {error}
        </p>
      ) : (
        hint && (
          <p id={`${id}-hint`} className="text-xs text-ink-subtle">
            {hint}
          </p>
        )
      )}
    </div>
  );
}

const controlClass = (invalid: boolean, withIcon: boolean) =>
  clsx(
    'h-10 w-full rounded-md border bg-surface px-3 text-sm text-ink transition-colors placeholder:text-ink-subtle',
    'focus:border-brand-600 focus:outline-none focus:ring-1 focus:ring-brand-600',
    'disabled:bg-canvas disabled:text-ink-subtle',
    invalid ? 'border-danger-600' : 'border-line-strong',
    withIcon && 'pl-9',
  );

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
  icon?: LucideIcon;
  trailing?: ReactNode;
  ref?: Ref<HTMLInputElement>;
}

export function TextField({
  label,
  error,
  hint,
  icon: Icon,
  trailing,
  id,
  className,
  ref,
  ...props
}: Readonly<TextFieldProps>) {
  const generated = useId();
  const inputId = id ?? generated;
  let describedBy: string | undefined;
  if (error) describedBy = `${inputId}-error`;
  else if (hint) describedBy = `${inputId}-hint`;
  return (
    <FieldShell id={inputId} label={label} error={error} hint={hint}>
      <div className="relative">
        {Icon && (
          <Icon
            size={16}
            strokeWidth={1.75}
            className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-ink-subtle"
            aria-hidden
          />
        )}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={clsx(controlClass(Boolean(error), Boolean(Icon)), trailing && 'pr-10', className)}
          {...props}
        />
        {trailing && <div className="absolute inset-y-0 right-1 flex items-center">{trailing}</div>}
      </div>
    </FieldShell>
  );
}

interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  error?: string;
  options: { value: string; label: string }[];
  ref?: Ref<HTMLSelectElement>;
}

export function SelectField({
  label,
  error,
  options,
  id,
  className,
  ref,
  ...props
}: Readonly<SelectFieldProps>) {
  const generated = useId();
  const selectId = id ?? generated;
  return (
    <FieldShell id={selectId} label={label} error={error}>
      <select
        ref={ref}
        id={selectId}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${selectId}-error` : undefined}
        className={clsx(controlClass(Boolean(error), false), className)}
        {...props}
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </FieldShell>
  );
}
