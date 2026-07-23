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
