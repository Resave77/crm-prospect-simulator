<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { getPipeline, getSalesExecutives, trashProspect, listTrashedProspects, restoreProspect, approveProspectDeletion, rejectProspectDeletion } from '../../../api/crm'
import { listCategories, type MasterDataCategory } from '../../../api/masterData'
import { fallbackCategories } from '../../../utils/masterDataFallback'
import { BOARD_STATUSES, filterProspects } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus, SalesExecutiveOption } from '../../../types/crm'

const crm = useCrmStore()
const router = useRouter()
const route = useRoute()
const prospects = ref<Prospect[]>([])
const sales = ref<SalesExecutiveOption[]>([])
const error = ref('')
const loading = ref(true)
const showFilters = ref(false)
const prospectTrashEnabled = true
const prospectPage = ref(1)
const prospectPageSize = ref(10)
const selectedProspectIds = ref<Set<string>>(new Set())
const selectedProspectCount = computed(() => selectedProspectIds.value.size)

const searchQuery = ref('')
const salesFilter = ref('')
const categoryFilter = ref('')
const statusFilter = ref('')
const masterCategories = ref<MasterDataCategory[]>(fallbackCategories)
const prospectTotalPages = computed(() => Math.max(1, Math.ceil(allFiltered.value.length / prospectPageSize.value)))
watch(() => route.query.search, (value) => {
  searchQuery.value = typeof value === 'string' ? value : ''
}, { immediate: true })

const deleteDialogVisible = ref(false)
const deleteTargetId = ref('')
const deleteTargetName = ref('')
const deleting = ref(false)
const trashVisible = ref(false)
const trashedProspects = ref<Prospect[]>([])
async function openTrash() {
  if (!prospectTrashEnabled) return
  if (selectedProspectIds.value.size) {
    confirmSelectedDelete()
    return
  }
  trashVisible.value = true
  try { trashedProspects.value = await listTrashedProspects() } catch (e) { error.value = crm.errorMessage(e) }
}

async function restoreTrashedProspect(id: string) {
  if (!prospectTrashEnabled) return
  try {
    await restoreProspect(id)
    trashedProspects.value = trashedProspects.value.filter((item) => item.id !== id)
    await crm.loadPipeline()
    prospects.value = crm.pipeline
  } catch (e) { error.value = crm.errorMessage(e) }
}

const deletionTarget = ref<Prospect | null>(null)
const showDeletionDialog = ref(false)
const deletionBusy = ref(false)

const actionDialogVisible = ref(false)
const actionTarget = ref<Prospect | null>(null)

const categoryOptions = computed(() => [{ label: 'All Categories', value: '' }, ...masterCategories.value.map((category) => ({ label: category.name, value: category.name }))])
const statusOptions = computed(() => [{ label: 'All Pipeline Statuses', value: '' }, ...BOARD_STATUSES.map((v) => ({ label: v.replaceAll('_', ' '), value: v }))])
const salesOptions = computed(() => [{ label: 'All Sales Executives', value: '' }, ...sales.value.map((s) => ({ label: s.fullName, value: s.id }))])

const allFiltered = computed(() => {
  const visibleProspects = prospects.value.filter((item) => item.status !== 'CONVERTED')
  return filterProspects(visibleProspects, {
    salesExecutiveId: salesFilter.value,
    industryGroup: '',
    category: categoryFilter.value,
    status: statusFilter.value,
    search: searchQuery.value,
  })
})

const paginatedProspects = computed(() => {
  const start = (prospectPage.value - 1) * prospectPageSize.value
  return allFiltered.value.slice(start, start + prospectPageSize.value)
})
const filtered = paginatedProspects

const allVisibleProspectsSelected = computed(() => (
  filtered.value.length > 0 && filtered.value.every((prospect) => selectedProspectIds.value.has(prospect.id))
))

function toggleProspectSelection(id: string) {
  const next = new Set(selectedProspectIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedProspectIds.value = next
}

function toggleAllVisibleProspects() {
  const next = new Set(selectedProspectIds.value)
  if (allVisibleProspectsSelected.value) filtered.value.forEach((prospect) => next.delete(prospect.id))
  else filtered.value.forEach((prospect) => next.add(prospect.id))
  selectedProspectIds.value = next
}

function updateProspectPageSize(size: number) {
  prospectPageSize.value = Math.max(1, size)
  prospectPage.value = 1
}

function goToProspectPage(page: number) {
  prospectPage.value = Math.max(1, Math.min(page, prospectTotalPages.value))
}

const totalActive = computed(() => prospects.value.filter((p) => !['CONVERTED', 'LOST'].includes(p.status)).length)
const totalWon = computed(() => prospects.value.filter((p) => p.status === 'WON').length)
const totalLost = computed(() => prospects.value.filter((p) => p.status === 'LOST').length)

watch([searchQuery, salesFilter, categoryFilter, statusFilter], () => {
  prospectPage.value = 1
})

watch(prospectTotalPages, (total) => {
  if (prospectPage.value > total) prospectPage.value = total
})

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
  categoryFilter.value = ''
  statusFilter.value = ''
}

