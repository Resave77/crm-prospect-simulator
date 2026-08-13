<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { askProspectAI, getProspectAIChatHistory } from '../../api/crm'
import { useAuthStore } from '../../stores/auth'
import type { ProspectAIChatHistory, ProspectChatResponse } from '../../types/crm'

const props = defineProps<{ prospectId: string; expanded?: boolean }>()
const emit = defineEmits<{ expand: [] }>()
const auth = useAuthStore()
const message = ref('')
const loading = ref(false)
const error = ref('')
const answer = ref<ProspectChatResponse | null>(null)
const history = ref<ProspectAIChatHistory[]>([])
const historyLoading = ref(false)
const historyError = ref('')
const threadRef = ref<HTMLElement | null>(null)
const visibleHistory = computed(() => props.expanded ? history.value : history.value.slice(-1))
const latestHistory = computed(() => history.value.at(-1) ?? null)
const chips = [{ label: 'Ringkas prospek', skill: 'PROSPECT_ANALYSIS' }, { label: 'Cari peluang yoghurt', skill: 'MENU_OPPORTUNITY' }, { label: 'Buat pitch visit', skill: 'SALES_PITCH' }]
const selectedSkill = ref('AUTO')
const messageLength = computed(() => message.value.trim().length)
const hasProspectId = computed(() => props.prospectId.trim().length > 0)
const permissionResult = computed(() => auth.hasPermission('use_prospect_ai_chat'))
const canSubmit = computed(() => messageLength.value > 0 && hasProspectId.value && permissionResult.value && !loading.value)
const roleLabels: Record<string, string> = { SUPER_ADMIN: 'Super Admin', ADMINISTRATOR: 'Administrator', SALES_MANAGER: 'Sales Manager', SALES_EXECUTIVE: 'Sales' }

function authorLabel(item: ProspectAIChatHistory) {
  const role = item.authorRole ? roleLabels[item.authorRole] || item.authorRole : ''
  return [item.authorName, role].filter(Boolean).join(' · ')
}

