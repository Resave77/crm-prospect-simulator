<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import { useAuthStore } from '../../../stores/auth'
import { useCrmStore } from '../../../stores/crm'
import { requestProspectDeletion } from '../../../api/crm'
import type { Prospect, ProspectStatus } from '../../../types/crm'
import {
  TAB_STAGE_MAP,
  TAB_LABELS,
  TAB_ORDER,
  type ProspectTab,
  pipelineStatusLabel,
  isActiveProspectStatus,
  firstNonEmptyTab,
  EMPTY_MESSAGES,
  PIPELINE_ROUTE_NAME,
} from '../../../utils/prospectPipeline'
import { haversineKm, formatDistance } from '../../../utils/maps'
import { initials } from '../../../utils/format'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')
const searchQuery = ref('')
const activeTab = ref<ProspectTab>('assigned')
const hasUserSelectedTab = ref(false)
const sortBy = ref<'distance' | 'name-asc' | 'name-desc' | 'recently-assigned' | 'last-updated'>('distance')
const userCoords = ref<{ lat: number; lng: number } | null>(null)
const gpsDenied = ref(false)
const showSortMenu = ref(false)

/* ── Filter state ──────────────────────────────────────────── */
const filterStage = ref('')
const filterCategory = ref('')
const filterIndustry = ref('')
const filterDistance = ref('')

const draftStage = ref('')
const draftCategory = ref('')
const draftIndustry = ref('')
const draftDistance = ref('')

const showFilterSheet = ref(false)

/* ── Delete state ─────────────────────────────────────────── */
const showDeleteDialog = ref(false)
const deleteTarget = ref<Prospect | null>(null)
const deleteBusy = ref(false)

function confirmDelete(p: Prospect) {
  deleteTarget.value = p
  showDeleteDialog.value = true
}

async function executeDelete() {
  if (!deleteTarget.value) return
  deleteBusy.value = true
  try {
    await requestProspectDeletion(deleteTarget.value.id)
    const idx = crm.myProspects.findIndex((p) => p.id === deleteTarget.value!.id)
    if (idx >= 0) {
      crm.myProspects[idx] = { ...crm.myProspects[idx], deletionRequested: true }
    }
    showDeleteDialog.value = false
    deleteTarget.value = null
  } catch (e: unknown) {
    error.value = crm.errorMessage(e)
  } finally {
    deleteBusy.value = false
  }
}

type SortOption = { label: string; value: typeof sortBy.value }
const sortOptions: SortOption[] = [
  { label: 'Nearest first', value: 'distance' },
  { label: 'Name A\u2013Z', value: 'name-asc' },
  { label: 'Name Z\u2013A', value: 'name-desc' },
  { label: 'Recently assigned', value: 'recently-assigned' },
  { label: 'Last updated', value: 'last-updated' },
]

function getDistance(p: Prospect): number | null {
  if (p.latitude == null || p.longitude == null || !userCoords.value) return null
  return haversineKm(userCoords.value.lat, userCoords.value.lng, p.latitude, p.longitude)
}

function statusSeverity(status: ProspectStatus): 'info' | 'warn' | 'success' | 'danger' | 'secondary' {
  switch (status) {
    case 'NEW_LEAD': return 'info'
    case 'CONTACTED': return 'info'
    case 'INTERESTED': return 'info'
    case 'QUALIFIED': return 'warn'
    case 'PROPOSAL_SENT': return 'info'
    case 'NEGOTIATION': return 'warn'
    case 'WON': return 'success'
    case 'LOST': return 'danger'
    default: return 'secondary'
  }
}

function tabLabel(tab: ProspectTab): string { return TAB_LABELS[tab] }

function hasCoordinates(p: Prospect): boolean {
  return p.latitude != null && p.longitude != null
}

function openGoogleMaps(p: Prospect) {
  if (p.latitude != null && p.longitude != null) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${p.latitude},${p.longitude}`, '_blank', 'noopener,noreferrer')
  } else if (p.formattedAddress) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(p.formattedAddress)}`, '_blank', 'noopener,noreferrer')
  }
}

