<template>
  <div class="sales-layout">
    <!-- Desktop sidebar -->
    <aside class="sales-sidebar">
      <div class="sidebar-header">
        <RouterLink to="/sales/dashboard" class="sidebar-brand">
          <span class="sidebar-logo">
            <img src="/yummy-logo.png" alt="Yummy Dairy" />
          </span>
          <span class="sidebar-brand-text">Yummy CRM</span>
        </RouterLink>
      </div>
      <nav class="sidebar-nav" aria-label="Sales navigation">
        <RouterLink v-for="item in visibleNavItems" :key="item.to" :to="item.to" class="sidebar-link">
          <i :class="item.icon" /><span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <div class="sidebar-footer">
        <div class="sidebar-user">
          <span class="sidebar-user-avatar">{{ initials }}</span>
          <div class="sidebar-user-info">
            <span class="sidebar-user-name">{{ auth.user?.fullName ?? '—' }}</span>
            <span class="sidebar-user-role">{{ roleLabel }}</span>
          </div>
        </div>
        <button class="sidebar-logout" @click="logout">
          <i class="pi pi-sign-out" /><span>Sign out</span>
        </button>
      </div>
    </aside>

    <!-- Main content area -->
    <div class="sales-shell">
      <Toast position="top-right" />
      <header class="sales-topbar">
        <button class="topbar-icon-btn" @click="goBack" title="Back" aria-label="Back to previous page">
          <i class="pi pi-arrow-left" />
        </button>
        <button class="topbar-icon-btn" @click="refreshPage" title="Refresh" aria-label="Refresh page">
          <i class="pi pi-refresh" />
        </button>
        <div class="topbar-title">{{ pageTitle }}</div>
        <div class="topbar-spacer" />
        <details class="profile-menu">
          <summary>
            <span class="avatar-initials">{{ initials }}</span>
            <div class="profile-info"><strong>{{ auth.user?.fullName }}</strong><small>{{ roleLabel }}</small></div>
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
      <main class="sales-content">
        <RouterView />
      </main>
    </div>

    <!-- Mobile bottom navigation -->
    <nav class="sales-nav" aria-label="Sales navigation">
      <RouterLink v-for="item in visibleNavItems" :key="item.to" :to="item.to" class="nav-item">
        <i :class="item.icon" /><span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Toast from 'primevue/toast'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const salesNavItems = [
  { label: 'Home', to: '/sales/dashboard', icon: 'pi pi-home', permission: 'view_sales_dashboard' },
  { label: 'Customer', to: '/sales/my-customers', icon: 'pi pi-users', permission: 'view_my_customers' },
  { label: 'Prospect Pipeline', to: '/sales/pipeline', icon: 'pi pi-chart-line', permission: 'menu_sales_pipeline' },
  { label: 'History', to: '/sales/history', icon: 'pi pi-history', permission: 'view_sales_history' },
  { label: 'Profile', to: '/sales/profile', icon: 'pi pi-user', permission: 'view_own_profile' },
]

const visibleNavItems = computed(() => salesNavItems.filter((item) => auth.hasPermission(item.permission)))

const pageTitle = computed(() => {
  const p = route.path
  if (p.includes('/check-in')) return 'Check In'
  if (p.includes('/check-out')) return 'Check Out'
  if (p.includes('/visit-result')) return 'Visit Result'
  if (p.startsWith('/sales/prospects/')) return 'Prospect Details'
  if (p.startsWith('/sales/customers/')) return 'Customer Details'
  const titles: Record<string, string> = {
    '/sales/dashboard': 'Home',
    '/sales/my-customers': 'My Customers',
    '/sales/my-prospects': 'My Prospects',
    '/sales/pipeline': 'Prospect Pipeline',
    '/sales/history': 'Visit History',
    '/sales/profile': 'My Profile',
  }
  return titles[p] ?? 'Yummy CRM'
})

const initials = computed(() => {
  const name = auth.user?.fullName ?? ''
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
})

