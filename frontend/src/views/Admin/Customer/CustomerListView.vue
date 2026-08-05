<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { useCustomerListStore } from '../../../stores/customerList'
import { deleteCustomer } from '../../../api/crm'

const store = useCustomerListStore()
const router = useRouter()
const error = ref('')
const activeTab = ref('site')
const deleteDialogVisible = ref(false)
const deleteTargetId = ref('')
const deleteTargetName = ref('')
const deleting = ref(false)
const companyDeleteDialogVisible = ref(false)
const companyDeleteTarget = ref<{ id: string; name: string; sites: number } | null>(null)

const tabs = [
  { key: 'site', label: 'Customer Site', icon: 'pi pi-map-marker' },
  { key: 'company', label: 'Company', icon: 'pi pi-building' },
  { key: 'master', label: 'Master Data', icon: 'pi pi-database' }
]

function selectTab(tabKey: string) {
  activeTab.value = tabKey
}

const sortOptions = [
  { label: 'Newest First', value: '' },
  { label: 'Oldest First', value: 'oldest' },
  { label: 'Customer Name', value: 'name' },
  { label: 'Customer Code', value: 'code' },
  { label: 'Converted Date', value: 'converted' },
  { label: 'Updated Date', value: 'updated' }
]

const segmentOptions = computed(() => {
  const segs = store.filterOptions?.segments ?? []
  return [{ label: 'All Segments', value: '' }, ...segs.map((s) => ({ label: s, value: s }))]
})
const categoryOptions = computed(() => {
  const cats = store.filterOptions?.categories ?? []
  return [{ label: 'All Categories', value: '' }, ...cats.map((c) => ({ label: c, value: c }))]
})
const regionOptions = computed(() => {
  const regs = store.filterOptions?.regions ?? []
  return [{ label: 'All Regions', value: '' }, ...regs.map((r) => ({ label: r, value: r }))]
})
const salesOptions = computed(() => {
  const sales = store.filterOptions?.salesExecutives ?? []
  return [{ label: 'All Sales', value: '' }, ...sales.map((s) => ({ label: s.fullName, value: s.fullName }))]
})

const selectedSort = computed({
  get: () => sortOptions.find((o) => o.value === store.params.sort) ?? sortOptions[0],
  set: (val) => store.setParam('sort', val.value)
})

const selectedSegment = computed({
  get: () => store.params.segment,
  set: (val) => { store.setParam('segment', val); store.setParam('page', 1) }
})
const selectedCategory = computed({
  get: () => store.params.category,
  set: (val) => { store.setParam('category', val); store.setParam('page', 1) }
})
const selectedRegion = computed({
  get: () => store.params.region,
  set: (val) => { store.setParam('region', val); store.setParam('page', 1) }
})
const selectedSales = computed({
  get: () => store.params.sales,
  set: (val) => { store.setParam('sales', val); store.setParam('page', 1) }
})

const companyKeyword = ref('')
const companySelectedRegion = ref('')
const companySelectedTier = ref('')
const companySelectedStatus = ref('')
const companySelectedLegal = ref('')
const companySelectedSort = ref('name')

const companySortOptions = [
  { label: 'Company name', value: 'name' },
  { label: 'Company code', value: 'code' },
  { label: 'Sites', value: 'sites' },
  { label: 'Registered location', value: 'location' }
]

const companyRegionOptions = computed(() => {
  const regions = [...new Set(store.allCustomers.map((c) => c.region).filter(Boolean))].sort()
  return [{ label: 'All Regions', value: '' }, ...regions.map((region) => ({ label: region, value: region }))]
})

const companyTierOptions = [
  { label: 'All Tiers', value: '' },
  { label: 'Tier 1', value: 'Tier 1' },
  { label: 'Tier 2', value: 'Tier 2' },
  { label: 'Tier 3', value: 'Tier 3' }
]

const companyStatusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'Active' },
  { label: 'Inactive', value: 'Inactive' }
]

