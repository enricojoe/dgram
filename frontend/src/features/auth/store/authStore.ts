import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth'

/**
 * Auth state, persisted to localStorage so a refresh keeps the user logged in.
 * The axios client reads `accessToken` from here to authorize requests, and
 * updates the tokens after a refresh.
 */
interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null

  setAuth: (user: User, accessToken: string, refreshToken: string) => void
  setTokens: (accessToken: string, refreshToken: string) => void
  updateUser: (partial: Partial<User>) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,

      setAuth: (user, accessToken, refreshToken) =>
        set({ user, accessToken, refreshToken }),
      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),
      updateUser: (partial) =>
        set((s) => (s.user ? { user: { ...s.user, ...partial } } : {})),
      logout: () => set({ user: null, accessToken: null, refreshToken: null }),
    }),
    { name: 'dgram-auth' },
  ),
)

/** Convenience selector used by guards/UI. */
export const isAuthenticated = () => !!useAuthStore.getState().accessToken
