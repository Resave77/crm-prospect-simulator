<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const error = ref('')
const customer = ref<any>(null)

onMounted(async () => {
  try {
    const id = route.params.id as string
    customer.value = await crm.loadAdminCustomer(id)
  } catch (e) { error.value = crm.errorMessage(e) }
})

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Customer Detail</p>
        <h1>{{ customer?.customer?.name ?? 'Customer Detail' }}</h1>
        <p class="muted">{{ customer?.customer?.customerCode }} &mdash; {{ customer?.parentCompany?.name }}</p>
      </div>
      <div style="display:flex;gap:0.5rem;">
        <Button label="Back to List" icon="pi pi-arrow-left" text size="small" @click="router.push('/admin/customers')" />
        <Button label="Edit" icon="pi pi-pencil" size="small" @click="router.push(`/admin/customers/${route.params.id}/edit`)" />
      </div>
    </div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="customer" style="display:grid;grid-template-columns:1fr 1fr;gap:1.25rem;">
      <div class="section-card">
        <h3 style="margin:0 0 1rem;font-size:1rem;">Customer Site</h3>
        <div class="source-grid">
          <div><strong>Name</strong><span>{{ customer.customer.name }}</span></div>
          <div><strong>Code</strong><span>{{ customer.customer.customerCode }}</span></div>
          <div><strong>Segment</strong><Tag :value="customer.customer.segment" /></div>
          <div><strong>Category</strong><span>{{ customer.customer.category }}</span></div>
          <div><strong>Region</strong><span>{{ customer.customer.region }}</span></div>
          <div><strong>Sales</strong><span>{{ customer.customer.salesExecutiveName }}</span></div>
          <div><strong>Converted</strong><span>{{ formatDate(customer.customer.convertedAt) }}</span></div>
          <div><strong>Address</strong><span>{{ customer.customer.address?.previewAddress }}</span></div>
        </div>
      </div>
      <div class="section-card">
        <h3 style="margin:0 0 1rem;font-size:1rem;">Parent Company</h3>
        <div class="source-grid">
          <div><strong>Name</strong><span>{{ customer.parentCompany.name }}</span></div>
          <div><strong>Code</strong><span>{{ customer.parentCompany.parentCode }}</span></div>
          <div><strong>Source Prospect</strong><span>{{ customer.sourceProspectName }}</span></div>
        </div>
      </div>
    </div>
  </section>
</template>
