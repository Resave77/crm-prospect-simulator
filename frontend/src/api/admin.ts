import { api } from './client'
import type { ApiEnvelope } from '../types/auth'
import type {
  AdminCreateUserInput,
  AdminManagerOption,
  AdminResetPasswordPayload,
  AdminResetPasswordResult,
  AdminUpdateUserInput,
  AdminUserDetail,
  AdminUserListParams,
  AdminUserListResult,
  AdminUserStatus,
} from '../types/admin'

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

export async function updateStatus(id: string, status: AdminUserStatus) {
  return (await api.patch<ApiEnvelope<AdminUserDetail>>(`/admin/users/${id}/status`, { status })).data.data
}

export async function resetUserPassword(userId: string, payload: AdminResetPasswordPayload) {
  return (await api.post<ApiEnvelope<AdminResetPasswordResult>>(`/admin/users/${userId}/reset-password`, payload)).data.data
}

export async function getManagerOptions() {
  return (await api.get<ApiEnvelope<AdminManagerOption[]>>('/admin/users/options/managers')).data.data
}
