import type { ChangePasswordPayload } from '../types/auth'

export interface PasswordRequirement {
  key: string
  label: string
  valid: boolean
}

export interface ChangePasswordValidation {
  requirements: PasswordRequirement[]
  valid: boolean
}

export function validateChangePassword(payload: ChangePasswordPayload): ChangePasswordValidation {
  const newPasswordLength = Array.from(payload.newPassword).length
  const requirements = [
    {
      key: 'minimum-length',
      label: 'minimum 6 characters',
      valid: payload.newPassword.trim() !== '' && newPasswordLength >= 6,
    },
    {
      key: 'different',
      label: 'new password differs from current password',
      valid: payload.newPassword.length > 0 && payload.newPassword !== payload.currentPassword,
    },
    {
      key: 'confirmation',
      label: 'confirmation matches',
      valid: payload.confirmPassword.length > 0 && payload.newPassword === payload.confirmPassword,
    },
  ]

  return {
    requirements,
    valid: requirements.every((requirement) => requirement.valid),
  }
}