function statusBadgeClass(status: ProspectStatus) {
  switch (status) {
    case 'LOST': return 'border-[#fecaca] bg-[#fff1f2] text-[#b91c1c]'
    case 'WON':
    case 'CONVERTED': return 'border-[#bbf7d0] bg-[#f0fdf4] text-[#166534]'
    case 'INTERESTED':
    case 'QUALIFIED': return 'border-[#bbf7d0] bg-[#f0fdf4] text-[#166534]'
    case 'PROPOSAL_SENT':
    case 'NEGOTIATION': return 'border-[#fed7aa] bg-[#fff7ed] text-[#c2410c]'
    default: return 'border-[#bfdbfe] bg-[#dbeafe] text-[#1d4ed8]'
  }
}

function segmentForProspect(prospect: Prospect) {
  const categoryName = (prospect.industryGroup || prospect.placeCategory || '').trim().toLowerCase()
  return masterCategories.value.find((category) => category.name.trim().toLowerCase() === categoryName)?.segmentName || ''
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

function confirmSelectedDelete() {
  if (!selectedProspectIds.value.size) return
  deleteTargetId.value = ''
  deleteTargetName.value = `${selectedProspectIds.value.size} selected prospect${selectedProspectIds.value.size === 1 ? '' : 's'}`
  deleteDialogVisible.value = true
}

async function executeDelete() {
  if (!prospectTrashEnabled) return
  if (!deleteTargetId.value && selectedProspectIds.value.size) {
    deleteDialogVisible.value = false
    await moveSelectedToTrash()
    return
  }
  deleting.value = true
  try {
    await trashProspect(deleteTargetId.value)
    prospects.value = prospects.value.filter((p) => p.id !== deleteTargetId.value)
    deleteDialogVisible.value = false
  } catch (e) {
    error.value = crm.errorMessage(e)
  } finally {
    deleting.value = false
  }
}

async function moveSelectedToTrash() {
  if (!prospectTrashEnabled || !selectedProspectIds.value.size) return
  deleting.value = true
  try {
    const ids = [...selectedProspectIds.value]
    await Promise.all(ids.map((id) => trashProspect(id)))
    const deletedIds = new Set(ids)
    prospects.value = prospects.value.filter((prospect) => !deletedIds.has(prospect.id))
    selectedProspectIds.value = new Set()
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
    try {
      const categoryResult = await listCategories()
      if (categoryResult?.length) masterCategories.value = categoryResult
    } catch {
      masterCategories.value = fallbackCategories
    }
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
        <div class="page-title-wrapper">
          <span class="eyebrow">Manajemen Prospek</span>
          <h1>Daftar Prospek</h1>
          <p class="muted">Kelola, tinjau, dan pantau seluruh prospek dalam proses penjualan.</p>
        </div>
      </div>

      <div class="summary-strip">
        <button type="button" class="summary-item">
          <i class="pi pi-inbox si-blue" />
          <span>Total</span>
          <strong>{{ prospects.filter((item) => item.status !== 'CONVERTED').length }}</strong>
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

    </header>

    <Message v-if="error" severity="error" class="page-message">{{ error }}</Message>

<nav class="prospect-erp-toolbar flex min-w-0 flex-wrap items-center gap-[10px] overflow-x-auto pb-[1px]" aria-label="Prospect management sections"><label class="prospect-search relative w-[260px] shrink-0 lg:w-[300px]"><i class="pi pi-search" /><input v-model="searchQuery" placeholder="Search prospect name, category, sales executive" /></label><Button label="More Filters" icon="pi pi-sliders-h" severity="secondary" outlined size="small" @click="showFilters = !showFilters" /><button v-if="prospectTrashEnabled" type="button" class="prospect-trash-button ml-auto flex h-[36px] w-[82px] shrink-0 items-center justify-center gap-[7px] rounded-[10px] border border-[#fecaca] bg-[#fff5f5] px-[12px] font-['Inter'] text-[12px] font-semibold text-[#b91c1b] transition-all hover:bg-[#fff1f2]" @click="openTrash"><i class="pi pi-trash text-[15px]" /><span>{{ selectedProspectCount ? 'Move to Trash' : 'Trash Prospect' }}</span></button></nav>
    <div class="panel-stack">
      <div v-if="showFilters" class="filter-panel">
        <div class="filter-grid">
          <div class="filter-field">
            <label>Sales Executive</label>
            <Select v-model="salesFilter" :options="salesOptions" optionLabel="label" optionValue="value" placeholder="All Sales" />
          </div>
          <div class="filter-field">
            <label>Category</label>
            <Select v-model="categoryFilter" :options="categoryOptions" optionLabel="label" optionValue="value" />
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
<div class="pagination-bar"><div class="flex items-center gap-[12px]"><div class="pagination-controls flex items-center gap-[4px]"><button type="button" class="p-button p-component p-button-icon-only p-button-rounded p-button-text p-button-sm" :disabled="prospectPage <= 1" aria-label="Previous page" @click="goToProspectPage(prospectPage - 1)"><i class="pi pi-angle-left" /></button><button v-for="page in prospectTotalPages" :key="page" type="button" :class="['p-button p-component p-button-rounded p-button-text p-button-sm pagination-num', { 'is-active': page === prospectPage }]" :aria-label="`Page ${page}`" @click="goToProspectPage(page)">{{ page }}</button><button type="button" class="p-button p-component p-button-icon-only p-button-rounded p-button-text p-button-sm" :disabled="prospectPage >= prospectTotalPages" aria-label="Next page" @click="goToProspectPage(prospectPage + 1)"><i class="pi pi-angle-right" /></button></div><span class="pagination-info ml-[8px]">Page {{ prospectPage }} of {{ prospectTotalPages }} / {{ allFiltered.length }} records</span></div><div class="pagination-settings"><label class="page-size-control"><span>Page size</span><select v-model="prospectPageSize" aria-label="Page size" @change="updateProspectPageSize(prospectPageSize)"><option v-for="size in [10, 20, 50]" :key="size" :value="size">{{ size }}</option></select></label><label><span>Go to</span><input v-model.number="prospectPage" type="number" min="1" :max="prospectTotalPages" aria-label="Go to page" @change="goToProspectPage(prospectPage)" /></label><button type="button" @click="goToProspectPage(prospectPage)">Set</button><button type="button" class="pagination-refresh" title="Refresh prospect data" @click="crm.loadPipeline()"><i class="pi pi-refresh" /></button></div></div>

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
                <th class="w-[56px] border-r border-[#e0e0e0] bg-[#f4f4f4] px-[12px] text-center"><button type="button" :class="['prospect-checkbox mx-auto flex size-[18px] items-center justify-center rounded-[4px] border bg-white shadow-sm', { 'prospect-checkbox-selected': allVisibleProspectsSelected }]" aria-label="Select all visible prospects" :aria-pressed="allVisibleProspectsSelected" @click.stop="toggleAllVisibleProspects"><i v-if="allVisibleProspectsSelected" class="pi pi-check text-[11px]" /></button></th>
                <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Place Name</span></th>
                <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Category</span></th>
                <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Sales Executive</span></th>
                <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Status</span></th>
                <th class="border-r border-[#e0e0e0] bg-[#f4f4f4] p-[12px] text-left last:border-r-0"><span class="select-none font-['Inter'] text-[11px] font-semibold uppercase tracking-wider text-black">Created</span></th>
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
                <td class="border-r border-[#e0e0e0] px-[12px] text-center"><button type="button" :class="['prospect-checkbox mx-auto flex size-[18px] items-center justify-center rounded-[4px] border bg-white shadow-sm', { 'prospect-checkbox-selected': selectedProspectIds.has(p.id) }]" :aria-label="`Select prospect ${p.placeName}`" :aria-pressed="selectedProspectIds.has(p.id)" @click.stop="toggleProspectSelection(p.id)"><i v-if="selectedProspectIds.has(p.id)" class="pi pi-check text-[11px]" /></button></td>
                <td class="prospect-name-cell">
                  <div class="min-w-0">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ p.placeName }}</p>
                    <p class="mt-[2px] truncate font-['Inter'] text-[12px] text-[#64748b]">{{ p.formattedAddress || 'Address unavailable' }}</p>
                  </div>
                </td>
                <td class="preview-cell">
                  <div class="min-w-0">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ p.placeCategory || 'Not provided' }}</p>
                    <p v-if="segmentForProspect(p)" class="mt-[2px] truncate font-['Inter'] text-[12px] text-[#64748b]">{{ segmentForProspect(p) }}</p>
                  </div>
                </td>
                <td class="preview-cell">
                  <div class="min-w-0">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ p.assignedSalesExecutive || 'Unassigned' }}</p>
                  </div>
                </td>
                <td class="preview-cell">
                  <div class="status-cell">
                    <span v-if="p.status === 'WON'" class="won-badge">
                      <i class="pi pi-trophy" />
                      Ready to Convert
                    </span>
                    <span
                      v-else
                      class="inline-flex w-fit items-center justify-center whitespace-nowrap rounded-full border px-[10px] py-[4px] font-['Inter'] text-[11px] font-semibold leading-[16px]"
                      :class="statusBadgeClass(p.status)"
                    >{{ p.status.replaceAll('_', ' ') }}</span>
                    <span v-if="p.deletionRequested" class="deletion-badge">Deletion Requested</span>
                  </div>
                </td>
                <td class="preview-cell">
                  <div class="min-w-0">
                    <p class="truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ formatDate(p.createdAt) }}</p>
                  </div>
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
              <strong>Move to Trash</strong>
              <small>Move this prospect to Trash. It can be restored later.</small>
            </span>
            <span class="action-arrow">›</span>
          </button>
        </div>
      </div>
    </Dialog>

    <Dialog v-if="prospectTrashEnabled" v-model:visible="trashVisible" header="Prospect Trash" modal :draggable="false" :style="{ width: 'min(620px, calc(100vw - 2rem))' }">
      <div v-if="trashedProspects.length" class="space-y-2">
        <div v-for="item in trashedProspects" :key="item.id" class="flex items-center justify-between gap-3 rounded-lg border border-slate-200 p-3">
          <div class="min-w-0"><strong class="block truncate">{{ item.placeName }}</strong><small class="text-slate-500">{{ item.formattedAddress }}</small></div>
          <Button label="Restore" icon="pi pi-refresh" size="small" outlined @click="restoreTrashedProspect(item.id)" />
        </div>
      </div>
      <p v-else class="m-0 text-sm text-slate-500">Trash is empty.</p>
      <template #footer><Button label="Close" severity="secondary" text @click="trashVisible = false" /></template>
    </Dialog>

    <Dialog v-if="prospectTrashEnabled" v-model:visible="deleteDialogVisible" header="Move to Trash" modal :draggable="false" :style="{ width: 'min(420px, calc(100vw - 2rem))' }">
      <p>Move <strong>{{ deleteTargetName }}</strong> to Trash? You can restore it later.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Move to Trash" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
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

