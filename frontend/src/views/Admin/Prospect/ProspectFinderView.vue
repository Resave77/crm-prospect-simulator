<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Slider from 'primevue/slider'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
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
const latitude = ref(0)
const longitude = ref(0)
const geoResolved = ref(false)
const results = ref<PlaceResult[]>([])
const resultSearch = ref('')
const selected = ref<PlaceResult | null>(null)
const sales = ref<SalesExecutiveOption[]>([])
const salesExecutiveId = ref('')
const industryGroup = ref('N&B / Kuliner')
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')
const detailOpen = ref(false)
const mapElement = ref<HTMLElement | null>(null)
const resultsScroll = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let searchCircle: L.Circle | null = null
const markers = new Map<string, L.Marker>()

const selectedSalesCount = computed(() => {
  const exec = sales.value.find(s => s.id === salesExecutiveId.value)
  return exec?.activeProspectCount ?? 0
})

const filteredResults = ref<PlaceResult[]>([])

watch([results, resultSearch], () => {
  const q = resultSearch.value.toLowerCase().trim()
  if (!q) { filteredResults.value = results.value; return }
  filteredResults.value = results.value.filter(r =>
    r.name.toLowerCase().includes(q) ||
    r.category.toLowerCase().includes(q) ||
    r.address.toLowerCase().includes(q)
  )
}, { immediate: true })

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
  map = L.map(mapElement.value, { zoomControl: true, preferCanvas: true }).setView([-6.2, 106.8], 12)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  }).addTo(map)
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
      .on('click', () => selectResult(item, true))
      .addTo(map)
    markers.set(item.googlePlaceId, marker)
    bounds.push(position)
  }
  drawSearchArea()
  if (bounds.length) map.fitBounds(L.latLngBounds(bounds), { padding: [48, 48], maxZoom: 16 })
  else map.setView([latitude.value, longitude.value], 14)
}

function selectResult(item: PlaceResult, focusMap = true) {
  selected.value = item
  detailOpen.value = true
  if (focusMap && map && item.latitude !== null && item.longitude !== null) {
    map.flyTo([item.latitude, item.longitude], Math.max(map.getZoom(), 16), { duration: 0.55 })
    markers.get(item.googlePlaceId)?.openTooltip()
  }
  nextTick(() => {
    const el = resultsScroll.value?.querySelector(`[data-place-id="${item.googlePlaceId}"]`)
    el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  })
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
    selected.value = null
    detailOpen.value = false
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
  if (!navigator.geolocation) { error.value = 'Geolocation is not available in this browser.'; geoResolved.value = true; return }
  navigator.geolocation.getCurrentPosition((position) => {
    latitude.value = position.coords.latitude
    longitude.value = position.coords.longitude
    geoResolved.value = true
    drawSearchArea()
    map?.flyTo([latitude.value, longitude.value], 15)
  }, () => {
    error.value = 'Location permission was denied. Please enable location access and try again.'
    geoResolved.value = true
  })
}

async function save() {
  if (!selected.value || !salesExecutiveId.value || !industryGroup.value) { error.value = 'Select a Place, Industry Group, and Sales Executive before saving.'; return }
  error.value = ''
  success.value = ''
  saving.value = true
  try {
    const item = await crmApi.saveProspect(selected.value, industryGroup.value, salesExecutiveId.value)
    success.value = `${item.placeName} saved as NEW_LEAD and assigned successfully.`
    detailOpen.value = false
  } catch (caught) {
    error.value = crmError(caught)
  } finally {
    saving.value = false
  }
}

function crmError(err: unknown) {
  const candidate = err as { response?: { data?: { error?: { message?: string } } }; message?: string }
  return candidate.response?.data?.error?.message ?? candidate.message ?? 'Prospect Finder request failed.'
}

