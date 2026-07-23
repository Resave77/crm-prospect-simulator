<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Card from 'primevue/card'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'

const crm = useCrmStore()
const error = ref('')
onMounted(async () => {
  try { await crm.loadMyCustomers() } catch (caught) { error.value = crm.errorMessage(caught) }
})
</script>

<template>
  <section class="mobile-page">
    <div class="page-heading compact-heading"><div><p class="eyebrow">Point 14</p><h1>My Customer</h1></div><Tag :value="`${crm.myCustomers.length} active`" severity="success" /></div>
    <p class="muted">Customer Existing records assigned to you after conversion.</p>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="crm.loading" class="empty-state">Loading customers…</div>
    <div v-else-if="!crm.myCustomers.length" class="empty-state"><i class="pi pi-users" /><strong>No Customer Existing yet</strong><span>Won prospects appear here after Administrator conversion.</span></div>
    <div v-else class="mobile-card-list">
      <Card v-for="customer in crm.myCustomers" :key="customer.id" class="customer-mobile-card">
        <template #title>{{ customer.name }}</template>
        <template #subtitle>{{ customer.customerCode }}</template>
        <template #content>
          <div class="detail-stack">
            <p><i class="pi pi-building" /> {{ customer.parentCompanyName }}</p>
            <p><i class="pi pi-map-marker" /> {{ customer.address.previewAddress }}</p>
            <div class="tag-row"><Tag :value="customer.segment" /><Tag :value="customer.category" severity="secondary" /></div>
            <RouterLink class="customer-detail-link" :to="`/sales/my-customers/${customer.id}`">View customer detail <i class="pi pi-arrow-right" /></RouterLink>
          </div>
        </template>
      </Card>
    </div>
  </section>
</template>