.si-blue { color: #e63946; }
.si-violet { color: #ef4e5d; }
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
  border-color: #e63946;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(230, 57, 70, 0.08);
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
  color: #d62839;
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

.act-view { color: #e63946 !important; }
.act-view:hover { background: #fff0f1 !important; }

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
  outline: 2px solid #e63946;
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
  border-color: #f4b3ba;
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

@media(min-width:901px){.prospect-page{width:100%;max-width:none;min-height:100%;box-sizing:border-box}.prospect-page .workspace-header,.prospect-page .filter-panel,.prospect-page .table-panel{width:100%;box-sizing:border-box}.prospect-page .filter-panel{display:flex;align-items:flex-end;justify-content:flex-start;gap:.6rem;padding:.55rem .7rem}.prospect-page .filter-grid{display:grid;grid-template-columns:repeat(3,minmax(145px,185px)) auto;flex:0 1 auto;gap:.5rem}.prospect-page .filter-field :deep(.p-select){height:36px;background:#fff}.prospect-page .data-table{min-width:0;width:100%;font-size:.74rem}.prospect-page .data-table thead th{height:44px;padding:.55rem .7rem;background:#f4f6f9;color:#3f5068}.prospect-page .data-table tbody td{height:58px;padding:.55rem .7rem}.prospect-page .pagination-bar{min-height:48px;padding:.55rem .75rem}.prospect-page .summary-strip{border-radius:10px}.prospect-page .page-heading-actions :deep(.p-button){min-height:34px;border-radius:8px}}
.prospect-name{display:block;max-width:100%;overflow:hidden;color:#172033;font-size:.76rem;line-height:1.3;text-overflow:ellipsis;white-space:nowrap}.prospect-row{cursor:pointer;transition:background .16s,box-shadow .16s}.prospect-row:hover{background:#fffafa}.prospect-row:focus-visible{outline:2px solid #e63946;outline-offset:-2px}.action-dialog-body{padding:.15rem}.action-summary{padding:.75rem;border:1px solid #e1e8f0;border-radius:12px;background:#f8fafc}.action-summary-main{display:grid;gap:.25rem}.action-summary-main strong{font-size:.9rem;color:#172033}.action-summary-main span{font-size:.68rem;color:#64748b;line-height:1.4}
.prospect-name-cell{position:relative;overflow:visible!important}.prospect-preview{position:absolute;z-index:50;left:.25rem;top:calc(100% - .1rem);display:grid;min-width:220px;gap:.16rem;padding:.55rem .65rem;border:1px solid #dce5f0;border-radius:8px;background:#fff;box-shadow:0 8px 18px rgba(15,23,42,.14);color:#52627a;font-size:.57rem;line-height:1.25}.prospect-preview strong{color:#075de3;font-size:.62rem}.prospect-preview span{white-space:nowrap}@media(min-width:901px){.table-panel:has(.prospect-name-cell),.table-panel:has(.prospect-name-cell) .table-scroll{overflow:visible}}
.preview-cell{position:relative;overflow:visible!important}.cell-preview{position:absolute;z-index:60;left:.25rem;top:calc(100% - .05rem);min-width:145px;padding:.5rem .6rem;border:1px solid #dce5f0;border-radius:7px;background:#fff;box-shadow:0 8px 18px rgba(15,23,42,.14);color:#52627a;font-size:.6rem;line-height:1.35;white-space:nowrap}.prospect-page .table-panel{overflow:visible}.prospect-page .table-scroll{overflow:visible}
.prospect-page .data-table{width:100%;min-width:0;table-layout:fixed}.prospect-page .data-table thead th{height:42px;padding:.55rem .7rem;background:#f4f6f9}.prospect-page .data-table tbody td{height:60px;padding:.55rem .7rem;border-right:1px solid #e5eaf0}.prospect-page .table-panel{width:100%;overflow:visible}
@media (max-width: 640px) {
  .prospect-page { padding:.5rem; gap:.5rem; background:#f8fafc; }
  .prospect-page .workspace-header { padding:.85rem; border-radius:14px; }
  .prospect-page .page-title-wrapper h1 { font-size:1.12rem; }
  .prospect-page .page-title-wrapper .muted { font-size:.64rem; line-height:1.35; white-space:normal; }
  .prospect-page .summary-strip { grid-template-columns:repeat(2,1fr); border-radius:10px; }
  .prospect-page .page-heading-actions { display:grid; grid-template-columns:repeat(2,1fr); gap:.4rem; }
  .prospect-page .page-heading-actions :deep(.p-button) { min-height:38px; font-size:.68rem; }
  .prospect-page .filter-panel { padding:.7rem; border-radius:12px; }
  .prospect-page .filter-grid { grid-template-columns:1fr; gap:.55rem; }
  .prospect-page .filter-field :deep(.p-select) { height:40px; border-radius:9px; }
  .prospect-page .table-panel { border-radius:12px; overflow:hidden; }
  .prospect-page .table-scroll { overflow:hidden; }
  .prospect-page .data-table { min-width:0; table-layout:fixed; font-size:.65rem; border-collapse:separate; border-spacing:0; }
  .prospect-page .data-table thead th { height:34px; padding:.45rem .35rem; font-size:.54rem; background:#fff8f8; color:#8b4b55; }
  .prospect-page .data-table tbody tr { background:#fff; }
  .prospect-page .data-table tbody tr + tr td { border-top:1px solid #f4e6e8; }
  .prospect-page .data-table tbody td { height:54px; padding:.45rem .35rem; vertical-align:middle; }
  .prospect-page .data-table th, .prospect-page .data-table td { padding:.6rem .45rem; }
  .prospect-page .data-table thead th:nth-child(2), .prospect-page .data-table tbody td:nth-child(2),
  .prospect-page .data-table thead th:nth-child(5), .prospect-page .data-table tbody td:nth-child(5) { display:none; }
  .prospect-page .data-table thead th { font-size:0; }
  .prospect-page .data-table thead th:nth-child(1)::after { content:'Prospek'; font-size:.56rem; }
  .prospect-page .data-table thead th:nth-child(3)::after { content:'Sales'; font-size:.56rem; }
  .prospect-page .data-table thead th:nth-child(4)::after { content:'Status'; font-size:.56rem; }
  .prospect-page .prospect-name { font-size:.7rem; }
  .prospect-page .cell-sub { display:block; max-width:150px; font-size:.55rem; }
  .prospect-page .cell-text { font-size:.62rem; }
  .prospect-page .p-tag, .prospect-page .won-badge { font-size:.54rem; padding:.18rem .3rem; white-space:normal; }
  .prospect-page .pagination-bar { flex-direction:column; gap:.4rem; padding:.65rem .7rem; }
  .prospect-page .pagination-info { font-size:.6rem; text-align:center; }
  .prospect-page .state-box { min-height:180px; padding:1.2rem; font-size:.68rem; }
}
.prospect-page { min-width:0; padding:0; gap:0; background:#fff; font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif; }
.prospect-page .workspace-header { display:grid; }
.prospect-page .table-heading { display:none; }
.prospect-erp-toolbar { display:flex; min-height:58px; height:58px; align-items:center; gap:10px; padding:10px 20px; border-bottom:1px solid #e2e8f0; background:#fff; }
.prospect-tabs { display:flex; align-items:center; gap:2px; padding:4px; border-radius:10px; background:#f1f5f9; }
.prospect-tab { display:inline-flex; height:34px; align-items:center; padding:0 14px; border-radius:10px; color:#40516a; font-size:12px; font-weight:500; white-space:nowrap; }
.prospect-tab.active { background:#dc2626; color:#fff; box-shadow:0 2px 6px rgba(220,38,38,.12); }
.prospect-search { display:flex; flex:0 0 270px; height:40px; align-items:center; gap:8px; padding:0 12px; border:1px solid #e2e8f0; border-radius:10px; color:#94a3b8; }
.prospect-search i { font-size:12px; }.prospect-search input { width:100%; height:32px; border:0; outline:0; background:transparent; color:#0f172a; font-size:13px; }.prospect-search input::placeholder { color:#94a3b8; }
.prospect-erp-toolbar > :deep(.p-button) { height:40px; min-height:40px; padding:0 14px; border-radius:10px; font-size:12px; font-weight:600; }
.prospect-page .panel-stack { gap:0; padding:0; }.prospect-page .filter-panel { display:flex; align-items:flex-end; gap:14px; padding:14px 20px 16px; border:0; border-bottom:1px solid #e2e8f0; border-radius:0; background:#fff; box-shadow:0 3px 12px rgba(15,23,42,.04); }.prospect-page .filter-grid { display:grid; grid-template-columns:repeat(3,minmax(160px,210px)) auto; flex:1; gap:12px; }.prospect-page .filter-field { gap:5px; }.prospect-page .filter-field label { color:#64748b; font-size:10px; font-weight:600; letter-spacing:.04em; text-transform:uppercase; }.prospect-page .filter-field :deep(.p-select) { height:40px; border:1px solid #e2e8f0; border-radius:10px; background:#fff; font-size:12px; }.prospect-page .filter-field :deep(.p-select-label) { padding:0 12px; }.prospect-page .filter-action { align-self:end; }.prospect-page .filter-action :deep(.p-button) { height:40px; min-height:40px; border-radius:10px; font-size:12px; }
.prospect-page .table-panel { border:0; border-radius:0; box-shadow:none; }.prospect-page .data-table { min-width:0; width:100%; table-layout:fixed; font-size:11px; }.prospect-page .data-table thead th { box-sizing:border-box; height:42px; padding:0 10px; border-right:1px solid #e2e2e2; background:#f4f4f4; color:#000; font-size:10px; font-weight:600; line-height:12px; letter-spacing:.07em; }.prospect-page .data-table tbody td { box-sizing:border-box; height:67px; padding:8px 10px; border-right:1px solid #e2e2e2; border-bottom:1px solid #e0e0e0; font-size:11px; line-height:13px; }.prospect-page .data-table tbody tr:hover { background:#fff; }.prospect-page .data-table .prospect-name { color:#0f172a; font-size:11px; font-weight:600; }.prospect-page .data-table .cell-sub,.prospect-page .data-table .cell-text,.prospect-page .data-table .cell-date { color:#64748b; font-size:10px; line-height:12px; }.prospect-page .data-table :deep(.p-tag) { border-radius:999px; padding:4px 10px; font-size:10px; font-weight:500; }.prospect-page .pagination-bar { min-height:42px; padding:0 20px; border-top:0; border-bottom:1px solid #e2e8f0; background:#fff; }
@media(max-width:900px){.prospect-erp-toolbar{height:auto;min-height:58px;flex-wrap:wrap;overflow:visible}.prospect-search{flex-basis:220px}.prospect-erp-toolbar>:deep(.p-button){width:100%}.prospect-page .filter-panel{display:grid;grid-template-columns:1fr;padding:12px}.prospect-page .filter-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:640px){.prospect-erp-toolbar{padding:10px 12px}.prospect-tabs{width:100%}.prospect-tab{flex:1;justify-content:center;padding:0 8px}.prospect-page .data-table{min-width:760px}.prospect-page .table-scroll{overflow-x:auto}}

/* Page-level ERP list alignment: keep the hierarchy visible and the data surface contained. */
.prospect-page .workspace-header { display: none; }
.prospect-page .prospect-erp-toolbar { margin-top: 0; }
.prospect-page .table-panel {
  margin: 0;
  overflow: visible;
  border: 0;
  border-radius: 0;
  box-shadow: none;
}
.prospect-page .data-table thead th { background: #f4f4f4; color: #000; }
.prospect-page .data-table tbody tr:hover { background: #fff; }
.prospect-page .data-table th:nth-child(1), .prospect-page .data-table td:nth-child(1) { width: 44%; }
.prospect-page .data-table th:nth-child(2), .prospect-page .data-table td:nth-child(2) { width: 18%; }
.prospect-page .data-table th:nth-child(3), .prospect-page .data-table td:nth-child(3) { width: 18%; }
.prospect-page .data-table th:nth-child(4), .prospect-page .data-table td:nth-child(4) { width: 12%; }
.prospect-page .data-table th:nth-child(5), .prospect-page .data-table td:nth-child(5) { width: 8%; }
@media(max-width:900px) {
  .prospect-page .prospect-erp-toolbar { margin-top: 10px; }
}
@media(max-width:640px) {
  .prospect-page .table-panel { border-radius: 0; }
}
.prospect-page .pagination-bar {
  min-height: 44px !important;
  height: 44px !important;
  box-sizing: border-box !important;
  padding: 5px 16px !important;
  gap: 12px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
}
.prospect-page .pagination-bar > div:first-child { gap: 12px !important; }
.prospect-page .pagination-bar > div:first-child > div { gap: 4px !important; }
.prospect-page .pagination-bar button { box-sizing: border-box !important; }
.prospect-page .pagination-bar > div:first-child > div > button {
  width: 30px !important;
  height: 30px !important;
  min-width: 30px !important;
  min-height: 30px !important;
  padding: 0 !important;
  font-size: 12px !important;
  line-height: 30px !important;
}
.prospect-page .pagination-bar .pagination-settings { gap: 10px !important; }
.prospect-page .pagination-bar .pagination-settings > div { height: 34px !important; }
.prospect-page .data-table th:nth-child(1), .prospect-page .data-table td:nth-child(1) { width: 56px !important; }
.prospect-page .data-table th:nth-child(2), .prospect-page .data-table td:nth-child(2) { width: 38% !important; }
.prospect-page .data-table th:nth-child(3), .prospect-page .data-table td:nth-child(3) { width: 17% !important; }
.prospect-page .data-table th:nth-child(4), .prospect-page .data-table td:nth-child(4) { width: 17% !important; }
.prospect-page .data-table th:nth-child(5), .prospect-page .data-table td:nth-child(5) { width: 12% !important; }
.prospect-page .data-table th:nth-child(6), .prospect-page .data-table td:nth-child(6) { width: 10% !important; }
.prospect-page { padding-left: 0 !important; }
.prospect-page .prospect-erp-toolbar { padding-left: 0 !important; }
.prospect-page .prospect-erp-toolbar { padding-left: 20px !important; }
.prospect-page .table-panel { margin-left: 0 !important; }
.prospect-page .data-table thead th:not(:first-child) {
  border-right: 1px solid #e0e0e0 !important;
  background: #f4f4f4 !important;
  padding: 12px !important;
  text-align: left !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 11px !important;
  font-weight: 600 !important;
  line-height: 1.25 !important;
  letter-spacing: 0.05em !important;
  text-transform: uppercase !important;
  color: #000 !important;
}
.prospect-page .status-cell > span:not(.won-badge):not(.deletion-badge) {
  display: inline-flex !important;
  width: fit-content !important;
  align-items: center !important;
  justify-content: center !important;
  border-radius: 999px !important;
  padding: 4px 10px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 11px !important;
  font-weight: 600 !important;
  line-height: 16px !important;
}
.prospect-page .prospect-checkbox {
  appearance: none !important;
  display: inline-flex !important;
  width: 18px !important;
  height: 18px !important;
  padding: 0 !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 4px !important;
  background: #fff !important;
  color: transparent !important;
}

.prospect-page .prospect-checkbox:focus-visible {
  outline: 2px solid #93c5fd !important;
  outline-offset: 2px;
}
.prospect-page .prospect-checkbox-selected {
  border-color: #ef4444 !important;
  background: #ef4444 !important;
  color: #fff !important;
}
.prospect-page .prospect-trash-button,
.prospect-page .prospect-trash-button span {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  line-height: 16px !important;
  letter-spacing: normal !important;
}
.prospect-page .prospect-erp-toolbar .prospect-trash-button {
  width: max-content !important;
  min-width: max-content !important;
  padding: 0 14px !important;
  white-space: nowrap;
}
.prospect-page .pagination-bar > div:first-child > div:nth-child(3) > button:first-child{display:none!important}
.prospect-page .pagination-bar{display:flex;align-items:center;justify-content:space-between;min-height:42px;padding:0 20px;border-bottom:1px solid #e2e8f0;background:#fff;font-family:Inter,ui-sans-serif,system-ui,sans-serif;font-size:12px;color:#64748b}.prospect-page .pagination-controls{display:flex;align-items:center;gap:4px}.prospect-page .pagination-controls button{display:inline-flex;align-items:center;justify-content:center;min-width:30px;height:30px;padding:0;border:0;border-radius:999px;background:transparent;color:#475569;cursor:pointer}.prospect-page .pagination-controls button:hover:not(:disabled),.prospect-page .pagination-controls button.pagination-num.is-active{background:#fff1f2;color:#991b1b;font-weight:700}.prospect-page .pagination-controls button:disabled{cursor:not-allowed;opacity:.45}.prospect-page .pagination-settings{display:flex;align-items:center;gap:10px}.prospect-page .pagination-settings label{position:relative;display:flex;width:80px;height:34px;align-items:center;padding:0 10px;border:1px solid #e2e8f0;border-radius:8px;color:#334155}.prospect-page .pagination-settings label span{position:absolute;top:-7px;left:8px;padding:0 4px;background:#fff;font-size:10px;font-weight:600;color:#475569}.prospect-page .pagination-settings select,.prospect-page .pagination-settings input{width:100%;border:0;outline:0;background:transparent;color:#334155;font-size:11px}.prospect-page .pagination-settings input{padding-top:4px}.prospect-page .pagination-settings>button{height:34px;padding:0 12px;border:1px solid #e2e8f0;border-radius:8px;background:#fff;color:#475569;cursor:pointer}.prospect-page .pagination-settings .pagination-refresh{width:30px;padding:0;border:0;border-radius:999px}
.prospect-page .data-table thead th{height:49px!important;padding:0 12px!important;border-right:1px solid #e2e2e2!important;border-bottom:1px solid #e0e0e0!important;background:#f4f4f4!important;color:#000!important;font-size:10px!important;font-weight:600!important;letter-spacing:.07em!important;text-transform:uppercase}.prospect-page .data-table tbody td{height:84px!important;padding:10px 12px!important;border-right:1px solid #e2e8f0!important;border-bottom:1px solid #e0e0e0!important;font-size:12px!important}.prospect-page .data-table tbody tr:hover{background:#fff!important}
.prospect-page .data-table{table-layout:fixed!important}
.prospect-page .data-table thead th,.prospect-page .data-table tbody td{padding:7px 10px!important}
.prospect-page .data-table tbody td{height:58px!important;line-height:16px!important}
.prospect-page .data-table th:nth-child(1),.prospect-page .data-table td:nth-child(1){width:48px!important}
.prospect-page .data-table th:nth-child(2),.prospect-page .data-table td:nth-child(2){width:38%!important}
.prospect-page .data-table th:nth-child(3),.prospect-page .data-table td:nth-child(3){width:18%!important}
.prospect-page .data-table th:nth-child(4),.prospect-page .data-table td:nth-child(4){width:18%!important}
.prospect-page .data-table th:nth-child(5),.prospect-page .data-table td:nth-child(5){width:14%!important}
.prospect-page .data-table th:nth-child(6),.prospect-page .data-table td:nth-child(6){width:12%!important}
</style>
