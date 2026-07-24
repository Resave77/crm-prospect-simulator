<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyCustomer } from '../../../api/crm'
import type { CustomerDetail } from '../../../types/crm'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'

const route = useRoute()
const detail = ref<CustomerDetail | null>(null)
const error = ref('')
const loading = ref(true)
const mapElement = ref<HTMLElement | null>(null)
let map: L.Map | null = null
const userCoords = ref<{ lat: number; lng: number } | null>(null)
let geoWatchId: number | null = null

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function customerPhone(): string {
  return detail.value?.customer.contacts?.[0]?.phone ?? ''
}

function customerEmail(): string {
  return detail.value?.customer.contacts?.[0]?.email ?? ''
}

function customerContactName(): string {
  return detail.value?.customer.contacts?.[0]?.name ?? ''
}

function customerContactPosition(): string {
  return detail.value?.customer.contacts?.[0]?.position ?? ''
}

function renderMap() {
  const item = detail.value?.customer
  if (!mapElement.value || item?.address.latitude == null || item.address.longitude == null) return
  map?.remove()
  map = L.map(mapElement.value, { zoomControl: true }).setView([item.address.latitude, item.address.longitude], 16)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap contributors' }).addTo(map)
  L.marker([item.address.latitude, item.address.longitude]).addTo(map).bindPopup(item.name).openPopup()
}

function navigate() {
  const c = detail.value?.customer
  if (!c) return
  openGoogleMapsNavigation({
    latitude: c.address?.latitude,
    longitude: c.address?.longitude,
    address: c.address?.previewAddress,
  })
}

