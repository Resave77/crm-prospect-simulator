<script setup lang="ts">
import type { Prospect, ProspectStatus } from '../../../types/crm'
import { stageTone, stageLabel } from './stageColors'
import { nextStage, previousStage } from '../../../domain/pipeline'

const props = defineProps<{
  item: Prospect
  highlight: boolean
}>()

const emit = defineEmits<{
  moveNext: [item: Prospect]
  movePrev: [item: Prospect]
  markLost: [item: Prospect]
  markWon: [item: Prospect]
  viewDetail: [item: Prospect]
  touchStart: [x: number]
  touchEnd: [item: Prospect, event: TouchEvent]
}>()

const next = () => nextStage(props.item.status)
const prev = () => previousStage(props.item.status)
const isTerminal = () => props.item.status === 'WON' || props.item.status === 'LOST'
const isNegotiation = () => props.item.status === 'NEGOTIATION'

function onTouchStart(e: TouchEvent) {
  emit('touchStart', e.touches[0].clientX)
}
</script>

<template>
  <article
    :id="`prospect-card-${item.id}`"
    class="pipeline-card"
    :class="{ 'pipeline-card--highlight': highlight }"
    @touchstart="onTouchStart"
    @touchend="emit('touchEnd', item, $event)"
  >
    <div class="card-header">
      <div class="card-avatar" :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg }">
        {{ item.placeName.charAt(0).toUpperCase() }}
      </div>
      <div class="card-identity">
        <h3 class="card-name">{{ item.placeName }}</h3>
        <span class="card-industry">{{ item.industryGroup }}</span>
      </div>
      <span
        class="card-stage-badge"
        :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg, borderColor: stageTone(item.status).border }"
      >
        {{ stageLabel(item.status) }}
      </span>
    </div>

    <div class="card-body">
      <p class="card-address"><i class="pi pi-map-marker" /> {{ item.formattedAddress }}</p>
      <p class="card-owner">{{ item.assignedSalesExecutive }}</p>

      <div v-if="!isTerminal()" class="card-progress">
        <span class="progress-label">Current: <strong>{{ stageLabel(item.status) }}</strong></span>
        <span v-if="next()" class="progress-next">Next: <strong>{{ stageLabel(next()!) }}</strong></span>
        <span v-else class="progress-next progress-terminal">Ready for final decision</span>
      </div>
      <div v-else class="card-progress">
        <span v-if="item.status === 'WON'" class="progress-terminal won">Waiting for Admin review</span>
        <span v-else class="progress-terminal lost">Prospect closed as lost</span>
      </div>
    </div>

    <div v-if="!isTerminal()" class="card-actions">
      <button
        v-if="prev()"
        class="action-btn action-secondary"
        :disabled="!prev()"
        @click="emit('movePrev', item)"
      >
        <i class="pi pi-arrow-left" />
        {{ stageLabel(prev()!) }}
      </button>
      <button class="action-btn action-outline" @click="emit('viewDetail', item)">
        <i class="pi pi-eye" />
        Detail
      </button>
      <button
        v-if="isNegotiation()"
        class="action-btn action-success"
        @click="emit('markWon', item)"
      >
        <i class="pi pi-check" />
        Won
      </button>
      <button
        v-else-if="next()"
        class="action-btn action-primary"
        @click="emit('moveNext', item)"
      >
        {{ stageLabel(next()!) }}
        <i class="pi pi-arrow-right" />
      </button>
    </div>
    <div v-else class="card-actions">
      <button class="action-btn action-outline" @click="emit('viewDetail', item)">
        <i class="pi pi-eye" />
        Detail
      </button>
    </div>

    <button class="card-lost-btn" @click="emit('markLost', item)">Lost</button>

    <div class="card-swipe-hint">Swipe to preview next/previous stage</div>
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

.pipeline-card:hover {
  box-shadow: 0 2px 8px rgba(0,0,0,0.07);
}

.pipeline-card--highlight {
  border-color: #2563eb !important;
  box-shadow: 0 0 0 2px rgba(37,99,235,0.18), 0 2px 8px rgba(37,99,235,0.1) !important;
  animation: pulse-highlight 0.6s ease;
}

@keyframes pulse-highlight {
  0% { box-shadow: 0 0 0 4px rgba(37,99,235,0.25); }
  100% { box-shadow: 0 0 0 2px rgba(37,99,235,0.18), 0 2px 8px rgba(37,99,235,0.1); }
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.card-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.85rem;
  flex-shrink: 0;
}

.card-identity {
  flex: 1;
  min-width: 0;
}

.card-name {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: #1e293b;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-industry {
  font-size: 0.7rem;
  color: #64748b;
  font-weight: 500;
}

.card-stage-badge {
  font-size: 0.62rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 8px;
  border: 1px solid;
  white-space: nowrap;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-address {
  margin: 0;
  font-size: 0.75rem;
  color: #64748b;
  line-height: 1.4;
}

.card-address i {
  font-size: 0.7rem;
  margin-right: 2px;
}

.card-owner {
  margin: 0;
  font-size: 0.7rem;
  color: #94a3b8;
  font-weight: 500;
}

.card-progress {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  margin-top: 6px;
  padding: 8px 10px;
  background: #f8fafc;
  border-radius: 10px;
}

.progress-label,
.progress-next {
  font-size: 0.7rem;
  color: #64748b;
}

.progress-label strong,
.progress-next strong {
  color: #1e293b;
  font-weight: 600;
}

.progress-terminal {
  font-size: 0.7rem;
  font-weight: 600;
}

.progress-terminal.won {
  color: #2e7d32;
}

.progress-terminal.lost {
  color: #c62828;
}

.card-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px solid #f1f5f9;
}

.action-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 9px 8px;
  border-radius: 10px;
  font-size: 0.72rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: background 0.12s, transform 0.1s;
  white-space: nowrap;
  min-height: 38px;
  -webkit-tap-highlight-color: transparent;
}

.action-btn:active {
  transform: scale(0.97);
}

.action-btn i {
  font-size: 0.7rem;
}

.action-primary {
  background: #2563eb;
  color: #fff;
}

.action-primary:hover {
  background: #1d4ed8;
}

.action-success {
  background: #16a34a;
  color: #fff;
}

.action-success:hover {
  background: #15803d;
}

.action-secondary {
  background: #f1f5f9;
  color: #475569;
}

.action-secondary:hover {
  background: #e2e8f0;
}

.action-outline {
  background: #fff;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.action-outline:hover {
  background: #f8fafc;
}

.card-lost-btn {
  width: 100%;
  margin-top: 8px;
  padding: 6px;
  background: none;
  border: 1px dashed #fca5a5;
  border-radius: 8px;
  color: #dc2626;
  font-size: 0.65rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.12s;
}

.card-lost-btn:hover {
  background: #fef2f2;
}

.card-swipe-hint {
  margin-top: 6px;
  text-align: center;
  font-size: 0.58rem;
  color: #cbd5e1;
  font-style: italic;
}
</style>
