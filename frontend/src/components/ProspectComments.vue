<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getProspectComments, addProspectComment, downloadCommentAttachment, getMentionUsers } from '../api/crm'
import type { ProspectComment, SalesExecutiveOption } from '../types/crm'
import type { UserRole } from '../types/auth'

const props = defineProps<{
  prospectId: string
  role: UserRole
}>()

const auth = useAuthStore()
const comments = ref<ProspectComment[]>([])
const newComment = ref('')
const files = ref<File[]>([])
const mentionUsers = ref<SalesExecutiveOption[]>([])
const fileInput = ref<HTMLInputElement | null>(null)
const loading = ref(true)
const sending = ref(false)
const error = ref('')
const listRef = ref<HTMLElement | null>(null)
const open = ref(false)
const imagePreviews = ref<Record<string, string>>({})
let pollId: number | undefined

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

async function ensureImagePreviews(items: ProspectComment[]) {
  const images = items.flatMap((comment) => comment.attachments ?? []).filter((file) => file.contentType.startsWith('image/'))
  await Promise.all(images.map(async (file) => {
    if (imagePreviews.value[file.id]) return
    try {
      const blob = await downloadCommentAttachment(props.prospectId, file.id, props.role)
      imagePreviews.value[file.id] = URL.createObjectURL(blob)
    } catch { /* the filename remains available when a preview cannot be loaded */ }
  }))
}

async function load() {
  try {
    const [items, users] = await Promise.all([getProspectComments(props.prospectId, props.role), getMentionUsers(props.role)])
    comments.value = items
    await ensureImagePreviews(items)
    mentionUsers.value = users.filter((u) => u.id !== auth.user?.id)
    scrollToBottom()
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Failed to load comments.'
  } finally {
    loading.value = false
  }
}

async function submit() {
  const text = newComment.value.trim()
  if ((!text && !files.value.length) || sending.value) return
  sending.value = true
  error.value = ''
  try {
    const item = await addProspectComment(props.prospectId, text, props.role, files.value)
    comments.value.push(item)
    await ensureImagePreviews([item])
    newComment.value = ''
    files.value = []
    if (fileInput.value) fileInput.value.value = ''
    scrollToBottom()
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Failed to send comment.'
  } finally {
    sending.value = false
  }
}

function mention(user: SalesExecutiveOption) {
  const before = newComment.value.replace(/@[^@\s]*$/, '')
  newComment.value = `${before}@${user.fullName} `
}

function selectFiles(event: Event) {
  const selected = Array.from((event.target as HTMLInputElement).files ?? [])
  error.value = ''
  if (selected.length > 5 || selected.some((file) => file.size > 5 * 1024 * 1024)) {
    error.value = 'Maximum 5 files, with a maximum size of 5 MB each.'
    return
  }
  files.value = selected
}

async function openAttachment(id: string, name: string) {
  try {
    const blob = await downloadCommentAttachment(props.prospectId, id, props.role)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a'); anchor.href = url; anchor.download = name; anchor.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  } catch { error.value = 'Failed to download attachment.' }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submit()
  }
}

watch(() => comments.value.length, () => scrollToBottom())
watch(open, (value) => { if (value) scrollToBottom() })
watch(() => props.prospectId, () => { loading.value = true; comments.value = []; load() })

onMounted(() => {
  load()
  pollId = window.setInterval(async () => {
    if (open.value || sending.value) return
    try {
      comments.value = await getProspectComments(props.prospectId, props.role)
      await ensureImagePreviews(comments.value)
    } catch { /* keep the last known badge count */ }
  }, 30000)
})
onBeforeUnmount(() => {
  if (pollId) window.clearInterval(pollId)
  Object.values(imagePreviews.value).forEach((url) => URL.revokeObjectURL(url))
})
</script>

