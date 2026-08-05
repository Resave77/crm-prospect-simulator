<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Textarea from 'primevue/textarea'
import { BOARD_STATUSES, PIPELINE_STAGES, nextStage, previousStage } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus } from '../../../types/crm'
import PipelineProspectCard from '../../../components/sales/pipeline/PipelineProspectCard.vue'
import { stageLabel } from '../../../components/sales/pipeline/stageColors'

type PipelineGroup = 'ALL' | 'NEW_LEAD' | 'IN_PROGRESS' | 'WON' | 'LOST'
type SecondaryStage = 'ALL_PROGRESS' | ProspectStatus

const GROUP_LABELS: Record<PipelineGroup, string> = {
  ALL: 'All', NEW_LEAD: 'New Lead', IN_PROGRESS: 'In Progress', WON: 'Won', LOST: 'Lost',
}

const IN_PROGRESS_STAGES: ProspectStatus[] = ['CONTACTED', 'INTERESTED', 'QUALIFIED', 'PROPOSAL_SENT', 'NEGOTIATION']

const SECONDARY_OPTIONS: { value: SecondaryStage; label: string }[] = [
  { value: 'ALL_PROGRESS', label: 'All Progress' },
  ...IN_PROGRESS_STAGES.map(s => ({ value: s, label: stageLabel(s) })),
]

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const toast = useToast()

const selected = ref<Prospect | null>(null)
const target = ref<ProspectStatus>('CONTACTED')
const notes = ref('')
const error = ref('')
const success = ref('')
const successTarget = ref<ProspectStatus | null>(null)
const successGroup = ref<PipelineGroup | null>(null)
const highlightId = ref<string | null>(null)

const activeGroup = ref<PipelineGroup>('ALL')
const secondaryStage = ref<SecondaryStage>('ALL_PROGRESS')
const searchQuery = ref('')
const sortOption = ref('updatedAt-desc')
const currentPage = ref(1)
const pageSize = ref(15)
const showFilters = ref(false)
const isDesktop = ref(window.innerWidth >= 1024)

const draftFilters = ref({ industryGroup: '', visitStatus: '' })
const appliedFilters = ref({ industryGroup: '', visitStatus: '' })

window.addEventListener('resize', () => {
  isDesktop.value = window.innerWidth >= 1024
})

const prospects = computed(() => crm.myProspects.filter(item => item.status !== 'CONVERTED'))

const countsByStatus = computed(() => {
  const map = new Map<ProspectStatus, number>()
  for (const s of BOARD_STATUSES) map.set(s, 0)
  for (const p of prospects.value) {
    map.set(p.status, (map.get(p.status) ?? 0) + 1)
  }
  return map
})

const groupCounts = computed(() => {
  const all = prospects.value.length
  const nl = countsByStatus.value.get('NEW_LEAD') ?? 0
  const ip = IN_PROGRESS_STAGES.reduce((sum, s) => sum + (countsByStatus.value.get(s) ?? 0), 0)
  const won = countsByStatus.value.get('WON') ?? 0
  const lost = countsByStatus.value.get('LOST') ?? 0
  return { ALL: all, NEW_LEAD: nl, IN_PROGRESS: ip, WON: won, LOST: lost } as Record<PipelineGroup, number>
})

const secondaryCounts = computed(() => {
  const result: Record<string, number> = { ALL_PROGRESS: 0 }
  for (const s of IN_PROGRESS_STAGES) {
    const count = countsByStatus.value.get(s) ?? 0
    result[s] = count
    result.ALL_PROGRESS += count
  }
  return result
})

const groupFiltered = computed(() => {
  switch (activeGroup.value) {
    case 'NEW_LEAD': return prospects.value.filter(p => p.status === 'NEW_LEAD')
    case 'IN_PROGRESS': return prospects.value.filter(p => IN_PROGRESS_STAGES.includes(p.status))
    case 'WON': return prospects.value.filter(p => p.status === 'WON')
    case 'LOST': return prospects.value.filter(p => p.status === 'LOST')
    default: return prospects.value
  }
})

const secondaryFiltered = computed(() => {
  if (activeGroup.value !== 'IN_PROGRESS' || secondaryStage.value === 'ALL_PROGRESS') return groupFiltered.value
  return groupFiltered.value.filter(p => p.status === secondaryStage.value)
})

