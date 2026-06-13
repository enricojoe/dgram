import { useState } from 'react'
import { AxiosError } from 'axios'
import { useAuthStore } from '@/features/auth/store/authStore'
import { useProfile } from '@/features/auth/hooks/useProfile'
import { displayNameOf, initialsOf } from '@/features/auth/identity'
import { useDiagrams } from '@/api/diagrams'

function errorMessage(err: unknown): string {
  if (err instanceof AxiosError) {
    return err.response?.data?.error ?? err.message
  }
  return 'Something went wrong'
}

/**
 * ProfilePage shows the signed-in user's account summary and lets them update
 * their display name and change their password. Both forms report inline
 * success/error and use the backend /me + /me/password endpoints.
 */
export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)
  const { updateProfile, changePassword } = useProfile()
  const { data: diagrams } = useDiagrams()

  const [displayName, setDisplayName] = useState(user?.displayName ?? '')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [pwError, setPwError] = useState<string | null>(null)
  const [pwDone, setPwDone] = useState(false)

  if (!user) return null

  const submitName = (e: React.FormEvent) => {
    e.preventDefault()
    updateProfile.mutate(displayName)
  }

  const submitPassword = (e: React.FormEvent) => {
    e.preventDefault()
    setPwError(null)
    setPwDone(false)
    if (newPassword !== confirm) {
      setPwError('New passwords do not match')
      return
    }
    changePassword.mutate(
      { oldPassword, newPassword },
      {
        onSuccess: () => {
          setPwDone(true)
          setOldPassword('')
          setNewPassword('')
          setConfirm('')
        },
        onError: (err) => setPwError(errorMessage(err)),
      },
    )
  }

  const memberSince = new Date(user.createdAt).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const inputClass =
    'rounded border border-gray-300 px-3 py-2 text-sm dark:border-gray-700 dark:bg-gray-900 dark:text-gray-100'
  const cardClass =
    'rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900'

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      {/* Summary */}
      <div className={`flex items-center gap-4 ${cardClass}`}>
        <span className="flex h-16 w-16 items-center justify-center rounded-full bg-indigo-600 text-xl font-semibold text-white">
          {initialsOf(user)}
        </span>
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold text-gray-900 dark:text-gray-100">
            {displayNameOf(user)}
          </h1>
          <p className="truncate text-sm text-gray-500 dark:text-gray-400">
            {user.email}
          </p>
          <p className="mt-1 text-xs text-gray-400 dark:text-gray-500">
            Member since {memberSince}
            {diagrams && ` · ${diagrams.length} diagram${diagrams.length === 1 ? '' : 's'}`}
          </p>
        </div>
      </div>

      {/* Display name */}
      <form onSubmit={submitName} className={`flex flex-col gap-3 ${cardClass}`}>
        <h2 className="text-sm font-semibold text-gray-800 dark:text-gray-200">
          Display name
        </h2>
        <input
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="Your name"
          maxLength={80}
          className={inputClass}
          aria-label="Display name"
        />
        <div className="flex items-center gap-3">
          <button
            type="submit"
            disabled={updateProfile.isPending}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {updateProfile.isPending ? 'Saving…' : 'Save name'}
          </button>
          {updateProfile.isSuccess && (
            <span className="text-sm text-green-600 dark:text-green-400">Saved</span>
          )}
          {updateProfile.isError && (
            <span className="text-sm text-red-600 dark:text-red-400">
              {errorMessage(updateProfile.error)}
            </span>
          )}
        </div>
      </form>

      {/* Change password */}
      <form onSubmit={submitPassword} className={`flex flex-col gap-3 ${cardClass}`}>
        <h2 className="text-sm font-semibold text-gray-800 dark:text-gray-200">
          Change password
        </h2>
        <input
          type="password"
          required
          placeholder="Current password"
          value={oldPassword}
          onChange={(e) => setOldPassword(e.target.value)}
          className={inputClass}
          aria-label="Current password"
        />
        <input
          type="password"
          required
          minLength={6}
          placeholder="New password (min 6 chars)"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          className={inputClass}
          aria-label="New password"
        />
        <input
          type="password"
          required
          minLength={6}
          placeholder="Confirm new password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          className={inputClass}
          aria-label="Confirm new password"
        />

        {pwError && <p className="text-sm text-red-600 dark:text-red-400">{pwError}</p>}
        {pwDone && (
          <p className="text-sm text-green-600 dark:text-green-400">
            Password updated
          </p>
        )}

        <div>
          <button
            type="submit"
            disabled={changePassword.isPending}
            className="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {changePassword.isPending ? 'Updating…' : 'Update password'}
          </button>
        </div>
      </form>
    </div>
  )
}
