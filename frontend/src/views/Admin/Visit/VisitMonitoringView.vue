<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import vTooltip from 'primevue/tooltip'
import { getAdminVisits, getSalesExecutives, deleteVisit as apiDeleteVisit } from '../../../api/crm'
import type { VisitMonitoringItem, VisitMonitoringFilters, SalesExecutiveOption, ProspectStatus, ProspectVisit } from '../../../types/crm'
import VisitSelfieModal from './VisitSelfieModal.vue'

interface GroupedVisitRow {
  prospectId: string
  customerName: string
  customerCategory: string
  industryGroup: string
  formattedAddress: string
  phoneNumber: string
  prospectLatitude: number | null
  prospectLongitude: number | null
  prospectStatus: ProspectStatus
  latestVisit: VisitMonitoringItem
  visitCount: number
}

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const error = ref('')
const visits = ref<VisitMonitoringItem[]>([])
const salesExecutives = ref<SalesExecutiveOption[]>([])
const selfieModalItem = ref<VisitMonitoringItem | null>(null)
const showDetailModal = ref(false)
const detailProspectName = ref('')
const detailProspectRow = ref<GroupedVisitRow | null>(null)
const prospectVisits = ref<ProspectVisit[]>([])
const prospectVisitsLoading = ref(false)
const detailError = ref('')
const deleteModalItem = ref<VisitMonitoringItem | null>(null)
const showDeleteModal = ref(false)
const deleteBusy = ref(false)
const showVisitResultModal = ref(false)
const visitResultItem = ref<VisitMonitoringItem | null>(null)
const actionDialogVisible = ref(false)
const actionTarget = ref<GroupedVisitRow | null>(null)
const showFilters = ref(false)
const selectedVisitIds = ref<Set<string>>(new Set())
const showBulkDeleteModal = ref(false)
const bulkDeleteBusy = ref(false)

const filters = ref<VisitMonitoringFilters>({
  dateFrom: '',
  dateTo: '',
  salesExecutiveId: '',
  customerName: '',
  radiusStatus: 'ALL',
})

const radiusOptions = [
  { label: 'All Visits', value: 'ALL' },
  { label: 'Inside Radius', value: 'INSIDE' },
  { label: 'Outside Radius', value: 'OUTSIDE' },
]

const salesOptions = computed(() => {
  return [{ label: 'All Sales', value: '' }, ...salesExecutives.value.map((s) => ({ label: s.fullName, value: s.id }))]
})

const selectedSales = ref('')
const selectedRadius = ref('ALL')
const visitPage = ref(1)
const visitPageSize = ref(10)
const visitGoToPage = ref(1)

const totalVisits = computed(() => visits.value.length)
const insideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'INSIDE').length)
const outsideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'OUTSIDE').length)
const openVisits = computed(() => visits.value.filter((v) => !v.checkOutAt).length)
const totalProspects = computed(() => groupedVisits.value.length)
const selectableVisitIds = computed(() => filteredGroupedVisits.value.map((row) => row.latestVisit.id))
const allVisitsSelected = computed(() => selectableVisitIds.value.length > 0 && selectableVisitIds.value.every((id) => selectedVisitIds.value.has(id)))

const customerSearch = ref('')
watch(() => route.query.search, (value) => { customerSearch.value = typeof value === 'string' ? value : '' }, { immediate: true })
const groupedVisits = computed(() => {
  const groups = new Map<string, VisitMonitoringItem[]>()
  for (const v of visits.value) {
    const list = groups.get(v.prospectId)
    if (list) list.push(v)
    else groups.set(v.prospectId, [v])
  }
  const result: GroupedVisitRow[] = []
  for (const [prospectId, items] of groups) {
    const sorted = [...items].sort((a, b) => new Date(b.checkInAt).getTime() - new Date(a.checkInAt).getTime())
    const latest = sorted[0]
    result.push({
      prospectId,
      customerName: latest.customerName,
      customerCategory: latest.customerCategory,
      industryGroup: latest.industryGroup,
      formattedAddress: latest.formattedAddress,
      phoneNumber: latest.phoneNumber,
      prospectLatitude: latest.prospectLatitude,
      prospectLongitude: latest.prospectLongitude,
      prospectStatus: latest.prospectStatus,
      latestVisit: latest,
      visitCount: latest.visitCount,
    })
  }
  return result
})

const filteredGroupedVisits = computed(() => {
  let result = groupedVisits.value
  if (customerSearch.value) {
    const q = customerSearch.value.toLowerCase()
    result = result.filter((v) => v.customerName.toLowerCase().includes(q) || v.industryGroup.toLowerCase().includes(q))
  }
  return result
})

const visitTotalPages = computed(() => Math.max(1, Math.ceil(filteredGroupedVisits.value.length / visitPageSize.value)))
const paginatedVisits = computed(() => {
  const start = (visitPage.value - 1) * visitPageSize.value
  return filteredGroupedVisits.value.slice(start, start + visitPageSize.value)
})

watch([customerSearch, selectedSales, selectedRadius], () => {
  visitPage.value = 1
  visitGoToPage.value = 1
})
watch(visitTotalPages, (total) => {
  if (visitPage.value > total) visitPage.value = total
  visitGoToPage.value = visitPage.value
})

function goToVisitPage(page: number) {
  const nextPage = Math.max(1, Math.min(page, visitTotalPages.value))
  visitPage.value = nextPage
  visitGoToPage.value = nextPage
}

function applyVisitPagination() {
  goToVisitPage(Number(visitGoToPage.value) || 1)
}

