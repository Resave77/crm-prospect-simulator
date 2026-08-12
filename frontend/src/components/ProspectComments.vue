<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getProspectComments, addProspectComment, deleteProspectComment, downloadCommentAttachment, getMentionUsers } from '../api/crm'
import type { ProspectComment, SalesExecutiveOption } from '../types/crm'
import type { UserRole } from '../types/auth'

const props = withDefaults(defineProps<{
  prospectId: string
  role: UserRole
  embedded?: boolean
}>(), { embedded: false })

const auth = useAuthStore()
const comments = ref<ProspectComment[]>([])
const newComment = ref('')
const files = ref<File[]>([])
const mentionUsers = ref<SalesExecutiveOption[]>([])
const fileInput = ref<HTMLInputElement | null>(null)
const loading = ref(true)
const sending = ref(false)
const deletingId = ref<string | null>(null)
const error = ref('')
const listRef = ref<HTMLElement | null>(null)
const open = ref(false)
const imagePreviews = ref<Record<string, string>>({})
const unavailableAttachments = ref<Record<string, boolean>>({})
const imageViewer = ref<{ url: string; name: string } | null>(null)
const previewImage = ref<{ name: string; url: string } | null>(null)
const newCount = ref(0)
let lastReadAt = ''
let pollId: number | undefined

function scrollToBottom() {
  nextTick(() => {
    if (listRef.value) {
      listRef.value.scrollTop = listRef.value.scrollHeight
    }
  })
}

function readKey() {
  return `yummy-crm:comments-read:${props.role}:${props.prospectId}`
}

function loadLastReadAt() {
  try { lastReadAt = localStorage.getItem(readKey()) ?? '' } catch { lastReadAt = '' }
}

function countUnread(items: ProspectComment[]) {
  if (!lastReadAt) return 0
  const readTime = new Date(lastReadAt).getTime()
  return items.filter((c) => c.createdAt && new Date(c.createdAt).getTime() > readTime).length
}

