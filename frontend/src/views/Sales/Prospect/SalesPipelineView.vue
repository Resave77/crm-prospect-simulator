<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Textarea from 'primevue/textarea'
import { BOARD_STATUSES, nextStage, previousStage } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus } from '../../../types/crm'
import PipelineStageStrip from '../../../components/sales/pipeline/PipelineStageStrip.vue'
import PipelineProspectCard from '../../../components/sales/pipeline/PipelineProspectCard.vue'
import { stageLabel } from '../../../components/sales/pipeline/stageColors'

type StageFilter = ProspectStatus | 'ALL'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()

const selected = ref<Prospect | null>(null)
const target = ref<ProspectStatus>('CONTACTED')
const notes = ref('')
const error = ref('')
const success = ref('')
const successTarget = ref<ProspectStatus | null>(null)
const touchStart = ref(0)
const highlightId = ref<string | null>(null)
const selectedStage = ref<StageFilter>('ALL')
const searchQuery = ref('')

const prospects = computed(() => crm.myProspects.filter((item) => item.status !== 'CONVERTED'))
const byStage = (stage: ProspectStatus) => prospects.value.filter((item) => item.status === stage)

const allStageCounts = computed(() => {
  const counts: Record<ProspectStatus, number> = {} as Record<ProspectStatus, number>
  for (const s of BOARD_STATUSES) counts[s] = 0
  for (const p of prospects.value) {
    if (counts[p.status] !== undefined) counts[p.status]++
  }
  return counts
})

const searchLower = computed(() => searchQuery.value.trim().toLowerCase())

const filteredProspects = computed(() => {
  let list = selectedStage.value === 'ALL'
    ? prospects.value
    : prospects.value.filter((p) => p.status === selectedStage.value)
  if (searchLower.value) {
    list = list.filter((p) =>
      `${p.placeName} ${p.industryGroup} ${p.formattedAddress} ${p.assignedSalesExecutive} ${stageLabel(p.status)}`.toLowerCase().includes(searchLower.value)
    )
  }
  return list
})

const stageProspects = computed(() => filteredProspects.value)

function viewTargetStage(stage: ProspectStatus) {
  selectedStage.value = stage
  success.value = ''
  successTarget.value = null
}

function openTransition(item: Prospect, status: ProspectStatus) {
  selected.value = item
  target.value = status
  notes.value = ''
  error.value = ''
}

function finishTouch(item: Prospect, event: TouchEvent) {
  const delta = event.changedTouches[0].clientX - touchStart.value
  if (delta < -70) { const stage = nextStage(item.status); if (stage) openTransition(item, stage) }
  if (delta > 70) { const stage = previousStage(item.status); if (stage) openTransition(item, stage) }
}

function viewDetail(item: Prospect) {
  router.push(`/sales/my-prospects/${item.id}`)
}

function selectStage(stage: StageFilter) {
  selectedStage.value = stage
}

async function submit() {
  if (!selected.value) return
  error.value = ''
  try {
    const prevStage = selected.value.status
    const item = await crm.transition(selected.value.id, target.value, notes.value)
    success.value = `${item.placeName} moved to ${stageLabel(item.status)}.`
    successTarget.value = item.status
    selected.value = null
  } catch (caught) { error.value = crm.errorMessage(caught) }
}

const stageHeaderLabel = computed(() => {
  if (selectedStage.value === 'ALL') return 'All prospects'
  return `${stageLabel(selectedStage.value)} prospects`
})

const stageHeaderCount = computed(() => {
  const n = stageProspects.value.length
  return `${n} ${n === 1 ? 'prospect' : 'prospects'}${selectedStage.value !== 'ALL' ? ' in this stage' : ''}`
})

