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
