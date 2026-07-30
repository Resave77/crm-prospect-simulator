<script setup lang="ts">
import { stageTone, stageLabel } from './stageColors'
import { nextStage, previousStage } from '../../../domain/pipeline'
import type { Prospect } from '../../../types/crm'

const props = withDefaults(defineProps<{
  item: Prospect
  highlight: boolean
  compact?: boolean
}>(), { compact: false })

const emit = defineEmits<{
  moveNext: [item: Prospect]
  movePrev: [item: Prospect]
  markLost: [item: Prospect]
  markWon: [item: Prospect]
  viewDetail: [item: Prospect]
}>()

const nxt = () => nextStage(props.item.status)
const prv = () => previousStage(props.item.status)
const isTerminal = () => props.item.status === 'WON' || props.item.status === 'LOST'
const isNewLead = () => props.item.status === 'NEW_LEAD'
const isNegotiation = () => props.item.status === 'NEGOTIATION'
</script>

<template>
  <article
    :id="`prospect-card-${item.id}`"
    class="pipeline-card"
    :class="{ 'pipeline-card--highlight': highlight, 'pipeline-card--compact': compact }"
  >
    <!-- ── COMPACT (DESKTOP) ── -->
    <template v-if="compact">
      <div class="cpt-body">
        <div class="cpt-row-top">
          <div
            class="cpt-avatar"
            :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg }"
          >{{ item.placeName.charAt(0).toUpperCase() }}</div>
          <div class="cpt-info">
            <span class="cpt-name">{{ item.placeName }}</span>
            <span class="cpt-industry">{{ item.industryGroup }}</span>
          </div>
          <div class="cpt-flow">
            <template v-if="prv()">
              <span class="cpt-flow-prev" :title="`Previous: ${stageLabel(prv()!)}`">{{ stageLabel(prv()!) }}</span>
              <i class="pi pi-arrow-right cpt-flow-arr" />
            </template>
            <span class="cpt-flow-current" :style="{ color: stageTone(item.status).fg }">{{ stageLabel(item.status) }}</span>
            <template v-if="nxt()">
              <i class="pi pi-arrow-right cpt-flow-arr" />
              <span class="cpt-flow-next" :title="`Next: ${stageLabel(nxt()!)}`">{{ stageLabel(nxt()!) }}</span>
            </template>
          </div>
        </div>
        <div class="cpt-actions">
          <button
            v-if="prv() && !isTerminal()"
            class="pact pact-back"
            :title="'Move this prospect one stage backward'"
            @click="emit('movePrev', item)"
          >
            <i class="pi pi-arrow-left" />
            <span>Back to {{ stageLabel(prv()!) }}</span>
          </button>
          <button
            class="pact pact-detail"
            @click="emit('viewDetail', item)"
          >
            <i class="pi pi-eye" />
            <span>View Detail</span>
          </button>
          <button
            v-if="!isTerminal() && nxt() && !isNegotiation()"
            class="pact pact-next"
            :title="'Advance this prospect to the next pipeline stage'"
            @click="emit('moveNext', item)"
          >
            <span>{{ isNewLead() ? 'Start Progress' : 'Move to' }} {{ stageLabel(nxt()!) }}</span>
            <i class="pi pi-arrow-right" />
          </button>
          <button
            v-if="isNegotiation()"
            class="pact pact-won"
            @click="emit('markWon', item)"
          >
            <i class="pi pi-check" />
            <span>Mark as Won</span>
          </button>
        </div>
      </div>
      <div v-if="!isTerminal()" class="cpt-lost-row">
        <button
          class="pact pact-lost"
          @click="emit('markLost', item)"
        >
          <i class="pi pi-times" />
          <span>Mark as Lost</span>
        </button>
      </div>
    </template>

    <!-- ── DEFAULT (MOBILE) ── -->
    <template v-else>
      <div class="card-header">
        <div
          class="card-avatar"
          :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg }"
        >{{ item.placeName.charAt(0).toUpperCase() }}</div>
        <div class="card-identity">
          <h3 class="card-name">{{ item.placeName }}</h3>
          <span class="card-industry">{{ item.industryGroup }}</span>
        </div>
        <span
          class="card-stage-badge"
          :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg, borderColor: stageTone(item.status).border }"
        >{{ stageLabel(item.status) }}</span>
      </div>

      <div class="card-body">
        <p class="card-address"><i class="pi pi-map-marker" /> {{ item.formattedAddress }}</p>
        <p class="card-owner"><i class="pi pi-user" /> {{ item.assignedSalesExecutive }}</p>
      </div>

      <div class="cm-flow">
        <template v-if="prv()">
          <span class="cm-flow-prev">{{ stageLabel(prv()!) }}</span>
          <i class="pi pi-arrow-right cm-flow-arr" />
        </template>
        <span class="cm-flow-current" :style="{ color: stageTone(item.status).fg }">{{ stageLabel(item.status) }}</span>
        <template v-if="nxt()">
          <i class="pi pi-arrow-right cm-flow-arr" />
          <span class="cm-flow-next">{{ stageLabel(nxt()!) }}</span>
        </template>
      </div>

      <div v-if="isTerminal()" class="card-progress">
        <span v-if="item.status === 'WON'" class="progress-terminal won">Waiting for Admin review</span>
        <span v-else class="progress-terminal lost">Sales process closed</span>
      </div>

      <div class="cm-actions">
        <div class="cm-actions-row">
          <button class="pact pact-detail" @click="emit('viewDetail', item)">
            <i class="pi pi-eye" />
            <span>View Detail</span>
          </button>
          <template v-if="!isTerminal()">
            <button
              v-if="nxt() && !isNegotiation()"
              class="pact pact-next"
              @click="emit('moveNext', item)"
            >
              <span>{{ isNewLead() ? 'Start Progress' : 'Move to' }} {{ stageLabel(nxt()!) }}</span>
              <i class="pi pi-arrow-right" />
            </button>
            <button
              v-if="isNegotiation()"
              class="pact pact-won"
              @click="emit('markWon', item)"
            >
              <i class="pi pi-check" />
              <span>Mark as Won</span>
            </button>
          </template>
        </div>
        <div v-if="!isTerminal()" class="cm-actions-row">
          <button
            v-if="prv()"
            class="pact pact-back"
            @click="emit('movePrev', item)"
          >
            <i class="pi pi-arrow-left" />
            <span>Back to {{ stageLabel(prv()!) }}</span>
          </button>
          <button
            class="pact pact-lost"
            @click="emit('markLost', item)"
          >
            <i class="pi pi-times" />
            <span>Mark as Lost</span>
          </button>
        </div>
      </div>
    </template>
  </article>
