<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

const store = useCustomerListStore()
const router = useRouter()
const error = ref('')
const activeTab = ref('site')

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
</script>

<template>
  <section class="admin-page">
    <!-- PAGE HEADER -->
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Customer Management</span>
        <h1>Customer List</h1>
        <p class="muted">Manage customer sites, corporate hierarchies, and master data configurations.</p>
      </div>
      <div class="page-heading-actions">
        <Button label="Export" icon="pi pi-download" severity="secondary" outlined size="small" />
        <Button
          :label="activeTab === 'company' ? 'Add Company' : 'Add Customer'"
          icon="pi pi-plus"
          size="small"
          @click="activeTab === 'company' ? router.push('/admin/companies/add') : router.push('/admin/customers/add')"
        />
      </div>
    </header>

    <!-- SUMMARY CARDS -->
    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-icon si-blue">
          <i class="pi pi-building" />
        </div>
        <div class="summary-data">
          <span class="summary-label">Total Companies</span>
          <span class="summary-value">{{ companyCount }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-violet">
          <i class="pi pi-map-marker" />
        </div>
        <div class="summary-data">
          <span class="summary-label">Customer Sites</span>
          <span class="summary-value">{{ store.allCustomers.length }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-emerald">
          <i class="pi pi-user-check" />
        </div>
        <div class="summary-data">
          <span class="summary-label">Assigned Sites</span>
          <span class="summary-value">{{ assignedCount }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-amber">
          <i class="pi pi-clock" />
        </div>
        <div class="summary-data">
          <span class="summary-label">ERP Pending</span>
          <span class="summary-value">{{ erpPendingCount }}</span>
        </div>
      </div>
    </div>

    <!-- ERROR -->
    <Message v-if="error" severity="error">{{ error }}</Message>

    <!-- TABS -->
    <nav class="tabs-bar">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        :class="['tab-item', { active: activeTab === tab.key }]"
        @click="selectTab(tab.key)"
      >
        <i :class="tab.icon" />
        {{ tab.label }}
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
                <tr v-for="c in store.items" :key="c.id">
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
                      <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" />
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
              <tr v-for="company in companies" :key="company.id">
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
                    <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" />
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
/* ── PAGE ─────────────────────────────────────────────────────────── */
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.75rem 2rem;
  min-height: 100vh;
}

/* ── PAGE HEADER ──────────────────────────────────────────────────── */
.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}
.page-title-wrapper {
  display: flex;
  flex-direction: column;
}
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--brand-green-light, #0b7766);
  margin-bottom: 0.35rem;
}
.page-title-wrapper h1 {
  font-size: 1.65rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 0.2rem;
  letter-spacing: -0.03em;
  line-height: 1.15;
}
.page-title-wrapper .muted {
  font-size: 0.85rem;
  color: var(--text-muted);
  max-width: 520px;
  line-height: 1.55;
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}

/* ── SUMMARY CARDS ────────────────────────────────────────────────── */
.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}
.summary-card {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  padding: 1.1rem 1.2rem;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  transition: box-shadow var(--transition-fast), transform var(--transition-fast);
}
.summary-card:hover {
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}
.summary-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: grid;
  place-content: center;
  font-size: 1.1rem;
  flex-shrink: 0;
}
.si-blue { background: #eff6ff; color: #2563eb; }
.si-violet { background: #eef2ff; color: #6366f1; }
.si-emerald { background: #ecfdf5; color: #059669; }
.si-amber { background: #fffbeb; color: #d97706; }

.summary-data {
  display: flex;
  flex-direction: column;
}
.summary-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  line-height: 1.3;
}
.summary-value {
  font-size: 1.45rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.25;
}

/* ── TABS ─────────────────────────────────────────────────────────── */
.tabs-bar {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border-light);
  padding: 0 0.15rem;
}
.tab-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.7rem 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  cursor: pointer;
  transition: color var(--transition-fast), border-color var(--transition-fast);
  white-space: nowrap;
}
.tab-item i {
  font-size: 0.85rem;
}
.tab-item:hover {
  color: var(--text-primary);
}
.tab-item.active {
  color: var(--brand-blue);
  border-bottom-color: var(--brand-blue);
}

/* ── PANEL STACK ──────────────────────────────────────────────────── */
.panel-stack {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* ── FILTER PANEL ─────────────────────────────────────────────────── */
.filter-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.1rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.search-row {
  margin-bottom: 0.9rem;
}
.search-field {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.9rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), background var(--transition-fast);
}
.search-field:focus-within {
  background: var(--surface-card);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}
.search-field i {
  color: var(--text-faint);
  font-size: 0.9rem;
}
.search-field input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 0.87rem;
  color: var(--text-primary);
}
.search-field input::placeholder {
  color: var(--text-faint);
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr) auto;
  gap: 0.75rem;
  align-items: end;
}
.filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.filter-field label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.filter-action {
  justify-content: flex-end;
  padding-bottom: 0.15rem;
}

/* ── TABLE PANEL ──────────────────────────────────────────────────── */
.table-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  overflow: hidden;
}
.table-scroll {
  overflow-x: auto;
}

