<script setup lang="ts">
import { computed, ref } from 'vue'
import { stageTone, stageLabel } from './stageColors'
import { nextStage, previousStage } from '../../../domain/pipeline'
import type { Prospect } from '../../../types/crm'

const props = withDefaults(defineProps<{
  item: Prospect
  highlight: boolean
  compact?: boolean
  newLeadAction?: 'sales' | 'admin'
}>(), { compact: false, newLeadAction: 'sales' })

const expanded = ref(false)

const emit = defineEmits<{
  moveNext: [item: Prospect]
  movePrev: [item: Prospect]
  markLost: [item: Prospect]
  markWon: [item: Prospect]
  deleteLost: [item: Prospect]
  cancelLostDeletion: [item: Prospect]
  viewDetail: [item: Prospect]
}>()

const nxt = () => nextStage(props.item.status)
const prv = () => previousStage(props.item.status)
const isTerminal = () => props.item.status === 'WON' || props.item.status === 'LOST'
const isNewLead = () => props.item.status === 'NEW_LEAD'
const isNegotiation = () => props.item.status === 'NEGOTIATION'
function cardTone() {
  if (props.item.status === 'LOST') return 'pipeline-card--lost'
  if (props.item.status === 'WON' || props.item.status === 'CONVERTED') return 'pipeline-card--won'
  if (props.item.status === 'NEW_LEAD') return 'pipeline-card--unvisited'
  if (props.item.status === 'NEGOTIATION' || props.item.status === 'PROPOSAL_SENT') return 'pipeline-card--attention'
  return 'pipeline-card--progress'
}
const leadInitials = computed(() => {
  const names = props.item.placeName.trim().split(/\s+/).filter(Boolean)
  return (names.slice(0, 2).map((name) => name.charAt(0).toUpperCase()).join('') || '?')
})
const leadCode = computed(() => {
  const numericPart = props.item.id.replace(/\D/g, '').slice(-3)
  const fallback = props.item.id.replace(/[^a-z0-9]/gi, '').slice(0, 3).toUpperCase().padEnd(3, '0')
  return `CTID-${numericPart ? numericPart.padStart(3, '0') : fallback}`
})
const leadDate = computed(() => {
  if (!props.item.updatedAt) return '—'
  const date = new Date(props.item.updatedAt)
  return Number.isNaN(date.getTime()) ? '—' : new Intl.DateTimeFormat('en-GB', { day: '2-digit', month: 'short' }).format(date)
})
function openDetail() {
  if (isNewLead()) {
    expanded.value = !expanded.value
    return
  }
  emit('viewDetail', props.item)
}
function onCardKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); openDetail() }
}
</script>

