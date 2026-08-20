<script setup lang="ts">
import type { Prospect } from '../../types/crm'

const props = defineProps<{ prospects: Prospect[] }>()

const groups = [
  { key: 'attention', label: 'Needs attention', tone: 'red', statuses: ['NEW_LEAD'] },
  { key: 'progress', label: 'In progress', tone: 'orange', statuses: ['CONTACTED', 'INTERESTED', 'QUALIFIED'] },
  { key: 'late', label: 'Proposal / negotiation', tone: 'yellow', statuses: ['PROPOSAL_SENT', 'NEGOTIATION'] },
  { key: 'healthy', label: 'Won / converted', tone: 'blue', statuses: ['WON', 'CONVERTED'] },
] as const

function count(statuses: readonly string[]) {
  return props.prospects.filter((item) => statuses.includes(item.status)).length
}
</script>

<template>
  <section class="health-card" aria-labelledby="prospect-health-title">
    <div class="health-heading">
      <div>
        <span class="health-eyebrow">Portfolio overview</span>
        <h2 id="prospect-health-title">Prospect Health</h2>
      </div>
      <span class="health-total">{{ prospects.length }} total</span>
    </div>

    <div class="health-track" aria-hidden="true">
      <span v-for="group in groups" :key="group.key" :class="`health-segment health-segment--${group.tone}`" :style="{ flex: Math.max(1, count(group.statuses)) }" />
    </div>

    <div class="health-grid">
      <div v-for="group in groups" :key="group.key" class="health-item">
        <span :class="`health-dot health-dot--${group.tone}`" />
        <span class="health-label">{{ group.label }}</span>
        <strong>{{ count(group.statuses) }}</strong>
      </div>
    </div>
    <p class="health-note"><i class="pi pi-info-circle" /> Visit urgency will appear here when visit schedules are enabled.</p>
  </section>
</template>

<style scoped>
.health-card { align-self: start; height: fit-content; padding: 1.05rem 1.15rem .9rem; border: 1px solid #e5eaf0; border-radius: 16px; background: #fff; box-shadow: 0 1px 3px rgba(15,23,42,.04); }
.health-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.health-eyebrow { color: #94a3b8; font-size: .58rem; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
h2 { margin: .18rem 0 0; color: #0f172a; font-size: 1.02rem; letter-spacing: -.01em; }
.health-total { padding: .3rem .6rem; border: 1px solid #e5eaf0; border-radius: 999px; background: #f8fafc; color: #64748b; font-size: .62rem; font-weight: 750; white-space: nowrap; }
.health-track { display: flex; gap: 4px; height: 8px; margin: 1.05rem 0 .9rem; overflow: hidden; border-radius: 99px; background: #f1f5f9; }
.health-segment { min-width: 8px; border-radius: 99px; }
.health-segment--red { background: #e35d6a; }.health-segment--orange { background: #f59e0b; }.health-segment--yellow { background: #facc15; }.health-segment--blue { background: #3b82f6; }
.health-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: .45rem; }
.health-item { display: grid; grid-template-columns: auto 1fr; gap: .15rem .35rem; align-items: center; min-width: 0; padding: .35rem .25rem; }
.health-item strong { grid-column: 2; color: #0f172a; font-size: 1rem; line-height: 1.1; }.health-label { overflow: hidden; color: #64748b; font-size: .6rem; line-height: 1.2; text-overflow: ellipsis; white-space: nowrap; }
.health-dot { width: 7px; height: 7px; grid-row: span 2; border-radius: 50%; }.health-dot--red { background:#e35d6a; }.health-dot--orange { background:#f59e0b; }.health-dot--yellow { background:#facc15; }.health-dot--blue { background:#3b82f6; }
.health-note { display: flex; gap: .35rem; margin: .75rem 0 0; padding-top: .65rem; border-top: 1px solid #f1f5f9; color: #94a3b8; font-size: .6rem; line-height: 1.4; }.health-note i { margin-top: 1px; }
@media (max-width: 480px) { .health-grid { grid-template-columns: repeat(2, 1fr); row-gap: .35rem; }.health-item { padding: .3rem .15rem; }.health-card { padding: .9rem; } }
</style>
