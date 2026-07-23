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
        { path: 'prospect-finder', name: 'AdminProspectFinder', component: () => import('../views/Admin/Prospect/ProspectFinderView.vue') },
        { path: 'prospects/pipeline', name: 'AdminProspectPipeline', component: () => import('../views/Admin/Prospect/ProspectPipelineView.vue') },
        { path: 'prospects', name: 'AdminProspects', redirect: '/admin/prospects/pipeline' },
        { path: 'prospects/won', name: 'AdminWonProspects', component: () => import('../views/Admin/Prospect/WonProspectsView.vue') },
        { path: 'prospects/:id/review', name: 'AdminProspectReview', component: () => import('../views/Admin/Prospect/ProspectReviewView.vue') },
        { path: 'prospects/:id/convert', name: 'AdminProspectConvert', component: () => import('../views/Admin/Prospect/ProspectConversionView.vue') },
        { path: 'customers', name: 'AdminCustomers', component: () => import('../views/Admin/Customer/CustomerListView.vue') },
        { path: 'customers/:id', name: 'AdminCustomerDetail', component: () => import('../views/Admin/Customer/CustomerDetailView.vue') },
        { path: 'customers/:id/edit', name: 'AdminCustomerEdit', component: () => import('../views/Admin/Customer/CustomerEditView.vue') },
        { path: 'companies', name: 'AdminCompanies', component: () => import('../views/Admin/Company/CompanyListView.vue') },
        { path: 'companies/:id', name: 'AdminCompanyDetail', component: () => import('../views/Admin/Company/CompanyDetailView.vue'), meta: { title: 'Company Detail', description: 'Company detail interface is under construction for the first release.' } },
        { path: 'companies/:id/edit', name: 'AdminCompanyEdit', component: () => import('../views/Admin/Company/CompanyEditView.vue') },
        { path: 'sales-executives', name: 'AdminSalesExecutives', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Sales Executive', description: 'Team management is represented by active seeded Sales Executive accounts in this simulation slice.' } },
        { path: 'customer-assignment', name: 'AdminCustomerAssignment', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Customer Assignment', description: 'Customer assignment workflow is deferred; converted customers retain their selected Sales Executive.' } },
        { path: 'visit-monitoring', name: 'AdminVisitMonitoring', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Visit Monitoring', description: 'GPS, selfie, check-in and check-out capture is planned for the attendance slice.' } },
        { path: 'prospect-assignment', name: 'AdminProspectAssignment', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Prospect Assignment', description: 'Assignment is available while saving a Place in Prospect Finder; bulk reassignment is deferred.' } },
        { path: 'reports', name: 'AdminReports', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Reports', description: 'Operational reporting is deferred until visit and attendance data exists.' } },
      ],
    },
    {
      path: '/sales', component: () => import('../layouts/SalesLayout.vue'), meta: { role: 'SALES_EXECUTIVE' },
      children: [
        { path: '', redirect: '/sales/dashboard' },
        { path: 'dashboard', name: 'SalesDashboard', component: () => import('../views/Sales/Dashboard/SalesDashboardView.vue') },
        { path: 'my-prospects', name: 'SalesMyProspects', component: () => import('../views/Sales/Prospect/MyProspectsView.vue') },
        { path: 'my-prospects/:id', name: 'SalesProspectDetail', component: () => import('../views/Sales/Prospect/ProspectDetailView.vue') },
        { path: 'my-customers', name: 'SalesMyCustomers', component: () => import('../views/Sales/Customer/MyCustomersView.vue') },
        { path: 'my-customers/:id', name: 'SalesCustomerDetail', component: () => import('../views/Sales/Customer/CustomerDetailView.vue') },
        { path: 'history', name: 'SalesHistory', component: () => import('../views/Sales/HistoryView.vue') },
        { path: 'profile', name: 'SalesProfile', component: () => import('../views/Sales/ProfileView.vue') },
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
