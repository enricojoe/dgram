import axios, {
  AxiosError,
  type InternalAxiosRequestConfig,
} from 'axios'
import { useAuthStore } from '@/features/auth/store/authStore'

// Base URL for all API calls. Defaults to "/api" (proxied to the backend in dev).
const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api'

/**
 * Shared axios instance. In dev, `/api/*` is proxied to the Go backend by Vite.
 * A request interceptor attaches the JWT; a response interceptor transparently
 * refreshes the token once on a 401 and retries the original request.
 */
export const api = axios.create({
  baseURL,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Bare client for the refresh call so it doesn't recurse through the
// response interceptor below.
const refreshClient = axios.create({ baseURL })

type RetriableConfig = InternalAxiosRequestConfig & { _retry?: boolean }

api.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config as RetriableConfig | undefined
    const { refreshToken, setTokens, logout } = useAuthStore.getState()

    const shouldRefresh =
      error.response?.status === 401 &&
      original &&
      !original._retry &&
      !!refreshToken &&
      !original.url?.includes('/auth/')

    if (!shouldRefresh) return Promise.reject(error)

    original._retry = true
    try {
      const { data } = await refreshClient.post<{
        accessToken: string
        refreshToken: string
      }>('/auth/refresh', { refreshToken })
      setTokens(data.accessToken, data.refreshToken)
      original.headers.Authorization = `Bearer ${data.accessToken}`
      return api(original)
    } catch (refreshErr) {
      logout()
      return Promise.reject(refreshErr)
    }
  },
)