const searchLower = computed(() => searchQuery.value.trim().toLowerCase())

const searchedFiltered = computed(() => {
  if (!searchLower.value) return secondaryFiltered.value
  return secondaryFiltered.value.filter(p =>
    `${p.placeName} ${p.industryGroup} ${p.formattedAddress} ${p.assignedSalesExecutive} ${stageLabel(p.status)} ${p.placeCategory || ''}`
      .toLowerCase().includes(searchLower.value)
  )
})

const additionalFiltered = computed(() => {
  const f = appliedFilters.value
  if (!f.industryGroup && !f.visitStatus) return searchedFiltered.value
  return searchedFiltered.value.filter(p => {
    if (f.industryGroup && p.industryGroup !== f.industryGroup) return false
    return true
  })
})

const sortedFiltered = computed(() => {
  const list = [...additionalFiltered.value]
  switch (sortOption.value) {
    case 'name-asc': return list.sort((a, b) => a.placeName.localeCompare(b.placeName))
    case 'name-desc': return list.sort((a, b) => b.placeName.localeCompare(a.placeName))
    case 'stage-order': return list.sort((a, b) => BOARD_STATUSES.indexOf(a.status) - BOARD_STATUSES.indexOf(b.status))
    case 'updatedAt-asc': return list.sort((a, b) => new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime())
    default: return list.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
  }
})

const totalFiltered = computed(() => sortedFiltered.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(totalFiltered.value / pageSize.value)))
const pageStart = computed(() => (currentPage.value - 1) * pageSize.value + 1)
const pageEnd = computed(() => Math.min(currentPage.value * pageSize.value, totalFiltered.value))
const limitReached = computed(() => pageEnd.value >= totalFiltered.value)

const paginatedProspects = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return sortedFiltered.value.slice(start, start + pageSize.value)
})

const stageHeaderLabel = computed(() => {
  if (activeGroup.value === 'ALL') return 'All prospects'
  if (activeGroup.value === 'IN_PROGRESS' && secondaryStage.value !== 'ALL_PROGRESS') return `${stageLabel(secondaryStage.value)} prospects`
  return `${GROUP_LABELS[activeGroup.value]} prospects`
})

const stageHeaderCount = computed(() => {
  const n = totalFiltered.value
  return `${n} ${n === 1 ? 'prospect' : 'prospects'}`
})

const industryOptions = computed(() => {
  const set = new Set<string>()
  for (const p of prospects.value) { if (p.industryGroup) set.add(p.industryGroup) }
  return [...set].sort()
})

const needsAttentionCount = computed(() => {
  return prospects.value.filter(p => p.status === 'NEW_LEAD').length
})

const needsAttentionProspects = computed(() => {
  if (activeGroup.value === 'ALL') {
    return prospects.value.filter(p => p.status === 'NEW_LEAD').slice(0, 3)
  }
  return []
})

watch([activeGroup, secondaryStage, searchQuery, appliedFilters, sortOption], () => {
  currentPage.value = 1
})

watch(success, (val) => {
  if (val) {
    toast.add({
      severity: 'success',
      summary: 'Pipeline updated',
      detail: val,
      life: 4500,
    })
    success.value = ''
  }
})

const dialogType = computed(() => {
  if (target.value === 'LOST') return 'lost'
  if (target.value === 'WON') return 'won'
  if (!selected.value) return 'forward'
  return PIPELINE_STAGES.indexOf(target.value) < PIPELINE_STAGES.indexOf(selected.value.status) ? 'backward' : 'forward'
})

function selectGroup(group: PipelineGroup) {
  activeGroup.value = group
  secondaryStage.value = 'ALL_PROGRESS'
  success.value = ''
  successTarget.value = null
  successGroup.value = null
}

function selectSecondary(stage: SecondaryStage) {
  secondaryStage.value = stage
}

function openTransition(item: Prospect, status: ProspectStatus) {
  selected.value = item
  target.value = status
  notes.value = ''
  error.value = ''
}

function viewDetail(item: Prospect) {
  router.push(`/sales/my-prospects/${item.id}`)
}

