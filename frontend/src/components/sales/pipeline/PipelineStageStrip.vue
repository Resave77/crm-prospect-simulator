<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ProspectStatus } from '../../../types/crm'
import { stageLabel } from './stageColors'

type StageFilter = ProspectStatus | 'ALL'

const props = defineProps<{
  allCount: number
  counts: Record<ProspectStatus, number>
  activeStage: StageFilter
}>()

const emit = defineEmits<{
  select: [stage: StageFilter]
}>()

const ALL_STAGES: ProspectStatus[] = [
  'NEW_LEAD', 'CONTACTED', 'INTERESTED', 'QUALIFIED',
  'PROPOSAL_SENT', 'NEGOTIATION', 'WON', 'LOST',
]

const stageScroller = ref<HTMLElement | null>(null)
const chipRefs = new Map<string, HTMLElement>()
const canScrollLeft = ref(false)
const canScrollRight = ref(false)
const hintVisible = ref(true)

let pointerDown = false
let dragging = false
let startX = 0
let startScroll = 0
const DRAG_THRESHOLD = 10

function setStageChipRef(stage: string, el: any) {
  if (el) chipRefs.set(stage, el as HTMLElement)
  else chipRefs.delete(stage)
}

function updateScrollState() {
  const el = stageScroller.value
  if (!el) return
  const eps = 2
  canScrollLeft.value = el.scrollLeft > eps
  canScrollRight.value = el.scrollLeft + el.clientWidth < el.scrollWidth - eps
}

function scrollStages(direction: -1 | 1) {
  const el = stageScroller.value
  if (!el) return
  const amount = Math.max(180, Math.round(el.clientWidth * 0.72))
  el.scrollBy({ left: direction * amount, behavior: 'smooth' })
}

function onPointerDown(e: PointerEvent) {
  if (e.button !== 0) return
  pointerDown = true
  dragging = false
  startX = e.clientX
  startScroll = stageScroller.value?.scrollLeft ?? 0
  window.addEventListener('pointermove', onPointerMove, { passive: false })
  window.addEventListener('pointerup', onPointerUp)
  window.addEventListener('pointercancel', onPointerUp)
}

function onPointerMove(e: PointerEvent) {
  if (!pointerDown) return
  const dx = e.clientX - startX
  if (!dragging && Math.abs(dx) >= DRAG_THRESHOLD) {
    dragging = true
    const el = stageScroller.value
    if (el) {
      el.style.cursor = 'grabbing'
      el.style.scrollBehavior = 'auto'
    }
  }
  if (dragging) {
    const el = stageScroller.value
    if (el) el.scrollLeft = startScroll - dx
    e.preventDefault()
  }
}

function onPointerUp() {
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerUp)

  const el = stageScroller.value
  if (el) {
    el.style.cursor = ''
    el.style.scrollBehavior = ''
  }

  const wasDragging = dragging
  pointerDown = false
  dragging = false

  if (wasDragging) {
    updateScrollState()
  }
}

function handleChipClick(stage: StageFilter) {
  if (dragging) return
  emit('select', stage)
  hintVisible.value = false
}

async function scrollToActive() {
  await nextTick()
  const chip = chipRefs.get(props.activeStage)
  if (chip) {
    chip.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
  }
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  updateScrollState()
  scrollToActive()
  const el = stageScroller.value
  if (el && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => updateScrollState())
    resizeObserver.observe(el)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerUp)
})

watch(() => props.activeStage, () => scrollToActive())
watch(() => props.allCount, () => nextTick(updateScrollState))
</script>

<template>
  <div class="stage-nav" :class="{ 'can-scroll-left': canScrollLeft, 'can-scroll-right': canScrollRight }">
    <button
      v-if="canScrollLeft"
      class="stage-arrow stage-arrow-left"
      type="button"
      aria-label="Scroll pipeline stages left"
      @click.stop="scrollStages(-1)"
    >
      <i class="pi pi-chevron-left" />
    </button>

    <div
      ref="stageScroller"
      class="stage-scroller"
      aria-label="Pipeline stages"
      role="toolbar"
      @scroll="updateScrollState"
      @pointerdown="onPointerDown"
    >
      <button
        class="stage-chip"
        :class="{ active: activeStage === 'ALL' }"
        type="button"
        :aria-pressed="activeStage === 'ALL'"
        @click="handleChipClick('ALL')"
      >
        <span class="stage-chip-label">All</span>
        <span class="stage-chip-count">{{ allCount }}</span>
      </button>

      <button
        v-for="stage in ALL_STAGES"
        :key="stage"
        :ref="(el) => setStageChipRef(stage, el)"
        class="stage-chip"
        :class="{
          active: activeStage === stage,
          empty: counts[stage] === 0,
          terminal: stage === 'WON' || stage === 'LOST',
        }"
        type="button"
        :aria-pressed="activeStage === stage"
        @click="handleChipClick(stage)"
      >
        <span class="stage-chip-label">{{ stageLabel(stage) }}</span>
        <span class="stage-chip-count">{{ counts[stage] }}</span>
      </button>
    </div>

    <button
      v-if="canScrollRight"
      class="stage-arrow stage-arrow-right"
      type="button"
      aria-label="Scroll pipeline stages right"
      @click.stop="scrollStages(1)"
    >
      <i class="pi pi-chevron-right" />
    </button>
  </div>

  <p v-if="hintVisible && canScrollRight" class="stage-hint">
    <i class="pi pi-arrows-h" />
    Swipe or drag to view all pipeline stages
  </p>
