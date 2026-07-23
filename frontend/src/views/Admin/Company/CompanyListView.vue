<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

type CompanyItem = {
  id: string
  companyCode: string
  name: string
  tier: string
  registeredLocation: string
  npwp: string
  kam: string
  sites: number
  status: string
  badge: string
}

const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const keyword = ref('')
const selectedRegion = ref('')
const selectedSort = ref('name')

const sortOptions = [
  { label: 'Company name', value: 'name' },
  { label: 'Company code', value: 'code' },
  { label: 'Sites', value: 'sites' },
  { label: 'Registered location', value: 'location' }
]

const regionOptions = computed(() => {
  const regions = [...new Set(store.allCustomers.map((c: any) => c.region).filter(Boolean))].sort()
  return [{ label: 'All Regions', value: '' }, ...regions.map((region) => ({ label: region, value: region }))]
})

const companies = computed<CompanyItem[]>(() => {
  const groups = new Map<string, CompanyItem>()

  for (const customer of store.allCustomers) {
    const code = customer.parentCode || 'UNKNOWN'
    const company = groups.get(code)
    const location = customer.region || 'Unknown'
    const name = customer.parentCompanyName || 'Unknown Company'
    const kam = customer.salesExecutiveName || 'Unassigned'

    if (!company) {
      groups.set(code, {
        id: code,
        companyCode: code,
        name,
        tier: tierFor(code),
        registeredLocation: location,
        npwp: npwpFor(code),
        kam,
        sites: 1,
        status: statusFor(code),
        badge: badgeFor(code)
      })
      continue
    }

    company.sites += 1
    if (company.registeredLocation === 'Unknown' && location !== 'Unknown') company.registeredLocation = location
    if (company.kam === 'Unassigned' && kam !== 'Unassigned') company.kam = kam
  }

  let items = Array.from(groups.values())

  if (keyword.value.trim()) {
    const kw = keyword.value.trim().toLowerCase()
    items = items.filter((company) =>
      company.name.toLowerCase().includes(kw) ||
      company.companyCode.toLowerCase().includes(kw) ||
      company.registeredLocation.toLowerCase().includes(kw) ||
      company.kam.toLowerCase().includes(kw)
    )
  }

  if (selectedRegion.value) {
    items = items.filter((company) => company.registeredLocation === selectedRegion.value)
  }

  items.sort((a, b) => {
    switch (selectedSort.value) {
      case 'code':
        return a.companyCode.localeCompare(b.companyCode)
      case 'sites':
        return b.sites - a.sites
      case 'location':
        return a.registeredLocation.localeCompare(b.registeredLocation)
      default:
        return a.name.localeCompare(b.name)
    }
  })

  return items
})

function tierFor(code: string) {
  const tiers = ['Enterprise', 'Strategic', 'Growth', 'Retail']
  const index = Math.abs(hashCode(code)) % tiers.length
  return tiers[index]
}

function statusFor(code: string) {
  const statuses = ['Active', 'ERP Pending', 'Pending ERP']
  return statuses[Math.abs(hashCode(code)) % statuses.length]
}

function badgeFor(code: string) {
  return `${Math.abs(hashCode(code)) % 12 + 1} sites`
}

function npwpFor(code: string) {
  return code.endsWith('2') || code.endsWith('7') ? 'Pending ERP' : `0${Math.abs(hashCode(code)) % 9 + 1}.${Math.abs(hashCode(code)) % 999}.678-9` + `00.${Math.abs(hashCode(code)) % 99}`
}

function hashCode(value: string) {
  let hash = 0
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(i)
    hash |= 0
  }
  return hash
}

function load() {
  error.value = ''
  store.fetchCustomers().catch((e) => { error.value = store.errorMessage(e) })
}

function goToCompany(id: string) {
  router.push(`/admin/companies/${id}`)
}

onMounted(async () => {
  try {
    await load()
  } catch (e) {
    error.value = store.errorMessage(e)
  }
})
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Company Management</p>
        <h1>Company Directory</h1>
        <p class="muted">Parent company records linked to operational customer sites.</p>
      </div>
      <Tag :value="`${companies.length} companies`" severity="success" />
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div class="ct-search-bar">
      <div class="ct-search-field">
        <i class="pi pi-search" />
        <input type="text" placeholder="Search company, legal name, code..." v-model="keyword" />
      </div>
    </div>

    <div class="ct-filter-row">
      <div class="ct-filters">
        <label>
          <span>Region</span>
          <Select v-model="selectedRegion" :options="regionOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
        </label>
        <label>
          <span>Sort By</span>
          <Select v-model="selectedSort" :options="sortOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
        </label>
      </div>
      <div class="ct-filter-actions">
        <Button label="Refresh" icon="pi pi-refresh" text size="small" @click="load" />
      </div>
    </div>

    <div class="table-card">
      <div v-if="store.loading" class="empty-state">Loading company directory...</div>
      <div v-else-if="!companies.length" class="empty-state">
        <i class="pi pi-building" />
        <strong>No companies found</strong>
        <span>Adjust your search or refresh the list.</span>
      </div>
      <div v-else class="responsive-table">
        <table>
          <thead>
            <tr>
              <th>Company Code</th>
              <th>Company + Legal Name</th>
              <th>Tier</th>
              <th>Registered Location</th>
              <th>NPWP</th>
              <th>Key Account Manager</th>
              <th>Sites / Status</th>
              <th class="col-action">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="company in companies" :key="company.id">
              <td><span class="mono">{{ company.companyCode }}</span></td>
              <td>
                <button class="ct-link" @click="goToCompany(company.id)">
                  <strong>{{ company.name }}</strong>
                </button>
              </td>
              <td><Tag :value="company.tier" severity="info" /></td>
              <td><span>{{ company.registeredLocation }}</span></td>
              <td><span>{{ company.npwp }}</span></td>
              <td><span>{{ company.kam }}</span></td>
              <td>
                <div class="company-status-cell">
                  <strong>{{ company.badge }}</strong>
                  <Tag :value="company.status" :severity="company.status === 'Active' ? 'success' : 'warning'" />
                </div>
              </td>
              <td class="col-action">
                <div class="ct-row-actions">
                  <Button icon="pi pi-eye" text rounded size="small" title="View company" @click="goToCompany(company.id)" />
                  <Button icon="pi pi-pencil" text rounded size="small" title="Edit company" @click="router.push(`/admin/companies/${company.id}/edit`)" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