onMounted(async () => {
  try {
    await crm.loadMyProspects()
    const qId = route.query.prospectId as string | undefined
    const qAction = route.query.action as string | undefined
    if (qId && qAction === 'update') {
      highlightId.value = qId
      const found = prospects.value.find((p) => p.id === qId)
      if (found) {
        selectedStage.value = found.status
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
      <span>Back</span>
    </RouterLink>

    <div class="pl-header">
      <div class="pl-header-text">
        <p class="pl-eyebrow">Track assigned prospects by stage</p>
        <h1 class="pl-title">Sales Pipeline</h1>
      </div>
      <span class="pl-total-badge">{{ prospects.length }} active</span>
    </div>

    <Message v-if="success" severity="success" closable @close="success = ''">
      {{ success }}
      <Button
        v-if="successTarget"
        :label="`View in ${stageLabel(successTarget)}`"
        size="small"
        text
        class="pl-view-target-btn"
        @click="viewTargetStage(successTarget!)"
      />
    </Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div v-if="crm.loading && !prospects.length" class="pl-loading">
      <i class="pi pi-spin pi-spinner" />
      <span>Loading prospects</span>
    </div>

    <template v-else>
      <div class="pl-strip-section">
        <PipelineStageStrip
          :all-count="prospects.length"
          :counts="allStageCounts"
          :active-stage="selectedStage"
          @select="selectStage"
        />
      </div>

      <div class="pl-search-row">
        <div class="pl-search">
          <i class="pi pi-search" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search pipeline prospects"
            class="pl-search-input"
          />
          <button v-if="searchQuery" class="pl-search-clear" @click="searchQuery = ''">
            <i class="pi pi-times" />
          </button>
        </div>
      </div>

      <div class="pl-stage-header">
        <span class="pl-stage-title">{{ stageHeaderLabel }}</span>
        <span class="pl-stage-count">{{ stageHeaderCount }}</span>
      </div>

      <div class="pl-list">
        <PipelineProspectCard
          v-for="item in stageProspects"
          :key="item.id"
          :item="item"
          :highlight="highlightId === item.id"
          @move-next="(i) => nextStage(i.status) && openTransition(i, nextStage(i.status)!)"
          @move-prev="(i) => previousStage(i.status) && openTransition(i, previousStage(i.status)!)"
          @mark-lost="(i) => openTransition(i, 'LOST')"
          @mark-won="(i) => openTransition(i, 'WON')"
          @view-detail="viewDetail"
          @touch-start="(x) => touchStart = x"
          @touch-end="(i, e) => finishTouch(i, e)"
        />

        <div v-if="!stageProspects.length" class="pl-empty">
          <div class="pl-empty-icon"><i class="pi pi-inbox" /></div>
          <strong class="pl-empty-title">
            {{ searchQuery ? 'No matching prospects' : `No prospects in ${selectedStage === 'ALL' ? 'pipeline' : stageLabel(selectedStage)}` }}
          </strong>
          <p class="pl-empty-text">
            {{ searchQuery ? 'Try a different search term.' : 'Prospects will appear here after they move into this stage.' }}
          </p>
          <div class="pl-empty-actions">
            <Button
              v-if="!searchQuery && selectedStage !== 'ALL' && prospects.length > 0"
              label="View all prospects"
              severity="secondary"
              text
              size="small"
              @click="selectStage('ALL')"
            />
          </div>
        </div>
      </div>
    </template>

    <Dialog
      :visible="selected !== null"
      modal
      :style="{ width: 'min(92vw, 560px)' }"
      :header="target === 'LOST' ? 'Record lost prospect' : target === 'WON' ? 'Confirm won prospect' : `Move to ${stageLabel(target)}`"
      :pt="{ mask: { style: 'backdrop-filter: blur(4px);' } }"
      @update:visible="(visible) => { if (!visible) selected = null }"
    >
      <div class="pl-dialog-body">
        <p class="pl-dialog-prospect" v-if="selected">
          <strong>{{ selected.placeName }}</strong>
        </p>
        <div class="pl-dialog-flow" v-if="selected && target !== 'LOST' && target !== 'WON'">
          <span class="pl-flow-from">{{ stageLabel(selected.status) }}</span>
          <i class="pi pi-arrow-right" />
          <span class="pl-flow-to">{{ stageLabel(target) }}</span>
        </div>
        <p v-if="target === 'LOST'" class="pl-dialog-warning">
          <i class="pi pi-exclamation-triangle" />
          This prospect will be marked as lost. This action will be stored in status history.
        </p>
        <p v-if="target === 'WON'" class="pl-dialog-success">
          <i class="pi pi-check-circle" />
          Won prospects will wait for Admin review before conversion.
        </p>
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
.pl-page {
  padding: 0 0 24px;
}

.pl-back {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #2563eb;
  font-size: 0.78rem;
  font-weight: 600;
  text-decoration: none;
  margin-bottom: 8px;
  transition: opacity 0.15s;
}

.pl-back:hover { opacity: 0.75; }

.pl-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.pl-header-text { flex: 1; }

.pl-eyebrow {
  margin: 0;
  font-size: 0.68rem;
  color: #64748b;
  font-weight: 500;
}

.pl-title {
  margin: 2px 0 0;
  font-size: 1.25rem;
  font-weight: 800;
  color: #0f172a;
  line-height: 1.3;
}

.pl-subtitle {
  margin: 4px 0 0;
  font-size: 0.75rem;
  color: #94a3b8;
  line-height: 1.4;
}

.pl-total-badge {
  flex-shrink: 0;
  padding: 5px 12px;
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid #bfdbfe;
  border-radius: 20px;
  font-size: 0.72rem;
  font-weight: 700;
  white-space: nowrap;
}

.pl-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 48px 0;
  color: #94a3b8;
  font-size: 0.82rem;
  font-weight: 600;
}

.pl-strip-section {
  margin-bottom: 6px;
}

.pl-search-row {
  margin-bottom: 8px;
}

.pl-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  transition: border-color 0.15s;
}

.pl-search:focus-within {
  border-color: #93c5fd;
  background: #fff;
}

.pl-search i {
  color: #94a3b8;
  font-size: 0.8rem;
}

.pl-search-input {
  flex: 1;
  border: none;
  background: none;
  outline: none;
  font-size: 0.8rem;
  color: #1e293b;
  font-family: inherit;
}

.pl-search-input::placeholder {
  color: #94a3b8;
}

.pl-search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: #e2e8f0;
  border-radius: 50%;
  cursor: pointer;
  color: #64748b;
  font-size: 0.6rem;
  transition: background 0.12s;
}

.pl-search-clear:hover {
  background: #cbd5e1;
}

.pl-stage-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0 4px;
}

.pl-stage-title {
  font-size: 0.82rem;
  font-weight: 700;
  color: #1e293b;
  text-transform: capitalize;
}

.pl-stage-count {
  font-size: 0.7rem;
  color: #94a3b8;
  font-weight: 500;
}

.pl-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pl-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  text-align: center;
}

