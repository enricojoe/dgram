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
