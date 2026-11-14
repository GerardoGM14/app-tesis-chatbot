import { useEffect, useState } from 'react';

const dateFormat = new Intl.DateTimeFormat('es-PE', { dateStyle: 'medium' });
const dateTimeFormat = new Intl.DateTimeFormat('es-PE', { dateStyle: 'medium', timeStyle: 'short' });

export function formatDate(iso: string): string {
  return dateFormat.format(new Date(iso));
}

export function formatDateTime(iso: string | null): string {
  return iso ? dateTimeFormat.format(new Date(iso)) : 'Nunca';
}

/** Devuelve el valor después de `delay` ms sin cambios (búsquedas mientras se escribe). */
export function useDebouncedValue<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = window.setTimeout(() => {
      setDebounced(value);
    }, delay);
    return () => {
      window.clearTimeout(id);
    };
  }, [value, delay]);
  return debounced;
}
