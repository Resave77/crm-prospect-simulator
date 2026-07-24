<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { checkInProspect, checkOutProspect, getMyProspect, transitionProspect } from '../../../api/crm'
import { nextStage } from '../../../domain/pipeline'
import type { ProspectReview } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import { getDistanceTo, formatDistance } from '../../../utils/maps'

const route = useRoute()
const router = useRouter()
const prospectId = computed(() => String(route.params.id))

const review = ref<ProspectReview | null>(null)
const error = ref('')
const success = ref('')
const loading = ref(true)
const visitBusy = ref(false)

const visitNotes = ref('')
const followUpNotes = ref('')
const selfiePlaceholder = ref(false)

const deviceCoords = ref<{ latitude: number; longitude: number } | null>(null)
const gpsReady = ref(false)
const userCoords = ref<{ lat: number; lng: number } | null>(null)
const elapsed = ref('—')
let elapsedTimer: ReturnType<typeof setInterval> | null = null

const pipelineNotes = ref('')
const transitionBusy = ref(false)
const transitionDone = ref(false)

const openVisit = computed(() => review.value?.visits.find((v) => !v.checkOutAt) ?? null)

const nextPipelineStage = computed(() => {
  if (!review.value) return null
  return nextStage(review.value.prospect.status)
})

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function message(err: unknown) {
  return (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message
    ?? (err instanceof Error ? err.message : 'Unable to complete the request.')
}

function coordinates() {
  return new Promise<GeolocationCoordinates>((resolve, reject) => {
    if (!navigator.geolocation) return reject(new Error('GPS is unavailable in this browser.'))
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve(pos.coords),
      () => reject(new Error('Allow location access to record this visit.')),
      { enableHighAccuracy: true, timeout: 12000 },
    )
  })
}

async function load() {
  review.value = await getMyProspect(prospectId.value)
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
    await checkInProspect(prospectId.value, {
      latitude: gps.latitude,
      longitude: gps.longitude,
      visitNotes: visitNotes.value,
      selfiePlaceholder: selfiePlaceholder.value,
    })
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
    await checkOutProspect(prospectId.value, openVisit.value.id, {
      latitude: gps.latitude,
      longitude: gps.longitude,
      followUpNotes: followUpNotes.value,
    })
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
    await transitionProspect(prospectId.value, nextPipelineStage.value, pipelineNotes.value)
    transitionDone.value = true
    success.value = `Pipeline updated to ${nextPipelineStage.value.replaceAll('_', ' ')}.`
    pipelineNotes.value = ''
    await load()
  } catch (caught) { error.value = message(caught) } finally { transitionBusy.value = false }
}

function goBack() {
  router.push(`/sales/my-prospects/${prospectId.value}`)
}

onMounted(async () => {
  try {
    await load()
    acquireDeviceCoords()
    if (openVisit.value) startElapsedTimer()
  } catch (caught) { error.value = message(caught) } finally { loading.value = false }
})

onBeforeUnmount(() => { if (elapsedTimer) clearInterval(elapsedTimer) })
</script>

