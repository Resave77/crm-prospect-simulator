<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

const store = useCustomerListStore()
const router = useRouter()
const error = ref('')
const activeTab = ref('site')

const tabs = [
  { key: 'site', label: 'Customer Site' },
  { key: 'company', label: 'Company' },
  { key: 'master', label: 'Master Data' }
]

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

function applyFilters() {
  store.setParam('page', 1)
  load()
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
    <div class="page-heading">
      <div>
        <p class="eyebrow">Customer Management</p>
        <h1>Customer Existing</h1>
        <p class="muted">Customer Site &amp; Parent Company Management</p>
      </div>
      <Tag :value="`${store.total} customers`" severity="success" />
    </div>

    <!-- TAB NAVIGATION -->
    <div class="ct-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        :class="['ct-tab', { active: activeTab === tab.key }]"
        @click="activeTab = tab.key"
      >{{ tab.label }}</button>
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <!-- CUSTOMER SITE TAB -->
    <template v-if="activeTab === 'site'">
      <!-- SEARCH BAR -->
      <div class="ct-search-bar">
        <div class="ct-search-field">
          <i class="pi pi-search" />
          <input
            type="text"
            placeholder="Search Customer..."
            :value="store.params.keyword"
            @input="onKeywordSearch(($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- FILTER ROW -->
      <div class="ct-filter-row">
        <div class="ct-filters">
          <label>
            <span>Region</span>
            <Select v-model="selectedRegion" :options="regionOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
          </label>
          <label>
            <span>Segment</span>
            <Select v-model="selectedSegment" :options="segmentOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
          </label>
          <label>
            <span>Category</span>
            <Select v-model="selectedCategory" :options="categoryOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
          </label>
          <label>
            <span>Sales Executive</span>
            <Select v-model="selectedSales" :options="salesOptions" optionLabel="label" optionValue="value" class="ct-filter-select" />
          </label>
          <label>
            <span>Sort By</span>
            <Select v-model="selectedSort" :options="sortOptions" optionLabel="label" class="ct-filter-select" />
          </label>
        </div>
        <div class="ct-filter-actions">
          <Button label="Reset" icon="pi pi-replay" text size="small" @click="resetAll" />
        </div>
      </div>

      <!-- ACTION ROW -->
      <div class="ct-action-row">
        <div class="ct-action-right">
          <Button icon="pi pi-trash" text size="small" disabled title="Trash" />
          <Button icon="pi pi-download" text size="small" disabled title="Export" />
          <Button icon="pi pi-refresh" text size="small" @click="load" title="Refresh" />
        </div>
      </div>

      <!-- TABLE -->
      <div class="table-card">
        <div v-if="store.loading" class="empty-state">Loading Customer Sites...</div>
        <div v-else-if="!store.items.length" class="empty-state">
          <i class="pi pi-users" />
          <strong>No customer sites found</strong>
          <span>Adjust your search or filters, or convert a Won Prospect first.</span>
        </div>
        <div v-else class="responsive-table">
          <table>
            <thead>
              <tr>
                <th class="col-check"><input type="checkbox" :checked="store.isAllSelected()" @change="store.toggleSelectAll()" /></th>
                <th>Customer Code</th>
                <th>Customer Site</th>
                <th>Parent Company</th>
                <th>Region</th>
                <th>Segment</th>
                <th>Category</th>
                <th>Sales Executive</th>
                <th>Pipeline Status</th>
                <th>Converted</th>
                <th>Updated</th>
                <th class="col-action">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in store.items" :key="c.id">
                <td class="col-check"><input type="checkbox" :checked="store.selectedIds.has(c.id)" @change="store.toggleSelect(c.id)" /></td>
                <td><span class="mono">{{ c.customerCode }}</span></td>
                <td>
                  <button class="ct-link" @click="router.push(`/admin/customers/${c.id}`)">
                    <strong>{{ c.name }}</strong>
                  </button>
                </td>
                <td>
                  <strong>{{ c.parentCompanyName }}</strong>
                  <span>{{ c.parentCode }}</span>
                </td>
                <td><span>{{ c.region || '—' }}</span></td>
                <td><Tag :value="c.segment" :severity="segmentSeverity(c.segment)" /></td>
                <td><span>{{ c.category }}</span></td>
                <td><span>{{ c.salesExecutiveName }}</span></td>
                <td><Tag value="Active" severity="success" /></td>
                <td><span>{{ formatDate(c.convertedAt) }}</span></td>
                <td><span>{{ formatDate(c.updatedAt) }}</span></td>
                <td class="col-action">
                  <div class="ct-row-actions">
                    <Button icon="pi pi-eye" text rounded size="small" title="View Detail" @click="router.push(`/admin/customers/${c.id}`)" />
                    <Button icon="pi pi-pencil" text rounded size="small" title="Edit" @click="router.push(`/admin/customers/${c.id}/edit`)" />
                    <Button icon="pi pi-trash" text rounded size="small" severity="danger" title="Delete" disabled />
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- PAGINATION -->
      <div v-if="store.pages > 1" class="ct-pagination">
        <span class="ct-page-info">
          Showing {{ (store.page - 1) * store.limit + 1 }}&ndash;{{ Math.min(store.page * store.limit, store.total) }} of {{ store.total }}
        </span>
        <div class="ct-page-controls">
          <Button icon="pi pi-angle-left" text rounded size="small" :disabled="store.page <= 1" @click="goToPage(store.page - 1)" />
          <template v-for="p in store.pages" :key="p">
            <Button
              :label="String(p)"
              text rounded size="small"
              :class="{ 'p-highlight': p === store.page }"
              @click="goToPage(p)"
            />
          </template>
          <Button icon="pi pi-angle-right" text rounded size="small" :disabled="store.page >= store.pages" @click="goToPage(store.page + 1)" />
        </div>
      </div>
    </template>

    <!-- COMPANY TAB (placeholder) -->
    <div v-if="activeTab === 'company'" class="ct-placeholder">
      <i class="pi pi-building" />
      <strong>Company Management</strong>
      <span>Parent company list and management will be available here.</span>
    </div>

    <!-- MASTER DATA TAB (placeholder) -->
    <div v-if="activeTab === 'master'" class="ct-placeholder">
      <i class="pi pi-database" />
      <strong>Master Data</strong>
      <span>Segment, Category, and other master data configuration will be available here.</span>
    </div>
  </section>
</template>

<style scoped>
.ct-tabs { display: flex; gap: 0; margin-bottom: 1.25rem; border-bottom: 2px solid var(--border-light); }
.ct-tab {
  padding: 0.65rem 1.25rem; font-size: 0.82rem; font-weight: 600; color: var(--text-muted);
  background: none; border: none; border-bottom: 2px solid transparent; margin-bottom: -2px;
  cursor: pointer; transition: color var(--transition-fast), border-color var(--transition-fast);
}
.ct-tab:hover { color: var(--text-primary); }
.ct-tab.active { color: var(--brand-blue); border-bottom-color: var(--brand-blue); }
.ct-search-bar { margin-bottom: 1rem; }
.ct-search-field {
  display: flex; align-items: center; gap: 0.6rem; padding: 0.55rem 0.85rem;
  background: var(--surface-card); border: 1px solid var(--border-default); border-radius: var(--radius-md);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.ct-search-field:focus-within { border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.ct-search-field i { color: var(--text-faint); font-size: 0.85rem; }
.ct-search-field input { flex: 1; border: none; outline: none; background: transparent; font-size: 0.85rem; color: var(--text-primary); }
.ct-search-field input::placeholder { color: var(--text-faint); }
.ct-filter-row {
  display: flex; align-items: flex-end; gap: 1rem; margin-bottom: 1rem; padding: 0.85rem;
  background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); box-shadow: var(--shadow-xs);
}
.ct-filters { flex: 1; display: flex; flex-wrap: wrap; gap: 0.75rem; }
.ct-filters label { display: grid; gap: 0.3rem; min-width: 140px; }
.ct-filters label > span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.ct-filter-select { width: 100%; }
.ct-filter-actions { flex-shrink: 0; }
.ct-action-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.85rem; }
.ct-action-left, .ct-action-right { display: flex; align-items: center; gap: 0.35rem; }
.col-check { width: 40px; text-align: center; }
.col-action { width: 120px; text-align: center; }
.ct-row-actions { display: flex; align-items: center; justify-content: center; gap: 0.15rem; }
.ct-link { background: none; border: none; padding: 0; cursor: pointer; text-align: left; font: inherit; }
.ct-link strong { color: var(--brand-blue); transition: color var(--transition-fast); }
.ct-link:hover strong { color: var(--brand-blue-light); text-decoration: underline; }
.mono { font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Roboto Mono', monospace; font-size: 0.8rem; color: var(--text-secondary); }
.ct-pagination { display: flex; align-items: center; justify-content: space-between; padding: 0.85rem 0; }
.ct-page-info { font-size: 0.78rem; color: var(--text-muted); }
.ct-page-controls { display: flex; align-items: center; gap: 0.2rem; }
.ct-page-controls :deep(.p-highlight) { background: var(--brand-blue); color: #fff; }
.ct-placeholder {
  min-height: 300px; display: grid; place-content: center; justify-items: center; gap: 0.5rem;
  text-align: center; color: var(--text-muted); background: var(--surface-card);
  border: 1px solid var(--border-light); border-radius: var(--radius-xl);
}
.ct-placeholder i { font-size: 2rem; color: var(--brand-blue); }
.ct-placeholder strong { color: var(--text-primary); }
@media (max-width: 1100px) {
  .ct-filter-row { flex-direction: column; align-items: stretch; }
  .ct-filters { flex-direction: column; }
  .ct-filters label { min-width: 100%; }
  .ct-action-row { flex-wrap: wrap; gap: 0.5rem; }
}
@media (max-width: 760px) {
  .ct-tabs { overflow-x: auto; }
  .ct-tab { white-space: nowrap; padding: 0.5rem 0.85rem; font-size: 0.78rem; }
  .ct-pagination { flex-direction: column; gap: 0.5rem; }
}
</style>
