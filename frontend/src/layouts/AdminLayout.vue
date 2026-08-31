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

const shortcutItems = computed(() => {
  const path = route.path
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
  } else if (path.includes('/prospects')) {
    items = [
      { label: 'Prospect Finder', to: '/admin/prospect-finder', permission: 'view_prospect_finder' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
    ]
  } else if (path.includes('/customers')) {
    items = [
      { label: 'Customer Site', to: { path: '/admin/customers', query: { tab: 'site' } }, permission: 'view_customers' },
      { label: 'Company', to: { path: '/admin/customers', query: { tab: 'company' } }, permission: 'view_customers' },
      { label: 'Master Data', to: { path: '/admin/customers', query: { tab: 'master' } }, permission: 'view_customers' },
      { label: 'Add Customer', to: '/admin/customers/add', permission: 'view_customers' },
      { label: 'Add Company', to: '/admin/companies/add', permission: 'view_customers' },
    ]
  } else if (path.includes('/accounts')) {
    items = [
      { label: 'Employee List', to: '/admin/accounts', permission: 'view_accounts' },
      { label: 'Create Employee', to: '/admin/accounts/create', permission: 'view_accounts' },
      { label: 'Role Management', to: '/admin/role-management', permission: 'view_roles' },
      { label: 'Sales Structure', to: '/admin/sales-structure', permission: 'view_sales_structure' },
    ]
  } else if (path.includes('/visit-monitoring')) {
    items = [
      { label: 'Customer List', to: '/admin/customers', permission: 'view_customers' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
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
      { label: 'Prospect Finder', to: '/admin/prospect-finder', permission: 'view_prospect_finder' },
      { label: 'Prospect List', to: '/admin/prospects/list', permission: 'view_prospect_list' },
      { label: 'Prospect Pipeline', to: '/admin/prospects/pipeline', permission: 'view_prospect_pipeline' },
    ]
  }

  return items.filter((item) => auth.hasPermission(item.permission))
})

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

function isShortcutActive(item: { to: string | { path: string; query?: Record<string, string> } }) {
  if (typeof item.to === 'string') return route.path === item.to || route.path.startsWith(`${item.to}/`)
  return route.path === item.to.path && (route.query.tab || 'site') === item.to.query?.tab
}

const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '220px')

