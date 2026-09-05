<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Toast from 'primevue/toast'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const search = ref('')
const sidebarOpen = ref(false)
const sidebarCollapsed = ref(false)
const navbarCollapsed = ref(false)
const searchOpen = ref(false)
const debugMode = ref(false)
const isOnline = ref(navigator.onLine)
const cloudSyncing = ref(false)

const userInitials = computed(() => {
  const name = auth.user?.fullName ?? ''
  return name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part.charAt(0).toUpperCase()).join('') || 'U'
})

/* shortcut navigation removed */
/* const shortcutItems = computed(() => {
  const path = route.path
  if (path === '/admin/prospect-finder' || path === '/admin/prospects/list' || path === '/admin/visit-monitoring') return []
  let items = []

  if (path.includes('billing')) {
    items = [
      { label: 'Billing Dashboard', to: '/admin/billing', permission: 'view_billing' },
      { label: 'SO Do List', to: '/admin/billing/so-do-list', permission: 'view_billing' },
      { label: 'DO Delivered', to: '/admin/billing/do-delivered', permission: 'view_billing' },
      { label: 'Invoice List', to: '/admin/billing/invoices', permission: 'view_billing' },
      { label: 'Kwitansi List', to: '/admin/billing/receipts', permission: 'view_billing' },
      { label: 'BLR', to: '/admin/billing/blr', permission: 'view_billing' },
      { label: 'Invoice Mapping', to: '/admin/billing/invoice-mapping', permission: 'view_billing' },
      { label: 'Bank Account', to: '/admin/billing/bank-account', permission: 'view_billing' },
    ]
  } else if (path === '/admin/prospects/list') {
    // Prospect List shortcuts stay focused on prospect workflows and are
    // available directly from the navbar without opening the sidebar.
    items = [
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
      { label: 'Visit Monitoring', to: '/admin/visit-monitoring', permission: 'view_visit_monitoring' },
    ]
  } else if (path.includes('/prospects')) {
    items = [
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
    ]
  } else if (path.includes('/customers')) {
    items = [
      { label: 'Customer Site', to: { path: '/admin/customers', query: { tab: 'site' } }, permission: 'view_customers' },
      { label: 'Company', to: { path: '/admin/customers', query: { tab: 'company' } }, permission: 'view_customers' },
      { label: 'Master Data', to: { path: '/admin/customers', query: { tab: 'master' } }, permission: 'view_customers' },
      { label: 'Add Segment', to: { path: '/admin/customers', query: { tab: 'master', action: 'add-segment' } }, permission: 'view_customers' },
      { label: 'Add Category', to: { path: '/admin/customers', query: { tab: 'master', action: 'add-category' } }, permission: 'view_customers' },
      { label: 'Master Data Trash', to: { path: '/admin/customers', query: { tab: 'master', action: 'trash' } }, permission: 'view_customers' },
      { label: 'Create Customer', to: '/admin/customers/add', permission: 'view_customers' },
    ]
  } else if (path.includes('/accounts')) {
    items = [
      { label: 'Employee List', to: '/admin/accounts', permission: 'view_accounts' },
      { label: 'Create Account', to: '/admin/accounts/create', permission: 'view_accounts' },
      { label: 'Trash', to: { path: '/admin/accounts', query: { view: 'trash' } }, permission: 'view_accounts' },
      { label: 'Active Accounts', to: { path: '/admin/accounts', query: { status: 'ACTIVE' } }, permission: 'view_accounts' },
      { label: 'Inactive Accounts', to: { path: '/admin/accounts', query: { status: 'INACTIVE' } }, permission: 'view_accounts' },
    ]
  } else if (path === '/admin/role-management' || path.startsWith('/admin/role-management/')) {
    items = [
      { label: 'Role Management', to: '/admin/role-management', permission: 'view_roles' },
      { label: 'Add Role', to: '/admin/role-management/create', permission: 'view_roles' },
      { label: 'Active Roles', to: { path: '/admin/role-management', query: { status: 'ACTIVE' } }, permission: 'view_roles' },
      { label: 'Inactive Roles', to: { path: '/admin/role-management', query: { status: 'INACTIVE' } }, permission: 'view_roles' },
      { label: 'Sales Structure', to: '/admin/sales-structure', permission: 'view_sales_structure' },
      { label: 'Employee Management', to: '/admin/accounts', permission: 'view_accounts' },
    ]
  } else if (path === '/admin/sales-structure' || path.startsWith('/admin/sales-structure/')) {
    items = [
      { label: 'Sales Structure', to: '/admin/sales-structure', permission: 'view_sales_structure' },
      { label: 'Role Management', to: '/admin/role-management', permission: 'view_roles' },
      { label: 'Employee Management', to: '/admin/accounts', permission: 'view_accounts' },
    ]
  } else if (path === '/admin/visit-monitoring') {
    items = [
      { label: 'Visit Monitoring', to: '/admin/visit-monitoring', permission: 'view_visit_monitoring' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Customer Existing', to: '/admin/customers', permission: 'view_customers' },
    ]
  } else if (path.includes('/visit-monitoring')) {
    items = [
      { label: 'Visit Monitoring', to: '/admin/visit-monitoring', permission: 'view_visit_monitoring' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Customer Existing', to: '/admin/customers', permission: 'view_customers' },
    ]
  } else if (path.includes('/reports')) {
    items = [
      { label: 'Dashboard', to: '/admin/dashboard', permission: 'view_admin_dashboard' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
      { label: 'Visit Monitoring', to: '/admin/visit-monitoring', permission: 'view_visit_monitoring' },
    ]
  } else {
    items = [
      { label: 'Customer List', to: '/admin/customers', permission: 'view_customers' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
    ]
  }

  return items.filter((item) => auth.hasPermission(item.permission))
}) */

const subPage = computed(() => {
  const path = route.path
  if (path.includes('/accounts/create') || path.includes('/accounts/') && path !== '/admin/accounts') {
    return { back: 'Employee Management', backTo: '/admin/accounts', title: path.includes('/edit') ? 'Edit Employee' : path.includes('/create') ? 'Create Employee' : 'Employee Detail', crumb: 'Employee Management > Employee' }
  }
  if (path.includes('/prospects/') && (path.endsWith('/review') || path.endsWith('/convert'))) {
    return { back: 'Prospect List', backTo: '/admin/prospects/list', title: path.endsWith('/convert') ? 'Convert Prospect' : 'Prospect Review', crumb: 'Prospect Management > Prospect' }
  }
  if (path.includes('/customers/add') || path.includes('/customers/')) {
    return { back: 'Customer List', backTo: '/admin/customers', title: path.includes('/edit') ? 'Edit Customer' : 'Add Customer', crumb: 'Customer Management > Customer' }
  }
  if (path.includes('/companies/add') || path.includes('/companies/')) {
    return { back: 'Customer List', backTo: '/admin/customers', title: path.includes('/edit') ? 'Edit Company' : 'Add Company', crumb: 'Customer Management > Company' }
  }
  const mainPages = ['/admin/dashboard', '/admin/accounts', '/admin/customers', '/admin/prospect-finder', '/admin/prospects/list', '/admin/prospects/pipeline', '/admin/visit-monitoring', '/admin/reports', '/admin/api-usage', '/admin/role-management', '/admin/sales-structure']
  if (!mainPages.includes(path)) {
    if (path.includes('/role-management/')) return { back: 'Role Management', backTo: '/admin/role-management', title: 'Role Detail', crumb: 'Management > Role' }
    if (path.includes('/prospects/')) return { back: 'Prospect List', backTo: '/admin/prospects/list', title: 'Prospect Detail', crumb: 'Prospect Management > Prospect' }
    if (path.includes('/accounts/')) return { back: 'Employee Management', backTo: '/admin/accounts', title: 'Employee Detail', crumb: 'Management > Employee' }
    return { back: 'Dashboard', backTo: '/admin/dashboard', title: String(route.meta.title || route.name || 'Page'), crumb: 'Admin Management' }
  }
  return null
})

