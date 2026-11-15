import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from 'react-router';
import { Toaster } from 'sonner';
import { ApiError } from '@/infrastructure/http/client';
import { router } from './router';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      // Los errores 4xx no se reintentan: repetirlos no cambia el resultado.
      retry: (failureCount, error) => !(error instanceof ApiError && error.status < 500) && failureCount < 2,
    },
  },
});

export function Providers() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster
        position="bottom-right"
        closeButton
        toastOptions={{
          unstyled: true,
          classNames: {
            toast:
              'flex w-[360px] items-start gap-3 rounded-lg border border-line bg-surface px-4 py-3 text-sm text-ink',
            title: 'font-medium',
            description: 'text-ink-muted',
            closeButton: 'text-ink-subtle hover:text-ink',
            success: '[&_[data-icon]]:text-success-600',
            error: 'border-danger-600/30 [&_[data-icon]]:text-danger-600',
            info: '[&_[data-icon]]:text-brand-600',
          },
        }}
      />
    </QueryClientProvider>
  );
}
