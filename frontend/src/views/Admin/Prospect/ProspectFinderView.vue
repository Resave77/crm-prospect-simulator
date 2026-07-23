<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Slider from 'primevue/slider'
import Tag from 'primevue/tag'
import * as crmApi from '../../../api/crm'
import type { PlaceResult, SalesExecutiveOption } from '../../../types/crm'

const categoryOptions = [
  ['food_drink', 'Food & Drink'], ['business', 'Business'], ['culture', 'Culture'], ['education', 'Education'],
  ['entertainment', 'Entertainment'], ['health', 'Health'], ['shopping', 'Shopping'], ['lodging', 'Lodging'], ['services', 'Services'],
] as const
const industries = ['N&B / Kuliner', 'Retail', 'Hospitality', 'Health & Beauty', 'Services', 'Other']
const keyword = ref('')
const categories = ref<string[]>(['food_drink', 'business', 'culture'])
const radius = ref(3000)
const latitude = ref(-6.229561)
const longitude = ref(106.848651)
const results = ref<PlaceResult[]>([])
const selected = ref<PlaceResult | null>(null)
const sales = ref<SalesExecutiveOption[]>([])
const salesExecutiveId = ref('')
const industryGroup = ref('N&B / Kuliner')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')
const mapElement = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let searchCircle: L.Circle | null = null
const markers = new Map<string, L.Marker>()

function markerIcon(item: PlaceResult, active = false) {
  const safeIcon = /^pi pi-[a-z-]+$/.test(item.markerIcon) ? item.markerIcon : 'pi pi-map-marker'
  const safeColor = /^#[0-9a-f]{6}$/i.test(item.markerColor) ? item.markerColor : '#2563eb'
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker${active ? ' is-selected' : ''}" style="--marker-color:${safeColor}"><i class="${safeIcon}"></i></span>`,
    iconSize: active ? [44, 50] : [36, 42],
    iconAnchor: active ? [22, 48] : [18, 40],
    popupAnchor: [0, -42],
  })
}

function initializeMap() {
  if (!mapElement.value || map) return
  map = L.map(mapElement.value, { zoomControl: true, preferCanvas: true }).setView([latitude.value, longitude.value], 14)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  }).addTo(map)
  drawSearchArea()
}

function drawSearchArea() {
  if (!map) return
  searchCircle?.remove()
  searchCircle = L.circle([latitude.value, longitude.value], {
    radius: radius.value,
    color: '#2563eb',
    weight: 1,
    fillColor: '#3b82f6',
    fillOpacity: 0.06,
  }).addTo(map)
}

function renderMarkers() {
  if (!map) return
  markers.forEach((marker) => marker.remove())
  markers.clear()
  const bounds: L.LatLngExpression[] = []
  for (const item of results.value) {
    if (item.latitude === null || item.longitude === null) continue
    const position: L.LatLngExpression = [item.latitude, item.longitude]
    const marker = L.marker(position, { icon: markerIcon(item, selected.value?.googlePlaceId === item.googlePlaceId), keyboard: true, title: item.name })
      .bindTooltip(item.name, { direction: 'top', offset: [0, -34] })
      .on('click', () => selectResult(item, false))
      .addTo(map)
    markers.set(item.googlePlaceId, marker)
    bounds.push(position)
  }
  drawSearchArea()
  if (bounds.length) map.fitBounds(L.latLngBounds(bounds), { padding: [48, 48], maxZoom: 16 })
  else map.setView([latitude.value, longitude.value], 14)
}

function selectResult(item: PlaceResult, focus = true) {
  selected.value = item
  if (focus && map && item.latitude !== null && item.longitude !== null) {
    map.flyTo([item.latitude, item.longitude], Math.max(map.getZoom(), 16), { duration: 0.55 })
    markers.get(item.googlePlaceId)?.openTooltip()
  }
}

watch(selected, (current, previous) => {
  if (previous) markers.get(previous.googlePlaceId)?.setIcon(markerIcon(previous, false))
  if (current) markers.get(current.googlePlaceId)?.setIcon(markerIcon(current, true))
})

watch(radius, drawSearchArea)

async function search() {
  error.value = ''
  success.value = ''
  loading.value = true
  try {
    results.value = await crmApi.searchPlaces({ keyword: keyword.value, categories: categories.value.join(','), radius: radius.value, latitude: latitude.value, longitude: longitude.value })
    selected.value = results.value[0] ?? null
    await nextTick()
    renderMarkers()
  } catch (caught) {
    error.value = crmError(caught)
  } finally {
    loading.value = false
  }
}

