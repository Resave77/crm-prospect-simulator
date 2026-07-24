<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { checkInProspect, checkOutProspect, getMyProspect, transitionProspect } from '../../../api/crm'
import { nextStage } from '../../../domain/pipeline'
import type { ProspectReview } from '../../../types/crm'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'

const route = useRoute()
const review = ref<ProspectReview | null>(null)
const error = ref(''); const success = ref(''); const loading = ref(true); const visitBusy = ref(false)
const visitNotes = ref(''); const followUpNotes = ref(''); const selfiePlaceholder = ref(false)
const mapElement = ref<HTMLElement | null>(null); let map: L.Map | null = null
const openVisit = computed(() => review.value?.visits.find((visit) => !visit.checkOutAt) ?? null)

const deviceCoords = ref<{ latitude: number; longitude: number } | null>(null)
const gpsReady = ref(false)
const userCoords = ref<{ lat: number; lng: number } | null>(null)
const elapsed = ref('—')
let elapsedTimer: ReturnType<typeof setInterval> | null = null

const pipelineNotes = ref('')
const transitionBusy = ref(false)
const transitionDone = ref(false)

const stepperSteps = ['Prospect', 'Check In', 'Visit', 'Pipeline', 'Check Out', 'Done']

const currentStep = computed(() => {
  if (!review.value) return 0
  if (openVisit.value) return 2
  if (review.value.visits.length > 0) return 5
  return 0
})

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

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function message(error: unknown) { return (error as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? (error instanceof Error ? error.message : 'Unable to complete the request.') }

function coordinates() {
  return new Promise<GeolocationCoordinates>((resolve, reject) => {
    if (!navigator.geolocation) return reject(new Error('GPS is unavailable in this browser.'))
    navigator.geolocation.getCurrentPosition((position) => resolve(position.coords), () => reject(new Error('Allow location access to record this visit.')), { enableHighAccuracy: true, timeout: 12000 })
  })
}

async function load() {
  review.value = await getMyProspect(String(route.params.id))
  await nextTick(); renderMap()
}

function renderMap() {
  const prospect = review.value?.prospect
  if (!mapElement.value || prospect?.latitude == null || prospect.longitude == null) return
  map?.remove()
  map = L.map(mapElement.value, { zoomControl: true }).setView([prospect.latitude, prospect.longitude], 16)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap contributors' }).addTo(map)
  L.marker([prospect.latitude, prospect.longitude]).addTo(map).bindPopup(prospect.placeName).openPopup()
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

function acquireDeviceCoords() {
  if (!navigator.geolocation) return
  navigator.geolocation.getCurrentPosition(
    (pos) => { deviceCoords.value = { latitude: pos.coords.latitude, longitude: pos.coords.longitude }; gpsReady.value = true },
    () => { gpsReady.value = false },
    { enableHighAccuracy: true, timeout: 10000 },
  )
  navigator.geolocation.watchPosition(
    (pos) => { userCoords.value = { lat: pos.coords.latitude, lng: pos.coords.longitude } },
    () => {},
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function startElapsedTimer() {
  updateElapsed()
  elapsedTimer = setInterval(updateElapsed, 30000)
}

function updateElapsed() {
  if (!openVisit.value) { elapsed.value = '—'; return }
  const diff = Date.now() - new Date(openVisit.value.checkInAt).getTime()
  const mins = Math.floor(diff / 60000)
  const hrs = Math.floor(mins / 60)
  const remMins = mins % 60
  elapsed.value = hrs > 0 ? `${hrs}h ${remMins}m` : `${remMins}m`
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

async function checkIn() {
  error.value = ''; success.value = ''; visitBusy.value = true
  try {
    const gps = await coordinates()
    deviceCoords.value = { latitude: gps.latitude, longitude: gps.longitude }
    gpsReady.value = true
    await checkInProspect(String(route.params.id), { latitude: gps.latitude, longitude: gps.longitude, visitNotes: visitNotes.value, selfiePlaceholder: selfiePlaceholder.value })
    success.value = 'Check-in recorded successfully.'
    visitNotes.value = ''; selfiePlaceholder.value = false
    await load()
    startElapsedTimer()
  } catch (caught) { error.value = message(caught) } finally { visitBusy.value = false }
}

async function checkOut() {
  if (!openVisit.value) return
  error.value = ''; success.value = ''; visitBusy.value = true
  try {
    const gps = await coordinates()
    deviceCoords.value = { latitude: gps.latitude, longitude: gps.longitude }
    await checkOutProspect(String(route.params.id), openVisit.value.id, { latitude: gps.latitude, longitude: gps.longitude, followUpNotes: followUpNotes.value })
    success.value = 'Check-out recorded successfully.'
    followUpNotes.value = ''; transitionDone.value = false
    if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null }
    elapsed.value = '—'
    await load()
  } catch (caught) { error.value = message(caught) } finally { visitBusy.value = false }
}

async function doTransition() {
  if (!review.value || !nextPipelineStage.value) return
  error.value = ''; success.value = ''; transitionBusy.value = true
  try {
    await transitionProspect(String(route.params.id), nextPipelineStage.value, pipelineNotes.value)
    transitionDone.value = true
    success.value = `Pipeline updated to ${nextPipelineStage.value.replaceAll('_', ' ')}.`
    pipelineNotes.value = ''
    await load()
  } catch (caught) { error.value = message(caught) } finally { transitionBusy.value = false }
}

onMounted(async () => {
  try {
    await load()
    acquireDeviceCoords()
    if (openVisit.value) startElapsedTimer()
  } catch (caught) { error.value = message(caught) } finally { loading.value = false }
})

onBeforeUnmount(() => { map?.remove(); map = null; if (elapsedTimer) clearInterval(elapsedTimer) })
</script>

<template>
  <section class="detail-page">
    <RouterLink class="back-link" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> My Prospects</RouterLink>

    <!-- Messages -->
    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <!-- Loading skeleton -->
    <div v-if="loading" class="detail-skeleton">
      <div class="sk-header"><div class="sk-circle" /><div class="sk-lines"><div class="sk-line w70" /><div class="sk-line w40" /></div></div>
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /></div>
    </div>

    <!-- Not found -->
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
          <div v-if="review.prospect.industryGroup" class="dcard-row"><i class="pi pi-tag" /><span>{{ review.prospect.industryGroup }}</span></div>
          <div v-if="review.prospect.placeCategory" class="dcard-row"><i class="pi pi-bookmark" /><span>{{ review.prospect.placeCategory }}</span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span>{{ review.prospect.assignedSalesExecutive }}</span></div>
          <div v-if="review.prospect.phoneNumber" class="dcard-row"><i class="pi pi-phone" /><a :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a></div>
          <div v-if="review.prospect.websiteUrl" class="dcard-row"><i class="pi pi-globe" /><a :href="review.prospect.websiteUrl" target="_blank" rel="noopener">Website</a></div>
        </div>
      </div>

      <!-- Visit Progress Stepper -->
      <div class="dcard dcard-stepper">
        <div v-for="(step, i) in stepperSteps" :key="i" class="pv-step" :class="{ done: i < currentStep, active: i === currentStep }">
          <div class="pv-step-dot"><i v-if="i < currentStep" class="pi pi-check" /><span v-else>{{ i + 1 }}</span></div>
          <span class="pv-step-label">{{ step }}</span>
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard">
        <h2>Location</h2>
        <div v-if="review.prospect.latitude != null && review.prospect.longitude != null" ref="mapElement" class="dcard-map" role="region" aria-label="Prospect location map" />
        <Message v-else severity="warn" :closable="false">No coordinates available for this prospect.</Message>
        <div class="dcard-location-rows">
          <div class="dcard-row"><i class="pi pi-map-marker" /><span>{{ review.prospect.formattedAddress || 'No address' }}</span></div>
          <div class="dcard-row"><i class="pi pi-compass" /><span>Prospect: {{ review.prospect.latitude ?? '—' }}, {{ review.prospect.longitude ?? '—' }}</span></div>
          <div v-if="deviceCoords" class="dcard-row"><i class="pi pi-map-marker" /><span>Device: {{ deviceCoords.latitude.toFixed(6) }}, {{ deviceCoords.longitude.toFixed(6) }}</span></div>
          <div v-if="review.prospect.latitude != null && review.prospect.longitude != null && userCoords" class="dcard-row dcard-distance">
            <i class="pi pi-compass" />
            <span>{{ formatDistance(getDistanceTo(review.prospect.latitude, review.prospect.longitude, userCoords.lat, userCoords.lng)!) }} away</span>
          </div>
          <div class="dcard-row"><i class="pi" :class="gpsReady ? 'pi-check-circle' : 'pi-info-circle'" :style="{ color: gpsReady ? '#22c55e' : undefined }" /><span>{{ gpsReady ? 'GPS ready' : 'GPS will be captured on check-in' }}</span></div>
        </div>
      </div>

      <!-- Check In / Active Visit Card -->
      <div class="dcard">
        <template v-if="!openVisit">
          <h2>Check In</h2>
          <p class="dcard-hint">Record your visit with GPS and optional notes.</p>
          <label class="dcard-field"><span>Visit Notes</span><Textarea v-model="visitNotes" rows="3" fluid placeholder="What do you plan to discuss?" /></label>
          <label class="dcard-checkbox" @click.prevent="selfiePlaceholder = !selfiePlaceholder">
            <input v-model="selfiePlaceholder" type="checkbox" @click.stop />
            <i class="pi pi-camera" />
            <span>Include selfie placeholder</span>
          </label>
          <Button label="GPS Check In" icon="pi pi-map-marker" :loading="visitBusy" fluid @click="checkIn" />
        </template>
        <template v-else>
          <div class="dcard-section-header">
            <h2>Active Visit</h2>
            <Tag value="Checked In" severity="success" />
          </div>
          <div class="dcard-stats">
            <div class="dcard-stat"><span>Check-in Time</span><strong>{{ new Date(openVisit.checkInAt).toLocaleString() }}</strong></div>
            <div class="dcard-stat"><span>Duration</span><strong>{{ elapsed }}</strong></div>
          </div>
          <div v-if="openVisit.visitNotes" class="dcard-note"><span>Visit Notes</span>{{ openVisit.visitNotes }}</div>
          <label class="dcard-field"><span>Follow-up Notes</span><Textarea v-model="followUpNotes" rows="3" fluid placeholder="Notes for after the visit..." /></label>
          <Button label="GPS Check Out" icon="pi pi-check-circle" :loading="visitBusy" fluid @click="checkOut" />
        </template>
      </div>

      <!-- Pipeline Update Card -->
      <div v-if="openVisit && nextPipelineStage && !transitionDone" class="dcard dcard-highlight">
        <h2>Pipeline Update</h2>
        <p class="dcard-hint">Move this prospect to the next pipeline stage.</p>
        <div class="dcard-pipeline-flow">
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" severity="secondary" />
          <i class="pi pi-arrow-right" />
          <Tag :value="nextPipelineStage.replaceAll('_', ' ')" severity="info" />
        </div>
        <label class="dcard-field"><span>Pipeline Notes</span><Textarea v-model="pipelineNotes" rows="2" fluid placeholder="Optional notes for this transition..." /></label>
        <Button label="Update Pipeline" icon="pi pi-arrow-right" :loading="transitionBusy" outlined fluid @click="doTransition" />
      </div>
      <div v-else-if="transitionDone && openVisit" class="dcard dcard-success">
        <div class="dcard-success-row"><i class="pi pi-check-circle" /><span>Pipeline updated for this visit.</span></div>
      </div>

      <!-- Visit History Card -->
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

      <!-- Status History Card -->
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
        <button class="dbar-btn dbar-update" :disabled="!nextPipelineStage || !openVisit" @click="doTransition" :title="!openVisit ? 'Start a visit to update pipeline' : !nextPipelineStage ? 'No next stage available' : 'Update pipeline'">
          <i class="pi pi-arrow-right" /> Update
        </button>
        <RouterLink class="dbar-btn dbar-checkin" :to="`/sales/my-prospects/${review.prospect.id}`">
          <i class="pi pi-sign-in" /> Check in
        </RouterLink>
      </div>
    </template>
  </section>
</template>

<style scoped>
.detail-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: 5.5rem; }

/* ── Skeleton ────────────────────────────────────────────── */
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

/* ── Empty ───────────────────────────────────────────────── */
.detail-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.detail-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.detail-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.detail-empty span { color: var(--text-muted); font-size: 0.8rem; max-width: 260px; }
.detail-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; }

/* ── Card ────────────────────────────────────────────────── */
.dcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.dcard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

/* Summary */
.dcard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.dcard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.dcard-avatar {
  width: 52px; height: 52px; display: grid; place-items: center; border-radius: 16px;
  color: #fff; font-weight: 800; font-size: 1rem; flex-shrink: 0;
}
.dcard-avatar-prospect { background: linear-gradient(135deg, #2563eb, #1d4ed8); box-shadow: 0 3px 10px rgba(37, 99, 235, 0.25); }
.dcard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-identity .eyebrow { margin: 0; }
.dcard-identity h1 { margin: 0; font-size: 1.2rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

/* Rows */
.dcard-location-rows, .dcard-rows { display: grid; gap: 0.45rem; }
.dcard-row { display: flex; align-items: flex-start; gap: 0.55rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.dcard-row i { color: var(--text-muted); font-size: 0.72rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.1rem; }
.dcard-row a { color: var(--brand-blue); text-decoration: none; }
.dcard-row a:hover { text-decoration: underline; }
.dcard-distance { color: var(--brand-blue); font-weight: 600; }

/* Map */
.dcard-map { height: 220px; border-radius: 12px; overflow: hidden; border: 1px solid var(--border-light); }

/* Stepper */
.dcard-stepper {
  padding: 0.85rem 0.65rem; display: flex; gap: 0; overflow-x: auto;
}
.pv-step { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 0.3rem; position: relative; min-width: 48px; }
.pv-step:not(:last-child)::after { content: ''; position: absolute; top: 13px; left: calc(50% + 15px); right: calc(-50% + 15px); height: 2px; background: var(--border-light); transition: background 0.2s ease; }
.pv-step.done:not(:last-child)::after { background: #22c55e; }
.pv-step-dot { width: 28px; height: 28px; display: grid; place-items: center; border-radius: 50%; background: #f1f5f9; border: 2px solid var(--border-light); color: var(--text-muted); font-size: 0.6rem; font-weight: 700; position: relative; z-index: 1; transition: all 0.2s ease; }
.pv-step.done .pv-step-dot { background: #22c55e; border-color: #22c55e; color: #fff; }
.pv-step.active .pv-step-dot { background: var(--brand-blue); border-color: var(--brand-blue); color: #fff; box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.15); }
.pv-step-label { font-size: 0.55rem; font-weight: 600; color: var(--text-muted); text-align: center; white-space: nowrap; }
.pv-step.done .pv-step-label { color: #22c55e; }
.pv-step.active .pv-step-label { color: var(--brand-blue); }

/* Field */
.dcard-field { display: flex; flex-direction: column; gap: 0.3rem; }
.dcard-field span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.dcard-hint { margin: -0.35rem 0 0; color: var(--text-muted); font-size: 0.78rem; }
.dcard-checkbox {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.65rem 0.85rem;
  border: 1px solid var(--border-light); border-radius: 12px; background: #f8fafc;
  color: var(--text-secondary); font-size: 0.82rem; cursor: pointer;
}
.dcard-checkbox:hover { background: #f1f5f9; }
.dcard-checkbox input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--brand-blue); }
.dcard-checkbox i { color: var(--text-muted); font-size: 0.85rem; }

/* Stats */
.dcard-section-header { display: flex; align-items: center; justify-content: space-between; }
.dcard-section-header h2 { margin: 0; }
.dcard-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
.dcard-stat { padding: 0.75rem; border-radius: 12px; background: #f8fafc; }
.dcard-stat span { display: block; color: var(--text-muted); font-size: 0.62rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.15rem; }
.dcard-stat strong { font-size: 0.82rem; color: var(--text-primary); }
.dcard-note { padding: 0.75rem; border-radius: 12px; background: #f8fafc; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.5; }
.dcard-note span { display: block; margin-bottom: 0.2rem; color: var(--text-muted); font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }

/* Pipeline highlight */
.dcard-highlight { border-color: var(--brand-blue-light); background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.dcard-pipeline-flow { display: flex; align-items: center; gap: 0.65rem; padding: 0.65rem; border-radius: 12px; background: var(--surface-card); border: 1px solid var(--border-light); justify-content: center; }
.dcard-pipeline-flow i { color: var(--brand-blue); font-size: 0.85rem; }
.dcard-success { border-color: #bbf7d0; background: #f0fdf4; }
.dcard-success-row { display: flex; align-items: center; gap: 0.5rem; color: #166534; font-size: 0.82rem; }
.dcard-success-row i { font-size: 1rem; }

/* Visit history */
.dcard-visit-list { display: grid; gap: 0.65rem; }
.dcard-visit { border: 1px solid var(--border-light); border-radius: 14px; overflow: hidden; }
.dcard-visit-header { display: flex; align-items: center; justify-content: space-between; padding: 0.55rem 0.85rem; background: #f8fafc; border-bottom: 1px solid var(--border-light); }
.dcard-visit-header span { color: var(--text-muted); font-size: 0.72rem; }
.dcard-visit-body { padding: 0.65rem 0.85rem; display: grid; gap: 0.35rem; }
.dcard-visit-detail { display: flex; align-items: flex-start; gap: 0.5rem; font-size: 0.78rem; color: var(--text-secondary); }
.dcard-visit-detail i { color: var(--text-muted); font-size: 0.68rem; margin-top: 0.18rem; flex-shrink: 0; }
.dcard-visit-exec { margin-top: 0.25rem; padding-top: 0.35rem; border-top: 1px solid var(--border-light); }

/* Timeline */
.dcard-timeline { display: grid; }
.dcard-timeline-entry { display: grid; grid-template-columns: 16px 1fr; gap: 0.75rem; padding-bottom: 1rem; position: relative; }
.dcard-timeline-entry:not(:last-child)::before { content: ''; position: absolute; left: 7px; top: 18px; bottom: 0; width: 2px; background: var(--border-light); }
.dcard-timeline-dot { width: 16px; height: 16px; border-radius: 50%; background: var(--brand-blue); border: 3px solid var(--brand-blue-bg); flex-shrink: 0; }
.dcard-timeline-content { display: grid; gap: 0.1rem; }
.dcard-timeline-content strong { font-size: 0.78rem; color: var(--text-primary); text-transform: capitalize; }
.dcard-timeline-content span { color: var(--text-muted); font-size: 0.68rem; }
.dcard-timeline-content p { margin: 0.2rem 0 0; font-size: 0.78rem; color: var(--text-secondary); line-height: 1.5; }

.dcard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1.5rem 0; }

/* ── Bottom Action Bar ───────────────────────────────────── */
.detail-bottom-bar {
  position: fixed; bottom: 0; left: 0; right: 0; z-index: 40;
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
.dbar-update { background: #f0fdf4; color: #059669; border: 1px solid #a7f3d0; }
.dbar-update:hover { background: #dcfce7; }
.dbar-update:disabled { opacity: 0.45; cursor: not-allowed; }
.dbar-checkin { background: #eff6ff; color: var(--brand-blue); border: 1px solid #bfdbfe; }
.dbar-checkin:hover { background: #dbeafe; }

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 480px) {
  .detail-page { gap: 0.7rem; }
  .dcard { padding: 1rem; }
  .dcard-identity h1 { font-size: 1.05rem; }
  .dcard-map { height: 180px; }
  .pv-step { min-width: 40px; }
  .pv-step-label { font-size: 0.48rem; }
  .pv-step-dot { width: 24px; height: 24px; font-size: 0.52rem; }
  .pv-step:not(:last-child)::after { top: 11px; left: calc(50% + 13px); right: calc(-50% + 13px); }
}
</style>