function acquireGPS() {
  if (!navigator.geolocation) return
  navigator.geolocation.getCurrentPosition(
    (pos) => { userCoords.value = { lat: pos.coords.latitude, lng: pos.coords.longitude }; gpsDenied.value = false },
    () => { gpsDenied.value = true },
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function matchesDistance(p: Prospect, dist: string): boolean {
  if (!dist) return true
  const d = getDistance(p)
  if (d === null) return false
  if (dist === 'lt1') return d < 1
  if (dist === '1to3') return d >= 1 && d <= 3
  if (dist === '3to5') return d > 3 && d <= 5
  return true
}

function applyFilters(list: Prospect[]): Prospect[] {
  let result = list
  if (filterStage.value) {
    const statuses = TAB_STAGE_MAP[filterStage.value as ProspectTab]
    if (statuses) result = result.filter((p) => statuses.includes(p.status))
  }
  if (filterIndustry.value) result = result.filter((p) => p.industryGroup === filterIndustry.value)
  if (filterCategory.value) result = result.filter((p) => p.placeCategory === filterCategory.value)
  if (filterDistance.value) result = result.filter((p) => matchesDistance(p, filterDistance.value))
  return result
}

function activeFilterCount(): number {
  let n = 0
  if (filterStage.value) n++
  if (filterIndustry.value) n++
  if (filterCategory.value) n++
  if (filterDistance.value) n++
  return n
}

function resetFilters() {
  filterStage.value = ''
  filterIndustry.value = ''
  filterCategory.value = ''
  filterDistance.value = ''
  draftStage.value = ''
  draftIndustry.value = ''
  draftCategory.value = ''
  draftDistance.value = ''
}

/* ── Bottom sheet ──────────────────────────────────────────── */
function openFilterSheet() {
  draftStage.value = filterStage.value
  draftIndustry.value = filterIndustry.value
  draftCategory.value = filterCategory.value
  draftDistance.value = filterDistance.value
  showFilterSheet.value = true
}

function applyDraftFilters() {
  filterStage.value = draftStage.value
  filterIndustry.value = draftIndustry.value
  filterCategory.value = draftCategory.value
  filterDistance.value = draftDistance.value
  showFilterSheet.value = false
}

function clearDraft() {
  draftStage.value = ''
  draftIndustry.value = ''
  draftCategory.value = ''
  draftDistance.value = ''
}

function draftFilterCount(): number {
  let n = 0
  if (draftStage.value) n++
  if (draftIndustry.value) n++
  if (draftCategory.value) n++
  if (draftDistance.value) n++
  return n
}

const draftFilteredCount = computed(() => {
  let list = activeProspects.value
  if (draftStage.value) {
    const statuses = TAB_STAGE_MAP[draftStage.value as ProspectTab]
    if (statuses) list = list.filter((p) => statuses.includes(p.status))
  }
  if (draftIndustry.value) list = list.filter((p) => p.industryGroup === draftIndustry.value)
  if (draftCategory.value) list = list.filter((p) => p.placeCategory === draftCategory.value)
  if (draftDistance.value) list = list.filter((p) => matchesDistance(p, draftDistance.value))
  return list.length
})

function draftCountForStage(stage: string): number {
  let list = activeProspects.value
  if (draftIndustry.value) list = list.filter((p) => p.industryGroup === draftIndustry.value)
  if (draftCategory.value) list = list.filter((p) => p.placeCategory === draftCategory.value)
  if (draftDistance.value) list = list.filter((p) => matchesDistance(p, draftDistance.value))
  if (stage) {
    const statuses = TAB_STAGE_MAP[stage as ProspectTab]
    if (statuses) return list.filter((p) => statuses.includes(p.status)).length
    return 0
  }
  return list.length
}

function draftCountForIndustry(ind: string): number {
  let list = activeProspects.value
  if (draftStage.value) {
    const statuses = TAB_STAGE_MAP[draftStage.value as ProspectTab]
    if (statuses) list = list.filter((p) => statuses.includes(p.status))
  }
  if (draftCategory.value) list = list.filter((p) => p.placeCategory === draftCategory.value)
  if (draftDistance.value) list = list.filter((p) => matchesDistance(p, draftDistance.value))
  if (ind) return list.filter((p) => p.industryGroup === ind).length
  return list.length
}

function draftCountForCategory(cat: string): number {
  let list = activeProspects.value
  if (draftStage.value) {
    const statuses = TAB_STAGE_MAP[draftStage.value as ProspectTab]
    if (statuses) list = list.filter((p) => statuses.includes(p.status))
  }
  if (draftIndustry.value) list = list.filter((p) => p.industryGroup === draftIndustry.value)
  if (draftDistance.value) list = list.filter((p) => matchesDistance(p, draftDistance.value))
  if (cat) return list.filter((p) => p.placeCategory === cat).length
  return list.length
}

function onSheetKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') showFilterSheet.value = false
}

watch(showFilterSheet, (open) => {
  if (open) {
    document.body.style.overflow = 'hidden'
    nextTick(() => document.getElementById('prospect-filter-close')?.focus())
    document.addEventListener('keydown', onSheetKeydown)
  } else {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', onSheetKeydown)
  }
})

onUnmounted(() => { document.body.style.overflow = ''; document.removeEventListener('keydown', onSheetKeydown) })

/* ── Data ──────────────────────────────────────────────────── */
function selectTab(tab: ProspectTab) {
  hasUserSelectedTab.value = true
  activeTab.value = tab
}

const activeProspects = computed(() =>
  crm.myProspects.filter((p) => isActiveProspectStatus(p.status)),
)

const tabCounts = computed(() => {
  const counts = {} as Record<ProspectTab, number>
  for (const tab of TAB_ORDER) {
    counts[tab] = activeProspects.value.filter((p) => TAB_STAGE_MAP[tab].includes(p.status)).length
  }
  return counts
})

const availableIndustries = computed(() =>
  [...new Set(activeProspects.value.map((p) => p.industryGroup).filter(Boolean))].sort(),
)

const availableCategories = computed(() =>
  [...new Set(activeProspects.value.map((p) => p.placeCategory).filter(Boolean))].sort(),
)

const currentTabProspects = computed(() => {
  const statuses = TAB_STAGE_MAP[activeTab.value]
  return activeProspects.value.filter((p) => statuses.includes(p.status))
})

const filteredProspects = computed(() => {
  let list = [...currentTabProspects.value]
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((p) => {
      const hay = `${p.placeName} ${p.formattedAddress} ${p.industryGroup} ${p.placeCategory} ${p.phoneNumber} ${p.assignedSalesExecutive}`.toLowerCase()
      return hay.includes(q)
    })
  }
  list = applyFilters(list)
  return list
})