function timestampLabel(value: string) {
  return new Intl.DateTimeFormat('id-ID', { hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

function scrollThreadToLatest() {
  nextTick(() => {
    const scroll = () => { if (threadRef.value) threadRef.value.scrollTop = threadRef.value.scrollHeight }
    scroll()
    requestAnimationFrame(scroll)
  })
}

function useChip(chip: { label: string; skill: string }) {
  message.value = chip.label
  selectedSkill.value = chip.skill
}
function handleCardClick() {
  if (!props.expanded && window.matchMedia('(max-width: 1199px)').matches) emit('expand')
}

async function submit() {
  if (!canSubmit.value) return
  loading.value = true; error.value = ''
  const question = message.value.trim()
  try {
    const result = await askProspectAI(props.prospectId, question, selectedSkill.value)
    const createdAt = new Date().toISOString()
    answer.value = result
    history.value.push({ ...result, id: `session-${Date.now()}`, message: question, userId: auth.user?.id, authorName: auth.user?.fullName, authorRole: auth.user?.role, createdAt })
    message.value = ''
    scrollThreadToLatest()
  }
  catch (caught: any) { error.value = caught?.response?.data?.error?.message || 'AI sedang tidak tersedia.' }
  finally { loading.value = false }
}
async function loadHistory() {
  if (!hasProspectId.value) return
  historyLoading.value = true
  historyError.value = ''
  try {
    history.value = await getProspectAIChatHistory(props.prospectId)
    answer.value = history.value.at(-1) || null
    scrollThreadToLatest()
  } catch {
    history.value = []
    answer.value = null
    historyError.value = 'Riwayat percakapan tidak dapat dimuat.'
  } finally {
    historyLoading.value = false
  }
}
onMounted(loadHistory)
watch(() => props.expanded, (value) => { if (value) scrollThreadToLatest() })
watch(() => history.value.length, () => { if (props.expanded) scrollThreadToLatest() })
</script>

<template>
  <article :class="['tanya-card', { 'tanya-card-expanded': expanded }]" @click="handleCardClick">
    <div class="tanya-head">
      <div class="tanya-title-group">
        <p class="ai-eyebrow"><i class="pi pi-comments" /> Tanya AI</p>
        <span>{{ loading ? 'Thinking...' : 'Sales copilot' }}</span>
      </div>
      <slot name="header-action" />
    </div>
    <p v-if="!expanded" class="tanya-mobile-hint">Ketuk untuk lihat percakapan</p>

    <div class="tanya-messages">
      <div v-if="loading" class="tanya-empty"><i class="pi pi-spin pi-spinner" /><strong>AI sedang menganalisis...</strong><span>Tunggu sebentar.</span></div>
      <div v-else-if="error" class="tanya-empty"><i class="pi pi-exclamation-triangle" /><strong>AI tidak tersedia.</strong><span>{{ error }}</span></div>
      <div v-else-if="historyLoading" class="tanya-empty"><i class="pi pi-spin pi-spinner" /><strong>Memuat percakapan...</strong></div>
      <div v-else-if="historyError" class="tanya-empty tanya-history-error"><i class="pi pi-exclamation-triangle" /><strong>Riwayat percakapan tidak dapat dimuat.</strong><button type="button" @click.stop="loadHistory">Coba lagi</button></div>
      <div v-else-if="expanded && history.length" ref="threadRef" class="tanya-history"><div v-for="item in visibleHistory" :key="item.id" class="tanya-turn"><div class="tanya-user-meta"><span v-if="authorLabel(item)">{{ authorLabel(item) }}</span><time :datetime="item.createdAt">{{ timestampLabel(item.createdAt) }}</time></div><div class="tanya-user-bubble">{{ item.message }}</div><div class="tanya-ai-bubble"><span class="tanya-ai-avatar"><i class="pi pi-sparkles" /></span><div class="tanya-answer"><strong>{{ item.answer }}</strong><span v-if="item.insight"><b>Insight</b>{{ item.insight }}</span><span v-if="item.why"><b>Why</b>{{ item.why }}</span><span v-if="item.recommendedAction"><b>Next step</b>{{ item.recommendedAction }}</span></div></div></div></div>
      <div v-else-if="latestHistory" class="tanya-preview">
        <div class="tanya-last-request"><span>Permintaan terakhir</span><p>{{ latestHistory.message }}</p></div>
        <div class="tanya-preview-heading"><span class="tanya-ai-avatar"><i class="pi pi-sparkles" /></span><strong>Respons AI terbaru</strong></div>
        <p class="tanya-preview-answer">{{ latestHistory.answer }}</p>
        <div class="tanya-preview-meta"><span v-if="authorLabel(latestHistory)">{{ authorLabel(latestHistory) }}</span><time :datetime="latestHistory.createdAt">{{ timestampLabel(latestHistory.createdAt) }}</time></div>
      </div>
      <div v-else class="tanya-empty">
        <i class="pi pi-sparkles" />
        <strong>Tanya AI siap digunakan.</strong>
        <span>Tanyakan hal spesifik tentang prospek ini.</span>
      </div>
    </div>

    <div class="tanya-chips" aria-label="AI prompt suggestions" @click.stop>
      <button v-for="chip in chips" :key="chip.label" type="button" @click="useChip(chip)" :disabled="loading">{{ chip.label }}</button>
    </div>

    <form class="tanya-input-row" @click.stop @submit.prevent="submit">
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
.tanya-card-expanded {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: #eef1f6;
  box-shadow: none;
}
.tanya-card-expanded .tanya-head {
  flex-shrink: 0;
  padding: .8rem clamp(1rem, 3vw, 1.5rem);
  background: #fff;
  border-bottom: 1px solid #e7ebf1;
  min-height: 60px;
}
.tanya-card-expanded .ai-eyebrow { font-size: .8rem; letter-spacing: -.01em; }
.tanya-card-expanded .tanya-head span {
  background: #fff0f1;
  border: 1px solid #ffd9dc;
  color: #b42332;
  font-size: .62rem;
  font-weight: 800;
}
.tanya-card-expanded .tanya-messages {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  overflow: hidden;
  border: 0;
  border-radius: 0;
  padding: clamp(1rem, 2.5vw, 1.75rem);
  background:
    radial-gradient(circle at 1px 1px, rgba(100, 116, 139, .055) 1px, transparent 0),
    linear-gradient(180deg, #f1f4f9 0%, #eaeef5 100%);
  background-size: 22px 22px, 100% 100%;
}
.tanya-card-expanded .tanya-history {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  height: auto;
  padding-right: .25rem;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, .5) transparent;
}
.tanya-card-expanded .tanya-history::-webkit-scrollbar { width: 8px; }
.tanya-card-expanded .tanya-history::-webkit-scrollbar-track { background: transparent; }
.tanya-card-expanded .tanya-history::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(148, 163, 184, .45);
  border: 2px solid transparent;
  background-clip: padding-box;
}
.tanya-card-expanded .tanya-turn {
  width: 100%;
  margin: 0;
}
.tanya-card-expanded .tanya-empty { margin: 0; padding-top: 1.5rem; min-height: 0; }
.tanya-card-expanded .tanya-answer { width: 100%; }
.tanya-card-expanded .tanya-user-bubble { max-width: 70%; }
.tanya-card-expanded .tanya-ai-bubble { max-width: 80%; }
.tanya-card-expanded .tanya-chips {
  flex-shrink: 0;
  padding: .7rem clamp(1rem, 3vw, 1.5rem);
  border-top: 1px solid #e7ebf1;
  background: #fff;
  gap: .5rem;
}
.tanya-card-expanded .tanya-chips button {
  min-height: 34px;
  border-color: #f3cdd2;
  background: #fff;
  color: #b42332;
  font-weight: 800;
}
.tanya-card-expanded .tanya-chips button:hover:not(:disabled) { background: #fff0f1; border-color: #e63946; }
.tanya-card-expanded .tanya-chips button:disabled { opacity: .5; cursor: not-allowed; }
.tanya-card-expanded .tanya-input-row {
  flex-shrink: 0;
  padding: .8rem clamp(1rem, 3vw, 1.5rem) calc(.9rem + env(safe-area-inset-bottom, 0px));
  border-top: 1px solid #e7ebf1;
  background: #fff;
  gap: .6rem;
}
.tanya-card-expanded .tanya-input-row textarea {
  min-height: 46px;
  max-height: 130px;
  resize: none;
  padding: .6rem 1rem;
  border-color: #e5e9f0;
  border-radius: 18px;
  background: #f4f6fa;
  font-size: .84rem;
  transition: border-color .15s ease, box-shadow .15s ease, background .15s ease;
}
.tanya-card-expanded .tanya-input-row textarea:focus {
  border-color: #f4b3ba;
  background: #fff;
  box-shadow: 0 0 0 4px rgba(230, 57, 70, .08);
}
.tanya-card-expanded .tanya-input-row button {
  min-height: 44px;
  padding: 0 1.15rem;
  border-radius: 999px;
  background: linear-gradient(145deg, #ef4e5d, #e63946 60%, #d62839);
  color: #fff;
  font-size: .76rem;
  font-weight: 800;
  box-shadow: 0 8px 18px -6px rgba(214, 40, 57, .45);
  transition: transform .15s ease, box-shadow .15s ease;
}
.tanya-card-expanded .tanya-input-row button:not(:disabled):hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 12px 22px -6px rgba(214, 40, 57, .5);
}
.tanya-card-expanded .tanya-input-row button:disabled { opacity: .45; cursor: not-allowed; box-shadow: none; }
.tanya-history { display: flex; flex-direction: column; gap: .9rem; min-width: 0; overflow-y: auto; scrollbar-width: thin; }
.tanya-turn { display: flex; flex-direction: column; gap: .5rem; min-width: 0; }
.tanya-user-meta { align-self: flex-end; display: flex; gap: .35rem; flex-wrap: wrap; justify-content: flex-end; color: #64748b; font-size: .68rem; line-height: 1.3; }
.tanya-user-meta span { font-weight: 700; color: #475569; }
.tanya-user-bubble { align-self: flex-end; max-width: 74%; padding: .65rem .8rem; border-radius: 16px 4px 16px 16px; background: linear-gradient(145deg, #ef4e5d, #d62839); color: #fff; font-size: .78rem; line-height: 1.45; box-shadow: 0 5px 14px rgba(214,40,57,.14); }
.tanya-ai-bubble { display: flex; align-items: flex-start; gap: .5rem; align-self: flex-start; max-width: 82%; min-width: 0; }
.tanya-ai-avatar { width: 28px; height: 28px; flex: 0 0 28px; display: grid; place-items: center; border-radius: 9px; background: #fff0f1; color: #e63946; }
.tanya-ai-bubble .tanya-answer { display: grid; gap: .35rem; padding: .7rem .8rem; border: 1px solid #e5e7eb; border-radius: 4px 16px 16px 16px; background: #fff; color: #334155; box-shadow: 0 3px 10px rgba(15,23,42,.04); line-height: 1.45; }
.tanya-ai-bubble .tanya-answer span { display: grid; gap: .1rem; color: #64748b; font-size: .72rem; }
.tanya-ai-bubble .tanya-answer b { color: #b42332; font-size: .65rem; text-transform: uppercase; letter-spacing: .04em; }
.tanya-card-expanded .tanya-history { padding-right: .25rem; }
.tanya-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}
.tanya-title-group { min-width: 0; display: flex; align-items: center; flex-wrap: wrap; gap: .45rem; }
.tanya-head .ai-eyebrow { min-width: 0; }
.tanya-head :deep(.expand-control) { flex: 0 0 auto; }

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

.tanya-mobile-hint {
  display: none;
  margin: -0.45rem 0 0;
  color: #a18f94;
  font-size: 0.68rem;
  line-height: 1.35;
  text-align: right;
}

.tanya-messages {
  min-height: 112px;
  display: grid;
  place-items: center;
  padding: 0.85rem;
  border: 1px dashed #e6dadd;
  border-radius: 12px;
  background: linear-gradient(180deg, #fff 0, #fcf9f9 100%);
}
.tanya-card:not(.tanya-card-expanded) .tanya-messages { min-height: 0; display: block; padding: 0; border: 0; background: transparent; }
.tanya-preview { width: 100%; min-width: 0; display: grid; gap: .65rem; padding: .85rem; box-sizing: border-box; border-radius: 12px; background: #fcf9f9; }
.tanya-last-request { min-width: 0; display: grid; gap: .18rem; padding-bottom: .6rem; border-bottom: 1px solid #eee4e6; }
.tanya-last-request span { color: #8d7d81; font-size: .7rem; font-weight: 700; }
.tanya-last-request p { margin: 0; overflow-wrap: anywhere; color: #5f4f54; font-size: .78rem; line-height: 1.45; }
.tanya-preview-heading { min-width: 0; display: flex; align-items: center; gap: .5rem; color: #b42332; font-size: .76rem; }
.tanya-preview-heading .tanya-ai-avatar { width: 26px; height: 26px; flex-basis: 26px; }
.tanya-preview-answer { margin: 0; display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 5; overflow: hidden; overflow-wrap: anywhere; white-space: normal; color: #273142; font-size: .875rem; line-height: 1.55; }
.tanya-preview-meta { min-width: 0; display: flex; flex-wrap: wrap; gap: .25rem .4rem; color: #817177; font-size: .7rem; line-height: 1.35; }
.tanya-preview-meta span { font-weight: 700; }
@media (max-width: 767px) {
  .tanya-card:not(.tanya-card-expanded) { cursor: pointer; transition: border-color .18s ease, box-shadow .18s ease; }
  .tanya-card:not(.tanya-card-expanded):active { border-color: #f1a3aa; box-shadow: 0 0 0 3px rgba(230,57,70,.08); }
  .tanya-card:not(.tanya-card-expanded) .tanya-mobile-hint { display: block; }
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
.tanya-history-error button {
  min-height: 32px;
  padding: 0 .75rem;
  border: 1px solid #f1b1b8;
  border-radius: 999px;
  background: #fff;
  color: #d62839;
  font-size: .72rem;
  font-weight: 800;
  cursor: pointer;
}

.tanya-chips {
  display: flex;
  gap: 0.45rem;
  min-width: 0;
  overflow-x: auto;
  padding-bottom: .2rem;
  -webkit-overflow-scrolling: touch;
}

.tanya-card:not(.tanya-card-expanded) .tanya-chips { flex-wrap: wrap; overflow-x: visible; padding-bottom: 0; }

.tanya-chips button {
  min-height: 34px;
  flex: 0 0 auto;
  max-width: 100%;
  padding: 0 0.65rem;
  border: 1px solid #eadde0;
  border-radius: 999px;
  background: #fff;
  color: #6b5d61;
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
  white-space: normal;
  overflow-wrap: anywhere;
  line-height: 1.25;
}
.tanya-card:not(.tanya-card-expanded) .tanya-chips button { flex: 1 1 120px; }

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
  flex: 1 1 180px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
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
  .tanya-user-bubble { max-width: 88%; }
  .tanya-ai-bubble { max-width: 92%; }

  .tanya-card-expanded .tanya-input-row {
    display: flex;
    align-items: flex-end;
    gap: .5rem;
    padding: .65rem .7rem calc(.75rem + env(safe-area-inset-bottom, 0px));
  }
  .tanya-card-expanded .tanya-input-row button {
    width: auto;
    min-height: 44px;
    padding: 0 1rem;
  }
  .tanya-card-expanded .tanya-turn { width: 100%; }
  .tanya-card-expanded .tanya-user-bubble { max-width: 92%; }
  .tanya-card-expanded .tanya-ai-bubble { max-width: 92%; }
  .tanya-card-expanded .tanya-messages { padding: 1rem .75rem; }
  .tanya-card-expanded .tanya-head { padding: .65rem .85rem; min-height: 54px; }
  .tanya-card-expanded .tanya-chips { padding: .6rem .7rem; }
}

@media (max-width: 1100px) {
  .tanya-card:not(.tanya-card-expanded) .tanya-input-row { flex-wrap: wrap; }
  .tanya-card:not(.tanya-card-expanded) .tanya-input-row button { margin-left: auto; }
}
</style>