function useGPS() {
  error.value = ''
  if (!navigator.geolocation) { error.value = 'Geolocation is not available in this browser.'; return }
  navigator.geolocation.getCurrentPosition((position) => {
    latitude.value = position.coords.latitude
    longitude.value = position.coords.longitude
    drawSearchArea()
    map?.flyTo([latitude.value, longitude.value], 15)
  }, () => { error.value = 'Location permission was denied or unavailable.' })
}

async function save() {
  if (!selected.value || !salesExecutiveId.value || !industryGroup.value) { error.value = 'Select a Place, Industry Group, and Sales Executive before saving.'; return }
  error.value = ''
  success.value = ''
  saving.value = true
  try {
    const item = await crmApi.saveProspect(selected.value, industryGroup.value, salesExecutiveId.value)
    success.value = `${item.placeName} saved as NEW_LEAD and assigned successfully.`
  } catch (caught) {
    error.value = crmError(caught)
  } finally {
    saving.value = false
  }
}

function crmError(error: unknown) {
  const candidate = error as { response?: { data?: { error?: { message?: string } } }; message?: string }
  return candidate.response?.data?.error?.message ?? candidate.message ?? 'Prospect Finder request failed.'
}

onMounted(async () => {
  await nextTick()
  initializeMap()
  try {
    sales.value = await crmApi.getSalesExecutives()
    salesExecutiveId.value = sales.value[0]?.id ?? ''
  } catch (caught) {
    error.value = crmError(caught)
  }
})

onBeforeUnmount(() => { map?.remove(); map = null; markers.clear() })
</script>

<template>
  <section class="finder-page">
    <div class="finder-heading"><div><p class="eyebrow">Google Places discovery</p><h1>Prospect Finder</h1><p>Google Places results on a keyless OpenStreetMap canvas.</p></div><Tag value="Leaflet + OpenStreetMap" severity="success" /></div>
    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div class="finder-desktop-shell">
      <aside class="finder-left-panel">
        <div class="finder-filter-scroll">
          <label class="field"><span>Keyword</span><InputText v-model="keyword" placeholder="Cafe, hotel, pharmacy..." @keyup.enter="search" /></label>
          <fieldset><legend>Categories</legend><label v-for="option in categoryOptions" :key="option[0]"><Checkbox v-model="categories" :input-id="option[0]" :value="option[0]" /><span>{{ option[1] }}</span></label></fieldset>
          <div class="radius-row"><strong>Search radius</strong><span>{{ (radius / 1000).toFixed(1) }} km</span></div><Slider v-model="radius" :min="500" :max="50000" :step="500" />
          <div class="coordinate-grid"><label class="field"><span>Latitude</span><input v-model.number="latitude" type="number" step="0.000001" /></label><label class="field"><span>Longitude</span><input v-model.number="longitude" type="number" step="0.000001" /></label></div>
          <div class="finder-filter-actions"><Button label="Use GPS" icon="pi pi-crosshairs" severity="secondary" outlined @click="useGPS" /><Button label="Search area" icon="pi pi-search" :loading="loading" :disabled="!categories.length" @click="search" /></div>
        </div>
        <div class="finder-results-header"><div><strong>{{ results.length }} results</strong><span>within {{ (radius / 1000).toFixed(1) }} km</span></div></div>
        <div v-if="loading" class="empty-state finder-result-state"><i class="pi pi-spin pi-spinner" /><strong>Searching Google Places</strong></div>
        <div v-else-if="!results.length" class="empty-state finder-result-state"><i class="pi pi-map-marker" /><strong>Choose filters and search</strong><span>Places results will appear here.</span></div>
        <div v-else class="finder-results">
          <button v-for="item in results" :key="item.googlePlaceId" class="result-card" :class="{ selected: selected?.googlePlaceId === item.googlePlaceId }" @click="selectResult(item)"><span class="result-marker" :style="{ background: item.markerColor }"><i :class="item.markerIcon" /></span><div><strong>{{ item.name }}</strong><span>{{ item.category }} · {{ Math.round(item.distance) }} m</span><small>{{ item.address }}</small></div><Tag :value="item.rating ? `★ ${item.rating}` : 'New'" severity="secondary" /></button>
        </div>
      </aside>
      <div class="finder-map-stage">
        <div ref="mapElement" class="leaflet-map" role="region" aria-label="OpenStreetMap with Google Places prospect markers" />
        <div class="map-source-badge"><i class="pi pi-shield" /><div><strong>Private Places search</strong><span>Google Places stays server-side. Map tiles use OpenStreetMap.</span></div></div>
        <aside v-if="selected" class="place-detail-overlay">
          <button class="detail-close" aria-label="Close selected Place details" @click="selected = null"><i class="pi pi-times" /></button>
          <span class="detail-hero" :style="{ background: selected.markerColor }"><i :class="selected.markerIcon" /></span>
          <p class="eyebrow">Selected business</p><h2>{{ selected.name }}</h2>
          <Tag :value="selected.businessStatus || 'BUSINESS STATUS UNKNOWN'" severity="info" />
          <div class="detail-list"><p><i class="pi pi-map-marker" />{{ selected.address }}</p><p v-if="selected.phone"><i class="pi pi-phone" />{{ selected.phone }}</p><a v-if="selected.website" :href="selected.website" target="_blank" rel="noreferrer"><i class="pi pi-globe" />Open website</a><a v-if="selected.googleMapsUrl" :href="selected.googleMapsUrl" target="_blank" rel="noreferrer"><i class="pi pi-external-link" />Open Place listing</a></div>
          <label class="field"><span>Industry Group</span><Select v-model="industryGroup" :options="industries" fluid /></label>
          <label class="field"><span>Assign Sales Executive</span><Select v-model="salesExecutiveId" :options="sales" option-label="fullName" option-value="id" placeholder="Select Sales Executive" fluid /></label>
          <Button label="Save as Prospect" icon="pi pi-save" :loading="saving" :disabled="!salesExecutiveId" fluid @click="save" />
        </aside>
      </div>
    </div>
  </section>
