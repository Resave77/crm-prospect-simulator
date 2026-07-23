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
    <div class="finder-heading">
      <div>
        <h1>Prospect Finder</h1>
        <p class="finder-subtitle">Search nearby businesses on OpenStreetMap and save qualified prospects.</p>
      </div>
    </div>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div class="finder-desktop-shell">
      <aside class="finder-left-panel">
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
            <p class="filter-section-title">Categories</p>
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
            <Button label="Use GPS" icon="pi pi-crosshairs" severity="secondary" outlined @click="useGPS" />
            <Button label="Search Area" icon="pi pi-search" :loading="loading" :disabled="!categories.length" @click="search" />
          </div>

          <p v-if="!results.length && !loading" class="filter-hint"><i class="pi pi-info-circle" /> Select categories and click Search Area to discover nearby businesses.</p>
        </div>

        <div class="finder-results-header">
          <div>
            <strong>{{ results.length }} results</strong>
            <span>within {{ (radius / 1000).toFixed(1) }} km</span>
          </div>
        </div>

        <div v-if="loading" class="empty-state finder-result-state">
          <div class="loading-pulse" />
          <strong>Searching Google Places</strong>
          <span>Scanning nearby businesses...</span>
        </div>
        <div v-else-if="!results.length" class="empty-state finder-result-state">
          <i class="pi pi-map-marker" />
          <strong>Choose filters and search</strong>
          <span>Places results will appear here.</span>
        </div>
        <div v-else class="finder-results">
          <button
            v-for="item in results"
            :key="item.googlePlaceId"
            class="result-card"
            :class="{ selected: selected?.googlePlaceId === item.googlePlaceId }"
            @click="selectResult(item)"
          >
            <span class="result-marker" :style="{ background: item.markerColor }"><i :class="item.markerIcon" /></span>
            <div class="result-info">
              <strong>{{ item.name }}</strong>
              <span class="result-meta">{{ item.category }} &middot; {{ Math.round(item.distance) }} m</span>
              <small>{{ item.address }}</small>
            </div>
            <Tag :value="item.rating ? `\u2605 ${item.rating}` : 'New'" severity="secondary" class="result-tag" />
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

        <aside v-if="selected" class="place-detail-overlay">
          <button class="detail-close" aria-label="Close selected Place details" @click="selected = null">
            <i class="pi pi-times" />
          </button>

          <div class="detail-header">
            <span class="detail-hero" :style="{ background: selected.markerColor }"><i :class="selected.markerIcon" /></span>
            <div>
              <p class="eyebrow">Selected business</p>
              <h2>{{ selected.name }}</h2>
              <Tag :value="selected.businessStatus || 'BUSINESS STATUS UNKNOWN'" severity="info" />
            </div>
          </div>

          <div class="detail-body">
            <div class="detail-list">
              <p><i class="pi pi-map-marker" />{{ selected.address }}</p>
              <p v-if="selected.phone"><i class="pi pi-phone" />{{ selected.phone }}</p>
              <a v-if="selected.website" :href="selected.website" target="_blank" rel="noreferrer"><i class="pi pi-globe" />Open website</a>
              <a v-if="selected.googleMapsUrl" :href="selected.googleMapsUrl" target="_blank" rel="noreferrer"><i class="pi pi-external-link" />Open Place listing</a>
            </div>
          </div>

          <div class="detail-actions">
            <label class="field"><span>Industry Group</span><Select v-model="industryGroup" :options="industries" fluid /></label>
            <label class="field"><span>Assign Sales Executive</span><Select v-model="salesExecutiveId" :options="sales" option-label="fullName" option-value="id" placeholder="Select Sales Executive" fluid /></label>
            <Button label="Save as Prospect" icon="pi pi-save" :loading="saving" :disabled="!salesExecutiveId" fluid @click="save" />
          </div>
        </aside>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   PROSPECT FINDER — Modern Redesign v2
   ════════════════════════════════════════════════════════════════ */

.finder-page {
  height: calc(100vh - 92px);
  min-height: 650px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

/* ── Heading ─────────────────────────────────────────────────── */
.finder-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.finder-heading h1 {
  margin: 0;
  font-size: 1.75rem;
  letter-spacing: -0.04em;
  font-weight: 800;
  background: linear-gradient(135deg, var(--text-primary) 0%, #334155 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.finder-heading .finder-subtitle {
  margin: 0.3rem 0 0;
  color: var(--text-muted);
  font-size: 0.88rem;
  line-height: 1.55;
}

.finder-heading .eyebrow { margin-bottom: 0.25rem; }

/* ── Desktop Shell ───────────────────────────────────────────── */
.finder-desktop-shell {
  min-height: 0;
  flex: 1;
  display: grid;
  grid-template-columns: 380px minmax(0, 1fr);
  overflow: hidden;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg),
              0 0 0 1px rgba(255, 255, 255, 0.8) inset;
}

/* ── Left Panel ──────────────────────────────────────────────── */
.finder-left-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr);
  background: linear-gradient(180deg, #fafbfc 0%, var(--surface-card) 100%);
  border-right: 1px solid var(--border-light);
}