</template>

<style scoped>
.pipeline-card {
  background: #fff;
  border: 1px solid #e8edf3;
  border-radius: 16px;
  padding: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.pipeline-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.07); }
.pipeline-card--highlight {
  border-color: #2563eb !important;
  box-shadow: 0 0 0 2px rgba(37,99,235,0.18), 0 2px 8px rgba(37,99,235,0.1) !important;
  animation: pulse-highlight 0.6s ease;
}
@keyframes pulse-highlight {
  0% { box-shadow: 0 0 0 4px rgba(37,99,235,0.25); }
  100% { box-shadow: 0 0 0 2px rgba(37,99,235,0.18), 0 2px 8px rgba(37,99,235,0.1); }
}

/* ── COMPACT (DESKTOP) ── */
.pipeline-card--compact {
  padding: 0;
  border-radius: 12px;
}
.cpt-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 12px;
}
.cpt-row-top {
  display: flex;
  align-items: center;
  gap: 10px;
}
.cpt-avatar {
  width: 30px; height: 30px;
  border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 0.7rem;
  flex-shrink: 0;
}
.cpt-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}
.cpt-name {
  font-size: 0.78rem; font-weight: 700; color: #1e293b;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.cpt-industry {
  font-size: 0.6rem; color: #94a3b8;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.cpt-flow {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.cpt-flow-prev, .cpt-flow-next {
  font-size: 0.6rem;
  font-weight: 500;
  color: #94a3b8;
  white-space: nowrap;
}
.cpt-flow-current {
  font-size: 0.65rem;
  font-weight: 800;
  white-space: nowrap;
  padding: 1px 6px;
  border-radius: 4px;
  background: #f1f5f9;
}
.cpt-flow-arr {
  font-size: 0.45rem;
  color: #cbd5e1;
}