onMounted(async () => {
  await nextTick()
  initializeMap()
  useGPS()
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
    <div class="finder-desktop-shell">
      <aside class="finder-left-panel">
        <div class="finder-panel-header">
          <div class="finder-panel-title">
            <i class="pi pi-compass" />
            <div>
              <h1>Prospect Finder</h1>
              <span>Discover &amp; save qualified prospects</span>
            </div>
          </div>
        </div>

        <div class="finder-filter-scroll">
          <div class="filter-section">
            <label class="field finder-keyword-field">
              <span>Keyword</span>
              <div class="keyword-input-wrap">
                <i class="pi pi-search keyword-icon" />
                <InputText v-model="keyword" placeholder="Cafe, hotel, pharmacy..." @keyup.enter="search" />
              </div>
            </label>
          </div>

          <div class="filter-section">
            <div class="filter-section-header">
              <span class="filter-section-title">Categories</span>
              <span class="category-count">{{ categories.length }} selected</span>
            </div>
            <div class="category-grid">
              <label v-for="option in categoryOptions" :key="option[0]" class="category-chip" :class="{ active: categories.includes(option[0]) }">
                <Checkbox v-model="categories" :input-id="option[0]" :value="option[0]" />
                <span>{{ option[1] }}</span>
              </label>
            </div>
          </div>

          <div class="filter-section">
            <div class="radius-header">
              <span class="filter-section-title">Search Radius</span>
              <span class="radius-value">{{ (radius / 1000).toFixed(1) }} km</span>
            </div>
            <Slider v-model="radius" :min="500" :max="50000" :step="500" class="finder-slider" />
            <div class="radius-range-labels">
              <span>0.5 km</span>
              <span>50 km</span>
            </div>
          </div>

          <div class="filter-section">
            <p class="filter-section-title">Coordinates</p>
            <div class="coordinate-grid">
              <label class="field"><span>Lat</span><input v-model.number="latitude" type="number" step="0.000001" /></label>
              <label class="field"><span>Lng</span><input v-model.number="longitude" type="number" step="0.000001" /></label>
            </div>
          </div>

          <div class="filter-actions">
            <Button label="GPS" icon="pi pi-crosshairs" severity="secondary" outlined @click="useGPS" />
            <Button :label="!geoResolved ? 'Detecting location...' : 'Search Area'" icon="pi pi-search" :loading="loading || !geoResolved" :disabled="!categories.length || !geoResolved" @click="search" />
          </div>
        </div>

        <div class="finder-results-header">
          <div class="results-header-row">
            <strong>{{ filteredResults.length }} result{{ filteredResults.length !== 1 ? 's' : '' }}</strong>
            <span v-if="results.length && resultSearch" class="results-filter-note">of {{ results.length }}</span>
          </div>
          <div v-if="results.length" class="result-search-wrap">
            <i class="pi pi-search" />
            <input v-model="resultSearch" placeholder="Filter results..." />
          </div>
        </div>

        <div v-if="loading" class="empty-state finder-result-state">
          <div class="loading-pulse" />
          <strong>Searching Places</strong>
          <span>Scanning nearby businesses...</span>
        </div>
        <div v-else-if="!results.length" class="empty-state finder-result-state">
          <i class="pi pi-map-marker" />
          <strong>Select categories and search</strong>
          <span>Business results will appear here.</span>
        </div>
        <div v-else-if="!filteredResults.length" class="empty-state finder-result-state">
          <i class="pi pi-filter" />
          <strong>No matching results</strong>
          <span>Try a different filter.</span>
        </div>
        <div v-else ref="resultsScroll" class="finder-results">
          <button
            v-for="item in filteredResults"
            :key="item.googlePlaceId"
            :data-place-id="item.googlePlaceId"
            class="result-card"
            :class="{ selected: selected?.googlePlaceId === item.googlePlaceId }"
            @click="selectResult(item, true)"
          >
            <span class="result-marker" :style="{ background: item.markerColor }"><i :class="item.markerIcon" /></span>
            <div class="result-info">
              <div class="result-name-row">
                <strong>{{ item.name }}</strong>
                <Tag v-if="item.rating" :value="`★ ${item.rating}`" severity="info" class="result-rating-tag" />
              </div>
              <span class="result-category">{{ item.category }}</span>
              <span class="result-address">{{ item.address }}</span>
              <div class="result-meta-row">
                <span v-if="item.distance" class="result-distance"><i class="pi pi-map-marker" /> {{ Math.round(item.distance) }} m</span>
                <Tag v-if="item.businessStatus" :value="item.businessStatus === 'OPERATIONAL' ? 'Open' : item.businessStatus" :severity="item.businessStatus === 'OPERATIONAL' ? 'success' : 'warn'" class="result-status-tag" />
              </div>
            </div>
            <i class="pi pi-chevron-right result-chevron" />
          </button>
        </div>
      </aside>

      <div class="finder-map-stage">
        <div ref="mapElement" class="leaflet-map" role="region" aria-label="OpenStreetMap with Google Places prospect markers" />

        <div class="map-source-badge">
          <i class="pi pi-shield" />
          <div>
            <strong>Private Places search</strong>
            <span>Google Places stays server-side. Map tiles use OpenStreetMap.</span>
          </div>
        </div>
      </div>
    </div>

    <Dialog v-model:visible="detailOpen" modal header="Place Details" :style="{ width: '420px' }" :closable="true" :breakpoints="{ '576px': '95vw' }">
      <div v-if="selected" class="detail-dialog">
        <div class="detail-hero-bar">
          <span class="detail-hero" :style="{ background: selected.markerColor }"><i :class="selected.markerIcon" /></span>
          <div class="detail-hero-info">
            <h2>{{ selected.name }}</h2>
            <div class="detail-hero-meta">
              <span>{{ selected.category }}</span>
              <Tag v-if="selected.rating" :value="`★ ${selected.rating}`" severity="info" />
              <Tag v-if="selected.userRatingCount" :value="`${selected.userRatingCount} reviews`" severity="secondary" />
            </div>
          </div>
        </div>

        <div class="detail-info-grid">
          <div class="detail-info-item">
            <i class="pi pi-map-marker" />
            <div>
              <span class="detail-info-label">Address</span>
              <span class="detail-info-value">{{ selected.address }}</span>
            </div>
          </div>
          <div v-if="selected.phone" class="detail-info-item">
            <i class="pi pi-phone" />
            <div>
              <span class="detail-info-label">Phone</span>
              <span class="detail-info-value">{{ selected.phone }}</span>
            </div>
          </div>
          <div v-if="selected.website" class="detail-info-item">
            <i class="pi pi-globe" />
            <div>
              <span class="detail-info-label">Website</span>
              <a :href="selected.website" target="_blank" rel="noreferrer" class="detail-info-link">Open website →</a>
            </div>
          </div>
          <div v-if="selected.googleMapsUrl" class="detail-info-item">
            <i class="pi pi-external-link" />
            <div>
              <span class="detail-info-label">Google Maps</span>
              <a :href="selected.googleMapsUrl" target="_blank" rel="noreferrer" class="detail-info-link">View listing →</a>
            </div>
          </div>
          <div class="detail-info-item">
            <i class="pi pi-info-circle" />
            <div>
              <span class="detail-info-label">Status</span>
              <Tag :value="selected.businessStatus || 'UNKNOWN'" :severity="selected.businessStatus === 'OPERATIONAL' ? 'success' : 'warn'" />
            </div>
          </div>
          <div class="detail-info-item">
            <i class="pi pi-tag" />
            <div>
              <span class="detail-info-label">Place Types</span>
              <span class="detail-info-value detail-types">{{ selected.placeTypes?.join(', ') || selected.markerCategory }}</span>
            </div>
          </div>
        </div>

        <div class="detail-assignment">
          <h3>Assignment</h3>
          <div class="detail-assignment-fields">
            <label class="field"><span>Industry Group</span><Select v-model="industryGroup" :options="industries" fluid /></label>
            <label class="field"><span>Assign Sales Executive</span><Select v-model="salesExecutiveId" :options="sales" option-label="fullName" option-value="id" placeholder="Select Sales Executive" fluid /></label>
            <Message v-if="selectedSalesCount > 0" severity="warn" :closable="false" class="assignment-warning">
              {{ sales.find(s => s.id === salesExecutiveId)?.fullName }} already has <strong>{{ selectedSalesCount }}</strong> active prospect{{ selectedSalesCount !== 1 ? 's' : '' }} assigned.
            </Message>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="detail-dialog-footer">
          <Button label="Cancel" severity="secondary" text @click="detailOpen = false" />
          <Button label="Save as Prospect" icon="pi pi-save" :loading="saving" :disabled="!salesExecutiveId || !industryGroup" @click="save" />
        </div>
      </template>
    </Dialog>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
  </section>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   PROSPECT FINDER — Workspace Layout
   ════════════════════════════════════════════════════════════════ */

.finder-page {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin: -1.5rem;
  margin-top: 0;
  min-height: 0;
}

/* ── Desktop Shell ───────────────────────────────────────────── */
.finder-desktop-shell {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 370px minmax(0, 1fr);
  overflow: hidden;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
}

/* ── Left Panel ──────────────────────────────────────────────── */
.finder-left-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto auto minmax(0, 1fr);
  background: var(--surface-card);
  border-right: 1px solid var(--border-light);
}

.finder-panel-header {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--border-light);
  background: linear-gradient(135deg, var(--brand-blue-50) 0%, #fff 100%);
}

.finder-panel-title {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.finder-panel-title > i {
  width: 2rem;
  height: 2rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: var(--brand-blue);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  flex-shrink: 0;
}

.finder-panel-title h1 {
  margin: 0;
  font-size: 1rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  line-height: 1.2;
}

.finder-panel-title span {
  display: block;
  margin-top: 0.1rem;
  color: var(--text-muted);
  font-size: 0.62rem;
  font-weight: 500;
}

.finder-filter-scroll {
  max-height: 380px;
  padding: 0.5rem 0.85rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.finder-filter-scroll::-webkit-scrollbar { width: 3px; }
.finder-filter-scroll::-webkit-scrollbar-track { background: transparent; }
.finder-filter-scroll::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 4px; }

/* ── Filter Sections ─────────────────────────────────────────── */
.filter-section {
  padding: 0.55rem 0;
  border-bottom: 1px solid #eef1f5;
}

.filter-section:last-of-type { border-bottom: 0; }

.filter-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.45rem;
}

.filter-section-title {
  margin: 0 0 0.45rem;
  color: var(--text-primary);
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.filter-section-header .filter-section-title { margin-bottom: 0; }

.category-count {
  padding: 0.1rem 0.45rem;
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: 1rem;
  font-size: 0.55rem;
  font-weight: 700;
}

/* Keyword */
.finder-keyword-field { gap: 0.3rem; }
.finder-keyword-field > span { color: var(--text-muted); font-size: 0.65rem; font-weight: 700; }

.keyword-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.keyword-icon {
  position: absolute;
  left: 0.65rem;
  color: var(--text-faint);
  font-size: 0.7rem;
  pointer-events: none;
}

.keyword-input-wrap :deep(.p-inputtext) {
  padding-left: 1.85rem;
  border-radius: var(--radius-sm);
  border-color: var(--border-default);
  font-size: 0.78rem;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
}

.keyword-input-wrap :deep(.p-inputtext:focus) {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* Categories */
.category-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.25rem;
}

.category-chip {
  display: flex;
  gap: 0.25rem;
  align-items: center;
  padding: 0.25rem 0.35rem;
  border-radius: var(--radius-sm);
  background: var(--surface-subtle);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.category-chip:hover { background: var(--surface-hover); border-color: var(--border-default); }

.category-chip.active {
  background: var(--brand-blue-50);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.12);
}

.category-chip span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 0.58rem;
  font-weight: 550;
  transition: color var(--transition-fast);
}

.category-chip.active span { color: var(--brand-blue); font-weight: 700; }

/* Radius */
.radius-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.4rem;
}

