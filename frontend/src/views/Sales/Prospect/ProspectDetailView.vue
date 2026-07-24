<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyProspect, transitionProspect } from '../../../api/crm'
import { nextStage } from '../../../domain/pipeline'
import type { ProspectReview } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'
import { formatPlaceType, businessStatusLabel, isValidWebsite, websiteDisplayUrl, isValidPhone, copyToClipboard } from '../../../utils/placeDetails'

const route = useRoute()
const review = ref<ProspectReview | null>(null)
const error = ref('')
const success = ref('')
const loading = ref(true)

const userCoords = ref<{ lat: number; lng: number } | null>(null)
let geoWatchId: number | null = null

const pipelineNotes = ref('')
const transitionBusy = ref(false)
const transitionDone = ref(false)

const openVisit = computed(() => review.value?.visits.find((v) => !v.checkOutAt) ?? null)

const nextPipelineStage = computed(() => {
  if (!review.value) return null
  return nextStage(review.value.prospect.status)
})

const statusSeverity = computed(() => {
  const s = review.value?.prospect.status
  if (s === 'WON' || s === 'CONVERTED') return 'success' as const
  if (s === 'LOST') return 'danger' as const
  if (s === 'NEW_LEAD') return 'info' as const
  if (s === 'QUALIFIED') return 'success' as const
  return 'warn' as const
})

