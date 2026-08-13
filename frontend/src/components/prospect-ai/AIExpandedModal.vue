<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  subtitle: string
  badge?: string
  variant?: 'summary' | 'chat' | 'discussion'
}>()

const emit = defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="ai-modal-overlay" @click.self="emit('close')">
      <section
        :class="['ai-modal-panel', `ai-modal-panel--${variant ?? 'summary'}`]"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ai-expanded-modal-title"
      >
        <header class="ai-modal-header">
          <div class="ai-modal-heading">
            <span id="ai-expanded-modal-title">{{ title }}</span>
            <strong>{{ subtitle }}</strong>
          </div>
          <div class="ai-modal-header-actions">
            <span v-if="badge" class="ai-modal-badge">{{ badge }}</span>
            <button type="button" aria-label="Tutup" @click="emit('close')"><i class="pi pi-times" /></button>
          </div>
        </header>
        <div class="ai-modal-content"><slot /></div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.ai-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 5000;
  display: grid;
  place-items: center;
  box-sizing: border-box;
  padding: clamp(12px, 3vw, 32px);
  background: rgba(24, 18, 20, .48);
  backdrop-filter: blur(4px);
  animation: ai-overlay-in .18s ease-out;
}

.ai-modal-panel {
  width: clamp(380px, 32vw, 460px);
  height: min(82dvh, 780px);
  max-width: 100%;
  max-height: calc(100dvh - 24px);
  min-width: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  border: 1px solid #eadfe1;
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 28px 80px rgba(15, 23, 42, .25);
  animation: ai-panel-in .18s ease-out;
}

.ai-modal-header {
  flex: 0 0 auto;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  padding: .8rem 1rem;
  border-bottom: 1px solid #e7ebf1;
  background: #fffafa;
}

.ai-modal-heading { min-width: 0; display: grid; gap: .15rem; }
.ai-modal-heading span { color: #b42332; font-size: .72rem; font-weight: 800; text-transform: uppercase; letter-spacing: .05em; }
.ai-modal-heading strong { overflow-wrap: anywhere; color: #1f2937; font-size: .86rem; line-height: 1.35; }
.ai-modal-header-actions { flex: 0 0 auto; display: flex; align-items: center; gap: .45rem; }
.ai-modal-badge { padding: .22rem .52rem; border: 1px solid #ffd9dc; border-radius: 999px; background: #fff0f1; color: #b42332; font-size: .62rem; font-weight: 800; }
.ai-modal-header button { width: 36px; height: 36px; display: grid; place-items: center; border: 1px solid #eadde0; border-radius: 10px; background: #fff; color: #b42332; cursor: pointer; }
.ai-modal-header button:hover { border-color: #f1a3aa; background: #fff0f1; }

.ai-modal-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  max-width: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 1rem;
}
.ai-modal-content > :deep(*) { width: 100%; min-width: 0; max-width: 100%; box-sizing: border-box; }
.ai-modal-content :deep(p), .ai-modal-content :deep(strong), .ai-modal-content :deep(span) { overflow-wrap: anywhere; white-space: normal; }
.ai-modal-panel--summary .ai-modal-content :deep(.ai-card) { box-shadow: none; }

.ai-modal-panel--chat .ai-modal-content,
.ai-modal-panel--discussion .ai-modal-content { display: flex; overflow: hidden; padding: 0; }
.ai-modal-panel--chat .ai-modal-content :deep(.tanya-card),
.ai-modal-panel--discussion .ai-modal-content :deep(.pc-floating) { flex: 1 1 auto; min-width: 0; min-height: 0; }
.ai-modal-panel--chat .ai-modal-content :deep(.tanya-card-expanded .tanya-head) { display: none; }
.ai-modal-panel--chat .ai-modal-content :deep(.tanya-card-expanded .tanya-user-bubble),
.ai-modal-panel--chat .ai-modal-content :deep(.tanya-card-expanded .tanya-ai-bubble) { max-width: 100%; overflow-wrap: anywhere; }
.ai-modal-panel--chat .ai-modal-content :deep(.tanya-chips) { max-width: 100%; overflow-x: auto; }
.ai-modal-panel--discussion .ai-modal-content :deep(.pc-wrap) { position: static; width: 100%; height: 100%; min-height: 0; border: 0; box-shadow: none; }
.ai-modal-panel--discussion .ai-modal-content :deep(.pc-list) { max-height: none !important; overflow-y: auto !important; justify-content: flex-start; }

/* Discussion keeps its established independent landscape dialog. */
.ai-modal-panel--discussion { width: min(1080px, calc(100vw - 80px)); height: min(760px, 78dvh); max-width: 100%; border-radius: 18px; }
.ai-modal-panel--discussion .ai-modal-header { background: #fff; }
.ai-modal-panel--discussion .ai-modal-header button { background: #fff0f1; border-color: #ffd9dc; }

@media (max-width: 767px) {
  .ai-modal-overlay { padding: 12px; }
  .ai-modal-panel { width: calc(100vw - 24px); max-width: none; height: min(90dvh, 780px); max-height: calc(100dvh - 24px); border-radius: 16px; }
  .ai-modal-header { padding: .68rem .75rem; }
  .ai-modal-heading strong { font-size: .8rem; }
  .ai-modal-badge { display: none; }
  .ai-modal-content { padding: .75rem; }
  .ai-modal-panel--chat .ai-modal-content,
  .ai-modal-panel--discussion .ai-modal-content { padding: 0; }
  .ai-modal-panel--discussion { width: 100%; height: min(760px, 92dvh); }
}

@media (max-width: 390px) {
  .ai-modal-overlay { padding: 8px; }
  .ai-modal-panel { width: calc(100vw - 16px); height: calc(100dvh - 16px); max-height: none; border-radius: 14px; }
}

@media (prefers-reduced-motion: reduce) {
  .ai-modal-overlay, .ai-modal-panel { animation: none; }
}

@keyframes ai-overlay-in { from { opacity: 0; } to { opacity: 1; } }
@keyframes ai-panel-in { from { opacity: 0; transform: scale(.98) translateY(6px); } to { opacity: 1; transform: none; } }
</style>
