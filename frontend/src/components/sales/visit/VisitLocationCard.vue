<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import type { GpsCoords } from '../../composables/sales/useVisitLocation'

const props = withDefaults(defineProps<{
  targetLatitude: number | null
  targetLongitude: number | null
  targetLabel?: string
  salesCoords: GpsCoords | null
  radiusMeters?: number
  height?: string
}>(), {
  targetLabel: 'Target',
  radiusMeters: 100,
  height: '190px',
})

const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let targetMarker: L.Marker | null = null
let salesMarker: L.Marker | null = null
let radiusCircle: L.Circle | null = null

function toNum(v: number | null): number | null {
  if (v == null) return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function targetIcon(): L.DivIcon {
  return L.divIcon({
    className: 'vlm-icon-host',
    html: `<span class="vlm-marker vlm-marker-target"><i class="pi pi-map-marker"></i></span>`,
    iconSize: [32, 38],
    iconAnchor: [16, 36],
  })
}

function salesIcon(): L.DivIcon {
  return L.divIcon({
    className: 'vlm-icon-host',
    html: `<span class="vlm-marker vlm-marker-sales"><i class="pi pi-user"></i></span>`,
    iconSize: [32, 38],
    iconAnchor: [16, 36],
  })
}

async function initMap() {
  if (!mapContainer.value) return
  const tLat = toNum(props.targetLatitude)
  const tLng = toNum(props.targetLongitude)
  if (tLat == null || tLng == null) return

  if (map) { map.remove(); map = null; targetMarker = null; salesMarker = null; radiusCircle = null }

  await nextTick()

  map = L.map(mapContainer.value, {
    zoomControl: false,
    preferCanvas: true,
    attributionControl: false,
    dragging: false,
    scrollWheelZoom: false,
    doubleClickZoom: false,
    touchZoom: false,
  }).setView([tLat, tLng], 16)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  targetMarker = L.marker([tLat, tLng], { icon: targetIcon() }).addTo(map)
  if (props.targetLabel) targetMarker.bindPopup(`<span style="font-size:0.78rem;font-weight:600">${props.targetLabel}</span>`)

  radiusCircle = L.circle([tLat, tLng], {
    radius: props.radiusMeters,
    color: '#2563eb',
    fillColor: '#2563eb',
    fillOpacity: 0.08,
    weight: 2,
    dashArray: '6 4',
  }).addTo(map)

  updateSalesMarker()
  fitBounds()

  setTimeout(() => { map?.invalidateSize() }, 150)
}

function updateSalesMarker() {
  if (!map) return
  const sLat = toNum(props.salesCoords?.latitude ?? null)
  const sLng = toNum(props.salesCoords?.longitude ?? null)
  if (sLat == null || sLng == null) return

  if (salesMarker) {
    salesMarker.setLatLng([sLat, sLng])
  } else {
    salesMarker = L.marker([sLat, sLng], { icon: salesIcon() }).addTo(map)
    salesMarker.bindPopup('<span style="font-size:0.78rem;font-weight:600">Your location</span>')
  }
}

function fitBounds() {
  if (!map) return
  const tLat = toNum(props.targetLatitude)
  const tLng = toNum(props.targetLongitude)
  const sLat = toNum(props.salesCoords?.latitude ?? null)
  const sLng = toNum(props.salesCoords?.longitude ?? null)

  if (tLat != null && tLng != null && sLat != null && sLng != null) {
    const bounds = L.latLngBounds([tLat, tLng], [sLat, sLng])
    map.fitBounds(bounds, { padding: [40, 40], maxZoom: 17 })
  } else if (tLat != null && tLng != null) {
    map.setView([tLat, tLng], 16)
  }
}

function cleanup() {
  if (map) { map.remove(); map = null; targetMarker = null; salesMarker = null; radiusCircle = null }
}

watch(() => props.salesCoords, () => { updateSalesMarker(); fitBounds() })
watch(() => [props.targetLatitude, props.targetLongitude], () => { cleanup(); initMap() })

onMounted(() => initMap())
onBeforeUnmount(() => cleanup())
</script>

<template>
  <div class="vlm-wrapper" :style="{ height }">
    <div v-if="targetLatitude != null && targetLongitude != null" ref="mapContainer" class="vlm-container" />
    <div v-else class="vlm-empty">
      <i class="pi pi-map-marker" />
      <span>No coordinates available</span>
    </div>
    <div class="vlm-legend">
      <span class="vlm-legend-item"><span class="vlm-dot vlm-dot-target" /> Target</span>
      <span v-if="salesCoords" class="vlm-legend-item"><span class="vlm-dot vlm-dot-sales" /> You</span>
      <span class="vlm-legend-item"><span class="vlm-ring" /> {{ radiusMeters }}m radius</span>
    </div>
  </div>
</template>

<style>
.vlm-icon-host { border: 0; background: transparent; }
.vlm-marker {
  width: 32px; height: 32px; display: grid; place-items: center;
  color: #fff; border: 3px solid #fff; border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg); box-shadow: 0 3px 12px rgba(0,0,0,0.25);
}
.vlm-marker i { transform: rotate(45deg); font-size: 0.7rem; }
.vlm-marker-target { background: #2563eb; }
.vlm-marker-sales { background: #dc2626; }
.vlm-wrapper .leaflet-popup { max-width: min(200px, 75vw) !important; }
.vlm-wrapper .leaflet-popup-content-wrapper { font-size: 0.75rem; border-radius: 10px; }
.vlm-wrapper .leaflet-popup-tip { display: none; }
</style>

<style scoped>
.vlm-wrapper { width: 100%; min-width: 0; border-radius: 12px; overflow: hidden; border: 1px solid var(--border-light, #e2e8f0); position: relative; isolation: isolate; z-index: 0; touch-action: pan-x pan-y; }
.vlm-container { width: 100%; height: 100%; }
.vlm-empty { width: 100%; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.5rem; background: #f1f5f9; color: #94a3b8; font-size: 0.78rem; }
.vlm-empty i { font-size: 1.4rem; }
.vlm-legend {
  position: absolute; bottom: 6px; left: 6px; z-index: 10;
  display: flex; gap: 0.55rem; padding: 0.3rem 0.6rem; border-radius: 8px;
  background: rgba(255,255,255,0.92); backdrop-filter: blur(4px);
  font-size: 0.6rem; color: #475569; box-shadow: 0 1px 4px rgba(0,0,0,0.1);
}
.vlm-legend-item { display: flex; align-items: center; gap: 0.25rem; white-space: nowrap; }
.vlm-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.vlm-dot-target { background: #2563eb; }
.vlm-dot-sales { background: #dc2626; }
.vlm-ring { width: 8px; height: 8px; border-radius: 50%; border: 1.5px dashed #2563eb; flex-shrink: 0; }
</style>
