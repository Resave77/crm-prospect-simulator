import type { AdminUserRole } from '../types/admin'

export function adminTemporaryPasswordError(value: string): string | null {
  if (value.trim() === '' || Array.from(value).length < 6) return 'Temporary password must be at least 6 characters.'
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