/* ── DATA TABLE ───────────────────────────────────────────────────── */
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}
.data-table thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--surface-subtle);
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.75rem 0.95rem;
  border-bottom: 1px solid var(--border-light);
  white-space: nowrap;
  text-align: left;
}
.data-table tbody td {
  padding: 0.75rem 0.95rem;
  border-bottom: 1px solid #f0f3f7;
  color: var(--text-primary);
  vertical-align: middle;
}
.data-table tbody tr:last-child td {
  border-bottom: none;
}
.data-table tbody tr {
  transition: background var(--transition-fast);
}
.data-table tbody tr:hover {
  background: #f8fafc;
}

.th-check, .td-check {
  width: 44px;
  text-align: center;
}
.th-action {
  width: 100px;
  text-align: center;
}

/* ── TABLE CELLS ──────────────────────────────────────────────────── */
.code-tag {
  display: inline-block;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: #f1f5f9;
  color: var(--text-secondary);
}
.code-blue {
  background: #eff6ff;
  color: #2563eb;
}

.link-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
  font: inherit;
  font-weight: 600;
  color: #2563eb;
  transition: color var(--transition-fast);
}
.link-btn:hover {
  color: #1d4ed8;
  text-decoration: underline;
}

.cell-stack {
  display: flex;
  flex-direction: column;
}
.cell-primary {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
}
.cell-sub {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}
.cell-text {
  font-size: 0.84rem;
  color: var(--text-secondary);
}
.cell-date {
  font-size: 0.8rem;
  color: var(--text-muted);
  white-space: nowrap;
}
.cell-badge {
  font-weight: 600;
  font-size: 0.78rem;
  color: var(--text-primary);
}

.site-status-cell {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.td-action {
  text-align: center;
}
.row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
}
.act-view {
  color: #2563eb !important;
}
.act-view:hover {
  background: #eff6ff !important;
}
.act-edit {
  color: #059669 !important;
}
.act-edit:hover {
  background: #ecfdf5 !important;
}
.act-delete {
  color: #dc2626 !important;
}
.act-delete:hover {
  background: #fef2f2 !important;
}

/* ── STATE BOXES (loading / empty) ────────────────────────────────── */
.state-box {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
}
.state-icon {
  font-size: 1.75rem;
  color: var(--brand-blue);
  margin-bottom: 0.25rem;
}
.state-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--surface-subtle);
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.state-icon-wrap i {
  font-size: 1.4rem;
  color: var(--text-faint);
}
.state-box strong {
  color: var(--text-primary);
  font-size: 0.95rem;
}

/* ── PAGINATION ───────────────────────────────────────────────────── */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.8rem 1.2rem;
  border-top: 1px solid var(--border-light);
  background: var(--surface-subtle);
}
.pagination-info {
  font-size: 0.8rem;
  color: var(--text-muted);
}
.pagination-info strong {
  color: var(--text-primary);
}
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.2rem;
}
.pagination-num {
  min-width: 32px;
  height: 32px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
}
.pagination-num:hover {
  background: var(--surface-hover) !important;
}
.pagination-num.is-active {
  background: var(--brand-blue) !important;
  color: #ffffff !important;
  font-weight: 700;
}
.pagination-dots {
  padding: 0 0.3rem;
  color: var(--text-faint);
  font-size: 0.85rem;
}

/* ── PLACEHOLDER (Master Data) ────────────────────────────────────── */
.placeholder-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  min-height: 340px;
  display: grid;
  place-content: center;
}
.placeholder-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  text-align: center;
  max-width: 340px;
}
.placeholder-icon {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-lg);
  background: var(--brand-blue-50, #eef5ff);
  display: grid;
  place-content: center;
  margin-bottom: 0.5rem;
}
.placeholder-icon i {
  font-size: 1.5rem;
  color: var(--brand-blue);
}
.placeholder-inner strong {
  color: var(--text-primary);
  font-size: 1rem;
}
.placeholder-inner span {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.55;
}

/* ── RESPONSIVE ───────────────────────────────────────────────────── */
@media (max-width: 1200px) {
  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 1024px) {
  .filter-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .filter-action {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }
}
@media (max-width: 768px) {
  .admin-page {
    padding: 1.25rem 1rem;
    gap: 1rem;
  }
  .page-heading {
    flex-direction: column;
    align-items: stretch;
  }
  .page-heading-actions {
    width: 100%;
    justify-content: flex-end;
  }
  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }
  .summary-card {
    padding: 0.9rem 1rem;
  }
  .summary-icon {
    width: 38px;
    height: 38px;
    font-size: 0.95rem;
  }
  .summary-value {
    font-size: 1.2rem;
  }
  .tabs-bar {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
  .tab-item {
    padding: 0.6rem 0.75rem;
    font-size: 0.8rem;
  }
  .filter-grid {
    grid-template-columns: 1fr 1fr;
  }
  .pagination-bar {
    flex-direction: column;
    gap: 0.6rem;
    padding: 1rem;
  }
}
@media (max-width: 480px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
  .filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