</template>

<style scoped>
.finder-page {
  height: calc(100vh - 92px);
  min-height: 650px;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.finder-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.finder-heading h1 {
  margin: 0;
  font-size: 1.5rem;
  letter-spacing: -0.035em;
  font-weight: 800;
}

.finder-heading p:not(.eyebrow) {
  margin: 0.2rem 0 0;
  color: var(--text-muted);
  font-size: 0.68rem;
}

.finder-heading .eyebrow { margin-bottom: 0.2rem; }

.finder-desktop-shell {
  min-height: 0;
  flex: 1;
  display: grid;
  grid-template-columns: 350px minmax(0, 1fr);
  overflow: hidden;
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
}

.finder-left-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  background: var(--surface-card);
  border-right: 1px solid var(--border-light);
}

.finder-filter-scroll {
  max-height: 325px;
  padding: 1rem;
  overflow-y: auto;
}

.finder-filter-scroll fieldset {
  margin: 0.7rem 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.35rem;
  border: 0;
}

.finder-filter-scroll legend {
  grid-column: 1 / -1;
  margin-bottom: 0.3rem;
  color: var(--text-primary);
  font-size: 0.68rem;
  font-weight: 800;
}

.finder-filter-scroll fieldset label {
  min-width: 0;
  display: flex;
  gap: 0.25rem;
  align-items: center;
  color: #5f6d83;
  font-size: 0.55rem;
}

.finder-filter-scroll fieldset label span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.finder-filter-actions {
  margin-top: 0.7rem;
  display: grid;
  grid-template-columns: 0.85fr 1.15fr;
  gap: 0.45rem;
}

.finder-filter-actions :deep(.p-button) {
  padding: 0.5rem;
  font-size: 0.62rem;
}

.finder-left-panel .finder-results-header {
  background: var(--surface-subtle);
}

.finder-results-header {
  padding: 0.6rem 0.85rem;
  border-top: 1px solid var(--border-light);
  border-bottom: 1px solid var(--border-light);
}

.finder-results-header div {
  display: flex;
  justify-content: space-between;
  font-size: 0.68rem;
}

.finder-results-header span { color: var(--text-muted); }

.finder-left-panel .finder-results {
  min-height: 0;
  padding: 0.6rem;
  overflow-y: auto;
  align-content: start;
}

.finder-result-state { min-height: 150px; }

.finder-results {
  padding: 0.6rem;
  overflow-y: auto;
  display: grid;
  gap: 0.4rem;
}

.result-card {
  width: 100%;
  padding: 0.6rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.6rem;
  align-items: start;
  text-align: left;
  color: var(--text-primary);
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color var(--transition-fast),
              background var(--transition-fast),
              box-shadow var(--transition-fast);
}

.result-card:hover {
  border-color: var(--border-default);
  background: var(--surface-subtle);
}

.result-card.selected {
  border-color: var(--brand-blue);
  background: #f5f8ff;
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.08);
}