function acquireGPS() {
  if (!navigator.geolocation) return
  geoWatchId = navigator.geolocation.watchPosition(
    (pos) => { userCoords.value = { lat: pos.coords.latitude, lng: pos.coords.longitude } },
    () => {},
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

onMounted(async () => {
  acquireGPS()
  try {
    detail.value = await getMyCustomer(String(route.params.id))
    await nextTick()
    renderMap()
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Unable to load customer.'
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => { map?.remove(); if (geoWatchId != null) navigator.geolocation?.clearWatch(geoWatchId) })
</script>

<template>
  <section class="detail-page">
    <RouterLink class="back-link" to="/sales/my-customers"><i class="pi pi-arrow-left" /> My Customers</RouterLink>

    <!-- Loading skeleton -->
    <div v-if="loading" class="detail-skeleton">
      <div class="sk-header"><div class="sk-circle" /><div class="sk-lines"><div class="sk-line w70" /><div class="sk-line w40" /></div></div>
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /><div class="sk-line w60" /></div>
    </div>

    <!-- Error -->
    <Message v-else-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <!-- Not found -->
    <div v-else-if="!detail" class="detail-empty">
      <div class="detail-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>Customer not found</strong>
      <span>This customer may have been removed or you don't have access.</span>
      <RouterLink class="detail-empty-btn" to="/sales/my-customers"><i class="pi pi-arrow-left" /> Back to customers</RouterLink>
    </div>

    <!-- Detail content -->
    <template v-else>
      <!-- Summary Card -->
      <div class="dcard dcard-summary">
        <div class="dcard-summary-top">
          <div class="dcard-avatar">{{ initials(detail.customer.name || 'Customer') }}</div>
          <div class="dcard-identity">
            <p class="eyebrow">Customer Existing</p>
            <h1>{{ detail.customer.name }}</h1>
            <small v-if="detail.customer.parentCompanyName">{{ detail.customer.parentCompanyName }}</small>
          </div>
        </div>
        <div class="dcard-codes">
          <div class="dcard-code-item"><span>Customer code</span><strong>{{ detail.customer.customerCode }}</strong></div>
          <div class="dcard-code-item"><span>Parent code</span><strong>{{ detail.customer.parentCode }}</strong></div>
        </div>
        <div class="dcard-tags">
          <Tag value="ACTIVE" severity="success" />
          <Tag v-if="detail.customer.segment" :value="detail.customer.segment" />
          <Tag v-if="detail.customer.category" :value="detail.customer.category" severity="secondary" />
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard">
        <h2>Location</h2>
        <div v-if="detail.customer.address.latitude != null && detail.customer.address.longitude != null" ref="mapElement" class="dcard-map" role="region" aria-label="Customer location map" />
        <Message v-else severity="warn" :closable="false">No saved coordinates for this customer.</Message>
        <div class="dcard-location-rows">
          <div class="dcard-row"><i class="pi pi-map-marker" /><span>{{ detail.customer.address.previewAddress || 'No address' }}</span></div>
          <div v-if="detail.customer.region" class="dcard-row"><i class="pi pi-globe" /><span>Region: {{ detail.customer.region }}</span></div>
          <div v-if="detail.customer.address.latitude != null && detail.customer.address.longitude != null && userCoords" class="dcard-row dcard-distance">
            <i class="pi pi-compass" />
            <span>{{ formatDistance(getDistanceTo(detail.customer.address.latitude, detail.customer.address.longitude, userCoords.lat, userCoords.lng)!) }} away</span>
          </div>
        </div>
      </div>

      <!-- Contact Card -->
      <div class="dcard">
        <h2>Contact</h2>
        <div v-if="detail.customer.contacts.length" class="dcard-contact-list">
          <div v-for="contact in detail.customer.contacts" :key="`${contact.name}-${contact.phone}`" class="dcard-contact">
            <div class="dcard-contact-info">
              <strong>{{ contact.name }}</strong>
              <span v-if="contact.position">{{ contact.position }}</span>
            </div>
            <div class="dcard-contact-actions">
              <a v-if="contact.phone" :href="`tel:${contact.phone}`" class="dcard-action-sm dcard-action-call"><i class="pi pi-phone" /> {{ contact.phone }}</a>
              <a v-if="contact.email" :href="`mailto:${contact.email}`" class="dcard-action-sm dcard-action-email"><i class="pi pi-envelope" /> {{ contact.email }}</a>
            </div>
          </div>
        </div>
        <p v-else class="dcard-empty-text">No contacts on file.</p>
      </div>

      <!-- Conversion Source -->
      <div class="dcard">
        <h2>Conversion Source</h2>
        <div class="dcard-rows">
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Prospect:</strong> {{ detail.sourceProspectName }}</span></div>
          <div class="dcard-row"><i class="pi pi-id-card" /><span><strong>Source ID:</strong> {{ detail.customer.sourceProspectId || '—' }}</span></div>
          <div class="dcard-row"><i class="pi pi-calendar" /><span><strong>Converted:</strong> {{ new Date(detail.customer.convertedAt).toLocaleDateString('en', { year: 'numeric', month: 'short', day: 'numeric' }) }}</span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Sales Executive:</strong> {{ detail.customer.salesExecutiveName }}</span></div>
        </div>
      </div>

      <!-- Bottom Action Bar -->
      <div class="detail-bottom-bar">
        <button class="dbar-btn dbar-navigate" :disabled="detail.customer.address?.latitude == null && detail.customer.address?.longitude == null && !detail.customer.address?.previewAddress" @click="navigate">
          <i class="pi pi-directions" /> Navigate
        </button>
        <a v-if="customerPhone()" :href="`tel:${customerPhone()}`" class="dbar-btn dbar-call">
          <i class="pi pi-phone" /> Call
        </a>
        <span v-else class="dbar-btn dbar-call dbar-disabled"><i class="pi pi-phone" /> Call</span>
        <RouterLink class="dbar-btn dbar-checkin" :to="`/sales/my-customers/${detail.customer.id}`">
          <i class="pi pi-sign-in" /> Check in
        </RouterLink>
      </div>
    </template>
  </section>
</template>

<style scoped>
.detail-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: 5.5rem; }

/* ── Skeleton ────────────────────────────────────────────── */
.detail-skeleton { display: grid; gap: 0.85rem; }
.sk-header { display: flex; align-items: center; gap: 0.7rem; }
.sk-circle { width: 48px; height: 48px; border-radius: 50%; background: #e2e8f0; flex-shrink: 0; }
.sk-lines { flex: 1; display: flex; flex-direction: column; gap: 0.4rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w70 { width: 70%; }
.sk-line.w60 { width: 60%; }
.sk-line.w50 { width: 50%; }
.sk-line.w40 { width: 40%; }
.sk-line.w80 { width: 80%; }
.sk-map { height: 180px; border-radius: 12px; background: #e2e8f0; }

/* ── Empty ───────────────────────────────────────────────── */
.detail-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.detail-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.detail-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.detail-empty span { color: var(--text-muted); font-size: 0.8rem; max-width: 260px; }
.detail-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; }

/* ── Card ────────────────────────────────────────────────── */
.dcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.dcard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

/* Summary */
.dcard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.dcard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.dcard-avatar {
  width: 52px; height: 52px; display: grid; place-items: center; border-radius: 16px;
  background: linear-gradient(135deg, #059669, #047857); color: #fff; font-weight: 800;
  font-size: 1rem; flex-shrink: 0; box-shadow: 0 3px 10px rgba(5, 150, 105, 0.25);
}
.dcard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-identity .eyebrow { margin: 0; }
.dcard-identity h1 { margin: 0; font-size: 1.2rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }
.dcard-identity small { color: var(--text-secondary); font-size: 0.75rem; }
.dcard-codes { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.dcard-code-item { padding: 0.55rem 0.65rem; background: rgba(255,255,255,0.7); border-radius: 10px; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-code-item span { color: var(--text-muted); font-size: 0.55rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.dcard-code-item strong { font-size: 0.78rem; color: var(--text-primary); font-weight: 700; }
.dcard-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }

/* Map */
.dcard-map { height: 220px; border-radius: 12px; overflow: hidden; border: 1px solid var(--border-light); }

/* Rows */
.dcard-location-rows, .dcard-rows { display: grid; gap: 0.45rem; }
.dcard-row { display: flex; align-items: flex-start; gap: 0.55rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.dcard-row i { color: var(--text-muted); font-size: 0.72rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.1rem; }
.dcard-distance { color: var(--brand-blue); font-weight: 600; }

/* Contact */
.dcard-contact-list { display: grid; gap: 0.5rem; }
.dcard-contact { padding: 0.7rem; background: #f8fafc; border-radius: 12px; display: flex; flex-direction: column; gap: 0.45rem; }
.dcard-contact-info { display: flex; flex-direction: column; gap: 0.05rem; }
.dcard-contact-info strong { font-size: 0.82rem; color: var(--text-primary); }
.dcard-contact-info span { font-size: 0.72rem; color: var(--text-muted); }
.dcard-contact-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.dcard-action-sm {
  display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.6rem;
  border-radius: 8px; border: 1px solid var(--border-light); background: #fff;
  color: var(--brand-blue); font-size: 0.72rem; font-weight: 600; text-decoration: none;
  transition: background 0.15s ease;
}
.dcard-action-sm:hover { background: var(--brand-blue-50); }
.dcard-action-sm i { font-size: 0.6rem; }
.dcard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1rem 0; }

/* ── Bottom Action Bar ───────────────────────────────────── */
.detail-bottom-bar {
  position: fixed; bottom: 0; left: 0; right: 0; z-index: 40;
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.5rem;
  padding: 0.75rem 1rem; padding-bottom: calc(0.75rem + env(safe-area-inset-bottom));
  background: var(--surface-card); border-top: 1px solid var(--border-light);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.dbar-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.65rem 0; border-radius: 12px; border: none;
  font-size: 0.72rem; font-weight: 700; cursor: pointer;
  text-decoration: none; text-align: center; transition: all 0.15s ease;
}
.dbar-navigate { background: var(--brand-blue); color: #fff; }
.dbar-navigate:hover { background: #1d4ed8; }
.dbar-navigate:disabled { background: #cbd5e1; cursor: not-allowed; }
.dbar-call { background: #f0fdf4; color: #059669; border: 1px solid #a7f3d0; }
.dbar-call:hover { background: #dcfce7; }
.dbar-disabled { opacity: 0.45; cursor: not-allowed; pointer-events: none; }
.dbar-checkin { background: #eff6ff; color: var(--brand-blue); border: 1px solid #bfdbfe; }
.dbar-checkin:hover { background: #dbeafe; }

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 480px) {
  .detail-page { gap: 0.7rem; }
  .dcard { padding: 1rem; }
  .dcard-identity h1 { font-size: 1.05rem; }
  .dcard-map { height: 180px; }
}
</style>
