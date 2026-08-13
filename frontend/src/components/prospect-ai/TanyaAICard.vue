<script setup lang="ts">
import { computed, ref } from 'vue'
import { askProspectAI } from '../../api/crm'
import { useAuthStore } from '../../stores/auth'
import type { ProspectChatResponse } from '../../types/crm'

const props = defineProps<{ prospectId: string }>()
const auth = useAuthStore()
const message = ref('')
const loading = ref(false)
const error = ref('')
const answer = ref<ProspectChatResponse | null>(null)
const chips = [{ label: 'Ringkas prospek', skill: 'PROSPECT_ANALYSIS' }, { label: 'Cari peluang yoghurt', skill: 'MENU_OPPORTUNITY' }, { label: 'Buat pitch visit', skill: 'SALES_PITCH' }]
const selectedSkill = ref('AUTO')
const messageLength = computed(() => message.value.trim().length)
const hasProspectId = computed(() => props.prospectId.trim().length > 0)
const permissionResult = computed(() => auth.hasPermission('use_prospect_ai_chat'))
const canSubmit = computed(() => messageLength.value > 0 && hasProspectId.value && permissionResult.value && !loading.value)

function useChip(chip: { label: string; skill: string }) {
  message.value = chip.label
  selectedSkill.value = chip.skill
}

async function submit() {
  if (!canSubmit.value) return
  loading.value = true; error.value = ''
  try { answer.value = await askProspectAI(props.prospectId, message.value.trim(), selectedSkill.value); message.value = '' }
  catch (caught: any) { error.value = caught?.response?.data?.error?.message || 'AI sedang tidak tersedia.' }
  finally { loading.value = false }
}
</script>

<template>
  <article class="tanya-card">
    <div class="tanya-head">
      <p class="ai-eyebrow"><i class="pi pi-comments" /> Tanya AI</p>
      <span>{{ loading ? 'Thinking...' : 'Sales copilot' }}</span>
    </div>

    <div class="tanya-messages">
      <div v-if="loading" class="tanya-empty"><i class="pi pi-spin pi-spinner" /><strong>AI sedang menganalisis...</strong><span>Tunggu sebentar.</span></div>
      <div v-else-if="error" class="tanya-empty"><i class="pi pi-exclamation-triangle" /><strong>AI tidak tersedia.</strong><span>{{ error }}</span></div>
      <div v-else-if="answer" class="tanya-answer"><strong>{{ answer.answer }}</strong><span v-if="answer.insight">Insight: {{ answer.insight }}</span><span v-if="answer.recommendedAction">Langkah berikutnya: {{ answer.recommendedAction }}</span></div>
      <div v-else class="tanya-empty">
        <i class="pi pi-sparkles" />
        <strong>Tanya AI siap digunakan.</strong>
        <span>Tanyakan hal spesifik tentang prospek ini.</span>
      </div>
    </div>

    <div class="tanya-chips" aria-label="AI prompt suggestions">
      <button v-for="chip in chips" :key="chip.label" type="button" @click="useChip(chip)" :disabled="loading">{{ chip.label }}</button>
    </div>

    <form class="tanya-input-row" @submit.prevent="submit">
      <textarea
        v-model="message"
        rows="2"
        placeholder="Tulis pertanyaan untuk AI..."
        @keydown.enter.exact.prevent="submit"
      />
      <button class="tanya-submit" type="submit" :disabled="!canSubmit"><i class="pi pi-send" /> {{ loading ? '...' : 'Ask AI' }}</button>
    </form>
  </article>
</template>

<style scoped>
.tanya-card {
  display: grid;
  gap: 0.85rem;
  min-width: 0;
  padding: 1rem;
  border: 1px solid #eadde0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(73, 34, 41, 0.06);
}

.tanya-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
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

.tanya-head span {
  flex-shrink: 0;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: #f1e8ea;
  color: #8d7d81;
  font-size: 0.62rem;
  font-weight: 800;
}

.tanya-messages {
  min-height: 150px;
  display: grid;
  place-items: center;
  padding: 0.85rem;
  border: 1px dashed #e6dadd;
  border-radius: 12px;
  background: linear-gradient(180deg, #fff 0, #fcf9f9 100%);
}

.tanya-empty {
  display: grid;
  justify-items: center;
  gap: 0.35rem;
  text-align: center;
}

.tanya-empty i {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 13px;
  background: #fff0f1;
  color: #e63946;
}

.tanya-empty strong {
  color: var(--text-primary);
  font-size: 0.82rem;
}

.tanya-empty span {
  max-width: 260px;
  color: var(--text-muted);
  font-size: 0.74rem;
  line-height: 1.45;
}

.tanya-chips {
  display: flex;
  gap: 0.45rem;
  overflow-x: auto;
  padding-bottom: 0.2rem;
  -webkit-overflow-scrolling: touch;
}

.tanya-chips button {
  min-height: 34px;
  flex: 0 0 auto;
  padding: 0 0.65rem;
  border: 1px solid #eadde0;
  border-radius: 999px;
  background: #fff;
  color: #6b5d61;
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
}

.tanya-chips button:hover {
  border-color: #f4b3ba;
  background: #fff0f1;
  color: #d62839;
}

.tanya-input-row {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  min-width: 0;
}

.tanya-input-row textarea {
  flex: 1;
  min-width: 0;
  min-height: 42px;
  max-height: 110px;
  resize: vertical;
  padding: 0.6rem 0.7rem;
  border: 1px solid #e6dadd;
  border-radius: 12px;
  background: #fcf9f9;
  color: var(--text-primary);
  font: inherit;
  font-size: 0.78rem;
  line-height: 1.4;
}

.tanya-input-row button {
  min-height: 42px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  flex-shrink: 0;
  padding: 0 0.75rem;
  border: 0;
  border-radius: 12px;
  background: #f1e8ea;
  color: #9b8b8f;
  font-size: 0.72rem;
  font-weight: 800;
  cursor: pointer;
}

.tanya-input-row button:disabled {
  cursor: not-allowed;
}

.tanya-input-row .tanya-submit:not(:disabled) {
  background: #e63946;
  color: #fff;
}

@media (max-width: 480px) {
  .tanya-input-row {
    display: grid;
  }

  .tanya-input-row button {
    width: 100%;
  }
}
</style>
