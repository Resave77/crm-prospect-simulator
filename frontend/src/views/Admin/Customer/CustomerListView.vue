<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { useCustomerListStore } from '../../../stores/customerList'
import { deleteCustomer, listTrashedCustomers, restoreCustomer } from '../../../api/crm'
import { listCategories, listSegments, type MasterDataCategory, type MasterDataSegment } from '../../../api/masterData'
import { fallbackCategories, fallbackSegments } from '../../../utils/masterDataFallback'
import MasterDataPanel from '../../../components/admin/MasterDataPanel.vue'
import type { CustomerSite } from '../../../types/crm'
import { DEV_CAPABILITIES } from '../../../config/capabilities'

const store = useCustomerListStore()
const router = useRouter()
const route = useRoute()
const error = ref('')
const pageSizeDraft = ref(10)
const goToPageDraft = ref(1)
const activeTab = ref('site')
const showFilters = ref(false)
const customerTrashEnabled = DEV_CAPABILITIES.customerTrash

watch(() => route.query.tab, (tab) => {
  if (tab === 'site' || tab === 'company' || tab === 'master') activeTab.value = tab
}, { immediate: true })
const deleteDialogVisible = ref(false)
const deleteTargetId = ref('')
const deleteTargetName = ref('')
const deleting = ref(false)
const trashVisible = ref(false)
const trashedCustomers = ref<CustomerSite[]>([])
const trashLoading = ref(false)
const restoringCustomerId = ref('')
async function openTrash() {
  if (!customerTrashEnabled) return
  trashVisible.value = true; trashLoading.value = true
  try { trashedCustomers.value = await listTrashedCustomers() } finally { trashLoading.value = false }
}
async function restoreCustomerItem(customer: CustomerSite) {
  if (!customerTrashEnabled) return
  restoringCustomerId.value = customer.id
  try { await restoreCustomer(customer.id); trashedCustomers.value = trashedCustomers.value.filter(c => c.id !== customer.id); await store.fetchCustomers() } finally { restoringCustomerId.value = '' }
}
const bulkMoving = ref(false)
const selectedCustomerId = ref('')
const customerActionVisible = ref(false)
const selectedCustomer = computed(() => store.items.find((c) => c.id === selectedCustomerId.value))
const companyDeleteDialogVisible = ref(false)
const bulkCompanyDeleteDialogVisible = ref(false)
const companyActionVisible = ref(false)
const selectedCompanyIds = ref<Set<string>>(new Set())
const selectedCompany = ref<typeof companies.value[number] | null>(null)
const companyDeleteTarget = ref<{ id: string; name: string; sites: number } | null>(null)
const masterCategories = ref<MasterDataCategory[]>(fallbackCategories)
const masterSegments = ref<MasterDataSegment[]>(fallbackSegments)

const tabs = [
  { key: 'site', label: 'Customer Site', icon: 'pi pi-map-marker' },
  { key: 'company', label: 'Company', icon: 'pi pi-building' },
  { key: 'master', label: 'Master Data', icon: 'pi pi-database' }
]

const segmentOptions = computed(() => [{ label: 'All Segments', value: '' }, ...masterSegments.value.map((s) => ({ label: s.name, value: s.name }))])
const categoryOptions = computed(() => [{ label: 'All Categories', value: '' }, ...masterCategories.value.map((c) => ({ label: c.name, value: c.name }))])

function selectTab(tabKey: string) {
  activeTab.value = tabKey
  router.replace({ query: { ...route.query, tab: tabKey } })
}

const sortOptions = [
  { label: 'Newest First', value: '' },
  { label: 'Oldest First', value: 'oldest' },
  { label: 'Customer Name', value: 'name' },
  { label: 'Customer Code', value: 'code' },
  { label: 'Converted Date', value: 'converted' },
  { label: 'Updated Date', value: 'updated' }
]

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
  { label: 'Key Account', value: 'Key Account' },
  { label: 'Regular Account', value: 'Regular Account' },
  { label: 'Strategic Account', value: 'Strategic Account' },
  { label: 'Small Account', value: 'Small Account' }
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
        tier: 'Key Account',
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
const allCompaniesSelected = computed(() => companies.value.length > 0 && companies.value.every((company) => selectedCompanyIds.value.has(company.id)))
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
  if (!dateStr) return 'â€”'
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