.finder-filter-scroll {
  max-height: 400px;
  padding: 1rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.finder-filter-scroll::-webkit-scrollbar { width: 4px; }
.finder-filter-scroll::-webkit-scrollbar-track { background: transparent; }
.finder-filter-scroll::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 4px; }

/* ── Filter Sections ─────────────────────────────────────────── */
.filter-section {
  padding: 0.75rem 0;
  border-bottom: 1px solid #eef1f5;
}

.filter-section:last-of-type {
  border-bottom: 0;
  padding-bottom: 0.25rem;
}

.filter-section-title {
  margin: 0 0 0.55rem;
  color: var(--text-primary);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

/* Keyword */
.finder-keyword-field {
  gap: 0.4rem;
}

.finder-keyword-field > span {
  color: var(--text-muted);
  font-size: 0.7rem;
  font-weight: 700;
}

.keyword-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.keyword-icon {
  position: absolute;
  left: 0.7rem;
  color: var(--text-faint);
  font-size: 0.75rem;
  pointer-events: none;
}

.keyword-input-wrap :deep(.p-inputtext) {
  padding-left: 2rem;
  border-radius: var(--radius-md);
  border-color: var(--border-default);
  font-size: 0.8rem;
  padding-top: 0.6rem;
  padding-bottom: 0.6rem;
}

.keyword-input-wrap :deep(.p-inputtext:focus) {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* Categories */
.category-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.3rem;
}

.category-chip {
  display: flex;
  gap: 0.3rem;
  align-items: center;
  padding: 0.35rem 0.45rem;
  border-radius: var(--radius-sm);
  background: var(--surface-subtle);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.category-chip:hover {
  background: var(--surface-hover);
  border-color: var(--border-default);
}

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
  font-size: 0.62rem;
  font-weight: 550;
  transition: color var(--transition-fast);
}

.category-chip.active span {
  color: var(--brand-blue);
  font-weight: 700;
}

/* Radius */
.radius-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.radius-value {
  padding: 0.2rem 0.65rem;
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: 1rem;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: -0.01em;
}

.radius-range-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.3rem;
  color: var(--text-faint);
  font-size: 0.58rem;
}

/* Coordinates */
.coordinate-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.55rem;
}

.coordinate-grid .field {
  gap: 0.3rem;
}

.coordinate-grid .field > span {
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 700;
}

.coordinate-grid input {
  width: 100%;
  padding: 0.5rem 0.65rem;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  color: var(--text-primary);
  background: var(--surface-card);
  transition: border-color var(--transition-fast),
              box-shadow var(--transition-fast);
}

