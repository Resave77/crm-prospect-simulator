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

const route = useRoute()
const review = ref<ProspectReview | null>(null)
const error = ref(''); const success = ref(''); const loading = ref(true); const visitBusy = ref(false)
const visitNotes = ref(''); const followUpNotes = ref(''); const selfiePlaceholder = ref(false)
const mapElement = ref<HTMLElement | null>(null); let map: L.Map | null = null
const openVisit = computed(() => review.value?.visits.find((visit) => !visit.checkOutAt) ?? null)

const deviceCoords = ref<{ latitude: number; longitude: number } | null>(null)
const gpsReady = ref(false)
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

function acquireDeviceCoords() {
  if (!navigator.geolocation) return
  navigator.geolocation.getCurrentPosition(
    (pos) => { deviceCoords.value = { latitude: pos.coords.latitude, longitude: pos.coords.longitude }; gpsReady.value = true },
    () => { gpsReady.value = false },
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
  <section class="pv-page">
    <RouterLink class="back-link" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> My Prospects</RouterLink>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div v-if="loading" class="empty-state"><i class="pi pi-spin pi-spinner" /></div>

    <template v-else-if="review">
      <!-- 1. Prospect Summary -->
      <div class="pv-summary-card">
        <div class="pv-summary-header">
          <div>
            <p class="eyebrow">Prospect</p>
            <h1>{{ review.prospect.placeName }}</h1>
          </div>
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" :severity="statusSeverity" />
        </div>
        <div class="pv-summary-details">
          <div class="pv-detail-row"><i class="pi pi-tag" /><span>{{ review.prospect.industryGroup }}</span></div>
          <div class="pv-detail-row"><i class="pi pi-user" /><span>{{ review.prospect.assignedSalesExecutive }}</span></div>
          <div class="pv-detail-row"><i class="pi pi-map-marker" /><span>{{ review.prospect.formattedAddress }}</span></div>
          <div v-if="review.prospect.phoneNumber" class="pv-detail-row"><i class="pi pi-phone" /><a :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a></div>
          <div v-if="review.prospect.websiteUrl" class="pv-detail-row"><i class="pi pi-globe" /><a :href="review.prospect.websiteUrl" target="_blank" rel="noreferrer">Website</a></div>
          <div v-if="review.prospect.googleMapsUrl" class="pv-detail-row"><i class="pi pi-external-link" /><a :href="review.prospect.googleMapsUrl" target="_blank" rel="noreferrer">Open in Google Maps</a></div>
        </div>
      </div>

      <!-- 2. Visit Progress Stepper -->
      <div class="pv-stepper">
        <div v-for="(step, i) in stepperSteps" :key="i" class="pv-step" :class="{ done: i < currentStep, active: i === currentStep }">
          <div class="pv-step-dot"><i v-if="i < currentStep" class="pi pi-check" /><span v-else>{{ i + 1 }}</span></div>
          <span class="pv-step-label">{{ step }}</span>
        </div>
      </div>

      <!-- 3. Location & Map -->
      <div class="pv-card">
        <h2>Location &amp; Map</h2>
        <div v-if="review.prospect.latitude != null && review.prospect.longitude != null" ref="mapElement" class="pv-map" role="region" aria-label="Prospect location map" />
        <Message v-else severity="warn" :closable="false">No coordinates available for this prospect.</Message>
        <div class="pv-location-info">
          <div class="pv-detail-row"><i class="pi pi-compass" /><span>Prospect: {{ review.prospect.latitude ?? '—' }}, {{ review.prospect.longitude ?? '—' }}</span></div>
          <div v-if="deviceCoords" class="pv-detail-row"><i class="pi pi-map-marker" /><span>Device: {{ deviceCoords.latitude.toFixed(6) }}, {{ deviceCoords.longitude.toFixed(6) }}</span></div>
          <div class="pv-detail-row"><i class="pi" :class="gpsReady ? 'pi-check-circle' : 'pi-info-circle'" :style="{ color: gpsReady ? '#22c55e' : undefined }" /><span>{{ gpsReady ? 'GPS ready' : 'GPS will be captured on check-in' }}</span></div>
        </div>
      </div>

      <!-- 4. Visit Action Card -->
      <div class="pv-card">
        <template v-if="!openVisit">
          <h2>Check In</h2>
          <p class="pv-hint">Record your visit with GPS and optional notes.</p>
          <label class="field"><span>Visit Notes</span><Textarea v-model="visitNotes" rows="3" fluid placeholder="What do you plan to discuss?" /></label>
          <label class="pv-selfie-option" @click.prevent="selfiePlaceholder = !selfiePlaceholder">
            <input v-model="selfiePlaceholder" type="checkbox" @click.stop />
            <i class="pi pi-camera" />
            <span>Include selfie placeholder</span>
          </label>
          <Button label="GPS Check In" icon="pi pi-map-marker" :loading="visitBusy" fluid @click="checkIn" />
        </template>
        <template v-else>
          <div class="pv-active-header">
            <h2>Active Visit</h2>
            <Tag value="Checked In" severity="success" />
          </div>
          <div class="pv-visit-stats">
            <div class="pv-stat"><span>Check-in Time</span><strong>{{ new Date(openVisit.checkInAt).toLocaleString() }}</strong></div>
            <div class="pv-stat"><span>Duration</span><strong>{{ elapsed }}</strong></div>
          </div>
          <div v-if="openVisit.visitNotes" class="pv-note-box"><span>Visit Notes</span>{{ openVisit.visitNotes }}</div>
          <label class="field"><span>Follow-up Notes</span><Textarea v-model="followUpNotes" rows="3" fluid placeholder="Notes for after the visit..." /></label>
          <Button label="GPS Check Out" icon="pi pi-check-circle" :loading="visitBusy" fluid @click="checkOut" />
        </template>
      </div>

      <!-- 5. Pipeline Section -->
      <div v-if="openVisit && nextPipelineStage && !transitionDone" class="pv-card pv-card-highlight">
        <h2>Pipeline Update</h2>
        <p class="pv-hint">Move this prospect to the next pipeline stage.</p>
        <div class="pv-pipeline-flow">
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" severity="secondary" />
          <i class="pi pi-arrow-right" />
          <Tag :value="nextPipelineStage.replaceAll('_', ' ')" severity="info" />
        </div>
        <label class="field"><span>Pipeline Notes</span><Textarea v-model="pipelineNotes" rows="2" fluid placeholder="Optional notes for this transition..." /></label>
        <Button label="Update Pipeline" icon="pi pi-arrow-right" :loading="transitionBusy" outlined fluid @click="doTransition" />
      </div>
      <div v-else-if="transitionDone && openVisit" class="pv-card pv-card-success">
        <div class="pv-success-row"><i class="pi pi-check-circle" /><span>Pipeline updated for this visit.</span></div>
      </div>

      <!-- 6. Visit History -->
      <div class="pv-card">
        <h2>Visit History</h2>
        <div v-if="review.visits.length" class="pv-visit-cards">
          <div v-for="visit in review.visits" :key="visit.id" class="pv-visit-card">
            <div class="pv-visit-card-header">
              <Tag :value="visit.checkOutAt ? 'Completed' : 'Active'" :severity="visit.checkOutAt ? 'secondary' : 'success'" />
              <span>{{ formatVisitDate(visit.checkInAt) }}</span>
            </div>
            <div class="pv-visit-card-body">
              <div class="pv-visit-detail"><i class="pi pi-sign-in" /><span>Check-in: {{ visit.checkInLatitude.toFixed(4) }}, {{ visit.checkInLongitude.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="pv-visit-detail"><i class="pi pi-sign-out" /><span>Check-out: {{ visit.checkOutLatitude?.toFixed(4) }}, {{ visit.checkOutLongitude?.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="pv-visit-detail"><i class="pi pi-clock" /><span>Duration: {{ calcDuration(visit.checkInAt, visit.checkOutAt) }}</span></div>
              <div v-if="visit.visitNotes" class="pv-visit-detail"><i class="pi pi-comment" /><span>{{ visit.visitNotes }}</span></div>
              <div v-if="visit.followUpNotes" class="pv-visit-detail"><i class="pi pi-directions" /><span>Follow-up: {{ visit.followUpNotes }}</span></div>
              <div v-if="visit.selfieReference" class="pv-visit-detail"><i class="pi pi-camera" /><span>Selfie placeholder recorded</span></div>
              <div class="pv-visit-detail pv-visit-exec"><i class="pi pi-user" /><span>{{ visit.salesExecutiveName }}</span></div>
            </div>
          </div>
        </div>
        <p v-else class="pv-empty">No visits recorded yet.</p>
      </div>

      <!-- 7. Status History -->
      <div class="pv-card">
        <h2>Status History</h2>
        <div v-if="review.history.length" class="pv-status-timeline">
          <div v-for="entry in review.history" :key="entry.id" class="pv-status-entry">
            <div class="pv-status-dot" />
            <div class="pv-status-content">
              <strong>{{ (entry.fromStatus || 'Created').replaceAll('_', ' ') }} → {{ entry.toStatus.replaceAll('_', ' ') }}</strong>
              <span>{{ new Date(entry.createdAt).toLocaleString() }} · {{ entry.changedByName }}</span>
              <p v-if="entry.notes">{{ entry.notes }}</p>
            </div>
          </div>
        </div>
        <p v-else class="pv-empty">No status changes recorded.</p>
      </div>
    </template>
  </section>
</template>

<style scoped>
.pv-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: 2rem; }
.pv-summary-card {
  position: sticky; top: 0; z-index: 10; padding: 1.25rem; border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); box-shadow: var(--shadow-sm);
}
.pv-summary-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.75rem; margin-bottom: 0.85rem; }
.pv-summary-header h1 { margin: 0; font-size: 1.3rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); }
.pv-summary-details { display: grid; gap: 0.45rem; }
.pv-detail-row { display: flex; align-items: center; gap: 0.6rem; color: var(--text-secondary); font-size: 0.82rem; }
.pv-detail-row i { color: var(--brand-blue); font-size: 0.78rem; width: 1rem; text-align: center; flex-shrink: 0; }
.pv-detail-row a { color: var(--brand-blue); text-decoration: none; }
.pv-detail-row a:hover { text-decoration: underline; }
.pv-stepper { display: flex; align-items: flex-start; gap: 0; padding: 0.85rem 0.65rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); box-shadow: var(--shadow-xs); overflow-x: auto; }
.pv-step { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 0.3rem; position: relative; min-width: 48px; }
.pv-step:not(:last-child)::after { content: ''; position: absolute; top: 13px; left: calc(50% + 15px); right: calc(-50% + 15px); height: 2px; background: var(--border-light); transition: background var(--transition-base); }
.pv-step.done:not(:last-child)::after { background: #22c55e; }
.pv-step-dot { width: 28px; height: 28px; display: grid; place-items: center; border-radius: 50%; background: var(--surface-subtle); border: 2px solid var(--border-light); color: var(--text-muted); font-size: 0.6rem; font-weight: 700; position: relative; z-index: 1; transition: all var(--transition-fast); }
.pv-step.done .pv-step-dot { background: #22c55e; border-color: #22c55e; color: #fff; }
.pv-step.active .pv-step-dot { background: var(--brand-blue); border-color: var(--brand-blue); color: #fff; box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.15); }
.pv-step-label { font-size: 0.55rem; font-weight: 600; color: var(--text-muted); text-align: center; white-space: nowrap; }
.pv-step.done .pv-step-label { color: #22c55e; }
.pv-step.active .pv-step-label { color: var(--brand-blue); }
.pv-card { padding: 1.25rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem; }
.pv-card h2 { margin: 0; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.pv-hint { margin: -0.35rem 0 0; color: var(--text-muted); font-size: 0.78rem; }
.pv-map { height: 260px; border-radius: var(--radius-md); overflow: hidden; border: 1px solid var(--border-light); }
.pv-location-info { display: grid; gap: 0.35rem; }
.pv-selfie-option {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.65rem 0.85rem;
  border: 1px solid var(--border-light); border-radius: var(--radius-md); background: var(--surface-subtle);
  color: var(--text-secondary); font-size: 0.82rem; cursor: pointer;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}
.pv-selfie-option:hover { background: var(--surface-hover); border-color: var(--border-default); }
.pv-selfie-option input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--brand-blue); }
.pv-selfie-option i { color: var(--text-muted); font-size: 0.85rem; }
.pv-active-header { display: flex; align-items: center; justify-content: space-between; }
.pv-active-header h2 { margin: 0; }
.pv-visit-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
.pv-stat { padding: 0.75rem; border-radius: var(--radius-md); background: var(--surface-subtle); }
.pv-stat span { display: block; color: var(--text-muted); font-size: 0.65rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.15rem; }
.pv-stat strong { font-size: 0.85rem; color: var(--text-primary); }
.pv-note-box { padding: 0.75rem; border-radius: var(--radius-md); background: var(--surface-subtle); color: var(--text-secondary); font-size: 0.82rem; line-height: 1.5; }
.pv-note-box span { display: block; margin-bottom: 0.2rem; color: var(--text-muted); font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.pv-card-highlight { border-color: var(--brand-blue-light); background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.pv-pipeline-flow { display: flex; align-items: center; gap: 0.65rem; padding: 0.65rem; border-radius: var(--radius-md); background: var(--surface-card); border: 1px solid var(--border-light); justify-content: center; }
.pv-pipeline-flow i { color: var(--brand-blue); font-size: 0.85rem; }
.pv-card-success { border-color: #bbf7d0; background: #f0fdf4; }
.pv-success-row { display: flex; align-items: center; gap: 0.5rem; color: #166534; font-size: 0.82rem; }
.pv-success-row i { font-size: 1rem; }
.pv-visit-cards { display: grid; gap: 0.65rem; }
.pv-visit-card { border: 1px solid var(--border-light); border-radius: var(--radius-lg); overflow: hidden; }
.pv-visit-card-header { display: flex; align-items: center; justify-content: space-between; padding: 0.55rem 0.85rem; background: var(--surface-subtle); border-bottom: 1px solid var(--border-light); }
.pv-visit-card-header span { color: var(--text-muted); font-size: 0.72rem; }
.pv-visit-card-body { padding: 0.65rem 0.85rem; display: grid; gap: 0.35rem; }
.pv-visit-detail { display: flex; align-items: flex-start; gap: 0.5rem; font-size: 0.78rem; color: var(--text-secondary); }
.pv-visit-detail i { color: var(--text-muted); font-size: 0.68rem; margin-top: 0.18rem; flex-shrink: 0; }
.pv-visit-exec { margin-top: 0.25rem; padding-top: 0.35rem; border-top: 1px solid var(--border-light); }
.pv-status-timeline { display: grid; }
.pv-status-entry { display: grid; grid-template-columns: 16px 1fr; gap: 0.75rem; padding-bottom: 1rem; position: relative; }
.pv-status-entry:not(:last-child)::before { content: ''; position: absolute; left: 7px; top: 18px; bottom: 0; width: 2px; background: var(--border-light); }
.pv-status-dot { width: 16px; height: 16px; border-radius: 50%; background: var(--brand-blue); border: 3px solid var(--brand-blue-bg); flex-shrink: 0; }
.pv-status-content { display: grid; gap: 0.1rem; }
.pv-status-content strong { font-size: 0.78rem; color: var(--text-primary); text-transform: capitalize; }
.pv-status-content span { color: var(--text-muted); font-size: 0.68rem; }
.pv-status-content p { margin: 0.2rem 0 0; font-size: 0.78rem; color: var(--text-secondary); line-height: 1.5; }
.pv-empty { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1.5rem 0; }
@media (max-width: 480px) {
  .pv-page { gap: 0.7rem; }
  .pv-card, .pv-summary-card { padding: 1rem; }
  .pv-summary-header h1 { font-size: 1.15rem; }
  .pv-stepper { padding: 0.7rem 0.35rem; }
  .pv-step { min-width: 40px; }
  .pv-step-label { font-size: 0.48rem; }
  .pv-step-dot { width: 24px; height: 24px; font-size: 0.52rem; }
  .pv-step:not(:last-child)::after { top: 11px; left: calc(50% + 13px); right: calc(-50% + 13px); }
  .pv-map { height: 220px; }
}
</style>
