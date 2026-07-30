<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Tooltip from 'primevue/tooltip'
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

const totalVisits = computed(() => visits.value.length)
const insideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'INSIDE').length)
const outsideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'OUTSIDE').length)
const openVisits = computed(() => visits.value.filter((v) => !v.checkOutAt).length)
const totalProspects = computed(() => groupedVisits.value.length)

const customerSearch = ref('')
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
  <section class="admin-page">
    <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.back()" title="Back" />
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Field Operations</span>
        <h1>Visit Monitoring</h1>
        <p class="muted">Track GPS check-ins, selfie evidence, and visit outcomes across your sales team.</p>
      </div>
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-icon si-blue"><i class="pi pi-map-marker" /></div>
        <div class="summary-body">
          <span>Total Prospects</span>
          <strong>{{ totalProspects }}</strong>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-emerald"><i class="pi pi-check-circle" /></div>
        <div class="summary-body">
          <span>Total Visits</span>
          <strong>{{ totalVisits }}</strong>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-red"><i class="pi pi-exclamation-circle" /></div>
        <div class="summary-body">
          <span>Outside Radius</span>
          <strong>{{ outsideCount }}</strong>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-amber"><i class="pi pi-clock" /></div>
        <div class="summary-body">
          <span>Open Visits</span>
          <strong>{{ openVisits }}</strong>
        </div>
      </div>
    </div>

    <div class="filter-panel">
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
          <label>Search</label>
          <InputText v-model="customerSearch" placeholder="Search customer / industry..." />
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

    <div class="table-panel">
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
              <th>Prospect</th>
              <th>Sales Executive</th>
              <th>Last Visit</th>
              <th>Status</th>
              <th>Visits</th>
              <th class="th-action">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in filteredGroupedVisits" :key="row.prospectId">
              <td>
                <div class="prospect-cell">
                  <strong class="prospect-name" @click="goToProspect(row)">{{ row.customerName }}</strong>
                  <small class="prospect-meta">{{ row.industryGroup }} · {{ row.customerCategory }}</small>
                  <small v-if="row.formattedAddress" class="prospect-address">{{ row.formattedAddress }}</small>
                  <small v-if="row.phoneNumber" class="prospect-phone"><i class="pi pi-phone" /> {{ row.phoneNumber }}</small>
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
                <Tag :value="statusLabel(row.prospectStatus)" :severity="statusSeverity(row.prospectStatus)" size="small" />
              </td>
              <td>
                <span class="visit-count-badge">{{ row.visitCount }}</span>
              </td>
              <td class="td-action">
                <div class="row-actions">
                  <Button v-tooltip.top="'View Details'" icon="pi pi-info-circle" text rounded size="small" class="act-detail" @click="openDetail(row)" />
                  <Button v-tooltip.top="'View Prospect'" icon="pi pi-eye" text rounded size="small" class="act-prospect" @click="goToProspect(row)" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <VisitSelfieModal :item="selfieModalItem" @close="selfieModalItem = null" />

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
    <Dialog v-model:visible="showDetailModal" modal :header="'Visit Details - ' + detailProspectName" :style="{ width: 'min(100%, 700px)' }" :closable="true">
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
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; }

.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; margin: 0; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand-green-light); margin-bottom: 0.35rem; }
.page-title-wrapper h1 { font-size: 1.65rem; font-weight: 800; color: var(--text-primary); margin: 0 0 0.2rem; letter-spacing: -0.03em; line-height: 1.15; }
.page-title-wrapper .muted { font-size: 0.85rem; color: var(--text-muted); max-width: 520px; line-height: 1.55; }

