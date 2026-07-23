<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
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
        <details class="profile-menu"><summary><span>{{ auth.user?.fullName?.slice(0, 1) }}</span><div><strong>{{ auth.user?.fullName }}</strong><small>Administrator</small></div></summary><Button label="Sign out" icon="pi pi-sign-out" severity="secondary" text @click="logout" /></details>
      </header>
      <main class="admin-content"><RouterView /></main>
    </div>
  </div>
</template>
