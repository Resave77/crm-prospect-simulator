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
  <div class="sales-shell">
    <header class="sales-header">
      <div class="shell-logo"><span>Y</span> Yummy CRM</div>
      <Button icon="pi pi-sign-out" aria-label="Sign out" severity="secondary" text rounded @click="logout" />
    </header>
    <main class="sales-content"><RouterView /></main>
    <nav class="sales-nav" aria-label="Sales navigation">
      <RouterLink to="/sales/dashboard"><i class="pi pi-home" /><span>Home</span></RouterLink>
      <span><i class="pi pi-briefcase" /><span>Prospects</span></span>
      <span><i class="pi pi-users" /><span>Customers</span></span>
      <span><i class="pi pi-user" /><span>Profile</span></span>
    </nav>
  </div>
</template>