function toggleCompanySelection(id: string) {
  const next = new Set(selectedCompanyIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedCompanyIds.value = next
}

function toggleAllCompanySelection() {
  selectedCompanyIds.value = allCompaniesSelected.value ? new Set() : new Set(companies.value.map((company) => company.id))
}

async function executeBulkDeleteCompany() {
  if (!customerTrashEnabled) return
  const ids = [...selectedCompanyIds.value]
  if (!ids.length) return
  deleting.value = true
  try {
    const sites = store.allCustomers.filter((customer) => ids.includes(customer.parentCode || ''))
    await Promise.all(sites.map((site) => deleteCustomer(site.id)))
    const removed = new Set(ids)
    store.allCustomers = store.allCustomers.filter((customer) => !removed.has(customer.parentCode || ''))
    store.items = store.items.filter((customer) => !removed.has(customer.parentCode || ''))
    store.total = store.allCustomers.length
    selectedCompanyIds.value = new Set()
    bulkCompanyDeleteDialogVisible.value = false
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    deleting.value = false
  }
}

function tierBadgeClass(tier: string) {
  switch (tier) {
    case 'Regular Account': return 'tier-regular'
    case 'Strategic Account': return 'tier-strategic'
    case 'Small Account': return 'tier-small'
    default: return 'tier-key'
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
    try {
      const [categories, segments] = await Promise.all([listCategories(), listSegments()])
      if (categories?.length) masterCategories.value = categories
      if (segments?.length) masterSegments.value = segments
    } catch (masterError) {
      console.warn('Master Data API unavailable; using frontend defaults.', masterError)
    }
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
  if (!customerTrashEnabled) return
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

function segmentLabel(segment: string) {
  const value = String(segment || '').toUpperCase()
  return value.includes('B2C') || value.includes('RETAIL') ? 'B2C' : 'B2B'
}

function applyPaginationSettings() {
  store.limit = pageSizeDraft.value
  goToPage(Math.max(1, Math.min(goToPageDraft.value, store.pages)))
}

async function moveSelectedToTrash() {
  const ids = [...store.selectedIds]
  if (!ids.length) return
  bulkMoving.value = true
  try {
    await Promise.all(ids.map((id) => deleteCustomer(id)))
    const removed = new Set(ids)
    store.allCustomers = store.allCustomers.filter((c) => !removed.has(c.id))
    store.items = store.items.filter((c) => !removed.has(c.id))
    store.total = store.allCustomers.length
    store.selectedIds = new Set()
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    bulkMoving.value = false
  }
}

function confirmDeleteCompany(company: { id: string; name: string; sites: number }) {
  companyDeleteTarget.value = company
  companyDeleteDialogVisible.value = true
}

async function executeDeleteCompany() {
  if (!customerTrashEnabled) return
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
  <section :class="['admin-page', { 'master-mode': activeTab === 'master' }]">
    <header v-if="activeTab !== 'site'" class="workspace-header">
      <div class="workspace-heading">
        <div class="page-title-wrapper">
          <span class="eyebrow">Customer Management</span>
          <h1>Customer & Perusahaan</h1>
          <p class="muted">Kelola site customer, akun perusahaan, kepemilikan, dan data referensi.</p>
        </div>
      </div>
      <div class="workspace-summary">
        <button type="button" class="summary-item" @click="selectTab('company')"><i class="pi pi-building si-blue" /><span>Perusahaan</span><strong>{{ companyCount }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('site')"><i class="pi pi-map-marker si-violet" /><span>Sites</span><strong>{{ store.allCustomers.length }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('site')"><i class="pi pi-user-check si-emerald" /><span>Assigned</span><strong>{{ assignedCount }}</strong></button>
        <button type="button" class="summary-item" @click="selectTab('company')"><i class="pi pi-clock si-amber" /><span>ERP Pending</span><strong>{{ erpPendingCount }}</strong></button>
      </div>
      <div class="page-heading-actions legacy-page-actions">
        <Button v-if="customerTrashEnabled" label="Trash" icon="pi pi-trash" severity="secondary" outlined size="small" @click="openTrash" />
        <Button label="Export" icon="pi pi-download" severity="secondary" outlined size="small" />
        <Button v-if="activeTab !== 'master'" label="Add Customer" icon="pi pi-plus" size="small" @click="router.push('/admin/customers/add')" />
      </div>
    </header>
    <Message v-if="error" severity="error" class="page-message">{{ error }}</Message>
    <nav class="flex min-w-0 flex-1 items-center gap-[10px] overflow-x-auto border-b border-[#e2e8f0] px-[24px] py-[10px] pb-[11px]" aria-label="Customer management sections">
      <div class="inline-flex shrink-0 max-w-full flex-nowrap gap-[4px] rounded-[16px] bg-[#f1f5f9] p-[3px]">
        <button v-for="tab in tabs" :key="tab.key" type="button" class="flex h-[36px] items-center shrink-0 rounded-[12px] px-[14px] font-['Inter'] text-[13px] font-semibold transition-all" :class="activeTab === tab.key ? 'bg-[#dc2626] text-white shadow-[0px_4px_16px_0px_rgba(220,38,38,0.18)]' : 'text-[#64748b] hover:text-[#0f172a]'" @click="selectTab(tab.key)">{{ tab.label }}</button>
      </div>
      <label v-if="activeTab === 'site'" class="relative w-[260px] shrink-0 lg:w-[300px]"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="pointer-events-none absolute left-[12px] top-1/2 size-[16px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg><input v-model="store.params.keyword" class="h-[36px] w-full rounded-[10px] border border-[#e2e8f0] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] placeholder:text-[#94a3b8] focus:border-[#dc2626] focus:outline-none focus:ring-1 focus:ring-[#dc2626]" placeholder="Search customer site code, name, company, city..." @keyup.enter="store.fetchCustomers()" /></label>
      <label v-if="activeTab === 'company'" class="relative w-[260px] shrink-0 lg:w-[300px]"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="pointer-events-none absolute left-[12px] top-1/2 size-[16px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg><input type="text" v-model="companyKeyword" placeholder="Search company code, company name, KAM..." class="h-[36px] w-full rounded-[10px] border border-[#e2e8f0] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] placeholder:text-[#94a3b8] focus:border-[#dc2626] focus:outline-none focus:ring-1 focus:ring-[#dc2626]" /></label>
      <button type="button" class="flex h-[36px] shrink-0 items-center gap-[8px] rounded-[10px] border border-[#e2e8f0] bg-white px-[12px] font-['Inter'] text-[12px] font-semibold text-[#475569] transition-all hover:border-[#cbd5e1] hover:bg-[#f8fafc]" @click="showFilters = !showFilters"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[15px]"><line x1="21" x2="14" y1="4" y2="4"></line><line x1="10" x2="3" y1="4" y2="4"></line><line x1="21" x2="12" y1="12" y2="12"></line><line x1="8" x2="3" y1="12" y2="12"></line><line x1="21" x2="16" y1="20" y2="20"></line><line x1="12" x2="3" y1="20" y2="20"></line><line x1="14" x2="14" y1="2" y2="6"></line><line x1="8" x2="8" y1="10" y2="14"></line><line x1="16" x2="16" y1="18" y2="22"></line></svg><span>More Filters</span></button>
      <div class="erp-customer-actions flex shrink-0 items-center gap-[8px]"><button type="button" class="flex h-[36px] shrink-0 items-center gap-[7px] rounded-[10px] border border-[#fecaca] bg-[#fff5f5] px-[12px] font-['Inter'] text-[12px] font-semibold text-[#b91c1b] transition-all hover:bg-[#fff1f2]" :disabled="activeTab === 'company' && !selectedCompanyIds.size" @click="activeTab === 'company' ? (bulkCompanyDeleteDialogVisible = true) : openTrash()"><i class="pi pi-trash text-[15px]" /><span>Trash</span></button><button v-if="activeTab !== 'master'" type="button" class="flex h-[36px] shrink-0 items-center gap-[8px] rounded-[10px] bg-[#dc2626] px-[12px] font-['Inter'] text-[12px] font-semibold text-white transition-all hover:bg-[#b91c1c]" @click="router.push('/admin/customers/add')"><i class="pi pi-plus text-[15px]" /><span>Create Customer Site</span></button></div>
    </nav>

    <!-- ======================== CUSTOMER SITE TAB ======================== -->
    <template v-if="activeTab === 'site'">
      <div v-if="store.selectedIds.size" class="selection-toolbar"><strong>{{ store.selectedIds.size }} selected</strong><Button label="Cancel" text size="small" :disabled="bulkMoving" @click="store.selectedIds = new Set()" /><Button label="Move Selected to Trash" icon="pi pi-trash" severity="danger" size="small" :loading="bulkMoving" @click="moveSelectedToTrash" /></div>
      <div class="panel-stack">
        <!-- FILTERS -->
        <div v-if="showFilters" class="filter-panel">
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
            <table class="data-table customer-site-table">
              <thead>
                <tr>
                  <th class="th-check">
                    <button type="button" class="customer-checkbox mx-auto flex size-[18px] items-center justify-center rounded-[4px] border border-[#cbd5e1] bg-white shadow-sm transition-colors" :class="{ 'customer-checkbox-selected': store.isAllSelected() }" aria-label="Select all customer sites" @click.stop="store.toggleSelectAll()"><i v-if="store.isAllSelected()" class="pi pi-check text-[11px]" /></button>
                  </th>
                  <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Site Code + Company Tier</span></th>
                  <th>Customer Site + Company</th>
                  <th>Region</th>
                  <th>Site Location</th>
                  <th>Segment</th>
                  <th>Sales Assignment</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="c in store.items" :key="c.id" :class="['clickable-row', { 'row-selected': selectedCustomerId === c.id }]" @click="selectedCustomerId = c.id; customerActionVisible = true">
                  <td class="td-check">
                    <button type="button" class="customer-checkbox mx-auto flex size-[18px] items-center justify-center rounded-[4px] border border-[#cbd5e1] bg-white shadow-sm transition-colors" :class="{ 'customer-checkbox-selected': store.selectedIds.has(c.id) }" :aria-label="`Select customer site ${c.name}`" @click.stop="store.toggleSelect(c.id)"><i v-if="store.selectedIds.has(c.id)" class="pi pi-check text-[11px]" /></button>
                  </td>
                  <td>
                    <div class="cell-stack"><code class="code-tag code-blue">{{ c.customerCode }}</code><Tag :value="segmentLabel(c.segment)" :severity="segmentSeverity(c.segment)" /></div>
                  </td>
                  <td>
                    <div class="cell-stack">
                      <button class="link-btn" @click.stop="selectedCustomerId = c.id">{{ c.name }}</button>
                      <span class="cell-sub">{{ c.parentCode }}</span>
                    </div>
                  </td>
                  <td><span class="cell-text">{{ c.region || '—' }}</span></td>
                  <td><div class="cell-stack"><span class="cell-primary">{{ c.address?.district || c.address?.subDistrict || '—' }}</span><span class="cell-sub">{{ c.address?.province || '' }}</span></div></td>
                  <td><Tag :value="segmentLabel(c.segment)" :severity="segmentSeverity(c.segment)" /></td>
                  <td>
                    <div class="min-w-0"><p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ c.salesExecutiveName || 'Unassigned' }}</p><p class="mt-[2px] truncate font-['Inter'] text-[12px] text-[#64748b]">KAM: {{ c.assignedSales?.fullName || c.salesExecutiveName || 'Unassigned' }}</p></div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- PAGINATION -->
          <div class="pagination-bar">
            <div class="flex items-center gap-[12px]">
              <div class="pagination-controls flex items-center gap-[4px]">
              <Button icon="pi pi-angle-left" text rounded size="small" :disabled="store.page <= 1" @click="goToPage(store.page - 1)" />
              <template v-for="(p, idx) in getPageNumbers()" :key="idx">
                <span v-if="p === '...'" class="pagination-dots">â€¦</span>
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
              <span class="pagination-info ml-[8px]">
                Page {{ store.page }} of {{ store.pages }} / {{ store.total }} records
              </span>
            </div>
            <div class="pagination-settings">
              <label class="page-size-control"><span>Page size</span><select v-model.number="pageSizeDraft" aria-label="Page size"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label>
              <label><span>Go to</span><input v-model.number="goToPageDraft" type="number" min="1" :max="store.pages" aria-label="Go to page" /></label>
              <button type="button" @click="applyPaginationSettings">Set</button>
              <button type="button" class="pagination-refresh" title="Refresh site data" @click="load"><i class="pi pi-refresh" /></button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- DELETE CONFIRMATION DIALOG -->
    <Dialog v-if="customerTrashEnabled" v-model:visible="deleteDialogVisible" header="Delete Customer" modal :style="{ width: '400px' }">
      <div class="delete-warning"><i class="pi pi-exclamation-triangle" /><div><strong>Move customer site to Trash?</strong><p>{{ deleteTargetName }} will be moved to Trash and can be restored later.</p></div></div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>

    <Dialog v-model:visible="companyActionVisible" modal :header="selectedCompany?.name || 'Company Actions'" :style="{ width: '460px' }" :breakpoints="{ '640px': '92vw' }">
      <div v-if="selectedCompany" class="customer-action-modal"><p>Choose what you want to do with this company record.</p><div><button class="customer-action-card" @click="companyActionVisible = false; goToCompany(selectedCompany!.id)"><i class="pi pi-eye" /><span><strong>View Detail Company</strong><small>Open the full company profile and review details.</small></span></button><button class="customer-action-card" @click="companyActionVisible = false; router.push(`/admin/companies/${selectedCompany!.id}/edit`)"><i class="pi pi-pencil" /><span><strong>Edit Company</strong><small>Open the company form and update the current information.</small></span></button><button class="customer-action-card danger" @click="companyActionVisible = false; confirmDeleteCompany(selectedCompany!)"><i class="pi pi-exclamation-triangle" /><span><strong>Move Company to Trash</strong><small>Move this company to Trash. All of its child customer sites will be moved too.</small></span></button></div></div>
    </Dialog>

    <!-- DELETE COMPANY CONFIRMATION DIALOG -->
  <Dialog v-model:visible="bulkCompanyDeleteDialogVisible" header="Delete Selected Companies" modal :style="{ width: '420px' }" :closable="!deleting">
    <div class="delete-warning"><i class="pi pi-exclamation-triangle" /><div><strong>Move selected companies to Trash?</strong><p>{{ selectedCompanyIds.size }} compan{{ selectedCompanyIds.size === 1 ? 'y' : 'ies' }} and all associated customer sites will be moved to Trash.</p></div></div>
    <template #footer><Button label="Cancel" severity="secondary" text :disabled="deleting" @click="bulkCompanyDeleteDialogVisible = false" /><Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeBulkDeleteCompany" /></template>
  </Dialog>

  <Dialog v-model:visible="companyDeleteDialogVisible" header="Delete Company" modal :style="{ width: '420px' }">
      <div class="delete-warning"><i class="pi pi-exclamation-triangle" /><div><strong>Move company to Trash?</strong><p>{{ companyDeleteTarget?.name }} and {{ companyDeleteTarget?.sites }} associated site(s) will be moved to Trash and can be restored later.</p></div></div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="companyDeleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDeleteCompany" />
      </template>
    </Dialog>

    <!-- ======================== COMPANY TAB ======================== -->
    <div v-if="activeTab === 'company'" class="panel-stack">
      <!-- FILTERS -->
      <div v-if="showFilters" class="filter-panel">
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

      <!-- PAGINATION -->
      <div v-if="activeTab === 'company'" class="pagination-bar company-pagination">
        <div class="pagination-controls">
          <Button icon="pi pi-angle-left" text rounded size="small" disabled />
          <Button label="1" text rounded size="small" class="pagination-num is-active" />
          <Button icon="pi pi-angle-right" text rounded size="small" disabled />
        </div>
        <span class="pagination-info">Page 1 of 1 / {{ companies.length }} records</span>
        <div class="pagination-settings">
          <label><span>Page size</span><select aria-label="Page size"><option>10</option><option>20</option><option>50</option></select></label>
          <label><span>Go to</span><input type="number" min="1" max="1" aria-label="Go to page" /></label>
          <button type="button">Set</button>
          <button type="button" class="pagination-refresh" title="Refresh site data" @click="load"><i class="pi pi-refresh" /></button>
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
          <table class="data-table company-table">
            <thead>
              <tr>
                <th class="th-check"><button type="button" class="customer-checkbox mx-auto" :class="{ 'customer-checkbox-selected': allCompaniesSelected }" aria-label="Select all companies" @click.stop="toggleAllCompanySelection"><i v-if="allCompaniesSelected" class="pi pi-check" /></button></th><th>Company Code</th><th>Company Name</th><th>Total Site</th><th>Company Tier</th><th>3M Avg Invoice</th><th>KAM</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="company in companies" :key="company.id" class="clickable-row" @click="selectedCompany = company; companyActionVisible = true">
                <td class="td-check"><button type="button" class="customer-checkbox mx-auto" :class="{ 'customer-checkbox-selected': selectedCompanyIds.has(company.id) }" :aria-label="`Select company ${company.name}`" @click.stop="toggleCompanySelection(company.id)"><i v-if="selectedCompanyIds.has(company.id)" class="pi pi-check" /></button></td><td class="border-r border-[#e0e0e0] p-[12px]">
                  <div class="min-w-0"><p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ company.companyCode }}</p></div>
                </td>
                <td>
                  <div class="cell-stack company-name-stack">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ company.name }}</p>
                    <p class="mt-[2px] truncate font-['Inter'] text-[12px] text-[#64748b]">{{ company.registeredLocation }}</p>
                  </div>
                </td>
                <td>
                  <div class="cell-stack">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ company.sites }} Customer Site{{ company.sites === 1 ? '' : 's' }}</p>
                    <p class="mt-[2px] truncate font-['Inter'] text-[12px] text-[#64748b]">{{ company.registeredLocation }}</p>
                  </div>
                </td>
                <td class="border-r border-[#e0e0e0] p-[12px]">
                  <Tag :value="company.tier" severity="secondary" :class="['company-tier-tag', tierBadgeClass(company.tier)]" />
                </td>
                <td class="border-r border-[#e0e0e0] p-[12px]">
                  <div class="min-w-0"><p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">Rp&nbsp;0</p></div>
                </td>
                <td class="border-r border-[#e0e0e0] p-[12px]">
                  <div class="min-w-0"><p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ company.kam }}</p></div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      </div>

    <!-- ======================== MASTER DATA TAB ======================== -->
    <MasterDataPanel v-if="activeTab === 'master'" />
  </section>
  <Dialog v-if="customerTrashEnabled" v-model:visible="trashVisible" modal header="Trash Customer Existing" :style="{ width: '680px' }" :breakpoints="{ '640px': '92vw' }">
    <div class="trash-modal-body">
      <div class="trash-empty-icon"><i class="pi pi-trash" /></div>
      <div v-if="trashLoading">Loading trash...</div>
      <template v-else-if="!trashedCustomers.length"><strong>Trash is empty</strong><span>Deleted customer sites will appear here.</span></template>
      <div v-else class="customer-trash-list"><div v-for="customer in trashedCustomers" :key="customer.id" class="customer-trash-row"><div><strong>{{ customer.name }}</strong><span>{{ customer.customerCode }} Â· {{ customer.parentCompanyName }}</span></div><Button label="Restore" icon="pi pi-undo" size="small" :loading="restoringCustomerId === customer.id" @click="restoreCustomerItem(customer)" /></div></div>
    </div>
  </Dialog>
  <Dialog v-model:visible="customerActionVisible" modal :header="selectedCustomer?.name || 'Customer Actions'" :style="{ width: '460px' }" :breakpoints="{ '640px': '92vw' }">
    <div v-if="selectedCustomer" class="customer-action-modal"><p>Choose what you want to do with this customer site / branch record.</p><div><button class="customer-action-card" @click="router.push(`/admin/customers/${selectedCustomer.id}`)"><i class="pi pi-eye" /><span><strong>View Detail Customer Site</strong><small>Open the full customer site profile and review details.</small></span></button><button class="customer-action-card" @click="router.push(`/admin/customers/${selectedCustomer.id}/edit`)"><i class="pi pi-pencil" /><span><strong>Edit Customer Site</strong><small>Open the customer site form and update the current information.</small></span></button><button class="customer-action-card danger" @click="customerActionVisible = false; confirmDelete(selectedCustomer.id, selectedCustomer.name)"><i class="pi pi-exclamation-triangle" /><span><strong>Move to Trash</strong><small>Move this customer site to Trash. It can be restored later.</small></span></button></div></div>
  </Dialog>