function markRead() {
  const latest = comments.value.reduce((max, c) => (c.createdAt && c.createdAt > max ? c.createdAt : max), '')
  lastReadAt = latest || new Date().toISOString()
  try { localStorage.setItem(readKey(), lastReadAt) } catch { /* ignore */ }
  newCount.value = 0
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

function canDeleteComment(comment: ProspectComment): boolean {
  return props.role === 'SUPER_ADMIN' || props.role === 'ADMINISTRATOR' || isOwnComment(comment)
}

async function ensureImagePreviews(items: ProspectComment[]) {
  const images = items.flatMap((comment) => comment.attachments ?? []).filter((file) => file.contentType.startsWith('image/'))
  await Promise.all(images.map(async (file) => {
    if (imagePreviews.value[file.id]) return
    try {
      const blob = await downloadCommentAttachment(props.prospectId, file.id, props.role)
      imagePreviews.value[file.id] = URL.createObjectURL(blob)
    } catch {
      unavailableAttachments.value = { ...unavailableAttachments.value, [file.id]: true }
    }
  }))
}

async function load() {
  try {
    const [items, users] = await Promise.all([
      getProspectComments(props.prospectId, props.role),
      getMentionUsers(props.role).catch(() => [] as SalesExecutiveOption[]),
    ])
    comments.value = items
    loadLastReadAt()
    newCount.value = lastReadAt ? countUnread(items) : 0
    if (!lastReadAt) markRead()
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
    markRead()
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
  if (unavailableAttachments.value[id]) {
    error.value = 'This attachment is no longer available. Please upload the photo again.'
    return
  }
  try {
    const blob = await downloadCommentAttachment(props.prospectId, id, props.role)
    const url = URL.createObjectURL(blob)
    if (blob.type.startsWith('image/')) {
      imageViewer.value = { url, name }
      return
    }
    const anchor = document.createElement('a'); anchor.href = url; anchor.download = name; anchor.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  } catch {
    unavailableAttachments.value = { ...unavailableAttachments.value, [id]: true }
    error.value = 'This attachment is no longer available. Please upload the photo again.'
  }
}

async function deleteComment(comment: ProspectComment) {
  if (deletingId.value || !window.confirm('Delete this comment and its attachments?')) return
  deletingId.value = comment.id
  error.value = ''
  try {
    await deleteProspectComment(props.prospectId, comment.id, props.role)
    for (const file of comment.attachments ?? []) {
      const preview = imagePreviews.value[file.id]
      if (preview) URL.revokeObjectURL(preview)
      delete imagePreviews.value[file.id]
    }
    comments.value = comments.value.filter((item) => item.id !== comment.id)
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Failed to delete comment.'
  } finally {
    deletingId.value = null
  }
}

function showImage(id: string, name: string) {
  const url = imagePreviews.value[id]
  if (url) previewImage.value = { name, url }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submit()
  }
}

watch(() => comments.value.length, () => scrollToBottom())
watch(open, (value) => {
  if (value) {
    markRead()
    scrollToBottom()
  }
})
watch(() => props.prospectId, () => {
  lastReadAt = ''
  newCount.value = 0
  Object.values(imagePreviews.value).forEach((url) => URL.revokeObjectURL(url))
  imagePreviews.value = {}
  unavailableAttachments.value = {}
  if (imageViewer.value) URL.revokeObjectURL(imageViewer.value.url)
  imageViewer.value = null
  loading.value = true
  comments.value = []
  load()
})

onMounted(() => {
  load()
  pollId = window.setInterval(async () => {
    if (open.value || sending.value) return
    try {
    const items = await getProspectComments(props.prospectId, props.role)
    newCount.value = countUnread(items)
    comments.value = items
    await ensureImagePreviews(comments.value)
    } catch { /* keep the last known badge count */ }
  }, 30000)
})
onBeforeUnmount(() => {
  if (pollId) window.clearInterval(pollId)
  Object.values(imagePreviews.value).forEach((url) => URL.revokeObjectURL(url))
  if (imageViewer.value) URL.revokeObjectURL(imageViewer.value.url)
})
</script>

<template>
  <div :class="['pc-floating', { 'pc-embedded': embedded }]">
    <button v-if="!embedded && !open" class="pc-launcher" aria-label="Open prospect discussion" @click="open = true">
      <i class="pi pi-comments" />
      <span v-if="newCount" class="pc-launcher-badge">{{ newCount > 99 ? '99+' : newCount }}</span>
    </button>
     <div v-if="open || embedded" class="pc-wrap">
    <div class="pc-header">
      <i class="pi pi-comments" />
      <span>Discussion</span>
      <span v-if="!loading" class="pc-count">{{ comments.length }}</span>
       <button v-if="!embedded" class="pc-close" aria-label="Close discussion" @click="open = false"><i class="pi pi-times" /></button>
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
            <button v-if="canDeleteComment(c)" class="pc-delete-btn" :disabled="deletingId === c.id" title="Delete comment" @click="deleteComment(c)">
              <i :class="deletingId === c.id ? 'pi pi-spin pi-spinner' : 'pi pi-trash'" />
            </button>
          </div>
          <p v-if="c.content" class="pc-msg-text">{{ c.content }}</p>
          <div v-if="c.attachments?.length" class="pc-attachments">
             <button v-for="file in c.attachments" :key="file.id" :disabled="unavailableAttachments[file.id]" @click="file.contentType.startsWith('image/') && imagePreviews[file.id] ? showImage(file.id, file.name) : openAttachment(file.id, file.name)">
               <img v-if="file.contentType.startsWith('image/') && imagePreviews[file.id]" :src="imagePreviews[file.id]" :alt="file.name" />
               <i v-else :class="unavailableAttachments[file.id] ? 'pi pi-exclamation-triangle' : (file.contentType.startsWith('image/') ? 'pi pi-image' : 'pi pi-file')" />
               <span>{{ unavailableAttachments[file.id] ? 'File unavailable' : file.name }}</span>
            </button>
     </div>
     <div v-if="imageViewer" class="pc-image-viewer" @click.self="imageViewer = null">
       <button class="pc-image-viewer-close" aria-label="Close image preview" @click="imageViewer = null"><i class="pi pi-times" /></button>
       <img :src="imageViewer.url" :alt="imageViewer.name" />
       <span>{{ imageViewer.name }}</span>
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
  <div v-if="previewImage" class="pc-lightbox" @click.self="previewImage = null">
    <button aria-label="Close image" @click="previewImage = null"><i class="pi pi-times" /></button>
    <img :src="previewImage.url" :alt="previewImage.name" />
    <span>{{ previewImage.name }}</span>
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
.pc-embedded { width: 100%; }
.pc-embedded .pc-wrap {
  position: static;
  width: 100%;
  max-height: none;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}
.pc-launcher { position:fixed; right:1.25rem; bottom:1.25rem; z-index:1101; width:58px; height:58px; border:0; border-radius:50%; display:grid; place-items:center; color:#fff; background:var(--brand-blue,#e63946); box-shadow:0 10px 30px rgba(230,57,70,.35); cursor:pointer; font-size:1.25rem; }
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
  border-radius: 9999px; background: var(--brand-blue-bg, #fff0f1);
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
  background: linear-gradient(135deg, #e63946, #d62839); color: #fff;
}

.pc-msg-body { flex: 1; min-width: 0; }
.pc-msg-meta {
  display: flex; align-items: baseline; gap: 0.4rem;
  margin-bottom: 0.15rem;
}
.pc-msg-meta strong { font-size: 0.75rem; color: var(--text-primary); }
.pc-msg-meta span { font-size: 0.62rem; color: var(--text-muted); }
.pc-msg-own .pc-msg-meta { flex-direction: row-reverse; }
.pc-delete-btn { margin-left: auto; padding: .15rem; border: 0; background: transparent; color: var(--text-muted); cursor: pointer; opacity: .5; }
.pc-delete-btn:hover:not(:disabled) { color: #dc2626; opacity: 1; }
.pc-delete-btn:disabled { cursor: wait; opacity: .5; }

.pc-msg-text {
  margin: 0; padding: 0.55rem 0.75rem;
  background: #f1f5f9; border-radius: 12px 12px 12px 4px;
  font-size: 0.8rem; line-height: 1.5; color: var(--text-secondary);
  word-break: break-word;
}
.pc-msg-own .pc-msg-text {
  background: var(--brand-blue, #e63946); color: #fff;
  border-radius: 12px 12px 4px 12px;
}
.pc-attachments { display:flex; flex-wrap:wrap; gap:.4rem; margin-top:.3rem; }
.pc-attachments button { display:flex; align-items:center; gap:.35rem; padding:.4rem .55rem; border:1px solid var(--border-light); border-radius:9px; color:var(--brand-blue); font-size:.7rem; background:#fff; cursor:pointer; }
.pc-attachments img { display:block; width:150px; max-width:100%; height:110px; object-fit:cover; border-radius:7px; }
.pc-attachments button:has(img) { flex-direction:column; align-items:flex-start; padding:.35rem; }
.pc-image-viewer { position:fixed; inset:0; z-index:1200; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:.6rem; padding:1.5rem; background:rgba(15,23,42,.92); }
.pc-image-viewer img { max-width:min(95vw,900px); max-height:85vh; object-fit:contain; border-radius:12px; box-shadow:0 20px 60px rgba(0,0,0,.45); }
.pc-image-viewer span { max-width:90vw; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:#fff; font-size:.75rem; }
.pc-image-viewer-close { position:absolute; top:1rem; right:1rem; width:40px; height:40px; display:grid; place-items:center; border:0; border-radius:50%; background:rgba(255,255,255,.15); color:#fff; cursor:pointer; }
.pc-mentions { display:grid; padding:.35rem .75rem; border-top:1px solid var(--border-light); background:#fff; }
.pc-mentions button { padding:.4rem; border:0; background:transparent; text-align:left; cursor:pointer; color:var(--text-primary); }
.pc-mentions button:hover { background:#fff0f1; }
.pc-selected-files { display:flex; gap:.35rem; flex-wrap:wrap; padding:.4rem .75rem 0; font-size:.65rem; }
.pc-selected-files span { padding:.25rem .4rem; background:#fff5f5; border-radius:6px; }
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
  border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(230, 57, 70, 0.08);
  background: #fff;
}
.pc-input-row textarea::placeholder { color: var(--text-muted); }
.pc-input-row textarea:disabled { opacity: 0.6; }

.pc-send-btn {
  width: 38px; height: 38px; flex-shrink: 0;
  display: grid; place-items: center; border-radius: 50%;
  border: 0; background: var(--brand-blue, #e63946); color: #fff;
  cursor: pointer; font-size: 0.85rem;
  transition: background 0.15s ease, transform 0.15s ease;
}
.pc-send-btn:hover:not(:disabled) { background: #d62839; transform: scale(1.05); }
.pc-send-btn:disabled { opacity: 0.45; cursor: not-allowed; }
.pc-lightbox { position:fixed; inset:0; z-index:2200; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:.65rem; padding:1rem; background:rgba(15,23,42,.92); }
.pc-lightbox img { max-width:min(960px,100%); max-height:calc(100dvh - 6rem); object-fit:contain; border-radius:12px; }
.pc-lightbox span { color:#fff; font-size:.75rem; }
.pc-lightbox button { position:absolute; top:1rem; right:1rem; width:40px; height:40px; display:grid; place-items:center; border:0; border-radius:50%; background:rgba(255,255,255,.15); color:#fff; cursor:pointer; }

/* ── Refined discussion experience (shared by Sales and Admin) ── */
.pc-wrap {
  width: min(390px, calc(100vw - 2rem));
  max-height: min(570px, calc(100dvh - 7rem));
  border-color: #eadde0;
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 24px 60px -18px rgba(73, 34, 41, 0.28), 0 8px 22px rgba(73, 34, 41, 0.08);
}
.pc-embedded .pc-wrap { border-radius: 18px; box-shadow: 0 8px 24px rgba(73,34,41,.07); }
.pc-launcher {
  width: 48px;
  height: 48px;
  border: 3px solid rgba(255,255,255,.9);
  background: linear-gradient(145deg, #ef4e5d, #e63946 58%, #d62839);
  box-shadow: 0 12px 28px rgba(214,40,57,.35);
  transition: transform .2s ease, box-shadow .2s ease;
}
.pc-launcher:hover { transform: translateY(-2px); box-shadow: 0 16px 32px rgba(214,40,57,.4); }
.pc-header {
  position: relative;
  min-height: 56px;
  padding: 0.7rem 0.8rem;
  gap: 0.65rem;
  overflow: hidden;
  border-bottom: 0;
  color: #fff;
  background: linear-gradient(120deg, #e63946, #d62839 64%, #d62839);
  font-size: 0.78rem;
  letter-spacing: 0;
  text-transform: none;
}
.pc-header::after { content:''; position:absolute; right:-34px; top:-55px; width:130px; height:130px; border-radius:50%; background:rgba(255,255,255,.09); }
.pc-header > i { width: 30px; height: 30px; display:grid; place-items:center; border-radius:9px; color:#e63946; background:#fff; font-size:.82rem; box-shadow:0 4px 10px rgba(135,28,40,.14); }
.pc-header > span:not(.pc-count) { font-size: .8rem; font-weight: 800; }
.pc-count { margin-left:auto; min-width:26px; height:24px; padding:0 7px; color:#fff; background:rgba(255,255,255,.17); border:1px solid rgba(255,255,255,.24); }
.pc-close { position:relative; z-index:1; width:32px; height:32px; color:#fff; background:rgba(255,255,255,.12); }
.pc-close:hover { background:rgba(255,255,255,.22); }
.pc-list {
  max-height: 320px;
  gap: 0.8rem;
  padding: 1rem;
  background: linear-gradient(180deg, #fff 0, #fdfafa 100%);
  scrollbar-width: thin;
  scrollbar-color: #eadde0 transparent;
}
.pc-msg { gap:.55rem; }
.pc-msg-avatar { width:30px; height:30px; border-radius:9px; border:1px solid #e6dadd; background:#fff; color:#625357; box-shadow:0 2px 6px rgba(73,34,41,.05); font-size:.65rem; }
.pc-msg-avatar-own { border-color:#ffd9dc; color:#fff; background:linear-gradient(145deg,#ef4e5d,#e63946); }
.pc-msg-body { max-width: calc(100% - 38px); }
.pc-msg-meta { margin-bottom:.28rem; gap:.45rem; }
.pc-msg-meta strong { font-size:.72rem; }
.pc-msg-meta span { font-size:.58rem; }
.pc-msg-text {
  padding:.55rem .68rem;
  border:1px solid #eee3e5;
  border-radius:4px 14px 14px 14px;
  background:#fff;
  color:#493c40;
  box-shadow:0 3px 10px rgba(73,34,41,.04);
}
.pc-msg-own .pc-msg-text {
  border-color:#e63946;
  border-radius:14px 4px 14px 14px;
  background:linear-gradient(145deg,#ef4e5d,#e63946);
  box-shadow:0 5px 14px rgba(214,40,57,.16);
}
.pc-delete-btn { width:24px; height:24px; display:grid; place-items:center; border-radius:7px; }
.pc-delete-btn:hover:not(:disabled) { background:#fef2f2; }
.pc-empty { min-height:210px; justify-content:center; gap:.55rem; }
.pc-empty i { width:52px; height:52px; display:grid; place-items:center; border-radius:16px; color:#e63946; background:#fff0f1; opacity:1; font-size:1.15rem; }
.pc-empty span { max-width:220px; line-height:1.5; }
.pc-attachments button { border-color:#eadde0; border-radius:11px; background:#fff; box-shadow:0 2px 7px rgba(73,34,41,.04); }
.pc-attachments img { border-radius:9px; }
.pc-selected-files { padding:.5rem .85rem; border-top:1px solid #f1e8ea; background:#fff9fa; }
.pc-selected-files span { display:inline-flex; align-items:center; gap:.25rem; border:1px solid #ffd9dc; border-radius:8px; background:#fff0f1; color:#d62839; }
.pc-mentions { padding:.4rem; border:1px solid #eadde0; border-bottom:0; background:#fff; }
.pc-mentions button { border-radius:8px; font-size:.72rem; }
.pc-input-row { gap:.4rem; padding:.55rem; border-top:1px solid #eadde0; background:#fff; }
.pc-attach-btn { width:36px; height:36px; border:1px solid #eadde0; border-radius:10px; background:#fcf9f9; color:#8d7d81; }
.pc-attach-btn:hover { color:#e63946; border-color:#f4b3ba; background:#fff0f1; }
.pc-input-row textarea { min-height:36px; padding:.5rem .65rem; border-color:#e6dadd; border-radius:10px; background:#fcf9f9; }
.pc-send-btn { width:36px; height:36px; border-radius:10px; background:linear-gradient(145deg,#ef4e5d,#e63946); box-shadow:0 5px 12px rgba(214,40,57,.2); }
.pc-send-btn:hover:not(:disabled) { background:#d62839; transform:translateY(-1px); }

@media (max-width: 640px) {
  .pc-wrap { left:.65rem; right:.65rem; width:auto; border-radius:20px; }
  .pc-header { min-height:60px; padding:.75rem .85rem; }
  .pc-list { padding:.8rem; gap:.7rem; }
  .pc-msg-body { max-width:calc(100% - 40px); }
  .pc-msg-text { font-size:.76rem; }
  .pc-input-row { padding:.6rem; }
}
</style>
