<script setup lang="ts">
import { computed } from 'vue'
import type { PlaceDetails } from '../../types/crm'
import type { ProspectInitialAnalysis } from '../../types/crm'

const props = defineProps<{
  placeDetails: PlaceDetails | null
  analysis?: ProspectInitialAnalysis | null
}>()

type MenuProfileRow = {
  menuName?: unknown
  menu?: unknown
  profile?: unknown
  yoghurtFit?: unknown
  opportunity?: unknown
}

const menuState = computed(() => String(props.analysis?.menu?.state ?? ''))
const menuUnavailable = computed(() => menuState.value === 'MENU_DATA_NOT_AVAILABLE')
const menuRows = computed(() => {
  const rows = props.analysis?.menu?.menus
  return Array.isArray(rows) ? (rows as MenuProfileRow[]).slice(0, 6) : []
})

function menuRowLabel(row: MenuProfileRow) {
  return String(row.menuName || row.menu || 'Data belum tersedia')
}
</script>

<template>
  <article class="ai-menu-card">
    <div class="ai-menu-head">
      <div>
        <p class="ai-eyebrow"><i class="pi pi-chart-line" /> AI Menu Profiling</p>
        <h2>Menu opportunity preview</h2>
      </div>
      <span class="ai-status-label">{{ analysis?.status === 'PENDING' ? 'Pending' : analysis?.status === 'FAILED' ? 'Unavailable' : 'Saved result' }}</span>
    </div>

    <div v-if="menuUnavailable" class="ai-menu-empty ai-menu-compact"><i class="pi pi-info-circle" /><div><strong>Menu data not available yet.</strong><span>Tagging general Google photos is not enough for AI menu profiling.</span></div></div>
    <div v-else-if="analysis?.status === 'SUCCESS' && analysis.menu" class="ai-menu-empty"><i class="pi pi-check-circle" /><div><strong>Menu profiling tersedia.</strong><span>{{ String(analysis.menu.topOpportunity || analysis.menu.recommendedAction || 'Profil menu tersimpan.') }}</span></div></div>
    <div v-else-if="menuRows.length" class="ai-menu-table" role="table" aria-label="AI menu profiling preview">
      <div class="ai-menu-row ai-menu-header" role="row">
        <span role="columnheader">Menu</span>
        <span role="columnheader">Profile</span>
        <span role="columnheader">Yoghurt Fit</span>
        <span role="columnheader">AI Opportunity</span>
      </div>
      <div v-for="row in menuRows" :key="menuRowLabel(row)" class="ai-menu-row" role="row">
        <strong role="cell">{{ menuRowLabel(row) }}</strong>
        <span role="cell">{{ String(row.profile || 'Data belum tersedia') }}</span>
        <span role="cell">{{ String(row.yoghurtFit || 'UNKNOWN') }}</span>
        <span role="cell">{{ String(row.opportunity || 'Data belum tersedia') }}</span>
      </div>
    </div>

    <div v-else class="ai-menu-empty">
      <i class="pi pi-book" />
      <div>
        <strong>Menu profiling belum dibuat.</strong>
        <span>Menu item data is required before AI menu profiling can run.</span>
      </div>
    </div>
  </article>
</template>

<style scoped>
.ai-menu-card {
  display: grid;
  gap: 0.9rem;
  min-width: 0;
  padding: 1rem;
  border: 1px solid #eadde0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(73, 34, 41, 0.06);
}

.ai-menu-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.ai-menu-head h2 {
  margin: 0.1rem 0 0;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.ai-eyebrow {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: #d62839;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.ai-primary-btn {
  min-height: 36px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  flex-shrink: 0;
  padding: 0 0.75rem;
  border: 0;
  border-radius: 11px;
  background: #f1e8ea;
  color: #9b8b8f;
  font-size: 0.72rem;
  font-weight: 800;
  cursor: not-allowed;
}

.ai-menu-table {
  display: grid;
  min-width: 0;
  overflow: hidden;
  border: 1px solid #eadde0;
  border-radius: 12px;
}

.ai-menu-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 0.85fr 1.1fr;
  min-width: 0;
  border-top: 1px solid #f1e8ea;
}

.ai-menu-row:first-child { border-top: 0; }

.ai-menu-row > * {
  min-width: 0;
  padding: 0.7rem;
  color: var(--text-secondary);
  font-size: 0.76rem;
  line-height: 1.4;
  overflow-wrap: anywhere;
}

.ai-menu-row strong {
  color: var(--text-primary);
}

.ai-menu-header {
  background: #fcf9f9;
}

.ai-menu-header span {
  color: #7a6a6e;
  font-size: 0.66rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.ai-menu-empty {
  display: flex;
  gap: 0.75rem;
  padding: 0.9rem;
  border: 1px dashed #e6dadd;
  border-radius: 12px;
  background: #fcf9f9;
}

.ai-menu-empty > i {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  border-radius: 10px;
  background: #fff0f1;
  color: #e63946;
}

.ai-menu-empty div {
  display: grid;
  gap: 0.2rem;
}

.ai-menu-empty strong {
  color: var(--text-primary);
  font-size: 0.82rem;
}

.ai-menu-empty span {
  color: var(--text-muted);
  font-size: 0.74rem;
  line-height: 1.45;
}

.ai-menu-compact {
  padding: 0.75rem;
}

@media (max-width: 767px) {
  .ai-menu-head {
    flex-direction: column;
  }

  .ai-primary-btn {
    width: 100%;
  }

  .ai-menu-table {
    gap: 0.65rem;
    border: 0;
    overflow: visible;
  }

  .ai-menu-header {
    display: none;
  }

  .ai-menu-row {
    display: grid;
    gap: 0.45rem;
    padding: 0.75rem;
    border: 1px solid #eadde0;
    border-radius: 12px;
    background: #fff;
  }

  .ai-menu-row > * {
    padding: 0;
  }

  .ai-menu-row > *::before {
    display: block;
    margin-bottom: 0.1rem;
    color: #9b8b8f;
    font-size: 0.62rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .ai-menu-row strong::before { content: 'Menu'; }
  .ai-menu-row span:nth-child(2)::before { content: 'Profile'; }
  .ai-menu-row span:nth-child(3)::before { content: 'Yoghurt Fit'; }
  .ai-menu-row span:nth-child(4)::before { content: 'AI Opportunity'; }
}
</style>
