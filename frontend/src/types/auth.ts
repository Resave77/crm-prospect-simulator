export type UserRole = 'ADMINISTRATOR' | 'SALES_EXECUTIVE'

export interface AuthUser {
  id: string
  email: string
  fullName: string
  role: UserRole
}

export interface AuthPayload {
  accessToken: string
  accessExpiresAt: string
  user: AuthUser
}

export interface ApiEnvelope<T> {
  data: T
  meta: { requestId?: string }
}

export interface ApiErrorEnvelope {
  error?: {
    code?: string
    message?: string
    fields?: Record<string, string>
    requestId?: string
  }
}
