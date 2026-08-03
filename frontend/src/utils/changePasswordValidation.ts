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
  const requirements = [
    {
      key: 'minimum-length',
      label: 'minimum 8 characters',
      valid: payload.newPassword.length >= 8,
    },
    {
      key: 'uppercase',
      label: 'one uppercase letter',
      valid: /[A-Z]/.test(payload.newPassword),
    },
    {
      key: 'lowercase',
      label: 'one lowercase letter',
      valid: /[a-z]/.test(payload.newPassword),
    },
    {
      key: 'digit',
      label: 'one digit',
      valid: /\d/.test(payload.newPassword),
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
