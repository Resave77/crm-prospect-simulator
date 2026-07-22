<script setup lang="ts">
import { useAuthStore } from '../../../stores/auth'
const auth = useAuthStore()
</script>

<template>
  <section class="dashboard-page">
    <p class="eyebrow">Administrator workspace</p>
    <h1>Good to see you, {{ auth.user?.fullName }}.</h1>
    <p class="muted">Authentication is active. CRM modules will be introduced in approved implementation phases.</p>
    <div class="foundation-card"><i class="pi pi-check-circle" /><div><strong>Secure foundation ready</strong><p>PostgreSQL sessions, JWT access, and Administrator authorization are connected.</p></div></div>
  </section>
</template>
