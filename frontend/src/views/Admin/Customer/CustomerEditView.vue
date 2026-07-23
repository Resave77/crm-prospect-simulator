<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'

const route = useRoute()
const router = useRouter()
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Edit Customer</p>
        <h1>Edit Customer Site</h1>
        <p class="muted">Update customer site information.</p>
      </div>
      <Button label="Back to List" icon="pi pi-arrow-left" text size="small" @click="router.push('/admin/customers')" />
    </div>
    <div class="section-card" style="min-height:300px;display:grid;place-content:center;justify-items:center;text-align:center;color:var(--text-muted);">
      <i class="pi pi-pencil" style="font-size:2rem;color:var(--brand-blue);" />
      <strong style="color:var(--text-primary);">Edit Form Coming Soon</strong>
      <span>Customer edit functionality will be implemented in a future update.</span>
    </div>
  </section>
</template>
