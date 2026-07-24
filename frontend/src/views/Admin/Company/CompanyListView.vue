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

const companyKeyword = ref('')
const companySelectedRegion = ref('')
const companySelectedTier = ref('')
const companySelectedStatus = ref('')
const companySelectedSort = ref('name')

const companySortOptions = [
  { label: 'Company Name', value: 'name' },
  { label: 'Company Code', value: 'code' },
  { label: 'Most Sites', value: 'sites' },
  { label: 'Region', value: 'location' }
]

const companyRegionOptions = computed(() => {
  const regions = [...new Set(store.allCustomers.map((c) => c.region).filter(Boolean))].sort()
  return [{ label: 'All Regions', value: '' }, ...regions.map((r) => ({ label: r, value: r }))]
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

const companies = computed(() => {
  const groups = new Map<string, {
    id: string
    companyCode: string
    name: string
    tier: string
    region: string
    npwp: string
    kam: string
    sites: number
    status: string
  }>()

  for (const customer of store.allCustomers) {
    const code = customer.parentCode || 'UNKNOWN'
    const company = groups.get(code)
    const region = customer.region || 'Unknown'
    const name = customer.parentCompanyName || 'Unknown Company'
    const kam = customer.salesExecutiveName || 'Unassigned'

    if (!company) {
      groups.set(code, {
        id: code,
        companyCode: code,
        name,
        tier: 'Tier 1',
        region,
        npwp: '00.000.000.0-000.000',
        kam,
        sites: 1,
        status: 'Active'
      })
      continue
    }

    company.sites += 1
    if (company.region === 'Unknown' && region !== 'Unknown') company.region = region
    if (company.kam === 'Unassigned' && kam !== 'Unassigned') company.kam = kam
  }

  let items = Array.from(groups.values())

  if (companyKeyword.value.trim()) {
    const kw = companyKeyword.value.trim().toLowerCase()
    items = items.filter((c) =>
      c.name.toLowerCase().includes(kw) ||
      c.companyCode.toLowerCase().includes(kw) ||
      c.kam.toLowerCase().includes(kw)
    )
  }
  if (companySelectedRegion.value) items = items.filter((c) => c.region === companySelectedRegion.value)
  if (companySelectedTier.value) items = items.filter((c) => c.tier === companySelectedTier.value)
  if (companySelectedStatus.value) items = items.filter((c) => c.status === companySelectedStatus.value)

  items.sort((a, b) => {
    switch (companySelectedSort.value) {
      case 'code': return a.companyCode.localeCompare(b.companyCode)
      case 'sites': return b.sites - a.sites
      case 'location': return a.region.localeCompare(b.region)
      default: return a.name.localeCompare(b.name)
    }
  })

  return items
})

const totalSites = computed(() => store.allCustomers.length)
const activeCompanies = computed(() => companies.value.filter((c) => c.status === 'Active').length)

function resetFilters() {
  companyKeyword.value = ''
  companySelectedRegion.value = ''
  companySelectedTier.value = ''
  companySelectedStatus.value = ''
  companySelectedSort.value = 'name'
}

onMounted(async () => {
  try {
    if (store.allCustomers.length === 0) await store.fetchCustomers()
  } catch (e) { error.value = store.errorMessage(e) }
})
</script>

<template>
  <section class="admin-page">
    <!-- PAGE HEADER -->
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Corporate Directory</span>
        <h1>Company List</h1>
        <p class="muted">View and manage parent companies, their associated sites, and corporate hierarchies.</p>
      </div>
      <div class="page-heading-actions">
        <Button label="Export" icon="pi pi-download" severity="secondary" outlined size="small" />
        <Button label="Add Company" icon="pi pi-plus" size="small" @click="router.push('/admin/companies/add')" />
      </div>
    </header>

    <!-- SUMMARY CARDS -->
    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-icon si-blue"><i class="pi pi-building" /></div>
        <div class="summary-data">
          <span class="summary-label">Total Companies</span>
          <span class="summary-value">{{ companies.length }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-violet"><i class="pi pi-map-marker" /></div>
        <div class="summary-data">
          <span class="summary-label">Total Sites</span>
          <span class="summary-value">{{ totalSites }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-emerald"><i class="pi pi-check-circle" /></div>
        <div class="summary-data">
          <span class="summary-label">Active</span>
          <span class="summary-value">{{ activeCompanies }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-amber"><i class="pi pi-shield" /></div>
        <div class="summary-data">
          <span class="summary-label">Tier 1</span>
          <span class="summary-value">{{ companies.filter((c) => c.tier === 'Tier 1').length }}</span>
        </div>
      </div>
    </div>

    <!-- ERROR -->
    <Message v-if="error" severity="error">{{ error }}</Message>

    <!-- FILTERS -->
    <div class="filter-panel">
      <div class="search-row">
        <div class="search-field">
          <i class="pi pi-search" />
          <input type="text" placeholder="Search company by name, code, KAM..." v-model="companyKeyword" />
        </div>
      </div>
      <div class="filter-grid">
        <div class="filter-field">
          <label>Tier</label>
          <Select v-model="companySelectedTier" :options="companyTierOptions" optionLabel="label" optionValue="value" />
        </div>
        <div class="filter-field">
          <label>Region</label>
          <Select v-model="companySelectedRegion" :options="companyRegionOptions" optionLabel="label" optionValue="value" />
        </div>
        <div class="filter-field">
          <label>Status</label>
          <Select v-model="companySelectedStatus" :options="companyStatusOptions" optionLabel="label" optionValue="value" />
        </div>
        <div class="filter-field">
          <label>Sort By</label>
          <Select v-model="companySelectedSort" :options="companySortOptions" optionLabel="label" optionValue="value" />
        </div>
        <div class="filter-field filter-action">
          <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetFilters" />
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
        <div class="state-icon-wrap"><i class="pi pi-building" /></div>
        <strong>No companies found</strong>
        <span class="muted">Try modifying your filters or add a new company.</span>
      </div>
      <div v-else class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>Code</th>
              <th>Company Name</th>
              <th>Tier</th>
              <th>Region</th>
              <th>NPWP</th>
              <th>KAM</th>
              <th>Sites</th>
              <th>Status</th>
              <th class="th-action">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="company in companies" :key="company.id">
              <td><code class="code-tag">{{ company.companyCode }}</code></td>
              <td>
                <button class="link-btn" @click="router.push(`/admin/companies/${company.id}`)">
                  {{ company.name }}
                </button>
              </td>
              <td><Tag :value="company.tier" severity="info" /></td>
              <td><span class="cell-text">{{ company.region }}</span></td>
              <td><code class="code-tag">{{ company.npwp }}</code></td>
              <td><span class="cell-text">{{ company.kam }}</span></td>
              <td>
                <span class="sites-badge">{{ company.sites }} {{ company.sites === 1 ? 'site' : 'sites' }}</span>
              </td>
              <td><Tag :value="company.status" :severity="company.status === 'Active' ? 'success' : 'warn'" size="small" /></td>
              <td class="td-action">
                <div class="row-actions">
                  <Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View" @click="router.push(`/admin/companies/${company.id}`)" />
                  <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="router.push(`/admin/companies/${company.id}/edit`)" />
                  <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
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
.page-title-wrapper { display: flex; flex-direction: column; }
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
.summary-data { display: flex; flex-direction: column; }
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

/* ── FILTER PANEL ─────────────────────────────────────────────────── */
.filter-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.1rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.search-row { margin-bottom: 0.9rem; }
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
.search-field i { color: var(--text-faint); font-size: 0.9rem; }
.search-field input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 0.87rem;
  color: var(--text-primary);
}
.search-field input::placeholder { color: var(--text-faint); }
.filter-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr) auto;
  gap: 0.75rem;
  align-items: end;
}
.filter-field { display: flex; flex-direction: column; gap: 0.3rem; }
.filter-field label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.filter-action { justify-content: flex-end; padding-bottom: 0.15rem; }

/* ── TABLE PANEL ──────────────────────────────────────────────────── */
.table-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  overflow: hidden;
}
.table-scroll { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
.data-table thead th {
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
.data-table tbody tr:last-child td { border-bottom: none; }
.data-table tbody tr { transition: background var(--transition-fast); }
.data-table tbody tr:hover { background: #f8fafc; }
.th-action { width: 110px; text-align: center; }

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
.link-btn:hover { color: #1d4ed8; text-decoration: underline; }
.cell-text { font-size: 0.84rem; color: var(--text-secondary); }
.sites-badge {
  display: inline-block;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: #f0fdf4;
  color: #166534;
}

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.td-action { text-align: center; }
.row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
}
.act-view { color: #2563eb !important; }
.act-view:hover { background: #eff6ff !important; }
.act-edit { color: #059669 !important; }
.act-edit:hover { background: #ecfdf5 !important; }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }

/* ── STATE BOXES ──────────────────────────────────────────────────── */
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
.state-icon { font-size: 1.75rem; color: var(--brand-blue); margin-bottom: 0.25rem; }
.state-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--surface-subtle);
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.state-icon-wrap i { font-size: 1.4rem; color: var(--text-faint); }
.state-box strong { color: var(--text-primary); font-size: 0.95rem; }

/* ── RESPONSIVE ───────────────────────────────────────────────────── */
@media (max-width: 1200px) { .summary-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; gap: 1rem; }
  .page-heading { flex-direction: column; align-items: stretch; }
  .page-heading-actions { width: 100%; justify-content: flex-end; }
  .summary-grid { grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
  .summary-card { padding: 0.9rem 1rem; }
  .summary-icon { width: 38px; height: 38px; font-size: 0.95rem; }
  .summary-value { font-size: 1.2rem; }
  .filter-grid { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 480px) {
  .summary-grid { grid-template-columns: 1fr; }
  .filter-grid { grid-template-columns: 1fr; }
}
</style>
