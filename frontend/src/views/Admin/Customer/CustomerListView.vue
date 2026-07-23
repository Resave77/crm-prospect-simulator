<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'

const crm = useCrmStore()
const error = ref('')
onMounted(async () => {
  try { await crm.loadAdminCustomers() } catch (caught) { error.value = crm.errorMessage(caught) }
})
</script>

<template>
  <section class="admin-page">
    <div class="page-heading"><div><p class="eyebrow">Point 14</p><h1>Customer Existing</h1><p class="muted">Converted Customer Sites and their Parent Companies.</p></div><Tag :value="`${crm.adminCustomers.length} customers`" /></div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div class="table-card responsive-table">
      <div v-if="crm.loading" class="empty-state">Loading Customer Existing…</div>
      <div v-else-if="!crm.adminCustomers.length" class="empty-state"><i class="pi pi-users" /><strong>No converted customers</strong><span>Complete a Won Prospect conversion to create the first record.</span></div>
      <table v-else><thead><tr><th>Customer Site</th><th>Parent Company</th><th>Segment / Category</th><th>Sales Executive</th><th>Converted</th></tr></thead><tbody><tr v-for="customer in crm.adminCustomers" :key="customer.id"><td><strong>{{ customer.name }}</strong><span>{{ customer.customerCode }}</span></td><td><strong>{{ customer.parentCompanyName }}</strong><span>{{ customer.parentCode }}</span></td><td><Tag :value="customer.segment" /><span>{{ customer.category }}</span></td><td>{{ customer.salesExecutiveName }}</td><td>{{ new Date(customer.convertedAt).toLocaleString() }}</td></tr></tbody></table>
    </div>
  </section>
</template>
