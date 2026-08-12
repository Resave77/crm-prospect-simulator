import type { Router } from 'vue-router'
import type { AuthUser, UserRole } from '../types/auth'

export const ADMIN_DASHBOARD_ROUTE = '/admin/dashboard'
export const SALES_DASHBOARD_ROUTE = '/sales/dashboard'
export const FORBIDDEN_ROUTE = '/forbidden'

export function isAdminRole(role: UserRole) {
  return role === 'SUPER_ADMIN' || role === 'ADMINISTRATOR'
}

export function homeFor(user: Pick<AuthUser, 'role' | 'salesRole'>) {
  const landing = user.salesRole?.landingPage
  if (isAdminRole(user.role)) {
    return landing && landing.startsWith('/admin') ? landing : ADMIN_DASHBOARD_ROUTE
  }
  if (landing && landing.startsWith('/sales')) return landing
  if (user.role === 'SALES_MANAGER') return FORBIDDEN_ROUTE
  return SALES_DASHBOARD_ROUTE
}

export function roleAllowed(required: UserRole, actual: UserRole, hasSalesRole = false) {
  if (required === 'ADMINISTRATOR') return actual === 'SUPER_ADMIN' || actual === 'ADMINISTRATOR'
  if (required === 'SALES_EXECUTIVE') {
    return actual === 'SALES_EXECUTIVE' || (actual === 'SALES_MANAGER' && hasSalesRole)
  }
  return required === actual
}

export function permissionKeysFor(user: Pick<AuthUser, 'role' | 'salesRole'>) {
  return user.salesRole?.permissionKeys ?? []
}

export function hasPermission(user: Pick<AuthUser, 'role' | 'salesRole'>, key?: string | null) {
  if (!key) return true
  if (user.role === 'SUPER_ADMIN') return true
  return permissionKeysFor(user).includes(key)
}

export function routeExists(router: Router, path: string) {
  const resolved = router.resolve(path)
  return resolved.matched.length > 0 && resolved.name !== 'NotFound'
}

export function routePermitted(router: Router, path: string, user: AuthUser) {
  const resolved = router.resolve(path)
  if (resolved.matched.length === 0 || resolved.name === 'NotFound') return false
  return resolved.matched.every((record) => {
    const required = record.meta.role as UserRole | undefined
    const permission = record.meta.permission as string | undefined
    return (!required || roleAllowed(required, user.role, Boolean(user.salesRole))) && hasPermission(user, permission)
  })
}

export function resolvePostLoginRoute(router: Router, user: AuthUser, intended?: string | null) {
  if (intended && intended.startsWith('/') && routePermitted(router, intended, user)) {
    return intended
  }
  const fallback = homeFor(user)
  if (routeExists(router, fallback) && routePermitted(router, fallback, user)) {
    return fallback
  }
  return isAdminRole(user.role) ? ADMIN_DASHBOARD_ROUTE : SALES_DASHBOARD_ROUTE
}