function runSearch() {
  const value = search.value.trim()
  if (route.path.includes('/prospects/list')) {
    router.replace({ path: route.path, query: value ? { ...route.query, search: value } : { ...route.query, search: undefined } })
  } else if (route.path.includes('/accounts')) {
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
  <div class="admin-shell" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
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
      <nav aria-label="Administrator navigation">
        <RouterLink v-if="auth.hasPermission('view_admin_dashboard')" to="/admin/dashboard" @click="closeSidebar" :title="sidebarCollapsed ? 'Dashboard' : ''">
          <i class="pi pi-home" /> <span>Dashboard</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">CUSTOMERS &amp; PROSPECTS</small>
        <RouterLink v-if="auth.hasPermission('view_customers')" to="/admin/customers" @click="closeSidebar" :title="sidebarCollapsed ? 'Customer Existing' : ''">
          <i class="pi pi-users" /> <span>Customer Existing</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_prospect_finder')" to="/admin/prospect-finder" @click="closeSidebar" :title="sidebarCollapsed ? 'Prospect Finder' : ''">
          <i class="pi pi-compass" /> <span>Prospect Finder</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_prospect_list')" to="/admin/prospects/list" @click="closeSidebar" :title="sidebarCollapsed ? 'Customer Prospect' : ''">
          <i class="pi pi-list" /> <span>Customer Prospect</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_prospect_pipeline')" to="/admin/prospects/pipeline" @click="closeSidebar" :title="sidebarCollapsed ? 'Prospect Pipeline' : ''">
          <i class="pi pi-th-large" /> <span>Prospect Pipeline</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_visit_monitoring')" to="/admin/visit-monitoring" @click="closeSidebar" :title="sidebarCollapsed ? 'Visit Monitoring' : ''">
          <i class="pi pi-map-marker" /> <span>Visit Monitoring</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">MANAGEMENT</small>
        <RouterLink v-if="auth.hasPermission('view_accounts')" to="/admin/accounts" @click="closeSidebar" :title="sidebarCollapsed ? 'Accounts' : ''">
          <i class="pi pi-user-edit" /> <span>Employee Management</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_roles')" to="/admin/role-management" @click="closeSidebar" :title="sidebarCollapsed ? 'Role Management' : ''">
          <i class="pi pi-id-card" /> <span>Role Management</span>
        </RouterLink>
        <RouterLink v-if="auth.hasPermission('view_sales_structure')" to="/admin/sales-structure" @click="closeSidebar" :title="sidebarCollapsed ? 'Sales Structure' : ''">
          <i class="pi pi-sitemap" /> <span>Sales Structure</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">REPORTS</small>
        <RouterLink v-if="auth.hasPermission('view_reports')" to="/admin/reports" @click="closeSidebar" :title="sidebarCollapsed ? 'Reports' : ''">
          <i class="pi pi-chart-bar" /> <span>Reports</span>
        </RouterLink>
        <RouterLink to="/admin/api-usage" @click="closeSidebar" :title="sidebarCollapsed ? 'Monitoring API' : ''">
          <i class="pi pi-chart-line" /> <span>Monitoring API</span>
        </RouterLink>
      </nav>
      <details class="sidebar-profile">
        <summary>
          <span class="sidebar-avatar">{{ auth.user?.fullName?.slice(0, 1) }}</span>
          <span class="sidebar-profile-info"><strong>{{ auth.user?.fullName }}</strong><small>Administrator</small></span>
          <i class="pi pi-ellipsis-v" />
        </summary>
        <div class="sidebar-profile-menu"><button class="signout-btn" @click="logout"><i class="pi pi-sign-out" /><span>Sign out</span></button></div>
      </details>
    </aside>
    <div class="admin-workspace">
      <Toast position="top-right" />
      <header class="admin-topbar">
        <button class="hamburger-btn" @click="sidebarOpen = !sidebarOpen" aria-label="Toggle navigation">
          <i :class="sidebarOpen ? 'pi pi-times' : 'pi pi-bars'" />
        </button>
        <nav v-if="!subPage" class="shortcut-nav" aria-label="Page shortcuts">
          <RouterLink v-for="item in shortcutItems" :key="`shortcut-${item.label}`" :to="item.to" :class="{ 'shortcut-active': isShortcutActive(item) }" active-class="" @click="closeSidebar">
            {{ item.label }}
          </RouterLink>
        </nav>
        <form v-if="searchOpen" class="global-search is-open" @submit.prevent="runSearch"><i class="pi pi-search" /><input v-model="search" autofocus aria-label="Search prospects" placeholder="Search prospects, customers..." /><button type="button" class="search-close" aria-label="Clear search" title="Clear search" @click="search = ''">×</button><button type="submit">Enter</button></form>
        <div class="topbar-actions">
          <button class="topbar-icon-btn navbar-collapse-btn" type="button" :title="navbarCollapsed ? 'Show shortcuts' : 'Minimize shortcuts'" :aria-label="navbarCollapsed ? 'Show shortcuts' : 'Minimize shortcuts'" @click="toggleNavbar"><i class="pi" :class="navbarCollapsed ? 'pi-chevron-left' : 'pi-chevron-right'" /></button>
          <button v-if="!navbarCollapsed" class="topbar-action-btn" type="button" title="Search" aria-label="Toggle search" @click="toggleSearch"><i class="pi pi-search" /> <span>Search</span></button>
          <button v-if="!navbarCollapsed" class="topbar-action-btn" :class="{ active: debugMode }" type="button" title="Toggle debug mode" aria-label="Toggle debug mode" @click="toggleDebug"><i class="pi pi-database" /> <span>Debug</span></button>
          <button class="topbar-icon-btn" type="button" title="Fullscreen" aria-label="Toggle fullscreen" @click="toggleFullscreen"><i class="pi pi-expand" /></button>
          <button class="topbar-icon-btn status-ok" :class="{ offline: !isOnline }" type="button" :title="isOnline ? 'System status: online' : 'System status: offline'" :aria-label="isOnline ? 'System status: online' : 'System status: offline'"><i class="pi" :class="isOnline ? 'pi-check' : 'pi-times'" /></button>
          <button class="topbar-icon-btn cloud-btn" :class="{ syncing: cloudSyncing }" type="button" :title="cloudSyncing ? 'Syncing...' : 'Sync cloud data'" aria-label="Sync cloud data" @click="syncCloud"><i class="pi" :class="cloudSyncing ? 'pi-spin pi-spinner' : 'pi-cloud-upload'" /><small>{{ cloudSyncing ? '…' : '0' }}</small></button>
        </div>
        <div class="topbar-spacer" />
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
      <div v-if="subPage" class="subpage-bar">
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
  background: #f7f9fb;
  transition: grid-template-columns 0.25s ease;
}

.admin-shell.sidebar-collapsed {
  grid-template-columns: 64px minmax(0, 1fr);
}

.admin-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  padding: 1.1rem 0.7rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  color: #1e293b;
  background: #ffffff;
  border-right: 1px solid #edf1f6;
  overflow-x: hidden;
  overflow-y: hidden;
  transition: width 0.25s ease;
}

