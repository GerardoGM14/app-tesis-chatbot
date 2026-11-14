import { X } from 'lucide-react';
import { type ReactNode, useEffect, useId, useRef } from 'react';
import { IconButton } from './Button';

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
}

/**
 * Diálogo modal sobre el elemento nativo <dialog>: el navegador se encarga del
 * foco atrapado, la tecla Escape y el fondo inerte.
 */
export function Dialog({ open, onClose, title, description, children, footer }: Readonly<DialogProps>) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      className="m-auto w-[calc(100%-2rem)] max-w-md rounded-lg border border-line bg-surface p-0 text-ink"
    >
      {open && (
        <div className="flex flex-col">
          <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4">
            <div>
              <h2 id={titleId} className="text-base font-semibold">
                {title}
              </h2>
              {description && (
                <p id={descriptionId} className="mt-1 text-sm text-ink-muted">
                  {description}
                </p>
              )}
            </div>
            <IconButton icon={X} label="Cerrar" onClick={onClose} />
          </header>
          <div className="px-5 py-4">{children}</div>
          {footer && (
            <footer className="flex justify-end gap-2 border-t border-line px-5 py-3">{footer}</footer>
          )}
        </div>
      )}
    </dialog>
  );
}
