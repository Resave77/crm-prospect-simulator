export type UserRole = 'SUPER_ADMIN' | 'ADMINISTRATOR' | 'SALES_MANAGER' | 'SALES_EXECUTIVE'

export type SalesRoleLevel = 1 | 2 | 3 | 4

export interface AuthSalesRole {
  id: string
  name: string
  level: SalesRoleLevel
  landingPage?: string | null
  permissionKeys?: string[]
}

export interface AuthUser {
  id: string
  email: string
  fullName: string
  employeeId?: string
  phone?: string
  role: UserRole
  mustChangePassword?: boolean
  managerId?: string
  salesRole?: AuthSalesRole | null
}

export interface AuthPayload {
  accessToken: string
  accessExpiresAt: string
  user: AuthUser
}

export interface ChangePasswordPayload {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

export interface ChangePasswordResult {
  passwordChanged: boolean
  mustChangePassword: boolean
  sessionsRevoked: number
  reauthenticationRequired: boolean
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

