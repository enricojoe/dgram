import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/store/authStore'

/**
 * ProtectedRoute redirects unauthenticated users to /login, remembering where
 * they were headed so login can send them back.
 */
export default function ProtectedRoute({
  children,
}: {
  children: React.ReactNode
}) {
  const token = useAuthStore((s) => s.accessToken)
  const location = useLocation()

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }
  return <>{children}</>
}
