import { Link, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/store/authStore'
import ThemeToggle from '@/features/theme/ThemeToggle'
import UserMenu from './UserMenu'

/**
 * AppShell is the global layout: a top header with brand + auth navigation, and
 * an Outlet that fills the rest of the viewport for the active route. Signed-in
 * users get an avatar dropdown (UserMenu); anonymous users get auth links plus a
 * standalone theme toggle.
 */
export default function AppShell() {
  const user = useAuthStore((s) => s.user)

  return (
    <div className="flex h-full flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-gray-100">
      <header className="flex items-center justify-between border-b border-gray-200 px-4 py-2 dark:border-gray-800">
        <Link
          to="/"
          className="flex items-center gap-2 text-lg font-bold text-indigo-700 dark:text-indigo-400"
        >
          <img src="/favicon.svg" alt="" className="h-6 w-6" />
          DGram
        </Link>

        <nav className="flex items-center gap-2 text-sm">
          {user ? (
            <UserMenu />
          ) : (
            <>
              <ThemeToggle />
              <Link
                to="/login"
                className="rounded px-2 py-1 text-gray-700 hover:text-indigo-700 dark:text-gray-300 dark:hover:text-indigo-400"
              >
                Log in
              </Link>
              <Link
                to="/register"
                className="rounded bg-indigo-600 px-3 py-1 font-medium text-white hover:bg-indigo-700"
              >
                Sign up
              </Link>
            </>
          )}
        </nav>
      </header>

      <main className="min-h-0 flex-1">
        <Outlet />
      </main>
    </div>
  )
}
