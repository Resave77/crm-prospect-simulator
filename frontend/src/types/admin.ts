export type AdminUserRole = 'SUPER_ADMIN' | 'ADMINISTRATOR' | 'SALES_MANAGER' | 'SALES_EXECUTIVE'
export type AdminUserStatus = 'ACTIVE' | 'INACTIVE'
export type SalesRoleLevel = 1 | 2 | 3 | 4

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
  organizationalRole?: string
  organizationalRoleLevel?: SalesRoleLevel
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

export interface SalesRole {
  id: string
  name: string
  level: SalesRoleLevel
  description: string
  isActive: boolean
  createdBy?: string | null
  updatedBy?: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateSalesRolePayload {
  name: string
  level: SalesRoleLevel
  description?: string
}

export interface UpdateSalesRolePayload {
  name?: string
  level?: SalesRoleLevel
  description?: string | null
}

export interface SalesStructureAssignment {
  assignmentId: string
  userId: string
  salesRoleId: string
  parentUserId?: string | null
  effectiveFrom: string
  effectiveTo?: string | null
}

export interface SalesStructureItem {
  assignmentId: string
  userId: string
  salesName: string
  systemRole: AdminUserRole
  salesRole: {
    id: string
    name: string
    level: SalesRoleLevel
  }
  parentUserId?: string | null
  parentName?: string | null
  effectiveFrom: string
  effectiveTo?: string | null
}

export interface CreateSalesAssignmentPayload {
  userId: string
  salesRoleId: string
  parentUserId?: string | null
  effectiveFrom: string
}
