<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyCustomer } from '../../../api/crm'
import type { CustomerDetail } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'
import { copyToClipboard } from '../../../utils/placeDetails'

const route = useRoute()
const detail = ref<CustomerDetail | null>(null)
const error = ref('')
const loading = ref(true)
const userCoords = ref<{ lat: number; lng: number } | null>(null)
let geoWatchId: number | null = null

const customer = computed(() => detail.value?.customer)
const parentCompany = computed(() => detail.value?.parentCompany)

const displayPhone = computed(() => {
  return customer.value?.contacts?.[0]?.phone ?? ''
})

const displayEmail = computed(() => {
  return customer.value?.contacts?.[0]?.email ?? ''
})

const displayContactName = computed(() => {
  return customer.value?.contacts?.[0]?.name ?? ''
})

const displayContactPosition = computed(() => {
  return customer.value?.contacts?.[0]?.position ?? ''
})

const hasCoords = computed(() => {
  return customer.value?.address?.latitude != null && customer.value?.address?.longitude != null
})

const distance = computed(() => {
  if (!hasCoords.value || !userCoords.value) return null
  return getDistanceTo(
    customer.value!.address.latitude!,
    customer.value!.address.longitude!,
    userCoords.value.lat,
    userCoords.value.lng,
  )
})

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function navigate() {
  if (!customer.value) return
  openGoogleMapsNavigation({
    latitude: customer.value.address?.latitude,
    longitude: customer.value.address?.longitude,
    address: customer.value.address?.previewAddress,
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

const copied = ref(false)
function handleCopy(text: string) {
  copyToClipboard(text)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

onMounted(async () => {
  acquireGPS()
  try {
    detail.value = await getMyCustomer(String(route.params.id))
  } catch (caught) {
    error.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ?? 'Unable to load customer.'
  } finally { loading.value = false }
})

onBeforeUnmount(() => { if (geoWatchId != null) navigator.geolocation?.clearWatch(geoWatchId) })
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
          <div class="dcard-avatar">{{ initials(customer?.name || 'Customer') }}</div>
          <div class="dcard-identity">
            <p class="eyebrow">Customer Existing</p>
            <h1>{{ customer?.name }}</h1>
            <small v-if="customer?.parentCompanyName">{{ customer.parentCompanyName }}</small>
          </div>
        </div>
        <div class="dcard-codes">
          <div class="dcard-code-item"><span>Customer code</span><strong>{{ customer?.customerCode }}</strong></div>
          <div class="dcard-code-item"><span>Parent code</span><strong>{{ customer?.parentCode }}</strong></div>
        </div>
        <div class="dcard-tags">
          <Tag value="ACTIVE" severity="success" />
          <Tag v-if="customer?.segment" :value="customer.segment" />
          <Tag v-if="customer?.category" :value="customer.category" severity="secondary" />
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard">
        <div class="dcard-header-row">
          <h2>Location</h2>
          <span v-if="distance != null" class="dcard-distance-pill">
            <i class="pi pi-compass" /> {{ formatDistance(distance) }} away
          </span>
        </div>
        <EntityLocationMap
          :latitude="customer?.address?.latitude ?? null"
          :longitude="customer?.address?.longitude ?? null"
          :label="customer?.name"
          :interactive="true"
          height="200px"
        />
        <div class="dcard-location-rows">
          <div class="dcard-row"><i class="pi pi-map-marker" /><span>{{ customer?.address?.previewAddress || 'No address' }}</span></div>
          <div v-if="customer?.region" class="dcard-row"><i class="pi pi-globe" /><span>Region: {{ customer.region }}</span></div>
          <div v-if="hasCoords" class="dcard-row dcard-row-coords">
            <i class="pi pi-compass" /><span>GPS: {{ customer?.address?.latitude?.toFixed(6) }}, {{ customer?.address?.longitude?.toFixed(6) }}</span>
          </div>
        </div>
      </div>

      <!-- Contact Card -->
      <div class="dcard">
        <h2>Contact & Address</h2>
        <div v-if="displayContactName || displayPhone || displayEmail" class="dcard-contact-section">
          <div v-if="displayContactName" class="dcard-row"><i class="pi pi-user" /><span><strong>{{ displayContactName }}</strong><template v-if="displayContactPosition"> · {{ displayContactPosition }}</template></span></div>
          <div v-if="displayPhone" class="dcard-row"><i class="pi pi-phone" /><a :href="`tel:${displayPhone}`">{{ displayPhone }}</a></div>
          <div v-if="displayEmail" class="dcard-row"><i class="pi pi-envelope" /><a :href="`mailto:${displayEmail}`">{{ displayEmail }}</a></div>
        </div>
        <p v-else class="dcard-empty-text">No contacts on file.</p>
        <div v-if="customer?.address?.previewAddress" class="dcard-address-block">
          <div class="dcard-address-label">Full Address</div>
          <p>{{ customer.address.previewAddress }}</p>
          <div v-if="customer.address.province || customer.address.district" class="dcard-address-detail">
            <span v-if="customer.address.village">{{ customer.address.village }}, </span>
            <span v-if="customer.address.subDistrict">{{ customer.address.subDistrict }}, </span>
            <span v-if="customer.address.district">{{ customer.address.district }}, </span>
            <span v-if="customer.address.province">{{ customer.address.province }}</span>
          </div>
        </div>
      </div>

      <!-- Conversion Source -->
      <div class="dcard">
        <h2>Conversion Source</h2>
        <div class="dcard-rows">
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Prospect:</strong> {{ detail.sourceProspectName }}</span></div>
          <div class="dcard-row"><i class="pi pi-id-card" /><span><strong>Source ID:</strong> {{ customer?.sourceProspectId || '—' }}</span></div>
          <div v-if="customer?.sourceGooglePlaceId" class="dcard-row">
            <i class="pi pi-info-circle" />
            <span class="dcard-place-id"><span>Google Place ID</span><code>{{ customer.sourceGooglePlaceId }}</code></span>
            <button class="dcard-copy-btn" title="Copy Place ID" @click="handleCopy(customer.sourceGooglePlaceId)"><i class="pi pi-copy" /></button>
          </div>
          <div class="dcard-row"><i class="pi pi-calendar" /><span><strong>Converted:</strong> {{ customer?.convertedAt ? new Date(customer.convertedAt).toLocaleDateString('en', { year: 'numeric', month: 'short', day: 'numeric' }) : '—' }}</span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Sales Executive:</strong> {{ customer?.salesExecutiveName }}</span></div>
        </div>
      </div>

      <!-- Bottom Action Bar -->
      <div class="detail-bottom-bar">
        <button class="dbar-btn dbar-navigate" :disabled="!hasCoords && !customer?.address?.previewAddress" @click="navigate">
          <i class="pi pi-directions" /> Navigate
        </button>
        <a v-if="displayPhone" :href="`tel:${displayPhone}`" class="dbar-btn dbar-call">
          <i class="pi pi-phone" /> Call
        </a>
        <span v-else class="dbar-btn dbar-call dbar-disabled"><i class="pi pi-phone" /> Call</span>
        <RouterLink class="dbar-btn dbar-checkin" :to="`/sales/my-customers/${customer?.id}/check-in`">
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

/* Header row */
.dcard-header-row { display: flex; align-items: center; justify-content: space-between; }
.dcard-header-row h2 { margin: 0; }
.dcard-distance-pill {
  display: inline-flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.55rem;
  border-radius: 9999px; background: #eff6ff; color: var(--brand-blue);
  font-size: 0.62rem; font-weight: 700; white-space: nowrap;
}

/* Rows */
.dcard-location-rows, .dcard-rows { display: grid; gap: 0.45rem; }
.dcard-row { display: flex; align-items: flex-start; gap: 0.55rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.dcard-row i { color: var(--text-muted); font-size: 0.72rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.1rem; }
.dcard-row a { color: var(--brand-blue); text-decoration: none; }
.dcard-row a:hover { text-decoration: underline; }
.dcard-distance { color: var(--brand-blue); font-weight: 600; }
.dcard-row-coords { color: var(--text-muted); font-size: 0.75rem; }

/* Contact */
.dcard-contact-section { display: grid; gap: 0.45rem; }
.dcard-address-block {
  padding: 0.75rem; background: #f8fafc; border-radius: 12px;
  border: 1px solid var(--border-light); margin-top: 0.25rem;
}
.dcard-address-label { color: var(--text-muted); font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.25rem; }
.dcard-address-block p { margin: 0; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.5; }
.dcard-address-detail { margin-top: 0.2rem; color: var(--text-muted); font-size: 0.75rem; }

.dcard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1rem 0; }

/* ── Bottom Action Bar ───────────────────────────────────── */
.detail-bottom-bar {
  position: fixed; bottom: 0; left: 50%; transform: translateX(-50%);
  width: min(100%, 440px); z-index: 40;
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
}
</style>
