import type { AdminUserRole } from '../types/admin'

export function adminTemporaryPasswordError(value: string): string | null {
  if (value.length < 8) return 'Temporary password must be at least 8 characters.'
  if (!/[A-Z]/.test(value)) return 'Temporary password must contain at least one uppercase letter.'
  if (!/[a-z]/.test(value)) return 'Temporary password must contain at least one lowercase letter.'
  if (!/[0-9]/.test(value)) return 'Temporary password must contain at least one number.'
  return null
}

export function adminRoleLabel(role: string) {
  switch (role) {
    case 'ADMINISTRATOR': return 'Administrator'
    case 'SALES_MANAGER': return 'Sales Manager'
    default: return 'Sales Executive'
  }
}

export function adminRoleSeverity(role: string) {
  switch (role) {
    case 'ADMINISTRATOR': return 'warn'
    case 'SALES_MANAGER': return 'info'
    default: return 'success'
  }
}

export function adminScopeSummary(role: AdminUserRole | '') {
  switch (role) {
    case 'ADMINISTRATOR':
      return {
        title: 'Administrator',
        scopes: [
          'Can manage accounts',
          'Can manage CRM',
          'ALL scope',
        ],
      }
    case 'SALES_MANAGER':
      return {
        title: 'Sales Manager',
        scopes: [
          'TEAM scope',
          'Read only',
        ],
      }
    case 'SALES_EXECUTIVE':
      return {
        title: 'Sales Executive',
        scopes: [
          'OWN scope',
          'Operational only',
        ],
      }
    default:
      return {
        title: 'No role selected',
        scopes: ['Select a role to see the account scope and permissions.'],
      }
  }
}