/* function isShortcutActive(item: { to: string | { path: string; query?: Record<string, string | undefined> } }) {
  if (typeof item.to === 'string') return route.path === item.to || route.path.startsWith(`${item.to}/`)
  return route.path === item.to.path && (route.query.tab || 'site') === item.to.query?.tab
}

function hasShortcutSeparator(label: string) {
  return ['Credit Limit', 'Daily Collect.', 'Pay. Hist. Summary', 'BLR', 'Invoice List', 'BLR List'].includes(label)
} */

const sidebarWidth = computed(() => sidebarCollapsed.value ? '80px' : '220px')

function runSearch() {
  const value = search.value.trim()
  if (route.path.includes('/prospects/list')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else if (route.path.includes('/accounts')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else if (route.path.includes('/role-management')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else if (route.path.includes('/sales-structure')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else if (route.path.includes('/visit-monitoring')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else {
    router.push({ path: '/admin/prospects/list', query: value ? { search: value } : {} })
  }
  searchOpen.value = false
}
function toggleSearch() { searchOpen.value = !searchOpen.value }
function toggleNavbar() {
  navbarCollapsed.value = !navbarCollapsed.value
  if (navbarCollapsed.value) searchOpen.value = false
}
function toggleDebug() { debugMode.value = !debugMode.value }
function closeDebug() { debugMode.value = false }
function syncCloud() {
  if (cloudSyncing.value) return
  cloudSyncing.value = true
  window.setTimeout(() => { cloudSyncing.value = false }, 900)
}
async function toggleFullscreen() {
  if (document.fullscreenElement) await document.exitFullscreen()
  else await document.documentElement.requestFullscreen()
}
function closeSidebar() { sidebarOpen.value = false }
function toggleCollapse() { sidebarCollapsed.value = !sidebarCollapsed.value }

async function logout() {
  closeSidebar()
  await auth.logout()
  await router.replace('/login')
}
</script>

<template>
  <div class="admin-shell bg-surface" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <div v-if="sidebarOpen" class="mobile-backdrop" @click="closeSidebar" />
    <aside class="admin-sidebar" :class="{ 'sidebar-open': sidebarOpen, collapsed: sidebarCollapsed }" :style="{ width: sidebarWidth }">
      <div class="sidebar-header">
        <div class="shell-logo">
          <span class="logo-mark">
            <img src="/yummy-logo.png" alt="Yummy Dairy" />
          </span>
          <div v-show="!sidebarCollapsed" class="logo-text">Yummy Food<small>Field Sales CRM</small></div>
        </div>
        <button class="collapse-btn" @click="toggleCollapse" :title="sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'">
          <i class="pi" :class="sidebarCollapsed ? 'pi-chevron-right' : 'pi-chevron-left'" />
        </button>
      </div>
      <small v-show="!sidebarCollapsed" class="nav-caption">MAIN</small>
      <nav class="min-h-0 flex-1 overflow-x-hidden overflow-y-auto overscroll-y-contain py-4 px-1" aria-label="Administrator navigation">
        <div class="space-y-4 pb-2">
          <div>
            <p v-show="!sidebarCollapsed" class="sidebar-subheader px-3 font-semibold uppercase tracking-wide text-slate-400">
              Pages
            </p>
            <ul class="mt-1 space-y-1">
              <li v-if="auth.hasPermission('view_admin_dashboard')" class="relative">
                <RouterLink
                  to="/admin/dashboard"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-home"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Dashboard</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_customers')" class="relative">
                <RouterLink
                  to="/admin/customers"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-users"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Customer List</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_prospect_finder')" class="relative">
                <RouterLink
                  to="/admin/prospect-finder"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-compass"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Prospect Finder</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_prospect_list')" class="relative">
                <RouterLink
                  to="/admin/prospects/list"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-list"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Customer Prospect</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_prospect_pipeline')" class="relative">
                <RouterLink
                  to="/admin/prospects/pipeline"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-th-large"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Prospect Pipeline</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_visit_monitoring')" class="relative">
                <RouterLink
                  to="/admin/visit-monitoring"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-map-marker"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Visit Monitoring</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_accounts')" class="relative">
                <RouterLink
                  to="/admin/accounts"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-user-edit"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Employee Management</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_roles')" class="relative">
                <RouterLink
                  to="/admin/role-management"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-shield"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Role Management</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_sales_structure')" class="relative">
                <RouterLink
                  to="/admin/sales-structure"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-sitemap"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Sales Structure</span>
                </RouterLink>
              </li>

              <li v-if="auth.hasPermission('view_reports')" class="relative">
                <RouterLink
                  to="/admin/reports"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-chart-bar"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Reports</span>
                </RouterLink>
              </li>

              <li class="relative">
                <RouterLink
                  to="/admin/api-usage"
                  class="flex rounded-lg px-2.5 font-semibold transition hover:bg-primary-50 hover:text-primary-700 text-slate-600"
                  :class="sidebarCollapsed ? 'flex-col items-center justify-center gap-[2px] text-center px-0 py-[3px] sidebar-compact' : 'flex-row items-center gap-1.5 px-2 py-1 sidebar-regular'"
                  active-class="bg-primary-50 text-primary-700 shadow-inner"
                  @click="closeSidebar"
                >
                  <span class="text-[15px] pi pi-chart-line"></span>
                  <span class="sidebar-label" :class="{ 'sidebar-label-compact text-center': sidebarCollapsed }">Monitoring API</span>
                </RouterLink>
              </li>
            </ul>
          </div>
        </div>
      </nav>
      <details class="sidebar-profile">
        <summary>
          <span class="sidebar-avatar"><span>{{ userInitials }}</span><i class="profile-online-dot" /></span>
          <span class="sidebar-profile-info"><strong>{{ auth.user?.fullName }}</strong><small><i class="pi pi-shield" /> Administrator</small></span>
          <i class="pi pi-chevron-up profile-chevron" />
        </summary>
        <div class="sidebar-profile-menu"><div class="profile-menu-heading"><span>ACCOUNT</span><strong>Siap bekerja hari ini?</strong></div><button class="signout-btn" @click="logout"><span class="logout-icon"><i class="pi pi-sign-out" /></span><span><strong>Keluar dari akun</strong><small>Amankan sesi Anda</small></span><i class="pi pi-arrow-right logout-arrow" /></button></div>
      </details>
    </aside>
    <div class="admin-workspace">
      <Toast position="top-right" />
      <header class="app-navbar flex h-[60px] shrink-0 items-center justify-between border-b border-slate-200 bg-white/90 px-3 py-2 text-navbar shadow-sm backdrop-blur sm:px-4">
        <!-- Left Side: Mobile Menu Button, Mobile Title, Desktop Quick Search Button -->
        <div class="ms-1 flex items-center gap-2">
          <button
            type="button"
            class="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-[13px] text-slate-700 transition hover:bg-slate-50 lg:hidden"
            aria-label="Open menu"
            @click="sidebarOpen = !sidebarOpen"
          >
            <span class="pi pi-th-large"></span>
          </button>

          <div class="lg:hidden">
            <p class="font-semibold uppercase tracking-wide text-slate-400 leading-none" style="font-size: calc(var(--font-navbar, 14px) * 0.65);">
              ERP
            </p>
            <h1 class="text-slate-800 leading-tight font-semibold" style="font-size: clamp(11px, 2.6vw, calc(var(--font-heading, 16px) * 0.86));">
              {{ String(route.meta.title || route.name || 'Dashboard') }}
            </h1>
          </div>

          <button
            class="hidden h-8 items-center gap-1.5 rounded-full border border-slate-200 bg-white/80 px-3 text-[12px] font-semibold text-slate-700 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 lg:flex"
            type="button"
            @click="toggleSearch"
          >
            <span class="pi pi-search text-[12px] text-primary-600"></span>
            <span>Quick navigation</span>
            <span class="flex items-center gap-0.5 rounded-md border border-slate-200 bg-white px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-slate-500">
              <span class="hidden sm:inline">Ctrl</span>
              <span class="hidden sm:inline text-slate-400">/</span>
              <span>⌘</span>
              <span>+</span>
              <span>K</span>
            </span>
          </button>

          <!-- Embedded Global Search Input when toggled -->
          <form v-if="searchOpen" class="global-search is-open" @submit.prevent="runSearch">
            <i class="pi pi-search" />
            <input v-model="search" autofocus aria-label="Search prospects" placeholder="Search prospects, customers..." />
            <button type="button" class="search-close" aria-label="Clear search" title="Clear search" @click="search = ''">×</button>
            <button type="submit">Enter</button>
          </form>
        </div>

        <!-- Middle Spacer -->
        <div class="flex items-center"></div>

        <!-- Right Side Action Buttons -->
        <div class="flex items-center gap-1.5">
          <button
            class="flex h-8 items-center justify-center gap-1.5 rounded-full border px-2.5 text-[12px] font-semibold shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition border-slate-200 bg-white/80 text-slate-600 hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 hidden lg:flex"
            :class="{ '!border-primary-400 !bg-primary-50 !text-primary-700': debugMode }"
            type="button"
            title="Debug Log"
            @click="toggleDebug"
          >
            <span class="pi pi-server text-[12px]"></span>
            <span class="hidden lg:inline">Debug Log</span>
          </button>

          <button
            class="hidden h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white/80 text-[12px] text-slate-600 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 lg:flex"
            type="button"
            :title="sidebarCollapsed ? 'Expand sidebar' : 'Compact mode'"
            @click="toggleCollapse"
          >
            <span class="pi text-[12px]" :class="sidebarCollapsed ? 'pi-angle-right' : 'pi-angle-left'"></span>
          </button>

          <button
            class="hidden h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white/80 text-[12px] text-slate-600 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 lg:flex"
            type="button"
            title="Fullscreen"
            @click="toggleFullscreen"
          >
            <span class="pi pi-window-maximize"></span>
          </button>

          <RouterLink
            to="/admin/reports"
            class="relative flex h-8 w-8 items-center justify-center rounded-full border text-[12px] shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition border-slate-200 bg-white/80 text-slate-600 hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700"
            title="Chat"
            aria-label="Open chat"
          >
            <span class="pi pi-comments text-[12px]"></span>
          </RouterLink>

          <!-- System Online Status Button -->
          <div class="relative">
            <button
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white/80 text-[12px] text-slate-600 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 animate-fade"
              :class="isOnline ? '!text-green-500 !border-green-200 !bg-green-50/80' : '!text-red-500 !border-red-200 !bg-red-50/80'"
              :title="isOnline ? 'Connected' : 'Disconnected'"
            >
              <i class="pi text-[12px]" :class="isOnline ? 'pi-check-circle' : 'pi-exclamation-circle'"></i>
            </button>
          </div>

          <!-- Cloud Sync Button -->
          <button
            type="button"
            class="relative flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-700 shadow-sm transition hover:border-slate-300 hover:shadow"
            :title="cloudSyncing ? 'Syncing...' : 'Sync Cloud'"
            @click="syncCloud"
          >
            <span class="pi text-[12px] text-sky-600" :class="cloudSyncing ? 'pi-spin pi-spinner' : 'pi-cloud-upload'"></span>
          </button>
        </div>
      </header>
      <div v-if="debugMode" class="debug-overlay" @click.self="closeDebug">
        <section class="debug-modal" role="dialog" aria-modal="true" aria-labelledby="debug-title">
          <header><div><span class="debug-kicker">SYSTEM DIAGNOSTICS</span><h2 id="debug-title">Debug Console</h2></div><button type="button" aria-label="Close debug" @click="closeDebug">×</button></header>
          <div class="debug-status"><i class="pi pi-check-circle" /><div><strong>No errors detected</strong><span>Frontend is running normally.</span></div></div>
          <div class="debug-detail"><span>Current route</span><code>{{ route.fullPath }}</code></div>
          <div class="debug-detail"><span>Network</span><strong :class="isOnline ? 'online' : 'offline'">{{ isOnline ? 'Online' : 'Offline' }}</strong></div>
          <footer><button type="button" @click="closeDebug">Close</button></footer>
        </section>
      </div>
      <div v-if="subPage && !route.path.endsWith('/create') && !route.path.includes('/role-management/') && !route.path.includes('/accounts/') && !route.path.endsWith('/add')" class="subpage-bar">
        <RouterLink :to="subPage.backTo" class="subpage-back">← {{ subPage.back }}</RouterLink>
        <span class="subpage-divider" />
        <strong>{{ subPage.title }}</strong>
        <small>{{ subPage.crumb }} &gt; {{ subPage.title.replace('Create ', '').replace('Add ', '') }}</small>
      </div>
      <main class="admin-content"><RouterView /></main>
    </div>
  </div>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   ADMIN SHELL — refined, professional sidebar (structure unchanged)
   ════════════════════════════════════════════════════════════════ */

/* ── Admin Shell ─────────────────────────────────────────────── */
.admin-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  /*background: #ffffff;*/
  transition: grid-template-columns 0.25s ease;
}

.admin-shell.sidebar-collapsed {
  grid-template-columns: 80px minmax(0, 1fr);
}

.admin-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  color: #1e293b;
  background: #ffffff;
  border-right: 1px solid #e2e8f0;
  overflow-x: hidden;
  overflow-y: hidden;
  transition: width 0.25s ease;
}

.admin-sidebar.collapsed {
  padding: 0;
  align-items: center;
}

.shell-logo {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.15rem 0.35rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  font-size: 0.85rem;
  min-height: 2rem;
}

.admin-sidebar.collapsed .shell-logo {
  padding: 0;
  justify-content: center;
}

.logo-mark {
  width: 3.1rem;
  height: 2rem;
  display: inline-grid;
  place-items: center;
  padding: 0.2rem;
  border-radius: 10px;
  background: #fff;
  border: 1px solid #edf1f6;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.05);
  flex-shrink: 0;
}

.logo-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.logo-text { display: grid; color: #0f172a; }
.logo-text small {
  color: #94a3b8;
  font-size: 0.55rem;
  font-weight: 550;
  line-height: 1.35;
}

/* ── Sidebar Header ─────────────────────────────────────────── */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  height: 56px;
  padding: 0 12px;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 0;
  position: relative;
}

.admin-sidebar.collapsed .sidebar-header {
  justify-content: center;
  padding: 0;
}

/* ── Collapse Toggle ────────────────────────────────────────── */
.collapse-btn {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 1px solid #edf1f6;
  border-radius: 999px;
  background: #ffffff;
  color: #94a3b8;
  cursor: pointer;
  position: absolute;
  right: -15px;
  top: 11px;
  z-index: 2;
  font-size: 0.55rem;
  flex-shrink: 0;
  transition: background 160ms ease, color 160ms ease, border-color 160ms ease;
}

.collapse-btn:hover {
  background: #e63946;
  color: #fff;
  border-color: #e63946;
}

.nav-caption {
  padding: 0.35rem 0.5rem 0.15rem;
  color: #a3adba;
  font-size: 0.6rem;
  font-weight: 750;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  white-space: nowrap;
  overflow: hidden;
}

.admin-sidebar nav .nav-caption {
  margin-top: 0.6rem;
  padding-top: 0.75rem;
  border-top: 1px solid #f1f4f8;
}

.admin-sidebar nav {
  display: grid;
  gap: 2px;
  flex: 1;
  min-height: 0;
  width: 100%;
  align-content: start;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 14px 10px 0;
}

.admin-sidebar nav::-webkit-scrollbar { width: 4px; }
.admin-sidebar nav::-webkit-scrollbar-track { background: transparent; }
.admin-sidebar nav::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }

.admin-sidebar nav a,
.nav-placeholder {
  display: flex;
  gap: 0.65rem;
  align-items: center;
  height: 32px;
  padding: 0 10px;
  border-radius: 8px;
  color: #64748b;
  text-decoration: none;
  font-size: 0.6875rem;
  font-weight: 550;
  white-space: nowrap;
  flex-shrink: 0;
  transition: color 150ms ease, background 150ms ease, padding 150ms ease;
}

.admin-sidebar.collapsed nav a {
  justify-content: center;
  flex-direction: column;
  gap: 2px;
  height: 42px;
  padding: 4px;
  border-radius: 8px;
}

.admin-sidebar nav a:hover {
  color: #e63946;
  background: #f4f7fc;
}

.admin-sidebar.collapsed nav a:hover {
  background: #fff0f1;
}

.admin-sidebar nav a.router-link-active {
  color: #0f172a;
  background: #fef2f2;
  font-weight: 700;
  box-shadow: inset 3px 0 0 #e63946;
}

.admin-sidebar.collapsed nav a.router-link-active {
  box-shadow: none;
  position: relative;
}

.admin-sidebar.collapsed nav a.router-link-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 1.2rem;
  background: #e63946;
  border-radius: 0 3px 3px 0;
}

.admin-sidebar nav i {
  width: 1.1rem;
  height: 1.1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  font-size: 0.78rem;
  flex-shrink: 0;
  color: #64748b;
  transition: color 150ms ease;
}

.admin-sidebar nav a:hover i { color: #e63946; }
.admin-sidebar nav a.router-link-active i { color: #e63946; }

.nav-placeholder { opacity: 0.5; }

.sidebar-note {
  padding: 0.85rem 0.8rem;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.3rem 0.6rem;
  color: #fff;
  background: linear-gradient(135deg, #16213b 0%, #1e293b 100%);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  box-shadow: 0 8px 20px -8px rgba(15, 23, 42, 0.5);
}

.sidebar-note > i {
  grid-row: 1 / 3;
  width: 1.85rem;
  height: 1.85rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, #e63946, #ef4e5d);
  border-radius: 9px;
  box-shadow: 0 4px 10px -2px rgba(230, 57, 70, 0.5);
}

.sidebar-note strong { font-size: 0.66rem; letter-spacing: -0.01em; }
.sidebar-note span {
  font-size: 0.55rem;
  line-height: 1.5;
  color: #9aa5b7;
}

/* ── Topbar ──────────────────────────────────────────────────── */
.admin-workspace {
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.admin-topbar {
  flex-shrink: 0;
  height: 56px;
  padding: 0 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.65rem;
  background: #ffffff;
  border-bottom: 1px solid #edf1f6;
  z-index: 10;
}

.global-search {
  width: min(520px, 55vw);
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.65rem;
  color: #a3adba;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  transition: border-color 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

.global-search:focus-within {
  background: #ffffff;
  border-color: #e63946;
  box-shadow: 0 0 0 4px rgba(230, 57, 70, 0.09);
}

.global-search input {
  width: 100%;
  border: 0;
  outline: 0;
  color: #0f172a;
  background: transparent;
  font-size: 0.72rem;
}

.global-search input::placeholder { color: #a3adba; }

.global-search button {
  padding: 0.25rem 0.5rem;
  border: 1px solid #e2e8f0;
  color: #94a3b8;
  background: #ffffff;
  border-radius: 6px;
  font-size: 0.52rem;
  cursor: pointer;
  transition: background 150ms ease;
}

.global-search button:hover { background: #f1f5f9; }
.global-search .search-close { padding: 0 .25rem; border: 0; background: transparent; color: #94a3b8; font-size: 1rem; line-height: 1; }
.global-search .search-close:hover { color: #dc2626; background: transparent; }
.debug-overlay { position: fixed; inset: 0; z-index: 200; display: grid; place-items: center; padding: 1rem; background: rgba(15,23,42,.38); backdrop-filter: blur(3px); }
.debug-modal { width: min(430px, 100%); overflow: hidden; border: 1px solid #e2e8f0; border-radius: 16px; background: #fff; box-shadow: 0 24px 70px rgba(15,23,42,.25); }
.debug-modal header { display:flex; align-items:center; justify-content:space-between; padding: 1.1rem 1.25rem; border-bottom: 1px solid #edf1f6; }
.debug-modal header h2 { margin: .2rem 0 0; color:#172033; font-size:1rem; }
.debug-kicker { color:#94a3b8; font-size:.55rem; font-weight:800; letter-spacing:.12em; }
.debug-modal header button { border:0; background:transparent; color:#64748b; font-size:1.4rem; cursor:pointer; }
.debug-status { display:flex; gap:.7rem; align-items:center; margin:1.1rem 1.25rem; padding:.85rem; border:1px solid #bbf7d0; border-radius:10px; background:#f0fdf4; color:#15803d; }
.debug-status i { font-size:1.2rem; }.debug-status div { display:grid; gap:.18rem; }.debug-status span { color:#64748b; font-size:.68rem; }
.debug-detail { display:flex; justify-content:space-between; gap:1rem; padding:.7rem 1.25rem; border-top:1px solid #f1f5f9; color:#64748b; font-size:.7rem; }.debug-detail code { color:#334155; font-size:.65rem; }.debug-detail .online { color:#16a34a; }.debug-detail .offline { color:#dc2626; }
.debug-modal footer { display:flex; justify-content:flex-end; padding:.9rem 1.25rem; background:#f8fafc; }.debug-modal footer button { padding:.45rem .8rem; border:1px solid #dbe3ee; border-radius:7px; background:#fff; color:#475569; cursor:pointer; }

.topbar-spacer { flex: 1; }

.profile-menu { position: relative; }

.profile-menu summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  list-style: none;
  cursor: pointer;
  padding: 0.25rem 0.4rem;
  border-radius: 10px;
  transition: background 150ms ease;
}

.profile-menu summary:hover { background: #f8fafc; }

.profile-menu summary > .avatar-initials {
  width: 2rem;
  height: 2rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, #e63946, #d62839);
  border-radius: 50%;
  font-size: 0.72rem;
  font-weight: 800;
  flex-shrink: 0;
  box-shadow: 0 3px 8px -2px rgba(230, 57, 70, 0.45);
}

.profile-menu summary .profile-info { display: grid; text-align: left; }
.profile-menu summary .profile-info strong { font-size: 0.68rem; color: #0f172a; }
.profile-menu summary .profile-info small { color: #94a3b8; font-size: 0.55rem; }

.profile-menu summary > i {
  font-size: 0.55rem;
  color: #a3adba;
  transition: transform 150ms ease;
}

.profile-menu[open] summary > i { transform: rotate(180deg); }

.profile-dropdown {
  position: absolute;
  z-index: 50;
  top: calc(100% + 6px);
  right: 0;
  min-width: 180px;
  padding: 0.35rem;
  background: #ffffff;
  border: 1px solid #edf1f6;
  border-radius: 12px;
  box-shadow: 0 12px 28px -10px rgba(15, 23, 42, 0.22);
}

.signout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 0.75rem;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: #dc2626;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 150ms ease, color 150ms ease;
}

.signout-btn i { font-size: 0.8rem; }

.signout-btn:hover {
  background: #fef2f2;
  color: #b91c1c;
}

.admin-content {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
  overflow-x: hidden;
}
.admin-content:has(.compact-admin-page) {
  padding: 0.75rem 1rem 1rem;
}
.admin-content:has(.admin-page),
.admin-content:has(.finder-page) {
  padding: 0;
}
.admin-content:has(.admin-page) > .admin-page,
.admin-content:has(.finder-page) > .finder-page {
  width: 100%;
  max-width: none;
  min-height: 100%;
  box-sizing: border-box;
}
@media (min-width: 901px) {
  .admin-content :deep(.p-button) {
    min-height: 34px;
    padding: .45rem .8rem;
    border-radius: 8px;
    font-size: .68rem;
    font-weight: 650;
    box-shadow: none;
    transition: transform 160ms ease, box-shadow 160ms ease, background 160ms ease;
  }
  .admin-content :deep(.p-button:not(.p-button-outlined):not(.p-button-text)) {
    border-color: #d62839;
    background: #d62839;
    color: #fff;
    box-shadow: 0 3px 8px rgba(214,40,57,.16);
  }
  .admin-content :deep(.p-button:not(.p-button-text):hover) { transform: translateY(-1px); box-shadow: 0 5px 12px rgba(15,23,42,.1); }
  .admin-content :deep(.p-button.p-button-outlined) { border-color: #d8e1ec; background: #fff; color: #475569; }
  .admin-content :deep(.p-button.p-button-outlined:hover) { border-color: #d62839; background: #fff7f7; color: #b4232d; }
  .admin-content :deep(.p-button.p-button-text) { color: #64748b; }
  .admin-content :deep(.p-button.p-button-danger) { border-color: #dc2626; background: #dc2626; }
  .admin-content :deep(.filter-panel) { gap: .55rem; padding: .6rem .7rem; }
  .admin-content :deep(.filter-field) { gap: .2rem; }
  .admin-content :deep(.filter-field label) { font-size: .55rem; letter-spacing: .06em; }
  .admin-content :deep(.filter-field .p-select),
  .admin-content :deep(.filter-field .p-inputtext) { min-height: 36px; border-radius: 8px; }
  .admin-content :deep(.filter-grid) { gap: .5rem; }
}

/* ── Mobile Hamburger ──────────────────────────────────────── */
.mobile-backdrop {
  display: none; position: fixed; inset: 0; z-index: 90;
  background: rgba(15, 23, 42, 0.45); backdrop-filter: blur(2px);
}

.hamburger-btn {
  display: none;
  align-items: center; justify-content: center;
  width: 36px; height: 36px; flex-shrink: 0;
  border: 1px solid #edf1f6; border-radius: 9px;
  background: #ffffff; color: #1e293b;
  cursor: pointer; font-size: 1rem;
  transition: background 150ms ease;
}
.hamburger-btn:hover { background: #f8fafc; }

.topbar-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border: 1px solid #edf1f6;
  border-radius: 9px;
  background: #ffffff;
  color: #475569;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 150ms ease, color 150ms ease, border-color 150ms ease;
}

.topbar-icon-btn:hover {
  background: #f8fafc;
  color: #e63946;
  border-color: #e2e8f0;
}

.ig-link:hover {
  background: linear-gradient(135deg, #fce18a, #ff5c87, #d942ff);
  color: #fff;
  border-color: transparent;
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 900px) {
  .admin-shell { grid-template-columns: 1fr; }
  .admin-shell.sidebar-collapsed { grid-template-columns: 1fr; }
  .admin-sidebar {
    position: fixed; top: 0; left: 0; z-index: 100;
    width: 260px; height: 100vh;
    transform: translateX(-100%);
    transition: transform 0.25s ease;
    display: flex;
    align-items: stretch;
    padding: 1.1rem 0.7rem;
    box-shadow: 12px 0 32px -12px rgba(15, 23, 42, 0.25);
  }
  .admin-sidebar.sidebar-open { transform: translateX(0); }
  .admin-sidebar.collapsed { width: 260px; align-items: stretch; padding: 1.1rem 0.7rem; }
  .admin-sidebar.collapsed .shell-logo { padding: 0.15rem 0.35rem; justify-content: flex-start; }
  .admin-sidebar.collapsed .logo-text { display: grid !important; }
  .admin-sidebar.collapsed nav a { justify-content: flex-start; padding: 0.55rem 0.65rem; }
  .admin-sidebar.collapsed nav a span { display: inline !important; }
  .admin-sidebar.collapsed nav a.router-link-active::before { display: none; }
  .admin-sidebar.collapsed nav a.router-link-active { box-shadow: inset 3px 0 0 #e63946; }
  .admin-sidebar.collapsed .nav-caption { display: block !important; }
  .admin-sidebar.collapsed .sidebar-note { display: grid !important; }
  .collapse-btn { display: none; }
  .mobile-backdrop { display: block; }
  .hamburger-btn { display: flex; }
  .admin-topbar { padding: 0 0.8rem; }
  .admin-content { padding: 0.8rem; }
  .profile-menu summary div { display: none; }
}

@media (max-width: 560px) {
  .global-search { width: 100%; }
  .profile-menu { display: none; }
}
.erp-icon { width: 15px; height: 15px; display: block; }
.sidebar-profile { position: relative; margin-top: auto; border-top: 1px solid #edf1f6; padding-top: .65rem; }
.sidebar-profile summary { display:flex; align-items:center; gap:.55rem; padding:.45rem .4rem; border-radius:9px; list-style:none; cursor:pointer; }
.sidebar-profile summary:hover { background:#f8fafc; }.sidebar-profile summary::-webkit-details-marker { display:none; }
.sidebar-avatar { display:grid; place-items:center; width:30px; height:30px; flex:none; border-radius:50%; background:linear-gradient(135deg,#e63946,#c92332); color:#fff; font-size:.68rem; font-weight:800; }
.sidebar-profile-info { display:grid; min-width:0; flex:1; gap:.1rem; }.sidebar-profile-info strong { overflow:hidden; color:#172033; font-size:.63rem; text-overflow:ellipsis; white-space:nowrap; }.sidebar-profile-info small { color:#94a3b8; font-size:.52rem; }.sidebar-profile summary > i { color:#94a3b8; font-size:.7rem; }
.sidebar-profile-menu { position:absolute; right:.35rem; bottom:calc(100% + .35rem); width:170px; padding:.3rem; border:1px solid #e5eaf0; border-radius:10px; background:#fff; box-shadow:0 12px 25px rgba(15,23,42,.14); }
.sidebar-profile { padding-top: .8rem; }
.sidebar-profile summary { padding: .55rem .5rem; border: 1px solid transparent; transition: .2s ease; }
.sidebar-profile summary:hover { background: linear-gradient(135deg, #fff7f7, #fff); border-color: #f4d5d8; }
.sidebar-avatar { position: relative; width: 36px; height: 36px; background: linear-gradient(135deg, #e63946, #8f1d2a); box-shadow: 0 5px 12px rgba(230,57,70,.22); }
.profile-online-dot { position: absolute; right: -1px; bottom: -1px; width: 9px; height: 9px; border: 2px solid #fff; border-radius: 50%; background: #22c55e; }
.sidebar-profile-info small { display: flex; align-items: center; gap: .25rem; }
.sidebar-profile-info small i { font-size: .5rem; color: #e63946; }
.profile-chevron { transition: transform .2s ease; }.sidebar-profile[open] .profile-chevron { transform: rotate(180deg); }
.sidebar-profile-menu { width: 218px; padding: .65rem; border-radius: 14px; }
.profile-menu-heading { display: grid; gap: .2rem; padding: .25rem .45rem .6rem; }.profile-menu-heading span { color: #e63946; font-size: .48rem; font-weight: 800; letter-spacing: .12em; }.profile-menu-heading strong { color: #172033; font-size: .65rem; }
.signout-btn { display: flex; align-items: center; gap: .55rem; width: 100%; padding: .55rem .45rem; border: 0; border-radius: 10px; background: #fff5f5; color: #b4232d; cursor: pointer; text-align: left; }.signout-btn:hover { background: #fee2e2; }.signout-btn > span:nth-child(2) { display: grid; gap: .12rem; flex: 1; }.signout-btn strong { font-size: .63rem; }.signout-btn small { color: #c26a72; font-size: .5rem; }.logout-icon { display: grid; place-items: center; width: 27px; height: 27px; border-radius: 8px; background: #fff; }.logout-arrow { font-size: .62rem; }
.topbar-actions { display: flex; align-items: center; gap: .45rem; margin-left: auto; }
.topbar-action-btn {
  height: 32px; display: inline-flex; align-items: center; gap: .4rem; padding: 0 .78rem;
  border: 1px solid #d7e0eb; border-radius: 999px; background: linear-gradient(#fff, #fbfcfe); color: #334155;
  font-size: .68rem; font-weight: 600; cursor: pointer; box-shadow: 0 1px 2px rgba(15,23,42,.04);
}
.topbar-action-btn:hover, .topbar-action-btn.active { color: #b4232d; border-color: #e8aeb4; background: #fff5f5; }
.topbar-action-btn i { display: grid; place-items: center; width: 16px; height: 16px; color: #526782; font-size: .72rem; }
.topbar-action-btn:hover i, .topbar-action-btn.active i { color: #d62839; }
.topbar-icon-btn { box-shadow: 0 1px 2px rgba(15,23,42,.04); }
.topbar-icon-btn.status-ok { color: #10b981; border-color: #aee9ca; background: #f2fff8; box-shadow: 0 0 0 3px rgba(16,185,129,.06); }
.topbar-icon-btn.status-ok:hover { color: #059669; border-color: #6ed9a5; background: #eafff3; }
.topbar-icon-btn.status-ok.offline { color: #dc2626; border-color: #fecaca; background: #fff5f5; }
.cloud-btn.syncing { color: #2563eb; border-color: #bfdbfe; background: #eff6ff; }
.topbar-icon-btn small { min-width: 12px; font-size: .5rem; margin-left: .12rem; }
.subpage-bar {
  min-height: 44px;
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0 1.5rem;
  background: #fff;
  border-bottom: 1px solid #e5eaf1;
  color: #172033;
}
.subpage-back { color: #51627b; text-decoration: none; font-size: 0.76rem; }
.subpage-back:hover { color: #c52b38; }
.subpage-divider { height: 24px; border-left: 1px solid #dce3ec; }
.subpage-bar strong { font-size: 0.82rem; }
.subpage-bar small { color: #8b9ab0; font-size: 0.58rem; }

/* Settlement-style shell: compact navigation chrome matching the reference UI. */
@media (min-width: 901px) {
  .admin-shell {
    grid-template-columns: 220px minmax(0, 1fr);
    transition: grid-template-columns 320ms cubic-bezier(0.22, 1, 0.36, 1);
  }

  .admin-shell.sidebar-collapsed {
    grid-template-columns: 80px minmax(0, 1fr);
  }

  .admin-sidebar {
    width: 220px !important;
    padding: 0 0.65rem 0.75rem;
    gap: 0.18rem;
    border-right: 1px solid #dfe5ed;
    overflow: visible;
    position: sticky;
    z-index: 50;
    transition: width 320ms cubic-bezier(0.22, 1, 0.36, 1), padding 320ms ease;
  }

  .admin-sidebar.collapsed {
    width: 80px !important;
  }

  .sidebar-header {
    position: relative;
    height: 50px;
    padding: 0 0.35rem;
    margin: 0;
    border-bottom: 1px solid #e2e8f0;
    overflow: visible;
  }

  .shell-logo { padding: 0; gap: 0.35rem; }
  .logo-mark {
    width: 3.15rem;
    height: 1.65rem;
    border: 0;
    box-shadow: none;
    padding: 0;
  }
  .logo-text { display: grid !important; }
  .collapse-btn {
    position: absolute;
    right: -28px;
    top: 50%;
    transform: translateY(-50%);
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #fff;
    color: #64748b;
    border-color: #d9e1ec;
    margin-right: 0;
    z-index: 60;
    box-shadow: 0 2px 7px rgba(15, 23, 42, 0.08);
    transition: right 320ms cubic-bezier(0.22, 1, 0.36, 1), background 160ms ease, color 160ms ease;
  }

  .collapse-btn:hover {
    background: #fff5f5;
    color: #b4232d;
    border-color: #e8aeb4;
  }

  .nav-caption {
    padding: 0.85rem 0.9rem 0.35rem;
    font-size: 0.54rem;
    color: #718096;
    letter-spacing: 0.14em;
  }
  .admin-sidebar nav { gap: 1px; padding: 0 0.35rem; }
  .admin-sidebar nav .nav-caption {
    margin-top: 0.4rem;
    padding: 0.75rem 0.35rem 0.3rem;
  }
  .admin-sidebar nav a {
    min-height: 29px;
    padding: 0.4rem 0.35rem;
    border-radius: 7px;
    color: #40516a;
    font-size: 0.68rem;
    font-weight: 500;
  }
  .admin-sidebar:not(.collapsed) nav a span {
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .admin-sidebar nav a i { font-size: 0.78rem; color: #64748b; }
  .admin-sidebar nav a.router-link-active {
    color: #b4232d;
    background: #fff0f1;
    box-shadow: none;
    font-weight: 650;
  }
  .admin-sidebar nav a.router-link-active i { color: #b4232d; }

  .admin-sidebar.collapsed nav a {
    min-height: 44px;
    padding: 0.28rem 0.05rem;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 0.12rem;
    text-align: center;
    line-height: 1.05;
  }
  .admin-sidebar.collapsed nav a span {
    display: block !important;
    width: 100%;
    overflow: hidden;
    white-space: normal;
    color: inherit;
    font-size: 0.47rem;
    font-weight: 550;
    max-height: 1.65rem;
  }
  .admin-sidebar.collapsed nav a i { width: auto; height: auto; font-size: 0.72rem; }
  .admin-sidebar.collapsed .nav-caption { display: none !important; }
  .admin-sidebar.collapsed {
    width: 80px !important;
    padding-left: 0.35rem;
    padding-right: 0.35rem;
    align-items: stretch;
  }
  .admin-sidebar.collapsed .sidebar-header { padding: 0 0.2rem; }
  .admin-sidebar.collapsed .logo-mark { width: 3.5rem; }
  .admin-sidebar.collapsed .collapse-btn { right: -14px; }
  .admin-sidebar.collapsed .logo-text { display: none !important; }
  .admin-sidebar.collapsed nav a {
    min-height: 42px;
    padding: 0.35rem 0.2rem;
    justify-content: center;
  }
  .admin-sidebar.collapsed nav a span {
    display: block !important;
    font-size: 0.5rem;
    line-height: 1.15;
    white-space: normal;
    text-align: center;
  }
  .admin-sidebar.collapsed nav a i { font-size: 0.76rem; }
  .admin-sidebar.collapsed .shell-logo { justify-content: center; }
  .admin-sidebar.collapsed .sidebar-profile summary { justify-content:center; padding:.45rem 0; }
  .admin-sidebar.collapsed .sidebar-profile-info, .admin-sidebar.collapsed .sidebar-profile summary > i { display:none; }
  .admin-sidebar.collapsed .sidebar-profile-menu { left:calc(100% + .35rem); right:auto; bottom:.35rem; }

  .admin-topbar {
    position: relative;
    height: 50px;
    padding: 0 0.7rem;
    gap: 0.45rem;
    border-bottom: 1px solid #dfe5ed;
  }
  .topbar-actions { margin-left: auto; }
  .topbar-spacer { display: none; }
  .topbar-icon-btn {
    width: 31px;
    height: 31px;
    border-radius: 50%;
    border-color: #d9e1ec;
    font-size: 0.72rem;
  }
  .global-search { display: none; }
  .global-search.is-open {
    position: relative;
    left: auto;
    top: auto;
    z-index: auto;
    display: flex;
    width: min(300px, 24vw);
    min-height: 38px;
    padding: 0.35rem 0.5rem;
    border-radius: 10px;
    background: #fff;
    box-shadow: none;
  }
  .global-search input { font-size: 0.67rem; }
  .profile-menu summary > .avatar-initials { width: 27px; height: 27px; font-size: 0.64rem; }
  .profile-menu summary .profile-info strong { font-size: 0.62rem; }
  .profile-menu summary .profile-info small { font-size: 0.5rem; }
  .admin-content { padding: 1.35rem 1.55rem; }
  .subpage-bar { padding: 0 1.55rem; }

  .shortcut-nav {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 0.12rem;
    min-width: 0;
    flex: 1;
    overflow: hidden;
    white-space: nowrap;
  }
  .shortcut-nav a {
    flex: 0 0 auto;
    min-height: 38px;
    display: inline-flex;
    align-items: center;
    padding: 0 0.78rem;
    color: #334155;
    background: transparent !important;
    border-right: 1px solid #dce3ec;
    text-decoration: none;
    font-size: 0.64rem;
    font-weight: 550;
    line-height: 1;
    transition: color 160ms ease, background 160ms ease, box-shadow 160ms ease;
  }
  .shortcut-nav a:hover { color: #b4232d; background: #fff5f5; }
  .shortcut-nav a.router-link-active:not(.shortcut-active) {
    color: #334155;
    background: transparent;
    border-right-color: #dce3ec;
    border-radius: 0;
    font-weight: 500;
  }
  .shortcut-nav a.shortcut-active {
    color: #fff;
    background: #df202b !important;
    border-right-color: #df202b;
    border-radius: 7px;
    font-weight: 700;
    padding-left: 0.95rem;
    padding-right: 0.95rem;
    box-shadow: 0 2px 5px rgba(223, 32, 43, 0.16);
  }
  .debug-popover { color: #15803d; font-size: .58rem; white-space: nowrap; }
  .shortcut-nav a:focus-visible {
    outline: 2px solid #e63946;
    outline-offset: -2px;
  }
}

.shortcut-nav { display: none; }
@media (min-width: 901px) {
  .shortcut-nav {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: center;
    gap: 1px;
    overflow-x: auto;
  }
  .shortcut-nav a {
    display: inline-flex;
    height: 34px;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    padding: 0 10px;
    border-radius: 9px;
    color: #475569;
    text-decoration: none;
    font-size: 0.625rem;
    font-weight: 600;
    white-space: nowrap;
    transition: background 160ms ease, color 160ms ease;
  }
  .shortcut-nav a:hover { background: #f8fafc; color: #0f172a; }
  .shortcut-separator {
    width: 1px;
    height: 22px;
    background: #cbd5e1;
    margin: 0 5px;
    flex: 0 0 auto;
  }
  .shortcut-nav a.shortcut-active {
    color: #ffffff;
    background: #dc2626;
    box-shadow: 0 2px 6px rgba(220, 38, 38, 0.12);
  }
  .admin-topbar { gap: 18px; padding: 0 24px; border-bottom-color: #e2e8f0; }
  .admin-shell.sidebar-collapsed { grid-template-columns: 64px minmax(0, 1fr); }
  .admin-sidebar.collapsed { width: 64px !important; padding: 0 !important; }
  .admin-sidebar.collapsed nav { padding: 14px 3px 0; align-items: center; }
  .admin-sidebar.collapsed .sidebar-header { height: 56px; }
  .admin-sidebar:not(.collapsed) nav { padding: 14px 10px 0; gap: 2px; }
  .admin-sidebar:not(.collapsed) nav a { height: 28px; min-height: 28px; padding: 0 8px; gap: 8px; border-radius: 7px; font-size: 0.625rem; }
  .admin-sidebar:not(.collapsed) nav a.router-link-active { background: #fef2f2; color: #0f172a; box-shadow: inset 0 1px 2px rgba(0,0,0,.04); }
  .admin-sidebar:not(.collapsed) nav a.router-link-active i { color: #475569; }
  .admin-sidebar.collapsed nav a { width: 100%; height: 42px; min-height: 42px; padding: 4px; gap: 2px; border-radius: 8px; }
  .admin-sidebar.collapsed nav a.router-link-active { background: #fef2f2; box-shadow: inset 0 2px 4px rgba(0,0,0,.04); }
  .admin-sidebar.collapsed nav a.router-link-active::before { display: none; }
  .admin-sidebar.collapsed nav a span { font-size: 0.5rem; line-height: 10px; font-weight: 500; text-align: center; }
  .admin-sidebar.collapsed nav a i { width: auto; height: auto; font-size: 0.875rem; color: #475569; }
  .admin-sidebar.collapsed .logo-mark { width: 52px; height: 40px; border: 0; box-shadow: none; }
  .admin-sidebar:not(.collapsed) .logo-mark { width: 114px; height: 40px; border: 0; box-shadow: none; }
  .admin-sidebar:not(.collapsed) .sidebar-header { height: 72px; justify-content: flex-start; padding: 0 10px; align-items: flex-start; padding-top: 8px; }
  .admin-sidebar:not(.collapsed) .shell-logo { align-items: flex-start; gap: 8px; }
  .admin-sidebar:not(.collapsed) .logo-mark { width: 78px; height: 38px; padding: 0; }
  .admin-sidebar:not(.collapsed) .logo-text { display: grid; width: 42px; padding-top: 1px; font-size: 0.58rem; line-height: 1.05; }
  .admin-sidebar:not(.collapsed) .logo-text small { display: block; margin-top: 2px; font-size: 0.46rem; line-height: 1.1; }
  .admin-sidebar:not(.collapsed) .collapse-btn { top: 21px; right: -15px; }
  .admin-sidebar .nav-caption { padding-left: 8px; font-size: 0.5rem; letter-spacing: 0.1em; color: #94a3b8; }
  .admin-sidebar.collapsed .sidebar-profile-info,
  .admin-sidebar.collapsed .sidebar-profile summary > i { display: none; }
  .admin-sidebar.collapsed .sidebar-profile summary { justify-content: center; padding: 8px 0; }
  .admin-sidebar nav { overflow: hidden; }
  .admin-sidebar:not(.collapsed) nav .nav-caption { margin-top: 0; padding-top: 5px; padding-bottom: 3px; }
  .admin-sidebar:not(.collapsed) nav a { min-height: 29px; }
  .admin-sidebar.collapsed nav a { min-height: 40px; }
  .admin-sidebar .sidebar-profile { flex: 0 0 auto; margin-top: auto; padding-top: 5px; }
  .admin-sidebar:not(.collapsed) .sidebar-profile { padding-left: 6px; padding-right: 6px; }
  .admin-sidebar:not(.collapsed) .sidebar-profile summary { padding: 5px 4px; gap: 6px; }
  .admin-sidebar:not(.collapsed) .sidebar-avatar { width: 32px; height: 32px; font-size: 0.58rem; }
  .admin-sidebar:not(.collapsed) .sidebar-profile-info strong { font-size: 0.56rem; }
  .admin-sidebar:not(.collapsed) .sidebar-profile-info small { font-size: 0.44rem; }
  .topbar-actions { gap: 10px; padding-left: 12px; border-left: 1px solid #eef2f7; }
  .topbar-icon-btn { width: 34px; height: 34px; border-radius: 999px; border-color: #d9e3ef; background: #fff; }
  .navbar-collapse-btn { width: 34px; height: 34px; }
  .topbar-action-btn { height: 34px; border-color: #d9e3ef; background: #fff; padding: 0 12px; gap: 8px; }
  .topbar-action-btn:nth-child(2) { min-width: 84px; }
  .topbar-action-btn:nth-child(3) { min-width: 80px; }
  .topbar-action-btn i { color: #475569; font-size: 0.75rem; }
  .topbar-action-btn:hover, .topbar-action-btn.active { color: #334155; border-color: #d9e3ef; background: #f8fafc; }
  .topbar-action-btn:hover i, .topbar-action-btn.active i { color: #475569; }
  .topbar-icon-btn:hover { color: #334155; border-color: #d9e3ef; background: #f8fafc; }
  .topbar-icon-btn.status-ok { width: 34px; height: 34px; border-color: #bbf7d0; background: #f0fdf4; color: #22c55e; box-shadow: 0 0 0 3px rgba(34,197,94,.06); }
  .topbar-icon-btn.cloud-btn { width: auto; min-width: 42px; height: 34px; padding: 0 8px; border-color: #d9e3ef; color: #0284c7; }
}
</style>
