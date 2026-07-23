<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { BOARD_STATUSES, nextStage, previousStage } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus } from '../../../types/crm'

const crm = useCrmStore()
const selected = ref<Prospect | null>(null)
const target = ref<ProspectStatus>('CONTACTED')
const notes = ref('')
const error = ref('')
const success = ref('')
const touchStart = ref(0)
const prospects = computed(() => crm.myProspects.filter((item) => item.status !== 'CONVERTED'))
const byStage = (stage: ProspectStatus) => prospects.value.filter((item) => item.status === stage)
const title = (stage: ProspectStatus) => stage.replaceAll('_', ' ')

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
async function submit() {
  if (!selected.value) return
  error.value = ''
  try {
    const item = await crm.transition(selected.value.id, target.value, notes.value)
    success.value = `${item.placeName} moved to ${title(item.status)}.`
    selected.value = null
  } catch (caught) { error.value = crm.errorMessage(caught) }
}
onMounted(async () => { try { await crm.loadMyProspects() } catch (caught) { error.value = crm.errorMessage(caught) } })
</script>

<template>
  <section class="mobile-page pipeline-page">
    <div class="mobile-title"><div><p class="eyebrow">My assigned prospects</p><h1>Pipeline</h1></div><Tag :value="`${prospects.length} total`" severity="info" /></div>
    <p class="mobile-hint">Swipe left/right or use the buttons to move exactly one stage. Tap a card for saved Place, visit, and history details.</p>
    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div v-if="crm.loading && !prospects.length" class="empty-state"><i class="pi pi-spin pi-spinner" /><strong>Loading prospects</strong></div>
    <div v-else class="pipeline-board sales-pipeline-board">
      <section v-for="stage in BOARD_STATUSES" :key="stage" class="pipeline-column">
        <header><strong>{{ title(stage) }}</strong><span>{{ byStage(stage).length }}</span></header>
        <div class="pipeline-column-body">
          <article v-for="item in byStage(stage)" :key="item.id" class="kanban-card" @touchstart="touchStart = $event.touches[0].clientX" @touchend="finishTouch(item, $event)">
            <RouterLink :to="`/sales/my-prospects/${item.id}`">
              <span class="industry-pill">{{ item.industryGroup }}</span><h2>{{ item.placeName }}</h2><p><i class="pi pi-map-marker" /> {{ item.formattedAddress }}</p><small>{{ item.assignedSalesExecutive }} · {{ title(item.status) }}</small>
            </RouterLink>
            <div v-if="!['LOST', 'WON'].includes(item.status)" class="kanban-controls">
              <Button icon="pi pi-arrow-left" aria-label="Previous stage" severity="secondary" text :disabled="!previousStage(item.status)" @click="previousStage(item.status) && openTransition(item, previousStage(item.status)!)" />
              <Button label="Lost" severity="danger" text @click="openTransition(item, 'LOST')" />
              <Button v-if="item.status === 'NEGOTIATION'" label="Mark Won" icon="pi pi-check" severity="success" @click="openTransition(item, 'WON')" />
              <Button v-else icon="pi pi-arrow-right" aria-label="Next stage" :disabled="!nextStage(item.status)" @click="nextStage(item.status) && openTransition(item, nextStage(item.status)!)" />
            </div>
          </article>
          <p v-if="!byStage(stage).length" class="pipeline-empty">No prospects</p>
        </div>
      </section>
    </div>
    <Dialog :visible="selected !== null" modal :header="target === 'LOST' ? 'Record lost prospect' : target === 'WON' ? 'Confirm won prospect' : `Move to ${title(target)}`" :style="{ width: 'min(92vw, 440px)' }" @update:visible="(visible) => { if (!visible) selected = null }">
      <p class="muted">{{ selected?.placeName }}. The change will be stored in status history.</p>
      <label class="field"><span>{{ target === 'LOST' ? 'Loss reason (required)' : target === 'WON' ? 'Win notes (required)' : 'Progress note (optional)' }}</span><Textarea v-model="notes" rows="4" fluid /></label>
      <template #footer><Button label="Cancel" severity="secondary" text @click="selected = null" /><Button label="Confirm" :severity="target === 'LOST' ? 'danger' : 'primary'" :disabled="(target === 'LOST' || target === 'WON') && !notes.trim()" :loading="crm.loading" @click="submit" /></template>
    </Dialog>
  </section>
</template>

<style scoped>
.pipeline-board { width: 100%; padding: 0.2rem 0 1rem; display: flex; gap: 0.8rem; overflow-x: auto; scroll-snap-type: x proximity; }
.pipeline-column { width: 285px; min-width: 285px; min-height: 360px; overflow: hidden; background: #eef3f9; border: 1px solid var(--border-light); border-radius: var(--radius-lg); scroll-snap-align: start; }
.pipeline-column > header { padding: 0.7rem 0.8rem; display: flex; align-items: center; justify-content: space-between; color: #26344b; background: var(--surface-card); border-bottom: 1px solid var(--border-light); font-size: 0.68rem; }
.pipeline-column > header strong { font-weight: 700; }
.pipeline-column > header span { min-width: 1.4rem; padding: 0.15rem 0.4rem; color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: 1rem; text-align: center; font-size: 0.62rem; font-weight: 700; }
.pipeline-column-body { padding: 0.6rem; display: grid; align-content: start; gap: 0.5rem; }
.kanban-card { padding: 0.7rem; background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-md); box-shadow: var(--shadow-xs); transition: box-shadow var(--transition-fast), border-color var(--transition-fast); }
.kanban-card:hover { box-shadow: var(--shadow-sm); border-color: var(--border-default); }
.kanban-card > a { color: inherit; text-decoration: none; }
.kanban-card h2 { margin: 0.35rem 0 0.2rem; font-size: 0.75rem; line-height: 1.35; font-weight: 700; }
.kanban-card p { margin: 0.25rem 0; color: #68758a; font-size: 0.58rem; line-height: 1.45; }
.kanban-card small { color: var(--text-muted); font-size: 0.52rem; }
.kanban-controls { margin-top: 0.5rem; display: flex; align-items: center; justify-content: space-between; border-top: 1px solid #edf1f5; padding-top: 0.4rem; }
.kanban-controls :deep(.p-button) { padding: 0.3rem; font-size: 0.54rem; }
.pipeline-empty { margin: 0.75rem 0; color: #8a96a8; text-align: center; font-size: 0.6rem; }
.mobile-hint { margin: 0; color: #758297; font-size: 0.58rem; line-height: 1.5; }
.sales-pipeline-board { margin-right: -0.8rem; width: calc(100% + 0.8rem); }
.industry-pill { display: inline-block; padding: 0.1rem 0.4rem; color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: 0.3rem; font-size: 0.48rem; font-weight: 600; letter-spacing: 0.02em; }
</style>
