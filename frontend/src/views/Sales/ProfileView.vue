<script setup lang="ts">import Button from 'primevue/button'; import { useRouter } from 'vue-router'; import { useAuthStore } from '../../stores/auth'; const auth = useAuthStore(); const router = useRouter(); async function logout(){ await auth.logout(); await router.replace('/login') }</script>
<template><section class="mobile-page"><div class="profile-hero"><span>{{ auth.user?.fullName?.slice(0,1) }}</span><h1>{{ auth.user?.fullName }}</h1><p>{{ auth.user?.email }}</p></div><div class="mobile-section"><div class="profile-row"><span>Application role</span><strong>Sales Executive</strong></div><div class="profile-row"><span>Session</span><strong>Authenticated</strong></div></div><Button label="Sign out securely" icon="pi pi-sign-out" severity="danger" outlined fluid @click="logout" /></section></template>

<style scoped>
.profile-hero {
  display: grid; justify-items: center; gap: 0.5rem; padding: 2.5rem 1.5rem; text-align: center;
  border-radius: var(--radius-xl); background: var(--surface-card); border: 1px solid var(--border-light); box-shadow: var(--shadow-sm);
}
.profile-hero span { width: 4rem; height: 4rem; display: grid; place-items: center; color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: 50%; font-size: 1.5rem; font-weight: 800; margin-bottom: 0.5rem; }
.profile-hero h1 { margin: 0; font-size: 1.4rem; font-weight: 800; }
.profile-hero p { margin: 0; color: var(--text-muted); font-size: 0.88rem; }
.profile-row { display: flex; justify-content: space-between; padding: 0.85rem 0; border-bottom: 1px solid var(--border-light); font-size: 0.9rem; }
.profile-row span { color: var(--text-muted); }
.profile-row strong { color: var(--text-primary); }
</style>