const roleLabel = computed(() => {
  if (auth.user?.salesRole?.name) return auth.user.salesRole.name
  if (auth.user?.role === 'ADMINISTRATOR') return 'Administrator'
  if (auth.user?.role === 'SALES_MANAGER') return 'Sales Manager'
  if (auth.user?.role === 'SALES_EXECUTIVE') return 'Sales Executive'
  return auth.user?.role ?? '—'
})

async function logout() {
  await auth.logout()
  await router.replace('/login')
}

function goBack() {
  if (window.history.length > 1) router.back()
  else router.replace('/sales/dashboard')
}

function refreshPage() {
  window.location.reload()
}
</script>

<style>
/* ── Non-scoped: global desktop adjustments for Sales fixed elements ── */
@media (min-width: 769px) {
  .sales-layout .detail-bottom-bar {
    left: var(--sales-shell-sidebar-w, 0px) !important;
    right: 0 !important;
    width: auto !important;
    padding-right: calc(7rem + env(safe-area-inset-right, 0px)) !important;
  }
  .sales-layout .mp-fab,
  .sales-layout .mc-fab,
  .sales-layout .sales-chat-fab,
  .sales-layout .chat-fab,
  .sales-layout .pc-launcher {
    bottom: calc(var(--desktop-action-bar-height, 0px) + 1rem);
    right: 1.5rem;
  }
  .sales-layout .p-toast-top-right {
    top: calc(var(--sales-topbar-h, 0px) + 1rem);
  }
}

