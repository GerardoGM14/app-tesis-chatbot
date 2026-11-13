import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { ChangePasswordValues, LoginValues } from '@/domain/schemas';
import { authApi } from '@/infrastructure/api/authApi';
import { logout } from '../session';
import { useSessionStore } from '../stores/sessionStore';

export function useCurrentUser() {
  return useSessionStore((s) => s.user);
}

export function useLogin() {
  const setSession = useSessionStore((s) => s.setSession);
  return useMutation({
    mutationFn: (values: LoginValues) => authApi.login(values.email, values.password),
    onSuccess: setSession,
  });
}

export function useChangePassword() {
  const setSession = useSessionStore((s) => s.setSession);
  return useMutation({
    mutationFn: (values: ChangePasswordValues) => authApi.changePassword(values.actual, values.nueva),
    onSuccess: setSession,
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: logout,
    onSettled: () => {
      // Ningún dato de la sesión anterior debe quedar en caché.
      queryClient.clear();
    },
  });
}