.result-marker {
  width: 2rem;
  height: 2rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 50%;
  font-size: 0.7rem;
}

.result-card div { display: grid; }
.result-card strong { font-size: 0.65rem; }
.result-card span { color: var(--text-muted); font-size: 0.52rem; }
.result-card small { color: #7f8c9b; font-size: 0.5rem; }

.finder-map-stage {
  position: relative;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  background: #e8eef5;
}

.leaflet-map {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
}

.leaflet-map :deep(.leaflet-control-zoom) {
  border: 0;
  box-shadow: 0 4px 12px rgba(30, 54, 84, 0.15);
  border-radius: var(--radius-sm) !important;
}

.leaflet-map :deep(.leaflet-control-zoom a) { color: #26344b; }

.leaflet-map :deep(.leaflet-control-attribution) {
  color: #617087;
  background: rgba(255, 255, 255, 0.88);
  font-size: 9px;
}

.map-source-badge {
  position: absolute;
  z-index: 500;
  left: 0.75rem;
  bottom: 1.5rem;
  max-width: 280px;
  padding: 0.6rem 0.75rem;
  display: flex;
  gap: 0.5rem;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(221, 229, 239, 0.9);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(8px);
}

.map-source-badge > i { color: #16a34a; }
.map-source-badge div { display: grid; gap: 0.1rem; }
.map-source-badge strong { font-size: 0.6rem; }
.map-source-badge span { color: #718096; font-size: 0.5rem; line-height: 1.4; }

.place-detail-overlay {
  position: absolute;
  z-index: 600;
  top: 0.75rem;
  right: 0.75rem;
  width: min(310px, calc(100% - 1.5rem));
  max-height: calc(100% - 1.5rem);
  padding: 0.9rem;
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  backdrop-filter: blur(8px);
}

.place-detail-overlay .detail-close {
  position: absolute;
  top: 0.6rem;
  right: 0.6rem;
  width: 1.8rem;
  height: 1.8rem;
  display: grid;
  place-items: center;
  color: #65738a;
  background: var(--surface-subtle);
  border: 0;
  border-radius: 50%;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.place-detail-overlay .detail-close:hover { background: var(--surface-hover); }

.detail-hero {
  width: 2.75rem;
  height: 2.75rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: var(--radius-md);
  font-size: 1rem;
}

.place-detail-overlay h2 {
  margin: 0.2rem 2rem 0.4rem 0;
  font-size: 1.1rem;
}

.place-detail-overlay .detail-list { margin: 0.7rem 0; }
.place-detail-overlay .field { margin-top: 0.7rem; }
.place-detail-overlay > :deep(.p-button) { margin-top: 0.8rem; }

.radius-row {
  margin: 0.8rem 0 0.6rem;
  display: flex;
  justify-content: space-between;
  font-size: 0.68rem;
}

.coordinate-grid {
  margin: 0.8rem 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.coordinate-grid input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
}

@media (max-width: 1100px) {
  .finder-desktop-shell { grid-template-columns: 310px minmax(0, 1fr); }
  .finder-filter-scroll fieldset { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 900px) {
  .finder-desktop-shell { min-height: 900px; grid-template-columns: 1fr; grid-template-rows: 480px 520px; overflow: visible; }
  .finder-left-panel { border-right: 0; border-bottom: 1px solid var(--border-light); }
  .finder-filter-scroll { max-height: 285px; }
}

@media (max-width: 760px) {
  .finder-page { height: auto; min-height: 0; }
  .finder-heading { align-items: flex-start; }
  .map-source-badge { right: 0.65rem; max-width: none; }
}
</style>

<style>
.finder-leaflet-icon-host { border: 0; background: transparent; }

.finder-leaflet-marker {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--marker-color);
  border: 3px solid #fff;
  border-radius: 50% 50% 50% 0;
  box-shadow: 0 4px 14px rgba(22, 41, 67, 0.3);
  transform: rotate(-45deg);
  transition: width var(--transition-fast),
              height var(--transition-fast),
              box-shadow var(--transition-fast);
}

.finder-leaflet-marker i { transform: rotate(45deg); font-size: 0.8rem; }

.finder-leaflet-marker.is-selected {
  width: 44px;
  height: 44px;
  box-shadow: 0 0 0 5px rgba(37, 99, 235, 0.18),
              0 6px 18px rgba(22, 41, 67, 0.35);
}

.finder-leaflet-marker.is-selected i { font-size: 1rem; }
</style>
