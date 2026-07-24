<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const search = ref('')

function runSearch() { router.push({ path: '/admin/prospects/pipeline', query: search.value.trim() ? { search: search.value.trim() } : {} }) }

async function logout() {
  await auth.logout()
  await router.replace('/login')
}
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="shell-logo"><span>Y</span><div>Yummy Food<small>Field Sales CRM</small></div></div>
      <small class="nav-caption">WORKSPACE</small>
      <nav aria-label="Administrator navigation">
        <RouterLink to="/admin/dashboard"><i class="pi pi-home" /> Dashboard</RouterLink>
        <RouterLink to="/admin/sales-executives"><i class="pi pi-user" /> Sales Executive</RouterLink>
        <RouterLink to="/admin/customers"><i class="pi pi-users" /> Customer</RouterLink>
        <RouterLink to="/admin/customer-assignment"><i class="pi pi-directions-alt" /> Customer Assignment</RouterLink>
        <RouterLink to="/admin/visit-monitoring"><i class="pi pi-map-marker" /> Visit Monitoring</RouterLink>
        <RouterLink to="/admin/prospect-finder"><i class="pi pi-compass" /> Prospect Finder</RouterLink>
        <RouterLink to="/admin/prospects/pipeline"><i class="pi pi-list" /> Prospect List</RouterLink>
        <RouterLink to="/admin/prospect-assignment"><i class="pi pi-id-card" /> Prospect Assignment</RouterLink>
        <RouterLink to="/admin/reports"><i class="pi pi-chart-bar" /> Reports</RouterLink>
      </nav>
      <div class="sidebar-note"><i class="pi pi-arrow-up-right" /><strong>Team performance</strong><span>Pipeline activity updates from real CRM records.</span></div>
    </aside>
    <div class="admin-workspace">
      <header class="admin-topbar">
        <form class="global-search" @submit.prevent="runSearch"><i class="pi pi-search" /><input v-model="search" aria-label="Search prospects" placeholder="Search prospects, customers..." /><button type="submit">Enter</button></form>
        <RouterLink class="icon-control" to="/admin/prospects/won" aria-label="Won prospect notifications"><i class="pi pi-bell" /></RouterLink>
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
/* ── Admin Shell ─────────────────────────────────────────────── */
.admin-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  background: var(--surface-page);
}

.admin-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  padding: 1.25rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  color: var(--text-primary);
  background: var(--surface-card);
  border-right: 1px solid var(--border-light);
  overflow-y: auto;
}

.shell-logo {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0 0.45rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  font-size: 0.85rem;
}

.shell-logo span {
  width: 2rem;
  height: 2rem;
  display: inline-grid;
  place-items: center;
  border-radius: var(--radius-sm);
  font-weight: 900;
  background: var(--brand-blue);
  color: #fff;
}

.shell-logo div { display: grid; }
.shell-logo small {
  color: var(--text-muted);
  font-size: 0.55rem;
  font-weight: 550;
  line-height: 1.3;
}

.nav-caption {
  padding: 0.6rem 0.55rem 0;
  color: var(--text-faint);
  font-size: 0.52rem;
  font-weight: 750;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.admin-sidebar nav {
  display: grid;
  gap: 2px;
  flex: 1;
}

.admin-sidebar nav a,
.nav-placeholder {
  display: flex;
  gap: 0.65rem;
  align-items: center;
  padding: 0.6rem 0.7rem;
  border-radius: var(--radius-sm);
  color: #647087;
  text-decoration: none;
  font-size: 0.72rem;
  font-weight: 500;
  transition: color var(--transition-fast),
              background var(--transition-fast);
}

.admin-sidebar nav a:hover {
  color: var(--brand-blue);
  background: var(--surface-hover);
}

.admin-sidebar nav a.router-link-active {
  color: var(--brand-blue);
  background: var(--brand-blue-bg);
  font-weight: 700;
  box-shadow: inset 3px 0 0 var(--brand-blue);
}

.admin-sidebar nav i {
  width: 1rem;
  text-align: center;
  font-size: 0.78rem;
}

.nav-placeholder { opacity: 0.5; }

.sidebar-note {
  padding: 0.75rem;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.25rem 0.55rem;
  color: #fff;
  background: #1a2540;
  border-radius: var(--radius-md);
}

.sidebar-note > i {
  grid-row: 1 / 3;
  width: 1.75rem;
  height: 1.75rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--brand-blue);
  border-radius: var(--radius-sm);
}

.sidebar-note strong { font-size: 0.65rem; }
.sidebar-note span {
  font-size: 0.52rem;
  line-height: 1.45;
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
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-light);
  z-index: 10;
}

.global-search {
  width: min(520px, 55vw);
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.65rem;
  color: var(--text-faint);
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast),
              box-shadow var(--transition-fast);
}

.global-search:focus-within {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}

.global-search input {
  width: 100%;
  border: 0;
  outline: 0;
  color: var(--text-primary);
  background: transparent;
  font-size: 0.72rem;
}

.global-search input::placeholder { color: var(--text-faint); }

.global-search button {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-default);
  color: var(--text-muted);
  background: var(--surface-card);
  border-radius: 4px;
  font-size: 0.52rem;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.global-search button:hover { background: var(--surface-subtle); }

.topbar-spacer { flex: 1; }

.profile-menu { position: relative; }

.profile-menu summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  list-style: none;
  cursor: pointer;
}

.profile-menu summary > .avatar-initials {
  width: 2rem;
  height: 2rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--brand-blue);
  border-radius: 50%;
  font-size: 0.72rem;
  font-weight: 800;
  flex-shrink: 0;
}

.profile-menu summary .profile-info { display: grid; text-align: left; }
.profile-menu summary .profile-info strong { font-size: 0.68rem; color: var(--text-primary); }
.profile-menu summary .profile-info small { color: var(--text-muted); font-size: 0.55rem; }

.profile-menu summary > i {
  font-size: 0.55rem;
  color: var(--text-faint);
  transition: transform var(--transition-fast);
}

.profile-menu[open] summary > i { transform: rotate(180deg); }

.profile-dropdown {
  position: absolute;
  z-index: 50;
  top: calc(100% + 6px);
  right: 0;
  min-width: 180px;
  padding: 0.35rem;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
}

.signout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 0.75rem;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: #dc2626;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: background var(--transition-fast),
              color var(--transition-fast);
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
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 900px) {
  .admin-shell { grid-template-columns: 1fr; }
  .admin-sidebar { display: none; }
  .admin-topbar { padding: 0 0.8rem; }
  .admin-content { padding: 0.8rem; }
  .profile-menu summary div { display: none; }
}

@media (max-width: 560px) {
  .global-search { width: 100%; }
  .profile-menu { display: none; }
}
</style>
