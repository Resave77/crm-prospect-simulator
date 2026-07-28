<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getProspectComments, addProspectComment } from '../api/crm'
import type { ProspectComment } from '../types/crm'
import type { UserRole } from '../types/auth'

const props = defineProps<{
  prospectId: string
  role: UserRole
}>()

const auth = useAuthStore()
const comments = ref<ProspectComment[]>([])
const newComment = ref('')
const loading = ref(true)
const sending = ref(false)
const error = ref('')
const listRef = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (listRef.value) {
      listRef.value.scrollTop = listRef.value.scrollHeight
    }
  })
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHrs = Math.floor(diffMins / 60)
  if (diffHrs < 24) return `${diffHrs}h ago`
  const diffDays = Math.floor(diffHrs / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return d.toLocaleDateString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function isOwnComment(comment: ProspectComment): boolean {
  return comment.userId === auth.user?.id
}

async function load() {
  try {
    comments.value = await getProspectComments(props.prospectId, props.role)
    scrollToBottom()
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Failed to load comments.'
  } finally {
    loading.value = false
  }
}

async function submit() {
  const text = newComment.value.trim()
  if (!text || sending.value) return
  sending.value = true
  error.value = ''
  try {
    const item = await addProspectComment(props.prospectId, text, props.role)
    comments.value.push(item)
    newComment.value = ''
    scrollToBottom()
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Failed to send comment.'
  } finally {
    sending.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submit()
  }
}

watch(() => comments.value.length, () => scrollToBottom())

onMounted(load)
</script>

<template>
  <div class="pc-wrap">
    <div class="pc-header">
      <i class="pi pi-comments" />
      <span>Discussion</span>
      <span v-if="!loading" class="pc-count">{{ comments.length }}</span>
    </div>

    <div v-if="loading" class="pc-loading">
      <i class="pi pi-spin pi-spinner" />
      <span>Loading comments...</span>
    </div>

    <div v-else-if="error && !comments.length" class="pc-error">{{ error }}</div>

    <div v-else ref="listRef" class="pc-list">
      <div v-if="!comments.length" class="pc-empty">
        <i class="pi pi-comment" />
        <span>No comments yet. Start the discussion.</span>
      </div>

      <div
        v-for="c in comments"
        :key="c.id"
        class="pc-msg"
        :class="{ 'pc-msg-own': isOwnComment(c) }"
      >
        <div class="pc-msg-avatar" :class="{ 'pc-msg-avatar-own': isOwnComment(c) }">
          {{ c.userName?.charAt(0)?.toUpperCase() || '?' }}
        </div>
        <div class="pc-msg-body">
          <div class="pc-msg-meta">
            <strong>{{ c.userName }}</strong>
            <span>{{ formatTime(c.createdAt) }}</span>
          </div>
          <p class="pc-msg-text">{{ c.content }}</p>
        </div>
      </div>
    </div>

    <div v-if="error && comments.length" class="pc-inline-error">{{ error }}</div>

    <div class="pc-input-row">
      <textarea
        v-model="newComment"
        placeholder="Write a comment..."
        rows="1"
        :disabled="sending"
        @keydown="onKeydown"
      />
      <button
        class="pc-send-btn"
        :disabled="!newComment.trim() || sending"
        @click="submit"
      >
        <i v-if="sending" class="pi pi-spin pi-spinner" />
        <i v-else class="pi pi-send" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.pc-wrap {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-xl);
  background: var(--surface-card);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.pc-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.85rem 1.15rem;
  border-bottom: 1px solid var(--border-light);
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-muted);
}
.pc-header i { font-size: 0.85rem; color: var(--brand-blue); }
.pc-count {
  min-width: 20px; height: 20px;
  display: inline-grid; place-items: center;
  border-radius: 9999px; background: var(--brand-blue-bg, #eff6ff);
  color: var(--brand-blue); font-size: 0.6rem; font-weight: 700;
  margin-left: auto;
}

.pc-loading {
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  padding: 2rem; color: var(--text-muted); font-size: 0.8rem;
}

.pc-error {
  padding: 0.75rem 1rem; margin: 0.5rem; border-radius: 10px;
  background: #fef2f2; color: #991b1b; font-size: 0.78rem;
}

.pc-inline-error {
  padding: 0.4rem 1rem; background: #fef2f2; color: #991b1b;
  font-size: 0.72rem;
}

.pc-list {
  flex: 1;
  max-height: 360px;
  overflow-y: auto;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  -webkit-overflow-scrolling: touch;
}

.pc-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.4rem;
  padding: 2rem 1rem; color: var(--text-muted); text-align: center;
}
.pc-empty i { font-size: 1.5rem; opacity: 0.35; }
.pc-empty span { font-size: 0.78rem; }

.pc-msg {
  display: flex; gap: 0.6rem; align-items: flex-start;
}
.pc-msg-own { flex-direction: row-reverse; }

.pc-msg-avatar {
  width: 32px; height: 32px; flex-shrink: 0;
  display: grid; place-items: center; border-radius: 50%;
  background: #e2e8f0; color: #475569;
  font-size: 0.72rem; font-weight: 800;
}
.pc-msg-avatar-own {
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff;
}

.pc-msg-body { flex: 1; min-width: 0; }
.pc-msg-meta {
  display: flex; align-items: baseline; gap: 0.4rem;
  margin-bottom: 0.15rem;
}
.pc-msg-meta strong { font-size: 0.75rem; color: var(--text-primary); }
.pc-msg-meta span { font-size: 0.62rem; color: var(--text-muted); }
.pc-msg-own .pc-msg-meta { flex-direction: row-reverse; }

.pc-msg-text {
  margin: 0; padding: 0.55rem 0.75rem;
  background: #f1f5f9; border-radius: 12px 12px 12px 4px;
  font-size: 0.8rem; line-height: 1.5; color: var(--text-secondary);
  word-break: break-word;
}
.pc-msg-own .pc-msg-text {
  background: var(--brand-blue, #2563eb); color: #fff;
  border-radius: 12px 12px 4px 12px;
}

.pc-input-row {
  display: flex; align-items: flex-end; gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  border-top: 1px solid var(--border-light);
  background: var(--surface-card);
}

.pc-input-row textarea {
  flex: 1; resize: none; border: 1px solid var(--border-light);
  border-radius: 12px; padding: 0.55rem 0.75rem;
  font-size: 0.8rem; font-family: inherit; line-height: 1.4;
  color: var(--text-primary); background: #f8fafc;
  outline: none; transition: border-color 0.15s ease, box-shadow 0.15s ease;
  max-height: 100px; min-height: 38px;
}
.pc-input-row textarea:focus {
  border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
  background: #fff;
}
.pc-input-row textarea::placeholder { color: var(--text-muted); }
.pc-input-row textarea:disabled { opacity: 0.6; }

.pc-send-btn {
  width: 38px; height: 38px; flex-shrink: 0;
  display: grid; place-items: center; border-radius: 50%;
  border: 0; background: var(--brand-blue, #2563eb); color: #fff;
  cursor: pointer; font-size: 0.85rem;
  transition: background 0.15s ease, transform 0.15s ease;
}
.pc-send-btn:hover:not(:disabled) { background: #1d4ed8; transform: scale(1.05); }
.pc-send-btn:disabled { opacity: 0.45; cursor: not-allowed; }
</style>