<template>
  <section class="checkin-page">
    <button class="back-link" @click="goBack"><i class="pi pi-arrow-left" /> Back to detail</button>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <!-- Loading -->
    <div v-if="loading" class="checkin-skeleton">
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
    </div>

    <div v-else-if="!review" class="checkin-empty">
      <div class="checkin-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>Prospect not found</strong>
      <button class="checkin-empty-btn" @click="goBack"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <template v-else>
      <!-- Prospect Summary -->
      <div class="cicard cicard-summary">
        <div class="cicard-summary-top">
          <div class="cicard-avatar">{{ initials(review.prospect.placeName || 'Prospect') }}</div>
          <div class="cicard-identity">
            <p class="eyebrow">Check-in</p>
            <h1>{{ review.prospect.placeName }}</h1>
          </div>
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" severity="info" />
        </div>
      </div>

      <!-- Location -->
      <div class="cicard">
        <h2>Location</h2>
        <EntityLocationMap
          :latitude="review.prospect.latitude"
          :longitude="review.prospect.longitude"
          :label="review.prospect.placeName"
          :interactive="false"
          height="180px"
        />
        <div class="cicard-location-rows">
          <div class="cicard-row"><i class="pi pi-map-marker" /><span>{{ review.prospect.formattedAddress || 'No address' }}</span></div>
          <div class="cicard-row"><i class="pi pi-compass" /><span>Prospect: {{ review.prospect.latitude ?? '—' }}, {{ review.prospect.longitude ?? '—' }}</span></div>
          <div v-if="deviceCoords" class="cicard-row"><i class="pi pi-map-marker" /><span>Device: {{ deviceCoords.latitude.toFixed(6) }}, {{ deviceCoords.longitude.toFixed(6) }}</span></div>
          <div v-if="review.prospect.latitude != null && review.prospect.longitude != null && userCoords" class="cicard-row cicard-distance">
            <i class="pi pi-compass" />
            <span>{{ formatDistance(getDistanceTo(review.prospect.latitude, review.prospect.longitude, userCoords.lat, userCoords.lng)!) }} away</span>
          </div>
          <div class="cicard-row"><i class="pi" :class="gpsReady ? 'pi-check-circle' : 'pi-info-circle'" :style="{ color: gpsReady ? '#22c55e' : undefined }" /><span>{{ gpsReady ? 'GPS ready' : 'GPS will be captured on check-in' }}</span></div>
        </div>
      </div>

      <!-- Check In Form / Active Visit -->
      <div class="cicard">
        <template v-if="!openVisit">
          <h2>GPS Check In</h2>
          <p class="cicard-hint">Record your visit with GPS location and optional notes.</p>
          <label class="cicard-field"><span>Visit Notes</span><Textarea v-model="visitNotes" rows="3" fluid placeholder="What do you plan to discuss?" /></label>
          <label class="cicard-checkbox" @click.prevent="selfiePlaceholder = !selfiePlaceholder">
            <input v-model="selfiePlaceholder" type="checkbox" @click.stop />
            <i class="pi pi-camera" />
            <span>Include selfie placeholder</span>
          </label>
          <Button label="GPS Check In" icon="pi pi-map-marker" :loading="visitBusy" fluid @click="checkIn" />
        </template>
        <template v-else>
          <div class="cicard-section-header">
            <h2>Active Visit</h2>
            <Tag value="Checked In" severity="success" />
          </div>
          <div class="cicard-stats">
            <div class="cicard-stat"><span>Check-in Time</span><strong>{{ new Date(openVisit.checkInAt).toLocaleString() }}</strong></div>
            <div class="cicard-stat"><span>Duration</span><strong>{{ elapsed }}</strong></div>
          </div>
          <div v-if="openVisit.visitNotes" class="cicard-note"><span>Visit Notes</span>{{ openVisit.visitNotes }}</div>
          <label class="cicard-field"><span>Follow-up Notes</span><Textarea v-model="followUpNotes" rows="3" fluid placeholder="Notes for after the visit..." /></label>
          <Button label="GPS Check Out" icon="pi pi-check-circle" :loading="visitBusy" fluid @click="checkOut" />
        </template>
      </div>

      <!-- Pipeline Update -->
      <div v-if="openVisit && nextPipelineStage && !transitionDone" class="cicard cicard-highlight">
        <h2>Pipeline Update</h2>
        <p class="cicard-hint">Move this prospect to the next pipeline stage.</p>
        <div class="cicard-pipeline-flow">
          <Tag :value="review.prospect.status.replaceAll('_', ' ')" severity="secondary" />
          <i class="pi pi-arrow-right" />
          <Tag :value="nextPipelineStage.replaceAll('_', ' ')" severity="info" />
        </div>
        <label class="cicard-field"><span>Pipeline Notes</span><Textarea v-model="pipelineNotes" rows="2" fluid placeholder="Optional notes for this transition..." /></label>
        <Button label="Update Pipeline" icon="pi pi-arrow-right" :loading="transitionBusy" outlined fluid @click="doTransition" />
      </div>
      <div v-else-if="transitionDone && openVisit" class="cicard cicard-success">
        <div class="cicard-success-row"><i class="pi pi-check-circle" /><span>Pipeline updated for this visit.</span></div>
      </div>

      <!-- Visit History -->
      <div class="cicard">
        <h2>Visit History</h2>
        <div v-if="review.visits.length" class="cicard-visit-list">
          <div v-for="visit in review.visits" :key="visit.id" class="cicard-visit">
            <div class="cicard-visit-header">
              <Tag :value="visit.checkOutAt ? 'Completed' : 'Active'" :severity="visit.checkOutAt ? 'secondary' : 'success'" />
              <span>{{ formatVisitDate(visit.checkInAt) }}</span>
            </div>
            <div class="cicard-visit-body">
              <div class="cicard-visit-detail"><i class="pi pi-sign-in" /><span>Check-in: {{ visit.checkInLatitude.toFixed(4) }}, {{ visit.checkInLongitude.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="cicard-visit-detail"><i class="pi pi-sign-out" /><span>Check-out: {{ visit.checkOutLatitude?.toFixed(4) }}, {{ visit.checkOutLongitude?.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="cicard-visit-detail"><i class="pi pi-clock" /><span>Duration: {{ calcDuration(visit.checkInAt, visit.checkOutAt) }}</span></div>
              <div v-if="visit.visitNotes" class="cicard-visit-detail"><i class="pi pi-comment" /><span>{{ visit.visitNotes }}</span></div>
              <div v-if="visit.followUpNotes" class="cicard-visit-detail"><i class="pi pi-directions" /><span>Follow-up: {{ visit.followUpNotes }}</span></div>
              <div class="cicard-visit-detail cicard-visit-exec"><i class="pi pi-user" /><span>{{ visit.salesExecutiveName }}</span></div>
            </div>
          </div>
        </div>
        <p v-else class="cicard-empty-text">No visits recorded yet.</p>
      </div>
    </template>
  </section>
</template>

<style scoped>
.checkin-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: 2rem; }

.back-link {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0;
  border: 0; background: transparent; color: var(--brand-blue, #2563eb);
  font-size: 0.8rem; font-weight: 600; cursor: pointer; text-decoration: none;
}
.back-link:hover { text-decoration: underline; }

.checkin-skeleton { display: grid; gap: 0.85rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w50 { width: 50%; }
.sk-line.w60 { width: 60%; }
.sk-line.w70 { width: 70%; }
.sk-line.w80 { width: 80%; }
.sk-map { height: 180px; border-radius: 12px; background: #e2e8f0; }

.checkin-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.checkin-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.checkin-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.checkin-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; border: 0; cursor: pointer; }

.cicard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.cicard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

.cicard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.cicard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.cicard-avatar {
  width: 52px; height: 52px; display: grid; place-items: center; border-radius: 16px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff; font-weight: 800;
  font-size: 1rem; flex-shrink: 0; box-shadow: 0 3px 10px rgba(37, 99, 235, 0.25);
}
.cicard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.cicard-identity .eyebrow { margin: 0; }
.cicard-identity h1 { margin: 0; font-size: 1.15rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

.cicard-location-rows { display: grid; gap: 0.45rem; }
.cicard-row { display: flex; align-items: flex-start; gap: 0.55rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.cicard-row i { color: var(--text-muted); font-size: 0.72rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.1rem; }
.cicard-distance { color: var(--brand-blue); font-weight: 600; }

.cicard-field { display: flex; flex-direction: column; gap: 0.3rem; }
.cicard-field span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.cicard-hint { margin: -0.35rem 0 0; color: var(--text-muted); font-size: 0.78rem; }
.cicard-checkbox {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.65rem 0.85rem;
  border: 1px solid var(--border-light); border-radius: 12px; background: #f8fafc;
  color: var(--text-secondary); font-size: 0.82rem; cursor: pointer;
}
.cicard-checkbox:hover { background: #f1f5f9; }
.cicard-checkbox input[type="checkbox"] { width: 16px; height: 16px; accent-color: var(--brand-blue); }
.cicard-checkbox i { color: var(--text-muted); font-size: 0.85rem; }

.cicard-section-header { display: flex; align-items: center; justify-content: space-between; }
.cicard-section-header h2 { margin: 0; }
.cicard-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
.cicard-stat { padding: 0.75rem; border-radius: 12px; background: #f8fafc; }
.cicard-stat span { display: block; color: var(--text-muted); font-size: 0.62rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.15rem; }
.cicard-stat strong { font-size: 0.82rem; color: var(--text-primary); }
.cicard-note { padding: 0.75rem; border-radius: 12px; background: #f8fafc; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.5; }
.cicard-note span { display: block; margin-bottom: 0.2rem; color: var(--text-muted); font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }

.cicard-highlight { border-color: var(--brand-blue-light); background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.cicard-pipeline-flow { display: flex; align-items: center; gap: 0.65rem; padding: 0.65rem; border-radius: 12px; background: var(--surface-card); border: 1px solid var(--border-light); justify-content: center; }
.cicard-pipeline-flow i { color: var(--brand-blue); font-size: 0.85rem; }
.cicard-success { border-color: #bbf7d0; background: #f0fdf4; }
.cicard-success-row { display: flex; align-items: center; gap: 0.5rem; color: #166534; font-size: 0.82rem; }
.cicard-success-row i { font-size: 1rem; }

.cicard-visit-list { display: grid; gap: 0.65rem; }
.cicard-visit { border: 1px solid var(--border-light); border-radius: 14px; overflow: hidden; }
.cicard-visit-header { display: flex; align-items: center; justify-content: space-between; padding: 0.55rem 0.85rem; background: #f8fafc; border-bottom: 1px solid var(--border-light); }
.cicard-visit-header span { color: var(--text-muted); font-size: 0.72rem; }
.cicard-visit-body { padding: 0.65rem 0.85rem; display: grid; gap: 0.35rem; }
.cicard-visit-detail { display: flex; align-items: flex-start; gap: 0.5rem; font-size: 0.78rem; color: var(--text-secondary); }
.cicard-visit-detail i { color: var(--text-muted); font-size: 0.68rem; margin-top: 0.18rem; flex-shrink: 0; }
.cicard-visit-exec { margin-top: 0.25rem; padding-top: 0.35rem; border-top: 1px solid var(--border-light); }

.cicard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1.5rem 0; }

@media (max-width: 480px) {
  .checkin-page { gap: 0.7rem; }
  .cicard { padding: 1rem; }
  .cicard-identity h1 { font-size: 1.05rem; }
}
</style>