const displayTypes = computed(() => {
  const types = review.value?.prospect.placeTypes
  if (!Array.isArray(types) || !types.length) return []
  const skip = new Set(['establishment', 'point_of_interest', 'food', 'store'])
  return types.filter((t) => !skip.has(t)).slice(0, 5)
})

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function message(err: unknown) {
  return (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message
    ?? (err instanceof Error ? err.message : 'Unable to complete the request.')
}

function navigate() {
  const p = review.value?.prospect
  if (!p) return
  openGoogleMapsNavigation({
    latitude: p.latitude,
    longitude: p.longitude,
    address: p.formattedAddress,
    googleMapsUrl: p.googleMapsUrl,
  })
}

function acquireGPS() {
  if (!navigator.geolocation) return
  geoWatchId = navigator.geolocation.watchPosition(
    (pos) => { userCoords.value = { lat: pos.coords.latitude, lng: pos.coords.longitude } },
    () => {},
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function formatVisitDate(iso: string) {
  return new Date(iso).toLocaleDateString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function calcDuration(checkIn: string, checkOut: string) {
  const diff = new Date(checkOut).getTime() - new Date(checkIn).getTime()
  const mins = Math.floor(diff / 60000)
  const hrs = Math.floor(mins / 60)
  const remMins = mins % 60
  return hrs > 0 ? `${hrs}h ${remMins}m` : `${remMins}m`
}

async function doTransition() {
  if (!review.value || !nextPipelineStage.value) return
  error.value = ''; success.value = ''; transitionBusy.value = true
  try {
    await transitionProspect(String(route.params.id), nextPipelineStage.value, pipelineNotes.value)
    transitionDone.value = true
    success.value = `Pipeline updated to ${nextPipelineStage.value.replaceAll('_', ' ')}.`
    pipelineNotes.value = ''
    review.value = await getMyProspect(String(route.params.id))
  } catch (caught) { error.value = message(caught) } finally { transitionBusy.value = false }
}

function handleCopy(text: string) {
  copyToClipboard(text)
  success.value = 'Copied to clipboard.'
  setTimeout(() => { if (success.value === 'Copied to clipboard.') success.value = '' }, 2000)
}

onMounted(async () => {
  acquireGPS()
  try {
    review.value = await getMyProspect(String(route.params.id))
  } catch (caught) { error.value = message(caught) } finally { loading.value = false }
})

onBeforeUnmount(() => { if (geoWatchId != null) navigator.geolocation?.clearWatch(geoWatchId) })
</script>

<template>
  <section class="detail-page">
    <RouterLink class="back-link" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> My Prospects</RouterLink>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div v-if="loading" class="detail-skeleton">
      <div class="sk-header"><div class="sk-circle" /><div class="sk-lines"><div class="sk-line w70" /><div class="sk-line w40" /></div></div>
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /></div>
    </div>

    <div v-else-if="!review" class="detail-empty">
      <div class="detail-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>Prospect not found</strong>
      <span>This prospect may have been removed or you don't have access.</span>
      <RouterLink class="detail-empty-btn" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> Back to prospects</RouterLink>
    </div>

    <template v-else>
      <!-- Summary Card -->
      <div class="dcard dcard-summary">
        <div class="dcard-summary-top">
          <div class="dcard-avatar dcard-avatar-prospect">{{ initials(review.prospect.placeName || 'Prospect') }}</div>
          <div class="dcard-identity">
            <p class="eyebrow">Prospect</p>
            <h1>{{ review.prospect.placeName }}</h1>
          </div>
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" :severity="statusSeverity" />
        </div>
        <div class="dcard-rows">
          <div v-if="review.prospect.placeCategory" class="dcard-row"><i class="pi pi-bookmark" /><span>{{ review.prospect.placeCategory }}</span></div>
          <div v-if="review.prospect.industryGroup" class="dcard-row"><i class="pi pi-tag" /><span>{{ review.prospect.industryGroup }}</span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span>{{ review.prospect.assignedSalesExecutive }}</span></div>
        </div>
        <div v-if="displayTypes.length" class="dcard-type-badges">
          <span v-for="t in displayTypes" :key="t" class="dcard-type-badge">{{ formatPlaceType(t) }}</span>
        </div>
      </div>

      <!-- Active Visit Alert -->
      <div v-if="openVisit" class="dcard dcard-active-visit">
        <div class="dcard-active-visit-row">
          <i class="pi pi-sign-in" />
          <span>You have an <strong>active visit</strong> in progress.</span>
          <RouterLink class="dcard-active-visit-link" :to="`/sales/my-prospects/${review.prospect.id}/check-out`">
            Check out <i class="pi pi-arrow-right" />
          </RouterLink>
        </div>
      </div>

      <!-- Contact Card -->
      <div v-if="review.prospect.phoneNumber || review.prospect.websiteUrl" class="dcard">
        <h2>Contact</h2>
        <div class="dcard-rows">
          <div v-if="review.prospect.phoneNumber && isValidPhone(review.prospect.phoneNumber)" class="dcard-row">
            <i class="pi pi-phone" /><a :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a>
          </div>
          <div v-if="review.prospect.websiteUrl && isValidWebsite(review.prospect.websiteUrl)" class="dcard-row">
            <i class="pi pi-globe" /><a :href="review.prospect.websiteUrl.startsWith('http') ? review.prospect.websiteUrl : `https://${review.prospect.websiteUrl}`" target="_blank" rel="noopener noreferrer">{{ websiteDisplayUrl(review.prospect.websiteUrl) }}</a>
          </div>
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard">
        <div class="dcard-header-row">
          <h2>Location</h2>
          <span v-if="review.prospect.latitude != null && review.prospect.longitude != null && userCoords" class="dcard-distance-pill">
            <i class="pi pi-compass" /> {{ formatDistance(getDistanceTo(review.prospect.latitude, review.prospect.longitude, userCoords.lat, userCoords.lng)!) }} away
          </span>
        </div>
        <EntityLocationMap
          :latitude="review.prospect.latitude"
          :longitude="review.prospect.longitude"
          :label="review.prospect.placeName"
          :interactive="false"
          height="200px"
        />
        <div class="dcard-location-rows">
          <div class="dcard-row"><i class="pi pi-map-marker" /><span>{{ review.prospect.formattedAddress || 'No address' }}</span></div>
          <div v-if="review.prospect.latitude != null && review.prospect.longitude != null" class="dcard-row dcard-row-coords">
            <i class="pi pi-compass" />
            <span>GPS: {{ review.prospect.latitude?.toFixed(6) }}, {{ review.prospect.longitude?.toFixed(6) }}</span>
            <button class="dcard-copy-btn" title="Copy coordinates" @click="handleCopy(`${review.prospect.latitude}, ${review.prospect.longitude}`)"><i class="pi pi-copy" /></button>
          </div>
          <a v-if="review.prospect.googleMapsUrl" :href="review.prospect.googleMapsUrl" target="_blank" rel="noopener noreferrer" class="dcard-row dcard-row-link">
            <i class="pi pi-external-link" /><span>Open in Google Maps</span>
          </a>
        </div>
      </div>

      <!-- Google Place Info -->
      <div v-if="review.prospect.googlePlaceId" class="dcard">
        <h2>Google Place</h2>
        <div class="dcard-rows">
          <div class="dcard-row">
            <i class="pi pi-info-circle" />
            <span class="dcard-place-id">
              <span>Place ID</span>
              <code>{{ review.prospect.googlePlaceId }}</code>
            </span>
            <button class="dcard-copy-btn" title="Copy Place ID" @click="handleCopy(review.prospect.googlePlaceId)"><i class="pi pi-copy" /></button>
          </div>
        </div>
      </div>

      <!-- Pipeline Update (only when visit active) -->
      <div v-if="openVisit && nextPipelineStage && !transitionDone" class="dcard dcard-highlight">
        <h2>Pipeline Update</h2>
        <p class="dcard-hint">Move this prospect to the next pipeline stage.</p>
        <div class="dcard-pipeline-flow">
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" severity="secondary" />
          <i class="pi pi-arrow-right" />
          <Tag :value="nextPipelineStage.replaceAll('_', ' ')" severity="info" />
        </div>
        <label class="dcard-field"><span>Pipeline Notes</span><textarea v-model="pipelineNotes" rows="2" class="dcard-textarea" placeholder="Optional notes for this transition..." /></label>
        <button class="dcard-btn dcard-btn-primary" :disabled="transitionBusy" @click="doTransition">
          <i class="pi pi-arrow-right" /> {{ transitionBusy ? 'Updating...' : 'Update Pipeline' }}
        </button>
      </div>
      <div v-else-if="transitionDone && openVisit" class="dcard dcard-success">
        <div class="dcard-success-row"><i class="pi pi-check-circle" /><span>Pipeline updated for this visit.</span></div>
      </div>

      <!-- Visit History -->
      <div class="dcard">
        <h2>Visit History</h2>
        <div v-if="review.visits.length" class="dcard-visit-list">
          <div v-for="visit in review.visits" :key="visit.id" class="dcard-visit">
            <div class="dcard-visit-header">
              <Tag :value="visit.checkOutAt ? 'Completed' : 'Active'" :severity="visit.checkOutAt ? 'secondary' : 'success'" />
              <span>{{ formatVisitDate(visit.checkInAt) }}</span>
            </div>
            <div class="dcard-visit-body">
              <div class="dcard-visit-detail"><i class="pi pi-sign-in" /><span>Check-in: {{ visit.checkInLatitude.toFixed(4) }}, {{ visit.checkInLongitude.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="dcard-visit-detail"><i class="pi pi-sign-out" /><span>Check-out: {{ visit.checkOutLatitude?.toFixed(4) }}, {{ visit.checkOutLongitude?.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="dcard-visit-detail"><i class="pi pi-clock" /><span>Duration: {{ calcDuration(visit.checkInAt, visit.checkOutAt) }}</span></div>
              <div v-if="visit.visitNotes" class="dcard-visit-detail"><i class="pi pi-comment" /><span>{{ visit.visitNotes }}</span></div>
              <div v-if="visit.followUpNotes" class="dcard-visit-detail"><i class="pi pi-directions" /><span>Follow-up: {{ visit.followUpNotes }}</span></div>
              <div class="dcard-visit-detail dcard-visit-exec"><i class="pi pi-user" /><span>{{ visit.salesExecutiveName }}</span></div>
            </div>
          </div>
        </div>
        <p v-else class="dcard-empty-text">No visits recorded yet.</p>
      </div>

      <!-- Status History -->
      <div class="dcard">
        <h2>Status History</h2>
        <div v-if="review.history.length" class="dcard-timeline">
          <div v-for="entry in review.history" :key="entry.id" class="dcard-timeline-entry">
            <div class="dcard-timeline-dot" />
            <div class="dcard-timeline-content">
              <strong>{{ (entry.fromStatus || 'Created').replaceAll('_', ' ') }} → {{ entry.toStatus.replaceAll('_', ' ') }}</strong>
              <span>{{ new Date(entry.createdAt).toLocaleString() }} · {{ entry.changedByName }}</span>
              <p v-if="entry.notes">{{ entry.notes }}</p>
            </div>
          </div>
        </div>
        <p v-else class="dcard-empty-text">No status changes recorded.</p>
      </div>

      <!-- Bottom Action Bar -->
      <div class="detail-bottom-bar">
        <button class="dbar-btn dbar-navigate" :disabled="review.prospect.latitude == null && review.prospect.longitude == null && !review.prospect.formattedAddress" @click="navigate">
          <i class="pi pi-directions" /> Navigate
        </button>
        <a v-if="review.prospect.phoneNumber && isValidPhone(review.prospect.phoneNumber)" :href="`tel:${review.prospect.phoneNumber}`" class="dbar-btn dbar-call">
          <i class="pi pi-phone" /> Call
        </a>
        <span v-else class="dbar-btn dbar-call dbar-disabled"><i class="pi pi-phone" /> Call</span>
        <RouterLink v-if="openVisit" class="dbar-btn dbar-checkout" :to="`/sales/my-prospects/${review.prospect.id}/check-out`">
          <i class="pi pi-sign-out" /> Check out
        </RouterLink>
        <RouterLink v-else class="dbar-btn dbar-checkin" :to="`/sales/my-prospects/${review.prospect.id}/check-in`">
          <i class="pi pi-sign-in" /> Check in
        </RouterLink>
      </div>
    </template>
  </section>
</template>

<style scoped>
.detail-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: 5.5rem; }

.detail-skeleton { display: grid; gap: 0.85rem; }
.sk-header { display: flex; align-items: center; gap: 0.7rem; }
.sk-circle { width: 48px; height: 48px; border-radius: 50%; background: #e2e8f0; flex-shrink: 0; }
.sk-lines { flex: 1; display: flex; flex-direction: column; gap: 0.4rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w70 { width: 70%; }
.sk-line.w60 { width: 60%; }
.sk-line.w50 { width: 50%; }
.sk-line.w40 { width: 40%; }
.sk-line.w80 { width: 80%; }
.sk-map { height: 180px; border-radius: 12px; background: #e2e8f0; }

.detail-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.detail-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.detail-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.detail-empty span { color: var(--text-muted); font-size: 0.8rem; max-width: 260px; }
.detail-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; }

.dcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.dcard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

.dcard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.dcard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.dcard-avatar { width: 52px; height: 52px; display: grid; place-items: center; border-radius: 16px; color: #fff; font-weight: 800; font-size: 1rem; flex-shrink: 0; }
.dcard-avatar-prospect { background: linear-gradient(135deg, #2563eb, #1d4ed8); box-shadow: 0 3px 10px rgba(37, 99, 235, 0.25); }
.dcard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-identity .eyebrow { margin: 0; }
.dcard-identity h1 { margin: 0; font-size: 1.2rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

.dcard-type-badges { display: flex; flex-wrap: wrap; gap: 0.3rem; }
.dcard-type-badge {
  display: inline-block; padding: 0.15rem 0.5rem; border-radius: 6px;
  background: #eff6ff; color: #1d4ed8; font-size: 0.58rem; font-weight: 600; line-height: 1.5;
}

.dcard-active-visit { border-color: #fbbf24; background: #fffbeb; }
.dcard-active-visit-row { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8rem; color: #92400e; flex-wrap: wrap; }
.dcard-active-visit-row > i { font-size: 1rem; color: #f59e0b; }
.dcard-active-visit-link {
  display: inline-flex; align-items: center; gap: 0.25rem; margin-left: auto;
  padding: 0.35rem 0.75rem; border-radius: 10px; background: #f59e0b; color: #fff;
  font-size: 0.7rem; font-weight: 700; text-decoration: none; white-space: nowrap;
}
.dcard-active-visit-link:hover { background: #d97706; }

.dcard-header-row { display: flex; align-items: center; justify-content: space-between; }
.dcard-header-row h2 { margin: 0; }
.dcard-distance-pill {
  display: inline-flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.55rem;
  border-radius: 9999px; background: #eff6ff; color: var(--brand-blue);
  font-size: 0.62rem; font-weight: 700; white-space: nowrap;
}

.dcard-location-rows, .dcard-rows { display: grid; gap: 0.45rem; }
.dcard-row { display: flex; align-items: flex-start; gap: 0.55rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.dcard-row i { color: var(--text-muted); font-size: 0.72rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.1rem; }
.dcard-row a { color: var(--brand-blue); text-decoration: none; }
.dcard-row a:hover { text-decoration: underline; }
.dcard-row-link { cursor: pointer; }
.dcard-distance { color: var(--brand-blue); font-weight: 600; }
.dcard-row-coords { color: var(--text-muted); font-size: 0.75rem; }
.dcard-row-coords code { font-size: 0.7rem; color: var(--text-muted); background: #f1f5f9; padding: 0.1rem 0.3rem; border-radius: 4px; }

.dcard-place-id { display: flex; flex-direction: column; gap: 0.15rem; flex: 1; min-width: 0; }
.dcard-place-id span { color: var(--text-muted); font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.dcard-place-id code {
  font-size: 0.7rem; color: var(--text-secondary); background: #f1f5f9;
  padding: 0.2rem 0.4rem; border-radius: 6px; word-break: break-all; line-height: 1.4;
}
.dcard-copy-btn {
  display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px;
  border-radius: 8px; border: 1px solid var(--border-light); background: #fff;
  color: var(--text-muted); cursor: pointer; font-size: 0.65rem; flex-shrink: 0; transition: all 0.15s ease;
}
.dcard-copy-btn:hover { color: var(--brand-blue); border-color: #bfdbfe; background: #eff6ff; }

.dcard-highlight { border-color: var(--brand-blue-light); background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.dcard-pipeline-flow { display: flex; align-items: center; gap: 0.65rem; padding: 0.65rem; border-radius: 12px; background: var(--surface-card); border: 1px solid var(--border-light); justify-content: center; }
.dcard-pipeline-flow i { color: var(--brand-blue); font-size: 0.85rem; }
.dcard-success { border-color: #bbf7d0; background: #f0fdf4; }
.dcard-success-row { display: flex; align-items: center; gap: 0.5rem; color: #166534; font-size: 0.82rem; }
.dcard-success-row i { font-size: 1rem; }

.dcard-field { display: flex; flex-direction: column; gap: 0.3rem; }
.dcard-field span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.dcard-hint { margin: -0.35rem 0 0; color: var(--text-muted); font-size: 0.78rem; }
.dcard-textarea {
  padding: 0.65rem 0.85rem; border: 1px solid var(--border-light); border-radius: 12px;
  background: #f8fafc; color: var(--text-primary); font-size: 0.82rem; font-family: inherit;
  resize: vertical; width: 100%; box-sizing: border-box;
}
.dcard-textarea:focus { outline: 0; border-color: var(--brand-blue); }
.dcard-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.4rem;
  padding: 0.7rem 1rem; border-radius: 12px; border: none;
  font-size: 0.78rem; font-weight: 700; cursor: pointer; transition: all 0.15s ease;
}
.dcard-btn-primary { background: var(--brand-blue); color: #fff; }
.dcard-btn-primary:hover { background: #1d4ed8; }
.dcard-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.dcard-visit-list { display: grid; gap: 0.65rem; }
.dcard-visit { border: 1px solid var(--border-light); border-radius: 14px; overflow: hidden; }
.dcard-visit-header { display: flex; align-items: center; justify-content: space-between; padding: 0.55rem 0.85rem; background: #f8fafc; border-bottom: 1px solid var(--border-light); }
.dcard-visit-header span { color: var(--text-muted); font-size: 0.72rem; }
.dcard-visit-body { padding: 0.65rem 0.85rem; display: grid; gap: 0.35rem; }
.dcard-visit-detail { display: flex; align-items: flex-start; gap: 0.5rem; font-size: 0.78rem; color: var(--text-secondary); }
.dcard-visit-detail i { color: var(--text-muted); font-size: 0.68rem; margin-top: 0.18rem; flex-shrink: 0; }
.dcard-visit-exec { margin-top: 0.25rem; padding-top: 0.35rem; border-top: 1px solid var(--border-light); }

.dcard-timeline { display: grid; }
.dcard-timeline-entry { display: grid; grid-template-columns: 16px 1fr; gap: 0.75rem; padding-bottom: 1rem; position: relative; }
.dcard-timeline-entry:not(:last-child)::before { content: ''; position: absolute; left: 7px; top: 18px; bottom: 0; width: 2px; background: var(--border-light); }
.dcard-timeline-dot { width: 16px; height: 16px; border-radius: 50%; background: var(--brand-blue); border: 3px solid var(--brand-blue-bg); flex-shrink: 0; }
.dcard-timeline-content { display: grid; gap: 0.1rem; }
.dcard-timeline-content strong { font-size: 0.78rem; color: var(--text-primary); text-transform: capitalize; }
.dcard-timeline-content span { color: var(--text-muted); font-size: 0.68rem; }
.dcard-timeline-content p { margin: 0.2rem 0 0; font-size: 0.78rem; color: var(--text-secondary); line-height: 1.5; }

.dcard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1.5rem 0; }

.detail-bottom-bar {
  position: fixed; bottom: 0; left: 50%; transform: translateX(-50%);
  width: min(100%, 440px); z-index: 40;
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.5rem;
  padding: 0.75rem 1rem; padding-bottom: calc(0.75rem + env(safe-area-inset-bottom));
  background: var(--surface-card); border-top: 1px solid var(--border-light);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.dbar-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.65rem 0; border-radius: 12px; border: none;
  font-size: 0.72rem; font-weight: 700; cursor: pointer;
  text-decoration: none; text-align: center; transition: all 0.15s ease;
}
.dbar-navigate { background: var(--brand-blue); color: #fff; }
.dbar-navigate:hover { background: #1d4ed8; }
.dbar-navigate:disabled { background: #cbd5e1; cursor: not-allowed; }
.dbar-call { background: #f0fdf4; color: #059669; border: 1px solid #a7f3d0; }
.dbar-call:hover { background: #dcfce7; }
.dbar-disabled { opacity: 0.45; cursor: not-allowed; pointer-events: none; }
.dbar-checkin { background: #eff6ff; color: var(--brand-blue); border: 1px solid #bfdbfe; }
.dbar-checkin:hover { background: #dbeafe; }
.dbar-checkout { background: #fff7ed; color: #c2410c; border: 1px solid #fed7aa; }
.dbar-checkout:hover { background: #ffedd5; }

@media (max-width: 480px) {
  .detail-page { gap: 0.7rem; }
  .dcard { padding: 1rem; }
  .dcard-identity h1 { font-size: 1.05rem; }
}
</style>