.admin-sidebar.collapsed {
  padding: 1.1rem 0.5rem;
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
  padding-bottom: 0.85rem;
  border-bottom: 1px solid #f1f4f8;
  margin-bottom: 0.15rem;
}

.admin-sidebar.collapsed .sidebar-header {
  flex-direction: column;
  gap: 0.55rem;
}

/* ── Collapse Toggle ────────────────────────────────────────── */
.collapse-btn {
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border: 1px solid #edf1f6;
  border-radius: 7px;
  background: #f8fafc;
  color: #94a3b8;
  cursor: pointer;
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
  padding-right: 2px;
}

.admin-sidebar nav::-webkit-scrollbar { width: 4px; }
.admin-sidebar nav::-webkit-scrollbar-track { background: transparent; }
.admin-sidebar nav::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 999px; }

.admin-sidebar nav a,
.nav-placeholder {
  display: flex;
  gap: 0.65rem;
  align-items: center;
  padding: 0.55rem 0.65rem;
  border-radius: 9px;
  color: #64748b;
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 550;
  white-space: nowrap;
  flex-shrink: 0;
  transition: color 150ms ease, background 150ms ease, padding 150ms ease;
}

.admin-sidebar.collapsed nav a {
  justify-content: center;
  padding: 0.6rem 0;
  border-radius: 9px;
}

.admin-sidebar nav a:hover {
  color: #e63946;
  background: #f4f7fc;
}

.admin-sidebar.collapsed nav a:hover {
  background: #fff0f1;
}

.admin-sidebar nav a.router-link-active {
  color: #e63946;
  background: linear-gradient(135deg, #fff0f1 0%, #fff5f5 100%);
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
  color: #94a3b8;
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
.sidebar-profile { position: relative; margin-top: auto; border-top: 1px solid #edf1f6; padding-top: .65rem; }
.sidebar-profile summary { display:flex; align-items:center; gap:.55rem; padding:.45rem .4rem; border-radius:9px; list-style:none; cursor:pointer; }
.sidebar-profile summary:hover { background:#f8fafc; }.sidebar-profile summary::-webkit-details-marker { display:none; }
.sidebar-avatar { display:grid; place-items:center; width:30px; height:30px; flex:none; border-radius:50%; background:linear-gradient(135deg,#e63946,#c92332); color:#fff; font-size:.68rem; font-weight:800; }
.sidebar-profile-info { display:grid; min-width:0; flex:1; gap:.1rem; }.sidebar-profile-info strong { overflow:hidden; color:#172033; font-size:.63rem; text-overflow:ellipsis; white-space:nowrap; }.sidebar-profile-info small { color:#94a3b8; font-size:.52rem; }.sidebar-profile summary > i { color:#94a3b8; font-size:.7rem; }
.sidebar-profile-menu { position:absolute; right:.35rem; bottom:calc(100% + .35rem); width:170px; padding:.3rem; border:1px solid #e5eaf0; border-radius:10px; background:#fff; box-shadow:0 12px 25px rgba(15,23,42,.14); }
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
    background: #f5f7fa;
    transition: grid-template-columns 320ms cubic-bezier(0.22, 1, 0.36, 1);
  }

  .admin-shell.sidebar-collapsed {
    grid-template-columns: 120px minmax(0, 1fr);
  }

  .admin-sidebar {
    width: 220px !important;
    padding: 0 0.65rem 0.75rem;
    gap: 0.18rem;
    border-right: 1px solid #dfe5ed;
    overflow: visible;
    position: sticky;
    z-index: 20;
    transition: width 320ms cubic-bezier(0.22, 1, 0.36, 1), padding 320ms ease;
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
    z-index: 3;
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
    width: 120px !important;
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
  .shortcut-nav { display: flex; }
}
</style>


