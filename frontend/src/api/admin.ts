import { api } from './client'
import type { ApiEnvelope } from '../types/auth'
import type {
  AdminCreateUserInput,
  AdminManagerOption,
  AdminPermission,
  AdminResetPasswordPayload,
  AdminResetPasswordResult,
  AdminUpdateUserInput,
  AdminUserDetail,
  AdminUserListParams,
  AdminUserListResult,
  AdminUserStatus,
  CreateSalesAssignmentPayload,
  CreateSalesRolePayload,
  SalesRole,
  SalesStructureAssignment,
  SalesStructureItem,
  UpdateSalesRolePayload,
} from '../types/admin'

function asArray<T>(value: unknown, fallbackMessage: string): T[] {
  if (Array.isArray(value)) return value as T[]
  throw new Error(fallbackMessage)
}

export async function getUsers(params: AdminUserListParams) {
  return (await api.get<ApiEnvelope<AdminUserListResult>>('/admin/users', { params })).data.data
}

export async function getUser(id: string) {
  return (await api.get<ApiEnvelope<AdminUserDetail>>(`/admin/users/${id}`)).data.data
}

export async function createUser(input: AdminCreateUserInput) {
  return (await api.post<ApiEnvelope<AdminUserDetail>>('/admin/users', input)).data.data
}

export async function updateUser(id: string, input: AdminUpdateUserInput) {
  return (await api.patch<ApiEnvelope<AdminUserDetail>>(`/admin/users/${id}`, input)).data.data
}

export async function deleteUser(id: string) {
  await api.delete(`/admin/users/${id}`)
}

export async function updateStatus(id: string, status: AdminUserStatus) {
  return (await api.patch<ApiEnvelope<AdminUserDetail>>(`/admin/users/${id}/status`, { status })).data.data
}

export async function resetUserPassword(userId: string, payload: AdminResetPasswordPayload) {
  return (await api.post<ApiEnvelope<AdminResetPasswordResult>>(`/admin/users/${userId}/reset-password`, payload)).data.data
}

export async function getManagerOptions() {
  return (await api.get<ApiEnvelope<AdminManagerOption[]>>('/admin/users/options/managers')).data.data
}

export async function getSalesRoles() {
  const response = await api.get<ApiEnvelope<unknown>>('/admin/sales-roles')
  return asArray<SalesRole>(response.data.data, 'Unable to load sales roles.')
}

export async function getSalesRole(id: string) {
  return (await api.get<ApiEnvelope<SalesRole>>(`/admin/sales-roles/${id}`)).data.data
}

export async function getPermissions(search?: string) {
  const response = await api.get<ApiEnvelope<unknown>>('/admin/permissions', {
    params: search ? { search } : undefined,
  })
  return asArray<AdminPermission>(response.data.data, 'Unable to load permissions.')
}

export async function createSalesRole(payload: CreateSalesRolePayload) {
  return (await api.post<ApiEnvelope<SalesRole>>('/admin/sales-roles', payload)).data.data
}

export async function updateSalesRole(id: string, payload: UpdateSalesRolePayload) {
  return (await api.patch<ApiEnvelope<SalesRole>>(`/admin/sales-roles/${id}`, payload)).data.data
}

export async function updateSalesRoleStatus(id: string, isActive: boolean) {
  return (
    await api.patch<ApiEnvelope<SalesRole>>(`/admin/sales-roles/${id}/status`, {
      isActive,
    })
  ).data.data
}

export async function deleteSalesRole(id: string) {
  await api.delete(`/admin/sales-roles/${id}`)
}

export async function getSalesStructure(effectiveDate: string) {
  const response = await api.get<ApiEnvelope<unknown>>('/admin/sales-structure', {
    params: { effectiveDate },
  })
  return asArray<SalesStructureItem>(
    response.data.data,
    'Unable to load sales structure.',
  )
}

export async function createSalesAssignment(payload: CreateSalesAssignmentPayload) {
  return (
    await api.post<ApiEnvelope<SalesStructureAssignment>>(
      '/admin/sales-structure/assignments',
      payload,
    )
  ).data.data
}

/**
 * Ends an assignment without deleting its history.
 *
 * Expected backend contract:
 * PATCH /api/v1/admin/sales-structure/assignments/:assignmentId/end
 * body: { effectiveTo: 'YYYY-MM-DD' }
 */
export async function endSalesAssignment(
  assignmentId: string,
  effectiveTo: string,
) {
  return (
    await api.patch<ApiEnvelope<SalesStructureAssignment>>(
      `/admin/sales-structure/assignments/${assignmentId}/end`,
      { effectiveTo },
    )
  ).data.data
}