<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

const props = withDefaults(defineProps<{
  latitude: number | null
  longitude: number | null
  label?: string
  height?: string
  zoom?: number
  interactive?: boolean
}>(), {
  label: 'Location',
  height: '200px',
  zoom: 16,
  interactive: true,
})

const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let marker: L.Marker | null = null

function toNumber(v: number | null): number | null {
  if (v == null) return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function createMarkerIcon(): L.DivIcon {
  return L.divIcon({
    className: 'entity-map-icon-host',
    html: `<span class="entity-map-marker"><i class="pi pi-map-marker"></i></span>`,
    iconSize: [36, 42],
    iconAnchor: [18, 40],
    popupAnchor: [0, -42],
  })
}

async function initMap() {
  if (!mapContainer.value) return
  const lat = toNumber(props.latitude)
  const lng = toNumber(props.longitude)
  if (lat == null || lng == null) return

  if (map) { map.remove(); map = null; marker = null }

  await nextTick()

  map = L.map(mapContainer.value, {
    zoomControl: true,
    preferCanvas: true,
    attributionControl: false,
    dragging: props.interactive,
    scrollWheelZoom: props.interactive,
    doubleClickZoom: props.interactive,
    touchZoom: props.interactive,
  }).setView([lat, lng], props.zoom)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  marker = L.marker([lat, lng], { icon: createMarkerIcon() }).addTo(map)
  if (props.label) marker.bindPopup(`<span style="font-size:0.8rem;font-weight:600">${props.label}</span>`).openPopup()

  setTimeout(() => { map?.invalidateSize() }, 150)
}

function cleanup() {
  if (map) { map.remove(); map = null; marker = null }
}

watch(() => [props.latitude, props.longitude], () => { cleanup(); initMap() })

onMounted(() => initMap())
onBeforeUnmount(() => cleanup())
</script>

<template>
  <div class="entity-map-wrapper" :style="{ height }">
    <div v-if="latitude != null && longitude != null" ref="mapContainer" class="entity-map-container" />
    <div v-else class="entity-map-empty">
      <i class="pi pi-map-marker" />
      <span>No coordinates available</span>
    </div>
  </div>
</template>

<style>
.entity-map-icon-host { border: 0; background: transparent; }
.entity-map-marker {
  width: 36px; height: 36px; display: grid; place-items: center;
  color: #fff; background: #2563eb; border: 3px solid #fff;
  border-radius: 50% 50% 50% 0; box-shadow: 0 4px 14px rgba(22, 41, 67, 0.3);
  transform: rotate(-45deg);
}
.entity-map-marker i { transform: rotate(45deg); font-size: 0.8rem; }
.entity-map-wrapper .leaflet-popup { max-width: min(220px, 80vw) !important; }
.entity-map-wrapper .leaflet-popup-content-wrapper { font-size: 0.78rem; border-radius: 10px; }
.entity-map-wrapper .leaflet-popup-tip { display: none; }
</style>

<style scoped>
.entity-map-wrapper {
  width: 100%; min-width: 0; border-radius: 12px; overflow: hidden;
  border: 1px solid var(--border-light, #e2e8f0);
  position: relative;
}
.entity-map-container {
  width: 100%; height: 100%;
}
.entity-map-empty {
  width: 100%; height: 100%; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 0.5rem;
  background: #f1f5f9; color: #94a3b8; font-size: 0.78rem;
}
.entity-map-empty i { font-size: 1.4rem; }
</style>
