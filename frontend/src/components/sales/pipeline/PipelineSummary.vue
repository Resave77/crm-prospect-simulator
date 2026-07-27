<script setup lang="ts">
import type { ProspectStatus } from '../../../types/crm'
import { stageTone, stageLabel } from './stageColors'

interface SummaryItem {
  stage: ProspectStatus
  count: number
}

defineProps<{
  items: SummaryItem[]
  activeStage: ProspectStatus | null
}>()

const emit = defineEmits<{
  select: [stage: ProspectStatus]
}>()
</script>

<template>
  <div class="pipeline-summary">
    <button
      v-for="item in items"
      :key="item.stage"
      class="summary-item"
      :class="{ active: activeStage === item.stage }"
      :style="{
        '--stage-bg': stageTone(item.stage).bg,
        '--stage-fg': stageTone(item.stage).fg,
        '--stage-border': stageTone(item.stage).border,
      }"
      @click="emit('select', item.stage)"
    >
      <span class="summary-count">{{ item.count }}</span>
      <span class="summary-label">{{ stageLabel(item.stage) }}</span>
    </button>
  </div>
</template>

<style scoped>
.pipeline-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  padding: 12px 0;
}

@media (max-width: 380px) {
  .pipeline-summary {
    grid-template-columns: repeat(2, 1fr);
  }
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 6px;
  background: var(--stage-bg);
  border: 1.5px solid var(--stage-border);
  border-radius: 12px;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.1s;
  -webkit-tap-highlight-color: transparent;
}

.summary-item:active {
  transform: scale(0.97);
}

.summary-item.active {
  border-color: var(--stage-fg);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--stage-fg) 20%, transparent);
}

.summary-count {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--stage-fg);
  line-height: 1.2;
}

.summary-label {
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--stage-fg);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  text-align: center;
  line-height: 1.2;
}
</style>