<template>
  <div class="pc-floating">
    <button class="pc-launcher" aria-label="Open prospect discussion" @click="open = !open">
      <i :class="open ? 'pi pi-times' : 'pi pi-comments'" />
      <span v-if="comments.length && !open" class="pc-launcher-badge">{{ comments.length > 99 ? '99+' : comments.length }}</span>
    </button>
  <div v-if="open" class="pc-wrap">
    <div class="pc-header">
      <i class="pi pi-comments" />
      <span>Discussion</span>
      <span v-if="!loading" class="pc-count">{{ comments.length }}</span>
      <button class="pc-close" aria-label="Close discussion" @click="open = false"><i class="pi pi-times" /></button>
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
          <p v-if="c.content" class="pc-msg-text">{{ c.content }}</p>
          <div v-if="c.attachments?.length" class="pc-attachments">
            <button v-for="file in c.attachments" :key="file.id" @click="openAttachment(file.id, file.name)">
              <img v-if="file.contentType.startsWith('image/') && imagePreviews[file.id]" :src="imagePreviews[file.id]" :alt="file.name" />
              <i v-else :class="file.contentType.startsWith('image/') ? 'pi pi-image' : 'pi pi-file'" />
              <span>{{ file.name }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="error && comments.length" class="pc-inline-error">{{ error }}</div>

    <div v-if="newComment.match(/@[^@\s]*$/)" class="pc-mentions">
      <button v-for="user in mentionUsers.filter(u => u.fullName.toLowerCase().includes((newComment.match(/@([^@\s]*)$/)?.[1] || '').toLowerCase())).slice(0, 6)" :key="user.id" @click="mention(user)">@{{ user.fullName }}</button>
    </div>
    <div v-if="files.length" class="pc-selected-files"><span v-for="file in files" :key="file.name"><i class="pi pi-paperclip" />{{ file.name }}</span></div>
    <div class="pc-input-row">
      <input ref="fileInput" class="pc-file-input" type="file" multiple accept="image/jpeg,image/png,.pdf,.docx,.xlsx" @change="selectFiles" />
      <button class="pc-attach-btn" title="Attach photo or document" @click="fileInput?.click()"><i class="pi pi-paperclip" /></button>
      <textarea
        v-model="newComment"
        placeholder="Write a comment..."
        rows="1"
        :disabled="sending"
        @keydown="onKeydown"
      />
      <button
        class="pc-send-btn"
        :disabled="(!newComment.trim() && !files.length) || sending"
        @click="submit"
      >
        <i v-if="sending" class="pi pi-spin pi-spinner" />
        <i v-else class="pi pi-send" />
      </button>
    </div>
  </div>
  </div>
</template>

<style scoped>
.pc-wrap {
  position: fixed;
  right: 1.25rem;
  bottom: 5.5rem;
  z-index: 1100;
  width: min(420px, calc(100vw - 2rem));
  max-height: min(620px, calc(100vh - 7rem));
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-xl);
  background: var(--surface-card);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
.pc-launcher { position:fixed; right:1.25rem; bottom:1.25rem; z-index:1101; width:58px; height:58px; border:0; border-radius:50%; display:grid; place-items:center; color:#fff; background:var(--brand-blue,#2563eb); box-shadow:0 10px 30px rgba(37,99,235,.35); cursor:pointer; font-size:1.25rem; }
.pc-launcher-badge { position:absolute; right:-3px; top:-4px; min-width:21px; height:21px; padding:0 5px; display:grid; place-items:center; border:2px solid #fff; border-radius:999px; background:#ef4444; color:#fff; font-size:.62rem; font-weight:800; }
.pc-close { margin-left:.15rem; width:28px; height:28px; display:grid; place-items:center; border:0; border-radius:50%; background:transparent; color:var(--text-muted); cursor:pointer; }

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
@media (max-width: 640px) {
  .pc-wrap { left:.75rem; right:.75rem; bottom:5rem; width:auto; max-height:calc(100dvh - 6.5rem); }
  .pc-list { max-height:calc(100dvh - 18rem); }
  .pc-launcher { right:1rem; bottom:1rem; }
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
.pc-attachments { display:flex; flex-wrap:wrap; gap:.4rem; margin-top:.3rem; }
.pc-attachments button { display:flex; align-items:center; gap:.35rem; padding:.4rem .55rem; border:1px solid var(--border-light); border-radius:9px; color:var(--brand-blue); font-size:.7rem; background:#fff; cursor:pointer; }
.pc-attachments img { display:block; width:150px; max-width:100%; height:110px; object-fit:cover; border-radius:7px; }
.pc-attachments button:has(img) { flex-direction:column; align-items:flex-start; padding:.35rem; }
.pc-mentions { display:grid; padding:.35rem .75rem; border-top:1px solid var(--border-light); background:#fff; }
.pc-mentions button { padding:.4rem; border:0; background:transparent; text-align:left; cursor:pointer; color:var(--text-primary); }
.pc-mentions button:hover { background:#eff6ff; }
.pc-selected-files { display:flex; gap:.35rem; flex-wrap:wrap; padding:.4rem .75rem 0; font-size:.65rem; }
.pc-selected-files span { padding:.25rem .4rem; background:#eef2ff; border-radius:6px; }
.pc-file-input { display:none; }
.pc-attach-btn { width:38px; height:38px; border:0; border-radius:50%; background:#eef2f7; color:var(--text-secondary); cursor:pointer; }

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