.pl-empty-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}

.pl-empty-icon i {
  font-size: 1.2rem;
  color: #94a3b8;
}

.pl-empty-title {
  font-size: 0.88rem;
  color: #334155;
  font-weight: 700;
  margin-bottom: 4px;
}

.pl-empty-text {
  margin: 0 0 12px;
  font-size: 0.75rem;
  color: #94a3b8;
}

.pl-empty-actions {
  display: flex;
  gap: 8px;
}

.pl-view-target-btn {
  margin-top: 6px;
}

.pl-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pl-dialog-prospect {
  margin: 0;
  font-size: 0.9rem;
}

.pl-dialog-prospect strong {
  color: #0f172a;
}

.pl-dialog-flow {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: #f8fafc;
  border-radius: 10px;
  font-size: 0.78rem;
  font-weight: 600;
  color: #475569;
}

.pl-dialog-flow i {
  color: #2563eb;
  font-size: 0.8rem;
}

.pl-flow-from {
  color: #64748b;
}

.pl-flow-to {
  color: #2563eb;
}

.pl-dialog-warning {
  margin: 0;
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 10px;
  font-size: 0.75rem;
  color: #991b1b;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.pl-dialog-warning i {
  margin-top: 1px;
  flex-shrink: 0;
}

.pl-dialog-success {
  margin: 0;
  padding: 10px 14px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 10px;
  font-size: 0.75rem;
  color: #166534;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.pl-dialog-success i {
  margin-top: 1px;
  flex-shrink: 0;
}

.pl-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pl-field span {
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
}

.pl-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