/* ── Toast positioning for Sales layout ── */
.sales-layout .p-toast-top-right {
  top: 1rem;
  right: 1.5rem;
}
@media (max-width: 768px) {
  .sales-layout .mp-fab,
  .sales-layout .mc-fab,
  .sales-layout .sales-chat-fab,
  .sales-layout .chat-fab,
  .sales-layout .pc-launcher {
    right: calc(1rem + env(safe-area-inset-right, 0px));
    bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 2rem + env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout .pc-wrap {
    bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 1rem + env(safe-area-inset-bottom, 0px)) !important;
    max-height: calc(100dvh - var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) - 2rem - env(safe-area-inset-top, 0px) - env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout .detail-page {
    padding-bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 5.75rem + env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout .detail-bottom-bar {
    left: 1rem;
    right: 1rem;
    bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 0.9rem + env(safe-area-inset-bottom, 0px)) !important;
    width: auto;
    z-index: 90;
    border: 1px solid rgba(191, 219, 254, 0.95);
    border-radius: 14px;
    padding: 0.55rem;
    background: rgba(255, 255, 255, 0.96);
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.12);
  }
  .sales-layout .detail-bottom-bar .dbar-btn {
    min-height: 44px;
  }
  .sales-layout:has(.detail-bottom-bar) .pc-launcher {
    bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 5.35rem + env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout:has(.detail-bottom-bar) .pc-wrap {
    bottom: calc(var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) + 4.35rem + env(safe-area-inset-bottom, 0px)) !important;
    max-height: calc(100dvh - var(--mobile-bottom-nav-height, var(--sales-nav-height, 78px)) - 5.35rem - env(safe-area-inset-top, 0px) - env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout .leaflet-container {
    z-index: 0;
  }
  .sales-layout .leaflet-tile-pane {
    z-index: 1 !important;
  }
  .sales-layout .leaflet-overlay-pane {
    z-index: 2 !important;
  }
  .sales-layout .leaflet-shadow-pane {
    z-index: 3 !important;
  }
  .sales-layout .leaflet-marker-pane {
    z-index: 4 !important;
  }
  .sales-layout .leaflet-tooltip-pane {
    z-index: 5 !important;
  }
  .sales-layout .leaflet-popup-pane,
  .sales-layout .leaflet-control-container {
    z-index: 6 !important;
  }
  .sales-layout .p-toast {
    left: 1rem;
    right: 1rem;
    width: auto;
  }
  .sales-layout .p-toast-top-right {
    top: calc(env(safe-area-inset-top) + 0.75rem);
    right: 1rem;
  }
  /* Keep the prospect discussion panel & launcher clear of the mobile bottom nav */
  .sales-layout .sales-shell .pc-launcher {
    bottom: calc(5.75rem + env(safe-area-inset-bottom, 0px));
  }
  /* Detail pages also have a fixed action bar; keep comments above it. */
  .sales-layout .sales-shell .detail-page .pc-launcher {
    right: 1rem;
    bottom: calc(10.25rem + env(safe-area-inset-bottom, 0px)) !important;
  }
  .sales-layout .sales-shell .pc-wrap {
    bottom: calc(9.75rem + env(safe-area-inset-bottom, 0px));
    max-height: calc(100dvh - 10.75rem - env(safe-area-inset-bottom, 0px));
  }
  .sales-layout .sales-shell .pc-list {
    max-height: calc(100dvh - 18rem - env(safe-area-inset-bottom, 0px));
  }
}
</style>

<style scoped>
/* ── Layout Root ──────────────────────────────────────────────── */
.sales-layout {
  --sales-shell-sidebar-w: 0px;
  --sales-nav-height: 78px;
  --mobile-bottom-nav-height: 78px;
  --desktop-action-bar-height: 72px;
  display: flex;
  min-height: 100vh;
  min-height: 100dvh;
  width: 100%;
  background:
    linear-gradient(180deg, #f3f7fd 0%, #f8fafc 42%, #f8fafc 100%);
}

/* ── Desktop Sidebar ──────────────────────────────────────────── */
.sales-sidebar {
  display: none;
}

/* ── Mobile Bottom Nav ────────────────────────────────────────── */
.sales-nav {
  position: fixed;
  bottom: max(0.65rem, env(safe-area-inset-bottom, 0px));
  left: 0.75rem;
  right: 0.75rem;
  width: auto;
  max-width: 460px;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid rgba(219, 226, 237, 0.9);
  border-radius: 24px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(64px, 1fr));
  min-height: var(--mobile-bottom-nav-height);
  padding: 0.42rem 0.45rem;
  box-shadow:
    0 18px 42px rgba(15, 23, 42, 0.14),
    0 4px 12px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(18px);
  z-index: 100;
}

.sales-nav .nav-item {
  min-width: 0;
  min-height: 54px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.2rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.66rem;
  font-weight: 700;
  border-radius: 18px;
  transition: all 0.2s ease;
}

.sales-nav .nav-item i {
  width: 20px;
  height: 20px;
  display: inline-grid;
  place-items: center;
  font-size: 1rem;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.sales-nav .nav-item.router-link-active {
  color: var(--brand-blue);
  background: #fff1f2;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(209, 67, 80, 0.05);
}

.sales-nav .nav-item.router-link-active i {
  transform: translateY(-1px);
}

.sales-nav .nav-item:not(.router-link-active):hover {
  color: #64748b;
}

/* ── Content Shell ────────────────────────────────────────────── */
.sales-shell {
  flex: 1;
  min-width: 0;
  width: 100%;
  padding-bottom: calc(var(--sales-nav-height) + 1.5rem + env(safe-area-inset-bottom, 0px));
}

.sales-content {
  width: 100%;
  max-width: 480px;
  margin: 0 auto;
  padding: max(1rem, env(safe-area-inset-top, 0px)) 1rem 1.25rem;
}

/* ── Desktop Breakpoint ───────────────────────────────────────── */
@media (min-width: 769px) {
  .sales-layout {
    --sales-shell-sidebar-w: 240px;
    --sales-nav-height: 0px;
    --mobile-bottom-nav-height: 0px;
    --sales-topbar-h: 56px;
  }

  .sales-sidebar {
    display: flex;
    flex-direction: column;
    width: 240px;
    position: sticky;
    top: 0;
    height: 100vh;
    height: 100dvh;
    background: var(--surface-card);
    border-right: 1px solid var(--border-light);
    overflow-y: auto;
    flex-shrink: 0;
  }

  .sales-nav {
    display: none;
  }

  .sales-topbar {
    display: flex;
  }

  .sales-shell {
    padding-bottom: 0;
  }

  .sales-content {
    max-width: 1280px;
    padding: 2rem;
  }
}

/* ── Desktop Topbar ──────────────────────────────────────────── */
.sales-topbar {
  display: none;
  align-items: center;
  gap: 0.5rem;
  height: var(--sales-topbar-h, 56px);
  padding: 0 1.5rem;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-light);
  position: sticky;
  top: 0;
  z-index: 30;
}

.topbar-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border: 1px solid var(--border-light);
  border-radius: 9px;
  background: #ffffff;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 150ms ease, color 150ms ease, border-color 150ms ease;
}

.topbar-icon-btn:hover {
  background: var(--surface-hover);
  color: var(--brand-blue);
  border-color: var(--border-default);
}

.topbar-title {
  margin-left: 0.35rem;
  font-size: 0.9rem;
  font-weight: 800;
  letter-spacing: -0.01em;
  color: var(--text-primary);
  white-space: nowrap;
}

.topbar-spacer {
  flex: 1;
}

.profile-menu {
  position: relative;
}

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

.profile-menu summary:hover {
  background: var(--surface-hover);
}

.profile-menu summary .avatar-initials {
  width: 2rem;
  height: 2rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, var(--brand-blue), #bb3342);
  border-radius: 50%;
  font-size: 0.72rem;
  font-weight: 800;
  flex-shrink: 0;
  box-shadow: 0 3px 8px -2px rgba(209, 67, 80, 0.45);
}

.profile-menu summary .profile-info {
  display: grid;
  text-align: left;
}

.profile-menu summary .profile-info strong {
  font-size: 0.68rem;
  color: var(--text-primary);
}

.profile-menu summary .profile-info small {
  color: var(--text-muted);
  font-size: 0.55rem;
}

.profile-menu summary > i {
  font-size: 0.55rem;
  color: var(--text-faint);
  transition: transform 150ms ease;
}

.profile-menu[open] summary > i {
  transform: rotate(180deg);
}

.profile-dropdown {
  position: absolute;
  z-index: 50;
  top: calc(100% + 6px);
  right: 0;
  min-width: 180px;
  padding: 0.35rem;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
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

.signout-btn i {
  font-size: 0.8rem;
}

.signout-btn:hover {
  background: #fef2f2;
  color: #b91c1c;
}

/* ── Sidebar Styles ───────────────────────────────────────────── */
.sidebar-header {
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border-light);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: var(--text-primary);
}

.sidebar-logo {
  width: 52px;
  height: 36px;
  display: grid;
  place-items: center;
  padding: 0.25rem;
  border-radius: 10px;
  background: #fff;
  border: 1px solid var(--border-light);
  flex-shrink: 0;
}

.sidebar-logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.sidebar-brand-text {
  font-size: 0.92rem;
  font-weight: 800;
  color: var(--text-primary);
}

.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  border-radius: 12px;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.82rem;
  font-weight: 600;
  transition: all 0.15s ease;
}

.sidebar-link i {
  font-size: 1rem;
  width: 20px;
  text-align: center;
}

.sidebar-link:hover {
  color: var(--text-primary);
  background: var(--surface-hover);
}

.sidebar-link.router-link-active {
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  font-weight: 700;
}

.sidebar-footer {
  padding: 0.75rem;
  border-top: 1px solid var(--border-light);
}

.sidebar-user {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.5rem 0.85rem;
  margin-bottom: 0.5rem;
}

.sidebar-user-avatar {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--brand-blue), #bb3342);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 800;
  flex-shrink: 0;
}

.sidebar-user-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.sidebar-user-name {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-user-role {
  font-size: 0.65rem;
  color: var(--text-muted);
}

.sidebar-logout {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.65rem 0.85rem;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.sidebar-logout:hover {
  color: #dc2626;
  background: #fef2f2;
}

.sidebar-logout i {
  font-size: 1rem;
  width: 20px;
  text-align: center;
}

</style>
