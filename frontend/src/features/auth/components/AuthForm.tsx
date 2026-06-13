import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AxiosError } from 'axios'
import { useAuth } from '../hooks/useAuth'

/**
 * AuthForm renders the login or register form depending on `mode`. On success
 * it navigates to the dashboard. Shared markup keeps the two pages in sync.
 */
export default function AuthForm({ mode }: { mode: 'login' | 'register' }) {
  const navigate = useNavigate()
  const { login, register } = useAuth()
  const mutation = mode === 'login' ? login : register

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate(
      { email, password },
      { onSuccess: () => navigate('/dashboard') },
    )
  }

  const error = mutation.error
    ? mutation.error instanceof AxiosError
      ? (mutation.error.response?.data?.error ?? mutation.error.message)
      : 'Something went wrong'
    : null

  return (
    <div className="mx-auto mt-20 w-full max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900">
      <h1 className="mb-4 text-xl font-semibold text-gray-800 dark:text-gray-100">
        {mode === 'login' ? 'Log in' : 'Create account'}
      </h1>

      <form onSubmit={submit} className="flex flex-col gap-3">
        <input
          type="email"
          required
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="rounded border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-950 dark:text-gray-100"
        />
        <input
          type="password"
          required
          minLength={6}
          placeholder="Password (min 6 chars)"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="rounded border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-950 dark:text-gray-100"
        />

        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}

        <button
          type="submit"
          disabled={mutation.isPending}
          className="rounded bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
        >
          {mutation.isPending
            ? 'Please wait…'
            : mode === 'login'
              ? 'Log in'
              : 'Sign up'}
        </button>
      </form>

      <p className="mt-4 text-center text-sm text-gray-500 dark:text-gray-400">
        {mode === 'login' ? (
          <>
            No account?{' '}
            <Link to="/register" className="text-indigo-600 hover:underline">
              Sign up
            </Link>
          </>
        ) : (
          <>
            Have an account?{' '}
            <Link to="/login" className="text-indigo-600 hover:underline">
              Log in
            </Link>
          </>
        )}
      </p>
    </div>
  )
}
