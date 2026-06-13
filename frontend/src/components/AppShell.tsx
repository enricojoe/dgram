import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/store/authStore'

/**
 * AppShell is the global layout: a top header with brand + auth navigation, and
 * an Outlet that fills the rest of the viewport for the active route.
 */
export default function AppShell() {
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()

  const onLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center justify-between border-b border-gray-200 px-4 py-2">
        <Link to="/" className="text-lg font-bold text-indigo-700">
          DGram
        </Link>

        <nav className="flex items-center gap-3 text-sm">
          {user ? (
            <>
              <Link to="/dashboard" className="text-gray-700 hover:text-indigo-700">
                My Diagrams
              </Link>
              <span className="text-gray-400">{user.email}</span>
              <button
                type="button"
                onClick={onLogout}
                className="rounded border border-gray-300 px-2 py-1 hover:bg-gray-50"
              >
                Log out
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="text-gray-700 hover:text-indigo-700">
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