const displayedProspects = computed(() => {
  let list = [...filteredProspects.value]
  if (sortBy.value === 'name-asc') list.sort((a, b) => a.placeName.localeCompare(b.placeName))
  else if (sortBy.value === 'name-desc') list.sort((a, b) => b.placeName.localeCompare(a.placeName))
  else if (sortBy.value === 'recently-assigned') list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  else if (sortBy.value === 'last-updated') list.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
  else if (sortBy.value === 'distance') {
    list.sort((a, b) => {
      const da = getDistance(a)
      const db = getDistance(b)
      if (da === null && db === null) return 0
      if (da === null) return 1
      if (db === null) return -1
      return da - db
    })
  }
  return list
})

const emptyMsg = computed(() => EMPTY_MESSAGES[activeTab.value])

onMounted(async () => {
  acquireGPS()
  try {
    await crm.loadMyProspects()
    if (!hasUserSelectedTab.value) {
      activeTab.value = firstNonEmptyTab(tabCounts.value)
    }
  } catch (e: unknown) { error.value = crm.errorMessage(e) }
})
</script>

<template>
  <section class="mp-page">
    <RouterLink class="mp-back" to="/sales/dashboard"><i class="pi pi-arrow-left" /></RouterLink>
    <!-- 1. Header -->
    <div class="mp-header">
      <div class="mp-header-left">
        <span class="mp-avatar">{{ auth.user?.fullName?.slice(0, 1) }}</span>
        <div class="mp-header-text">
          <strong>My prospects</strong>
          <small>{{ activeProspects.length }} active &middot; {{ displayedProspects.length }} shown</small>
        </div>
      </div>
      <div class="mp-header-actions">
        <RouterLink class="mp-pipeline-btn" :to="{ name: PIPELINE_ROUTE_NAME }">
          <i class="pi pi-chart-bar" />
          <span>Sales Pipeline</span>
        </RouterLink>
        <button class="mp-header-action" @click="openFilterSheet" aria-label="Open filters">
          <i class="pi pi-sliders-h" />
          <span v-if="activeFilterCount()" class="mp-notif-dot">{{ activeFilterCount() }}</span>
        </button>
      </div>
    </div>

    <!-- 2. Pipeline summary mini card -->
    <RouterLink class="mp-pipeline-summary" :to="{ name: PIPELINE_ROUTE_NAME }">
      <div class="mp-pipeline-summary-label"><i class="pi pi-chart-bar" /> Sales Pipeline</div>
      <div class="mp-pipeline-summary-counts">
        <span v-for="tab in TAB_ORDER" :key="tab" class="mp-ps-item">
          {{ tabLabel(tab) }} <strong>{{ tabCounts[tab] }}</strong>
        </span>
      </div>
    </RouterLink>

    <!-- 3. Search -->
    <div class="mp-search">
      <i class="pi pi-search" />
      <input v-model="searchQuery" placeholder="Search business, category or area" aria-label="Search prospects" />
      <button v-if="searchQuery" class="mp-search-clear" aria-label="Clear search" @click="searchQuery = ''">
        <i class="pi pi-times" />
      </button>
    </div>

    <!-- 4. Tabs -->
    <div class="mp-tabs">
      <button v-for="tab in TAB_ORDER" :key="tab" class="mp-tab" :class="{ active: activeTab === tab }" @click="selectTab(tab)">
        {{ tabLabel(tab) }} <span class="mp-tab-badge">{{ tabCounts[tab] }}</span>
      </button>
    </div>

    <!-- Messages -->
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <Message v-if="gpsDenied" severity="warn" :closable="false">
      Location access was denied. Distance sorting requires GPS permission.
    </Message>

    <!-- Loading skeleton -->
    <div v-if="crm.loading && !activeProspects.length" class="mp-skeleton-list">
      <div v-for="n in 4" :key="n" class="mp-skeleton-card">
        <div class="sk-row"><div class="sk-avatar" /><div class="sk-lines"><div class="sk-line w60" /><div class="sk-line w40" /></div></div>
        <div class="sk-line w80" /><div class="sk-line w50" />
        <div class="sk-row sk-actions"><div class="sk-btn" /><div class="sk-btn" /><div class="sk-btn" /></div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="!displayedProspects.length && !error" class="mp-empty">
      <div class="mp-empty-icon"><i class="pi pi-briefcase" /></div>
      <template v-if="searchQuery || activeFilterCount()">
        <strong>No prospects found</strong>
        <span>Try changing your search or filters.</span>
        <button class="mp-reset-btn" @click="searchQuery = ''; resetFilters()">
          <i class="pi pi-refresh" /> Reset filters
        </button>
      </template>
      <template v-else>
        <strong>{{ emptyMsg.title }}</strong>
        <span>{{ emptyMsg.subtitle }}</span>
        <RouterLink class="mp-reset-btn" :to="{ name: PIPELINE_ROUTE_NAME }">
          <i class="pi pi-chart-bar" /> Open Sales Pipeline
        </RouterLink>
      </template>
    </div>

    <!-- Sort + Cards -->
    <template v-else>
      <div class="mp-section-header">
        <strong>{{ tabLabel(activeTab) }} prospects</strong>
        <button class="mp-sort-trigger" @click="showSortMenu = !showSortMenu">
          <i class="pi pi-sort-alt" /> {{ sortOptions.find((s) => s.value === sortBy)?.label }}
          <i class="pi pi-chevron-down" />
        </button>
      </div>

      <div v-if="showSortMenu" class="mp-sort-menu">
        <button v-for="opt in sortOptions" :key="opt.value" :class="{ active: sortBy === opt.value }" @click="sortBy = opt.value; showSortMenu = false">
          {{ opt.label }}
        </button>
      </div>

      <div class="mp-card-list">
        <article v-for="prospect in displayedProspects" :key="prospect.id" class="mp-card">
          <div class="mp-card-top">
            <span class="mp-card-avatar">{{ initials(prospect.placeName || 'Unnamed') }}</span>
            <div class="mp-card-identity">
              <strong>{{ prospect.placeName || 'Unnamed prospect' }}</strong>
              <small>{{ prospect.industryGroup || prospect.placeCategory || 'Uncategorized' }}</small>
            </div>
            <button
              v-if="prospect.deletionRequested"
              class="mp-card-delete mp-card-delete-pending"
              title="Deletion requested — awaiting admin approval"
              disabled
            >
              <i class="pi pi-clock" />
            </button>
            <button
              v-else
              class="mp-card-delete"
              title="Request deletion"
              @click.stop="confirmDelete(prospect)"
            >
              <i class="pi pi-trash" />
            </button>
            <span v-if="getDistance(prospect) !== null" class="mp-distance-pill">
              <i class="pi pi-map-marker" /> {{ formatDistance(getDistance(prospect)!) }}
            </span>
          </div>

          <div class="mp-card-middle">
            <div class="mp-card-address">
              <i class="pi pi-map-marker" />
              <span>{{ prospect.formattedAddress || 'No address' }}</span>
            </div>
            <div class="mp-card-tags">
              <Tag :value="pipelineStatusLabel(prospect.status)" :severity="statusSeverity(prospect.status)" />
              <span v-if="prospect.deletionRequested" class="mp-deletion-badge">Deletion Requested</span>
              <span v-if="prospect.status === 'WON'" class="mp-review-tag">Waiting for Admin Review</span>
            </div>
          </div>

          <!-- Assigned: Navigate, Detail, Visit -->
          <div v-if="activeTab === 'assigned'" class="mp-card-actions">
            <button class="mp-action-btn mp-action-navigate" :disabled="!hasCoordinates(prospect) && !prospect.formattedAddress" @click="openGoogleMaps(prospect)">
              <i class="pi pi-directions" /> Navigate
            </button>
            <RouterLink class="mp-action-btn mp-action-detail" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-eye" /> Detail
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-visit" :to="`/sales/my-prospects/${prospect.id}/check-in`">
              <i class="pi pi-sign-in" /> Visit
            </RouterLink>
          </div>

          <!-- Visited: Detail, Update Pipeline, Visit -->
          <div v-else-if="activeTab === 'visited'" class="mp-card-actions">
            <RouterLink class="mp-action-btn mp-action-detail" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-eye" /> Detail
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-pipeline" :to="{ name: PIPELINE_ROUTE_NAME, query: { prospectId: prospect.id, action: 'update' } }">
              <i class="pi pi-chart-bar" /> Pipeline
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-visit" :to="`/sales/my-prospects/${prospect.id}/check-in`">
              <i class="pi pi-sign-in" /> Visit
            </RouterLink>
          </div>

          <!-- Follow Up: Detail, Update Pipeline, Visit -->
          <div v-else-if="activeTab === 'followup'" class="mp-card-actions">
            <RouterLink class="mp-action-btn mp-action-detail" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-eye" /> Detail
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-pipeline" :to="{ name: PIPELINE_ROUTE_NAME, query: { prospectId: prospect.id, action: 'update' } }">
              <i class="pi pi-chart-bar" /> Pipeline
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-visit" :to="`/sales/my-prospects/${prospect.id}/check-in`">
              <i class="pi pi-sign-in" /> Visit
            </RouterLink>
          </div>

          <!-- Won: Detail, View Pipeline -->
          <div v-else-if="activeTab === 'won'" class="mp-card-actions">
            <RouterLink class="mp-action-btn mp-action-detail" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-eye" /> Detail
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-pipeline" :to="{ name: PIPELINE_ROUTE_NAME, query: { prospectId: prospect.id, action: 'update' } }">
              <i class="pi pi-chart-bar" /> Pipeline
            </RouterLink>
            <button class="mp-action-btn mp-action-navigate" :disabled="!hasCoordinates(prospect) && !prospect.formattedAddress" @click="openGoogleMaps(prospect)">
              <i class="pi pi-directions" /> Navigate
            </button>
          </div>
        </article>
      </div>
    </template>

    <!-- Request Deletion Confirmation -->
    <Dialog v-model:visible="showDeleteDialog" modal header="Request Deletion" :style="{ width: 'min(100%, 400px)' }" :closable="!deleteBusy">
      <p v-if="deleteTarget" style="margin:0;font-size:0.85rem;line-height:1.5;">
        Request deletion of <strong>{{ deleteTarget.placeName }}</strong>?
      </p>
      <p style="margin:0.5rem 0 0;font-size:0.78rem;color:var(--text-muted);">Admin will review and approve or reject this request.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined :disabled="deleteBusy" @click="showDeleteDialog = false" />
        <Button label="Request Deletion" icon="pi pi-trash" severity="danger" :loading="deleteBusy" @click="executeDelete" />
      </template>
    </Dialog>

    <!-- FAB filter -->
    <button class="mp-fab" @click="openFilterSheet" aria-label="Open filter panel">
      <i class="pi pi-filter" />
      <span v-if="activeFilterCount()" class="mp-fab-badge">{{ activeFilterCount() }}</span>
    </button>

    <!-- Filter Bottom Sheet -->
    <Teleport to="body">
      <div v-if="showFilterSheet" class="mp-sheet-overlay" @click.self="showFilterSheet = false" />
      <div v-if="showFilterSheet" class="mp-sheet" role="dialog" aria-labelledby="prospect-filter-title">
        <div class="mp-sheet-handle" />

        <div class="mp-sheet-header">
          <div class="mp-sheet-header-text">
            <strong id="prospect-filter-title">Filter prospects</strong>
            <small>Refine prospects by pipeline, category, and distance</small>
          </div>
          <div class="mp-sheet-header-actions">
            <button class="mp-sheet-clear" @click="clearDraft">Clear all</button>
            <button id="prospect-filter-close" class="mp-sheet-close" aria-label="Close filters" @click="showFilterSheet = false">
              <i class="pi pi-times" />
            </button>
          </div>
        </div>

        <div class="mp-sheet-body">
          <!-- Pipeline Stage -->
          <div class="mp-sheet-group">
            <span class="mp-sheet-group-label">Pipeline Stage</span>
            <div class="mp-sheet-chips">
              <button class="mp-chip" :class="{ active: !draftStage }" :aria-pressed="!draftStage" @click="draftStage = ''">
                All stages <span class="mp-chip-count">{{ draftCountForStage('') }}</span>
              </button>
              <button v-for="tab in TAB_ORDER" :key="tab" class="mp-chip" :class="{ active: draftStage === tab }" :aria-pressed="draftStage === tab" @click="draftStage = draftStage === tab ? '' : tab">
                {{ tabLabel(tab) }} <span class="mp-chip-count">{{ draftCountForStage(draftStage === tab ? '' : tab) }}</span>
              </button>
            </div>
          </div>

          <!-- Industry -->
          <div v-if="availableIndustries.length" class="mp-sheet-group">
            <span class="mp-sheet-group-label">Industry</span>
            <div class="mp-sheet-chips">
              <button class="mp-chip" :class="{ active: !draftIndustry }" :aria-pressed="!draftIndustry" @click="draftIndustry = ''">
                All industries <span class="mp-chip-count">{{ draftCountForIndustry('') }}</span>
              </button>
              <button v-for="ind in availableIndustries" :key="ind" class="mp-chip" :class="{ active: draftIndustry === ind }" :aria-pressed="draftIndustry === ind" @click="draftIndustry = draftIndustry === ind ? '' : ind">
                {{ ind }} <span class="mp-chip-count">{{ draftCountForIndustry(draftIndustry === ind ? '' : ind) }}</span>
              </button>
            </div>
          </div>

          <!-- Category -->
          <div v-if="availableCategories.length" class="mp-sheet-group">
            <span class="mp-sheet-group-label">Category</span>
            <div class="mp-sheet-chips">
              <button class="mp-chip" :class="{ active: !draftCategory }" :aria-pressed="!draftCategory" @click="draftCategory = ''">
                All categories <span class="mp-chip-count">{{ draftCountForCategory('') }}</span>
              </button>
              <button v-for="cat in availableCategories" :key="cat" class="mp-chip" :class="{ active: draftCategory === cat }" :aria-pressed="draftCategory === cat" @click="draftCategory = draftCategory === cat ? '' : cat">
                {{ cat }} <span class="mp-chip-count">{{ draftCountForCategory(draftCategory === cat ? '' : cat) }}</span>
              </button>
            </div>
          </div>

          <!-- Distance -->
          <div v-if="userCoords" class="mp-sheet-group">
            <span class="mp-sheet-group-label">Distance from you</span>
            <div class="mp-sheet-chips">
              <button class="mp-chip" :class="{ active: !draftDistance }" @click="draftDistance = ''">
                Any
              </button>
              <button class="mp-chip" :class="{ active: draftDistance === 'lt1' }" @click="draftDistance = draftDistance === 'lt1' ? '' : 'lt1'">
                &lt; 1 km
              </button>
              <button class="mp-chip" :class="{ active: draftDistance === '1to3' }" @click="draftDistance = draftDistance === '1to3' ? '' : '1to3'">
                1&ndash;3 km
              </button>
              <button class="mp-chip" :class="{ active: draftDistance === '3to5' }" @click="draftDistance = draftDistance === '3to5' ? '' : '3to5'">
                3&ndash;5 km
              </button>
            </div>
          </div>

          <div v-if="!userCoords" class="mp-sheet-helper">
            <i class="pi pi-info-circle" /> Enable location to filter prospects by distance.
          </div>

          <div v-if="!availableIndustries.length && !availableCategories.length" class="mp-sheet-empty">
            <span>No filter options available yet.</span>
          </div>
        </div>

        <div class="mp-sheet-footer">
          <button class="mp-sheet-reset" @click="clearDraft">Reset</button>
          <button class="mp-sheet-apply" @click="applyDraftFilters">
            Show {{ draftFilteredCount }} prospect{{ draftFilteredCount !== 1 ? 's' : '' }}
          </button>
        </div>
      </div>
    </Teleport>
  </section>
