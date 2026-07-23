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