.radius-value {
  padding: 0.15rem 0.5rem;
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: 1rem;
  font-size: 0.65rem;
  font-weight: 800;
}

.radius-range-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.2rem;
  color: var(--text-faint);
  font-size: 0.52rem;
}

/* Coordinates */
.coordinate-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.45rem;
}

.coordinate-grid .field { gap: 0.2rem; }
.coordinate-grid .field > span { color: var(--text-muted); font-size: 0.58rem; font-weight: 700; }

.coordinate-grid input {
  width: 100%;
  padding: 0.4rem 0.55rem;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 0.7rem;
  color: var(--text-primary);
  background: var(--surface-card);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.coordinate-grid input:focus {
  outline: none;
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* Filter Actions */
.filter-actions {
  margin-top: 0.5rem;
  display: grid;
  grid-template-columns: 0.8fr 1.2fr;
  gap: 0.4rem;
}

.filter-actions :deep(.p-button) {
  padding: 0.5rem;
  font-size: 0.68rem;
  font-weight: 700;
  border-radius: var(--radius-sm);
}

/* ── Results Header ──────────────────────────────────────────── */
.finder-results-header {
  padding: 0.5rem 0.85rem;
  border-top: 1px solid var(--border-light);
  border-bottom: 1px solid var(--border-light);
  background: var(--surface-subtle);
}

.results-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.35rem;
}

.finder-results-header strong {
  font-size: 0.72rem;
  font-weight: 800;
  color: var(--text-primary);
}

.results-filter-note {
  color: var(--text-muted);
  font-size: 0.6rem;
  font-weight: 500;
}

.result-search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.result-search-wrap i {
  position: absolute;
  left: 0.55rem;
  color: var(--text-faint);
  font-size: 0.65rem;
  pointer-events: none;
}

.result-search-wrap input {
  width: 100%;
  padding: 0.35rem 0.55rem 0.35rem 1.5rem;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-card);
  font-size: 0.68rem;
  color: var(--text-primary);
  transition: border-color var(--transition-fast);
}