.summary-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; }
.summary-card {
  display: flex; align-items: center; gap: 0.9rem; padding: 1.1rem 1.2rem;
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-lg); box-shadow: var(--shadow-xs);
  transition: box-shadow var(--transition-smooth), transform var(--transition-smooth), border-color var(--transition-smooth);
}
.summary-card:hover { box-shadow: var(--shadow-md); transform: translateY(-2px); border-color: var(--border-strong); }
.summary-icon { width: 44px; height: 44px; border-radius: var(--radius-md); display: grid; place-content: center; font-size: 1.1rem; flex-shrink: 0; }
.si-blue { background: #eff6ff; color: #2563eb; }
.si-emerald { background: #ecfdf5; color: #059669; }
.si-red { background: #fef2f2; color: #dc2626; }
.si-amber { background: #fffbeb; color: #d97706; }
.summary-body { display: flex; flex-direction: column; gap: 0.15rem; }
.summary-body span { font-size: 0.62rem; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.04em; }
.summary-body strong { font-size: 1.35rem; font-weight: 800; letter-spacing: -0.03em; line-height: 1.2; }

.filter-panel {
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-lg); padding: 1.15rem 1.3rem; box-shadow: var(--shadow-xs);
}
.filter-grid { display: grid; grid-template-columns: repeat(5, 1fr) auto; gap: 0.75rem; align-items: end; }
.filter-field label { display: block; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.35rem; }
.filter-field :deep(.p-inputtext),
.filter-field :deep(.p-select) { width: 100%; }
.filter-action { display: flex; align-items: end; padding-bottom: 2px; }

.date-input {
  width: 100%; padding: 0.5rem 0.75rem; font-size: 0.82rem;
  border: 1px solid var(--border-light); border-radius: var(--radius-sm);
  background: var(--surface-card); color: var(--text-primary);
  outline: none; transition: border-color 0.15s, box-shadow 0.15s;
}
.date-input:focus { border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }

.table-panel {
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-lg); box-shadow: var(--shadow-xs); overflow: hidden;
}

.state-box { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.5rem; padding: 3rem 1rem; color: var(--text-muted); }
.state-icon { font-size: 2rem; }
.state-icon-wrap { width: 56px; height: 56px; border-radius: var(--radius-lg); background: var(--surface-subtle); display: grid; place-content: center; }
.state-icon-wrap i { font-size: 1.4rem; color: var(--text-faint); }
.state-box strong { font-size: 0.95rem; color: var(--text-primary); }
.state-box .muted { font-size: 0.78rem; }

.table-scroll { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
.data-table thead th {
  background: #f1f5f9; color: var(--text-muted); font-size: 0.68rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.06em;
  padding: 0.85rem 0.95rem; border-bottom: 1px solid var(--border-light);
  white-space: nowrap; text-align: left;
}
.data-table tbody td { padding: 0.85rem 0.95rem; border-bottom: 1px solid #f0f3f7; color: var(--text-primary); vertical-align: middle; }
.data-table tbody tr:last-child td { border-bottom: none; }
.data-table tbody tr { transition: background var(--transition-fast); }
.data-table tbody tr:hover { background: #f8fafc; }
.th-action { text-align: center !important; width: 150px; }

.prospect-cell { display: grid; gap: 0.15rem; }
.prospect-name { font-size: 0.82rem; font-weight: 600; cursor: pointer; color: var(--brand-blue); transition: color var(--transition-fast); }
.prospect-name:hover { color: #1d4ed8; text-decoration: underline; }
.prospect-meta { font-size: 0.68rem; color: var(--text-muted); }
.prospect-address { font-size: 0.68rem; color: var(--text-muted); max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prospect-phone { font-size: 0.68rem; color: var(--text-muted); font-family: 'SF Mono', 'Fira Code', monospace; }
.prospect-phone i { font-size: 0.6rem; margin-right: 0.2rem; }

.datetime-cell { display: grid; gap: 0.05rem; }
.date-val { font-size: 0.78rem; color: var(--text-primary); font-weight: 500; }
.time-val { font-size: 0.72rem; color: var(--text-muted); font-family: 'SF Mono', 'Fira Code', monospace; }

.open-badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: 20px; background: #fffbeb; color: #92400e; font-size: 0.65rem; font-weight: 700; letter-spacing: 0.02em; }

.visit-count-badge {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 1.6rem; height: 1.6rem; padding: 0 0.4rem; border-radius: 50%;
  background: #e0e7ff; color: #4338ca; font-size: 0.72rem; font-weight: 700;
}

.cell-text { font-size: 0.82rem; color: var(--text-secondary); }

.td-action { text-align: center; }
.row-actions { display: flex; align-items: center; justify-content: center; gap: 0.2rem; }
.act-prospect { color: #7c3aed !important; }
.act-prospect:hover { background: #f5f3ff !important; }
.act-view { color: var(--brand-blue) !important; }
.act-view:hover { background: #eff6ff !important; }
.act-edit { color: #059669 !important; }
.act-edit:hover { background: #ecfdf5 !important; border-color: #a7f3d0; }
.act-map { color: #ea580c !important; }
.act-map:hover { background: #fff7ed !important; }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }
.act-result { color: #059669 !important; }
.act-result:hover { background: #ecfdf5 !important; }

.result-modal-body { display: grid; gap: 1rem; }
.result-customer { padding-bottom: 0.75rem; border-bottom: 1px solid var(--border-light); }
.result-customer strong { display: block; font-size: 1rem; }
.result-customer span { color: var(--text-muted); font-size: 0.78rem; }
.result-fields { display: grid; gap: 0.75rem; }
.result-field { display: grid; gap: 0.15rem; }
.result-label { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.result-value { font-size: 0.88rem; color: var(--text-primary); line-height: 1.5; white-space: pre-wrap; }
.result-empty { color: var(--text-muted); font-style: italic; }
.act-detail { color: #0d9488 !important; }
.act-detail:hover { background: #f0fdfa !important; }

.detail-row { display: flex; flex-direction: column; gap: 0.1rem; }
.detail-label { font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text-muted); }
.detail-row strong { font-size: 0.82rem; color: var(--text-primary); }

.detail-visits-list { display: grid; gap: 0.85rem; max-height: 60vh; overflow-y: auto; padding-right: 0.25rem; }
.detail-visit-card {
  display: grid; gap: 0.65rem; padding: 1rem 1.15rem;
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-md); box-shadow: var(--shadow-xs);
  transition: box-shadow var(--transition-fast), border-color var(--transition-fast);
}
.detail-visit-card:hover { box-shadow: var(--shadow-sm); border-color: var(--border-strong); }
.visit-card-header { display: flex; align-items: center; gap: 0.5rem; }
.visit-card-num { font-size: 0.75rem; font-weight: 700; color: var(--text-primary); letter-spacing: -0.01em; }
.visit-card-body { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.visit-card-col { display: grid; gap: 0.4rem; }
.visit-card-actions {
  display: flex; align-items: center; gap: 0.25rem;
  padding-top: 0.5rem; border-top: 1px solid var(--border-light);
}

@media (max-width: 1200px) { .summary-grid { grid-template-columns: repeat(2, 1fr); } .filter-grid { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 768px) { .admin-page { padding: 1.25rem 1rem; } .filter-grid { grid-template-columns: repeat(2, 1fr); } .visit-card-body { grid-template-columns: 1fr; } }
@media (max-width: 480px) { .summary-grid { grid-template-columns: 1fr; } .filter-grid { grid-template-columns: 1fr; } }
</style>
