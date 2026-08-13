<script setup lang="ts">
import type { ProspectInitialAnalysis } from '../../types/crm'

defineProps<{
  prospectName?: string
  analysis?: ProspectInitialAnalysis | null
  loading?: boolean
  error?: string
  context?: 'prospect' | 'customer'
}>()

defineEmits<{ retry: [] }>()
</script>

<template>
  <article class="ai-card ai-summary-card">
    <div class="ai-card-head">
      <div>
        <p class="ai-eyebrow"><i class="pi pi-sparkles" /> AI Summary</p>
        <h2>{{ context === 'customer' ? 'Customer Insight' : 'Prospect insight' }}</h2>
      </div>
      <span v-if="context === 'customer'" class="historical-note">Insight AI tersimpan dari fase prospect</span>
    </div>

    <div v-if="loading" class="ai-state ai-empty">
      <i class="pi pi-spin pi-spinner" />
      <div>
        <strong>{{ context === 'customer' ? 'Memuat insight customer...' : 'Membuat ringkasan prospect...' }}</strong>
        <span>{{ context === 'customer' ? 'Mengambil insight tersimpan dari riwayat prospect.' : 'AI sedang menganalisis data prospect.' }}</span>
      </div>
    </div>
    <div v-else-if="analysis?.summary" class="ai-state ai-empty"><i class="pi pi-check-circle" /><div><strong>{{ String(analysis.summary.summary || 'AI Summary tersedia.') }}</strong><span>Potential: {{ String(analysis.summary.potential || 'UNKNOWN') }}</span></div></div>
    <div v-else-if="error || analysis?.status === 'FAILED'" class="ai-state ai-empty"><i class="pi pi-exclamation-triangle" /><div><strong>AI Summary belum dapat dibuat.</strong><span>{{ error || 'Analisis tetap aman tanpa mengganggu data CRM.' }}</span><button v-if="context !== 'customer'" class="ai-retry-btn" type="button" @click="$emit('retry')">Coba lagi</button></div></div>
    <div v-else class="ai-state ai-empty"><i class="pi pi-file-edit" /><div><strong>AI Summary belum dibuat.</strong><span>{{ context === 'customer' ? 'Insight customer belum tersedia dari riwayat prospect.' : `${prospectName || 'Prospect'} siap untuk diringkas.` }}</span></div></div>
  </article>
</template>

<style scoped>
.ai-card {
  display: grid;
  gap: 0.9rem;
  min-width: 0;
  padding: 1rem;
  border: 1px solid #eadde0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(73, 34, 41, 0.06);
}

.ai-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}
.historical-note{display:block;margin-top:.35rem;color:#8a7b7f;font-size:.68rem;line-height:1.4}

.ai-card-head h2 {
  margin: 0.1rem 0 0;
  color: var(--text-primary);
  font-size: 0.95rem;
  line-height: 1.25;
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

.ai-icon-btn {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  border: 1px solid #eadde0;
  border-radius: 10px;
  background: #fcf9f9;
  color: #9b8b8f;
}

.ai-state {
  display: flex;
  gap: 0.75rem;
  min-width: 0;
  padding: 0.85rem;
  border: 1px dashed #e6dadd;
  border-radius: 12px;
  background: #fcf9f9;
}

.ai-state > i {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  border-radius: 10px;
  background: #fff0f1;
  color: #e63946;
}

.ai-state div {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
}

.ai-state strong {
  color: var(--text-primary);
  font-size: 0.82rem;
}

.ai-state span {
  color: var(--text-muted);
  font-size: 0.74rem;
  line-height: 1.45;
}

.ai-primary-btn {
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 0;
  border-radius: 12px;
  background: #f1e8ea;
  color: #9b8b8f;
  font-size: 0.75rem;
  font-weight: 800;
  cursor: not-allowed;
}
.ai-retry-btn { width: fit-content; margin-top: 0.35rem; padding: 0.4rem 0.65rem; border: 1px solid #e63946; border-radius: 9px; background: #fff; color: #d62839; font-size: 0.72rem; font-weight: 800; cursor: pointer; }
</style>
