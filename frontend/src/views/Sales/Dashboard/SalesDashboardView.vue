<script setup lang="ts">
import { useAuthStore } from '../../../stores/auth'
const auth = useAuthStore()
</script>

<template>
  <section class="dashboard-page sales-dashboard">
    <p class="eyebrow">Sales Executive</p>
    <h1>Hello, {{ auth.user?.fullName }}.</h1>
    <p class="muted">Your mobile field-sales workspace is securely connected.</p>
    <div class="foundation-card"><i class="pi pi-mobile" /><div><strong>Mobile workspace ready</strong><p>Assigned prospects and customers will appear here in the next business module phase.</p></div></div>
  </section>
</template>
