import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { pinia } from '../stores/pinia'
import { useAuthStore } from '../stores/auth'
import type { UserRole } from '../types/auth'

declare module 'vue-router' {
  interface RouteMeta {
    public?: boolean
    role?: UserRole
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', name: 'Login', component: () => import('../views/Login/LoginView.vue'), meta: { public: true } },
    {
      path: '/admin', component: () => import('../layouts/AdminLayout.vue'), meta: { role: 'ADMINISTRATOR' },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', name: 'AdminDashboard', component: () => import('../views/Admin/Dashboard/AdminDashboardView.vue') },
      ],
    },
    {
      path: '/sales', component: () => import('../layouts/SalesLayout.vue'), meta: { role: 'SALES_EXECUTIVE' },
      children: [
        { path: '', redirect: '/sales/dashboard' },
        { path: 'dashboard', name: 'SalesDashboard', component: () => import('../views/Sales/Dashboard/SalesDashboardView.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('../views/NotFoundView.vue'), meta: { public: true } },
  ],
})

function homeFor(role: UserRole) {
  return role === 'ADMINISTRATOR' ? '/admin/dashboard' : '/sales/dashboard'
}

router.beforeEach(async (to: RouteLocationNormalized) => {
  const auth = useAuthStore(pinia)
  await auth.bootstrap()

  if (to.name === 'Login' && auth.role) return homeFor(auth.role)
  if (to.meta.public) return true
  if (!auth.authenticated) return { name: 'Login', query: { redirect: to.fullPath } }
  if (to.meta.role && to.meta.role !== auth.role) return homeFor(auth.role!)
  return true
})

export default router
