<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { checkInProspect, checkOutProspect, getMyProspect } from '../../../api/crm'
import type { ProspectReview } from '../../../types/crm'

const route = useRoute()
const review = ref<ProspectReview | null>(null)
const error = ref(''); const success = ref(''); const loading = ref(true); const visitBusy = ref(false)
const visitNotes = ref(''); const followUpNotes = ref(''); const selfiePlaceholder = ref(false)
const mapElement = ref<HTMLElement | null>(null); let map: L.Map | null = null
const openVisit = computed(() => review.value?.visits.find((visit) => !visit.checkOutAt) ?? null)

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
async function checkIn() {
  error.value = ''; success.value = ''; visitBusy.value = true
  try { const gps = await coordinates(); await checkInProspect(String(route.params.id), { latitude: gps.latitude, longitude: gps.longitude, visitNotes: visitNotes.value, selfiePlaceholder: selfiePlaceholder.value }); success.value = 'Check-in and GPS activity were recorded.'; visitNotes.value = ''; await load() } catch (caught) { error.value = message(caught) } finally { visitBusy.value = false }
}
async function checkOut() {
  if (!openVisit.value) return
  error.value = ''; success.value = ''; visitBusy.value = true
  try { const gps = await coordinates(); await checkOutProspect(String(route.params.id), openVisit.value.id, { latitude: gps.latitude, longitude: gps.longitude, followUpNotes: followUpNotes.value }); success.value = 'Check-out and follow-up were recorded.'; followUpNotes.value = ''; await load() } catch (caught) { error.value = message(caught) } finally { visitBusy.value = false }
}
onMounted(async () => { try { await load() } catch (caught) { error.value = message(caught) } finally { loading.value = false } })
onBeforeUnmount(() => map?.remove())
</script>

<template>
  <section class="mobile-page">
    <RouterLink class="back-link" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> My Prospect</RouterLink>
    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message><Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div v-if="loading" class="empty-state"><i class="pi pi-spin pi-spinner" /></div>
    <template v-else-if="review">
      <div class="mobile-title"><div><p class="eyebrow">Saved Place snapshot</p><h1>{{ review.prospect.placeName }}</h1></div><Tag :value="review.prospect.status.replaceAll('_', ' ')" /></div>
      <div class="mobile-section snapshot-detail">
        <div class="tag-row"><Tag :value="review.prospect.industryGroup" /><Tag :value="review.prospect.placeCategory || 'Uncategorized'" severity="secondary" /></div>
        <p><i class="pi pi-map-marker" /> {{ review.prospect.formattedAddress }}</p><p><i class="pi pi-compass" /> {{ review.prospect.latitude ?? '—' }}, {{ review.prospect.longitude ?? '—' }}</p>
        <p v-if="review.prospect.placeTypes.length"><i class="pi pi-tags" /> {{ review.prospect.placeTypes.join(', ') }}</p><p><i class="pi pi-user" /> {{ review.prospect.assignedSalesExecutive }}</p>
        <p v-if="review.prospect.phoneNumber"><i class="pi pi-phone" /> <a :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a></p><p v-if="review.prospect.websiteUrl"><i class="pi pi-globe" /> <a :href="review.prospect.websiteUrl" target="_blank" rel="noreferrer">Website</a></p>
        <div v-if="review.prospect.latitude != null && review.prospect.longitude != null" ref="mapElement" class="snapshot-map" aria-label="Saved prospect location map" /><Message v-else severity="warn" :closable="false">No saved coordinates are available for this prospect.</Message>
      </div>
      <div class="mobile-section visit-recorder"><h2>Simulated prospect visit</h2><p class="muted">GPS is captured as an activity record. It does not change the pipeline stage.</p>
        <template v-if="!openVisit"><label class="field"><span>Visit notes</span><Textarea v-model="visitNotes" rows="3" fluid /></label><label class="selfie-option"><input v-model="selfiePlaceholder" type="checkbox" /> Include optional selfie placeholder</label><Button label="GPS check-in" icon="pi pi-map-marker" :loading="visitBusy" @click="checkIn" /></template>
        <template v-else><Message severity="info" :closable="false">Checked in {{ new Date(openVisit.checkInAt).toLocaleString() }}</Message><label class="field"><span>Follow-up / checkout notes</span><Textarea v-model="followUpNotes" rows="3" fluid /></label><Button label="GPS check-out" icon="pi pi-check-circle" :loading="visitBusy" @click="checkOut" /></template>
      </div>
      <div class="mobile-section"><h2>Saved notes</h2><p><strong>Visit:</strong> {{ review.prospect.visitNotes || 'No notes yet.' }}</p><p><strong>Follow-up:</strong> {{ review.prospect.followUpNotes || 'No notes yet.' }}</p></div>
      <div class="mobile-section"><h2>Visit records</h2><div v-if="review.visits.length" class="timeline"><div v-for="visit in review.visits" :key="visit.id"><i class="pi pi-circle-fill" /><div><strong>{{ visit.checkOutAt ? 'Completed visit' : 'Checked in' }}</strong><span>{{ new Date(visit.checkInAt).toLocaleString() }} · {{ visit.salesExecutiveName }}</span><p>In: {{ visit.checkInLatitude }}, {{ visit.checkInLongitude }}</p><p v-if="visit.checkOutAt">Out: {{ visit.checkOutLatitude }}, {{ visit.checkOutLongitude }} · {{ new Date(visit.checkOutAt).toLocaleString() }}</p><p v-if="visit.selfieReference">Selfie: simulated placeholder recorded</p><p v-if="visit.visitNotes">{{ visit.visitNotes }}</p><p v-if="visit.followUpNotes">Follow-up: {{ visit.followUpNotes }}</p></div></div></div><p v-else class="muted">No visits recorded.</p></div>
      <div class="mobile-section"><h2>Status history</h2><div class="timeline"><div v-for="entry in review.history" :key="entry.id"><i class="pi pi-circle-fill" /><div><strong>{{ entry.fromStatus?.replaceAll('_', ' ') || 'Created' }} → {{ entry.toStatus.replaceAll('_', ' ') }}</strong><span>{{ new Date(entry.createdAt).toLocaleString() }} · {{ entry.changedByName }}</span><p v-if="entry.notes">{{ entry.notes }}</p></div></div></div></div>
    </template>
  </section>
</template>