</template>

<style scoped>
.stage-nav {
  position: relative;
  padding: 4px 0 2px;
}

/* Fade edges */
.stage-nav::before,
.stage-nav::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  width: 44px;
  pointer-events: none;
  z-index: 2;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.stage-nav::before {
  left: 0;
  background: linear-gradient(90deg, rgba(248,250,252,1) 0%, rgba(248,250,252,0) 100%);
}

.stage-nav::after {
  right: 0;
  background: linear-gradient(270deg, rgba(248,250,252,1) 0%, rgba(248,250,252,0) 100%);
}

.stage-nav.can-scroll-left::before { opacity: 1; }
.stage-nav.can-scroll-right::after { opacity: 1; }

/* Scroller */
.stage-scroller {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 4px 3rem 4px 0.5rem;
  scroll-snap-type: x proximity;
  cursor: grab;
  user-select: none;
  touch-action: pan-x;
  -webkit-overflow-scrolling: touch;
}

.stage-scroller::-webkit-scrollbar { display: none; }

/* Arrow buttons */
.stage-arrow {
  position: absolute;
  top: 50%;
  z-index: 4;
  width: 32px;
  height: 32px;
  transform: translateY(-50%);
  display: grid;
  place-items: center;
  border: 1px solid #dbe3ef;
  border-radius: 50%;
  background: rgba(255,255,255,0.96);
  color: #e63946;
  box-shadow: 0 2px 8px rgba(15,23,42,0.1);
  cursor: pointer;
  transition: background 0.12s, box-shadow 0.12s, opacity 0.15s;
  font-size: 0.7rem;
}

.stage-arrow:hover {
  background: #fff;
  box-shadow: 0 4px 12px rgba(15,23,42,0.14);
}

.stage-arrow:active {
  transform: translateY(-50%) scale(0.93);
}

.stage-arrow-left { left: 0; }
.stage-arrow-right { right: 0; }

/* Chips */
.stage-chip {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 120px;
  min-height: 48px;
  padding: 10px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 14px;
  background: #fff;
  cursor: pointer;
  font-family: inherit;
  -webkit-tap-highlight-color: transparent;
  transition: background 0.12s, border-color 0.12s, box-shadow 0.15s, transform 0.1s, opacity 0.12s;
  white-space: nowrap;
  scroll-snap-align: start;
}

.stage-chip:active { transform: scale(0.97); }

.stage-chip-label {
  font-size: 0.72rem;
  font-weight: 600;
  color: #475569;
  line-height: 1.2;
}

.stage-chip-count {
  min-width: 24px;
  height: 24px;
  padding: 0 6px;
  display: inline-grid;
  place-items: center;
  border-radius: 9999px;
  background: #fff0f1;
  color: #e63946;
  font-size: 0.68rem;
  font-weight: 800;
  line-height: 1;
  flex-shrink: 0;
}

/* Active */
.stage-chip.active {
  background: #e63946;
  border-color: #e63946;
  color: #fff;
  box-shadow: 0 4px 14px rgba(230,57,70,0.22);
}

.stage-chip.active .stage-chip-label {
  color: #fff;
}

.stage-chip.active .stage-chip-count {
  background: rgba(255,255,255,0.22);
  color: #fff;
}

/* Empty */
.stage-chip.empty:not(.active) {
  opacity: 0.6;
}

.stage-chip.empty:not(.active):hover {
  opacity: 0.8;
}

/* Terminal */
.stage-chip.terminal {
  border-style: dashed;
}

.stage-chip.terminal.active {
  border-style: solid;
}

/* Hint */
.stage-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin: 2px 0 0;
  font-size: 0.65rem;
  color: #94a3b8;
  text-align: center;
}

.stage-hint i {
  color: #f5a0a9;
  font-size: 0.68rem;
}
</style>