<template>
  <article
    :id="`prospect-card-${item.id}`"
    class="pipeline-card"
    :class="[cardTone(), { 'pipeline-card--highlight': highlight, 'pipeline-card--compact': compact, 'pipeline-card--new-lead': isNewLead() }]"
    :role="isNewLead() ? 'button' : 'link'"
    tabindex="0"
    :aria-expanded="isNewLead() ? expanded : undefined"
    @click="openDetail"
    @keydown="onCardKeydown"
  >
    <!-- ── COLLAPSED NEW LEAD ── -->
    <template v-if="isNewLead()">
      <div class="nl-preview">
        <div class="nl-preview-topline">
          <span class="nl-stage-label">New lead</span>
          <button
            class="nl-menu-button"
            type="button"
            title="Open prospect detail"
            aria-label="Open prospect detail"
            @click.stop="emit('viewDetail', item)"
          >⋮</button>
        </div>

        <div class="nl-identity-row">
          <div class="nl-avatar-stack" aria-hidden="true">
            <span class="nl-avatar nl-avatar-primary">{{ leadInitials.charAt(0) }}</span>
            <span v-if="leadInitials.length > 1" class="nl-avatar nl-avatar-secondary">{{ leadInitials.charAt(1) }}</span>
          </div>
          <div class="nl-name-block">
            <span class="nl-name">{{ item.placeName }}</span>
            <span class="nl-helper">{{ expanded ? 'Click to hide details' : 'Click to view prospect' }}</span>
          </div>
          <i class="pi pi-bell nl-activity-icon" title="New lead activity" />
        </div>

        <div class="nl-preview-meta">
          <span class="nl-code"><i class="pi pi-tag" /> {{ leadCode }}</span>
          <time class="nl-date" :datetime="item.updatedAt">{{ leadDate }}</time>
        </div>
      </div>

      <!-- ── EXPANDED NEW LEAD ── -->
      <div v-if="expanded" class="nl-details" @click.stop>
        <div class="nl-details-heading">
          <div>
            <span class="nl-details-kicker">Prospect details</span>
            <strong>{{ item.placeName }}</strong>
          </div>
          <span class="nl-status">{{ stageLabel(item.status) }}</span>
        </div>

        <div class="nl-detail-grid">
          <div class="nl-detail-item nl-detail-item-wide">
            <span><i class="pi pi-map-marker" /> Address</span>
            <strong>{{ item.formattedAddress || '—' }}</strong>
          </div>
          <div class="nl-detail-item">
            <span><i class="pi pi-briefcase" /> Category</span>
            <strong>{{ item.placeCategory || item.industryGroup || '—' }}</strong>
          </div>
          <div class="nl-detail-item">
            <span><i class="pi pi-user" /> Assigned to</span>
            <strong>{{ item.assignedSalesExecutive || 'Assigned to you' }}</strong>
          </div>
          <div v-if="item.phoneNumber" class="nl-detail-item">
            <span><i class="pi pi-phone" /> Phone</span>
            <strong>{{ item.phoneNumber }}</strong>
          </div>
        </div>

        <p v-if="item.visitNotes || item.followUpNotes" class="nl-notes">
          {{ item.followUpNotes || item.visitNotes }}
        </p>

        <div class="nl-detail-actions">
          <button class="nl-open-detail" type="button" @click.stop="emit('viewDetail', item)">
            <i class="pi pi-eye" /> {{ newLeadAction === 'admin' ? 'Open ticketing' : 'Buka detail prospect' }}
          </button>
          <button v-if="newLeadAction === 'sales' && nxt()" class="nl-start-progress" type="button" @click.stop="emit('moveNext', item)">
            Mulai progress <i class="pi pi-arrow-right" />
          </button>
        </div>
      </div>
    </template>

    <!-- ── COMPACT (DESKTOP) ── -->
    <template v-else-if="compact">
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
        </div>
        <p class="cpt-address"><i class="pi pi-map-marker" /> <span>{{ item.formattedAddress }}</span></p>
        <div class="cpt-meta">
          <span><i class="pi pi-user" /> {{ item.assignedSalesExecutive || 'Assigned to you' }}</span>
          <span
            class="cpt-stage"
            :style="{ background: stageTone(item.status).bg, color: stageTone(item.status).fg, borderColor: stageTone(item.status).border }"
          >{{ stageLabel(item.status) }}</span>
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
        <div class="cpt-actions">
          <button
            v-if="item.status === 'LOST' && !item.deletionRequested"
            class="pact pact-delete-lost"
            title="Request deletion to admin"
            aria-label="Request deletion to admin"
            @click.stop="emit('deleteLost', item)"
          ><i class="pi pi-trash" /></button>
          <span v-else-if="item.status === 'LOST' && item.deletionRequested" class="deletion-pending"><i class="pi pi-clock" title="Waiting for admin approval" /><button type="button" title="Cancel deletion request" aria-label="Cancel deletion request" @click.stop="emit('cancelLostDeletion', item)"><i class="pi pi-undo" /></button></span>
          <button
            v-if="prv() && !isTerminal()"
            class="pact pact-back"
            :title="'Move this prospect one stage backward'"
            @click.stop="emit('movePrev', item)"
          >
            <i class="pi pi-arrow-left" />
            <span>Back to {{ stageLabel(prv()!) }}</span>
          </button>
          <button
            class="pact pact-detail"
            @click.stop="emit('viewDetail', item)"
          >
            <i class="pi pi-eye" />
            <span>View Detail</span>
          </button>
          <button
            v-if="!isTerminal() && nxt() && !isNegotiation()"
            class="pact pact-next"
            :title="'Advance this prospect to the next pipeline stage'"
            @click.stop="emit('moveNext', item)"
          >
            <span>{{ isNewLead() ? 'Start Progress' : 'Move to' }} {{ stageLabel(nxt()!) }}</span>
            <i class="pi pi-arrow-right" />
          </button>
          <button
            v-if="isNegotiation()"
            class="pact pact-won"
            @click.stop="emit('markWon', item)"
          >
            <i class="pi pi-check" />
            <span>Mark as Won</span>
          </button>
        </div>
      </div>
      <div v-if="!isTerminal()" class="cpt-lost-row">
        <button
          class="pact pact-lost"
          @click.stop="emit('markLost', item)"
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
          <button v-if="item.status === 'LOST' && !item.deletionRequested" class="pact pact-delete-lost" title="Request deletion to admin" aria-label="Request deletion to admin" @click.stop="emit('deleteLost', item)"><i class="pi pi-trash" /></button>
          <span v-else-if="item.status === 'LOST' && item.deletionRequested" class="deletion-pending"><i class="pi pi-clock" title="Waiting for admin approval" /><button type="button" title="Cancel deletion request" aria-label="Cancel deletion request" @click.stop="emit('cancelLostDeletion', item)"><i class="pi pi-undo" /></button></span>
          <button class="pact pact-detail" @click.stop="emit('viewDetail', item)">
            <i class="pi pi-eye" />
            <span>View Detail</span>
          </button>
          <template v-if="!isTerminal()">
            <button
              v-if="nxt() && !isNegotiation()"
              class="pact pact-next"
              @click.stop="emit('moveNext', item)"
            >
              <span>{{ isNewLead() ? 'Start Progress' : 'Move to' }} {{ stageLabel(nxt()!) }}</span>
              <i class="pi pi-arrow-right" />
            </button>
            <button
              v-if="isNegotiation()"
              class="pact pact-won"
              @click.stop="emit('markWon', item)"
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
            @click.stop="emit('movePrev', item)"
          >
            <i class="pi pi-arrow-left" />
            <span>Back to {{ stageLabel(prv()!) }}</span>
          </button>
          <button
            class="pact pact-lost"
            @click.stop="emit('markLost', item)"
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
.pact-delete-lost, .deletion-pending { margin-left:auto; flex:0 0 auto !important; }
.pact-delete-lost { width:28px; min-width:28px; padding:.3rem !important; justify-content:center; color:#b4232f; }
.deletion-pending { display:inline-flex; align-items:center; gap:.35rem; color:#d97706; font-size:.75rem; }.deletion-pending button { display:grid; place-items:center; width:24px; height:24px; padding:0; border:1px solid #fde68a; border-radius:7px; background:#fffbeb; color:#b45309; cursor:pointer; }.deletion-pending button:hover { background:#fef3c7; }
.pipeline-card {
  background: #fff;
  border: 1px solid #e8edf3;
  border-radius: 16px;
  padding: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.pipeline-card--lost { border-color: #fecaca; background: linear-gradient(180deg, #fff 0%, #fff7f7 100%); }
.pipeline-card--won { border-color: #bbf7d0; background: linear-gradient(180deg, #fff 0%, #f7fff9 100%); }
.pipeline-card--unvisited { border-color: #cbd5e1; background: linear-gradient(180deg, #fff 0%, #f8fafc 100%); }
.pipeline-card--progress { border-color: #bfdbfe; background: linear-gradient(180deg, #fff 0%, #f8fbff 100%); }
.pipeline-card--attention { border-color: #fed7aa; background: linear-gradient(180deg, #fff 0%, #fffaf3 100%); }
.pipeline-card--lost::before,
.pipeline-card--won::before,
.pipeline-card--unvisited::before,
.pipeline-card--progress::before,
.pipeline-card--attention::before { content: ''; display: block; height: 3px; margin: -14px -14px 12px; border-radius: 16px 16px 0 0; }
.pipeline-card--lost::before { background: #ef4444; }
.pipeline-card--won::before { background: #22c55e; }
.pipeline-card--unvisited::before { background: #94a3b8; }
.pipeline-card--progress::before { background: #3b82f6; }
.pipeline-card--attention::before { background: #f59e0b; }
.pipeline-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.07); }
.pipeline-card[role="link"] { cursor: pointer; }
.pipeline-card[role="link"]:focus-visible { outline: 3px solid rgba(37,99,235,.3); outline-offset: 2px; }
.pipeline-card[role="link"]:hover { border-color: #cbd5e1; box-shadow: 0 5px 14px rgba(15,23,42,.09); }
.pipeline-card--highlight {
  border-color: #e63946 !important;
  box-shadow: 0 0 0 2px rgba(230,57,70,0.18), 0 2px 8px rgba(230,57,70,0.1) !important;
  animation: pulse-highlight 0.6s ease;
}
@keyframes pulse-highlight {
  0% { box-shadow: 0 0 0 4px rgba(230,57,70,0.25); }
  100% { box-shadow: 0 0 0 2px rgba(230,57,70,0.18), 0 2px 8px rgba(230,57,70,0.1); }
}

/* ── COMPACT (DESKTOP) ── */
.pipeline-card--compact {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  padding: 0;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  overflow: hidden;
}
.pipeline-card--compact::before { margin: 0; border-radius: 8px 8px 0 0; }
.cpt-body {
  box-sizing: border-box;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
  padding: 0.62rem;
}
.cpt-row-top {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
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
  font-size: 0.72rem; font-weight: 750; color: #0f172a;
  line-height: 1.35;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.cpt-industry {
  width: fit-content;
  max-width: 100%;
  padding: 0.13rem 0.4rem;
  border-radius: 5px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 0.5rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.cpt-address {
  min-width: 0;
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 0.28rem;
  margin: 0;
  color: #64748b;
  font-size: 0.58rem;
  line-height: 1.45;
}
.cpt-address i {
  margin-top: 0.1rem;
  color: #94a3b8;
  font-size: 0.58rem;
}
.cpt-address span {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.cpt-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem;
  min-width: 0;
  color: #94a3b8;
  font-size: 0.56rem;
}
.cpt-meta > span:first-child {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cpt-meta i {
  font-size: 0.56rem;
}
.cpt-stage {
  flex: 0 0 auto;
  max-width: 46%;
  overflow: hidden;
  padding: 0.12rem 0.38rem;
  border: 1px solid;
  border-radius: 999px;
  font-size: 0.52rem;
  font-weight: 800;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}
.cpt-flow {
  display: flex;
  align-items: center;
  gap: 0.24rem;
  min-width: 0;
  padding: 0.34rem 0.45rem;
  border-radius: 8px;
  background: #f8fafc;
  overflow: hidden;
}
.cpt-flow-prev, .cpt-flow-next {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.56rem;
  font-weight: 500;
  color: #94a3b8;
  white-space: nowrap;
}
.cpt-flow-current {
  flex: 0 0 auto;
  font-size: 0.58rem;
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
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 0.34rem;
  padding-top: 0.38rem;
  border-top: 1px solid #edf1f5;
}
.cpt-lost-row {
  box-sizing: border-box;
  min-width: 0;
  display: flex;
  justify-content: flex-end;
  padding: 0 0.62rem 0.54rem;
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
.pipeline-card--compact .pact {
  flex: 1 1 0;
  min-width: 0;
  min-height: 30px;
  padding: 0.36rem 0.46rem;
  border-radius: 7px;
  font-size: 0.56rem;
}
.pipeline-card--compact .pact i { font-size: 0.58rem; }
.pipeline-card--compact .pact-detail { flex-basis: 100%; }
.pipeline-card--compact .pact-lost { min-height: 28px; }
.pipeline-card--compact .pact span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  background: #e63946;
  color: #fff;
}
.pact-next:hover { background: #d62839; }
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

/* ── NEW LEAD SUMMARY ── */
.pipeline-card--new-lead {
  padding: 0;
  overflow: hidden;
  border-color: #e7e9ee;
  background: #fff;
}
.pipeline-card--new-lead::before {
  margin: 0;
  border-radius: 15px 15px 0 0;
  background: #ef4444;
}
.pipeline-card--new-lead:hover { border-color: #cbd5e1; }
.nl-preview { padding: 0.72rem 0.78rem 0.68rem; }
.nl-preview-topline,
.nl-identity-row,
.nl-preview-meta,
.nl-details-heading,
.nl-detail-actions {
  display: flex;
  align-items: center;
}
.nl-preview-topline { justify-content: space-between; margin-bottom: 0.65rem; }
.nl-stage-label {
  color: #ef4444;
  font-size: 0.57rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.nl-menu-button {
  display: grid;
  place-items: center;
  width: 1.45rem;
  height: 1.45rem;
  padding: 0 0 0.35rem;
  border: 0;
  border-radius: 0.45rem;
  background: transparent;
  color: #94a3b8;
  font-family: inherit;
  font-size: 1.1rem;
  line-height: 1;
  cursor: pointer;
}
.nl-menu-button:hover { background: #f1f5f9; color: #475569; }
.nl-identity-row { gap: 0.56rem; min-width: 0; }
.nl-avatar-stack { display: flex; min-width: 2rem; }
.nl-avatar {
  display: grid;
  place-items: center;
  width: 1.78rem;
  height: 1.78rem;
  border: 2px solid #fff;
  border-radius: 999px;
  color: #fff;
  font-size: 0.6rem;
  font-weight: 800;
}
.nl-avatar-primary { background: #3548a8; z-index: 1; }
.nl-avatar-secondary { margin-left: -0.55rem; background: #ef6b73; }
.nl-name-block { display: grid; min-width: 0; flex: 1; gap: 0.08rem; }
.nl-name {
  overflow: hidden;
  color: #172033;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nl-helper {
  color: #a0a7b4;
  font-size: 0.54rem;
  font-weight: 500;
}
.nl-activity-icon {
  display: grid;
  place-items: center;
  width: 1.6rem;
  height: 1.6rem;
  flex: 0 0 auto;
  border-radius: 0.5rem;
  background: #fff0f1;
  color: #e63946;
  font-size: 0.7rem;
}
.nl-preview-meta {
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.78rem;
  padding-top: 0.58rem;
  border-top: 1px solid #f0f1f4;
}
.nl-code,
.nl-date {
  color: #8b93a1;
  font-size: 0.55rem;
  font-weight: 600;
}
.nl-code { display: inline-flex; align-items: center; gap: 0.24rem; }
.nl-code i { color: #b7beca; font-size: 0.55rem; }
.nl-date { white-space: nowrap; }
.nl-details {
  padding: 0.72rem 0.78rem 0.78rem;
  border-top: 1px solid #edf0f4;
  background: #fbfcfe;
}
.nl-details-heading { justify-content: space-between; gap: 0.5rem; margin-bottom: 0.62rem; }
.nl-details-heading > div { display: grid; min-width: 0; gap: 0.12rem; }
.nl-details-kicker {
  color: #94a3b8;
  font-size: 0.53rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
.nl-details-heading strong {
  overflow: hidden;
  color: #1e293b;
  font-size: 0.7rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nl-status {
  flex: 0 0 auto;
  padding: 0.18rem 0.38rem;
  border: 1px solid #fecdd3;
  border-radius: 999px;
  background: #fff1f2;
  color: #be123c;
  font-size: 0.49rem;
  font-weight: 800;
  text-transform: uppercase;
}
.nl-detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.48rem; }
.nl-detail-item { display: grid; min-width: 0; gap: 0.18rem; }
.nl-detail-item-wide { grid-column: 1 / -1; }
.nl-detail-item span {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: #9aa3b1;
  font-size: 0.51rem;
  font-weight: 600;
}
.nl-detail-item span i { font-size: 0.5rem; }
.nl-detail-item strong {
  overflow: hidden;
  color: #475569;
  font-size: 0.59rem;
  font-weight: 650;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nl-notes {
  margin: 0.58rem 0 0;
  padding: 0.48rem 0.55rem;
  border-radius: 0.45rem;
  background: #f1f5f9;
  color: #64748b;
  font-size: 0.58rem;
  line-height: 1.4;
}
.nl-detail-actions { gap: 0.4rem; margin-top: 0.68rem; }
.nl-detail-actions button {
  min-width: 0;
  min-height: 1.8rem;
  padding: 0.35rem 0.5rem;
  border-radius: 0.45rem;
  font-family: inherit;
  font-size: 0.55rem;
  font-weight: 700;
  cursor: pointer;
}
.nl-open-detail { flex: 1; border: 1px solid #e2e8f0; background: #fff; color: #475569; }
.nl-open-detail:hover { background: #f8fafc; }
.nl-start-progress { flex: 1; border: 1px solid #e63946; background: #e63946; color: #fff; }
.nl-start-progress:hover { background: #d62839; }

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
@media (max-width: 768px) {
  .pipeline-card--compact .cpt-body,
  .pipeline-card--compact .cpt-lost-row {
    display: none;
  }
}
</style>