.result-search-wrap input:focus {
  outline: none;
  border-color: var(--brand-blue);
}

.result-search-wrap input::placeholder { color: var(--text-faint); }

/* ── Results List ────────────────────────────────────────────── */
.finder-result-state {
  min-height: 140px;
  gap: 0.5rem;
  padding: 1rem;
}

.finder-result-state span { font-size: 0.72rem; }
.finder-result-state i { font-size: 1.5rem; }

.loading-pulse {
  width: 32px;
  height: 32px;
  border: 3px solid var(--brand-blue-100);
  border-top-color: var(--brand-blue);
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

@keyframes finder-spin { to { transform: rotate(360deg); } }

.finder-results {
  min-height: 0;
  padding: 0.45rem;
  overflow-y: auto;
  align-content: start;
  display: grid;
  gap: 0.3rem;
}

.finder-results::-webkit-scrollbar { width: 3px; }
.finder-results::-webkit-scrollbar-track { background: transparent; }
.finder-results::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 4px; }

.result-card {
  width: 100%;
  padding: 0.6rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.55rem;
  align-items: center;
  text-align: left;
  color: var(--text-primary);
  background: var(--surface-card);
  border: 1px solid #eef1f5;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-base);
}

.result-card:hover {
  border-color: var(--brand-blue-100);
  background: #f8faff;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.06);
}