const companyLegalOptions = [
  { label: 'All Legal Types', value: '' },
  { label: 'PT', value: 'PT' },
  { label: 'CV', value: 'CV' },
  { label: 'Firma', value: 'Firma' }
]

const companies = computed(() => {
  const groups = new Map<string, {
    id: string
    companyCode: string
    name: string
    tier: string
    registeredLocation: string
    npwp: string
    kam: string
    sites: number
    status: string
    legal: string
    badge: string
  }>()

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
        tier: 'Tier 1',
        registeredLocation: location,
        npwp: '00.000.000.0-000.000',
        kam,
        sites: 1,
        status: 'Active',
        legal: 'PT',
        badge: `${1} Site`
      })
      continue
    }

    company.sites += 1
    company.badge = `${company.sites} Sites`
    if (company.registeredLocation === 'Unknown' && location !== 'Unknown') company.registeredLocation = location
    if (company.kam === 'Unassigned' && kam !== 'Unassigned') company.kam = kam
  }

  let items = Array.from(groups.values())

  if (companyKeyword.value.trim()) {
    const kw = companyKeyword.value.trim().toLowerCase()
    items = items.filter((company) =>
      company.name.toLowerCase().includes(kw) ||
      company.companyCode.toLowerCase().includes(kw) ||
      company.registeredLocation.toLowerCase().includes(kw) ||
      company.kam.toLowerCase().includes(kw)
    )
  }

  if (companySelectedRegion.value) {
    items = items.filter((company) => company.registeredLocation === companySelectedRegion.value)
  }

  if (companySelectedTier.value) {
    items = items.filter((company) => company.tier === companySelectedTier.value)
  }

  if (companySelectedStatus.value) {
    items = items.filter((company) => company.status === companySelectedStatus.value)
  }

  if (companySelectedLegal.value) {
    items = items.filter((company) => company.legal === companySelectedLegal.value)
  }

  items.sort((a, b) => {
    switch (companySelectedSort.value) {
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

function goToCompany(id: string) {
  router.push(`/admin/companies/${id}`)
}

const companyCount = computed(() => new Set(store.allCustomers.map((c) => c.parentCode || 'UNKNOWN')).size)
const assignedCount = computed(() => store.allCustomers.filter((c) => c.salesExecutiveName).length)
const erpPendingCount = computed(() => companies.value.filter((c) => c.status !== 'Active').length)

function resetCompanyFilters() {
  companyKeyword.value = ''
  companySelectedRegion.value = ''
  companySelectedTier.value = ''
  companySelectedStatus.value = ''
  companySelectedLegal.value = ''
  companySelectedSort.value = 'name'
}

let searchTimeout: ReturnType<typeof setTimeout> | null = null
function onKeywordSearch(value: string) {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    store.setParam('keyword', value)
    store.setParam('page', 1)
    load()
  }, 350)
}

function load() {
  error.value = ''
  store.fetchCustomers().catch((e) => { error.value = store.errorMessage(e) })
}

function resetAll() {
  store.resetFilters()
  load()
}

function goToPage(p: number) {
  store.setPage(p)
  load()
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function segmentSeverity(seg: string) {
  switch (seg) {
    case 'Key Account': return 'warn'
    case 'Modern Trade': return 'info'
    case 'Food Service': return 'success'
    default: return 'secondary'
  }
}

function getPageNumbers(): (number | '...')[] {
  const total = store.pages
  const current = store.page
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | '...')[] = [1]
  if (current > 3) pages.push('...')
  for (let i = Math.max(2, current - 1); i <= Math.min(total - 1, current + 1); i++) {
    pages.push(i)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
}

onMounted(async () => {
  try {
    await store.fetchFilterOptions()
    await load()
  } catch (e) { error.value = store.errorMessage(e) }
})

function confirmDelete(id: string, name: string) {
  deleteTargetId.value = id
  deleteTargetName.value = name
  deleteDialogVisible.value = true
}

async function executeDelete() {
  deleting.value = true
  try {
    await deleteCustomer(deleteTargetId.value)
    store.allCustomers = store.allCustomers.filter((c) => c.id !== deleteTargetId.value)
    store.items = store.items.filter((c) => c.id !== deleteTargetId.value)
    store.total = store.allCustomers.length
    deleteDialogVisible.value = false
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    deleting.value = false
  }
}

function confirmDeleteCompany(company: { id: string; name: string; sites: number }) {
  companyDeleteTarget.value = company
  companyDeleteDialogVisible.value = true
}

async function executeDeleteCompany() {
  if (!companyDeleteTarget.value) return
  deleting.value = true
  try {
    const companyId = companyDeleteTarget.value.id
    const companySites = store.allCustomers.filter((c) => c.parentCode === companyId)
    for (const site of companySites) {
      await deleteCustomer(site.id)
    }
    store.allCustomers = store.allCustomers.filter((c) => c.parentCode !== companyId)
    store.items = store.items.filter((c) => c.parentCode !== companyId)
    store.total = store.allCustomers.length
    companyDeleteDialogVisible.value = false
  } catch (e) {
    error.value = store.errorMessage(e)
    companyDeleteDialogVisible.value = false
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <section class="admin-page">
    <header class="workspace-header">
      <div class="workspace-heading">
        <Button icon="pi pi-arrow-left" severity="secondary" text rounded class="back-button" @click="router.back()" title="Back" />
        <div class="page-title-wrapper">
          <span class="eyebrow">Customer Management</span>
          <h1>Customers & Companies</h1>
          <p class="muted">Manage customer sites, corporate accounts, ownership, and reference data.</p>
        </div>
      </div>
      <div class="workspace-summary">
        <button type="button" class="summary-item" @click="selectTab('company')"><i class="pi pi-building si-blue" /><span>Companies</span><strong>{{ companyCount }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('site')"><i class="pi pi-map-marker si-violet" /><span>Sites</span><strong>{{ store.allCustomers.length }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('site')"><i class="pi pi-user-check si-emerald" /><span>Assigned</span><strong>{{ assignedCount }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('company')"><i class="pi pi-clock si-amber" /><span>ERP Pending</span><strong>{{ erpPendingCount }}</strong></button>
      </div>
      <div class="page-heading-actions">
        <Button label="Export" icon="pi pi-download" severity="secondary" outlined size="small" />
        <Button :label="activeTab === 'company' ? 'Add Company' : 'Add Customer'" icon="pi pi-plus" size="small" @click="activeTab === 'company' ? router.push('/admin/companies/add') : router.push('/admin/customers/add')" />
      </div>
    </header>
    <Message v-if="error" severity="error" class="page-message">{{ error }}</Message>
    <nav class="tabs-bar" aria-label="Customer management sections">
      <button v-for="tab in tabs" :key="tab.key" type="button" :class="['tab-item', { active: activeTab === tab.key }]" @click="selectTab(tab.key)">
        <i :class="tab.icon" /><span>{{ tab.label }}</span><strong v-if="tab.key === 'site'">{{ store.allCustomers.length }}</strong><strong v-else-if="tab.key === 'company'">{{ companyCount }}</strong>
      </button>
    </nav>

    <!-- ======================== CUSTOMER SITE TAB ======================== -->
    <template v-if="activeTab === 'site'">
      <div class="panel-stack">
        <!-- FILTERS -->
        <div class="filter-panel">
          <div class="search-row">
            <div class="search-field">
              <i class="pi pi-search" />
              <input
                type="text"
                placeholder="Search by customer name, code, company..."
                :value="store.params.keyword"
                @input="onKeywordSearch(($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>

          <div class="filter-grid">
            <div class="filter-field">
              <label>Region</label>
              <Select v-model="selectedRegion" :options="regionOptions" optionLabel="label" optionValue="value" placeholder="All Regions" />
            </div>
            <div class="filter-field">
              <label>Segment</label>
              <Select v-model="selectedSegment" :options="segmentOptions" optionLabel="label" optionValue="value" placeholder="All Segments" />
            </div>
            <div class="filter-field">
              <label>Category</label>
              <Select v-model="selectedCategory" :options="categoryOptions" optionLabel="label" optionValue="value" placeholder="All Categories" />
            </div>
            <div class="filter-field">
              <label>Sales Executive</label>
              <Select v-model="selectedSales" :options="salesOptions" optionLabel="label" optionValue="value" placeholder="All Sales" />
            </div>
            <div class="filter-field">
              <label>Sort By</label>
              <Select v-model="selectedSort" :options="sortOptions" optionLabel="label" />
            </div>
            <div class="filter-field filter-action">
              <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetAll" />
            </div>
          </div>
        </div>

        <!-- TABLE -->
        <div class="table-panel">
          <!-- Loading -->
          <div v-if="store.loading" class="state-box">
            <i class="pi pi-spin pi-spinner state-icon" />
            <span>Loading customer sites...</span>
          </div>

          <!-- Empty -->
          <div v-else-if="!store.items.length" class="state-box">
            <div class="state-icon-wrap">
              <i class="pi pi-users" />
            </div>
            <strong>No customer sites found</strong>
            <span class="muted">Adjust your search or filters to view results.</span>
          </div>

          <!-- Table -->
          <div v-else class="table-scroll">
            <table class="data-table">
              <thead>
                <tr>
                  <th class="th-check">
                    <input type="checkbox" :checked="store.isAllSelected()" @change="store.toggleSelectAll()" />
                  </th>
                  <th>Code</th>
                  <th>Customer Site</th>
                  <th>Parent Company</th>
                  <th>Region</th>
                  <th>Segment</th>
                  <th>Category</th>
                  <th>Sales Executive</th>
                  <th>Converted</th>
                  <th class="th-action">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="c in store.items" :key="c.id" class="clickable-row" @dblclick="router.push(`/admin/customers/${c.id}`)">
                  <td class="td-check">
                    <input type="checkbox" :checked="store.selectedIds.has(c.id)" @change="store.toggleSelect(c.id)" />
                  </td>
                  <td>
                    <code class="code-tag code-blue">{{ c.customerCode }}</code>
                  </td>
                  <td>
                    <button class="link-btn" @click="router.push(`/admin/customers/${c.id}`)">
                      {{ c.name }}
                    </button>
                  </td>
                  <td>
                    <div class="cell-stack">
                      <span class="cell-primary">{{ c.parentCompanyName }}</span>
                      <span class="cell-sub">{{ c.parentCode }}</span>
                    </div>
                  </td>
                  <td>
                    <span class="cell-text">{{ c.region || '—' }}</span>
                  </td>
                  <td>
                    <Tag :value="c.segment" :severity="segmentSeverity(c.segment)" />
                  </td>
                  <td>
                    <span class="cell-text">{{ c.category }}</span>
                  </td>
                  <td>
                    <span class="cell-text">{{ c.salesExecutiveName || 'Unassigned' }}</span>
                  </td>
                  <td>
                    <span class="cell-date">{{ formatDate(c.convertedAt) }}</span>
                  </td>
                  <td class="td-action">
                    <div class="row-actions">
                      <Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View" @click="router.push(`/admin/customers/${c.id}`)" />
                      <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="router.push(`/admin/customers/${c.id}/edit`)" />
                      <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete(c.id, c.name)" />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- PAGINATION -->
          <div v-if="store.pages > 1" class="pagination-bar">
            <span class="pagination-info">
              Showing <strong>{{ (store.page - 1) * store.limit + 1 }}</strong>–<strong>{{ Math.min(store.page * store.limit, store.total) }}</strong> of <strong>{{ store.total }}</strong>
            </span>
            <div class="pagination-controls">
              <Button icon="pi pi-angle-left" text rounded size="small" :disabled="store.page <= 1" @click="goToPage(store.page - 1)" />
              <template v-for="(p, idx) in getPageNumbers()" :key="idx">
                <span v-if="p === '...'" class="pagination-dots">…</span>
                <Button
                  v-else
                  :label="String(p)"
                  text rounded size="small"
                  :class="['pagination-num', { 'is-active': p === store.page }]"
                  @click="goToPage(p as number)"
                />
              </template>
              <Button icon="pi pi-angle-right" text rounded size="small" :disabled="store.page >= store.pages" @click="goToPage(store.page + 1)" />
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- DELETE CONFIRMATION DIALOG -->
    <Dialog v-model:visible="deleteDialogVisible" header="Delete Customer" modal :style="{ width: '400px' }">
      <p>Are you sure you want to delete <strong>{{ deleteTargetName }}</strong>? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>

    <!-- DELETE COMPANY CONFIRMATION DIALOG -->
    <Dialog v-model:visible="companyDeleteDialogVisible" header="Delete Company" modal :style="{ width: '420px' }">
      <p>Are you sure you want to delete <strong>{{ companyDeleteTarget?.name }}</strong> and all <strong>{{ companyDeleteTarget?.sites }}</strong> of its site(s)? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="companyDeleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDeleteCompany" />
      </template>
    </Dialog>

    <!-- ======================== COMPANY TAB ======================== -->
    <div v-if="activeTab === 'company'" class="panel-stack">
      <!-- FILTERS -->
      <div class="filter-panel">
        <div class="search-row">
          <div class="search-field">
            <i class="pi pi-search" />
            <input type="text" placeholder="Search company by name, code, location..." v-model="companyKeyword" />
          </div>
        </div>

        <div class="filter-grid">
          <div class="filter-field">
            <label>Legal Type</label>
            <Select v-model="companySelectedLegal" :options="companyLegalOptions" optionLabel="label" optionValue="value" placeholder="All Legal Types" />
          </div>
          <div class="filter-field">
            <label>Tier</label>
            <Select v-model="companySelectedTier" :options="companyTierOptions" optionLabel="label" optionValue="value" placeholder="All Tiers" />
          </div>
          <div class="filter-field">
            <label>Region</label>
            <Select v-model="companySelectedRegion" :options="companyRegionOptions" optionLabel="label" optionValue="value" placeholder="All Regions" />
          </div>
          <div class="filter-field">
            <label>Status</label>
            <Select v-model="companySelectedStatus" :options="companyStatusOptions" optionLabel="label" optionValue="value" placeholder="All Status" />
          </div>
          <div class="filter-field">
            <label>Sort By</label>
            <Select v-model="companySelectedSort" :options="companySortOptions" optionLabel="label" optionValue="value" />
          </div>
          <div class="filter-field filter-action">
            <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetCompanyFilters" />
          </div>
        </div>
      </div>

      <!-- TABLE -->
      <div class="table-panel">
        <div v-if="store.loading" class="state-box">
          <i class="pi pi-spin pi-spinner state-icon" />
          <span>Loading companies...</span>
        </div>
        <div v-else-if="!companies.length" class="state-box">
          <div class="state-icon-wrap">
            <i class="pi pi-building" />
          </div>
          <strong>No companies found</strong>
          <span class="muted">Try modifying your filters to find company accounts.</span>
        </div>
        <div v-else class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th>Code</th>
                <th>Company Name</th>
                <th>Legal</th>
                <th>Tier</th>
                <th>Region</th>
                <th>NPWP</th>
                <th>KAM</th>
                <th>Sites</th>
                <th class="th-action">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="company in companies" :key="company.id" class="clickable-row" @dblclick="goToCompany(company.id)">
                <td>
                  <code class="code-tag">{{ company.companyCode }}</code>
                </td>
                <td>
                  <button class="link-btn" @click="goToCompany(company.id)">
                    {{ company.name }}
                  </button>
                </td>
                <td>
                  <span class="cell-badge">{{ company.legal }}</span>
                </td>
                <td>
                  <Tag :value="company.tier" severity="info" />
                </td>
                <td>
                  <span class="cell-text">{{ company.registeredLocation }}</span>
                </td>
                <td>
                  <code class="code-tag">{{ company.npwp }}</code>
                </td>
                <td>
                  <span class="cell-text">{{ company.kam }}</span>
                </td>
                <td>
                  <div class="site-status-cell">
                    <span class="cell-badge">{{ company.badge }}</span>
                    <Tag :value="company.status" :severity="company.status === 'Active' ? 'success' : 'warn'" size="small" />
                  </div>
                </td>
                <td class="td-action">
                  <div class="row-actions">
                    <Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View" @click="goToCompany(company.id)" />
                    <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="router.push(`/admin/companies/${company.id}/edit`)" />
                    <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDeleteCompany(company)" />
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ======================== MASTER DATA TAB ======================== -->
    <div v-if="activeTab === 'master'" class="placeholder-panel">
      <div class="placeholder-inner">
        <div class="placeholder-icon">
          <i class="pi pi-database" />
        </div>
        <strong>Master Data Management</strong>
        <span>Segment configurations, categories, and reference data will be available here soon.</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.admin-page{box-sizing:border-box;display:flex;min-width:0;min-height:calc(100dvh - 4rem);flex-direction:column;gap:.75rem;padding:.85rem 1rem 1rem;overflow-x:hidden;background:#f8fafc}.workspace-header{display:grid;grid-template-columns:minmax(260px,1fr) auto auto;align-items:center;gap:.9rem;padding:.72rem .85rem;border:1px solid #e5eaf0;border-radius:12px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.workspace-heading{display:flex;min-width:0;align-items:center;gap:.65rem}.back-button{flex:0 0 auto}.page-title-wrapper{min-width:0;display:grid;gap:.04rem}.page-title-wrapper .eyebrow{color:#64748b;font-size:.56rem;font-weight:800;letter-spacing:.08em;text-transform:uppercase}.page-title-wrapper h1{margin:0;color:#0f172a;font-size:1.08rem;line-height:1.2;letter-spacing:-.02em}.page-title-wrapper .muted{margin:0;overflow:hidden;color:#94a3b8;font-size:.67rem;line-height:1.4;text-overflow:ellipsis;white-space:nowrap}.workspace-summary{display:grid;grid-template-columns:repeat(4,minmax(74px,auto));overflow:hidden;border:1px solid #e5eaf0;border-radius:9px;background:#f8fafc}.summary-item{display:grid;grid-template-columns:18px auto;grid-template-rows:auto auto;column-gap:.3rem;min-width:78px;padding:.4rem .55rem;border:0;border-right:1px solid #e5eaf0;background:transparent;text-align:left;cursor:pointer}.summary-item:last-child{border-right:0}.summary-item:hover{background:#eff6ff}.summary-item i{grid-row:1/3;align-self:center;font-size:.74rem}.summary-item span{color:#94a3b8;font-size:.5rem;font-weight:800;text-transform:uppercase}.summary-item strong{color:#0f172a;font-size:.78rem}.si-blue{color:#2563eb}.si-violet{color:#6366f1}.si-emerald{color:#059669}.si-amber{color:#d97706}.page-heading-actions{display:flex;align-items:center;justify-content:flex-end;gap:.45rem}.page-message{margin:0}.tabs-bar{display:flex;min-width:0;gap:.35rem;padding:.32rem;overflow-x:auto;border:1px solid #e5eaf0;border-radius:10px;background:#fff;scrollbar-width:none}.tabs-bar::-webkit-scrollbar{display:none}.tab-item{display:inline-flex;min-height:34px;align-items:center;gap:.42rem;padding:.42rem .72rem;border:1px solid transparent;border-radius:8px;background:transparent;color:#64748b;font-size:.72rem;font-weight:700;white-space:nowrap;cursor:pointer}.tab-item:hover{background:#f8fafc;color:#0f172a}.tab-item.active{border-color:#bfdbfe;background:#eff6ff;color:#1d4ed8}.tab-item strong{min-width:20px;padding:.08rem .35rem;border-radius:999px;background:rgba(148,163,184,.14);font-size:.58rem;text-align:center}.tab-item.active strong{background:#dbeafe}.panel-stack{display:flex;min-width:0;flex-direction:column;gap:.65rem}.filter-panel{display:grid;grid-template-columns:minmax(250px,1.35fr) minmax(0,3.65fr);align-items:end;gap:.7rem;padding:.65rem;border:1px solid #e5eaf0;border-radius:10px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.search-row{min-width:0;margin:0}.search-field{display:flex;min-height:38px;align-items:center;gap:.55rem;padding:.45rem .7rem;border:1px solid #dbe3ee;border-radius:8px;background:#f8fafc}.search-field:focus-within{border-color:#2563eb;background:#fff;box-shadow:0 0 0 3px rgba(37,99,235,.08)}.search-field i{color:#94a3b8;font-size:.76rem}.search-field input{width:100%;min-width:0;border:0;outline:0;background:transparent;color:#0f172a;font-size:.76rem}.search-field input::placeholder{color:#94a3b8}.filter-grid{display:grid;grid-template-columns:repeat(5,minmax(110px,1fr)) auto;align-items:end;gap:.5rem;min-width:0}.filter-field{display:grid;min-width:0;gap:.22rem}.filter-field label{color:#64748b;font-size:.55rem;font-weight:800;letter-spacing:.05em;text-transform:uppercase}.filter-field :deep(.p-select){min-width:0;height:38px;border-radius:8px;font-size:.72rem}.filter-field :deep(.p-select-label){padding-block:.5rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.filter-action{display:flex;align-items:flex-end;justify-content:flex-end;padding-bottom:.02rem}.table-panel{min-width:0;overflow:hidden;border:1px solid #e5eaf0;border-radius:10px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.table-scroll{width:100%;min-width:0;overflow-x:auto;scrollbar-width:thin}.data-table{width:100%;min-width:980px;table-layout:fixed;border-collapse:collapse;font-size:.74rem}.data-table thead th{position:sticky;top:0;z-index:2;padding:.58rem .65rem;border-bottom:1px solid #e5eaf0;background:#f8fafc;color:#64748b;font-size:.58rem;font-weight:800;letter-spacing:.05em;text-align:left;text-transform:uppercase;white-space:nowrap}.data-table tbody td{height:48px;padding:.48rem .65rem;overflow:hidden;border-bottom:1px solid #edf1f6;color:#1e293b;text-overflow:ellipsis;vertical-align:middle}.data-table tbody tr:last-child td{border-bottom:0}.data-table tbody tr:hover{background:#f8fbff}.clickable-row{cursor:default}.th-check,.td-check{width:42px;text-align:center}.th-action{width:112px;text-align:center}.code-tag{display:inline-block;max-width:100%;overflow:hidden;padding:.14rem .42rem;border-radius:5px;background:#f1f5f9;color:#475569;font-family:'SF Mono','Fira Code',Consolas,monospace;font-size:.66rem;font-weight:650;text-overflow:ellipsis;white-space:nowrap}.code-blue{background:#eff6ff;color:#2563eb}.link-btn{max-width:100%;overflow:hidden;border:0;background:transparent;color:#1d4ed8;font:inherit;font-weight:750;text-align:left;text-overflow:ellipsis;white-space:nowrap;cursor:pointer}.link-btn:hover{text-decoration:underline;text-underline-offset:2px}.cell-stack{display:grid;min-width:0;gap:.04rem}.cell-primary,.cell-text,.cell-date,.cell-badge{display:block;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.cell-primary{color:#0f172a;font-size:.73rem;font-weight:700}.cell-sub{color:#94a3b8;font-family:Consolas,monospace;font-size:.6rem}.cell-text{color:#475569;font-size:.7rem}.cell-date{color:#64748b;font-size:.68rem}.cell-badge{color:#334155;font-size:.68rem;font-weight:700}.site-status-cell{display:flex;min-width:0;align-items:center;gap:.3rem}.td-action{text-align:center}.row-actions{display:inline-flex;align-items:center;justify-content:center;gap:.12rem;padding:.12rem;border:1px solid #e5eaf0;border-radius:999px;background:#fff}.row-actions :deep(.p-button){width:1.85rem;height:1.85rem}.act-view{color:#2563eb!important}.act-view:hover{background:#eff6ff!important}.act-edit{color:#059669!important}.act-edit:hover{background:#ecfdf5!important}.act-delete{color:#dc2626!important}.act-delete:hover{background:#fef2f2!important}.state-box{min-height:240px;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:.4rem;padding:2rem;color:#64748b;text-align:center}.state-icon{color:#2563eb;font-size:1.5rem}.state-icon-wrap{display:grid;width:50px;height:50px;place-content:center;border-radius:13px;background:#f1f5f9;color:#94a3b8}.state-box strong{color:#0f172a;font-size:.86rem}.pagination-bar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;padding:.58rem .75rem;border-top:1px solid #e5eaf0;background:#f8fafc}.pagination-info{color:#64748b;font-size:.67rem}.pagination-info strong{color:#0f172a}.pagination-controls{display:flex;align-items:center;gap:.12rem}.pagination-num{min-width:28px;height:28px;color:#475569;font-size:.68rem;font-weight:700}.pagination-num.is-active{background:#2563eb!important;color:#fff!important}.pagination-dots{padding:0 .2rem;color:#94a3b8}.placeholder-panel{min-height:320px;display:grid;place-content:center;border:1px solid #e5eaf0;border-radius:10px;background:#fff}.placeholder-inner{display:flex;max-width:360px;flex-direction:column;align-items:center;gap:.45rem;color:#64748b;text-align:center}.placeholder-icon{display:grid;width:58px;height:58px;place-content:center;margin-bottom:.25rem;border-radius:15px;background:#eff6ff;color:#2563eb}.placeholder-inner strong{color:#0f172a;font-size:.9rem}.placeholder-inner span{font-size:.75rem;line-height:1.5}@media(max-width:1280px){.workspace-header{grid-template-columns:minmax(240px,1fr) auto}.workspace-summary{grid-column:1/-1;grid-row:2}.filter-panel{grid-template-columns:1fr}.filter-grid{grid-template-columns:repeat(3,minmax(130px,1fr)) auto}}@media(max-width:900px){.admin-page{min-height:auto;padding:.75rem;overflow:visible}.workspace-header{grid-template-columns:1fr auto}.workspace-summary{grid-template-columns:repeat(4,1fr);width:100%}.page-title-wrapper .muted{white-space:normal}.filter-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.filter-action{grid-column:1/-1}.data-table{min-width:920px}}@media(max-width:640px){.admin-page{padding:.6rem;gap:.6rem}.workspace-header{grid-template-columns:1fr;align-items:stretch}.workspace-heading{align-items:flex-start}.page-heading-actions{display:grid;grid-template-columns:1fr 1fr}.page-heading-actions :deep(.p-button){width:100%}.workspace-summary{grid-template-columns:repeat(2,1fr)}.summary-item:nth-child(2){border-right:0}.summary-item:nth-child(-n+2){border-bottom:1px solid #e5eaf0}.filter-panel{padding:.55rem}.filter-grid{grid-template-columns:1fr}.filter-action{grid-column:auto}.pagination-bar{flex-direction:column;align-items:stretch}.pagination-controls{justify-content:center}.data-table{min-width:860px}}
</style>

