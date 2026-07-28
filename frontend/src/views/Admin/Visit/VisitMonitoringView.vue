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
import type { VisitMonitoringItem, VisitMonitoringFilters, SalesExecutiveOption, ProspectStatus } from '../../../types/crm'
import VisitSelfieModal from './VisitSelfieModal.vue'

const router = useRouter()

const loading = ref(false)
const error = ref('')
const visits = ref<VisitMonitoringItem[]>([])
const salesExecutives = ref<SalesExecutiveOption[]>([])
const selfieModalItem = ref<VisitMonitoringItem | null>(null)
const detailModalItem = ref<VisitMonitoringItem | null>(null)
const showDetailModal = ref(false)
const deleteModalItem = ref<VisitMonitoringItem | null>(null)
const showDeleteModal = ref(false)
const deleteBusy = ref(false)

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

const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'New Lead', value: 'NEW_LEAD' },
  { label: 'Contacted', value: 'CONTACTED' },
  { label: 'Interested', value: 'INTERESTED' },
  { label: 'Qualified', value: 'QUALIFIED' },
  { label: 'Proposal Sent', value: 'PROPOSAL_SENT' },
  { label: 'Negotiation', value: 'NEGOTIATION' },
  { label: 'Won', value: 'WON' },
  { label: 'Lost', value: 'LOST' },
  { label: 'Converted', value: 'CONVERTED' },
]

const salesOptions = computed(() => {
  return [{ label: 'All Sales', value: '' }, ...salesExecutives.value.map((s) => ({ label: s.fullName, value: s.id }))]
})

const selectedSales = ref('')
const selectedRadius = ref('ALL')
const selectedStatus = ref('')

const totalVisits = computed(() => visits.value.length)
const insideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'INSIDE').length)
const outsideCount = computed(() => visits.value.filter((v) => v.radiusStatus === 'OUTSIDE').length)
const openVisits = computed(() => visits.value.filter((v) => !v.checkOutAt).length)

