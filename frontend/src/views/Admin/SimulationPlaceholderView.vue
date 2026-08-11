<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>
<template><section class="admin-page placeholder-page"><RouterLink class="placeholder-back" to="/admin/dashboard"><i class="pi pi-arrow-left" /></RouterLink><div class="placeholder-icon"><i class="pi pi-clock" /></div><p class="eyebrow">Simulation scope</p><h1>{{ route.meta.title }}</h1><p class="muted">{{ route.meta.description }}</p><div class="placeholder-actions"><RouterLink to="/admin/dashboard">Return to Dashboard</RouterLink><RouterLink to="/admin/prospects/pipeline">Open working Pipeline</RouterLink></div></section></template>
<style scoped>
.placeholder-back { display: inline-flex; align-items: center; justify-content: center; width: 2rem; height: 2rem; color: var(--brand-blue); background: var(--brand-blue-bg); border: 1px solid transparent; border-radius: var(--radius-md); text-decoration: none; font-size: 0.9rem; margin-bottom: 0.75rem; transition: background var(--transition-fast), border-color var(--transition-fast); }
.placeholder-back:hover { background: #ffd9dd; border-color: var(--brand-blue); }
</style>
