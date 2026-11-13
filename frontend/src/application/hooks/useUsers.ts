import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UserCreateValues } from '@/domain/schemas';
import { type UserPatch, usersApi } from '@/infrastructure/api/usersApi';

const keys = { all: ['usuarios'] as const };

export function useUsers(enabled = true) {
  return useQuery({
    queryKey: keys.all,
    queryFn: usersApi.list,
    select: (data) => data.items,
    enabled,
  });
}

function useInvalidateUsers() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: keys.all });
}

export function useCreateUser() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: (values: UserCreateValues) => usersApi.create(values),
    onSuccess: invalidate,
  });
}

export function useUpdateUser() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: UserPatch }) => usersApi.update(id, patch),
    onSuccess: invalidate,
  });
}

/** Restablece la contraseña y refresca la lista para mostrar el cambio obligatorio pendiente. */
export function useResetPassword() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: (id: string) => usersApi.resetPassword(id),
    onSuccess: invalidate,
  });
}