/* ── COMPACT ACTIONS ── */
.cpt-actions {
  display: flex;
  gap: 6px;
}
.cpt-lost-row {
  display: flex;
  justify-content: flex-end;
  padding: 0 12px 6px;
}

/* ── SHARED ACTION BUTTONS ── */
.pact {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 0.68rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: background 0.12s, transform 0.1s;
  white-space: nowrap;
  min-height: 36px;
  font-family: inherit;
  -webkit-tap-highlight-color: transparent;
}
.pact:active { transform: scale(0.97); }
.pact i { font-size: 0.65rem; }
.pact-back {
  background: #f1f5f9;
  color: #475569;
}
.pact-back:hover { background: #e2e8f0; }
.pact-detail {
  background: #fff;
  color: #475569;
  border: 1px solid #e2e8f0;
}
.pact-detail:hover { background: #f8fafc; }
.pact-next {
  background: #2563eb;
  color: #fff;
}
.pact-next:hover { background: #1d4ed8; }
.pact-won {
  background: #16a34a;
  color: #fff;
}
.pact-won:hover { background: #15803d; }
.pact-lost {
  background: #fff;
  color: #dc2626;
  border: 1px solid #fecaca;
  font-size: 0.65rem;
}
.pact-lost:hover { background: #fef2f2; }

/* ── DEFAULT (MOBILE) ── */
.card-header {
  display: flex; align-items: center; gap: 10px; margin-bottom: 10px;
}
.card-avatar {
  width: 36px; height: 36px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 0.85rem; flex-shrink: 0;
}
.card-identity { flex: 1; min-width: 0; }
.card-name {
  margin: 0; font-size: 0.9rem; font-weight: 700; color: #1e293b;
  line-height: 1.3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.card-industry { font-size: 0.7rem; color: #64748b; font-weight: 500; }
.card-stage-badge {
  font-size: 0.62rem; font-weight: 700; padding: 3px 8px; border-radius: 8px;
  border: 1px solid; white-space: nowrap; text-transform: uppercase;
  letter-spacing: 0.03em;
}
.card-body { display: flex; flex-direction: column; gap: 4px; }
.card-address {
  margin: 0; font-size: 0.75rem; color: #64748b; line-height: 1.4;
}
.card-address i { font-size: 0.7rem; margin-right: 2px; }
.card-owner {
  margin: 0; font-size: 0.7rem; color: #94a3b8; font-weight: 500;
  display: flex; align-items: center; gap: 3px;
}
.card-owner i { font-size: 0.6rem; color: #94a3b8; }
.card-progress {
  margin-top: 8px; padding: 8px 10px;
  background: #f8fafc; border-radius: 10px;
}
.progress-terminal { font-size: 0.7rem; font-weight: 600; }
.progress-terminal.won { color: #2e7d32; }
.progress-terminal.lost { color: #c62828; }

/* ── MOBILE DIRECTION FLOW ── */
.cm-flow {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 0 4px;
  padding: 6px 10px;
  background: #f8fafc;
  border-radius: 10px;
}
.cm-flow-prev, .cm-flow-next {
  font-size: 0.68rem;
  font-weight: 500;
  color: #64748b;
}
.cm-flow-current {
  font-size: 0.72rem;
  font-weight: 800;
  white-space: nowrap;
}
.cm-flow-arr {
  font-size: 0.5rem;
  color: #94a3b8;
}

/* ── MOBILE ACTIONS ── */
.cm-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f1f5f9;
}
.cm-actions-row {
  display: flex;
  gap: 8px;
}
.cm-actions-row .pact { flex: 1; }

/* ── RESPONSIVE ── */
@media (max-width: 1023px) {
  .pipeline-card--compact .cpt-body,
  .pipeline-card--compact .cpt-lost-row {
    display: none;
  }
}
</style>
