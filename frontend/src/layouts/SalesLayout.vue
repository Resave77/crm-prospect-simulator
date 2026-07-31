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
        <RouterLink to="/sales/dashboard" class="sidebar-link">
          <i class="pi pi-home" /><span>Home</span>
        </RouterLink>
        <RouterLink to="/sales/my-customers" class="sidebar-link">
          <i class="pi pi-users" /><span>Customer</span>
        </RouterLink>
        <RouterLink to="/sales/my-prospects" class="sidebar-link">
          <i class="pi pi-briefcase" /><span>Prospect</span>
        </RouterLink>
        <RouterLink to="/sales/history" class="sidebar-link">
          <i class="pi pi-history" /><span>History</span>
        </RouterLink>
        <RouterLink to="/sales/profile" class="sidebar-link">
          <i class="pi pi-user" /><span>Profile</span>
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
      <main class="sales-content">
        <RouterView />
      </main>
    </div>

    <!-- Mobile bottom navigation -->
    <nav class="sales-nav" aria-label="Sales navigation">
      <RouterLink to="/sales/dashboard" class="nav-item">
        <i class="pi pi-home" /><span>Home</span>
      </RouterLink>
      <RouterLink to="/sales/my-customers" class="nav-item">
        <i class="pi pi-users" /><span>Customer</span>
      </RouterLink>
      <RouterLink to="/sales/my-prospects" class="nav-item">
        <i class="pi pi-briefcase" /><span>Prospect</span>
      </RouterLink>
      <RouterLink to="/sales/history" class="nav-item">
        <i class="pi pi-history" /><span>History</span>
      </RouterLink>
      <RouterLink to="/sales/profile" class="nav-item">
        <i class="pi pi-user" /><span>Profile</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import Toast from 'primevue/toast'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const initials = computed(() => {
  const name = auth.user?.fullName ?? ''
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
})

const roleLabel = computed(() => {
  if (auth.user?.role === 'ADMINISTRATOR') return 'Administrator'
  if (auth.user?.role === 'SALES_MANAGER') return 'Sales Manager'
  if (auth.user?.role === 'SALES_EXECUTIVE') return 'Sales Executive'
  return auth.user?.role ?? '—'
})

async function logout() {
  await auth.logout()
  await router.replace('/login')
}
</script>

<style>
/* ── Non-scoped: global desktop adjustments for Sales fixed elements ── */
@media (min-width: 768px) {
  .sales-layout .detail-bottom-bar {
    left: var(--sales-shell-sidebar-w, 0px);
    right: 0;
    width: auto;
  }
  .sales-layout .mp-fab,
  .sales-layout .mc-fab {
    bottom: 1.5rem;
    right: 1.5rem;
  }
}

/* ── Toast positioning for Sales layout ── */
.sales-layout .p-toast-top-right {
  top: 1rem;
  right: 1.5rem;
}
@media (max-width: 767px) {
  .sales-layout .p-toast {
    left: 1rem;
    right: 1rem;
    width: auto;
  }
  .sales-layout .p-toast-top-right {
    top: calc(env(safe-area-inset-top) + 0.75rem);
    right: 1rem;
  }
}
</style>

<style scoped>
/* ── Layout Root ──────────────────────────────────────────────── */
.sales-layout {
  --sales-shell-sidebar-w: 0px;
  --sales-nav-height: 78px;
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
  grid-template-columns: repeat(5, minmax(0, 1fr));
  padding: 0.45rem;
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
  gap: 0.18rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.66rem;
  font-weight: 700;
  border-radius: 18px;
  transition: all 0.2s ease;
}

.sales-nav .nav-item i {
  font-size: 1.18rem;
  transition:
    color 0.2s ease,
    transform 0.2s ease;
}

.sales-nav .nav-item.router-link-active {
  color: var(--brand-blue);
  background: #eef5ff;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(37, 99, 235, 0.05);
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
@media (min-width: 768px) {
  .sales-layout {
    --sales-shell-sidebar-w: 240px;
    --sales-nav-height: 0px;
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

  .sales-shell {
    padding-bottom: 0;
  }

  .sales-content {
    max-width: 1280px;
    padding: 2rem;
  }
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
  background: linear-gradient(135deg, var(--brand-blue), #1d4ed8);
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