</template>

<style scoped>
.erp-customer-actions > button:first-child,
.customer-action-card.danger { display: none !important; }
.admin-page{box-sizing:border-box;display:flex;min-width:0;min-height:calc(100dvh - 4rem);flex-direction:column;gap:.75rem;padding:.85rem 1rem 1rem;overflow-x:hidden;background:#f8fafc}.workspace-header{display:grid;grid-template-columns:minmax(260px,1fr) auto auto;align-items:center;gap:.9rem;padding:.72rem .85rem;border:1px solid #e5eaf0;border-radius:12px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.workspace-heading{display:flex;min-width:0;align-items:center;gap:.65rem}.back-button{flex:0 0 auto}.page-title-wrapper{min-width:0;display:grid;gap:.04rem}.page-title-wrapper .eyebrow{color:#64748b;font-size:.56rem;font-weight:800;letter-spacing:.08em;text-transform:uppercase}.page-title-wrapper h1{margin:0;color:#0f172a;font-size:1.08rem;line-height:1.2;letter-spacing:-.02em}.page-title-wrapper .muted{margin:0;overflow:hidden;color:#94a3b8;font-size:.67rem;line-height:1.4;text-overflow:ellipsis;white-space:nowrap}.workspace-summary{display:grid;grid-template-columns:repeat(4,minmax(74px,auto));overflow:hidden;border:1px solid #e5eaf0;border-radius:9px;background:#f8fafc}.summary-item{display:grid;grid-template-columns:18px auto;grid-template-rows:auto auto;column-gap:.3rem;min-width:78px;padding:.4rem .55rem;border:0;border-right:1px solid #e5eaf0;background:transparent;text-align:left;cursor:pointer}.summary-item:last-child{border-right:0}.summary-item:hover{background:#fff0f1}.summary-item i{grid-row:1/3;align-self:center;font-size:.74rem}.summary-item span{color:#94a3b8;font-size:.5rem;font-weight:800;text-transform:uppercase}.summary-item strong{color:#0f172a;font-size:.78rem}.si-blue{color:#e63946}.si-violet{color:#ef4e5d}.si-emerald{color:#059669}.si-amber{color:#d97706}.page-heading-actions{display:flex;align-items:center;justify-content:flex-end;gap:.45rem}.page-message{margin:0}.tabs-bar{display:flex;min-width:0;gap:.35rem;padding:.32rem;overflow-x:auto;border:1px solid #e5eaf0;border-radius:10px;background:#fff;scrollbar-width:none}.tabs-bar::-webkit-scrollbar{display:none}.tab-item{display:inline-flex;min-height:34px;align-items:center;gap:.42rem;padding:.42rem .72rem;border:1px solid transparent;border-radius:8px;background:transparent;color:#64748b;font-size:.72rem;font-weight:700;white-space:nowrap;cursor:pointer}.tab-item:hover{background:#f8fafc;color:#0f172a}.tab-item.active{border-color:#f4b3ba;background:#fff0f1;color:#d62839}.tab-item strong{min-width:20px;padding:.08rem .35rem;border-radius:999px;background:rgba(148,163,184,.14);font-size:.58rem;text-align:center}.tab-item.active strong{background:#ffd9dc}.panel-stack{display:flex;min-width:0;flex-direction:column;gap:.65rem}.filter-panel{display:grid;grid-template-columns:minmax(250px,1.35fr) minmax(0,3.65fr);align-items:end;gap:.7rem;padding:.65rem;border:1px solid #e5eaf0;border-radius:10px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.search-row{min-width:0;margin:0}.search-field{display:flex;min-height:38px;align-items:center;gap:.55rem;padding:.45rem .7rem;border:1px solid #dbe3ee;border-radius:8px;background:#f8fafc}.search-field:focus-within{border-color:#e63946;background:#fff;box-shadow:0 0 0 3px rgba(230,57,70,.08)}.search-field i{color:#94a3b8;font-size:.76rem}.search-field input{width:100%;min-width:0;border:0;outline:0;background:transparent;color:#0f172a;font-size:.76rem}.search-field input::placeholder{color:#94a3b8}.filter-grid{display:grid;grid-template-columns:repeat(5,minmax(110px,1fr)) auto;align-items:end;gap:.5rem;min-width:0}.filter-field{display:grid;min-width:0;gap:.22rem}.filter-field label{color:#64748b;font-size:.55rem;font-weight:800;letter-spacing:.05em;text-transform:uppercase}.filter-field :deep(.p-select){min-width:0;height:38px;border-radius:8px;font-size:.72rem}.filter-field :deep(.p-select-label){padding-block:.5rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.filter-action{display:flex;align-items:flex-end;justify-content:flex-end;padding-bottom:.02rem}.table-panel{min-width:0;overflow:hidden;border:1px solid #e5eaf0;border-radius:10px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.table-scroll{width:100%;min-width:0;overflow-x:auto;scrollbar-width:thin}.data-table{width:100%;min-width:980px;table-layout:fixed;border-collapse:collapse;font-size:.74rem}.data-table thead th{position:sticky;top:0;z-index:2;padding:.58rem .65rem;border-bottom:1px solid #e5eaf0;background:#f8fafc;color:#64748b;font-size:.58rem;font-weight:800;letter-spacing:.05em;text-align:left;text-transform:uppercase;white-space:nowrap}.data-table tbody td{height:48px;padding:.48rem .65rem;overflow:hidden;border-bottom:1px solid #edf1f6;color:#1e293b;text-overflow:ellipsis;vertical-align:middle}.data-table tbody tr:last-child td{border-bottom:0}.data-table tbody tr:hover{background:#fffbfb}.clickable-row{cursor:default}.th-check,.td-check{width:42px;text-align:center}.th-action{width:112px;text-align:center}.code-tag{display:inline-block;max-width:100%;overflow:hidden;padding:.14rem .42rem;border-radius:5px;background:#f1f5f9;color:#475569;font-family:'SF Mono','Fira Code',Consolas,monospace;font-size:.66rem;font-weight:650;text-overflow:ellipsis;white-space:nowrap}.code-blue{background:#fff0f1;color:#e63946}.link-btn{max-width:100%;overflow:hidden;border:0;background:transparent;color:#d62839;font:inherit;font-weight:750;text-align:left;text-overflow:ellipsis;white-space:nowrap;cursor:pointer}.link-btn:hover{text-decoration:underline;text-underline-offset:2px}.cell-stack{display:grid;min-width:0;gap:.04rem}.cell-primary,.cell-text,.cell-date,.cell-badge{display:block;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.cell-primary{color:#0f172a;font-size:.73rem;font-weight:700}.cell-sub{color:#94a3b8;font-family:Consolas,monospace;font-size:.6rem}.cell-text{color:#475569;font-size:.7rem}.cell-date{color:#64748b;font-size:.68rem}.cell-badge{color:#334155;font-size:.68rem;font-weight:700}.site-status-cell{display:flex;min-width:0;align-items:center;gap:.3rem}.td-action{text-align:center}.row-actions{display:inline-flex;align-items:center;justify-content:center;gap:.12rem;padding:.12rem;border:1px solid #e5eaf0;border-radius:999px;background:#fff}.row-actions :deep(.p-button){width:1.85rem;height:1.85rem}.act-view{color:#e63946!important}.act-view:hover{background:#fff0f1!important}.act-edit{color:#059669!important}.act-edit:hover{background:#ecfdf5!important}.act-delete{color:#dc2626!important}.act-delete:hover{background:#fef2f2!important}.state-box{min-height:240px;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:.4rem;padding:2rem;color:#64748b;text-align:center}.state-icon{color:#e63946;font-size:1.5rem}.state-icon-wrap{display:grid;width:50px;height:50px;place-content:center;border-radius:13px;background:#f1f5f9;color:#94a3b8}.state-box strong{color:#0f172a;font-size:.86rem}.pagination-bar{display:flex;align-items:center;justify-content:space-between;gap:.75rem;padding:.58rem .75rem;border-top:1px solid #e5eaf0;background:#f8fafc}.pagination-info{color:#64748b;font-size:.67rem}.pagination-info strong{color:#0f172a}.pagination-controls{display:flex;align-items:center;gap:.12rem}.pagination-num{min-width:28px;height:28px;color:#475569;font-size:.68rem;font-weight:700}.pagination-num.is-active{background:#e63946!important;color:#fff!important}.pagination-dots{padding:0 .2rem;color:#94a3b8}.placeholder-panel{min-height:320px;display:grid;place-content:center;border:1px solid #e5eaf0;border-radius:10px;background:#fff}.placeholder-inner{display:flex;max-width:360px;flex-direction:column;align-items:center;gap:.45rem;color:#64748b;text-align:center}.placeholder-icon{display:grid;width:58px;height:58px;place-content:center;margin-bottom:.25rem;border-radius:15px;background:#fff0f1;color:#e63946}.placeholder-inner strong{color:#0f172a;font-size:.9rem}.placeholder-inner span{font-size:.75rem;line-height:1.5}@media(max-width:1280px){.workspace-header{grid-template-columns:minmax(240px,1fr) auto}.workspace-summary{grid-column:1/-1;grid-row:2}.filter-panel{grid-template-columns:1fr}.filter-grid{grid-template-columns:repeat(3,minmax(130px,1fr)) auto}}@media(max-width:900px){.admin-page{min-height:auto;padding:.75rem;overflow:visible}.workspace-header{grid-template-columns:1fr auto}.workspace-summary{grid-template-columns:repeat(4,1fr);width:100%}.page-title-wrapper .muted{white-space:normal}.filter-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.filter-action{grid-column:1/-1}.data-table{min-width:920px}}@media(max-width:640px){.admin-page{padding:.6rem;gap:.6rem}.workspace-header{grid-template-columns:1fr;align-items:stretch}.workspace-heading{align-items:flex-start}.page-heading-actions{display:grid;grid-template-columns:1fr 1fr}.page-heading-actions :deep(.p-button){width:100%}.workspace-summary{grid-template-columns:repeat(2,1fr)}.summary-item:nth-child(2){border-right:0}.summary-item:nth-child(-n+2){border-bottom:1px solid #e5eaf0}.filter-panel{padding:.55rem}.filter-grid{grid-template-columns:1fr}.filter-action{grid-column:auto}.pagination-bar{flex-direction:column;align-items:stretch}.pagination-controls{justify-content:center}.data-table{min-width:860px}}
.selection-toolbar{display:flex;align-items:center;gap:.5rem;padding:.5rem .75rem;border:1px solid #fecaca;border-radius:9px;background:#fff5f5;color:#991b1b}.selection-toolbar strong{margin-right:auto;font-size:.72rem}.selection-toolbar :deep(.p-button){font-size:.68rem}
.data-table tbody tr.row-selected{background:#fff8f8;box-shadow:inset 3px 0 #e63946}.data-table tbody tr{cursor:pointer}
.trash-modal-body{display:flex;flex-direction:column;align-items:center;gap:.45rem;padding:1.4rem 1rem 1.7rem;text-align:center;color:#64748b}.trash-empty-icon{display:grid;place-items:center;width:58px;height:58px;margin-bottom:.25rem;border-radius:16px;background:#fff1f2;color:#dc2626;font-size:1.35rem}.trash-modal-body strong{color:#172033;font-size:.95rem}.trash-modal-body span{font-size:.72rem}
.table-panel>.state-box{min-height:120px;padding:1.25rem}.table-panel>.state-box .state-icon{font-size:1.1rem}.table-panel>.state-box span{font-size:.7rem}
.customer-action-modal{display:grid;gap:.35rem}.customer-action-modal strong{color:#172033;font-size:.95rem}.customer-action-modal>span{color:#94a3b8;font-family:monospace;font-size:.68rem}.customer-action-modal>div{display:grid;gap:.5rem;margin-top:.75rem}.customer-action-modal :deep(.p-button){justify-content:center;width:100%}
.customer-action-modal>p{margin:0 0 .7rem;color:#64748b;font-size:.72rem;line-height:1.5}.customer-action-card{display:flex;align-items:center;gap:.75rem;width:100%;padding:.85rem .8rem;border:1px solid #dfe6ef;border-radius:13px;background:#fff;color:#172033;text-align:left;cursor:pointer;transition:border-color .16s,background .16s,box-shadow .16s}.customer-action-card:hover{border-color:#b8c9df;background:#f8fbff;box-shadow:0 3px 10px rgba(15,23,42,.06)}.customer-action-card>i{display:grid;place-items:center;width:34px;height:34px;flex:none;border-radius:50%;background:#eff6ff;color:#2563eb;font-size:.85rem}.customer-action-card>span{display:grid;gap:.2rem;min-width:0}.customer-action-card strong{font-size:.74rem}.customer-action-card small{color:#64748b;font-size:.64rem;line-height:1.35}.customer-action-card.danger{border-color:#fecaca;background:#fff7f7}.customer-action-card.danger:hover{background:#fff1f2;border-color:#fca5a5}.customer-action-card.danger>i{background:#fff;color:#dc2626}.customer-action-card.danger strong,.customer-action-card.danger small{color:#b91c1c}
@media(min-width:901px){.panel-stack .filter-panel{display:flex;align-items:flex-end;gap:.5rem;padding:.5rem .6rem}.panel-stack .filter-panel .search-row{flex:0 1 280px}.panel-stack .filter-panel .filter-grid{display:grid;grid-template-columns:repeat(5,minmax(100px,1fr)) auto;flex:1;gap:.4rem}.panel-stack .filter-field :deep(.p-select){height:36px}.panel-stack .filter-field label{font-size:.52rem}}
.admin-page .data-table{width:100%;table-layout:fixed}.admin-page .data-table thead th{height:42px;padding:0 10px;background:#f4f4f4}.admin-page .data-table tbody td{height:67px;padding:8px 10px;border-bottom:1px solid #e2e8f0}.admin-page .table-panel{width:100%}
@media (max-width: 640px) {
  .admin-page { background:#f8fafc; padding:.5rem; gap:.5rem; }
  .workspace-header { padding:.85rem; border-radius:14px; gap:.7rem; }
  .page-title-wrapper h1 { font-size:1.12rem; }
  .page-title-wrapper .muted { font-size:.64rem; line-height:1.35; white-space:normal; }
  .workspace-summary { border-radius:10px; }
  .summary-item { min-width:0; padding:.5rem .4rem; grid-template-columns:16px auto; column-gap:.2rem; }
  .summary-item span { font-size:.48rem; }
  .page-heading-actions { grid-template-columns:repeat(2,1fr); gap:.4rem; }
  .page-heading-actions :deep(.p-button) { min-height:38px; font-size:.68rem; }
  .page-heading-actions :deep(.p-button:last-child) { grid-column:1/-1; background:#e63946; border-color:#e63946; }
  .tab-item { flex:1; justify-content:center; min-height:38px; padding:.4rem .35rem; font-size:.64rem; }
  .filter-panel { padding:.7rem; border-radius:12px; }
  .filter-field :deep(.p-select) { height:40px; border-radius:9px; }
  .filter-action :deep(.p-button) { width:100%; min-height:36px; }
  .table-panel { border-radius:12px; }
  .table-scroll { overflow:hidden; }
  .data-table { min-width:0; table-layout:fixed; font-size:.65rem; border-collapse:separate; border-spacing:0; }\n  .data-table thead th { height:34px; padding:.45rem .35rem; font-size:.54rem; background:#fff8f8; color:#8b4b55; }\n  .data-table tbody tr { background:#fff; }\n  .data-table tbody tr + tr td { border-top:1px solid #f4e6e8; }\n  .data-table tbody td { height:54px; padding:.45rem .35rem; vertical-align:middle; }
  .data-table th, .data-table td { padding:.55rem .4rem; }
  .data-table thead th:nth-child(4), .data-table tbody td:nth-child(4),
  .data-table thead th:nth-child(5), .data-table tbody td:nth-child(5),
  .data-table thead th:nth-child(6), .data-table tbody td:nth-child(6) { display:none; }
  .th-check, .td-check { width:30px; }
  .data-table thead th:nth-child(2), .data-table thead th:nth-child(3), .data-table thead th:nth-child(7) { font-size:0; }
  .data-table thead th:nth-child(2)::after { content:'Kode'; font-size:.56rem; }
  .data-table thead th:nth-child(3)::after { content:'Customer'; font-size:.56rem; }
  .data-table thead th:nth-child(7)::after { content:'Sales'; font-size:.56rem; }
  .cell-primary { font-size:.69rem; }
  .cell-sub { font-size:.55rem; }
  .link-btn { font-size:.68rem; }
  .data-table .code-tag, .data-table .p-tag { font-size:.55rem; padding:.12rem .3rem; }
  .pagination-bar { padding:.65rem .7rem; gap:.4rem; }
  .pagination-info { font-size:.6rem; text-align:center; }
  .state-box { min-height:180px; padding:1.2rem; font-size:.68rem; }
}
/* ERP visual clone â€” scoped to Customer Existing/Company content only. */
.legacy-page-actions,.workspace-heading,.workspace-summary,.workspace-header { display:none; }
.admin-page { padding:0; gap:0; background:#fff; font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif; }
.tabs-bar { display:flex; align-items:center; min-height:58px; height:58px; padding:10px 20px; gap:10px; border:0; border-bottom:1px solid #e2e8f0; border-radius:0; background:#fff; overflow:hidden; }
.tab-item { height:34px; min-height:34px; padding:0 12px; border-radius:10px; color:#40516a; font-size:12px; font-weight:500; }
.tab-item.active { color:#fff; background:#dc2626; border-color:#dc2626; box-shadow:0 2px 6px rgba(220,38,38,.12); }
.tab-item i,.tab-item strong { display:none; }
.erp-customer-actions { display:flex; align-items:center; gap:8px; margin-left:auto; }
.erp-customer-actions :deep(.p-button) { height:40px; min-height:40px; padding:0 12px; border-radius:10px; font-size:12px; font-weight:600; }
.erp-customer-actions :deep(.p-button:first-child) { min-width:108px; }
.erp-customer-actions :deep(.p-button:nth-child(2)) { min-width:72px; }
.erp-customer-actions :deep(.p-button:nth-child(3)) { min-width:158px; }
.erp-customer-actions :deep(.p-button:last-child) { background:#dc2626; border-color:#dc2626; }
.panel-stack { gap:0; padding:0; }
.filter-panel { display:flex; align-items:flex-end; gap:10px; padding:16px; border:0; border-bottom:1px solid #e2e8f0; border-radius:0; box-shadow:none; }
.filter-grid { gap:8px; }
.filter-field label { font-size:10px; }
.filter-field :deep(.p-select),.search-field { height:40px; border-radius:10px; }
.table-panel { display:flex; flex-direction:column; border:0; border-radius:0; box-shadow:none; }
.table-panel > .pagination-bar { order:-1; height:42px; min-height:42px; padding:0 20px; border:0; border-bottom:1px solid #e2e8f0; background:#fff; }
.pagination-bar { justify-content:flex-start; }
.pagination-info { order:2; margin-left:14px; font-size:10px; }
.pagination-controls { order:1; }
.pagination-controls :deep(.p-button),.pagination-num { width:28px; min-width:28px; height:28px; border-radius:50%; font-size:10px; }
.data-table { min-width:0; width:100%; table-layout:fixed; font-size:11px; }
.data-table thead th { box-sizing:border-box; height:42px; padding:0 10px; border-right:1px solid #e2e2e2; border-bottom:1px solid #e2e2e0; background:#f4f4f4; color:#000; font-size:10px; font-weight:600; line-height:12px; letter-spacing:.07em; text-transform:uppercase; }
.data-table tbody td { box-sizing:border-box; height:67px; padding:8px 10px; border-right:1px solid #e2e2e2; border-bottom:1px solid #e0e0e0; color:#0f172a; font-size:11px; line-height:13px; }
.data-table tbody tr:hover { background:#fff; }
.data-table .cell-primary { color:#0f172a; font-size:11px; font-weight:600; line-height:13px; }
.data-table .cell-text,.data-table .cell-sub { color:#64748b; font-size:10px; font-weight:400; line-height:12px; }
.data-table .code-tag { font-family:Inter,sans-serif; font-size:10px; font-weight:600; }
.data-table :deep(.p-tag) { border-radius:999px; padding:3px 9px; font-size:10px; font-weight:500; }
.data-table input[type="checkbox"] { width:16px; height:16px; accent-color:#e63946; }
.customer-site-table .customer-checkbox {
  width: 18px !important;
  height: 18px !important;
  padding: 0 !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 4px !important;
  background: #fff !important;
  color: transparent !important;
}
.customer-site-table .customer-checkbox.customer-checkbox-selected {
  border-color: #ef4444 !important;
  background: #ef4444 !important;
  color: #fff !important;
}
.company-table .customer-checkbox{display:flex;width:18px!important;height:18px!important;margin:0 auto;align-items:center;justify-content:center;padding:0;border:1px solid #cbd5e1!important;border-radius:4px;background:#fff!important;color:transparent}.company-table .customer-checkbox-selected{border-color:#ef4444!important;background:#ef4444!important;color:#fff!important}
.customer-site-table th:nth-child(1),.customer-site-table td:nth-child(1){width:4%}.customer-site-table th:nth-child(2),.customer-site-table td:nth-child(2){width:18%}.customer-site-table th:nth-child(3),.customer-site-table td:nth-child(3){width:31%}.customer-site-table th:nth-child(4),.customer-site-table td:nth-child(4){width:18%}.customer-site-table th:nth-child(5),.customer-site-table td:nth-child(5){width:14%}.customer-site-table th:nth-child(6),.customer-site-table td:nth-child(6){width:15%}.customer-site-table th:nth-child(7),.customer-site-table td:nth-child(7){width:15%}.customer-site-table th:nth-child(6),.customer-site-table td:nth-child(6){display:none}
.customer-site-table td:nth-child(2) .cell-stack { gap:4px; }.customer-site-table td:nth-child(2) .code-tag { display:block; width:max-content; max-width:100%; padding:3px 8px; border-radius:6px; background:#fff0f1; color:#e63946; }.customer-site-table td:nth-child(2) :deep(.p-tag) { display:inline-flex; width:max-content; max-width:100%; align-items:center; border:1px solid transparent; padding:4px 10px; border-radius:999px; font-size:10px; font-weight:500; line-height:11px; text-align:center; }.customer-site-table td:nth-child(2) :deep(.p-tag-info) { border-color:#bfdbfe; background:#dbeafe; color:#1d4ed8; }.customer-site-table td:nth-child(2) :deep(.p-tag-success) { border-color:#bbf7d0; background:#dcfce7; color:#166534; }.customer-site-table td:nth-child(2) :deep(.p-tag-warn) { border-color:#fed7aa; background:#fff7ed; color:#c2410c; }.customer-site-table td:nth-child(2) :deep(.p-tag-secondary) { border-color:#ddd6fe; background:#ede9fe; color:#6d28d9; }.customer-site-table .link-btn { color:#e63946; font-size:12px; font-weight:600; }
.company-tier-tag { display:inline-flex; width:max-content; align-items:center; justify-content:center; border:1px solid #ddd6fe; border-radius:999px; background:#ede9fe; color:#7c3aed; padding:4px 10px; font-size:11px; font-weight:600; line-height:normal; white-space:nowrap; }
.company-tier-tag.tier-key { border-color:#ddd6fe; background:#ede9fe; color:#7c3aed; }
.company-tier-tag.tier-regular { border-color:#bbf7d0; background:#dcfce7; color:#166534; }
.company-tier-tag.tier-strategic { border-color:#bfdbfe; background:#dbeafe; color:#1d4ed8; }
.company-tier-tag.tier-small { border-color:#fed7aa; background:#fff7ed; color:#c2410c; }
.company-table :deep(.company-tier-tag.tier-key) { border-color:#ddd6fe !important; background:#ede9fe !important; color:#7c3aed !important; }
.company-table :deep(.company-tier-tag.tier-regular) { border-color:#bbf7d0 !important; background:#dcfce7 !important; color:#166534 !important; }
.company-table :deep(.company-tier-tag.tier-strategic) { border-color:#bfdbfe !important; background:#dbeafe !important; color:#1d4ed8 !important; }
.company-table :deep(.company-tier-tag.tier-small) { border-color:#fed7aa !important; background:#fff7ed !important; color:#c2410c !important; }
.company-table :deep(.company-tier-tag .p-tag-label) { font-size:11px !important; font-weight:600 !important; }
.company-table :deep(.company-tier-tag.tier-key .p-tag-label) { color:#7c3aed !important; }
.company-table :deep(.company-tier-tag.tier-regular .p-tag-label) { color:#166534 !important; }
.company-table :deep(.company-tier-tag.tier-strategic .p-tag-label) { color:#1d4ed8 !important; }
.company-table :deep(.company-tier-tag.tier-small .p-tag-label) { color:#c2410c !important; }
.company-table th:nth-child(1),.company-table td:nth-child(1){width:4%}.company-table th:nth-child(2),.company-table td:nth-child(2){width:13%}.company-table th:nth-child(3),.company-table td:nth-child(3){width:25%}.company-table th:nth-child(4),.company-table td:nth-child(4){width:16%}.company-table th:nth-child(5),.company-table td:nth-child(5){width:14%}.company-table th:nth-child(6),.company-table td:nth-child(6){width:12%}.company-table th:nth-child(7),.company-table td:nth-child(7){width:16%}.company-table th:nth-child(n+8),.company-table td:nth-child(n+8){display:none}
.company-table { border-top: 1px solid #e0e0e0; }
.company-table thead th { border-right: 1px solid #e0e0e0; border-bottom: 1px solid #e0e0e0; background: #f4f4f4; color: #000; }
.company-table thead th:last-child { border-right: 0; }
.company-table tbody td { border-right: 1px solid #e0e0e0; }
.company-table tbody td:last-child { border-right: 0; }
.site-search,.company-search { display:flex; flex:0 0 270px; height:40px; min-width:270px; align-items:center; gap:8px; padding:0 12px; border:1px solid #e2e8f0; border-radius:10px; background:#fff; color:#94a3b8; }.site-search i,.company-search i{font-size:12px}.site-search input,.company-search input{width:100%;height:32px;min-height:32px;padding:0;border:0;outline:0;background:transparent;color:#0f172a;font-size:13px}.site-search input::placeholder,.company-search input::placeholder{color:#94a3b8}
@media(max-width:900px){.tabs-bar{height:auto;min-height:58px;flex-wrap:wrap;overflow:visible}.site-search,.company-search{flex-basis:220px}.erp-customer-actions{width:100%;margin-left:0}.filter-panel{display:grid;grid-template-columns:1fr;padding:12px}.filter-grid{grid-template-columns:repeat(2,minmax(0,1fr));}}
@media(max-width:640px){.tabs-bar{padding:10px 12px}.erp-customer-actions :deep(.p-button){flex:1}.data-table{min-width:860px}.table-scroll{overflow-x:auto}.filter-grid{grid-template-columns:1fr}}

/* Keep the customer workspace content-flow compact; the table must follow the toolbar. */
.admin-page {
  min-height: 0 !important;
  height: auto !important;
  justify-content: flex-start !important;
}
.admin-page > .tabs-bar {
  flex: 0 0 auto;
}
.admin-page .panel-stack,
.admin-page .table-panel {
  flex: 0 0 auto;
}
.admin-page > nav[aria-label="Customer management sections"] {
  width: 100%;
  box-sizing: border-box;
  justify-content: flex-start;
}
.admin-page > nav[aria-label="Customer management sections"] > .erp-customer-actions {
  margin-left: auto;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button {
  white-space: nowrap;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:first-child {
  min-width: 132px;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:nth-child(2) {
  min-width: 82px;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:last-child {
  width: 170px;
  min-width: 170px;
  justify-content: center;
  height: 36px;
  gap: 7px;
  border-radius: 10px;
  background: #dc2626;
  padding: 0 12px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  box-shadow: 0 4px 16px rgba(220, 38, 38, .22);
  transition: all 150ms ease;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:last-child:hover {
  filter: brightness(1.1);
}
.admin-page .pagination-controls {
  gap: 4px;
}
.admin-page .pagination-controls .pagination-num,
.admin-page .pagination-controls :deep(.p-button) {
  display: inline-flex;
  width: 30px;
  min-width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  transition: all 150ms ease;
}
.admin-page .pagination-controls .pagination-num {
  color: #64748b;
}
.admin-page .pagination-controls .pagination-num:hover {
  background: #f8fafc;
  color: #1e293b;
}
.admin-page .pagination-controls .pagination-num.is-active {
  background: #fff1f2 !important;
  color: #991b1b !important;
  font-weight: 700;
}
.admin-page .table-panel > .pagination-bar {
  justify-content: flex-start;
  gap: 14px;
}
.admin-page .pagination-bar .pagination-info {
  order: 2;
  margin-left: 0;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  color: #64748b;
}
.admin-page .pagination-controls {
  order: 1;
}
.admin-page .pagination-controls > .p-button:first-child,
.admin-page .pagination-controls > .p-button:last-child {
  display: none;
}
.admin-page .data-table thead th {
  border-right: 1px solid #e0e0e0;
  background: #f4f4f4;
  padding: 12px;
  text-align: left;
  color: #000;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .05em;
  text-transform: uppercase;
  user-select: none;
}
.admin-page .data-table thead th:last-child {
  border-right: 0;
}
.admin-page .customer-site-table thead th:nth-child(6),
.admin-page .customer-site-table tbody td:nth-child(6) {
  display: table-cell;
}
.admin-page .customer-site-table tbody td:nth-child(6) {
  display: table-cell;
}
.admin-page .customer-site-table thead th:nth-child(6) {
  font-size: 11px;
}
.admin-page .customer-site-table thead th:nth-child(6)::before {
  content: none;
}
.admin-page .customer-site-table tbody td:nth-child(6) :deep(.p-tag) {
  display: inline-flex;
}
.admin-page .customer-site-table tbody td:nth-child(7) .cell-text {
  color: #0f172a;
  font-size: 13px;
  font-weight: 600;
}
.admin-page .customer-site-table td:nth-child(2) .cell-stack {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
}
.admin-page .customer-site-table td:nth-child(2) .code-tag {
  padding: 0;
  background: transparent;
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 700;
}
.admin-page .customer-site-table td:nth-child(2) :deep(.p-tag) {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
  gap: 4px;
  overflow: hidden;
  margin-top: 4px;
  border-radius: 999px;
  padding: 4px 10px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 11px;
  font-weight: 600;
  transition: color 150ms ease, box-shadow 150ms ease;
}
.admin-page .customer-site-table td:nth-child(2) :deep(.p-tag-warn) {
  border-color: #ddd6fe;
  background: #ede9fe;
  color: #7c3aed;
}
.admin-page .customer-site-table td:nth-child(2) :deep(.p-tag-info) {
  border-color: #bfdbfe !important;
  background: #dbeafe !important;
  color: #1d4ed8 !important;
}
.admin-page .customer-site-table td:nth-child(2) :deep(.p-tag.p-tag-info),
.admin-page .customer-site-table td:nth-child(2) :deep(.p-tag-info) {
  background: #dbeafe !important;
  border: 1px solid #bfdbfe !important;
  color: #1d4ed8 !important;
  -webkit-text-fill-color: #1d4ed8 !important;
}
.admin-page .customer-site-table tbody td:nth-child(n+3) {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
}
.admin-page .customer-site-table tbody td:nth-child(3) .link-btn,
.admin-page .customer-site-table tbody td:nth-child(3) .cell-primary {
  display: block;
  margin: 0;
  max-width: 100%;
  overflow: hidden;
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.admin-page .customer-site-table tbody td:nth-child(n+3) .cell-sub,
.admin-page .customer-site-table tbody td:nth-child(n+3) .cell-text {
  display: block;
  margin-top: 2px;
  color: #64748b;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  line-height: 1.25;
}
.admin-page .customer-site-table tbody td:nth-child(3) .link-btn:hover {
  color: #0f172a;
  text-decoration: underline;
}
.admin-page .customer-site-table tbody td:nth-child(4) .cell-text {
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.25;
}
.admin-page .customer-site-table tbody td:nth-child(5) .cell-primary {
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.25;
}
.admin-page .customer-site-table tbody td:nth-child(5) .cell-sub {
  margin-top: 2px;
  color: #64748b;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  line-height: 1.25;
}
.admin-page .customer-site-table th:nth-child(1), .admin-page .customer-site-table td:nth-child(1) { width: 3.5%; }
.admin-page .customer-site-table th:nth-child(2), .admin-page .customer-site-table td:nth-child(2) { width: 16%; }
.admin-page .customer-site-table th:nth-child(3), .admin-page .customer-site-table td:nth-child(3) { width: 27%; }
.admin-page .customer-site-table th:nth-child(4), .admin-page .customer-site-table td:nth-child(4) { width: 16%; }
.admin-page .customer-site-table th:nth-child(5), .admin-page .customer-site-table td:nth-child(5) { width: 13%; }
.admin-page .customer-site-table th:nth-child(6), .admin-page .customer-site-table td:nth-child(6) { width: 13%; }
.admin-page .customer-site-table th:nth-child(7), .admin-page .customer-site-table td:nth-child(7) { width: 14%; }
.admin-page .customer-site-table tbody td:nth-child(6) :deep(.p-tag) {
  margin: 0;
  border: 0 !important;
  background: transparent !important;
  padding: 0;
  color: #0f172a !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.25;
}
.admin-page .customer-site-table thead th:nth-child(6)::before {
  content: 'SEGMENT';
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .05em;
}
.admin-page .customer-site-table thead th:nth-child(6) {
  font-size: 0;
}
.pagination-settings {
  order: 3;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
}
.pagination-settings > label:first-child { width: 80px; min-width: 80px; }
.pagination-settings > label:nth-child(2) { width: 65px; min-width: 65px; }
.pagination-settings label {
  position: relative;
  display: flex;
  cursor: pointer;
  width: 80px;
  height: 34px;
  align-items: center;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0 10px;
  background: #fff;
}
.pagination-settings label:hover { border-color: #cbd5e1; }
.pagination-settings label:nth-child(2) { width: 65px; }
.pagination-settings label span {
  position: absolute;
  top: -7px;
  left: 8px;
  padding: 0 4px;
  background: #fff;
  color: #475569;
  font-size: 10px;
  font-weight: 600;
  line-height: 1;
}
.pagination-settings label span,
.pagination-settings label > span {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 10px;
  font-weight: 600;
  line-height: 10px;
}
.pagination-settings select,
.pagination-settings input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: #334155;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  line-height: 16px;
}
.pagination-settings select { cursor: pointer; }
.pagination-settings .page-size-control {
  position: relative;
  display: flex;
  width: 80px;
  height: 34px;
  align-items: center;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0 10px;
  background: #fff;
}
.pagination-settings .page-size-control select {
  height: 32px;
  padding: 0;
  appearance: none;
}
.pagination-settings > button:not(.pagination-refresh) {
  box-sizing: border-box;
  height: 34px;
  width: 48px;
  min-width: 48px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0 12px;
  background: #fff;
  color: #475569;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 16px;
}
.pagination-refresh {
  display: flex;
  width: 30px;
  height: 30px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: #475569;
  font-size: 18px;
}
.pagination-refresh:hover { background: #f1f5f9; color: #1e293b; }
.company-pagination {
  width: 100%;
  height: 42px;
  min-height: 42px;
  padding: 0 20px;
  align-items: center;
  gap: 12px;
  border-top: 1px solid #e2e8f0;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
  justify-content: flex-start;
  box-sizing: border-box;
}
.company-pagination .pagination-controls {
  order: 1;
  display: flex;
  height: 30px;
  align-items: center;
  gap: 4px;
}
.company-pagination .pagination-info {
  order: 2;
  margin-left: 0;
  color: #64748b;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  line-height: 16px;
}
.company-pagination .pagination-settings {
  order: 3;
  height: 34px;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}
.company-pagination .pagination-controls > :deep(.p-button:first-child),
.company-pagination .pagination-controls > :deep(.p-button:last-child) {
  display: none;
}
.company-pagination .pagination-num.is-active {
  width: 30px;
  min-width: 30px;
  height: 30px;
  min-height: 30px;
  padding: 0 !important;
  border: 0 !important;
  border-radius: 999px;
  background: #fff1f2 !important;
  color: #991b1b !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 700;
  box-shadow: none !important;
}
.company-pagination .pagination-settings label {
  height: 34px;
}
.company-pagination .pagination-settings > button:not(.pagination-refresh) {
  height: 34px;
}
@media (max-width: 900px) {
  .pagination-settings { margin-left: 0; flex-wrap: wrap; }
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:first-child {
  width: 82px;
  min-width: 82px;
  justify-content: center;
  gap: 7px;
  border-color: #fecaca;
  background: #fff5f5;
  color: #b91c1b;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 600;
  transition: all 150ms ease;
}
.admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions button:first-child:hover {
  background: #fff1f2;
}
.admin-page > nav[aria-label="Customer management sections"] > label input {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 400 !important;
  letter-spacing: 0 !important;
}
.admin-page > nav[aria-label="Customer management sections"] > button {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0;
  white-space: nowrap;
}
.admin-page > nav[aria-label="Customer management sections"] > div:first-child {
  display: inline-flex;
  flex-shrink: 0;
  max-width: 100%;
  flex-wrap: nowrap;
  gap: 4px;
  border-radius: 16px;
  background: #f1f5f9;
  padding: 3px;
}
.admin-page > nav[aria-label="Customer management sections"] > div:first-child > button {
  display: flex;
  height: 36px;
  flex-shrink: 0;
  align-items: center;
  border-radius: 12px;
  padding: 0 14px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  transition: all 150ms ease;
}
.admin-page > nav[aria-label="Customer management sections"] > div:first-child > button:not(.bg-\[\#dc2626\]) {
  color: #64748b;
}
@media (max-width: 900px) {
  .admin-page > nav[aria-label="Customer management sections"] {
    flex-wrap: wrap;
  }
  .admin-page > nav[aria-label="Customer management sections"] .erp-customer-actions {
    width: 100%;
    margin-left: 0;
  }
}
.delete-warning{display:flex;align-items:flex-start;gap:12px;padding:14px;border:1px solid #fecaca;border-radius:12px;background:#fff7f7;color:#334155}.delete-warning>i{display:flex;width:32px;height:32px;flex:none;align-items:center;justify-content:center;border-radius:999px;background:#fee2e2;color:#dc2626;font-size:15px}.delete-warning strong{display:block;color:#0f172a;font-size:13px;font-weight:700}.delete-warning p{margin:4px 0 0;color:#64748b;font-size:12px;line-height:18px}.admin-page :deep(.p-dialog-footer){display:flex;justify-content:flex-end;gap:8px}.admin-page :deep(.p-dialog-footer .p-button){min-height:38px;border-radius:10px}
/* â”€â”€ Master Data mode: full-width stretch layout â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
.admin-page.master-mode {
  padding: 0.4rem 0.6rem 0.6rem;
  gap: 0.4rem;
}
.admin-page.master-mode .workspace-header { display: none; }
.admin-page.master-mode .tabs-bar { flex-shrink: 0; }
.admin-page.master-mode > .md-root {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>

