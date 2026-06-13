import { api } from './client'
import type { AuthResponse, User } from '@/types/auth'

/** Register a new account; returns the user and a fresh token pair. */
export async function register(
  email: string,
  password: string,
): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/register', {
    email,
    password,
  })
  return data
}

/** Log in with email + password. */
export async function login(
  email: string,
  password: string,
): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>('/auth/login', {
    email,
    password,
  })
  return data
}

/** Fetch the currently authenticated user. */
export async function getMe(): Promise<User> {
  const { data } = await api.get<User>('/me')
  return data
}

/** Update the current user's profile (display name). */
export async function updateProfile(displayName: string): Promise<User> {
  const { data } = await api.patch<User>('/me', { displayName })
  return data
}

/** Change the current user's password. Requires the current password. */
export async function changePassword(
  oldPassword: string,
  newPassword: string,
): Promise<void> {
  await api.post('/me/password', { oldPassword, newPassword })
}
