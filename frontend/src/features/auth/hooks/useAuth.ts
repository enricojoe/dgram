import { useMutation } from '@tanstack/react-query'
import { login as loginApi, register as registerApi } from '@/api/auth'
import { useAuthStore } from '../store/authStore'
import type { AuthResponse } from '@/types/auth'

interface Credentials {
  email: string
  password: string
}

/**
 * useAuth exposes login/register mutations that, on success, store the user and
 * tokens in the auth store, plus the current user and a logout action.
 */
export function useAuth() {
  const setAuth = useAuthStore((s) => s.setAuth)
  const logout = useAuthStore((s) => s.logout)
  const user = useAuthStore((s) => s.user)

  const onSuccess = (d: AuthResponse) =>
    setAuth(d.user, d.accessToken, d.refreshToken)

  const login = useMutation({
    mutationFn: ({ email, password }: Credentials) => loginApi(email, password),
    onSuccess,
  })

  const register = useMutation({
    mutationFn: ({ email, password }: Credentials) =>
      registerApi(email, password),
    onSuccess,
  })

  return { user, login, register, logout }
}
