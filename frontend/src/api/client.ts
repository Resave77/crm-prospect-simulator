import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { ApiEnvelope, AuthPayload } from '../types/auth'

interface RetryableRequest extends InternalAxiosRequestConfig {
  _retry?: boolean
}

export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: { Accept: 'application/json' },
})

let accessToken: string | null = null
let refreshPromise: Promise<AuthPayload> | null = null
let sessionObserver: ((payload: AuthPayload | null) => void) | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function observeSession(observer: (payload: AuthPayload | null) => void) {
  sessionObserver = observer
}

export async function refreshSession(): Promise<AuthPayload> {
  if (!refreshPromise) {
    refreshPromise = axios
      .post<ApiEnvelope<AuthPayload>>('/api/v1/auth/refresh', {}, { withCredentials: true })
      .then((response) => {
        setAccessToken(response.data.data.accessToken)
        sessionObserver?.(response.data.data)
        return response.data.data
      })
      .catch((error) => {
        setAccessToken(null)
        sessionObserver?.(null)
        throw error
      })
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

api.interceptors.request.use((config) => {
  if (accessToken) config.headers.Authorization = `Bearer ${accessToken}`
  return config
})

api.interceptors.response.use(undefined, async (error: AxiosError) => {
  const request = error.config as RetryableRequest | undefined
  const isAuthEndpoint = request?.url?.startsWith('/auth/') ?? false
  if (error.response?.status !== 401 || !request || request._retry || isAuthEndpoint) {
    throw error
  }
  request._retry = true
  const payload = await refreshSession()
  request.headers.Authorization = `Bearer ${payload.accessToken}`
  return api(request)
})