</template>

<style scoped>
/* ── Page ───────────────────────────────────────────────────── */
.mp-page { display: flex; flex-direction: column; gap: 0.85rem; padding-bottom: 1.5rem; }
.mp-back { display: inline-flex; align-items: center; justify-content: center; width: 2rem; height: 2rem; color: var(--brand-blue); background: var(--brand-blue-bg); border: 1px solid transparent; border-radius: var(--radius-md); text-decoration: none; font-size: 0.9rem; transition: background var(--transition-fast), border-color var(--transition-fast); }
.mp-back:hover { background: #dbeafe; border-color: var(--brand-blue); }

/* ── 1. Header ─────────────────────────────────────────────── */
.mp-header { display: flex; align-items: center; justify-content: space-between; padding: 0.15rem 0; }
.mp-header-left { display: flex; align-items: center; gap: 0.7rem; }
.mp-header-actions { display: flex; align-items: center; gap: 0.5rem; }
.mp-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 50%; background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff; font-weight: 800; font-size: 0.85rem; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.25);
}
.mp-header-text { display: flex; flex-direction: column; gap: 0.05rem; }
.mp-header-text strong { font-size: 1rem; font-weight: 800; color: #0f172a; }
.mp-header-text small { color: #64748b; font-size: 0.7rem; font-weight: 500; }
.mp-pipeline-btn {
  display: inline-flex; align-items: center; gap: 0.35rem;
  padding: 0.4rem 0.75rem; border-radius: 9999px;
  background: #eff6ff; color: #2563eb; border: 1px solid #bfdbfe;
  font-size: 0.7rem; font-weight: 700; text-decoration: none; white-space: nowrap;
  transition: all 0.15s ease;
}
.mp-pipeline-btn:hover { background: #dbeafe; border-color: #93c5fd; }
.mp-pipeline-btn i { font-size: 0.7rem; }
.mp-header-action {
  position: relative; width: 36px; height: 36px; display: grid; place-items: center;
  border-radius: 50%; border: 1px solid #e2e8f0; background: #fff; color: #64748b;
  cursor: pointer; font-size: 0.9rem; transition: all 0.15s ease;
}
.mp-header-action:hover { color: #2563eb; border-color: #cbd5e1; background: #eff6ff; }
.mp-notif-dot {
  position: absolute; top: -3px; right: -3px; min-width: 16px; height: 16px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px; border: 2px solid #fff;
}

/* ── 2. Pipeline summary ────────────────────────────────────── */
.mp-pipeline-summary {
  display: flex; flex-direction: column; gap: 0.5rem;
  padding: 0.75rem 1rem; background: #f0f7ff; border: 1px solid #bfdbfe;
  border-radius: 14px; text-decoration: none; color: inherit; transition: all 0.15s ease;
}
.mp-pipeline-summary:hover { background: #e0efff; border-color: #93c5fd; }
.mp-pipeline-summary-label {
  display: flex; align-items: center; gap: 0.35rem;
  font-size: 0.72rem; font-weight: 700; color: #2563eb;
}
.mp-pipeline-summary-label i { font-size: 0.7rem; }
.mp-pipeline-summary-counts { display: flex; gap: 0.75rem; flex-wrap: wrap; }
.mp-ps-item { font-size: 0.65rem; color: #64748b; }
.mp-ps-item strong { color: #0f172a; font-weight: 700; }

/* ── 3. Search ─────────────────────────────────────────────── */
.mp-search {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.55rem 0.8rem; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 14px; transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mp-search:focus-within { border-color: #2563eb; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.mp-search i { color: #94a3b8; font-size: 0.85rem; flex-shrink: 0; }
.mp-search input {
  flex: 1; border: 0; outline: 0; background: transparent; color: #0f172a;
  font-size: 0.8rem; font-weight: 500;
}
.mp-search input::placeholder { color: #94a3b8; }
.mp-search-clear {
  width: 22px; height: 22px; display: grid; place-items: center;
  border-radius: 50%; border: 0; background: #f1f5f9; color: #64748b;
  cursor: pointer; font-size: 0.65rem; transition: background 0.15s ease;
}
.mp-search-clear:hover { background: #e2e8f0; }

/* ── 4. Tabs ───────────────────────────────────────────────── */
.mp-tabs { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.35rem; }
.mp-tab {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.5rem 0; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-tab.active { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; font-weight: 700; }
.mp-tab:not(.active):hover { background: #f8fafc; border-color: #cbd5e1; }
.mp-tab-badge {
  min-width: 16px; height: 16px; display: inline-grid; place-items: center;
  border-radius: 9999px; background: #f1f5f9; font-size: 0.55rem; font-weight: 700; padding: 0 3px;
}
.mp-tab.active .mp-tab-badge { background: #dbeafe; color: #2563eb; }

/* ── 5. Section header + sort ──────────────────────────────── */
.mp-section-header { display: flex; align-items: center; justify-content: space-between; }
.mp-section-header strong { font-size: 0.82rem; font-weight: 800; color: #0f172a; }
.mp-sort-trigger {
  display: flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.6rem;
  border-radius: 10px; border: 1px solid #e2e8f0; background: #fff;
  color: #64748b; font-size: 0.68rem; font-weight: 600; cursor: pointer; transition: all 0.15s ease;
}
.mp-sort-trigger:hover { border-color: #cbd5e1; background: #f8fafc; }
.mp-sort-trigger i { font-size: 0.6rem; }
.mp-sort-menu { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.mp-sort-menu button {
  padding: 0.4rem 0.7rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-sort-menu button.active { background: #2563eb; color: #fff; border-color: #2563eb; }
.mp-sort-menu button:not(.active):hover { border-color: #cbd5e1; background: #f8fafc; }

/* ── 6. Card list ──────────────────────────────────────────── */
.mp-card-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mp-card {
  padding: 0.9rem 1rem; background: #fff; border: 1px solid #eef1f6;
  border-radius: 16px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
  display: flex; flex-direction: column; gap: 0.6rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mp-card:hover { border-color: #d6dce6; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }

/* Card top */
.mp-card-top { display: flex; align-items: center; gap: 0.65rem; }
.mp-card-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 12px; background: #eff6ff; color: #2563eb;
  font-weight: 800; font-size: 0.72rem; flex-shrink: 0;
}
.mp-card-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.05rem; }
.mp-card-identity strong { font-size: 0.82rem; font-weight: 700; color: #0f172a; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mp-card-identity small { color: #64748b; font-size: 0.65rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mp-distance-pill {
  display: flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.5rem;
  border-radius: 9999px; background: #eff6ff; color: #2563eb;
  font-size: 0.6rem; font-weight: 700; white-space: nowrap; flex-shrink: 0;
}
.mp-distance-pill i { font-size: 0.55rem; }

.mp-card-delete {
  width: 28px; height: 28px; display: grid; place-items: center;
  border-radius: 8px; border: 1px solid #fee2e2; background: #fff;
  color: #dc2626; font-size: 0.72rem; cursor: pointer; flex-shrink: 0;
  transition: all 0.15s ease;
}
.mp-card-delete:hover { background: #fef2f2; border-color: #fca5a5; color: #b91c1c; }
.mp-card-delete-pending {
  border-color: #fde68a; background: #fffbeb; color: #d97706; cursor: default;
}
.mp-card-delete-pending:hover { background: #fffbeb; border-color: #fde68a; color: #d97706; }

.mp-deletion-badge {
  display: inline-block; padding: 0.15rem 0.5rem; border-radius: 9999px;
  background: #fef3c7; color: #92400e; font-size: 0.58rem; font-weight: 600;
}

/* Card middle */
.mp-card-middle { display: flex; flex-direction: column; gap: 0.35rem; }
.mp-card-address { display: flex; align-items: flex-start; gap: 0.4rem; color: #64748b; font-size: 0.72rem; line-height: 1.4; }
.mp-card-address i { margin-top: 0.1rem; font-size: 0.68rem; flex-shrink: 0; color: #94a3b8; }
.mp-card-tags { display: flex; flex-wrap: wrap; gap: 0.3rem; align-items: center; }
.mp-review-tag {
  display: inline-block; padding: 0.15rem 0.5rem; border-radius: 9999px;
  background: #fef3c7; color: #92400e; font-size: 0.58rem; font-weight: 600;
}

/* Card actions */
.mp-card-actions {
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.4rem;
  padding-top: 0.45rem; border-top: 1px solid #f1f5f9;
}
.mp-action-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.45rem 0; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; font-size: 0.65rem; font-weight: 600; cursor: pointer;
  text-decoration: none; transition: all 0.15s ease;
}
.mp-action-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.mp-action-navigate { color: #2563eb; }
.mp-action-navigate:not(:disabled):hover { background: #eff6ff; border-color: #bfdbfe; }
.mp-action-detail { color: #0f172a; }
.mp-action-detail:hover { background: #f8fafc; border-color: #cbd5e1; }
.mp-action-visit { color: #059669; }
.mp-action-visit:hover { background: #ecfdf5; border-color: #a7f3d0; }
.mp-action-pipeline { color: #7c3aed; }
.mp-action-pipeline:hover { background: #f5f3ff; border-color: #ddd6fe; }

/* ── Empty state ───────────────────────────────────────────── */
.mp-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  padding: 2.5rem 1rem; text-align: center;
}
.mp-empty-icon {
  width: 52px; height: 52px; display: grid; place-items: center;
  border-radius: 16px; background: #eff6ff; color: #2563eb; font-size: 1.4rem;
}
.mp-empty strong { color: #0f172a; font-size: 0.88rem; }
.mp-empty span { color: #64748b; font-size: 0.75rem; line-height: 1.5; }
.mp-reset-btn {
  display: inline-flex; align-items: center; gap: 0.3rem; margin-top: 0.3rem;
  padding: 0.45rem 0.85rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #2563eb; font-size: 0.72rem; font-weight: 600;
  cursor: pointer; text-decoration: none; transition: all 0.15s ease;
}
.mp-reset-btn:hover { background: #eff6ff; border-color: #bfdbfe; }

/* ── Skeleton ──────────────────────────────────────────────── */
.mp-skeleton-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mp-skeleton-card {
  padding: 1rem; background: #fff; border: 1px solid #eef1f6; border-radius: 16px;
  display: flex; flex-direction: column; gap: 0.6rem;
  animation: mp-pulse 1.5s ease-in-out infinite;
}
.sk-row { display: flex; align-items: center; gap: 0.65rem; }
.sk-avatar { width: 38px; height: 38px; border-radius: 12px; background: #e2e8f0; flex-shrink: 0; }
.sk-lines { flex: 1; display: flex; flex-direction: column; gap: 0.3rem; }
.sk-line { height: 10px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w60 { width: 60%; }
.sk-line.w40 { width: 40%; }
.sk-line.w80 { width: 80%; }
.sk-line.w50 { width: 50%; }
.sk-actions { justify-content: flex-end; }
.sk-btn { width: 70px; height: 28px; border-radius: 8px; background: #e2e8f0; }
@keyframes mp-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

/* ── FAB ───────────────────────────────────────────────────── */
.mp-fab {
  position: fixed; bottom: calc(80px + env(safe-area-inset-bottom, 0px));
  right: 1rem;
  width: 48px; height: 48px; border-radius: 50%; border: 0;
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff;
  font-size: 1.1rem; cursor: pointer; z-index: 50;
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.35);
  display: grid; place-items: center; transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.mp-fab:hover { transform: scale(1.08); box-shadow: 0 8px 24px rgba(37, 99, 235, 0.45); }
.mp-fab-badge {
  position: absolute; top: -2px; right: -2px; min-width: 18px; height: 18px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px; border: 2px solid #fff;
}

/* ── Filter Bottom Sheet ───────────────────────────────────── */
.mp-sheet-overlay {
  position: fixed; inset: 0; background: rgba(15, 23, 42, 0.45);
  z-index: 200; animation: mp-fade-in 0.2s ease;
}
.mp-sheet {
  position: fixed; left: 50%; bottom: 0;
  width: 100%; max-width: 480px; max-height: 82dvh;
  transform: translateX(-50%);
  background: #fff; border-radius: 20px 20px 0 0;
  box-shadow: 0 -8px 32px rgba(15, 23, 42, 0.12);
  z-index: 201; display: flex; flex-direction: column;
  animation: mp-sheet-up 0.25s ease;
}
.mp-sheet-handle {
  width: 36px; height: 4px; border-radius: 2px;
  background: #d1d5db; margin: 0.65rem auto 0; flex-shrink: 0;
}
.mp-sheet-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  padding: 0.75rem 1.25rem 0.5rem; flex-shrink: 0;
}
.mp-sheet-header-text { display: flex; flex-direction: column; gap: 0.15rem; }
.mp-sheet-header-text strong { font-size: 0.95rem; font-weight: 800; color: #0f172a; }
.mp-sheet-header-text small { font-size: 0.72rem; color: #64748b; }
.mp-sheet-header-actions { display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
.mp-sheet-clear {
  padding: 0.3rem 0.6rem; border-radius: 8px; border: 0;
  background: transparent; color: #2563eb; font-size: 0.72rem; font-weight: 700;
  cursor: pointer; transition: background 0.15s ease;
}
.mp-sheet-clear:hover { background: #eff6ff; }
.mp-sheet-close {
  width: 32px; height: 32px; display: grid; place-items: center;
  border-radius: 50%; border: 0; background: #f1f5f9; color: #64748b;
  cursor: pointer; font-size: 0.85rem; transition: all 0.15s ease;
}
.mp-sheet-close:hover { background: #e2e8f0; color: #0f172a; }
.mp-sheet-body {
  flex: 1; overflow-y: auto; padding: 0.25rem 1.25rem 1rem;
  -webkit-overflow-scrolling: touch;
}
.mp-sheet-group { margin-bottom: 1rem; }
.mp-sheet-group-label {
  display: block; margin-bottom: 0.5rem;
  font-size: 0.7rem; font-weight: 700; color: #64748b;
  text-transform: uppercase; letter-spacing: 0.04em;
}
.mp-sheet-chips { display: flex; flex-wrap: wrap; gap: 0.4rem; }
.mp-chip {
  display: inline-flex; align-items: center; gap: 0.3rem;
  padding: 0.4rem 0.75rem; border-radius: 9999px;
  border: 1px solid #e2e8f0; background: #fff; color: #475569;
  font-size: 0.72rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s ease;
}
.mp-chip:hover { border-color: #cbd5e1; background: #f8fafc; }
.mp-chip.active {
  background: #2563eb; border-color: #2563eb; color: #fff;
}
.mp-chip.active:hover { background: #1d4ed8; }
.mp-chip-count { font-size: 0.6rem; font-weight: 700; opacity: 0.7; }
.mp-chip.active .mp-chip-count { opacity: 1; }
.mp-sheet-empty {
  padding: 1.5rem 0; text-align: center; color: #94a3b8; font-size: 0.8rem;
}
.mp-sheet-helper {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.75rem 1rem; background: #fffbeb; border: 1px solid #fde68a;
  border-radius: 12px; color: #92400e; font-size: 0.72rem; font-weight: 500;
}
.mp-sheet-helper i { font-size: 0.85rem; color: #d97706; flex-shrink: 0; }
.mp-sheet-footer {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.75rem 1.25rem calc(0.75rem + env(safe-area-inset-bottom, 0px));
  border-top: 1px solid #f1f5f9; flex-shrink: 0;
}
.mp-sheet-reset {
  padding: 0.65rem 1rem; border-radius: 12px; border: 1px solid #e2e8f0;
  background: #fff; color: #475569; font-size: 0.78rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-sheet-reset:hover { background: #f8fafc; border-color: #cbd5e1; }
.mp-sheet-apply {
  flex: 1; padding: 0.65rem 1rem; border-radius: 12px; border: 0;
  background: #2563eb; color: #fff; font-size: 0.78rem; font-weight: 700;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-sheet-apply:hover { background: #1d4ed8; }

@keyframes mp-fade-in { from { opacity: 0; } to { opacity: 1; } }
@keyframes mp-sheet-up { from { transform: translateX(-50%) translateY(100%); } to { transform: translateX(-50%) translateY(0); } }

/* ── Responsive ────────────────────────────────────────────── */
@media (max-width: 767px) {
  .mp-pipeline-btn span { display: none; }
  .mp-pipeline-btn { padding: 0.4rem; }
}
</style>
