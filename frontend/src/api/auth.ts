import { api } from './client'
import type { ApiEnvelope, AuthUser, ChangePasswordPayload, ChangePasswordResult } from '../types/auth'

export async function getCurrentUser(): Promise<AuthUser> {
  return (await api.get<ApiEnvelope<AuthUser>>('/auth/me')).data.data
}

export async function changePassword(payload: ChangePasswordPayload): Promise<ChangePasswordResult> {
  const response = await api.post<ApiEnvelope<ChangePasswordResult>>('/auth/change-password', payload)
  return response.data.data
}
