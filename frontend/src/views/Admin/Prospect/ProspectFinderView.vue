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
import { useToast } from 'primevue/usetoast'
import * as crmApi from '../../../api/crm'
import type { CustomerMarker, PlaceDetails, PlaceResult, SalesExecutiveOption } from '../../../types/crm'

const toast = useToast()

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
const customerMarkers = ref<CustomerMarker[]>([])
const customersLoading = ref(false)
const placeDetails = ref<PlaceDetails | null>(null)
const placeDetailsLoading = ref(false)
const placeDetailsError = ref('')
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
let customerLayer: L.LayerGroup | null = null
const markers = new Map<string, L.Marker>()
const customerMarkerMap = new Map<string, L.Marker>()

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
  if (item.isCustomer) {
    return customerMarkerIcon(active)
  }
  const safeIcon = /^pi pi-[a-z-]+$/.test(item.markerIcon) ? item.markerIcon : 'pi pi-map-marker'
  const safeColor = /^#[0-9a-f]{6}$/i.test(item.markerColor) ? item.markerColor : '#d14350'
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker${active ? ' is-selected' : ''}" style="--marker-color:${safeColor}"><i class="${safeIcon}"></i></span>`,
    iconSize: active ? [44, 50] : [36, 42],
    iconAnchor: active ? [22, 48] : [18, 40],
    popupAnchor: [0, -42],
  })
}

function customerMarkerIcon(active = false) {
  return L.divIcon({
    className: 'finder-leaflet-icon-host',
    html: `<span class="finder-leaflet-marker is-customer${active ? ' is-selected' : ''}" title="Existing customer"><b>Y</b></span>`,
    iconSize: active ? [46, 52] : [38, 44],
    iconAnchor: active ? [23, 50] : [19, 42],
    popupAnchor: [0, -44],
  })
}

function customerToResult(c: CustomerMarker): PlaceResult {
  return {
    customerId: c.customerId,
    googlePlaceId: c.googlePlaceId,
    name: c.name,
    address: c.address,
    category: 'Existing Customer',
    distance: 0,
    rating: 0,
    userRatingCount: 0,
    businessStatus: 'OPERATIONAL',
    latitude: c.latitude,
    longitude: c.longitude,
    phone: '',
    website: '',
    googleMapsUrl: '',
    markerCategory: 'customer',
    markerColor: '#16a34a',
    markerIcon: '',
    placeTypes: [],
    isCustomer: true,
  }
}

function initializeMap() {
  if (!mapElement.value || map) return
  map = L.map(mapElement.value, { zoomControl: true, preferCanvas: true }).setView([-6.2, 106.8], 12)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
  }).addTo(map)
  customerLayer = L.layerGroup().addTo(map)
}

function renderCustomerMarkers() {
  if (!map || !customerLayer) return
  customerLayer.clearLayers()
  customerMarkerMap.clear()
  const bounds: L.LatLngExpression[] = []
  for (const item of customerMarkers.value) {
    if (item.latitude === null || item.longitude === null) continue
    const position: L.LatLngExpression = [item.latitude, item.longitude]
    const marker = L.marker(position, { icon: customerMarkerIcon(selected.value?.customerId === item.customerId), keyboard: true, title: item.name })
      .bindTooltip(item.name, { direction: 'top', offset: [0, -42] })
      .on('click', () => selectCustomer(item))
      .addTo(customerLayer)
    customerMarkerMap.set(item.customerId, marker)
    bounds.push(position)
  }
  if (!results.value.length && bounds.length) {
    map.fitBounds(L.latLngBounds(bounds), { padding: [48, 48], maxZoom: 14 })
  }
}

async function loadCustomerMarkers() {
  customersLoading.value = true
  try {
    customerMarkers.value = await crmApi.getCustomerMarkers()
  } catch (caught) {
    toast.add({ severity: 'warn', summary: 'Customers unavailable', detail: crmError(caught), life: 5000 })
  } finally {
    customersLoading.value = false
    renderCustomerMarkers()
  }
}

function drawSearchArea() {
  if (!map) return
  searchCircle?.remove()
  searchCircle = L.circle([latitude.value, longitude.value], {
    radius: radius.value,
    color: '#d14350',
    weight: 1,
    fillColor: '#df5a66',
    fillOpacity: 0.06,
  }).addTo(map)
}

