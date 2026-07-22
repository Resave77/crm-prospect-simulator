<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function logout() {
  await auth.logout()
  await router.replace('/login')
}
</script>

<template>
  <div class="admin-shell">
    <aside class="admin-sidebar">
      <div class="shell-logo"><span>Y</span> Yummy CRM</div>
      <nav aria-label="Administrator navigation">
        <RouterLink to="/admin/dashboard"><i class="pi pi-home" /> Dashboard</RouterLink>
        <span class="nav-placeholder"><i class="pi pi-map" /> Prospect Finder</span>
        <span class="nav-placeholder"><i class="pi pi-users" /> Existing Customers</span>
      </nav>
      <Button label="Sign out" icon="pi pi-sign-out" severity="secondary" text @click="logout" />
    </aside>
    <main class="admin-content"><RouterView /></main>
  </div>
</template>
