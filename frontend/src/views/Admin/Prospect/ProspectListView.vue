<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { getPipeline, getSalesExecutives, deleteProspect } from '../../../api/crm'
import { BOARD_STATUSES, filterProspects } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus, SalesExecutiveOption } from '../../../types/crm'

const crm = useCrmStore()
const router = useRouter()
const prospects = ref<Prospect[]>([])
const sales = ref<SalesExecutiveOption[]>([])
const error = ref('')
const loading = ref(true)

const searchQuery = ref('')
const salesFilter = ref('')
const industryFilter = ref('')
const statusFilter = ref('')

const deleteDialogVisible = ref(false)
const deleteTargetId = ref('')
const deleteTargetName = ref('')
const deleting = ref(false)

const industries = ['N&B / Kuliner', 'Retail', 'Hospitality', 'Health & Beauty', 'Services', 'Other']
const industryOptions = computed(() => [{ label: 'All Business Segments', value: '' }, ...industries.map((v) => ({ label: v, value: v }))])
const statusOptions = computed(() => [{ label: 'All Pipeline Statuses', value: '' }, ...BOARD_STATUSES.map((v) => ({ label: v.replaceAll('_', ' '), value: v }))])
const salesOptions = computed(() => [{ label: 'All Sales Executives', value: '' }, ...sales.value.map((s) => ({ label: s.fullName, value: s.id }))])

const filtered = computed(() => {
  return filterProspects(prospects.value, {
    salesExecutiveId: salesFilter.value,
    industryGroup: industryFilter.value,
    status: statusFilter.value,
    search: searchQuery.value,
  })
})

const totalActive = computed(() => prospects.value.filter((p) => !['CONVERTED', 'LOST'].includes(p.status)).length)
const totalWon = computed(() => prospects.value.filter((p) => p.status === 'WON').length)
const totalLost = computed(() => prospects.value.filter((p) => p.status === 'LOST').length)

