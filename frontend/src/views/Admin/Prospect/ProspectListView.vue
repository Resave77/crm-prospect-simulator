<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { getPipeline, getSalesExecutives, deleteProspect, approveProspectDeletion, rejectProspectDeletion } from '../../../api/crm'
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

const deletionTarget = ref<Prospect | null>(null)
const showDeletionDialog = ref(false)
const deletionBusy = ref(false)

const actionDialogVisible = ref(false)
const actionTarget = ref<Prospect | null>(null)

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

function rowClass(status: ProspectStatus) {
  return {
    'prospect-row-won': status === 'WON',
    'prospect-row-lost': status === 'LOST',
    'prospect-row-converted': status === 'CONVERTED',
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

function viewProspect(id: string) {
  actionDialogVisible.value = false
  router.push({ name: 'AdminProspectReview', params: { id } })
}

function openActions(prospect: Prospect) {
  actionTarget.value = prospect
  actionDialogVisible.value = true
}

function closeActions() {
  actionDialogVisible.value = false
  actionTarget.value = null
}

function convertProspect(prospect: Prospect) {
  actionDialogVisible.value = false
  router.push(`/admin/prospects/${prospect.id}/convert`)
}

function requestDeleteFromActions(prospect: Prospect) {
  actionDialogVisible.value = false
  confirmDelete(prospect.id, prospect.placeName)
}

function requestApproveDeletionFromActions(prospect: Prospect) {
  actionDialogVisible.value = false
  confirmApproveDeletion(prospect)
}

async function requestRejectDeletionFromActions(prospect: Prospect) {
  actionDialogVisible.value = false
  await executeRejectDeletion(prospect)
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

function confirmApproveDeletion(p: Prospect) {
  deletionTarget.value = p
  showDeletionDialog.value = true
}

async function executeApproveDeletion() {
  if (!deletionTarget.value) return
  deletionBusy.value = true
  try {
    await approveProspectDeletion(deletionTarget.value.id)
    prospects.value = prospects.value.filter((p) => p.id !== deletionTarget.value!.id)
    showDeletionDialog.value = false
    deletionTarget.value = null
  } catch (e) {
    error.value = crm.errorMessage(e)
  } finally {
    deletionBusy.value = false
  }
}

async function executeRejectDeletion(p: Prospect) {
  try {
    await rejectProspectDeletion(p.id)
    const idx = prospects.value.findIndex((item) => item.id === p.id)
    if (idx >= 0) {
      prospects.value[idx] = { ...prospects.value[idx], deletionRequested: false }
    }
  } catch (e) {
    error.value = crm.errorMessage(e)
  }
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
  <section class="prospect-page">
    <header class="workspace-header">
      <div class="workspace-heading">
        <Button
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          rounded
          class="back-button"
          @click="router.back()"
          title="Back"
        />
        <div class="page-title-wrapper">
          <span class="eyebrow">Prospect Management</span>
          <h1>Prospect List</h1>
          <p class="muted">Manage, review, and track every prospect across the sales lifecycle.</p>
        </div>
      </div>

      <div class="summary-strip">
        <button type="button" class="summary-item">
          <i class="pi pi-inbox si-blue" />
          <span>Total</span>
          <strong>{{ prospects.length }}</strong>
        </button>
        <button type="button" class="summary-item">
          <i class="pi pi-chart-line si-violet" />
          <span>Active</span>
          <strong>{{ totalActive }}</strong>
        </button>
        <button type="button" class="summary-item won-summary">
          <i class="pi pi-trophy si-emerald" />
          <span>Won</span>
          <strong>{{ totalWon }}</strong>
        </button>
        <button type="button" class="summary-item">
          <i class="pi pi-times-circle si-red" />
          <span>Lost</span>
          <strong>{{ totalLost }}</strong>
        </button>
      </div>

      <div class="page-heading-actions">
        <Button
          label="Prospect Finder"
          icon="pi pi-search"
          severity="success"
          outlined
          size="small"
          @click="router.push('/admin/prospect-finder')"
        />
        <Button
          label="View Pipeline"
          icon="pi pi-columns"
          severity="secondary"
          outlined
          size="small"
          @click="router.push('/admin/prospects/pipeline')"
        />
      </div>
    </header>

    <Message v-if="error" severity="error" class="page-message">{{ error }}</Message>

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

      <section class="table-panel">
        <div class="table-heading">
          <div>
            <strong>Prospect Overview</strong>
            <span>{{ filtered.length }} records matching current filters</span>
          </div>
        </div>

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
              <tr
                  v-for="p in filtered"
                  :key="p.id"
                  :class="rowClass(p.status)"
                  class="prospect-row"
                  tabindex="0"
                  @click="openActions(p)"
                  @keydown.enter="openActions(p)"
                >
                <td>
                  <div class="cell-stack">
                    <button class="link-btn" type="button" @click.stop="openActions(p)">{{ p.placeName }}</button>
                    <span class="cell-sub">{{ p.formattedAddress }}</span>
                  </div>
                </td>
                <td><span class="cell-text">{{ p.placeCategory }}</span></td>
                <td><span class="cell-text">{{ p.industryGroup }}</span></td>
                <td><span class="cell-text">{{ p.assignedSalesExecutive }}</span></td>
                <td>
                  <div class="status-cell">
                    <span v-if="p.status === 'WON'" class="won-badge">
                      <i class="pi pi-trophy" />
                      Ready to Convert
                    </span>
                    <Tag
                      v-else
                      :value="p.status.replaceAll('_', ' ')"
                      :severity="statusSeverity(p.status)"
                    />
                    <span v-if="p.deletionRequested" class="deletion-badge">Deletion Requested</span>
                  </div>
                </td>
                <td><span class="cell-date">{{ formatDate(p.createdAt) }}</span></td>
                <td class="td-action">
                  <Button
                    label="Manage"
                    severity="secondary"
                    outlined
                    size="small"
                    class="manage-button"
                    @click.stop="openActions(p)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <Dialog
      v-model:visible="actionDialogVisible"
      modal
      :draggable="false"
      :dismissable-mask="true"
      header="Prospect Actions"
      class="prospect-action-dialog"
      :style="{ width: 'min(460px, calc(100vw - 2rem))' }"
      @hide="closeActions"
    >
      <div v-if="actionTarget" class="action-dialog-body">
        <div
          class="action-summary"
          :class="{
            won: actionTarget.status === 'WON',
            lost: actionTarget.status === 'LOST',
          }"
        >
          <div class="action-summary-main">
            <strong>{{ actionTarget.placeName }}</strong>
            <span>{{ actionTarget.formattedAddress || 'Address unavailable' }}</span>
          </div>
          <span
            v-if="actionTarget.status === 'WON'"
            class="action-status won"
          >
            WON · Ready to Convert
          </span>
          <Tag
            v-else
            :value="actionTarget.status.replaceAll('_', ' ')"
            :severity="statusSeverity(actionTarget.status)"
          />
        </div>

        <div class="action-menu">
          <button
            type="button"
            class="action-menu-item"
            @click="viewProspect(actionTarget.id)"
          >
            <span>
              <strong>View Details</strong>
              <small>Open prospect review, ticketing, and comments.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>

          <button
            v-if="actionTarget.status === 'WON'"
            type="button"
            class="action-menu-item primary"
            @click="convertProspect(actionTarget)"
          >
            <span>
              <strong>Convert to Customer</strong>
              <small>Create a customer record from this won prospect.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>

          <button
            v-if="actionTarget.deletionRequested"
            type="button"
            class="action-menu-item success"
            @click="requestApproveDeletionFromActions(actionTarget)"
          >
            <span>
              <strong>Approve Deletion Request</strong>
              <small>Approve the submitted deletion request.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>

          <button
            v-if="actionTarget.deletionRequested"
            type="button"
            class="action-menu-item warning"
            @click="requestRejectDeletionFromActions(actionTarget)"
          >
            <span>
              <strong>Reject Deletion Request</strong>
              <small>Keep the prospect and cancel the request.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>

          <button
            type="button"
            class="action-menu-item danger"
            @click="requestDeleteFromActions(actionTarget)"
          >
            <span>
              <strong>Delete Prospect</strong>
              <small>Permanently remove this prospect and related data.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>
        </div>
      </div>
    </Dialog>

    <Dialog v-model:visible="deleteDialogVisible" header="Delete Prospect" modal :draggable="false" :style="{ width: 'min(420px, calc(100vw - 2rem))' }">
      <p>Are you sure you want to delete <strong>{{ deleteTargetName }}</strong>? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>

    <Dialog v-model:visible="showDeletionDialog" header="Approve Deletion" modal :draggable="false" :style="{ width: 'min(420px, calc(100vw - 2rem))' }">
      <p v-if="deletionTarget" style="margin:0;font-size:0.85rem;line-height:1.5;">
        Approve deletion of <strong>{{ deletionTarget.placeName }}</strong>?
      </p>
      <p style="margin:0.5rem 0 0;font-size:0.78rem;color:var(--text-muted);">This will permanently remove the prospect and all associated data.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="showDeletionDialog = false" :disabled="deletionBusy" />
        <Button label="Approve & Delete" severity="danger" icon="pi pi-trash" :loading="deletionBusy" @click="executeApproveDeletion" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.prospect-page {
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
  grid-template-columns: minmax(280px, 1fr) auto auto;
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
  grid-template-columns: repeat(4, minmax(78px, auto));
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
  min-width: 78px;
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

.won-summary {
  background: #f0fdf4;
}

.won-summary span,
.won-summary strong {
  color: #15803d;
}

.si-blue { color: #d14350; }
.si-violet { color: #d15a66; }
.si-emerald { color: #16a34a; }
.si-red { color: #dc2626; }

.page-heading-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.45rem;
}

.page-message { margin: 0; }

.panel-stack {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.65rem;
}

.filter-panel {
  display: grid;
  grid-template-columns: minmax(250px, 1.4fr) minmax(0, 2.6fr);
  align-items: end;
  gap: 0.7rem;
  padding: 0.65rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.search-row { margin: 0; }

.search-field {
  display: flex;
  min-height: 38px;
  align-items: center;
  gap: 0.55rem;
  padding: 0.45rem 0.7rem;
  border: 1px solid #dbe3ee;
  border-radius: 8px;
  background: #f8fafc;
}

.search-field:focus-within {
  border-color: #d14350;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(209, 67, 80, 0.08);
}

.search-field i { color: #94a3b8; font-size: 0.76rem; }

.search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #0f172a;
  font-size: 0.76rem;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(130px, 1fr)) auto;
  align-items: end;
  gap: 0.5rem;
}

.filter-field {
  display: grid;
  min-width: 0;
  gap: 0.22rem;
}

.filter-field label {
  color: #64748b;
  font-size: 0.55rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.filter-field :deep(.p-select) {
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

.legend {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  padding: 0.2rem 0.48rem;
  border-radius: 999px;
  font-size: 0.58rem !important;
  font-weight: 800;
}

.legend-item.won {
  background: #dcfce7;
  color: #15803d !important;
}

.legend-item.lost {
  background: #fee2e2;
  color: #b91c1c !important;
}

.table-scroll {
  width: 100%;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: thin;
}

.data-table {
  width: 100%;
  min-width: 980px;
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
.th-action { width: 100px; text-align: center; }

.prospect-row-won {
  background: linear-gradient(90deg, #f0fdf4 0%, #ffffff 48%);
  box-shadow: inset 4px 0 0 #22c55e;
}

.prospect-row-won:hover {
  background: linear-gradient(90deg, #dcfce7 0%, #f8fffa 48%) !important;
}

.prospect-row-lost {
  background: linear-gradient(90deg, #fff7f7 0%, #ffffff 42%);
  box-shadow: inset 4px 0 0 #ef4444;
}

.prospect-row-converted {
  opacity: 0.72;
  background: #f8fafc;
}

.cell-stack {
  display: grid;
  min-width: 0;
  gap: 0.05rem;
}

.cell-sub,
.cell-text,
.cell-date {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-sub { max-width: 320px; color: #94a3b8; font-size: 0.62rem; }
.cell-text { color: #475569; font-size: 0.7rem; }
.cell-date { color: #64748b; font-size: 0.68rem; }

.link-btn {
  display: block;
  max-width: 100%;
  overflow: hidden;
  color: #bb3342;
  font-weight: 750;
  text-decoration: none;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-btn:hover {
  text-decoration: underline;
  text-underline-offset: 2px;
}

.status-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.28rem;
}

.won-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.32rem;
  padding: 0.24rem 0.55rem;
  border: 1px solid #86efac;
  border-radius: 999px;
  background: #dcfce7;
  color: #15803d;
  font-size: 0.58rem;
  font-weight: 850;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.won-badge i {
  color: #16a34a;
  font-size: 0.62rem;
}

.deletion-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  width: fit-content;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: #fef3c7;
  color: #92400e;
  font-size: 0.58rem;
  font-weight: 800;
  white-space: nowrap;
}

.deletion-badge::before {
  content: '';
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #d97706;
}

.td-action { text-align: center; }

.row-actions {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
}

.row-actions :deep(.p-button) {
  min-height: 1.9rem;
}

.act-view { color: #d14350 !important; }
.act-view:hover { background: #fff1f2 !important; }

.act-convert {
  min-width: 84px !important;
  border: 1px solid #86efac !important;
  background: #16a34a !important;
  color: #fff !important;
  font-size: 0.62rem !important;
  font-weight: 800 !important;
}

.act-convert:hover {
  border-color: #15803d !important;
  background: #15803d !important;
}
.act-convert :deep(.p-button-label),
.act-convert :deep(.p-button-icon) {
  color: #ffffff !important;
}

.act-convert:deep(*) {
  color: #ffffff !important;
}
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }

.act-approve {
  color: #16a34a !important;
  border: 1px solid #bbf7d0 !important;
}

.act-reject {
  color: #d97706 !important;
  border: 1px solid #fde68a !important;
}

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

.state-icon { color: #d14350; font-size: 1.5rem; }

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

@media (max-width: 1280px) {
  .workspace-header {
    grid-template-columns: minmax(240px, 1fr) auto;
  }

  .summary-strip {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  .filter-panel {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .prospect-page {
    min-height: auto;
    padding: 0.75rem;
    overflow: visible;
  }

  .workspace-header {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .page-title-wrapper .muted {
    white-space: normal;
  }

  .page-heading-actions {
    justify-content: stretch;
  }

  .page-heading-actions :deep(.p-button) {
    flex: 1;
  }

  .filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filter-action {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .prospect-page {
    padding: 0.6rem;
  }

  .summary-strip {
    grid-template-columns: repeat(2, 1fr);
  }

  .summary-item:nth-child(2) { border-right: 0; }
  .summary-item:nth-child(-n + 2) { border-bottom: 1px solid #e5eaf0; }

  .page-heading-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .filter-grid {
    grid-template-columns: 1fr;
  }

  .filter-action {
    grid-column: auto;
  }

  .table-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}

.prospect-row {
  cursor: pointer;
}

.prospect-row:focus-visible {
  outline: 2px solid #d14350;
  outline-offset: -2px;
}

.manage-button {
  min-width: 72px !important;
  height: 30px !important;
  border-color: #cbd5e1 !important;
  background: #fff !important;
  color: #475569 !important;
  font-size: 0.62rem !important;
  font-weight: 750 !important;
}

.manage-button:hover {
  border-color: #94a3b8 !important;
  background: #f8fafc !important;
  color: #0f172a !important;
}

.prospect-action-dialog :deep(.p-dialog) {
  border-radius: 14px;
}

.action-dialog-body {
  display: grid;
  gap: 0.75rem;
}

.action-summary {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #f8fafc;
}

.action-summary.won {
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.action-summary.lost {
  border-color: #fecaca;
  background: #fff7f7;
}

.action-summary-main {
  display: grid;
  gap: 0.12rem;
}

.action-summary-main strong {
  color: #0f172a;
  font-size: 0.85rem;
}

.action-summary-main span {
  display: -webkit-box;
  overflow: hidden;
  color: #64748b;
  font-size: 0.68rem;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.action-status {
  width: fit-content;
  padding: 0.22rem 0.5rem;
  border-radius: 999px;
  font-size: 0.58rem;
  font-weight: 850;
}

.action-status.won {
  border: 1px solid #86efac;
  background: #dcfce7;
  color: #15803d;
}

.action-menu {
  display: grid;
  gap: 0.45rem;
}

.action-menu-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 20px;
  align-items: center;
  gap: 0.7rem;
  width: 100%;
  padding: 0.72rem 0.75rem;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #fff;
  text-align: left;
  cursor: pointer;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    transform 150ms ease;
}

.action-menu-item:hover {
  transform: translateY(-1px);
  border-color: #f3b9c0;
  background: #fffbfb;
}

.action-menu-item > span:first-child {
  display: grid;
  min-width: 0;
  gap: 0.1rem;
}

.action-menu-item strong {
  color: #0f172a;
  font-size: 0.74rem;
}

.action-menu-item small {
  color: #64748b;
  font-size: 0.64rem;
  line-height: 1.4;
}

.action-arrow {
  color: #94a3b8;
  font-size: 1rem;
  font-weight: 800;
  text-align: right;
}

.action-menu-item.primary {
  border-color: #86efac;
  background: #f0fdf4;
}

.action-menu-item.primary strong,
.action-menu-item.success strong {
  color: #15803d;
}

.action-menu-item.success {
  border-color: #bbf7d0;
}

.action-menu-item.warning {
  border-color: #fde68a;
  background: #fffbeb;
}

.action-menu-item.warning strong {
  color: #a16207;
}

.action-menu-item.danger {
  border-color: #fecaca;
  background: #fff7f7;
}

.action-menu-item.danger strong {
  color: #b91c1c;
}

</style>
