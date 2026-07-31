export type AdminUserRole = 'ADMINISTRATOR' | 'SALES_MANAGER' | 'SALES_EXECUTIVE'
export type AdminUserStatus = 'ACTIVE' | 'INACTIVE'

export interface AdminUserListItem {
  id: string
  employeeId: string
  fullName: string
  email: string
  phone: string
  role: AdminUserRole
  status: AdminUserStatus
  managerId: string | null
  managerName: string
  mustChangePassword: boolean
  createdAt: string
  updatedAt: string
}

export interface AdminUserDetail extends AdminUserListItem {
  createdBy: string | null
  updatedBy: string | null
}

export interface AdminManagerOption {
  id: string
  employeeId: string
  name: string
  email: string
}

export interface AdminUserListParams {
  page: number
  limit: number
  search: string
  role: AdminUserRole | ''
  status: AdminUserStatus | ''
}

export interface AdminUserListResult {
  data: AdminUserListItem[]
  total: number
  page: number
  limit: number
  pages: number
}

export interface AdminCreateUserInput {
  employeeId: string
  name: string
  email: string
  phone: string
  role: AdminUserRole
  managerId: string | null
  temporaryPassword: string
}

export interface AdminUpdateUserInput {
  employeeId?: string
  name?: string
  email?: string
  phone?: string
  role?: AdminUserRole
  managerId?: string | null
}

export type AdminResetPasswordMode = 'AUTO' | 'MANUAL'

export interface AdminResetPasswordPayload {
  mode: AdminResetPasswordMode
  temporaryPassword?: string
}

export interface AdminResetPasswordResult {
  userId: string
  temporaryPassword: string
  mustChangePassword: boolean
  sessionsRevoked: number
}
