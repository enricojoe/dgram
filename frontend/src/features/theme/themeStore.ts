import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type Theme = 'light' | 'dark'

interface ThemeState {
  theme: Theme
  setTheme: (theme: Theme) => void
  toggle: () => void
}

/** True when the OS prefers a dark color scheme (used as the first-visit default). */
function osPrefersDark(): boolean {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches
  )
}

/**
 * Theme state, persisted to localStorage. On first visit (no persisted value)
 * the initializer falls back to the OS preference. The `.dark` class on <html>
 * is applied by `useApplyTheme`, which subscribes to this store.
 */
export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: osPrefersDark() ? 'dark' : 'light',
      setTheme: (theme) => set({ theme }),
      toggle: () => set({ theme: get().theme === 'dark' ? 'light' : 'dark' }),
    }),
    { name: 'dgram-theme' },
  ),
)
