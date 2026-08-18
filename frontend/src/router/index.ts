import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { pinia } from '../stores/pinia'
import { useAuthStore } from '../stores/auth'
import type { UserRole } from '../types/auth'
import { FORBIDDEN_ROUTE, hasPermission, homeFor, roleAllowed, routePermitted } from '../utils/navigation'

declare module 'vue-router' {
  interface RouteMeta {
    public?: boolean
    role?: UserRole
    permission?: string
    entityType?: 'prospect' | 'customer'
    title?: string
    description?: string
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', name: 'Login', component: () => import('../views/Login/LoginView.vue'), meta: { public: true } },
    { path: '/change-password', name: 'ChangePassword', component: () => import('../views/Auth/ChangePasswordView.vue') },
    { path: '/forbidden', name: 'Forbidden', component: () => import('../views/NotFoundView.vue'), meta: { public: true } },
    {
      path: '/admin', component: () => import('../layouts/AdminLayout.vue'), meta: { role: 'ADMINISTRATOR' },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', name: 'AdminDashboard', component: () => import('../views/Admin/Dashboard/AdminDashboardView.vue'), meta: { permission: 'view_admin_dashboard' } },
        { path: 'accounts', name: 'AdminAccounts', component: () => import('../views/Admin/Accounts/AdminAccountsView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'view_accounts' } },
        { path: 'accounts/create', name: 'AdminAccountCreate', component: () => import('../views/Admin/Accounts/AdminAccountCreateView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'create_account' } },
        { path: 'accounts/:id', name: 'AdminAccountDetail', component: () => import('../views/Admin/Accounts/AdminAccountDetailView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'view_accounts' } },
        { path: 'accounts/:id/edit', name: 'AdminAccountEdit', component: () => import('../views/Admin/Accounts/AdminAccountEditView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'update_account' } },
        { path: 'role-management', name: 'AdminRoleManagement', component: () => import('../views/Admin/Roles/AdminRoleManagementView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'view_roles' } },
        { path: 'role-management/create', name: 'AdminRoleCreate', component: () => import('../views/Admin/Roles/AdminRoleFormView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'create_role' } },
        { path: 'role-management/:id', name: 'AdminRoleDetail', component: () => import('../views/Admin/Roles/AdminRoleDetailView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'view_roles' } },
        { path: 'role-management/:id/edit', name: 'AdminRoleEdit', component: () => import('../views/Admin/Roles/AdminRoleFormView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'update_role' } },
        { path: 'sales-structure', name: 'AdminSalesStructure', component: () => import('../views/Admin/Roles/AdminSalesStructureView.vue'), meta: { role: 'ADMINISTRATOR', permission: 'view_sales_structure' } },
        { path: 'prospect-finder', name: 'AdminProspectFinder', component: () => import('../views/Admin/Prospect/ProspectFinderView.vue'), meta: { permission: 'view_prospect_finder' } },
        { path: 'prospects/pipeline', name: 'AdminProspectPipeline', component: () => import('../views/Admin/Prospect/ProspectPipelineView.vue'), meta: { permission: 'view_prospect_pipeline' } },
        { path: 'prospects/list', name: 'AdminProspectList', component: () => import('../views/Admin/Prospect/ProspectListView.vue'), meta: { permission: 'view_prospect_list' } },
        { path: 'prospects', name: 'AdminProspects', redirect: '/admin/prospects/pipeline' },
        { path: 'prospects/:id/review', name: 'AdminProspectReview', component: () => import('../views/Admin/Prospect/ProspectReviewView.vue'), meta: { permission: 'view_prospect_pipeline' } },
        { path: 'prospects/:id/convert', name: 'AdminProspectConvert', component: () => import('../views/Admin/Prospect/ProspectConversionView.vue'), meta: { permission: 'convert_prospect' } },
        { path: 'customers', name: 'AdminCustomers', component: () => import('../views/Admin/Customer/CustomerListView.vue'), meta: { permission: 'view_customers' } },
        { path: 'customers/add', name: 'AdminCustomerAdd', component: () => import('../views/Admin/Customer/CustomerAddView.vue'), meta: { permission: 'create_customer' } },
        { path: 'customers/:id', name: 'AdminCustomerDetail', component: () => import('../views/Admin/Customer/CustomerDetailView.vue'), meta: { permission: 'view_customers' } },
        { path: 'customers/:id/edit', name: 'AdminCustomerEdit', component: () => import('../views/Admin/Customer/CustomerEditView.vue'), meta: { permission: 'update_customer' } },
        { path: 'companies/add', name: 'AdminCompanyAdd', component: () => import('../views/Admin/Company/CompanyAddView.vue'), meta: { permission: 'create_customer' } },
        { path: 'companies/:id', name: 'AdminCompanyDetail', component: () => import('../views/Admin/Company/CompanyDetailView.vue'), meta: { permission: 'view_customers', title: 'Company Detail', description: 'Company detail interface is under construction for the first release.' } },
        { path: 'companies/:id/edit', name: 'AdminCompanyEdit', component: () => import('../views/Admin/Company/CompanyEditView.vue'), meta: { permission: 'update_customer' } },
        { path: 'visit-monitoring', name: 'AdminVisitMonitoring', component: () => import('../views/Admin/Visit/VisitMonitoringView.vue'), meta: { permission: 'view_visit_monitoring' } },
        { path: 'prospect-assignment', name: 'AdminProspectAssignment', component: () => import('../views/Admin/SimulationPlaceholderView.vue'), meta: { title: 'Prospect Assignment', description: 'Assignment is available while saving a Place in Prospect Finder; bulk reassignment is deferred.' } },
        { path: 'reports', name: 'AdminReports', component: () => import('../views/Admin/Reports/ReportsAnalyticsView.vue'), meta: { permission: 'view_reports' } },
      ],
    },
    {
      path: '/sales', component: () => import('../layouts/SalesLayout.vue'), meta: { role: 'SALES_EXECUTIVE' },
      children: [
        { path: '', redirect: '/sales/dashboard' },
        { path: 'dashboard', name: 'SalesDashboard', component: () => import('../views/Sales/Dashboard/SalesDashboardView.vue'), meta: { permission: 'view_sales_dashboard' } },
        { path: 'my-prospects', redirect: '/sales/pipeline' },
        { path: 'my-prospects/:id', name: 'SalesProspectDetail', component: () => import('../views/Sales/Prospect/ProspectDetailView.vue'), meta: { permission: 'view_my_prospect_detail' } },
        { path: 'my-prospects/:id/check-in', name: 'SalesProspectCheckIn', component: () => import('../views/Sales/Visit/CheckInView.vue'), meta: { permission: 'check_in_prospect', entityType: 'prospect' } },
        { path: 'my-prospects/:id/check-in/success', name: 'SalesProspectCheckInSuccess', component: () => import('../views/Sales/Visit/CheckInSuccessView.vue'), meta: { permission: 'check_in_prospect', entityType: 'prospect' } },
        { path: 'my-prospects/:id/visit-result', name: 'SalesProspectVisitResult', component: () => import('../views/Sales/Visit/VisitResultView.vue'), meta: { permission: 'update_visit_result', entityType: 'prospect' } },
        { path: 'my-prospects/:id/check-out', name: 'SalesProspectCheckOut', component: () => import('../views/Sales/Visit/CheckOutView.vue'), meta: { permission: 'check_out_prospect', entityType: 'prospect' } },
        { path: 'my-prospects/:id/check-out/success', name: 'SalesProspectCheckOutSuccess', component: () => import('../views/Sales/Visit/CheckOutSuccessView.vue'), meta: { permission: 'check_out_prospect', entityType: 'prospect' } },
        { path: 'pipeline', name: 'SalesPipeline', component: () => import('../views/Sales/Prospect/SalesPipelineView.vue'), meta: { permission: 'menu_sales_pipeline' } },
        { path: 'my-customers', name: 'SalesMyCustomers', component: () => import('../views/Sales/Customer/MyCustomersView.vue'), meta: { permission: 'view_my_customers' } },
        { path: 'my-customers/:id', name: 'SalesCustomerDetail', component: () => import('../views/Sales/Customer/CustomerDetailView.vue'), meta: { permission: 'view_my_customer_detail' } },
        { path: 'my-customers/:id/check-in', name: 'SalesCustomerCheckIn', component: () => import('../views/Sales/Visit/CheckInView.vue'), meta: { permission: 'check_in_customer', entityType: 'customer' } },
        { path: 'my-customers/:id/check-in/success', name: 'SalesCustomerCheckInSuccess', component: () => import('../views/Sales/Visit/CheckInSuccessView.vue'), meta: { permission: 'check_in_customer', entityType: 'customer' } },
        { path: 'my-customers/:id/visit-result', name: 'SalesCustomerVisitResult', component: () => import('../views/Sales/Visit/VisitResultView.vue'), meta: { permission: 'update_visit_result', entityType: 'customer' } },
        { path: 'my-customers/:id/check-out', name: 'SalesCustomerCheckOut', component: () => import('../views/Sales/Visit/CheckOutView.vue'), meta: { permission: 'check_out_customer', entityType: 'customer' } },
        { path: 'my-customers/:id/check-out/success', name: 'SalesCustomerCheckOutSuccess', component: () => import('../views/Sales/Visit/CheckOutSuccessView.vue'), meta: { permission: 'check_out_customer', entityType: 'customer' } },
        { path: 'history', name: 'SalesHistory', component: () => import('../views/Sales/HistoryView.vue'), meta: { permission: 'view_sales_history' } },
        { path: 'profile', name: 'SalesProfile', component: () => import('../views/Sales/ProfileView.vue'), meta: { permission: 'view_own_profile' } },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('../views/NotFoundView.vue'), meta: { public: true } },
  ],
})

router.beforeEach(async (to: RouteLocationNormalized) => {
  const auth = useAuthStore(pinia)
  await auth.bootstrap()
  const user = auth.user

  if (to.name === 'Login' && user && homeFor(user) !== FORBIDDEN_ROUTE) return homeFor(user)
  if (to.meta.public) return true
  if (!user) return { name: 'Login', query: { redirect: to.fullPath } }
  if (user.mustChangePassword && to.name !== 'ChangePassword') return { name: 'ChangePassword' }
  if (to.meta.role && !roleAllowed(to.meta.role, user.role, Boolean(user.salesRole))) return homeFor(user)
  if (to.meta.permission && !hasPermission(user, to.meta.permission)) {
    const fallback = homeFor(user)
    if (fallback !== to.fullPath && routePermitted(router, fallback, user)) return fallback
    return FORBIDDEN_ROUTE
  }
  return true
})

export default router