function statusSeverity(status: ProspectStatus) {
  switch (status) {
    case 'NEW_LEAD': return 'info'
    case 'CONTACTED': return 'info'
    case 'INTERESTED': return 'success'
    case 'QUALIFIED': return 'success'
    case 'PROPOSAL_SENT': return 'warn'
    case 'NEGOTIATION': return 'warn'
    case 'WON': return 'success'
    case 'LOST': return 'danger'
    case 'CONVERTED': return 'secondary'
    default: return 'secondary'
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function resetFilters() {
  searchQuery.value = ''
  salesFilter.value = ''
  industryFilter.value = ''
  statusFilter.value = ''
}

function confirmDelete(id: string, name: string) {
  deleteTargetId.value = id
  deleteTargetName.value = name
  deleteDialogVisible.value = true
}

async function executeDelete() {
  deleting.value = true
  try {
    await deleteProspect(deleteTargetId.value)
    prospects.value = prospects.value.filter((p) => p.id !== deleteTargetId.value)
    deleteDialogVisible.value = false
  } catch (e) {
    error.value = crm.errorMessage(e)
  } finally {
    deleting.value = false
  }
}

function detailRoute(p: Prospect) {
  if (p.status === 'WON') return `/admin/prospects/${p.id}/review`
  return `/admin/prospects/${p.id}/review`
}

onMounted(async () => {
  try {
    const [salesResult] = await Promise.all([getSalesExecutives(), crm.loadPipeline()])
    sales.value = salesResult
    prospects.value = crm.pipeline
  } catch (e) {
    error.value = crm.errorMessage(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="admin-page">
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Prospect Management</span>
        <h1>Prospect List</h1>
        <p class="muted">All prospects across the sales pipeline. Manage, review, and track prospect lifecycle.</p>
      </div>
      <div class="page-heading-actions">
        <Button label="Prospect Finder" icon="pi pi-search" severity="success" outlined size="small" @click="router.push('/admin/prospect-finder')" />
        <Button label="View Pipeline" icon="pi pi-columns" severity="secondary" outlined size="small" @click="router.push('/admin/prospects/pipeline')" />
      </div>
    </header>

    <div class="summary-grid">
      <div class="summary-card">
        <div class="summary-icon si-blue"><i class="pi pi-inbox" /></div>
        <div class="summary-data">
          <span class="summary-label">Total Prospects</span>
          <span class="summary-value">{{ prospects.length }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-violet"><i class="pi pi-spin pi-spinner" v-if="false" /><i class="pi pi-chart-line" /></div>
        <div class="summary-data">
          <span class="summary-label">Active Pipeline</span>
          <span class="summary-value">{{ totalActive }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-emerald"><i class="pi pi-check-circle" /></div>
        <div class="summary-data">
          <span class="summary-label">Won</span>
          <span class="summary-value">{{ totalWon }}</span>
        </div>
      </div>
      <div class="summary-card">
        <div class="summary-icon si-amber"><i class="pi pi-times-circle" /></div>
        <div class="summary-data">
          <span class="summary-label">Lost</span>
          <span class="summary-value">{{ totalLost }}</span>
        </div>
      </div>
    </div>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div class="panel-stack">
      <div class="filter-panel">
        <div class="search-row">
          <div class="search-field">
            <i class="pi pi-search" />
            <input type="text" v-model="searchQuery" placeholder="Search by place name, address..." />
          </div>
        </div>
        <div class="filter-grid">
          <div class="filter-field">
            <label>Sales Executive</label>
            <Select v-model="salesFilter" :options="salesOptions" optionLabel="label" optionValue="value" placeholder="All Sales" />
          </div>
          <div class="filter-field">
            <label>Industry Group</label>
            <Select v-model="industryFilter" :options="industryOptions" optionLabel="label" optionValue="value" />
          </div>
          <div class="filter-field">
            <label>Status</label>
            <Select v-model="statusFilter" :options="statusOptions" optionLabel="label" optionValue="value" />
          </div>
          <div class="filter-field filter-action">
            <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetFilters" />
          </div>
        </div>
      </div>

      <div class="table-panel">
        <div v-if="loading" class="state-box">
          <i class="pi pi-spin pi-spinner state-icon" />
          <span>Loading prospects...</span>
        </div>
        <div v-else-if="!filtered.length" class="state-box">
          <div class="state-icon-wrap"><i class="pi pi-inbox" /></div>
          <strong>No prospects found</strong>
          <span class="muted">Adjust your search or filters to view results.</span>
        </div>
        <div v-else class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th>Place Name</th>
                <th>Category</th>
                <th>Industry Group</th>
                <th>Sales Executive</th>
                <th>Status</th>
                <th>Created</th>
                <th class="th-action">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in filtered" :key="p.id">
                <td>
                  <div class="cell-stack">
                    <button class="link-btn" @click="router.push(detailRoute(p))">{{ p.placeName }}</button>
                    <span class="cell-sub">{{ p.formattedAddress }}</span>
                  </div>
                </td>
                <td><span class="cell-text">{{ p.placeCategory }}</span></td>
                <td><span class="cell-text">{{ p.industryGroup }}</span></td>
                <td><span class="cell-text">{{ p.assignedSalesExecutive }}</span></td>
                <td><Tag :value="p.status.replaceAll('_', ' ')" :severity="statusSeverity(p.status)" /></td>
                <td><span class="cell-date">{{ formatDate(p.createdAt) }}</span></td>
                <td class="td-action">
                  <div class="row-actions">
                    <Button icon="pi pi-eye" text rounded size="small" class="act-view" title="Detail" @click="router.push(detailRoute(p))" />
                    <Button icon="pi pi-trash" text rounded size="small" class="act-delete" title="Delete" @click="confirmDelete(p.id, p.placeName)" />
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <Dialog v-model:visible="deleteDialogVisible" header="Delete Prospect" modal :style="{ width: '400px' }">
      <p>Are you sure you want to delete <strong>{{ deleteTargetName }}</strong>? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper { display: flex; flex-direction: column; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand-green-light, #0b7766); margin-bottom: 0.35rem; }
.page-title-wrapper h1 { font-size: 1.65rem; font-weight: 800; color: var(--text-primary); margin: 0 0 0.2rem; letter-spacing: -0.03em; line-height: 1.15; }
.page-title-wrapper .muted { font-size: 0.85rem; color: var(--text-muted); max-width: 520px; line-height: 1.55; }
.page-heading-actions { display: flex; gap: 0.5rem; align-items: center; padding-top: 0.15rem; }

.summary-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; }
.summary-card { display: flex; align-items: center; gap: 0.9rem; padding: 1.1rem 1.2rem; background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); box-shadow: var(--shadow-xs); transition: box-shadow var(--transition-fast), transform var(--transition-fast); }
.summary-card:hover { box-shadow: var(--shadow-sm); transform: translateY(-1px); }
.summary-icon { width: 44px; height: 44px; border-radius: var(--radius-md); display: grid; place-content: center; font-size: 1.1rem; flex-shrink: 0; }
.si-blue { background: #eff6ff; color: #2563eb; }
.si-violet { background: #eef2ff; color: #6366f1; }
.si-emerald { background: #ecfdf5; color: #059669; }
.si-amber { background: #fffbeb; color: #d97706; }
.summary-data { display: flex; flex-direction: column; }
.summary-label { font-size: 0.7rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text-muted); line-height: 1.3; }
.summary-value { font-size: 1.45rem; font-weight: 700; color: var(--text-primary); line-height: 1.25; }

.panel-stack { display: flex; flex-direction: column; gap: 1rem; }

.filter-panel { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 1.1rem 1.25rem; box-shadow: var(--shadow-xs); }
.search-row { margin-bottom: 0.9rem; }
.search-field { display: flex; align-items: center; gap: 0.65rem; padding: 0.6rem 0.9rem; background: var(--surface-subtle); border: 1px solid var(--border-default); border-radius: var(--radius-sm); transition: border-color var(--transition-fast), box-shadow var(--transition-fast), background var(--transition-fast); }
.search-field:focus-within { background: var(--surface-card); border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.search-field i { color: var(--text-faint); font-size: 0.9rem; }
.search-field input { flex: 1; border: none; outline: none; background: transparent; font-size: 0.87rem; color: var(--text-primary); }
.search-field input::placeholder { color: var(--text-faint); }
.filter-grid { display: grid; grid-template-columns: repeat(3, 1fr) auto; gap: 0.75rem; align-items: end; }
.filter-field { display: flex; flex-direction: column; gap: 0.3rem; }
.filter-field label { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.filter-action { justify-content: flex-end; padding-bottom: 0.15rem; }

.table-panel { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); box-shadow: var(--shadow-xs); overflow: hidden; }
.table-scroll { overflow-x: auto; }

.data-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
.data-table thead th { position: sticky; top: 0; z-index: 1; background: var(--surface-subtle); color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; padding: 0.75rem 0.95rem; border-bottom: 1px solid var(--border-light); white-space: nowrap; text-align: left; }
.data-table tbody td { padding: 0.75rem 0.95rem; border-bottom: 1px solid #f0f3f7; color: var(--text-primary); vertical-align: middle; }
.data-table tbody tr:last-child td { border-bottom: none; }
.data-table tbody tr { transition: background var(--transition-fast); }
.data-table tbody tr:hover { background: #f8fafc; }
.th-action { width: 100px; text-align: center; }

.cell-stack { display: flex; flex-direction: column; }
.cell-sub { font-size: 0.72rem; color: var(--text-muted); margin-top: 0.1rem; max-width: 320px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cell-text { font-size: 0.84rem; color: var(--text-secondary); }
.cell-date { font-size: 0.8rem; color: var(--text-muted); white-space: nowrap; }

.link-btn { background: none; border: none; padding: 0; cursor: pointer; text-align: left; font: inherit; font-weight: 600; color: #2563eb; transition: color var(--transition-fast); }
.link-btn:hover { color: #1d4ed8; text-decoration: underline; }

.td-action { text-align: center; }
.row-actions { display: flex; align-items: center; justify-content: center; gap: 0.15rem; }
.act-view { color: #2563eb !important; }
.act-view:hover { background: #eff6ff !important; }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }

.state-box { min-height: 260px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.4rem; padding: 2rem; text-align: center; color: var(--text-muted); }
.state-icon { font-size: 1.75rem; color: var(--brand-blue); margin-bottom: 0.25rem; }
.state-icon-wrap { width: 56px; height: 56px; border-radius: var(--radius-lg); background: var(--surface-subtle); display: grid; place-content: center; margin-bottom: 0.35rem; }
.state-icon-wrap i { font-size: 1.4rem; color: var(--text-faint); }
.state-box strong { color: var(--text-primary); font-size: 0.95rem; }

@media (max-width: 1200px) { .summary-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; gap: 1rem; }
  .page-heading { flex-direction: column; align-items: stretch; }
  .summary-grid { grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
  .filter-grid { grid-template-columns: 1fr 1fr; }
}
</style>
