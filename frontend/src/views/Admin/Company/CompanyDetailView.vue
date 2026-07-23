<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

const route = useRoute()
const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const company = ref<any>(null)

function computeCompany(id: string) {
  const all = store.allCustomers
  const rows = all.filter((c: any) => c.parentCode === id)
  if (!rows.length) return null

  const representative = rows[0]
  return {
    id,
    companyCode: id,
    name: representative.parentCompanyName || 'Unknown Company',
    registeredLocation: representative.region || 'Unknown',
    tier: 'Enterprise',
    npwp: representative.parentCode ? `0${representative.parentCode.slice(-3)}.000.678-9` : 'N/A',
    kam: representative.salesExecutiveName || 'Unassigned',
    sites: rows.length,
    status: 'Active',
    parentCompany: representative.parentCompanyName,
    locations: [...new Set(rows.map((r) => r.region || 'Unknown'))].join(', '),
    customers: rows
  }
}

onMounted(async () => {
  try {
    if (!store.allCustomers.length) {
      await store.fetchCustomers()
    }
    const id = route.params.id as string
    company.value = computeCompany(id)
    if (!company.value) {
      error.value = 'Company not found.'
    }
  } catch (e) {
    error.value = store.errorMessage(e)
  }
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
        <p class="eyebrow">Company Detail</p>
        <h1>{{ company?.name ?? 'Company Detail' }}</h1>
        <p class="muted">Parent company overview and related customer sites.</p>
      </div>
      <div style="display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap;">
        <Button label="Back to Companies" icon="pi pi-arrow-left" text size="small" @click="router.push('/admin/companies')" />
        <Button label="Edit Company" icon="pi pi-pencil" size="small" @click="router.push(`/admin/companies/${route.params.id}/edit`)" />
      </div>
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div v-if="company" class="grid-columns-2" style="gap:1rem;">
      <div class="section-card">
        <h3>Company Summary</h3>
        <div class="source-grid">
          <div><strong>Company Code</strong><span>{{ company.companyCode }}</span></div>
          <div><strong>Legal Name</strong><span>{{ company.name }}</span></div>
          <div><strong>Tier</strong><Tag :value="company.tier" /></div>
          <div><strong>Registered Location</strong><span>{{ company.registeredLocation }}</span></div>
          <div><strong>NPWP</strong><span>{{ company.npwp }}</span></div>
          <div><strong>Key Account Manager</strong><span>{{ company.kam }}</span></div>
          <div><strong>Status</strong><Tag :value="company.status" /></div>
          <div><strong>Customer Sites</strong><span>{{ company.sites }}</span></div>
        </div>
      </div>

      <div class="section-card">
        <h3>Related Customer Sites</h3>
        <div class="source-grid">
          <div><strong>Primary Location</strong><span>{{ company.locations }}</span></div>
          <div><strong>Company</strong><span>{{ company.parentCompany }}</span></div>
          <div class="full-width" style="grid-column:1/-1;">
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Site Name</th>
                  <th>Region</th>
                  <th>Segment</th>
                  <th>Sales Executive</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="site in company.customers" :key="site.id">
                  <td>{{ site.name }}</td>
                  <td>{{ site.region || '—' }}</td>
                  <td>{{ site.segment }}</td>
                  <td>{{ site.salesExecutiveName || '—' }}</td>
                  <td>
                    <Button icon="pi pi-eye" text rounded size="small" @click="router.push(`/admin/customers/${site.id}`)" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
