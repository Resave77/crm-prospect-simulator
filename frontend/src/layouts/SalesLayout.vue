<script setup lang="ts">
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="sales-page-wrapper">
    <div class="sales-shell">
      <header class="sales-header">
        <RouterLink class="sales-identity" to="/sales/profile">
          <span>{{ auth.user?.fullName?.slice(0, 1) }}</span>
          <div>
            <strong>Good morning, {{ auth.user?.fullName?.split(' ')[0] }}</strong>
            <small>{{ new Intl.DateTimeFormat('en', { weekday: 'long', day: 'numeric', month: 'short' }).format(new Date()) }}</small>
          </div>
        </RouterLink>

        <RouterLink class="icon-control" to="/sales/profile" aria-label="Open settings">
          <i class="pi pi-sliders-h" />
        </RouterLink>
      </header>

      <main class="sales-content">
        <RouterView />
      </main>

      <nav class="sales-nav" aria-label="Sales navigation">
        <RouterLink to="/sales/dashboard">
          <i class="pi pi-home" />
          <span>Home</span>
        </RouterLink>
        <RouterLink to="/sales/my-customers">
          <span class="nav-symbol">C</span>
          <span>Customer</span>
        </RouterLink>
        <RouterLink to="/sales/my-prospects">
          <span class="nav-symbol">P</span>
          <span>Prospect</span>
        </RouterLink>
        <RouterLink to="/sales/history">
          <i class="pi pi-history" />
          <span>History</span>
        </RouterLink>
        <RouterLink to="/sales/profile">
          <i class="pi pi-user" />
          <span>Profile</span>
        </RouterLink>
      </nav>
    </div>
  </div>
</template>

<style scoped>
/* ── Sales Mobile App Frame ──────────────────────────────────── */
.sales-page-wrapper {
  width: 100%;
  min-height: 100vh;
  background: #edf2f7;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 1.5rem 0;
}

.sales-shell {
  width: 100%;
  max-width: 440px;
  min-height: 100vh;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 28px;
  position: relative;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  padding-bottom: 80px;
  overflow: hidden;
  color: #0f172a;
}

.sales-header {
  height: 68px;
  padding: 0.85rem 1.15rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #ffffff;
  border-bottom: 1px solid #f1f5f9;
}

.sales-identity {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  color: #0f172a;
  text-decoration: none;
}

.sales-identity span {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: #2563eb;
  color: #ffffff;
  font-weight: 800;
  font-size: 1.05rem;
  flex-shrink: 0;
}

.sales-identity div { display: flex; flex-direction: column; }
.sales-identity strong { font-size: 0.95rem; font-weight: 800; color: #0f172a; line-height: 1.2; }
.sales-identity small {
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 500;
  margin-top: 2px;
}

.sales-header :deep(.icon-control) {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  color: #64748b;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  text-decoration: none;
  font-size: 0.95rem;
  transition: all 0.2s ease;
}

.sales-header :deep(.icon-control:hover) {
  color: #2563eb;
  border-color: #cbd5e1;
}

.sales-content {
  flex: 1;
  padding: 1.15rem;
}

.sales-nav {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: min(100%, 440px);
  height: 68px;
  background: #ffffff;
  border-top: 1px solid #e2e8f0;
  border-radius: 20px 20px 0 0;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  padding: 0.35rem 0.5rem;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.05);
  z-index: 100;
}

.sales-nav a {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.2rem;
  color: #64748b;
  text-decoration: none;
  font-size: 0.68rem;
  font-weight: 600;
  border-radius: 14px;
  padding: 0.3rem 0;
  transition: all 0.2s ease;
}

.sales-nav a i {
  font-size: 1.15rem;
}

.sales-nav a .nav-symbol {
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1.15;
}

.sales-nav a.router-link-active {
  color: #2563eb;
  background: #eff6ff;
  font-weight: 800;
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 480px) {
  .sales-page-wrapper {
    padding: 0;
    background: #f8fafc;
  }
  .sales-shell {
    max-width: 100%;
    border-radius: 0;
    border: none;
    box-shadow: none;
  }
}
</style>
