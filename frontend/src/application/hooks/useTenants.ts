import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { TenantCreateValues } from '@/domain/schemas';
import { tenantsApi } from '@/infrastructure/api/tenantsApi';

const keys = {
  all: ['empresas'] as const,
  list: (search: string) => ['empresas', 'lista', search] as const,
  admins: (id: string) => ['empresas', id, 'administradores'] as const,
};

export function useTenants(search: string) {
  return useQuery({
    queryKey: keys.list(search),
    queryFn: () => tenantsApi.list(search),
    select: (data) => data.items,
  });
}

export function useTenantAdmins(tenantId: string | null) {
  return useQuery({
    queryKey: keys.admins(tenantId ?? ''),
    queryFn: () => tenantsApi.admins(tenantId ?? ''),
    select: (data) => data.items,
    enabled: tenantId !== null,
  });
}

function useInvalidateTenants() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: keys.all });
}

export function useCreateTenant() {
  const invalidate = useInvalidateTenants();
  return useMutation({
    mutationFn: (values: TenantCreateValues) => tenantsApi.create(values),
    onSuccess: invalidate,
  });
}

export function useSetTenantStatus() {
  const invalidate = useInvalidateTenants();
  return useMutation({
    mutationFn: ({ id, suspend }: { id: string; suspend: boolean }) =>
      suspend ? tenantsApi.suspend(id) : tenantsApi.reactivate(id),
    onSuccess: invalidate,
  });
}

export function useRenameTenant() {
  const invalidate = useInvalidateTenants();
  return useMutation({
    mutationFn: ({ id, nombre }: { id: string; nombre: string }) => tenantsApi.rename(id, nombre),
    onSuccess: invalidate,
  });
}
