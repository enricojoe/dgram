import { useMutation } from '@tanstack/react-query'
import {
  changePassword as changePasswordApi,
  updateProfile as updateProfileApi,
} from '@/api/auth'
import { useAuthStore } from '../store/authStore'

/**
 * useProfile exposes mutations for the profile page: updating the display name
 * (which syncs the cached user) and changing the password.
 */
export function useProfile() {
  const updateUser = useAuthStore((s) => s.updateUser)

  const updateProfile = useMutation({
    mutationFn: (displayName: string) => updateProfileApi(displayName),
    onSuccess: (user) => updateUser(user),
  })

  const changePassword = useMutation({
    mutationFn: ({
      oldPassword,
      newPassword,
    }: {
      oldPassword: string
      newPassword: string
    }) => changePasswordApi(oldPassword, newPassword),
  })

  return { updateProfile, changePassword }
}