const customerSearch = ref('')
const filteredVisits = computed(() => {
  let result = visits.value
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
  selectedStatus.value = ''
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

function openSelfie(item: VisitMonitoringItem) {
  selfieModalItem.value = item
}

function openDetail(item: VisitMonitoringItem) {
  detailModalItem.value = item
  showDetailModal.value = true
}

function goToProspect(item: VisitMonitoringItem) {
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

function viewGpsLocation(item: VisitMonitoringItem) {
  window.open(`https://www.google.com/maps?q=${item.checkInLatitude},${item.checkInLongitude}`, '_blank')
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
          <span>Total Visits</span>
          <strong>{{ totalVisits }}</strong>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-emerald"><i class="pi pi-check-circle" /></div>
        <div class="summary-body">
          <span>Inside Radius</span>
          <strong>{{ insideCount }}</strong>
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
      <div v-else-if="!filteredVisits.length" class="state-box">
        <div class="state-icon-wrap"><i class="pi pi-map-marker" /></div>
        <strong>No visits found</strong>
        <span class="muted">No visit records match your current filters.</span>
      </div>
      <div v-else class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>Prospect</th>
              <th>Sales Executive</th>
              <th>Check In</th>
              <th>Check Out</th>
              <th>Duration</th>
              <th>Distance</th>
              <th>Radius</th>
              <th>Status</th>
              <th>Visits</th>
              <th class="th-action">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="visit in filteredVisits" :key="visit.id">
              <td>
                <div class="prospect-cell">
                  <strong class="prospect-name" @click="goToProspect(visit)">{{ visit.customerName }}</strong>
                  <small class="prospect-meta">{{ visit.industryGroup }} · {{ visit.customerCategory }}</small>
                  <small v-if="visit.formattedAddress" class="prospect-address">{{ visit.formattedAddress }}</small>
                  <small v-if="visit.phoneNumber" class="prospect-phone"><i class="pi pi-phone" /> {{ visit.phoneNumber }}</small>
                </div>
              </td>
              <td>
                <span class="cell-text">{{ visit.salesExecutiveName }}</span>
              </td>
              <td>
                <div class="datetime-cell">
                  <span class="date-val">{{ formatDateShort(visit.checkInAt) }}</span>
                  <span class="time-val">{{ formatTime(visit.checkInAt) }}</span>
                </div>
              </td>
              <td>
                <div v-if="visit.checkOutAt" class="datetime-cell">
                  <span class="date-val">{{ formatDateShort(visit.checkOutAt) }}</span>
                  <span class="time-val">{{ formatTime(visit.checkOutAt) }}</span>
                </div>
                <span v-else class="open-badge">Open</span>
              </td>
              <td>
                <span class="cell-text">{{ formatDuration(visit.durationSeconds) }}</span>
              </td>
              <td>
                <span class="cell-text">{{ formatDistance(visit.distanceMeters) }}</span>
              </td>
              <td>
                <Tag
                  :value="visit.radiusStatus === 'INSIDE' ? 'Inside' : visit.radiusStatus === 'OUTSIDE' ? 'Outside' : 'Unknown'"
                  :severity="visit.radiusStatus === 'INSIDE' ? 'success' : visit.radiusStatus === 'OUTSIDE' ? 'danger' : 'secondary'"
                  size="small"
                />
              </td>
              <td>
                <Tag :value="statusLabel(visit.prospectStatus)" :severity="statusSeverity(visit.prospectStatus)" size="small" />
              </td>
              <td>
                <span class="visit-count-badge">{{ visit.visitCount }}</span>
              </td>
              <td class="td-action">
                <div class="row-actions">
                  <Button v-tooltip.top="'Visit Details'" icon="pi pi-info-circle" text rounded size="small" class="act-detail" @click="openDetail(visit)" />
                  <Button v-tooltip.top="'View Prospect'" icon="pi pi-eye" text rounded size="small" class="act-prospect" @click="goToProspect(visit)" />
                  <Button v-tooltip.top="'Selfie Evidence'" icon="pi pi-camera" text rounded size="small" class="act-view" @click="openSelfie(visit)" />
                  <Button v-tooltip.top="'Download Evidence'" icon="pi pi-download" text rounded size="small" class="act-edit" @click="downloadVisitData(visit)" />
                  <Button v-tooltip.top="'View GPS'" icon="pi pi-map" text rounded size="small" class="act-map" @click="viewGpsLocation(visit)" />
                  <Button v-tooltip.top="'Delete Visit'" icon="pi pi-trash" text rounded size="small" class="act-delete" @click="confirmDelete(visit)" />
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

    <!-- Visit Detail Modal -->
    <Dialog v-model:visible="showDetailModal" modal header="Visit Details" :style="{ width: 'min(100%, 520px)' }" :closable="true">
      <template v-if="detailModalItem">
        <div class="detail-grid">
          <div class="detail-section">
            <h3 class="detail-heading">Prospect Information</h3>
            <div class="detail-row"><span class="detail-label">Name</span><strong>{{ detailModalItem.customerName }}</strong></div>
            <div class="detail-row"><span class="detail-label">Industry</span><strong>{{ detailModalItem.industryGroup || '—' }}</strong></div>
            <div class="detail-row"><span class="detail-label">Category</span><strong>{{ detailModalItem.customerCategory || '—' }}</strong></div>
            <div class="detail-row"><span class="detail-label">Address</span><strong>{{ detailModalItem.formattedAddress || '—' }}</strong></div>
            <div class="detail-row"><span class="detail-label">Phone</span><strong>{{ detailModalItem.phoneNumber || '—' }}</strong></div>
            <div class="detail-row"><span class="detail-label">Status</span>
              <Tag :value="statusLabel(detailModalItem.prospectStatus)" :severity="statusSeverity(detailModalItem.prospectStatus)" size="small" />
            </div>
          </div>

          <div class="detail-section">
            <h3 class="detail-heading">Visit Information</h3>
            <div class="detail-row"><span class="detail-label">Sales Executive</span><strong>{{ detailModalItem.salesExecutiveName }}</strong></div>
            <div class="detail-row"><span class="detail-label">Check In</span><strong>{{ formatDateShort(detailModalItem.checkInAt) }} {{ formatTime(detailModalItem.checkInAt) }}</strong></div>
            <div class="detail-row"><span class="detail-label">Check Out</span><strong v-if="detailModalItem.checkOutAt">{{ formatDateShort(detailModalItem.checkOutAt) }} {{ formatTime(detailModalItem.checkOutAt) }}</strong><span v-else class="open-badge">Open</span></div>
            <div class="detail-row"><span class="detail-label">Duration</span><strong>{{ formatDuration(detailModalItem.durationSeconds) }}</strong></div>
            <div class="detail-row"><span class="detail-label">Total Visits</span><strong>{{ detailModalItem.visitCount }}</strong></div>
          </div>

          <div class="detail-section">
            <h3 class="detail-heading">GPS & Location</h3>
            <div class="detail-row"><span class="detail-label">Check-in GPS</span><strong>{{ detailModalItem.checkInLatitude.toFixed(6) }}, {{ detailModalItem.checkInLongitude.toFixed(6) }}</strong></div>
            <div class="detail-row" v-if="detailModalItem.checkOutLatitude"><span class="detail-label">Check-out GPS</span><strong>{{ detailModalItem.checkOutLatitude?.toFixed(6) }}, {{ detailModalItem.checkOutLongitude?.toFixed(6) }}</strong></div>
            <div class="detail-row"><span class="detail-label">Distance from target</span><strong>{{ formatDistance(detailModalItem.distanceMeters) }}</strong></div>
            <div class="detail-row"><span class="detail-label">Radius Status</span>
              <Tag
                :value="detailModalItem.radiusStatus === 'INSIDE' ? 'Inside' : detailModalItem.radiusStatus === 'OUTSIDE' ? 'Outside' : 'Unknown'"
                :severity="detailModalItem.radiusStatus === 'INSIDE' ? 'success' : detailModalItem.radiusStatus === 'OUTSIDE' ? 'danger' : 'secondary'"
                size="small"
              />
            </div>
          </div>

          <div class="detail-section" v-if="detailModalItem.visitNotes || detailModalItem.followUpNotes">
            <h3 class="detail-heading">Notes</h3>
            <div v-if="detailModalItem.visitNotes" class="detail-row"><span class="detail-label">Visit Notes</span><strong>{{ detailModalItem.visitNotes }}</strong></div>
            <div v-if="detailModalItem.followUpNotes" class="detail-row"><span class="detail-label">Follow-up Notes</span><strong>{{ detailModalItem.followUpNotes }}</strong></div>
          </div>
        </div>
      </template>
      <template #footer>
        <Button label="Close" severity="secondary" outlined @click="showDetailModal = false" />
        <Button label="View Prospect" icon="pi pi-eye" @click="detailModalItem && goToProspect(detailModalItem)" />
        <Button label="View on Map" icon="pi pi-map" @click="detailModalItem && viewGpsLocation(detailModalItem)" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; }

.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand-green-light, #0b7766); margin-bottom: 0.35rem; }
.page-title-wrapper h1 { font-size: 1.65rem; font-weight: 800; color: var(--text-primary); margin: 0 0 0.2rem; letter-spacing: -0.03em; line-height: 1.15; }
.page-title-wrapper .muted { font-size: 0.85rem; color: var(--text-muted); max-width: 520px; line-height: 1.55; }

.summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.85rem; }
.summary-card {
  display: flex; align-items: center; gap: 0.85rem; padding: 1rem 1.15rem;
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-lg); box-shadow: var(--shadow-xs);
}
.summary-icon {
  width: 2.4rem; height: 2.4rem; border-radius: var(--radius-sm);
  display: grid; place-items: center; font-size: 0.9rem; flex-shrink: 0;
}
.si-blue { background: #eff6ff; color: #2563eb; }
.si-emerald { background: #ecfdf5; color: #059669; }
.si-red { background: #fef2f2; color: #dc2626; }
.si-amber { background: #fffbeb; color: #d97706; }
.summary-body { display: grid; }
.summary-body span { font-size: 0.62rem; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.04em; }
.summary-body strong { font-size: 1.35rem; font-weight: 800; letter-spacing: -0.03em; }

.filter-panel { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 1.1rem 1.25rem; box-shadow: var(--shadow-xs); }
.filter-grid { display: grid; grid-template-columns: repeat(5, 1fr) auto; gap: 0.75rem; align-items: end; }
.filter-field label { display: block; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.35rem; }
.filter-field :deep(.p-inputtext),
.filter-field :deep(.p-select) { width: 100%; }
.filter-action { display: flex; align-items: end; padding-bottom: 2px; }

.date-input {
  width: 100%; padding: 0.5rem 0.75rem; font-size: 0.82rem;
  border: 1px solid var(--border-light, #d1d5db); border-radius: 8px;
  background: var(--surface-card, #fff); color: var(--text-primary, #111827);
  outline: none; transition: border-color 0.15s;
}
.date-input:focus { border-color: #2563eb; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }

.table-panel { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 1.1rem 1.25rem; box-shadow: var(--shadow-xs); }

.state-box { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.6rem; padding: 3rem 1rem; color: var(--text-muted); }
.state-icon { font-size: 2rem; }
.state-icon-wrap { width: 3rem; height: 3rem; display: grid; place-items: center; background: var(--brand-blue-bg); color: var(--brand-blue); border-radius: 50%; font-size: 1rem; }
.state-box strong { font-size: 0.9rem; color: var(--text-primary); }
.state-box .muted { font-size: 0.78rem; }

.table-scroll { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
.data-table thead th { background: var(--surface-subtle, #f8fafc); color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; padding: 0.75rem 0.95rem; border-bottom: 1px solid var(--border-light); white-space: nowrap; text-align: left; }
.data-table tbody td { padding: 0.75rem 0.95rem; border-bottom: 1px solid #f0f3f7; color: var(--text-primary); vertical-align: middle; }
.data-table tbody tr:hover { background: #f8fafc; }
.th-action { text-align: center !important; width: 170px; }

.prospect-cell { display: grid; gap: 0.1rem; }
.prospect-name { font-size: 0.82rem; font-weight: 600; cursor: pointer; color: #2563eb; }
.prospect-name:hover { text-decoration: underline; }
.prospect-meta { font-size: 0.68rem; color: var(--text-muted); }
.prospect-address { font-size: 0.68rem; color: var(--text-muted); max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prospect-phone { font-size: 0.68rem; color: var(--text-muted); font-family: 'SF Mono', 'Fira Code', monospace; }
.prospect-phone i { font-size: 0.6rem; margin-right: 0.2rem; }

.datetime-cell { display: grid; gap: 0.05rem; }
.date-val { font-size: 0.78rem; color: var(--text-primary); }
.time-val { font-size: 0.72rem; color: var(--text-muted); font-family: 'SF Mono', 'Fira Code', monospace; }

.open-badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 20px; background: #fffbeb; color: #92400e; font-size: 0.68rem; font-weight: 600; }

.visit-count-badge {
  display: inline-flex; align-items: center; justify-content: center;
  min-width: 1.5rem; height: 1.5rem; padding: 0 0.35rem; border-radius: 50%;
  background: #f1f5f9; color: #475569; font-size: 0.72rem; font-weight: 700;
}

.cell-text { font-size: 0.82rem; color: var(--text-primary); }

.td-action { text-align: center; }
.row-actions { display: flex; align-items: center; justify-content: center; gap: 0.15rem; }
.act-prospect { color: #7c3aed !important; }
.act-prospect:hover { background: #f5f3ff !important; }
.act-view { color: #2563eb !important; }
.act-view:hover { background: #eff6ff !important; }
.act-edit { color: #059669 !important; }
.act-edit:hover { background: #ecfdf5 !important; }
.act-map { color: #ea580c !important; }
.act-map:hover { background: #fff7ed !important; }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }
.act-detail { color: #0d9488 !important; }
.act-detail:hover { background: #f0fdfa !important; }

.detail-grid { display: grid; gap: 1rem; }
.detail-section { display: grid; gap: 0.45rem; }
.detail-heading { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); padding-bottom: 0.35rem; border-bottom: 1px solid var(--border-light); }
.detail-row { display: flex; flex-direction: column; gap: 0.1rem; }
.detail-label { font-size: 0.68rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text-muted); }
.detail-row strong { font-size: 0.82rem; color: var(--text-primary); }

@media (max-width: 1200px) { .summary-grid { grid-template-columns: repeat(2, 1fr); } .filter-grid { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 768px) { .admin-page { padding: 1.25rem 1rem; } .filter-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 480px) { .summary-grid { grid-template-columns: 1fr; } .filter-grid { grid-template-columns: 1fr; } }
</style>
