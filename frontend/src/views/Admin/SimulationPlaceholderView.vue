<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>
<template><section class="admin-page placeholder-page"><div class="placeholder-icon"><i class="pi pi-clock" /></div><p class="eyebrow">Simulation scope</p><h1>{{ route.meta.title }}</h1><p class="muted">{{ route.meta.description }}</p><div class="placeholder-actions"><RouterLink to="/admin/dashboard">Return to Dashboard</RouterLink><RouterLink to="/admin/prospects/pipeline">Open working Pipeline</RouterLink></div></section></template>