function renderMarkers() {
  if (!map) return
  markers.forEach((marker) => marker.remove())
  markers.clear()
  const bounds: L.LatLngExpression[] = []
  for (const item of results.value) {
    if (item.isCustomer) continue
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

function closeResults() { results.value = []; filteredResults.value = []; resultSearch.value = ''; detailOpen.value = false; placeDetails.value = null }

async function selectResult(item: PlaceResult, focusMap = true) {
  selected.value = item
  placeDetails.value = null
  placeDetailsError.value = ''
  detailOpen.value = true
  if (item.googlePlaceId) {
    placeDetailsLoading.value = true
    try {
      placeDetails.value = await crmApi.getPlaceDetails(item.googlePlaceId)
    } catch (caught) {
      placeDetailsError.value = crmError(caught)
    } finally {
      placeDetailsLoading.value = false
    }
  }
  if (focusMap && map && item.latitude !== null && item.longitude !== null) {
    map.flyTo([item.latitude, item.longitude], Math.max(map.getZoom(), 16), { duration: 0.55 })
    const cid = item.customerId
    if (cid) customerMarkerMap.get(cid)?.openTooltip()
    else markers.get(item.googlePlaceId)?.openTooltip()
  }
  nextTick(() => {
    if (!item.googlePlaceId) return
    const el = resultsScroll.value?.querySelector(`[data-place-id="${item.googlePlaceId}"]`)
    el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  })
}

function selectCustomer(item: CustomerMarker) {
  selectResult(customerToResult(item), true)
}

watch(selected, (current, previous) => {
  if (previous) {
    if (previous.customerId) customerMarkerMap.get(previous.customerId)?.setIcon(customerMarkerIcon(false))
    else markers.get(previous.googlePlaceId)?.setIcon(markerIcon(previous, false))
  }
  if (current) {
    if (current.customerId) customerMarkerMap.get(current.customerId)?.setIcon(customerMarkerIcon(true))
    else markers.get(current.googlePlaceId)?.setIcon(markerIcon(current, true))
  }
})

watch(radius, drawSearchArea)

async function search() {
  error.value = ''
  success.value = ''
  loading.value = true
  try {
    results.value = await crmApi.searchPlaces({ keyword: keyword.value, categories: categories.value.join(','), radius: radius.value, latitude: latitude.value, longitude: longitude.value })
    selected.value = null
    placeDetails.value = null
    detailOpen.value = false
    await nextTick()
    renderMarkers()
  } catch (caught) {
    error.value = crmError(caught)
    toast.add({ severity: 'error', summary: 'Search failed', detail: error.value, life: 6000 })
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
  if (!selected.value || !salesExecutiveId.value || !industryGroup.value) {
    toast.add({ severity: 'warn', summary: 'Missing information', detail: 'Select a Place, Industry Group, and Sales Executive before saving.', life: 4000 })
    return
  }
  if (selected.value.isCustomer) {
    toast.add({ severity: 'warn', summary: 'Existing customer', detail: 'This place is already an existing customer and cannot be assigned to sales.', life: 4000 })
    return
  }
  error.value = ''
  success.value = ''
  saving.value = true
  try {
    const item = await crmApi.saveProspect(selected.value, industryGroup.value, salesExecutiveId.value)
    success.value = `${item.placeName} saved as NEW_LEAD and assigned successfully.`
    toast.add({ severity: 'success', summary: 'Prospect saved', detail: `${item.placeName} was saved as NEW_LEAD and assigned successfully.`, life: 5000 })
    detailOpen.value = false
  } catch (caught) {
    error.value = crmError(caught)
    toast.add({ severity: 'error', summary: 'Save failed', detail: error.value, life: 6000 })
  } finally {
    saving.value = false
  }
}

function crmError(err: unknown) {
  const candidate = err as { response?: { data?: { error?: { message?: string } } }; message?: string }
  return candidate.response?.data?.error?.message ?? candidate.message ?? 'Prospect Finder request failed.'
}

function parkingLabel(key: string) {
  const labels: Record<string, string> = { freeStreetParking:'Free Street', paidStreetParking:'Paid Street', freeParkingLot:'Free Lot', paidParkingLot:'Paid Lot', valetParking:'Valet', garageParking:'Garage' }
  return labels[key] || key
}

function paymentLabel(key: string) {
  const labels: Record<string, string> = { cashOnly:'Cash Only', creditCardOnly:'Credit Card', debitCardOnly:'Debit Card', nfcOnly:'NFC' }
  return labels[key] || key
}

function accessibilityLabel(key: string) {
  const labels: Record<string, string> = { wheelchairAccessibleEntrance:'Entrance', wheelchairAccessibleParking:'Parking', wheelchairAccessibleRestroom:'Restroom', wheelchairAccessibleSeating:'Seating' }
  return labels[key] || key
}

function optionActive(details: PlaceDetails | null, key: string): boolean {
  if (!details) return false
  const parking = details.parkingOptions as Record<string, boolean> | null
  const payment = details.paymentOptions as Record<string, boolean> | null
  const access = details.accessibilityOptions as Record<string, boolean> | null
  return !!(parking?.[key] || payment?.[key] || access?.[key])
}

onMounted(async () => {
  await nextTick()
  initializeMap()
  useGPS()
  loadCustomerMarkers()
  try {
    sales.value = await crmApi.getSalesExecutives()
    salesExecutiveId.value = sales.value[0]?.id ?? ''
  } catch (caught) {
    error.value = crmError(caught)
  }
})

onBeforeUnmount(() => {
  map?.remove()
  map = null
  customerLayer = null
  markers.clear()
  customerMarkerMap.clear()
})
</script>

<template>
  <section class="finder-page">
    <header class="finder-page-header">
      <div class="finder-heading-left">
        <Button
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          rounded
          class="finder-back"
          @click="$router.back()"
          title="Back"
        />
        <div class="finder-page-title">
          <span class="finder-eyebrow">Prospect Management</span>
          <h1>Prospect Finder</h1>
          <p>Discover nearby businesses, review place details, and assign qualified prospects.</p>
        </div>
      </div>

      <div class="finder-page-stats">
        <div class="finder-stat">
          <span>Location</span>
          <strong>{{ geoResolved ? 'Ready' : 'Detecting' }}</strong>
        </div>
        <div class="finder-stat">
          <span>Radius</span>
          <strong>{{ (radius / 1000).toFixed(1) }} km</strong>
        </div>
        <div class="finder-stat">
          <span>Categories</span>
          <strong>{{ categories.length }}</strong>
        </div>
        <div class="finder-stat">
          <span>Results</span>
          <strong>{{ results.length }}</strong>
        </div>
      </div>
    </header>

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
            <Slider v-model="radius" :min="500" :max="20000" :step="500" class="finder-slider" />
            <div class="radius-range-labels">
              <span>0.5 km</span>
              <span>20 km</span>
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

      </aside>

      <div class="finder-map-stage">
        <div ref="mapElement" class="leaflet-map" role="region" aria-label="OpenStreetMap with Google Places prospect markers" />

        

        <div v-if="loading" class="map-loading-state">
          <span class="map-loading-spinner" />
          <strong>Finding nearby prospects...</strong>
        </div>

        <div v-if="results.length" class="finder-floating-results">
          <div class="floating-results-header">
            <div class="floating-results-title">
              <strong>{{ filteredResults.length }} result{{ filteredResults.length !== 1 ? 's' : '' }}</strong>
              <span v-if="results.length && resultSearch" class="results-filter-note">of {{ results.length }}</span>
            </div>
            <button class="floating-results-close" @click="closeResults" title="Close results"><i class="pi pi-times" /></button>
          </div>
          <div class="floating-results-search">
            <i class="pi pi-search" />
            <input v-model="resultSearch" placeholder="Filter results..." />
          </div>
          <div ref="resultsScroll" class="floating-results-list">
            <button
              v-for="item in filteredResults"
              :key="item.googlePlaceId"
              :data-place-id="item.googlePlaceId"
              class="result-card"
              :class="{ selected: selected?.googlePlaceId === item.googlePlaceId }"
              @click="selectResult(item, true)"
            >
              <span class="result-marker" :class="{ 'is-customer': item.isCustomer }" :style="item.isCustomer ? undefined : { background: item.markerColor }">
                <b v-if="item.isCustomer">Y</b>
                <i v-else :class="item.markerIcon" />
              </span>
              <div class="result-info">
                <div class="result-name-row">
                  <strong>{{ item.name }}</strong>
                  <Tag v-if="item.rating" :value="`★ ${item.rating}`" severity="info" class="result-rating-tag" />
                  <Tag v-if="item.isCustomer" value="Existing Customer" severity="success" class="result-customer-tag" />
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
          <div v-if="!filteredResults.length" class="floating-results-empty">
            <i class="pi pi-filter" />
            <span>No matching results</span>
          </div>
        </div>

        <div class="map-source-badge">
          <i class="pi pi-shield" />
          <div>
            <strong>Private Places search</strong>
            <span>Google Places stays server-side. Map tiles use OpenStreetMap.</span>
          </div>
        </div>

        <div v-if="customerMarkers.length || customersLoading" class="map-customer-badge" :class="{ 'is-loading': customersLoading }">
          <span class="map-customer-dot"><b>Y</b></span>
          <div>
            <strong>{{ customersLoading ? 'Loading existing customers...' : `${customerMarkers.length} existing customer${customerMarkers.length !== 1 ? 's' : ''}` }}</strong>
            <span>Existing customers are always visible and cannot be assigned to sales.</span>
          </div>
        </div>
      </div>
    </div>

    <Dialog v-model:visible="detailOpen" modal header="Place Details" :style="{ width: '520px' }" :closable="true" :breakpoints="{ '576px': '95vw' }">
      <div v-if="placeDetailsLoading" class="dialog-loading"><div class="loading-pulse" /><span>Loading full details...</span></div>
      <div v-else-if="selected" class="detail-dialog">
        <div class="detail-hero-bar">
          <span class="detail-hero" :class="{ 'is-customer': selected.isCustomer }" :style="selected.isCustomer ? undefined : { background: selected.markerColor }">
            <b v-if="selected.isCustomer">Y</b>
            <i v-else :class="selected.markerIcon" />
          </span>
          <div class="detail-hero-info">
            <h2>{{ placeDetails?.placeName || selected.name }}</h2>
            <div class="detail-hero-meta">
              <span>{{ placeDetails?.placeCategory || selected.category }}</span>
              <Tag v-if="selected.isCustomer" value="Existing Customer" severity="success" />
              <Tag v-else-if="(placeDetails?.rating || selected.rating)!" :value="`★ ${(placeDetails?.rating || selected.rating)!.toFixed(1)}`" severity="info" />
              <Tag v-if="placeDetails?.userRatingCount || selected.userRatingCount" :value="`${placeDetails?.userRatingCount || selected.userRatingCount} reviews`" severity="secondary" />
              <Tag v-if="placeDetails?.priceLevel" :value="placeDetails.priceLevel" severity="contrast" />
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.editorialSummary" class="detail-editorial">
          <i class="pi pi-quote-left" /> {{ placeDetails.editorialSummary }}
        </div>

        <div class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-info-circle" /> Basic Info</h3>
          <div class="detail-info-grid">
            <div class="detail-info-item">
              <i class="pi pi-map-marker" />
              <div>
                <span class="detail-info-label">Address</span>
                <span class="detail-info-value">{{ placeDetails?.formattedAddress || selected.address }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.phoneNumber || selected.phone" class="detail-info-item">
              <i class="pi pi-phone" />
              <div>
                <span class="detail-info-label">Phone</span>
                <span class="detail-info-value">{{ placeDetails?.phoneNumber || selected.phone }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.internationalPhone" class="detail-info-item">
              <i class="pi pi-phone" />
              <div>
                <span class="detail-info-label">International</span>
                <span class="detail-info-value">{{ placeDetails.internationalPhone }}</span>
              </div>
            </div>
            <div v-if="placeDetails?.websiteUrl || selected.website" class="detail-info-item">
              <i class="pi pi-globe" />
              <div>
                <span class="detail-info-label">Website</span>
                <a :href="placeDetails?.websiteUrl || selected.website" target="_blank" rel="noreferrer" class="detail-info-link">Open website →</a>
              </div>
            </div>
            <div v-if="placeDetails?.googleMapsUrl || selected.googleMapsUrl" class="detail-info-item">
              <i class="pi pi-external-link" />
              <div>
                <span class="detail-info-label">Google Maps</span>
                <a :href="placeDetails?.googleMapsUrl || selected.googleMapsUrl" target="_blank" rel="noreferrer" class="detail-info-link">View listing →</a>
              </div>
            </div>
            <div class="detail-info-item">
              <i class="pi pi-tag" />
              <div>
                <span class="detail-info-label">Place Types</span>
                <span class="detail-info-value detail-types">{{ placeDetails?.placeTypes?.join(', ') || selected.placeTypes?.join(', ') || selected.markerCategory }}</span>
              </div>
            </div>
            <div class="detail-info-item">
              <i class="pi pi-info-circle" />
              <div>
                <span class="detail-info-label">Status</span>
                <Tag :value="(placeDetails?.businessStatus || selected.businessStatus || 'UNKNOWN')" :severity="(placeDetails?.businessStatus || selected.businessStatus) === 'OPERATIONAL' ? 'success' : 'warn'" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.openingHours" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-clock" /> Opening Hours</h3>
          <div class="detail-hours-grid">
            <div class="detail-hours-badge" :class="{ 'is-open': placeDetails.openingHours.openNow }">
              <span class="hours-dot" />
              {{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}
            </div>
            <div v-for="day in placeDetails.openingHours.weekdays" :key="day" class="detail-hours-day">{{ day }}</div>
          </div>
        </div>

        <div v-if="placeDetails" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-cog" /> Services &amp; Options</h3>
          <div class="detail-options-grid">
            <div v-if="placeDetails.delivery !== undefined" class="detail-option-chip" :class="{ active: placeDetails.delivery }">
              <i class="pi pi-truck" /> <span>Delivery</span>
            </div>
            <div v-if="placeDetails.dineIn !== undefined" class="detail-option-chip" :class="{ active: placeDetails.dineIn }">
              <i class="pi pi-building" /> <span>Dine In</span>
            </div>
            <div v-if="placeDetails.takeout !== undefined" class="detail-option-chip" :class="{ active: placeDetails.takeout }">
              <i class="pi pi-box" /> <span>Takeout</span>
            </div>
            <div v-if="placeDetails.curbsidePickup !== undefined" class="detail-option-chip" :class="{ active: placeDetails.curbsidePickup }">
              <i class="pi pi-car" /> <span>Curbside Pickup</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.parkingOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-car" /> Parking</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['freeStreetParking','paidStreetParking','freeParkingLot','paidParkingLot','valetParking','garageParking']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>{{ parkingLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.paymentOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-credit-card" /> Payment</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['cashOnly','creditCardOnly','debitCardOnly','nfcOnly']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>{{ paymentLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.accessibilityOptions" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-universal-access" /> Accessibility</h3>
          <div class="detail-options-grid">
            <div v-for="key in ['wheelchairAccessibleEntrance','wheelchairAccessibleParking','wheelchairAccessibleRestroom','wheelchairAccessibleSeating']" :key="key" class="detail-option-chip" :class="{ active: optionActive(placeDetails, key) }">
              <i class="pi pi-check-circle" /> <span>Wheelchair {{ accessibilityLabel(key) }}</span>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.reviews?.length" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-comments" /> Reviews ({{ placeDetails.reviews.length }})</h3>
          <div class="detail-reviews-list">
            <div v-for="review in placeDetails.reviews.slice(0, 5)" :key="review.authorName + review.time" class="detail-review">
              <div class="review-author">
                <img v-if="review.authorPhoto" :src="review.authorPhoto" alt="" class="review-author-photo" />
                <span class="review-author-name">{{ review.authorName }}</span>
                <Tag :value="`★ ${review.rating.toFixed(1)}`" severity="info" class="review-rating" />
                <small class="review-time">{{ review.time }}</small>
              </div>
              <p v-if="review.text" class="review-text">{{ review.text }}</p>
            </div>
          </div>
        </div>

        <div v-if="placeDetails?.photos?.length" class="detail-section">
          <h3 class="detail-section-title"><i class="pi pi-images" /> Photos</h3>
          <div class="detail-photos-row">
            <img v-for="photo in placeDetails.photos.slice(0, 6)" :key="photo.name" :src="photo.photoUrl" alt="Place photo" class="detail-photo" loading="lazy" />
          </div>
        </div>

        <div v-if="placeDetailsError" class="detail-section">
          <Message severity="warn" :closable="false">{{ placeDetailsError }}</Message>
        </div>

        <div class="detail-assignment">
          <h3>Assignment</h3>
          <Message v-if="selected?.isCustomer" severity="success" :closable="false" class="assignment-warning">
            <strong>Existing customer</strong> — this place has already been converted to a customer and can no longer be assigned to sales.
          </Message>
          <div v-else class="detail-assignment-fields">
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
          <Button v-if="selected?.isCustomer" label="Existing Customer" icon="pi pi-check" severity="success" disabled />
          <Button v-else label="Save as Prospect" icon="pi pi-save" :loading="saving" :disabled="!salesExecutiveId || !industryGroup" @click="save" />
        </div>
      </template>
    </Dialog>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
  </section>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   PROSPECT FINDER — Workspace Layout (modernized visual pass)
   ════════════════════════════════════════════════════════════════ */

.finder-page {
  box-sizing: border-box;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 0;
  width: calc(100% + 3rem);
  min-width: 0;
  height: calc(100dvh - 4rem);
  min-height: 0;
  margin: -1.5rem;
  padding: 0;
  overflow: hidden;
  background: #f8fafc;
}

.finder-page-header {
  z-index: 5;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.8rem;
  min-height: 64px;
  padding: 0.52rem 0.85rem;
  border: 0;
  border-bottom: 1px solid #e5eaf0;
  border-radius: 0;
  background: #ffffff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.finder-heading-left {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.65rem;
}

.finder-back {
  flex: 0 0 auto;
}

.finder-page-title {
  min-width: 0;
}

.finder-eyebrow {
  color: #64748b;
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.finder-page-title h1 {
  margin: 0.08rem 0 0;
  color: #0f172a;
  font-size: 1.08rem;
  line-height: 1.2;
}

.finder-page-title p {
  margin: 0.12rem 0 0;
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.68rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.finder-page-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(76px, 1fr));
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #f8fafc;
}

.finder-stat {
  display: grid;
  gap: 0.04rem;
  min-width: 76px;
  padding: 0.45rem 0.65rem;
  border-right: 1px solid #e5eaf0;
  text-align: center;
}

.finder-stat:last-child { border-right: 0; }
.finder-stat span { color: #94a3b8; font-size: 0.55rem; font-weight: 800; text-transform: uppercase; }
.finder-stat strong { color: #0f172a; font-size: 0.78rem; }

/* ── Desktop Shell ───────────────────────────────────────────── */
.finder-desktop-shell {
  min-height: 0;
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  overflow: hidden;
  background: var(--surface-card);
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

/* ── Left Panel ──────────────────────────────────────────────── */
.finder-left-panel {
  min-height: 0;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  background: var(--surface-card);
  border-right: 1px solid var(--border-light);
}

.finder-panel-header {
  padding: 0.65rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  background: linear-gradient(150deg, var(--brand-blue-50) 0%, #ffffff 70%);
}

.finder-panel-title {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}

.finder-panel-title > i {
  width: 2.15rem;
  height: 2.15rem;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, var(--brand-blue) 0%, #bb3342 100%);
  border-radius: 0.7rem;
  font-size: 0.85rem;
  flex-shrink: 0;
  box-shadow: 0 4px 12px -2px rgba(209, 67, 80, 0.45);
}

.finder-panel-title h1 {
  margin: 0;
  font-size: 1.02rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  line-height: 1.2;
}

.finder-panel-title span {
  display: block;
  margin-top: 0.15rem;
  color: var(--text-muted);
  font-size: 0.63rem;
  font-weight: 500;
}

.finder-filter-scroll {
  padding: 0.35rem 0.75rem 0.55rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.finder-filter-scroll::-webkit-scrollbar { width: 4px; }
.finder-filter-scroll::-webkit-scrollbar-track { background: transparent; }
.finder-filter-scroll::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }
.finder-filter-scroll::-webkit-scrollbar-thumb:hover { background: var(--text-faint); }

/* ── Filter Sections ─────────────────────────────────────────── */
.filter-section {
  padding: 0.45rem 0;
  border-bottom: 1px solid #eef1f5;
}

.filter-section:last-of-type { border-bottom: 0; }

.filter-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.filter-section-title {
  margin: 0 0 0.5rem;
  color: var(--text-primary);
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.filter-section-header .filter-section-title { margin-bottom: 0; }

.category-count {
  padding: 0.12rem 0.5rem;
  color: var(--brand-blue);
  background: var(--brand-blue-50);
  border: 1px solid var(--brand-blue-100);
  border-radius: 999px;
  font-size: 0.55rem;
  font-weight: 700;
}

/* Keyword */
.finder-keyword-field { gap: 0.32rem; }
.finder-keyword-field > span { color: var(--text-muted); font-size: 0.65rem; font-weight: 700; }

.keyword-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.keyword-icon {
  position: absolute;
  left: 0.7rem;
  color: var(--text-faint);
  font-size: 0.7rem;
  pointer-events: none;
}

.keyword-input-wrap :deep(.p-inputtext) {
  padding-left: 1.9rem;
  border-radius: 0.6rem;
  border-color: var(--border-default);
  font-size: 0.78rem;
  padding-top: 0.52rem;
  padding-bottom: 0.52rem;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.keyword-input-wrap :deep(.p-inputtext:focus) {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(209, 67, 80, 0.10);
}

/* Categories */
.category-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.3rem;
}

.category-chip {
  display: flex;
  gap: 0.28rem;
  align-items: center;
  padding: 0.3rem 0.4rem;
  border-radius: 0.55rem;
  background: var(--surface-subtle);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 160ms ease, border-color 160ms ease, transform 160ms ease;
}

.category-chip:hover { background: var(--surface-hover); border-color: var(--border-default); transform: translateY(-1px); }

.category-chip.active {
  background: linear-gradient(135deg, var(--brand-blue-50) 0%, #fff5f6 100%);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 1px rgba(209, 67, 80, 0.14), 0 2px 6px -2px rgba(209, 67, 80, 0.25);
}

.category-chip span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 0.58rem;
  font-weight: 550;
  transition: color 160ms ease;
}

.category-chip.active span { color: var(--brand-blue); font-weight: 700; }

/* Radius */
.radius-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.45rem;
}

.radius-value {
  padding: 0.16rem 0.55rem;
  color: var(--brand-blue);
  background: linear-gradient(135deg, var(--brand-blue-50), #fff5f6);
  border: 1px solid var(--brand-blue-100);
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 800;
}

.radius-range-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.25rem;
  color: var(--text-faint);
  font-size: 0.52rem;
}

/* Coordinates */
.coordinate-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.coordinate-grid .field { gap: 0.22rem; }
.coordinate-grid .field > span { color: var(--text-muted); font-size: 0.58rem; font-weight: 700; }

.coordinate-grid input {
  width: 100%;
  padding: 0.42rem 0.6rem;
  border: 1px solid var(--border-default);
  border-radius: 0.55rem;
  font-size: 0.7rem;
  color: var(--text-primary);
  background: var(--surface-card);
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.coordinate-grid input:focus {
  outline: none;
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(209, 67, 80, 0.10);
}

/* Filter Actions */
.filter-actions {
  margin-top: 0.55rem;
  display: grid;
  grid-template-columns: 0.8fr 1.2fr;
  gap: 0.45rem;
}

.filter-actions :deep(.p-button) {
  padding: 0.52rem;
  font-size: 0.68rem;
  font-weight: 700;
  border-radius: 0.6rem;
  transition: transform 160ms ease, box-shadow 160ms ease, filter 160ms ease;
}

.filter-actions :deep(.p-button:not(:disabled):hover) {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

/* ── Floating Results Panel ──────────────────────────────────── */
.finder-floating-results {
  position: absolute;
  z-index: 600;
  top: 0.75rem;
  right: 0.75rem;
  width: 340px;
  max-height: calc(100% - 1.5rem);
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(221, 229, 239, 0.9);
  border-radius: 1rem;
  box-shadow: 0 8px 32px rgba(16, 24, 40, 0.14), 0 2px 8px rgba(16, 24, 40, 0.06);
  backdrop-filter: blur(16px) saturate(1.4);
  -webkit-backdrop-filter: blur(16px) saturate(1.4);
  overflow: hidden;
  animation: float-results-in 0.25s ease both;
}

@keyframes float-results-in {
  from { opacity: 0; transform: translateX(12px) scale(0.98); }
  to { opacity: 1; transform: translateX(0) scale(1); }
}

.floating-results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.floating-results-title {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.floating-results-title strong {
  font-size: 0.72rem;
  font-weight: 800;
  color: var(--text-primary);
}

.floating-results-title .results-filter-note {
  color: var(--text-muted);
  font-size: 0.6rem;
  font-weight: 500;
}

.floating-results-close {
  width: 24px;
  height: 24px;
  display: grid;
  place-items: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-faint);
  cursor: pointer;
  font-size: 0.6rem;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.floating-results-close:hover {
  background: var(--surface-hover);
  color: var(--text-primary);
}

.floating-results-search {
  position: relative;
  padding: 0.5rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.floating-results-search i {
  position: absolute;
  left: 1.15rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-faint);
  font-size: 0.65rem;
  pointer-events: none;
}

.floating-results-search input {
  width: 100%;
  padding: 0.38rem 0.6rem 0.38rem 1.55rem;
  border: 1px solid var(--border-default);
  border-radius: 0.55rem;
  background: var(--surface-subtle);
  font-size: 0.68rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.floating-results-search input:focus {
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 4px rgba(209, 67, 80, 0.10);
}

.floating-results-search input::placeholder { color: var(--text-faint); }

.floating-results-list {
  flex: 1;
  min-height: 0;
  padding: 0.5rem;
  overflow-y: auto;
  display: grid;
  gap: 0.35rem;
  align-content: start;
}

.floating-results-list::-webkit-scrollbar { width: 4px; }
.floating-results-list::-webkit-scrollbar-track { background: transparent; }
.floating-results-list::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }
.floating-results-list::-webkit-scrollbar-thumb:hover { background: var(--text-faint); }

.floating-results-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 1.5rem;
  color: var(--text-faint);
}

.floating-results-empty i { font-size: 1.2rem; }
.floating-results-empty span { font-size: 0.72rem; }

.result-card {
  width: 100%;
  padding: 0.65rem;
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 0.6rem;
  align-items: center;
  text-align: left;
  color: var(--text-primary);
  background: var(--surface-card);
  border: 1px solid #eef1f5;
  border-radius: 0.85rem;
  cursor: pointer;
  transition: border-color 180ms ease, background 180ms ease, box-shadow 180ms ease, transform 180ms ease;
}

.result-card:hover {
  border-color: var(--brand-blue-100);
  background: #f8faff;
  box-shadow: 0 4px 14px -4px rgba(209, 67, 80, 0.16);
  transform: translateY(-1px);
}

.result-card.selected {
  border-color: var(--brand-blue);
  background: linear-gradient(135deg, #f5f8ff 0%, #eef3ff 100%);
  box-shadow: 0 0 0 2px rgba(209, 67, 80, 0.12), 0 6px 16px -4px rgba(209, 67, 80, 0.18);
}

.result-marker {
  width: 1.95rem;
  height: 1.95rem;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 50%;
  font-size: 0.6rem;
  flex-shrink: 0;
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.16);
}

.result-marker.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  box-shadow: 0 3px 8px rgba(21, 128, 61, 0.35);
}

.result-marker.is-customer b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 0.95rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
}

.result-customer-tag { transform: scale(0.85); transform-origin: left; }

.result-info {
  min-width: 0;
  display: grid;
  gap: 0.12rem;
}

.result-name-row {
  display: flex;
  align-items: center;
  gap: 0.38rem;
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
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-meta-row {
  display: flex;
  align-items: center;
  gap: 0.38rem;
  margin-top: 0.12rem;
}

.result-distance {
  display: flex;
  align-items: center;
  gap: 0.22rem;
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
  transition: color 160ms ease, transform 160ms ease;
}

.result-card:hover .result-chevron,
.result-card.selected .result-chevron { color: var(--brand-blue); transform: translateX(2px); }

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
  box-shadow: 0 6px 20px rgba(30, 54, 84, 0.16);
  border-radius: 0.7rem !important;
  overflow: hidden;
}

.leaflet-map :deep(.leaflet-control-zoom a) { color: #26344b; }

.leaflet-map :deep(.leaflet-control-attribution) {
  color: #617087;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 0.4rem 0 0 0;
  font-size: 9px;
}

/* ── Map Source Badge ────────────────────────────────────────── */
.map-source-badge {
  position: absolute;
  z-index: 500;
  left: 0.8rem;
  bottom: 0.8rem;
  max-width: 260px;
  padding: 0.55rem 0.75rem;
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(221, 229, 239, 0.9);
  border-radius: 0.7rem;
  box-shadow: 0 6px 20px rgba(30, 54, 84, 0.12);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
}

.map-source-badge > i {
  margin-top: 0.02rem;
  color: #16a34a;
  font-size: 0.75rem;
  filter: drop-shadow(0 1px 1px rgba(22, 163, 74, 0.25));
}
.map-source-badge div { display: grid; gap: 0.12rem; }
.map-source-badge strong { font-size: 0.58rem; font-weight: 700; }
.map-source-badge span { color: #718096; font-size: 0.52rem; line-height: 1.5; }

/* ── Existing Customer Badge ─────────────────────────────────── */
.map-customer-badge {
  position: absolute;
  z-index: 500;
  left: 0.8rem;
  bottom: 4.6rem;
  max-width: 280px;
  padding: 0.55rem 0.75rem;
  display: flex;
  gap: 0.55rem;
  align-items: flex-start;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(187, 247, 208, 0.9);
  border-radius: 0.7rem;
  box-shadow: 0 6px 20px rgba(21, 128, 61, 0.14);
  backdrop-filter: blur(14px) saturate(1.4);
  -webkit-backdrop-filter: blur(14px) saturate(1.4);
}

.map-customer-badge.is-loading { opacity: 0.7; }

.map-customer-dot {
  width: 1.7rem;
  height: 1.7rem;
  margin-top: 0.02rem;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-radius: 50%;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.22), 0 3px 8px rgba(21, 128, 61, 0.35);
}

.map-customer-dot b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 0.95rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
}

.map-customer-badge div { display: grid; gap: 0.1rem; }
.map-customer-badge strong { font-size: 0.6rem; font-weight: 700; color: var(--text-primary); }
.map-customer-badge span { color: #718096; font-size: 0.52rem; line-height: 1.5; }

.map-empty-state,
.map-loading-state {
  position: absolute;
  z-index: 450;
  left: 50%;
  top: 50%;
  display: flex;
  width: min(340px, calc(100% - 2rem));
  transform: translate(-50%, -50%);
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  padding: 1.1rem 1.25rem;
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.93);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.12);
  text-align: center;
  backdrop-filter: blur(12px);
}

.map-empty-icon {
  display: grid;
  width: 42px;
  height: 42px;
  place-content: center;
  border-radius: 12px;
  background: #fff1f2;
  color: #d14350;
}

.map-empty-state strong,
.map-loading-state strong { color: #0f172a; font-size: 0.82rem; }
.map-empty-state p { margin: 0; color: #64748b; font-size: 0.68rem; line-height: 1.45; }

.map-loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid #ffd9dd;
  border-top-color: #d14350;
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

/* ── Detail Dialog ───────────────────────────────────────────── */
.detail-dialog {
  display: grid;
  gap: 1rem;
}

.dialog-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--text-muted);
  font-size: 0.75rem;
}

.loading-pulse {
  width: 28px;
  height: 28px;
  border: 3px solid var(--brand-blue-100);
  border-top-color: var(--brand-blue);
  border-radius: 50%;
  animation: finder-spin 0.75s linear infinite;
}

@keyframes finder-spin { to { transform: rotate(360deg); } }

.detail-hero-bar {
  display: flex;
  gap: 0.8rem;
  align-items: flex-start;
}

.detail-hero {
  width: 2.85rem;
  height: 2.85rem;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  color: #fff;
  border-radius: 0.85rem;
  font-size: 1rem;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.16);
}

.detail-hero.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.2), 0 6px 16px rgba(21, 128, 61, 0.35);
}

.detail-hero.is-customer b {
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 1.35rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
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
  gap: 0.38rem;
  margin-top: 0.32rem;
  flex-wrap: wrap;
}

.detail-hero-meta > span {
  color: var(--text-muted);
  font-size: 0.7rem;
  font-weight: 500;
}

.detail-editorial {
  display: flex;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-style: italic;
  line-height: 1.5;
}

.detail-editorial i {
  color: var(--brand-blue);
  font-size: 0.65rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.detail-section {
  display: grid;
  gap: 0.65rem;
}

.detail-section-title {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.detail-section-title i {
  font-size: 0.6rem;
  color: var(--brand-blue);
}

.detail-info-grid {
  display: grid;
  gap: 0.6rem;
}

.detail-info-item {
  display: flex;
  gap: 0.7rem;
  align-items: flex-start;
}

.detail-info-item > i {
  margin-top: 0.12rem;
  color: var(--brand-blue);
  font-size: 0.72rem;
  width: 1rem;
  text-align: center;
  flex-shrink: 0;
}

.detail-info-item > div {
  display: grid;
  gap: 0.08rem;
}

.detail-info-label {
  color: var(--text-muted);
  font-size: 0.55rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-info-value {
  color: var(--text-primary);
  font-size: 0.78rem;
  line-height: 1.5;
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
  transition: opacity 160ms ease;
}

.detail-info-link:hover { opacity: 0.72; }

/* Hours */
.detail-hours-grid {
  display: grid;
  gap: 0.3rem;
}

.detail-hours-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 700;
  width: fit-content;
  background: #fef2f2;
  color: #dc2626;
}

.detail-hours-badge.is-open {
  background: #f0fdf4;
  color: #16a34a;
}

.hours-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.detail-hours-day {
  padding: 0.2rem 0.65rem;
  color: var(--text-secondary);
  font-size: 0.62rem;
  line-height: 1.5;
  border-bottom: 1px solid var(--border-light);
}

.detail-hours-day:last-child { border-bottom: 0; }

/* Options grid */
.detail-options-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.detail-option-chip {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.55rem;
  border-radius: 999px;
  font-size: 0.6rem;
  font-weight: 500;
  color: var(--text-faint);
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
}

.detail-option-chip.active {
  color: #16a34a;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.detail-option-chip i { font-size: 0.55rem; }

/* Reviews */
.detail-reviews-list {
  display: grid;
  gap: 0.65rem;
}

.detail-review {
  padding: 0.65rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.review-author {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.review-author-photo {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  object-fit: cover;
}

.review-author-name {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-primary);
}

.review-rating { transform: scale(0.8); transform-origin: left; }

.review-time {
  color: var(--text-faint);
  font-size: 0.55rem;
}

.review-text {
  margin: 0.4rem 0 0;
  color: var(--text-secondary);
  font-size: 0.65rem;
  line-height: 1.5;
}

/* Photos */
.detail-photos-row {
  display: flex;
  gap: 0.4rem;
  overflow-x: auto;
  padding-bottom: 0.25rem;
}

.detail-photo {
  width: 80px;
  height: 80px;
  border-radius: var(--radius-sm);
  object-fit: cover;
  flex-shrink: 0;
  border: 1px solid var(--border-light);
}

.detail-photo::-webkit-scrollbar { height: 3px; }
.detail-photo::-webkit-scrollbar-thumb { background: var(--border-default); border-radius: 999px; }

.detail-assignment {
  padding-top: 0.75rem;
  border-top: 1px solid var(--border-light);
}

.detail-assignment h3 {
  margin: 0 0 0.7rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.detail-assignment-fields {
  display: grid;
  gap: 0.7rem;
}

.detail-assignment-fields .field {
  display: grid;
  gap: 0.32rem;
}

.detail-assignment-fields .field > span {
  color: var(--text-muted);
  font-size: 0.65rem;
  font-weight: 700;
}

.detail-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.55rem;
}

.assignment-warning {
  font-size: 0.72rem;
  margin-top: 0.18rem;
  border-radius: 0.6rem;
}

/* ── Responsive ──────────────────────────────────────────────── */
@media (max-width: 1180px) {
  .finder-page-header { grid-template-columns: 1fr; }
  .finder-page-stats { width: 100%; }
  .finder-desktop-shell { grid-template-columns: 310px minmax(0, 1fr); }
  .category-grid { grid-template-columns: 1fr 1fr; }
  .finder-floating-results { width: 300px; }
}

@media (max-width: 900px) {
  .finder-page {
    height: auto;
    min-height: 100vh;
    overflow: visible;
    padding: 0.75rem;
  }
  .finder-desktop-shell {
    min-height: 780px;
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(500px, 1fr);
    overflow: visible;
  }
  .finder-left-panel { border-right: 0; border-bottom: 1px solid var(--border-light); }
  .finder-filter-scroll { overflow: visible; }
  .finder-floating-results {
    top: 0.5rem;
    right: 0.5rem;
    width: calc(100% - 1rem);
    max-height: calc(100% - 1rem);
  }
}

@media (max-width: 640px) {
  .finder-page-header { align-items: stretch; }
  .finder-heading-left { align-items: flex-start; }
  .finder-page-title p { white-space: normal; }
  .finder-page-stats { grid-template-columns: repeat(2, 1fr); }
  .finder-stat:nth-child(2) { border-right: 0; }
  .finder-stat:nth-child(-n + 2) { border-bottom: 1px solid #e5eaf0; }
  .finder-panel-title h1 { font-size: 0.9rem; }
  .map-source-badge { max-width: calc(100% - 1.6rem); }
  .category-grid { grid-template-columns: repeat(2, 1fr); }
  .coordinate-grid, .filter-actions { grid-template-columns: 1fr; }
  .finder-floating-results {
    top: 0;
    right: 0;
    width: 100%;
    max-height: 100%;
    border-radius: 0;
    border: none;
  }
}

@supports not (height: 100dvh) {
  .finder-page {
    height: calc(100vh - 4rem);
  }
}

@media (max-height: 760px) and (min-width: 901px) {
  .finder-page-header {
    min-height: 56px;
    padding-block: 0.38rem;
  }

  .finder-page-title p,
  .finder-panel-title span {
    display: none;
  }

  .finder-panel-header {
    padding: 0.45rem 0.75rem;
  }

  .finder-filter-scroll {
    padding-top: 0.2rem;
  }

  .filter-section {
    padding: 0.32rem 0;
  }
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
  box-shadow: 0 6px 16px rgba(22, 41, 67, 0.32), 0 0 0 1px rgba(22, 41, 67, 0.04);
  transform: rotate(-45deg);
  transition: width 160ms cubic-bezier(0.4, 0, 0.2, 1),
              height 160ms cubic-bezier(0.4, 0, 0.2, 1),
              box-shadow 160ms cubic-bezier(0.4, 0, 0.2, 1);
}

.finder-leaflet-marker i { transform: rotate(45deg); font-size: 0.8rem; }

.finder-leaflet-marker.is-customer {
  background: linear-gradient(135deg, #22c55e 0%, #15803d 100%);
  border-color: #ffffff;
  border-radius: 50%;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.25), 0 6px 16px rgba(21, 128, 61, 0.4);
}

.finder-leaflet-marker.is-customer b {
  transform: rotate(45deg);
  font-family: 'Segoe UI', system-ui, sans-serif;
  font-size: 1.15rem;
  font-weight: 800;
  font-style: italic;
  line-height: 1;
  letter-spacing: -0.04em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
}

.finder-leaflet-marker.is-selected {
  width: 44px;
  height: 44px;
  box-shadow: 0 0 0 6px rgba(209, 67, 80, 0.16),
              0 8px 22px rgba(22, 41, 67, 0.38);
}

.finder-leaflet-marker.is-customer.is-selected {
  width: 46px;
  height: 46px;
  box-shadow: 0 0 0 6px rgba(34, 197, 94, 0.2), 0 8px 22px rgba(21, 128, 61, 0.42);
}

.finder-leaflet-marker.is-selected i { font-size: 1rem; }
.finder-leaflet-marker.is-customer.is-selected b { font-size: 1.4rem; }
</style>