.result-card.selected {
  border-color: var(--brand-blue);
  background: linear-gradient(135deg, #f5f8ff 0%, #eef3ff 100%);
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.1), 0 2px 8px rgba(37, 99, 235, 0.08);
}

.result-marker {
  width: 1.85rem;
  height: 1.85rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 50%;
  font-size: 0.6rem;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.result-info {
  min-width: 0;
  display: grid;
  gap: 0.1rem;
}

.result-name-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.result-name-row strong {
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-rating-tag { transform: scale(0.85); transform-origin: left; }

.result-category {
  color: var(--brand-blue);
  font-size: 0.55rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.result-address {
  color: var(--text-faint);
  font-size: 0.58rem;
  line-height: 1.35;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-meta-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.1rem;
}

.result-distance {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  color: var(--text-muted);
  font-size: 0.55rem;
  font-weight: 500;
}

.result-distance i { font-size: 0.5rem; }

.result-status-tag { transform: scale(0.8); transform-origin: left; }

.result-chevron {
  color: var(--text-faint);
  font-size: 0.6rem;
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.result-card:hover .result-chevron,
.result-card.selected .result-chevron { color: var(--brand-blue); }

/* ── Map Stage ───────────────────────────────────────────────── */
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
  box-shadow: 0 4px 16px rgba(30, 54, 84, 0.18);
  border-radius: var(--radius-md) !important;
  overflow: hidden;
}

.leaflet-map :deep(.leaflet-control-zoom a) { color: #26344b; }

.leaflet-map :deep(.leaflet-control-attribution) {
  color: #617087;
  background: rgba(255, 255, 255, 0.9);
  border-radius: var(--radius-sm) 0 0 0;
  font-size: 9px;
}

/* ── Map Source Badge ────────────────────────────────────────── */
.map-source-badge {
  position: absolute;
  z-index: 500;
  left: 0.75rem;
  bottom: 0.75rem;
  max-width: 260px;
  padding: 0.5rem 0.7rem;
  display: flex;
  gap: 0.45rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(221, 229, 239, 0.95);
  border-radius: var(--radius-sm);
  box-shadow: 0 2px 10px rgba(30, 54, 84, 0.1);
  backdrop-filter: blur(12px);
}

.map-source-badge > i { margin-top: 0.02rem; color: #16a34a; font-size: 0.75rem; }
.map-source-badge div { display: grid; gap: 0.1rem; }
.map-source-badge strong { font-size: 0.58rem; font-weight: 700; }
.map-source-badge span { color: #718096; font-size: 0.52rem; line-height: 1.45; }

/* ── Detail Dialog ───────────────────────────────────────────── */
.detail-dialog {
  display: grid;
  gap: 1rem;
}

.detail-hero-bar {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
}

.detail-hero {
  width: 2.75rem;
  height: 2.75rem;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: var(--radius-md);
  font-size: 1rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.detail-hero-info h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1.3;
  color: var(--text-primary);
}

.detail-hero-meta {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 0.3rem;
  flex-wrap: wrap;
}

.detail-hero-meta > span {
  color: var(--text-muted);
  font-size: 0.7rem;
  font-weight: 500;
}

.detail-info-grid {
  display: grid;
  gap: 0.65rem;
}

.detail-info-item {
  display: flex;
  gap: 0.65rem;
  align-items: flex-start;
}

.detail-info-item > i {
  margin-top: 0.1rem;
  color: var(--brand-blue);
  font-size: 0.72rem;
  width: 1rem;
  text-align: center;
  flex-shrink: 0;
}

.detail-info-item > div {
  display: grid;
  gap: 0.1rem;
}

.detail-info-label {
  color: var(--text-muted);
  font-size: 0.58rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.detail-info-value {
  color: var(--text-primary);
  font-size: 0.78rem;
  line-height: 1.45;
}

.detail-info-value.detail-types {
  color: var(--text-secondary);
  font-size: 0.7rem;
}

.detail-info-link {
  color: var(--brand-blue);
  font-size: 0.75rem;
  font-weight: 600;
  text-decoration: none;
  transition: opacity var(--transition-fast);
}

.detail-info-link:hover { opacity: 0.75; }

.detail-assignment {
  padding-top: 0.85rem;
  border-top: 1px solid var(--border-light);
}

.detail-assignment h3 {
  margin: 0 0 0.65rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

.detail-assignment-fields {
  display: grid;
  gap: 0.65rem;
}

.detail-assignment-fields .field {
  display: grid;
  gap: 0.3rem;
}

.detail-assignment-fields .field > span {
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 700;
}

.detail-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.assignment-warning {
  font-size: 0.72rem;
  margin-top: 0.15rem;
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 1100px) {
  .finder-desktop-shell { grid-template-columns: 340px minmax(0, 1fr); }
  .category-grid { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 900px) {
  .finder-page { margin: -0.8rem; }
  .finder-desktop-shell {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(400px, 1fr);
  }
  .finder-left-panel { border-right: 0; border-bottom: 1px solid var(--border-light); }
  .finder-filter-scroll { max-height: 300px; }
}

@media (max-width: 760px) {
  .finder-page { margin: -0.8rem; }
  .finder-panel-title h1 { font-size: 0.9rem; }
  .map-source-badge { max-width: none; }
  .category-grid { grid-template-columns: repeat(3, 1fr); }
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
  transition: width 150ms cubic-bezier(0.4, 0, 0.2, 1),
              height 150ms cubic-bezier(0.4, 0, 0.2, 1),
              box-shadow 150ms cubic-bezier(0.4, 0, 0.2, 1);
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
