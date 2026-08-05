import type { Router } from 'vue-router'
import type { UserRole } from '../types/auth'

export const ADMIN_DASHBOARD_ROUTE = '/admin/dashboard'
export const SALES_DASHBOARD_ROUTE = '/sales/dashboard'
export const FORBIDDEN_ROUTE = '/forbidden'

export function homeFor(role: UserRole) {
  if (role === 'SUPER_ADMIN' || role === 'ADMINISTRATOR') return ADMIN_DASHBOARD_ROUTE
  if (role === 'SALES_MANAGER') return FORBIDDEN_ROUTE
  return SALES_DASHBOARD_ROUTE
}

export function roleAllowed(required: UserRole, actual: UserRole) {
  if (required === 'ADMINISTRATOR') return actual === 'SUPER_ADMIN' || actual === 'ADMINISTRATOR'
  return required === actual
}

export function routeExists(router: Router, path: string) {
  const resolved = router.resolve(path)
  return resolved.matched.length > 0 && resolved.name !== 'NotFound'
}

export function routePermitted(router: Router, path: string, role: UserRole) {
  const resolved = router.resolve(path)
  if (resolved.matched.length === 0 || resolved.name === 'NotFound') return false
  return resolved.matched.every((record) => {
    const required = record.meta.role as UserRole | undefined
    return !required || roleAllowed(required, role)
  })
}

export function resolvePostLoginRoute(router: Router, role: UserRole, intended?: string | null) {
  if (intended && intended.startsWith('/') && routePermitted(router, intended, role)) {
    return intended
  }
  const fallback = homeFor(role)
  if (routeExists(router, fallback) && routePermitted(router, fallback, role)) {
    return fallback
  }
  return role === 'SUPER_ADMIN' || role === 'ADMINISTRATOR' ? ADMIN_DASHBOARD_ROUTE : SALES_DASHBOARD_ROUTE
}
