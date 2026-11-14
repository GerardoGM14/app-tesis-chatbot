import type { FieldValues, Path, UseFormSetError } from 'react-hook-form';
import { toast } from 'sonner';
import { ApiError } from '@/infrastructure/http/client';

/**
 * Lleva los errores de validación del backend a los campos del formulario.
 * Si el error no corresponde a un campo conocido, lo muestra como toast.
 */
export function applyApiError<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
  knownFields: readonly Path<T>[],
): void {
  if (error instanceof ApiError) {
    const entries = Object.entries(error.fields).filter(([field]) =>
      (knownFields as readonly string[]).includes(field),
    );
    if (entries.length > 0) {
      for (const [field, message] of entries) {
        setError(field as Path<T>, { type: 'server', message });
      }
      return;
    }
    toast.error(error.message);
    return;
  }
  toast.error('Ocurrió un error inesperado');
}

export function errorMessage(error: unknown): string {
  return error instanceof ApiError ? error.message : 'Ocurrió un error inesperado';
}
