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