function formatDuration(seconds: number | undefined): string {
  if (seconds == null) return '—'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function formatDistance(meters: number): string {
  if (meters >= 1000) return `${(meters / 1000).toFixed(1)} km`
  return `${Math.round(meters)} m`
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatDateShort(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function siteLocationParts(address: string) {
  const parts = address.split(',').map((part) => part.trim()).filter(Boolean)
  if (parts.length < 2) return { city: address || '—', province: '' }
  const hasCountry = ['indonesia', 'indonésie'].includes((parts.at(-1) || '').toLowerCase())
  const provinceIndex = hasCountry ? -2 : -1
  return { city: parts.at(provinceIndex - 1) || parts[0], province: parts.at(provinceIndex) || '' }
}

function statusSeverity(status: string): string {
  switch (status) {
    case 'WON': case 'CONVERTED': return 'success'
    case 'LOST': return 'danger'
    case 'NEW_LEAD': return 'info'
    case 'CONTACTED': case 'INTERESTED': return 'warn'
    default: return 'secondary'
  }
}

function statusLabel(status: string): string {
  return status.replace(/_/g, ' ')
}

function statusBadgeClass(status: string) {
  switch (status) {
    case 'LOST': return 'status-badge-lost'
    case 'WON':
    case 'CONVERTED':
    case 'INTERESTED':
    case 'QUALIFIED': return 'status-badge-success'
    case 'PROPOSAL_SENT':
    case 'NEGOTIATION': return 'status-badge-warning'
    default: return 'status-badge-info'
  }
}

function applyFilters() {
  filters.value.salesExecutiveId = selectedSales.value
  filters.value.radiusStatus = selectedRadius.value
  fetchData()
}

function resetFilters() {
  filters.value = { dateFrom: '', dateTo: '', salesExecutiveId: '', customerName: '', radiusStatus: 'ALL' }
  selectedSales.value = ''
  selectedRadius.value = 'ALL'
  fetchData()
}

async function fetchData() {
  loading.value = true
  error.value = ''
  try {
    visits.value = await getAdminVisits(filters.value)
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to load visit data.'
  } finally {
    loading.value = false
  }
}

async function loadSalesExecutives() {
  try {
    salesExecutives.value = await getSalesExecutives()
  } catch { /* silent */ }
}

async function openDetail(item: GroupedVisitRow) {
  detailProspectName.value = item.customerName
  detailProspectRow.value = item
  detailError.value = ''
  showDetailModal.value = true
  prospectVisitsLoading.value = false
  const raw = visits.value.filter((v) => v.prospectId === item.prospectId)
  prospectVisits.value = raw
  if (!raw.length) detailError.value = 'No visit records found for this prospect.'
}

function haversineDistance(lat1: number, lng1: number, lat2: number, lng2: number): number {
  const R = 6371000
  const dLat = (lat2 - lat1) * Math.PI / 180
  const dLng = (lng2 - lng1) * Math.PI / 180
  const a = Math.sin(dLat / 2) ** 2 + Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * Math.sin(dLng / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function makeVisitItem(visit: ProspectVisit): VisitMonitoringItem {
  const row = detailProspectRow.value
  let dist = 0
  let radius: 'INSIDE' | 'OUTSIDE' | 'UNKNOWN' = 'UNKNOWN'
  if (row?.prospectLatitude != null && row?.prospectLongitude != null) {
    dist = haversineDistance(visit.checkInLatitude, visit.checkInLongitude, row.prospectLatitude, row.prospectLongitude)
    radius = dist <= 100 ? 'INSIDE' : 'OUTSIDE'
  }
  return {
    ...visit,
    customerName: row?.customerName || '',
    customerCategory: row?.customerCategory || '',
    industryGroup: row?.industryGroup || '',
    formattedAddress: row?.formattedAddress || '',
    phoneNumber: row?.phoneNumber || '',
    prospectLatitude: row?.prospectLatitude || null,
    prospectLongitude: row?.prospectLongitude || null,
    distanceMeters: Math.round(dist),
    radiusStatus: radius,
    prospectStatus: row?.prospectStatus || 'NEW_LEAD',
    visitCount: row?.visitCount || 0,
  } as VisitMonitoringItem
}

function openSelfie(item: VisitMonitoringItem) {
  selfieModalItem.value = item
}

function openVisitResult(item: VisitMonitoringItem) {
  visitResultItem.value = item
  showVisitResultModal.value = true
}

function goToProspect(item: GroupedVisitRow | VisitMonitoringItem) {
  router.push({ name: 'AdminProspectReview', params: { id: item.prospectId } })
}

function toggleVisitSelection(id: string) {
  const next = new Set(selectedVisitIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedVisitIds.value = next
}

function toggleAllVisitSelection() {
  selectedVisitIds.value = allVisitsSelected.value ? new Set() : new Set(selectableVisitIds.value)
}

async function executeBulkDelete() {
  if (!selectedVisitIds.value.size) return
  bulkDeleteBusy.value = true
  try {
    await Promise.all([...selectedVisitIds.value].map((id) => apiDeleteVisit(id)))
    selectedVisitIds.value = new Set()
    showBulkDeleteModal.value = false
    await fetchData()
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to delete selected visits.'
  } finally {
    bulkDeleteBusy.value = false
  }
}

function openActions(item: GroupedVisitRow) {
  actionTarget.value = item
  actionDialogVisible.value = true
}

function closeActions() {
  actionDialogVisible.value = false
  actionTarget.value = null
}

function viewSelectedDetails() {
  if (!actionTarget.value) return
  const target = actionTarget.value
  closeActions()
  openDetail(target)
}

function viewSelectedProspect() {
  if (!actionTarget.value) return
  const target = actionTarget.value
  closeActions()
  goToProspect(target)
}

function downloadVisitData(item: VisitMonitoringItem) {
  const data = {
    evidenceId: item.id,
    prospectId: item.prospectId,
    salesExecutive: item.salesExecutiveName,
    customer: item.customerName,
    industryGroup: item.industryGroup,
    formattedAddress: item.formattedAddress,
    phoneNumber: item.phoneNumber,
    checkIn: item.checkInAt,
    checkOut: item.checkOutAt,
    visitNotes: item.visitNotes,
    followUpNotes: item.followUpNotes,
    distance: item.distanceMeters,
    radiusStatus: item.radiusStatus,
    status: item.prospectStatus,
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `visit-${item.id.slice(0, 8)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function viewGpsLocation(lat: number, lng: number) {
  window.open(`https://www.google.com/maps?q=${lat},${lng}`, '_blank')
}

function confirmDelete(item: VisitMonitoringItem) {
  deleteModalItem.value = item
  showDeleteModal.value = true
}

async function executeDelete() {
  if (!deleteModalItem.value) return
  deleteBusy.value = true
  try {
    await apiDeleteVisit(deleteModalItem.value.id)
    visits.value = visits.value.filter((v) => v.id !== deleteModalItem.value!.id)
    prospectVisits.value = prospectVisits.value.filter((v) => v.id !== deleteModalItem.value!.id)
    showDeleteModal.value = false
    deleteModalItem.value = null
  } catch (e: any) {
    error.value = e?.response?.data?.error?.message || 'Failed to delete visit.'
  } finally {
    deleteBusy.value = false
  }
}

onMounted(() => {
  fetchData()
  loadSalesExecutives()
})
</script>

<template>
  <section class="visit-page admin-page">
    <header class="workspace-header">
      <div class="workspace-heading">
        <div class="page-title-wrapper">
          <span class="eyebrow">Field Operations</span>
          <h1>Visit Monitoring</h1>
          <p class="muted">Track GPS check-ins, selfie evidence, visit duration, and outcomes across your sales team.</p>
        </div>
      </div>

      <div class="summary-strip">
        <button type="button" class="summary-item">
          <i class="pi pi-users si-blue" />
          <span>Prospects</span>
          <strong>{{ totalProspects }}</strong>
        </button>
        <button type="button" class="summary-item">
          <i class="pi pi-map-marker si-emerald" />
          <span>Visits</span>
          <strong>{{ totalVisits }}</strong>
        </button>
        <button type="button" class="summary-item">
          <i class="pi pi-exclamation-circle si-red" />
          <span>Outside</span>
          <strong>{{ outsideCount }}</strong>
        </button>
        <button type="button" class="summary-item">
          <i class="pi pi-clock si-amber" />
          <span>Open</span>
          <strong>{{ openVisits }}</strong>
        </button>
      </div>
    </header>

    <Message v-if="error" severity="error" class="page-message">{{ error }}</Message>

    <nav class="visit-erp-toolbar" aria-label="Visit monitoring controls">
      <label class="visit-search"><i class="pi pi-search" /><input v-model="customerSearch" placeholder="Search customer or industry group" /></label>
      <Button label="More Filters" icon="pi pi-sliders-h" severity="secondary" outlined size="small" @click="showFilters = !showFilters" />
      <button type="button" class="visit-trash-button" :disabled="!selectedVisitIds.size" :class="{ 'visit-trash-disabled': !selectedVisitIds.size }" @click="showBulkDeleteModal = true"><i class="pi pi-trash" /><span>Trash{{ selectedVisitIds.size ? ` (${selectedVisitIds.size})` : '' }}</span></button>
    </nav>

    <div v-if="showFilters" class="filter-panel">
      <div class="filter-grid">
        <div class="filter-field">
          <label>Date From</label>
          <input type="date" class="date-input" v-model="filters.dateFrom" @change="fetchData" />
        </div>
        <div class="filter-field">
          <label>Date To</label>
          <input type="date" class="date-input" v-model="filters.dateTo" @change="fetchData" />
        </div>
        <div class="filter-field">
          <label>Sales Executive</label>
          <Select v-model="selectedSales" :options="salesOptions" optionLabel="label" optionValue="value" placeholder="All Sales" @change="applyFilters" />
        </div>
        <div class="filter-field">
          <label>Radius</label>
          <Select v-model="selectedRadius" :options="radiusOptions" optionLabel="label" optionValue="value" @change="applyFilters" />
        </div>
        <div class="filter-field filter-action">
          <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetFilters" />
        </div>
      </div>
    </div>

    <section class="table-panel">
      <div class="visit-pagination-bar"><div class="visit-pagination-info"><button type="button" class="visit-page-number" :disabled="visitPage <= 1" @click="goToVisitPage(visitPage - 1)">{{ visitPage }}</button><span>Page {{ visitPage }} of {{ visitTotalPages }} / {{ filteredGroupedVisits.length }} records</span></div><div class="visit-pagination-settings"><label><span>Page size</span><select v-model.number="visitPageSize"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label><label><span>Go to</span><input v-model.number="visitGoToPage" type="number" min="1" :max="visitTotalPages" /></label><button type="button" @click="applyVisitPagination">Set</button><button type="button" title="Refresh visit data" @click="fetchData"><i class="pi pi-refresh" /></button></div></div>
      <div class="table-heading">
        <div>
          <strong>Prospect Visit Overview</strong>
          <span>{{ filteredGroupedVisits.length }} prospects with visit activity</span>
        </div>
        <div class="table-heading-meta">
          <Tag :value="`${insideCount} inside`" severity="success" />
          <Tag :value="`${outsideCount} outside`" severity="danger" />
        </div>
      </div>

      <div v-if="loading" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading visits...</span>
      </div>
      <div v-else-if="!filteredGroupedVisits.length" class="state-box">
        <div class="state-icon-wrap"><i class="pi pi-map-marker" /></div>
        <strong>No prospects found</strong>
        <span class="muted">No prospects with matching visit records.</span>
      </div>
      <div v-else class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th class="th-check"><button type="button" class="visit-checkbox" :class="{ 'visit-checkbox-selected': allVisitsSelected }" aria-label="Select all visits" @click.stop="toggleAllVisitSelection"><i v-if="allVisitsSelected" class="pi pi-check" /></button></th>
              <th>Prospect</th>
              <th>Category</th>
              <th>Site Location</th>
              <th>Sales Executive</th>
              <th>Last Visit</th>
              <th>Status</th>
              <th>Visits</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in paginatedVisits" :key="row.prospectId" class="visit-row" @click="openActions(row)">
              <td class="td-check"><button type="button" class="visit-checkbox" :class="{ 'visit-checkbox-selected': selectedVisitIds.has(row.latestVisit.id) }" :aria-label="`Select visit ${row.customerName}`" @click.stop="toggleVisitSelection(row.latestVisit.id)"><i v-if="selectedVisitIds.has(row.latestVisit.id)" class="pi pi-check" /></button></td>
              <td>
                <div class="prospect-cell">
                  <strong class="prospect-name">{{ row.customerName }}</strong>
                  <small class="prospect-meta">{{ row.industryGroup }} · {{ row.customerCategory }}</small>
                  <small v-if="row.formattedAddress" class="prospect-address">{{ row.formattedAddress }}</small>
                  <small v-if="row.phoneNumber" class="prospect-phone"><i class="pi pi-phone" /> {{ row.phoneNumber }}</small>
                </div>
              </td>
              <td>
                <div class="cell-stack">
                  <span class="cell-primary">{{ row.customerCategory || '—' }}</span>
                  <span class="cell-sub">{{ row.industryGroup || '—' }}</span>
                </div>
              </td>
              <td>
                <div class="cell-stack">
                  <span class="cell-primary">{{ siteLocationParts(row.formattedAddress).city }}</span>
                  <span class="cell-sub">{{ siteLocationParts(row.formattedAddress).province }}</span>
                </div>
              </td>
              <td>
                <span class="cell-text">{{ row.latestVisit.salesExecutiveName }}</span>
              </td>
              <td>
                <div class="datetime-cell">
                  <span class="date-val">{{ formatDateShort(row.latestVisit.checkInAt) }}</span>
                  <span class="time-val">{{ formatTime(row.latestVisit.checkInAt) }}</span>
                </div>
              </td>
              <td>
                <span class="status-badge" :class="statusBadgeClass(row.prospectStatus)">{{ statusLabel(row.prospectStatus) }}</span>
              </td>
              <td>
                <span class="visit-count-badge">{{ row.visitCount }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <VisitSelfieModal :item="selfieModalItem" @close="selfieModalItem = null" />

    <Dialog v-model:visible="actionDialogVisible" modal :draggable="false" header="Visit Actions" :style="{ width: 'min(420px, calc(100vw - 2rem))' }" @hide="closeActions">
      <p v-if="actionTarget" class="action-dialog-subtitle">Choose an action for <strong>{{ actionTarget.customerName }}</strong>.</p>
      <div class="visit-action-list">
        <button type="button" class="visit-action-card" @click="viewSelectedDetails">
          <span class="visit-action-icon"><i class="pi pi-info-circle" /></span>
          <span><strong>View Visit Details</strong><small>Review all visit records and evidence.</small></span>
          <i class="pi pi-chevron-right" />
        </button>
        <button type="button" class="visit-action-card" @click="viewSelectedProspect">
          <span class="visit-action-icon"><i class="pi pi-eye" /></span>
          <span><strong>View Prospect</strong><small>Open the full prospect profile.</small></span>
          <i class="pi pi-chevron-right" />
        </button>
      </div>
    </Dialog>

    <Dialog v-model:visible="showBulkDeleteModal" modal header="Delete Selected Visits" class="bulk-delete-dialog" :style="{ width: 'min(100%, 420px)' }" :closable="!bulkDeleteBusy">
      <div class="bulk-delete-warning"><i class="pi pi-exclamation-triangle" /><div><strong>Delete selected visits?</strong><p>This action will permanently remove {{ selectedVisitIds.size }} selected visit{{ selectedVisitIds.size === 1 ? '' : 's' }} and cannot be undone.</p></div></div>
      <template #footer><Button label="Cancel" severity="secondary" outlined :disabled="bulkDeleteBusy" @click="showBulkDeleteModal = false" /><Button label="Delete" icon="pi pi-trash" severity="danger" :loading="bulkDeleteBusy" @click="executeBulkDelete" /></template>
    </Dialog>

    <Dialog v-model:visible="showDeleteModal" modal header="Delete Visit" :style="{ width: 'min(100%, 420px)' }" :closable="!deleteBusy">
      <p v-if="deleteModalItem" style="margin:0;font-size:0.85rem;line-height:1.5;">
        Delete visit record for <strong>{{ deleteModalItem.customerName }}</strong> checked in by <strong>{{ deleteModalItem.salesExecutiveName }}</strong> on <strong>{{ formatDateShort(deleteModalItem.checkInAt) }}</strong>?
      </p>
      <p style="margin:0.5rem 0 0;font-size:0.78rem;color:var(--text-muted);">This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined :disabled="deleteBusy" @click="showDeleteModal = false" />
        <Button label="Delete" icon="pi pi-trash" severity="danger" :loading="deleteBusy" @click="executeDelete" />
      </template>
    </Dialog>

    <!-- Visit Result Modal -->
    <Dialog v-model:visible="showVisitResultModal" modal header="Visit Result" :style="{ width: 'min(100%, 420px)' }" :closable="true">
      <template v-if="visitResultItem">
        <div class="result-modal-body">
          <div class="result-customer">
            <strong>{{ visitResultItem.customerName }}</strong>
            <span>{{ visitResultItem.salesExecutiveName }} · {{ formatDateShort(visitResultItem.checkInAt) }}</span>
          </div>
          <div class="result-fields">
            <div class="result-field">
              <span class="result-label">Visit Result</span>
              <span class="result-value" :class="{ 'result-empty': !visitResultItem.visitResult }">{{ visitResultItem.visitResult || '—' }}</span>
            </div>
            <div class="result-field">
              <span class="result-label">Visit Outcome</span>
              <span class="result-value" :class="{ 'result-empty': !visitResultItem.visitOutcome }">{{ visitResultItem.visitOutcome || '—' }}</span>
            </div>
            <div class="result-field" v-if="visitResultItem.visitNotes">
              <span class="result-label">Visit Notes</span>
              <span class="result-value">{{ visitResultItem.visitNotes }}</span>
            </div>
            <div class="result-field" v-if="visitResultItem.followUpNotes">
              <span class="result-label">Follow-up Notes</span>
              <span class="result-value">{{ visitResultItem.followUpNotes }}</span>
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showVisitResultModal = false" />
      </template>
    </Dialog>

    <!-- Visit Detail Modal -->
    <Dialog v-model:visible="showDetailModal" modal :style="{ width: 'min(100%, 760px)' }" :closable="true" class="visit-detail-dialog">
      <template #header>
        <div class="visit-detail-header">
          <span class="visit-detail-icon"><i class="pi pi-map-marker" /></span>
          <div><span>VISIT ACTIVITY</span><strong>{{ detailProspectName }}</strong><small>{{ prospectVisits.length }} visit record{{ prospectVisits.length === 1 ? '' : 's' }} · Evidence timeline</small></div>
        </div>
      </template>
      <template v-if="prospectVisitsLoading">
        <div class="state-box">
          <i class="pi pi-spin pi-spinner state-icon" />
          <span>Loading visit details...</span>
        </div>
      </template>
      <template v-else-if="detailError">
        <div class="state-box">
          <div class="state-icon-wrap"><i class="pi pi-exclamation-triangle" /></div>
          <strong>Error loading visits</strong>
          <span class="muted">{{ detailError }}</span>
        </div>
      </template>
      <template v-else-if="prospectVisits.length === 0">
        <div class="state-box">
          <div class="state-icon-wrap"><i class="pi pi-map-marker" /></div>
          <strong>No visits found</strong>
          <span class="muted">No visit records for this prospect.</span>
        </div>
      </template>
      <template v-else>
        <div class="detail-visits-list">
          <div v-for="(pv, idx) in prospectVisits" :key="pv.id" class="detail-visit-card">
            <div class="visit-card-header">
              <span class="visit-card-num">Visit #{{ prospectVisits.length - idx }}</span>
              <span v-if="!pv.checkOutAt" class="open-badge">Open</span>
            </div>
            <div class="visit-card-body">
              <div class="visit-card-col">
                <div class="detail-row"><span class="detail-label">Check In</span><strong>{{ formatDateShort(pv.checkInAt) }} {{ formatTime(pv.checkInAt) }}</strong></div>
                <div class="detail-row" v-if="pv.checkOutAt"><span class="detail-label">Check Out</span><strong>{{ formatDateShort(pv.checkOutAt) }} {{ formatTime(pv.checkOutAt) }}</strong></div>
                <div class="detail-row"><span class="detail-label">Sales Executive</span><strong>{{ pv.salesExecutiveName }}</strong></div>
                <div class="detail-row" v-if="pv.visitResult"><span class="detail-label">Visit Result</span><strong>{{ pv.visitResult }}</strong></div>
                <div class="detail-row" v-if="pv.visitOutcome"><span class="detail-label">Visit Outcome</span><strong>{{ pv.visitOutcome }}</strong></div>
                <div class="detail-row" v-if="pv.visitNotes"><span class="detail-label">Visit Notes</span><strong>{{ pv.visitNotes }}</strong></div>
                <div class="detail-row" v-if="pv.followUpNotes"><span class="detail-label">Follow-up Notes</span><strong>{{ pv.followUpNotes }}</strong></div>
              </div>
              <div class="visit-card-col">
                <div class="detail-row"><span class="detail-label">Check-in GPS</span><strong>{{ pv.checkInLatitude.toFixed(6) }}, {{ pv.checkInLongitude.toFixed(6) }}</strong></div>
                <div class="detail-row" v-if="pv.checkOutLatitude"><span class="detail-label">Check-out GPS</span><strong>{{ pv.checkOutLatitude.toFixed(6) }}, {{ pv.checkOutLongitude?.toFixed(6) }}</strong></div>
                <div class="detail-row"><span class="detail-label">Distance</span><strong>{{ formatDistance(makeVisitItem(pv).distanceMeters) }}</strong></div>
                <div class="detail-row"><span class="detail-label">Radius</span>
                  <Tag
                    :value="makeVisitItem(pv).radiusStatus === 'INSIDE' ? 'Inside' : makeVisitItem(pv).radiusStatus === 'OUTSIDE' ? 'Outside' : 'Unknown'"
                    :severity="makeVisitItem(pv).radiusStatus === 'INSIDE' ? 'success' : makeVisitItem(pv).radiusStatus === 'OUTSIDE' ? 'danger' : 'secondary'"
                    size="small"
                  />
                </div>
              </div>
            </div>
            <div class="visit-card-actions">
              <Button v-tooltip.top="'Selfie Evidence'" icon="pi pi-camera" text rounded size="small" class="act-view" @click="openSelfie(makeVisitItem(pv))" />
              <Button v-tooltip.top="'Download Evidence'" icon="pi pi-download" text rounded size="small" class="act-edit" @click="downloadVisitData(makeVisitItem(pv))" />
              <Button v-tooltip.top="'View GPS'" icon="pi pi-map" text rounded size="small" class="act-map" @click="viewGpsLocation(pv.checkInLatitude, pv.checkInLongitude)" />
              <Button v-tooltip.top="'Delete Visit'" icon="pi pi-trash" text rounded size="small" class="act-delete" @click="confirmDelete(makeVisitItem(pv))" />
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetailModal = false" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.visit-page {
  box-sizing: border-box;
  display: flex;
  min-width: 0;
  min-height: calc(100dvh - 4rem);
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem 1rem 1rem;
  overflow-x: hidden;
  background: #f8fafc;
}

.workspace-header {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) auto;
  align-items: center;
  gap: 0.9rem;
  padding: 0.72rem 0.85rem;
  border: 1px solid #e5eaf0;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.workspace-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.65rem;
}

.back-button { flex: 0 0 auto; }

.page-title-wrapper {
  min-width: 0;
  display: grid;
  gap: 0.05rem;
}

.page-title-wrapper .eyebrow {
  color: #64748b;
  font-size: 0.56rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.page-title-wrapper h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.08rem;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.page-title-wrapper .muted {
  margin: 0;
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.67rem;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(82px, auto));
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #f8fafc;
}

.summary-item {
  display: grid;
  grid-template-columns: 18px auto;
  grid-template-rows: auto auto;
  column-gap: 0.32rem;
  min-width: 82px;
  padding: 0.42rem 0.58rem;
  border: 0;
  border-right: 1px solid #e5eaf0;
  background: transparent;
  text-align: left;
}

.summary-item:last-child { border-right: 0; }
.summary-item i { grid-row: 1 / 3; align-self: center; font-size: 0.72rem; }
.summary-item span { color: #94a3b8; font-size: 0.5rem; font-weight: 800; text-transform: uppercase; }
.summary-item strong { color: #0f172a; font-size: 0.8rem; }
.si-blue { color: #e63946; }
.si-emerald { color: #059669; }
.si-red { color: #dc2626; }
.si-amber { color: #d97706; }

.page-message { margin: 0; }

.filter-panel {
  padding: 0.65rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(135px, 1fr)) auto;
  gap: 0.55rem;
  align-items: end;
}

.filter-field {
  min-width: 0;
}

.filter-field label {
  display: block;
  margin-bottom: 0.24rem;
  color: #64748b;
  font-size: 0.55rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.filter-field :deep(.p-inputtext),
.filter-field :deep(.p-select) {
  width: 100%;
  min-width: 0;
  height: 38px;
  border-radius: 8px;
  font-size: 0.72rem;
}

.filter-field :deep(.p-select-label) {
  padding-block: 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.filter-action {
  display: flex;
  align-items: flex-end;
  justify-content: flex-end;
}

.date-input {
  box-sizing: border-box;
  width: 100%;
  height: 38px;
  padding: 0.45rem 0.65rem;
  border: 1px solid #dbe3ee;
  border-radius: 8px;
  outline: none;
  background: #fff;
  color: #0f172a;
  font-size: 0.72rem;
}

.date-input:focus {
  border-color: #e63946;
  box-shadow: 0 0 0 3px rgba(230, 57, 70, 0.08);
}

.table-panel {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.table-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 0.62rem 0.75rem;
  border-bottom: 1px solid #e5eaf0;
  background: #f8fafc;
}

.table-heading > div:first-child {
  display: grid;
  gap: 0.06rem;
}

.table-heading strong {
  color: #0f172a;
  font-size: 0.78rem;
}

.table-heading span {
  color: #94a3b8;
  font-size: 0.62rem;
}

.table-heading-meta {
  display: flex;
  gap: 0.35rem;
  align-items: center;
}

.table-scroll {
  width: 100%;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: thin;
}

.data-table {
  width: 100%;
  min-width: 900px;
  table-layout: fixed;
  border-collapse: collapse;
  font-size: 0.74rem;
}

.data-table thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  padding: 0.58rem 0.65rem;
  border-bottom: 1px solid #e5eaf0;
  background: #f8fafc;
  color: #64748b;
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-align: left;
  text-transform: uppercase;
  white-space: nowrap;
}

.data-table tbody td {
  height: 50px;
  padding: 0.5rem 0.65rem;
  overflow: hidden;
  border-bottom: 1px solid #edf1f6;
  color: #1e293b;
  text-overflow: ellipsis;
  vertical-align: middle;
}

.data-table tbody tr:last-child td { border-bottom: 0; }
.data-table tbody tr:hover { background: #fffbfb; }
.visit-row { cursor: default; }
.th-action { width: 110px; text-align: center !important; }

.prospect-cell {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.prospect-name {
  overflow: hidden;
  color: #d62839;
  font-size: 0.74rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.prospect-name:hover { text-decoration: underline; text-underline-offset: 2px; }

.prospect-meta,
.prospect-address,
.prospect-phone {
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.6rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prospect-address { max-width: 320px; }
.prospect-phone { font-family: Consolas, monospace; }

.datetime-cell {
  display: grid;
  gap: 0.02rem;
}

.date-val {
  color: #334155;
  font-size: 0.7rem;
  font-weight: 700;
}

.time-val {
  color: #94a3b8;
  font-family: Consolas, monospace;
  font-size: 0.62rem;
}

.visit-count-badge {
  display: inline-grid;
  min-width: 1.55rem;
  height: 1.55rem;
  padding: 0 0.35rem;
  place-content: center;
  border-radius: 999px;
  background: #fff5f5;
  color: #a51e2d;
  font-size: 0.66rem;
  font-weight: 800;
}

.cell-text {
  display: block;
  overflow: hidden;
  color: #475569;
  font-size: 0.7rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-action { text-align: center; }

.row-actions {
  display: inline-flex;
  gap: 0.12rem;
  align-items: center;
  padding: 0.12rem;
  border: 1px solid #e5eaf0;
  border-radius: 999px;
  background: #fff;
}

.row-actions :deep(.p-button) {
  width: 1.85rem;
  height: 1.85rem;
}

.act-detail { color: #0d9488 !important; }
.act-detail:hover { background: #f0fdfa !important; }
.act-prospect { color: #c54b59 !important; }
.act-prospect:hover { background: #f5f3ff !important; }
.act-view { color: #e63946 !important; }
.act-view:hover { background: #fff0f1 !important; }
.act-edit { color: #059669 !important; }
.act-edit:hover { background: #ecfdf5 !important; }
.act-map { color: #ea580c !important; }
.act-map:hover { background: #fff7ed !important; }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }

.state-box {
  min-height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 2rem;
  color: #64748b;
  text-align: center;
}

.state-icon { color: #e63946; font-size: 1.5rem; }

.state-icon-wrap {
  display: grid;
  width: 50px;
  height: 50px;
  place-content: center;
  border-radius: 13px;
  background: #f1f5f9;
  color: #94a3b8;
}

.state-box strong { color: #0f172a; font-size: 0.86rem; }
.state-box .muted { font-size: 0.7rem; }

.open-badge {
  display: inline-block;
  padding: 0.16rem 0.48rem;
  border-radius: 999px;
  background: #fffbeb;
  color: #92400e;
  font-size: 0.6rem;
  font-weight: 800;
}

.result-modal-body { display: grid; gap: 1rem; }
.result-customer { padding-bottom: 0.75rem; border-bottom: 1px solid #e5eaf0; }
.result-customer strong { display: block; font-size: 1rem; }
.result-customer span { color: #64748b; font-size: 0.78rem; }
.result-fields { display: grid; gap: 0.75rem; }
.result-field { display: grid; gap: 0.15rem; }
.result-label { color: #64748b; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; }
.result-value { color: #0f172a; font-size: 0.88rem; line-height: 1.5; white-space: pre-wrap; }
.result-empty { color: #94a3b8; font-style: italic; }

.detail-visits-list {
  display: grid;
  max-height: 62vh;
  gap: 0.7rem;
  padding-right: 0.2rem;
  overflow-y: auto;
}

.detail-visit-card {
  display: grid;
  gap: 0.6rem;
  padding: 0.85rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
}

.visit-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.visit-card-num {
  color: #0f172a;
  font-size: 0.72rem;
  font-weight: 800;
}

.visit-card-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
}

.visit-card-col {
  display: grid;
  gap: 0.38rem;
}

.detail-row {
  display: grid;
  gap: 0.08rem;
}

.detail-label {
  color: #64748b;
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.detail-row strong {
  color: #0f172a;
  font-size: 0.74rem;
}

.visit-card-actions {
  display: flex;
  gap: 0.2rem;
  align-items: center;
  padding-top: 0.45rem;
  border-top: 1px solid #edf1f6;
}

@media (max-width: 1200px) {
  .workspace-header {
    grid-template-columns: 1fr;
  }

  .summary-strip {
    width: 100%;
  }

  .filter-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .filter-action {
    grid-column: 1 / -1;
  }
}

@media (max-width: 768px) {
  .visit-page {
    min-height: auto;
    padding: 0.7rem;
    overflow: visible;
  }

  .page-title-wrapper .muted {
    white-space: normal;
  }

  .summary-strip {
    grid-template-columns: repeat(2, 1fr);
  }

  .summary-item:nth-child(2) { border-right: 0; }
  .summary-item:nth-child(-n + 2) { border-bottom: 1px solid #e5eaf0; }

  .filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filter-action {
    grid-column: 1 / -1;
  }

  .table-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .visit-card-body {
    grid-template-columns: 1fr;
  }

  .data-table {
    min-width: 850px;
  }
}

@media (max-width: 480px) {
  .visit-page {
    padding: 0.55rem;
  }

  .workspace-heading {
    align-items: flex-start;
  }

  .filter-grid {
    grid-template-columns: 1fr;
  }

  .filter-action {
    grid-column: auto;
  }
}
.filter-panel{padding:.55rem .65rem}.filter-grid{grid-template-columns:repeat(5,minmax(115px,1fr)) auto;gap:.45rem}.filter-field :deep(.p-select),.date-input{height:35px}.filter-field label{margin-bottom:.18rem}.filter-action :deep(.p-button){height:35px;white-space:nowrap}.summary-strip{gap:.4rem}.summary-item{padding:.38rem .55rem}.table-panel{width:100%;overflow:hidden}.table-scroll{width:100%;overflow-x:auto}.data-table{width:100%;min-width:1120px;table-layout:fixed;border-collapse:collapse}.data-table thead th{height:42px;padding:.55rem .7rem;background:#f4f6f9;color:#475569;font-size:.58rem;letter-spacing:.05em;text-transform:uppercase;white-space:nowrap}.data-table tbody td{height:60px;padding:.55rem .7rem;border-bottom:1px solid #e5eaf0;border-right:1px solid #e5eaf0;vertical-align:middle}.data-table tbody tr:hover{background:#fffafa}.data-table tbody tr{cursor:pointer}.cell-stack{display:grid;min-width:0;gap:.08rem}.cell-primary,.cell-sub{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.cell-primary{color:#0f172a;font-size:.72rem;font-weight:700}.cell-sub{color:#64748b;font-size:.62rem}.prospect-name{display:block;overflow:hidden;color:#e63946;font-size:.76rem;font-weight:750;text-overflow:ellipsis;white-space:nowrap}.prospect-meta,.prospect-address,.prospect-phone{display:block;overflow:hidden;color:#64748b;font-size:.62rem;text-overflow:ellipsis;white-space:nowrap}.action-dialog-subtitle{margin:0 0 1rem;color:#64748b;font-size:.78rem}.visit-action-list{display:grid;gap:.65rem}.visit-action-card{display:grid;grid-template-columns:38px 1fr 16px;align-items:center;gap:.65rem;width:100%;padding:.75rem;border:1px solid #dbe3ee;border-radius:12px;background:#fff;text-align:left;cursor:pointer}.visit-action-card:hover{border-color:#f3a4ad;background:#fff8f8}.visit-action-card>span:nth-child(2){display:grid;gap:.15rem}.visit-action-card strong{color:#0f172a;font-size:.78rem}.visit-action-card small{color:#64748b;font-size:.66rem}.visit-action-card>i{color:#94a3b8;font-size:.7rem}.visit-action-icon{display:grid;width:38px;height:38px;place-items:center;border-radius:50%;background:#eff6ff;color:#2563eb}.visit-detail-header{display:flex;align-items:center;gap:.7rem}.visit-detail-icon{display:grid;width:40px;height:40px;place-items:center;border-radius:12px;background:#fff0f1;color:#e63946}.visit-detail-header div{display:grid;gap:.12rem}.visit-detail-header span{color:#e63946;font-size:.56rem;font-weight:800;letter-spacing:.08em}.visit-detail-header strong{color:#0f172a;font-size:.95rem}.visit-detail-header small{color:#64748b;font-size:.68rem}.visit-detail-dialog :deep(.p-dialog-content){padding-top:.35rem}.detail-visits-list{display:grid;gap:.7rem}.detail-visit-card{overflow:hidden;border:1px solid #e2e8f0;border-radius:13px;background:#fff;box-shadow:0 3px 10px rgba(15,23,42,.04)}.visit-card-header{display:flex;align-items:center;justify-content:space-between;padding:.65rem .8rem;border-bottom:1px solid #edf1f6;background:#f8fafc}.visit-card-num{color:#0f172a;font-size:.72rem;font-weight:800}.open-badge{padding:.18rem .48rem;border-radius:999px;background:#dcfce7;color:#15803d;font-size:.6rem;font-weight:800}.visit-card-body{display:grid;grid-template-columns:1fr 1fr;gap:1rem;padding:.75rem .8rem}.detail-row{display:grid;gap:.18rem;padding:.28rem 0;border-bottom:1px solid #f1f5f9}.detail-label{color:#94a3b8;font-size:.6rem;font-weight:700;text-transform:uppercase;letter-spacing:.04em}.detail-row strong{color:#334155;font-size:.7rem;line-height:1.35}.visit-card-actions{display:flex;justify-content:flex-end;gap:.25rem;padding:.45rem .65rem;border-top:1px solid #edf1f6;background:#fbfdff}@media(max-width:1050px){.filter-grid{grid-template-columns:repeat(3,minmax(130px,1fr)) auto}}@media(max-width:700px){.filter-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.filter-action{grid-column:1/-1}.visit-card-body{grid-template-columns:1fr}}
@media (max-width: 640px) {
  .visit-page { padding:.5rem; gap:.5rem; background:#f8fafc; }
  .visit-page .workspace-header { padding:.85rem; border-radius:14px; }
  .visit-page .page-title-wrapper h1 { font-size:1.12rem; }
  .visit-page .page-title-wrapper .muted { font-size:.64rem; line-height:1.35; white-space:normal; }
  .visit-page .summary-strip { grid-template-columns:repeat(2,1fr); gap:.35rem; }
  .visit-page .summary-item { min-width:0; padding:.5rem .4rem; }
  .visit-page .summary-item span { font-size:.48rem; }
  .visit-page .page-heading-actions { display:grid; grid-template-columns:1fr 1fr; gap:.4rem; }
  .visit-page .page-heading-actions :deep(.p-button) { min-height:38px; font-size:.68rem; }
  .visit-page .filter-panel { padding:.7rem; border-radius:12px; }
  .visit-page .filter-grid { grid-template-columns:1fr; gap:.55rem; }
  .visit-page .filter-field :deep(.p-select), .visit-page .date-input { height:40px; border-radius:9px; }
  .visit-page .filter-action :deep(.p-button) { width:100%; min-height:36px; }
  .visit-page .table-panel { border-radius:12px; overflow:hidden; }
  .visit-page .table-scroll { overflow:hidden; }
  .visit-page .data-table { min-width:0; table-layout:fixed; font-size:.65rem; border-collapse:separate; border-spacing:0; }
  .visit-page .data-table thead th { height:34px; padding:.45rem .35rem; font-size:.54rem; background:#fff8f8; color:#8b4b55; }
  .visit-page .data-table tbody tr { background:#fff; }
  .visit-page .data-table tbody tr + tr td { border-top:1px solid #f4e6e8; }
  .visit-page .data-table tbody td { height:54px; padding:.45rem .35rem; vertical-align:middle; }
  .visit-page .data-table th, .visit-page .data-table td { padding:.6rem .45rem; }
  .visit-page .data-table thead th:nth-child(2), .visit-page .data-table tbody td:nth-child(2),
  .visit-page .data-table thead th:nth-child(3), .visit-page .data-table tbody td:nth-child(3),
  .visit-page .data-table thead th:nth-child(5), .visit-page .data-table tbody td:nth-child(5),
  .visit-page .data-table thead th:nth-child(7), .visit-page .data-table tbody td:nth-child(7) { display:none; }
  .visit-page .data-table thead th { font-size:0; }
  .visit-page .data-table thead th:nth-child(1)::after { content:'Customer'; font-size:.56rem; }
  .visit-page .data-table thead th:nth-child(4)::after { content:'Sales'; font-size:.56rem; }
  .visit-page .data-table thead th:nth-child(6)::after { content:'Status'; font-size:.56rem; }
  .visit-page .prospect-name { font-size:.7rem; }
  .visit-page .prospect-meta, .visit-page .prospect-address, .visit-page .prospect-phone { font-size:.55rem; }
  .visit-page .cell-text { font-size:.62rem; }
  .visit-page .p-tag { font-size:.54rem; padding:.18rem .3rem; white-space:normal; }
  .visit-page .state-box { min-height:180px; padding:1.2rem; font-size:.68rem; }
  .visit-action-card { min-height:54px; padding:.65rem; }
}
.visit-page { min-width:0; padding:0; gap:0; background:#fff; font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif; }
.visit-page .workspace-header,.visit-page .table-heading { display:none; }
.visit-erp-toolbar { display:flex; min-height:58px; height:58px; align-items:center; gap:10px; padding:10px 20px; border-bottom:1px solid #e2e8f0; background:#fff; }
.visit-section-title { display:inline-flex; height:34px; align-items:center; padding:0 14px; border-radius:10px; background:#dc2626; color:#fff; font-size:12px; font-weight:500; white-space:nowrap; }
.visit-search { display:flex; flex:0 0 270px; height:40px; align-items:center; gap:8px; padding:0 12px; border:1px solid #e2e8f0; border-radius:10px; color:#94a3b8; }.visit-search i{font-size:12px}.visit-search input{width:100%;height:32px;border:0;outline:0;background:transparent;color:#0f172a;font-size:13px}.visit-search input::placeholder{color:#94a3b8}
.visit-erp-toolbar>:deep(.p-button){height:40px;min-height:40px;padding:0 14px;border-radius:10px;font-size:12px;font-weight:600}
.visit-trash-button{display:flex;min-width:82px;height:36px;margin-left:auto;align-items:center;justify-content:center;gap:7px;padding:0 12px;border:1px solid #fecaca;border-radius:10px;background:#fff5f5;color:#b91c1b;font-family:Inter,ui-sans-serif,system-ui,sans-serif;font-size:12px;font-weight:600;line-height:16px;cursor:pointer}.visit-trash-button:hover{background:#fff1f2}.visit-trash-button i{font-size:15px}
.visit-page .panel-stack{gap:0;padding:0}.visit-page .filter-panel{display:flex;align-items:flex-end;gap:14px;padding:14px 20px 16px;border:0;border-bottom:1px solid #e2e8f0;border-radius:0;background:#fff;box-shadow:0 3px 12px rgba(15,23,42,.04)}.visit-page .filter-grid{display:grid;grid-template-columns:repeat(4,minmax(150px,210px)) auto;flex:1;gap:12px}.visit-page .filter-field{gap:5px}.visit-page .filter-field label{color:#64748b;font-size:10px;font-weight:600;letter-spacing:.04em;text-transform:uppercase}.visit-page .filter-field :deep(.p-select),.visit-page .date-input{height:40px;border:1px solid #e2e8f0;border-radius:10px;font-size:12px}.visit-page .filter-field :deep(.p-select-label){padding:0 12px}.visit-page .filter-action :deep(.p-button){height:40px;min-height:40px;border-radius:10px;font-size:12px}
.visit-page .table-panel{border:0;border-radius:0;box-shadow:none}.visit-page .data-table{min-width:0;width:100%;table-layout:fixed;font-size:11px}.visit-page .data-table thead th{box-sizing:border-box;height:42px;padding:0 10px;border-right:1px solid #e2e2e2;background:#f4f4f4;color:#000;font-size:10px;font-weight:600;line-height:12px;letter-spacing:.07em}.visit-page .data-table tbody td{box-sizing:border-box;height:67px;padding:8px 10px;border-right:1px solid #e2e2e2;border-bottom:1px solid #e0e0e0;font-size:11px;line-height:13px}.visit-page .data-table tbody tr:hover{background:#fff}.visit-page .data-table :deep(.p-tag){border-radius:999px;padding:4px 10px;font-size:10px;font-weight:500}.visit-page .pagination-bar{min-height:42px;padding:0 20px;border-top:0;border-bottom:1px solid #e2e8f0;background:#fff}
@media(max-width:900px){.visit-erp-toolbar{height:auto;min-height:58px;flex-wrap:wrap;overflow:visible}.visit-search{flex-basis:220px}.visit-erp-toolbar>:deep(.p-button){width:100%}.visit-page .filter-panel{display:grid;grid-template-columns:1fr;padding:12px}.visit-page .filter-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:640px){.visit-erp-toolbar{padding:10px 12px}.visit-section-title{width:100%;justify-content:center}.visit-page .data-table{min-width:900px}.visit-page .table-scroll{overflow-x:auto}.visit-page .filter-grid{grid-template-columns:1fr}}
.visit-page .visit-checkbox{display:flex;width:18px;height:18px;margin:0 auto;align-items:center;justify-content:center;padding:0;border:1px solid #cbd5e1;border-radius:4px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.08);cursor:pointer}
.visit-page .status-badge{display:inline-flex;width:fit-content;align-items:center;justify-content:center;padding:4px 10px;border:1px solid;border-radius:999px;font-family:Inter,ui-sans-serif,system-ui,sans-serif;font-size:11px;font-weight:600;line-height:16px;white-space:nowrap}.visit-page .status-badge-info{border-color:#bfdbfe;background:#dbeafe;color:#1d4ed8}.visit-page .status-badge-success{border-color:#bbf7d0;background:#dcfce7;color:#166534}.visit-page .status-badge-warning{border-color:#fed7aa;background:#fff7ed;color:#c2410c}.visit-page .status-badge-lost{border-color:#fecaca;background:#fff1f2;color:#b91c1c}
.visit-page .visit-checkbox-selected{border-color:#ef4444;background:#ef4444;color:#fff}.visit-trash-button:disabled{cursor:not-allowed;opacity:.55}.visit-trash-disabled{pointer-events:none}
.visit-page .data-table th:first-child,.visit-page .data-table td:first-child{width:54px !important;min-width:54px !important;padding:0 !important;text-align:center !important}
.visit-pagination-bar{display:flex;min-height:44px;align-items:center;justify-content:space-between;padding:5px 16px;border-bottom:1px solid #e2e8f0;background:#fff;font-family:Inter,ui-sans-serif,system-ui,sans-serif;font-size:12px;color:#64748b}.visit-pagination-info{display:flex;align-items:center;gap:12px}.visit-page-number{display:flex;width:30px;height:30px;align-items:center;justify-content:center;border-radius:999px;background:#fff1f2;color:#991b1b;font-weight:600}.visit-pagination-settings{display:flex;align-items:center;gap:10px}.visit-pagination-settings label{position:relative;display:flex;width:80px;height:34px;align-items:center;padding:0 10px;border:1px solid #e2e8f0;border-radius:8px;color:#334155}.visit-pagination-settings label span{position:absolute;top:-7px;left:8px;padding:0 4px;background:#fff;font-size:10px;font-weight:600;color:#475569}.visit-pagination-settings label strong{font-weight:400}.visit-pagination-settings button{height:34px;padding:0 12px;border:1px solid #e2e8f0;border-radius:8px;background:#fff;color:#475569}.visit-pagination-settings button:last-child{width:30px;padding:0;border:0;border-radius:999px}
.bulk-delete-warning{display:flex;align-items:flex-start;gap:12px;padding:14px;border:1px solid #fecaca;border-radius:12px;background:#fff7f7;color:#334155}.bulk-delete-warning>i{display:flex;width:32px;height:32px;flex:none;align-items:center;justify-content:center;border-radius:999px;background:#fee2e2;color:#dc2626;font-size:15px}.bulk-delete-warning strong{display:block;font-size:13px;font-weight:700;color:#0f172a}.bulk-delete-warning p{margin:4px 0 0;font-size:12px;line-height:18px;color:#64748b}.bulk-delete-dialog :deep(.p-dialog-footer){display:flex;justify-content:flex-end;gap:8px;padding-top:4px}.bulk-delete-dialog :deep(.p-dialog-footer .p-button){min-height:38px;border-radius:10px}
.visit-page .data-table { border-collapse: collapse; }
.visit-page .data-table thead th { box-sizing:border-box; height:42px; padding:0 10px; border-right:1px solid #e2e2e2; border-bottom:1px solid #e2e2e0; background:#f4f4f4; color:#000; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:11px; font-weight:600; line-height:12px; letter-spacing:.07em; text-align:left; text-transform:uppercase; vertical-align:middle; }
.visit-page .data-table tbody td { box-sizing:border-box; height:67px; padding:8px 10px; border-right:1px solid #e2e2e2; border-bottom:1px solid #e0e0e0; color:#0f172a; font-size:11px; line-height:13px; }

/* Keep the page chrome fixed and limit scrolling to the visit data area. */
.visit-page { display:flex; height:100%; min-height:0; flex-direction:column; gap:0; padding:0!important; overflow:hidden; }
.visit-page > .workspace-header,
.visit-page > .visit-erp-toolbar,
.visit-page > .filter-panel { flex:none; }
.visit-page > .visit-erp-toolbar { position:sticky; top:0; z-index:20; background:#fff; }
.visit-page > .table-panel { display:flex; min-height:0; flex:1; flex-direction:column; overflow:hidden; border-radius:0; }
.visit-page > .table-panel > .visit-pagination-bar,
.visit-page > .table-panel > .table-heading { flex:none; }
.visit-page > .table-panel > .table-scroll { min-height:0; flex:1; overflow:auto; }
.visit-page > .table-panel > .state-box { flex:1; min-height:0; overflow:auto; }
@media(max-width:900px){
  .visit-page { height:auto; min-height:100%; overflow:visible; }
  .visit-page > .table-panel { overflow:visible; }
  .visit-page > .table-panel > .table-scroll { max-height:calc(100dvh - 220px); overflow:auto; }
}
</style>