.coordinate-grid input:focus {
  outline: none;
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* Filter Actions */
.filter-actions {
  margin-top: 0.75rem;
  display: grid;
  grid-template-columns: 0.85fr 1.15fr;
  gap: 0.5rem;
}

.filter-actions :deep(.p-button) {
  padding: 0.6rem;
  font-size: 0.72rem;
  font-weight: 700;
  border-radius: var(--radius-md);
}

/* Hint */
.filter-hint {
  margin: 0.65rem 0 0;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.5rem 0.65rem;
  color: var(--text-muted);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: var(--radius-sm);
  font-size: 0.65rem;
  line-height: 1.5;
}

.filter-hint i {
  color: var(--brand-blue);
  font-size: 0.7rem;
  flex-shrink: 0;
}

/* ── Results Header ──────────────────────────────────────────── */
.finder-results-header {
  padding: 0.65rem 1rem;
  border-top: 1px solid #eef1f5;
  border-bottom: 1px solid #eef1f5;
  background: linear-gradient(135deg, var(--surface-subtle) 0%, #f0f3f7 100%);
}

.finder-results-header div {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.finder-results-header strong {
  font-size: 0.78rem;
  font-weight: 800;
  color: var(--text-primary);
}

.finder-results-header span {
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 500;
}

/* ── Results List ────────────────────────────────────────────── */
.finder-left-panel .finder-results {
  min-height: 0;
  padding: 0.6rem;
  overflow-y: auto;
  align-content: start;
}

.finder-left-panel .finder-results::-webkit-scrollbar { width: 4px; }
.finder-left-panel .finder-results::-webkit-scrollbar-track { background: transparent; }
.finder-left-panel .finder-results::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 4px; }

.finder-result-state {
  min-height: 180px;
  gap: 0.6rem;
}

.finder-result-state span {
  font-size: 0.78rem;
}

.finder-results {
  padding: 0.6rem;
  overflow-y: auto;
  display: grid;
  gap: 0.45rem;
}

.loading-pulse {
  width: 40px;
  height: 40px;
  border: 3px solid var(--brand-blue-100);
  border-top-color: var(--brand-blue);
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

@keyframes finder-spin {
  to { transform: rotate(360deg); }
}

.result-card {
  width: 100%;
  padding: 0.75rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.7rem;
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
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.06),
              0 0 0 1px rgba(37, 99, 235, 0.04);
  transform: translateY(-1px);
}

.result-card.selected {
  border-color: var(--brand-blue);
  background: linear-gradient(135deg, #f5f8ff 0%, #eef3ff 100%);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1),
              0 4px 12px rgba(37, 99, 235, 0.08);
}

.result-marker {
  width: 2.15rem;
  height: 2.15rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 50%;
  font-size: 0.72rem;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.result-info {
  min-width: 0;
  display: grid;
  gap: 0.15rem;
}

.result-info strong {
  font-size: 0.78rem;
  font-weight: 700;
  line-height: 1.35;
}

.result-meta {
  color: var(--text-muted);
  font-size: 0.62rem;
  font-weight: 500;
}

.result-info small {
  color: var(--text-faint);
  font-size: 0.6rem;
  line-height: 1.4;
}

.result-tag {
  flex-shrink: 0;
}

/* ── Map Stage ───────────────────────────────────────────────── */
.finder-map-stage {
  position: relative;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  background: linear-gradient(135deg, #e8eef5 0%, #dfe6ef 100%);
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
  left: 0.85rem;
  bottom: 1.5rem;
  max-width: 290px;
  padding: 0.6rem 0.8rem;
  display: flex;
  gap: 0.55rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(221, 229, 239, 0.95);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 16px rgba(30, 54, 84, 0.12);
  backdrop-filter: blur(12px);
}

.map-source-badge > i {
  margin-top: 0.05rem;
  color: #16a34a;
  font-size: 0.85rem;
}

.map-source-badge div { display: grid; gap: 0.15rem; }
.map-source-badge strong { font-size: 0.65rem; font-weight: 700; }
.map-source-badge span { color: #718096; font-size: 0.58rem; line-height: 1.5; }

/* ── Place Detail Overlay ────────────────────────────────────── */
.place-detail-overlay {
  position: absolute;
  z-index: 600;
  top: 0.85rem;
  right: 0.85rem;
  width: min(330px, calc(100% - 1.7rem));
  max-height: calc(100% - 1.7rem);
  overflow-y: auto;
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid rgba(229, 234, 243, 0.95);
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.12),
              0 0 0 1px rgba(255, 255, 255, 0.8) inset;
  backdrop-filter: blur(16px);
  display: flex;
  flex-direction: column;
}

.place-detail-overlay .detail-close {
  position: absolute;
  top: 0.7rem;
  right: 0.7rem;
  width: 1.85rem;
  height: 1.85rem;
  display: grid;
  place-items: center;
  color: var(--text-muted);
  background: rgba(245, 247, 250, 0.9);
  border: 1px solid var(--border-light);
  border-radius: 50%;
  cursor: pointer;
  transition: all var(--transition-fast);
  z-index: 1;
}

.place-detail-overlay .detail-close:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
  transform: scale(1.05);
}

.detail-header {
  padding: 1rem 1rem 0.85rem;
  display: flex;
  gap: 0.8rem;
  align-items: flex-start;
  background: linear-gradient(135deg, #f8faff 0%, #f5f8ff 100%);
  border-bottom: 1px solid #eef1f5;
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

.detail-header .eyebrow {
  margin-bottom: 0.15rem;
  font-size: 0.6rem;
}

.place-detail-overlay h2 {
  margin: 0 2.5rem 0.3rem 0;
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: -0.025em;
  line-height: 1.3;
  color: var(--text-primary);
}

.detail-body {
  padding: 0.85rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.detail-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-list p,
.detail-list a {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.75rem;
  line-height: 1.5;
}

.detail-list p i,
.detail-list a i {
  margin-top: 0.12rem;
  color: var(--brand-blue);
  font-size: 0.72rem;
  flex-shrink: 0;
}

.detail-list a {
  color: var(--brand-blue);
  text-decoration: none;
  font-weight: 600;
  transition: all var(--transition-fast);
}

.detail-list a:hover { opacity: 0.75; transform: translateX(2px); }

.detail-actions {
  padding: 0.85rem 1rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  border-top: 1px solid #eef1f5;
}

.place-detail-overlay .field > span {
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
}

.place-detail-overlay > .detail-actions > :deep(.p-button) {
  margin-top: 0.2rem;
  padding: 0.65rem;
  font-weight: 700;
  border-radius: var(--radius-md);
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 1100px) {
  .finder-desktop-shell { grid-template-columns: 340px minmax(0, 1fr); }
  .category-grid { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 900px) {
  .finder-desktop-shell {
    min-height: 940px;
    grid-template-columns: 1fr;
    grid-template-rows: 520px 520px;
    overflow: visible;
  }
  .finder-left-panel { border-right: 0; border-bottom: 1px solid var(--border-light); }
  .finder-filter-scroll { max-height: 360px; }
}

@media (max-width: 760px) {
  .finder-page { height: auto; min-height: 0; }
  .finder-heading { align-items: flex-start; }
  .finder-heading h1 { font-size: 1.45rem; }
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
