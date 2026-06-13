import { useEffect } from 'react'
import { useThemeStore } from './themeStore'

/**
 * Syncs the active theme to the `<html>` element's class list so Tailwind's
 * `.dark` variant takes effect. Mounted once near the app root.
 */
export function useApplyTheme() {
  const theme = useThemeStore((s) => s.theme)

  useEffect(() => {
    const root = document.documentElement
    root.classList.toggle('dark', theme === 'dark')
    root.style.colorScheme = theme
  }, [theme])
}
