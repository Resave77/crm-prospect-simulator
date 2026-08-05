<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import Toast from 'primevue/toast'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const search = ref('')
const sidebarOpen = ref(false)
const sidebarCollapsed = ref(false)

const sidebarWidth = computed(() => sidebarCollapsed.value ? '64px' : '220px')

function runSearch() { router.push({ path: '/admin/prospects/list', query: search.value.trim() ? { search: search.value.trim() } : {} }) }
function closeSidebar() { sidebarOpen.value = false }
function toggleCollapse() { sidebarCollapsed.value = !sidebarCollapsed.value }

function goBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/admin/dashboard')
}

function refreshPage() {
  window.location.reload()
}

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
        <RouterLink to="/admin/dashboard" @click="closeSidebar" :title="sidebarCollapsed ? 'Dashboard' : ''">
          <i class="pi pi-home" /> <span v-show="!sidebarCollapsed">Dashboard</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">CUSTOMERS &amp; PROSPECTS</small>
        <RouterLink to="/admin/customers" @click="closeSidebar" :title="sidebarCollapsed ? 'Customer Existing' : ''">
          <i class="pi pi-users" /> <span v-show="!sidebarCollapsed">Customer Existing</span>
        </RouterLink>
        <RouterLink to="/admin/prospect-finder" @click="closeSidebar" :title="sidebarCollapsed ? 'Prospect Finder' : ''">
          <i class="pi pi-compass" /> <span v-show="!sidebarCollapsed">Prospect Finder</span>
        </RouterLink>
        <RouterLink to="/admin/prospects/list" @click="closeSidebar" :title="sidebarCollapsed ? 'Customer Prospect' : ''">
          <i class="pi pi-list" /> <span v-show="!sidebarCollapsed">Customer Prospect</span>
        </RouterLink>
        <RouterLink to="/admin/prospects/pipeline" @click="closeSidebar" :title="sidebarCollapsed ? 'Prospect Pipeline' : ''">
          <i class="pi pi-th-large" /> <span v-show="!sidebarCollapsed">Prospect Pipeline</span>
        </RouterLink>
        <RouterLink to="/admin/visit-monitoring" @click="closeSidebar" :title="sidebarCollapsed ? 'Visit Monitoring' : ''">
          <i class="pi pi-map-marker" /> <span v-show="!sidebarCollapsed">Visit Monitoring</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">MANAGEMENT</small>
        <RouterLink to="/admin/accounts" @click="closeSidebar" :title="sidebarCollapsed ? 'Accounts' : ''">
          <i class="pi pi-user-edit" /> <span v-show="!sidebarCollapsed">Accounts</span>
        </RouterLink>
        <RouterLink to="/admin/role-management" @click="closeSidebar" :title="sidebarCollapsed ? 'Role Management' : ''">
          <i class="pi pi-id-card" /> <span v-show="!sidebarCollapsed">Role Management</span>
        </RouterLink>
        <RouterLink to="/admin/sales-structure" @click="closeSidebar" :title="sidebarCollapsed ? 'Sales Structure' : ''">
          <i class="pi pi-sitemap" /> <span v-show="!sidebarCollapsed">Sales Structure</span>
        </RouterLink>

        <small v-show="!sidebarCollapsed" class="nav-caption">REPORTS</small>
        <RouterLink to="/admin/reports" @click="closeSidebar" :title="sidebarCollapsed ? 'Reports' : ''">
          <i class="pi pi-chart-bar" /> <span v-show="!sidebarCollapsed">Reports</span>
        </RouterLink>
      </nav>
    </aside>
    <div class="admin-workspace">
      <Toast position="top-right" />
      <header class="admin-topbar">
        <button class="hamburger-btn" @click="sidebarOpen = !sidebarOpen" aria-label="Toggle navigation">
          <i :class="sidebarOpen ? 'pi pi-times' : 'pi pi-bars'" />
        </button>
        <button class="topbar-icon-btn" @click="goBack" title="Back" aria-label="Back to previous page">
          <i class="pi pi-arrow-left" />
        </button>
        <button class="topbar-icon-btn" @click="refreshPage" title="Refresh" aria-label="Refresh page">
          <i class="pi pi-refresh" />
        </button>
        <form class="global-search" @submit.prevent="runSearch"><i class="pi pi-search" /><input v-model="search" aria-label="Search prospects" placeholder="Search prospects, customers..." /><button type="submit">Enter</button></form>
        <div class="topbar-spacer" />
        <details class="profile-menu">
          <summary>
            <span class="avatar-initials">{{ auth.user?.fullName?.slice(0, 1) }}</span>
            <div class="profile-info"><strong>{{ auth.user?.fullName }}</strong><small>Administrator</small></div>
            <i class="pi pi-chevron-down" />
          </summary>
          <div class="profile-dropdown">
            <button class="signout-btn" @click="logout">
              <i class="pi pi-sign-out" />
              <span>Sign out</span>
            </button>
          </div>
        </details>
      </header>
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
  background: #2563eb;
  color: #fff;
  border-color: #2563eb;
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
  color: #2563eb;
  background: #f4f7fc;
}

.admin-sidebar.collapsed nav a:hover {
  background: #eef4ff;
}

.admin-sidebar nav a.router-link-active {
  color: #2563eb;
  background: linear-gradient(135deg, #eef4ff 0%, #eaf1ff 100%);
  font-weight: 700;
  box-shadow: inset 3px 0 0 #2563eb;
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
  background: #2563eb;
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

.admin-sidebar nav a:hover i { color: #2563eb; }
.admin-sidebar nav a.router-link-active i { color: #2563eb; }

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
  background: linear-gradient(135deg, #2563eb, #6366f1);
  border-radius: 9px;
  box-shadow: 0 4px 10px -2px rgba(37, 99, 235, 0.5);
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
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.09);
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
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  border-radius: 50%;
  font-size: 0.72rem;
  font-weight: 800;
  flex-shrink: 0;
  box-shadow: 0 3px 8px -2px rgba(37, 99, 235, 0.45);
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
  color: #2563eb;
  border-color: #e2e8f0;
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
  .admin-sidebar.collapsed nav a.router-link-active { box-shadow: inset 3px 0 0 #2563eb; }
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
</style>


