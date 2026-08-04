export type AdminUserRole = 'SUPER_ADMIN' | 'ADMINISTRATOR' | 'SALES_MANAGER' | 'SALES_EXECUTIVE'
export type AdminUserStatus = 'ACTIVE' | 'INACTIVE'
export type SalesRoleLevel = 1 | 2 | 3 | 4
export type PermissionNodeType = 'GROUP' | 'MENU' | 'ACTION'

export interface AdminPermission {
  id: string
  key: string
  name: string
  description?: string
  groupKey: string
  parentKey?: string | null
  nodeType: PermissionNodeType
  routePath?: string | null
  isActive: boolean
  sortOrder: number
}

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
  organizationalRole: AdminOrganizationalRoleSummary | null
  mustChangePassword: boolean
  createdAt: string
  updatedAt: string
}

export interface AdminOrganizationalRoleSummary {
  id: string
  name: string
  level: SalesRoleLevel
  landingPage?: string | null
  permissionCount: number
  isActive: boolean
  description?: string
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
  salesRoleId: string | null
  managerId: string | null
  temporaryPassword: string
}

export interface AdminUpdateUserInput {
  employeeId?: string
  name?: string
  email?: string
  phone?: string
  salesRoleId?: string | null
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
  landingPage?: string | null
  permissionCount?: number
  permissions?: AdminPermission[]
  createdBy?: string | null
  updatedBy?: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateSalesRolePayload {
  name: string
  level: SalesRoleLevel
  description?: string
  landingPage: string
  permissionKeys: string[]
}

export interface UpdateSalesRolePayload {
  name?: string
  level?: SalesRoleLevel
  description?: string | null
  landingPage: string
  permissionKeys: string[]
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
