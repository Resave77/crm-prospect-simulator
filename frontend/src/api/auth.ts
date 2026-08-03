import { api } from './client'
import type { ApiEnvelope, ChangePasswordPayload, ChangePasswordResult } from '../types/auth'

export async function changePassword(payload: ChangePasswordPayload): Promise<ChangePasswordResult> {
  const response = await api.post<ApiEnvelope<ChangePasswordResult>>('/auth/change-password', payload)
  return response.data.data
}