function viewTargetStage(stage: ProspectStatus) {
  if (stage === 'NEW_LEAD') activeGroup.value = 'NEW_LEAD'
  else if (IN_PROGRESS_STAGES.includes(stage)) { activeGroup.value = 'IN_PROGRESS'; secondaryStage.value = stage }
  else if (stage === 'WON') activeGroup.value = 'WON'
  else if (stage === 'LOST') activeGroup.value = 'LOST'
  else activeGroup.value = 'ALL'
  secondaryStage.value = 'ALL_PROGRESS'
  success.value = ''
  successTarget.value = null
  successGroup.value = null
}

async function submit() {
  if (!selected.value) return
  error.value = ''
  try {
    const item = await crm.transition(selected.value.id, target.value, notes.value)
    success.value = `${item.placeName} moved to ${stageLabel(item.status)}.`
    successTarget.value = item.status
    selected.value = null
  } catch (caught) { error.value = crm.errorMessage(caught) }
}

function changePage(page: number) {
  currentPage.value = page
  const el = document.querySelector('.pl-list')
  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function loadMore() {
  pageSize.value += 15
  currentPage.value = 1
}

function openFilterDialog() {
  draftFilters.value = { ...appliedFilters.value }
  showFilters.value = true
}

function applyFilters() {
  appliedFilters.value = { ...draftFilters.value }
  showFilters.value = false
}

function resetFilters() {
  draftFilters.value = { industryGroup: '', visitStatus: '' }
  appliedFilters.value = { industryGroup: '', visitStatus: '' }
  showFilters.value = false
}

function groupFromStatus(s: ProspectStatus): PipelineGroup {
  if (s === 'NEW_LEAD') return 'NEW_LEAD'
  if (IN_PROGRESS_STAGES.includes(s)) return 'IN_PROGRESS'
  if (s === 'WON') return 'WON'
  if (s === 'LOST') return 'LOST'
  return 'ALL'
}

const SORT_OPTIONS = [
  { value: 'updatedAt-desc', label: 'Recently updated' },
  { value: 'updatedAt-asc', label: 'Oldest updated' },
  { value: 'name-asc', label: 'Name A–Z' },
  { value: 'name-desc', label: 'Name Z–A' },
  { value: 'stage-order', label: 'Stage order' },
]

onMounted(async () => {
  try {
    await crm.loadMyProspects()
    const qId = route.query.prospectId as string | undefined
    const qAction = route.query.action as string | undefined
    if (qId && qAction === 'update') {
      highlightId.value = qId
      const found = prospects.value.find(p => p.id === qId)
      if (found) {
        const g = groupFromStatus(found.status)
        activeGroup.value = g
        if (g === 'IN_PROGRESS') secondaryStage.value = found.status
        await nextTick()
        const el = document.getElementById(`prospect-card-${qId}`)
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        const name = found.placeName || 'prospect'
        success.value = `Visit completed. Review ${name} and move it to the next stage when ready.`
      } else {
        success.value = 'Prospect updated — review and move to the next stage when ready.'
      }
      setTimeout(() => { highlightId.value = null }, 4000)
    }
  } catch (caught) { error.value = crm.errorMessage(caught) }
})
</script>

<template>
  <section class="pl-page">
    <RouterLink class="pl-back" to="/sales/dashboard">
      <i class="pi pi-arrow-left" />
    </RouterLink>

    <div class="pl-header">
      <div class="pl-header-text">
        <p class="pl-eyebrow">Track assigned prospects by stage</p>
        <h1 class="pl-title">Sales Pipeline</h1>
      </div>
      <span class="pl-total-badge">{{ prospects.length }} active</span>
    </div>

    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div v-if="crm.loading && !prospects.length" class="pl-loading">
      <i class="pi pi-spin pi-spinner" />
      <span>Loading prospects</span>
    </div>

    <template v-else>
      <!-- ── GROUP TABS ── -->
      <div class="pg-tabs">
        <button
          v-for="group in (['ALL', 'NEW_LEAD', 'IN_PROGRESS', 'WON', 'LOST'] as PipelineGroup[])"
          :key="group"
          class="pg-tab"
          :class="{ active: activeGroup === group }"
          @click="selectGroup(group)"
        >
          <span class="pg-tab-label">{{ GROUP_LABELS[group] }}</span>
          <span class="pg-tab-count">{{ groupCounts[group] }}</span>
        </button>
      </div>

      <!-- ── SECONDARY STAGE SELECTOR (IN PROGRESS only) ── -->
      <div v-if="activeGroup === 'IN_PROGRESS'" class="pl-secondary">
        <template v-if="isDesktop">
          <div class="ps-chips">
            <button
              v-for="opt in SECONDARY_OPTIONS"
              :key="opt.value"
              class="ps-chip"
              :class="{ active: secondaryStage === opt.value }"
              @click="selectSecondary(opt.value)"
            >
              {{ opt.label }}
              <span class="ps-chip-count">{{ secondaryCounts[opt.value] }}</span>
            </button>
          </div>
        </template>
        <template v-else>
          <label class="ps-select-label">Stage:</label>
          <select
            class="ps-select"
            :value="secondaryStage"
            @change="selectSecondary(($event.target as HTMLSelectElement).value as SecondaryStage)"
          >
            <option
              v-for="opt in SECONDARY_OPTIONS"
              :key="opt.value"
              :value="opt.value"
            >{{ opt.label }} ({{ secondaryCounts[opt.value] }})</option>
          </select>
        </template>
      </div>

      <!-- ── NEEDS ATTENTION ── -->
      <div v-if="activeGroup === 'ALL' && needsAttentionProspects.length" class="pl-attention">
        <div class="pl-attention-head">
          <i class="pi pi-exclamation-circle" />
          <strong>Needs attention</strong>
          <span class="pl-attention-count">{{ needsAttentionCount }} prospects</span>
        </div>
        <div class="pl-attention-list">
          <div v-for="p in needsAttentionProspects" :key="p.id" class="pl-attention-item">
            <i class="pi pi-arrow-right" />
            <span>{{ p.placeName }}</span>
            <span class="pl-attention-stage">New Lead</span>
          </div>
        </div>
      </div>

      <!-- ── TOOLBAR ── -->
      <div class="pl-toolbar">
        <div class="pl-search">
          <i class="pi pi-search" />
          <input v-model="searchQuery" type="text" placeholder="Search pipeline prospects" class="pl-search-input" />
          <button v-if="searchQuery" class="pl-search-clear" @click="searchQuery = ''"><i class="pi pi-times" /></button>
        </div>
        <div class="pl-toolbar-actions">
          <button class="pl-tb-btn" @click="openFilterDialog" title="Filters">
            <i class="pi pi-filter" />
          </button>
          <div class="pl-sort-wrap">
            <select v-model="sortOption" class="pl-sort-select" title="Sort">
              <option v-for="opt in SORT_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
        </div>
      </div>

      <!-- ── STAGE HEADER ── -->
      <div class="pl-stage-header">
        <span class="pl-stage-title">{{ stageHeaderLabel }}</span>
        <span class="pl-stage-count">{{ stageHeaderCount }}</span>
      </div>

      <!-- ── LIST ── -->
      <div class="pl-list">
        <template v-for="item in paginatedProspects" :key="item.id">
          <PipelineProspectCard
            v-if="isDesktop"
            :item="item"
            :highlight="highlightId === item.id"
            :compact="true"
            @move-next="(i) => { const s = nextStage(i.status); if (s) openTransition(i, s) }"
            @move-prev="(i) => { const s = previousStage(i.status); if (s) openTransition(i, s) }"
            @mark-lost="(i) => openTransition(i, 'LOST')"
            @mark-won="(i) => openTransition(i, 'WON')"
            @view-detail="viewDetail"
          />
          <PipelineProspectCard
            v-else
            :item="item"
            :highlight="highlightId === item.id"
            @move-next="(i) => { const s = nextStage(i.status); if (s) openTransition(i, s) }"
            @move-prev="(i) => { const s = previousStage(i.status); if (s) openTransition(i, s) }"
            @mark-lost="(i) => openTransition(i, 'LOST')"
            @mark-won="(i) => openTransition(i, 'WON')"
            @view-detail="viewDetail"
          />
        </template>

        <div v-if="!paginatedProspects.length" class="pl-empty">
          <div class="pl-empty-icon"><i class="pi pi-inbox" /></div>
          <strong class="pl-empty-title">
            <template v-if="searchQuery">No matching prospects</template>
            <template v-else-if="activeGroup === 'NEW_LEAD'">No new leads</template>
            <template v-else-if="activeGroup === 'IN_PROGRESS'">No prospects in progress</template>
            <template v-else-if="activeGroup === 'WON'">No won prospects yet</template>
            <template v-else-if="activeGroup === 'LOST'">No lost prospects</template>
            <template v-else>No prospects in pipeline</template>
          </strong>
          <p class="pl-empty-text">
            <template v-if="searchQuery">Try changing your search or filters.</template>
            <template v-else-if="activeGroup === 'NEW_LEAD'">Newly assigned prospects will appear here.</template>
            <template v-else-if="activeGroup === 'IN_PROGRESS'">Move a New Lead forward to begin the sales process.</template>
            <template v-else-if="activeGroup === 'WON'">Successful prospects will appear here.</template>
            <template v-else-if="activeGroup === 'LOST'">Closed prospects will appear here.</template>
            <template v-else>Prospects will appear here after they are assigned.</template>
          </p>
        </div>
      </div>

      <!-- ── PAGINATION ── -->
      <div v-if="totalFiltered > 0" class="pl-pagination">
        <template v-if="isDesktop">
          <span class="pl-page-info">Showing {{ pageStart }}–{{ pageEnd }} of {{ totalFiltered }}</span>
          <div class="pl-page-controls">
            <button class="pl-page-btn" :disabled="currentPage <= 1" @click="changePage(currentPage - 1)">Previous</button>
            <span class="pl-page-num">Page {{ currentPage }} of {{ totalPages }}</span>
            <button class="pl-page-btn" :disabled="currentPage >= totalPages" @click="changePage(currentPage + 1)">Next</button>
          </div>
        </template>
        <template v-else>
          <span class="pl-page-info">Showing {{ pageEnd }} of {{ totalFiltered }}</span>
          <button v-if="!limitReached" class="pl-load-btn" @click="loadMore">Load more prospects</button>
        </template>
      </div>
    </template>

    <!-- ── FILTER DIALOG ── -->
    <Dialog
      :visible="showFilters"
      modal
      header="Filters"
      :style="{ width: 'min(92vw, 480px)' }"
      :pt="{ mask: { style: 'backdrop-filter: blur(4px);' } }"
      @update:visible="(v) => { if (!v) showFilters = false }"
    >
      <div class="pl-filter-body">
        <label class="pl-filter-field">
          <span>Industry / Category</span>
          <select v-model="draftFilters.industryGroup" class="pl-filter-select">
            <option value="">All industries</option>
            <option v-for="ind in industryOptions" :key="ind" :value="ind">{{ ind }}</option>
          </select>
        </label>
      </div>
      <template #footer>
        <div class="pl-filter-footer">
          <span class="pl-filter-total">{{ totalFiltered }} prospects</span>
          <div class="pl-filter-actions">
            <Button label="Reset" severity="secondary" text size="small" @click="resetFilters" />
            <Button label="Apply" size="small" @click="applyFilters" />
          </div>
        </div>
      </template>
    </Dialog>

    <!-- ── TRANSITION DIALOG ── -->
    <Dialog
      :visible="selected !== null"
      modal
      :style="{ width: 'min(92vw, 560px)' }"
      :header="dialogType === 'lost' ? 'Mark prospect as Lost' : dialogType === 'won' ? 'Mark prospect as Won' : dialogType === 'backward' ? 'Move prospect back' : 'Move Prospect Stage'"
      :pt="{ mask: { style: 'backdrop-filter: blur(4px);' } }"
      @update:visible="(visible) => { if (!visible) selected = null }"
    >
      <div class="pl-dialog-body">
        <p class="pl-dialog-prospect" v-if="selected"><strong>{{ selected.placeName }}</strong></p>
        <div class="pl-dialog-flow" v-if="selected && target !== 'LOST' && target !== 'WON'">
          <span class="pl-flow-from">{{ stageLabel(selected.status) }}</span>
          <i class="pi pi-arrow-right" />
          <span class="pl-flow-to">{{ stageLabel(target) }}</span>
        </div>
        <p v-if="dialogType === 'forward' && selected" class="pl-dialog-message">
          You are moving <strong>{{ selected.placeName }}</strong> from <strong>{{ stageLabel(selected.status) }}</strong> to <strong>{{ stageLabel(target) }}</strong>.
        </p>
        <p v-if="dialogType === 'backward'" class="pl-dialog-message">
          Use this when the prospect needs to return to the previous sales stage.
        </p>
        <p v-if="dialogType === 'won'" class="pl-dialog-success"><i class="pi pi-check-circle" /> Won prospects will wait for Admin review before conversion.</p>
        <p v-if="dialogType === 'lost'" class="pl-dialog-warning"><i class="pi pi-exclamation-triangle" /> This closes the active sales process for this prospect.</p>
        <label class="pl-field">
          <span>{{ target === 'LOST' ? 'Loss reason (required)' : target === 'WON' ? 'Win notes (required)' : 'Progress note (optional)' }}</span>
          <Textarea v-model="notes" rows="3" fluid placeholder="Add details about this transition..." />
        </label>
      </div>
      <template #footer>
        <div class="pl-dialog-footer">
          <Button label="Cancel" severity="secondary" text @click="selected = null" />
          <Button
            :label="target === 'LOST' ? 'Confirm lost' : target === 'WON' ? 'Confirm won' : 'Confirm move'"
            :severity="target === 'LOST' ? 'danger' : target === 'WON' ? 'success' : 'primary'"
            :disabled="(target === 'LOST' || target === 'WON') && !notes.trim()"
            :loading="crm.loading"
            @click="submit"
          />
        </div>
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.pl-page { padding: 0 0 24px; }

/* ── BACK ── */
.pl-back {
  display: inline-flex; align-items: center; justify-content: center;
  width: 2rem; height: 2rem; color: var(--brand-blue); background: var(--brand-blue-bg);
  border: 1px solid transparent; border-radius: var(--radius-md);
  text-decoration: none; font-size: 0.9rem; margin-bottom: 8px;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}
.pl-back:hover { background: #dbeafe; border-color: var(--brand-blue); }

/* ── HEADER ── */
.pl-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 10px; }
.pl-header-text { flex: 1; }
.pl-eyebrow { margin: 0; font-size: 0.68rem; color: #64748b; font-weight: 500; }
.pl-title { margin: 2px 0 0; font-size: 1.25rem; font-weight: 800; color: #0f172a; line-height: 1.3; }
.pl-total-badge {
  flex-shrink: 0; padding: 5px 12px; background: #eff6ff; color: #2563eb;
  border: 1px solid #bfdbfe; border-radius: 20px; font-size: 0.72rem; font-weight: 700; white-space: nowrap;
}
.pl-loading { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 48px 0; color: #94a3b8; font-size: 0.82rem; font-weight: 600; }
.pl-view-target-btn { margin-top: 6px; }

/* ── GROUP TABS ── */
.pg-tabs {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 0.5rem;
  margin-bottom: 10px;
}
.pg-tab {
  min-height: 46px;
  padding: 0.6rem 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  border-radius: 12px;
  border: 1.5px solid #e2e8f0;
  background: #fff;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.12s;
  -webkit-tap-highlight-color: transparent;
  min-width: 0;
}
.pg-tab:active { transform: scale(0.97); }
.pg-tab-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: #475569;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pg-tab-count {
  min-width: 22px;
  height: 22px;
  padding: 0 5px;
  display: inline-grid;
  place-items: center;
  border-radius: 9999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 0.62rem;
  font-weight: 800;
  line-height: 1;
  flex-shrink: 0;
}
.pg-tab.active {
  background: #2563eb;
  border-color: #2563eb;
}
.pg-tab.active .pg-tab-label { color: #fff; }
.pg-tab.active .pg-tab-count { background: rgba(255,255,255,0.22); color: #fff; }

@media (max-width: 480px) {
  .pg-tabs { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .pg-tab { min-height: 44px; padding: 0.5rem 0.3rem; }
  .pg-tab-label { font-size: 0.65rem; }
}

/* ── SECONDARY STAGE ── */
.pl-secondary { margin-bottom: 8px; }
.ps-chips { display: flex; flex-wrap: wrap; gap: 6px; }
.ps-chip {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 10px; border-radius: 8px;
  border: 1px solid #e2e8f0; background: #fff;
  font-size: 0.68rem; font-weight: 600; color: #475569;
  cursor: pointer; transition: all 0.12s;
  font-family: inherit;
}
.ps-chip:hover { background: #f8fafc; }
.ps-chip.active { background: #eef2ff; border-color: #6366f1; color: #4338ca; }
.ps-chip-count {
  padding: 1px 5px; border-radius: 9999px;
  background: #f1f5f9; color: #64748b;
  font-size: 0.6rem; font-weight: 700;
}
.ps-chip.active .ps-chip-count { background: rgba(99,102,241,0.12); color: #4338ca; }
.ps-select-label { font-size: 0.72rem; font-weight: 600; color: #475569; margin-right: 6px; }
.ps-select {
  width: 100%; padding: 8px 10px; border: 1px solid #e2e8f0;
  border-radius: 10px; background: #fff; font-size: 0.8rem;
  font-family: inherit; color: #1e293b;
}

/* ── NEEDS ATTENTION ── */
.pl-attention {
  border: 1px solid #fef3c7; border-radius: 14px;
  background: #fffbeb; padding: 10px 14px; margin-bottom: 8px;
}
.pl-attention-head {
  display: flex; align-items: center; gap: 6px; font-size: 0.75rem;
  color: #92400e; margin-bottom: 6px;
}
.pl-attention-head i { font-size: 0.75rem; color: #f59e0b; }
.pl-attention-count { color: #a16207; font-weight: 500; margin-left: auto; font-size: 0.65rem; }
.pl-attention-list { display: flex; flex-direction: column; gap: 3px; }
.pl-attention-item {
  display: flex; align-items: center; gap: 6px;
  font-size: 0.72rem; color: #78350f; padding: 2px 0;
}
.pl-attention-item i { font-size: 0.5rem; color: #f59e0b; }
.pl-attention-stage {
  margin-left: auto; padding: 1px 6px; border-radius: 4px;
  background: #fef3c7; font-size: 0.6rem; font-weight: 600; color: #92400e;
}

/* ── TOOLBAR ── */
.pl-toolbar { display: flex; gap: 8px; margin-bottom: 8px; align-items: center; }
.pl-search {
  flex: 1; display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; background: #f8fafc; border: 1px solid #e2e8f0;
  border-radius: 12px; transition: border-color 0.15s;
}
.pl-search:focus-within { border-color: #93c5fd; background: #fff; }
.pl-search i { color: #94a3b8; font-size: 0.8rem; }
.pl-search-input { flex: 1; border: none; background: none; outline: none; font-size: 0.8rem; color: #1e293b; font-family: inherit; }
.pl-search-input::placeholder { color: #94a3b8; }
.pl-search-clear {
  display: flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border: none; background: #e2e8f0;
  border-radius: 50%; cursor: pointer; color: #64748b; font-size: 0.6rem; transition: background 0.12s;
}
.pl-search-clear:hover { background: #cbd5e1; }
.pl-toolbar-actions { display: flex; gap: 6px; flex-shrink: 0; }
.pl-tb-btn {
  display: flex; align-items: center; justify-content: center;
  width: 36px; height: 36px; border: 1px solid #e2e8f0;
  border-radius: 10px; background: #fff; cursor: pointer;
  color: #64748b; font-size: 0.8rem; transition: background 0.12s;
}
.pl-tb-btn:hover { background: #f8fafc; }
.pl-sort-select {
  padding: 6px 8px; border: 1px solid #e2e8f0; border-radius: 10px;
  background: #fff; font-size: 0.7rem; font-family: inherit;
  color: #475569; cursor: pointer; height: 36px;
}

/* ── STAGE HEADER ── */
.pl-stage-header { display: flex; align-items: center; justify-content: space-between; padding: 6px 0 4px; }
.pl-stage-title { font-size: 0.82rem; font-weight: 700; color: #1e293b; text-transform: capitalize; }
.pl-stage-count { font-size: 0.7rem; color: #94a3b8; font-weight: 500; }

/* ── LIST ── */
.pl-list { display: flex; flex-direction: column; gap: 8px; }

/* ── EMPTY ── */
.pl-empty { display: flex; flex-direction: column; align-items: center; padding: 40px 20px; text-align: center; }
.pl-empty-icon { width: 48px; height: 48px; border-radius: 14px; background: #f1f5f9; display: flex; align-items: center; justify-content: center; margin-bottom: 12px; }
.pl-empty-icon i { font-size: 1.2rem; color: #94a3b8; }
.pl-empty-title { font-size: 0.88rem; color: #334155; font-weight: 700; margin-bottom: 4px; }
.pl-empty-text { margin: 0; font-size: 0.75rem; color: #94a3b8; }

/* ── PAGINATION ── */
.pl-pagination {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 0 4px; gap: 12px;
}
.pl-page-info { font-size: 0.72rem; color: #94a3b8; font-weight: 500; }
.pl-page-controls { display: flex; align-items: center; gap: 10px; }
.pl-page-btn {
  padding: 6px 14px; border: 1px solid #e2e8f0; border-radius: 8px;
  background: #fff; font-size: 0.7rem; font-weight: 600; color: #475569;
  cursor: pointer; transition: background 0.12s; font-family: inherit;
}
.pl-page-btn:hover:not(:disabled) { background: #f8fafc; }
.pl-page-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.pl-page-num { font-size: 0.7rem; color: #64748b; font-weight: 500; }
.pl-load-btn {
  padding: 8px 20px; border: 1px solid #2563eb; border-radius: 10px;
  background: #fff; color: #2563eb; font-size: 0.75rem; font-weight: 600;
  cursor: pointer; transition: background 0.12s; font-family: inherit;
}
.pl-load-btn:hover { background: #eff6ff; }

/* ── FILTER DIALOG ── */
.pl-filter-body { display: flex; flex-direction: column; gap: 12px; padding: 4px 0; }
.pl-filter-field { display: flex; flex-direction: column; gap: 4px; }
.pl-filter-field span { font-size: 0.75rem; font-weight: 600; color: #475569; }
.pl-filter-select {
  width: 100%; padding: 8px 10px; border: 1px solid #e2e8f0;
  border-radius: 10px; background: #fff; font-size: 0.8rem;
  font-family: inherit; color: #1e293b;
}
.pl-filter-footer { display: flex; align-items: center; justify-content: space-between; width: 100%; }
.pl-filter-total { font-size: 0.72rem; color: #94a3b8; font-weight: 500; }
.pl-filter-actions { display: flex; gap: 8px; }

/* ── DIALOG ── */
.pl-dialog-body { display: flex; flex-direction: column; gap: 10px; }
.pl-dialog-prospect { margin: 0; font-size: 0.9rem; }
.pl-dialog-prospect strong { color: #0f172a; }
.pl-dialog-flow { display: flex; align-items: center; gap: 10px; padding: 10px 14px; background: #f8fafc; border-radius: 10px; font-size: 0.78rem; font-weight: 600; color: #475569; }
.pl-dialog-flow i { color: #2563eb; font-size: 0.8rem; }
.pl-flow-from { color: #64748b; }
.pl-flow-to { color: #2563eb; }
.pl-dialog-message { margin: 0; font-size: 0.78rem; color: #475569; line-height: 1.5; }
.pl-dialog-message strong { color: #1e293b; }
.pl-dialog-warning { margin: 0; padding: 10px 14px; background: #fef2f2; border: 1px solid #fecaca; border-radius: 10px; font-size: 0.75rem; color: #991b1b; display: flex; align-items: flex-start; gap: 8px; }
.pl-dialog-warning i { margin-top: 1px; flex-shrink: 0; }
.pl-dialog-success { margin: 0; padding: 10px 14px; background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 10px; font-size: 0.75rem; color: #166534; display: flex; align-items: flex-start; gap: 8px; }
.pl-dialog-success i { margin-top: 1px; flex-shrink: 0; }
.pl-field { display: flex; flex-direction: column; gap: 4px; }
.pl-field span { font-size: 0.75rem; font-weight: 600; color: #475569; }
.pl-dialog-footer { display: flex; justify-content: flex-end; gap: 8px; }

/* ── RESPONSIVE ── */
@media (max-width: 480px) {
  .pg-tabs { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .pg-tab:nth-child(4),
  .pg-tab:nth-child(5) { grid-column: span 1; }
  .pl-toolbar { flex-direction: column; }
  .pl-toolbar-actions { width: 100%; }
  .pl-sort-select { flex: 1; }
  .pl-pagination { flex-direction: column; align-items: flex-start; }
}

/* ── DESKTOP ── */
@media (min-width: 1024px) {
  .pl-back { display: none; }
  .pl-page { padding: 0 0 32px; }
  .pl-title { font-size: 1.5rem; }
  .pl-eyebrow { font-size: 0.75rem; }
  .pl-total-badge { font-size: 0.75rem; }

  .pg-tab { min-height: 48px; }
  .pg-tab-label { font-size: 0.74rem; }

  .pl-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 10px;
    align-items: start;
  }

  .pl-stage-header { padding: 8px 0 6px; }
  .pl-stage-title { font-size: 0.9rem; }
}
</style>
