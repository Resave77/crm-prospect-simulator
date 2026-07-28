<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyCustomer, getMyCustomerPlaceDetails } from '../../../api/crm'
import type { CustomerDetail, PlaceDetails } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'
import { copyToClipboard } from '../../../utils/placeDetails'
import DataSourceBadge from '../../../components/sales/detail/DataSourceBadge.vue'
import { initials, formatErrorMessage, formatVisitDate, calcDuration } from '../../../utils/format'
import { priceLevelLabel, priceLevelSeverity, businessStatusLabel, businessStatusSeverity, stars } from '../../../utils/placeLabels'

const route = useRoute()
const detail = ref<CustomerDetail | null>(null)
const placeDetails = ref<PlaceDetails | null>(null)
const error = ref('')
const loading = ref(true)
const userCoords = ref<{ lat: number; lng: number } | null>(null)
let geoWatchId: number | null = null
const activePhotoIdx = ref(0)
const showAllHours = ref(false)

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
const showLegend = ref(false)
function handleCopy(text: string) {
  copyToClipboard(text)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

onMounted(async () => {
  acquireGPS()
  try {
    const customerId = String(route.params.id)
    const [cust, place] = await Promise.all([
      getMyCustomer(customerId),
      getMyCustomerPlaceDetails(customerId).catch(() => null),
    ])
    detail.value = cust
    placeDetails.value = place
  } catch (caught) {
    error.value = formatErrorMessage(caught) ?? 'Unable to load customer.'
  } finally { loading.value = false }
})

onBeforeUnmount(() => { if (geoWatchId != null) navigator.geolocation?.clearWatch(geoWatchId) })
</script>

<template>
  <section class="detail-page">
    <RouterLink class="back-link" to="/sales/my-customers"><i class="pi pi-arrow-left" /></RouterLink>

    <!-- Data Source Legend -->
    <div class="ds-legend">
      <button class="ds-legend-toggle" @click="showLegend = !showLegend">
        <i class="pi pi-info-circle" /> Data source legend
        <i :class="showLegend ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
      </button>
      <div v-if="showLegend" class="ds-legend-body">
        <div class="ds-legend-item"><DataSourceBadge source="google" /> <span>Data imported from Google Maps / Places</span></div>
        <div class="ds-legend-item"><DataSourceBadge source="manual" /> <span>Data entered by Admin during conversion</span></div>
        <div class="ds-legend-item"><DataSourceBadge source="system" /> <span>Generated automatically by CRM</span></div>
        <div class="ds-legend-item"><DataSourceBadge source="prospect" /> <span>Carried over from the source Prospect</span></div>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="detail-skeleton">
      <div class="sk-header"><div class="sk-circle" /><div class="sk-lines"><div class="sk-line w70" /><div class="sk-line w40" /></div></div>
      <div class="sk-gallery" />
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
            <h1>{{ customer?.name }} <DataSourceBadge source="google" /></h1>
            <small v-if="customer?.parentCompanyName">{{ customer.parentCompanyName }} <DataSourceBadge source="manual" /></small>
          </div>
        </div>
        <div class="dcard-codes">
          <div class="dcard-code-item"><span>Customer code <DataSourceBadge source="system" /></span><strong>{{ customer?.customerCode }}</strong></div>
          <div class="dcard-code-item"><span>Parent code <DataSourceBadge source="system" /></span><strong>{{ customer?.parentCode }}</strong></div>
        </div>
        <div class="dcard-tags">
          <Tag value="ACTIVE" severity="success" /> <DataSourceBadge source="system" />
          <Tag v-if="customer?.segment" :value="customer.segment" /> <DataSourceBadge source="manual" />
          <Tag v-if="customer?.category" :value="customer.category" severity="secondary" /> <DataSourceBadge source="google" />
          <template v-if="placeDetails">
            <Tag v-if="placeDetails.businessStatus" :value="businessStatusLabel(placeDetails.businessStatus)" :severity="businessStatusSeverity(placeDetails.businessStatus)" />
            <Tag v-if="placeDetails.priceLevel" :value="priceLevelLabel(placeDetails.priceLevel)" :severity="priceLevelSeverity(placeDetails.priceLevel)" />
          </template>
        </div>
      </div>

      <!-- Google Place Info Card -->
      <div v-if="placeDetails" class="dcard dcard-google-info">
        <h2>Google Maps Info <DataSourceBadge source="google" /></h2>

        <!-- Editorial Summary -->
        <p v-if="placeDetails.editorialSummary" class="dcard-editorial">{{ placeDetails.editorialSummary }}</p>

        <!-- Rating + Website Row -->
        <div class="dcard-info-grid">
          <div v-if="placeDetails.rating > 0" class="dcard-info-item">
            <div class="dcard-rating">
              <span class="dcard-rating-num">{{ placeDetails.rating.toFixed(1) }}</span>
              <div class="dcard-stars">
                <i v-for="(s, i) in stars(placeDetails.rating)" :key="i" :class="['pi', s]" />
              </div>
              <span class="dcard-rating-count">({{ placeDetails.userRatingCount.toLocaleString() }} reviews)</span>
            </div>
          </div>
          <div v-if="placeDetails.websiteUrl" class="dcard-info-item">
            <a :href="placeDetails.websiteUrl" target="_blank" rel="noopener" class="dcard-link">
              <i class="pi pi-external-link" /> Website
            </a>
          </div>
          <div v-if="placeDetails.googleMapsUrl" class="dcard-info-item">
            <a :href="placeDetails.googleMapsUrl" target="_blank" rel="noopener" class="dcard-link">
              <i class="pi pi-map" /> View on Google Maps
            </a>
          </div>
          <div v-if="placeDetails.phoneNumber" class="dcard-info-item">
            <a :href="`tel:${placeDetails.phoneNumber}`" class="dcard-link">
              <i class="pi pi-phone" /> {{ placeDetails.phoneNumber }}
            </a>
            <span v-if="placeDetails.internationalPhone && placeDetails.internationalPhone !== placeDetails.phoneNumber" class="dcard-intl-phone">
              International: {{ placeDetails.internationalPhone }}
            </span>
          </div>
        </div>

        <!-- Place Types -->
        <div v-if="placeDetails.placeTypes?.length" class="dcard-types">
          <span class="dcard-types-label">Categories:</span>
          <Tag v-for="t in placeDetails.placeTypes.slice(0, 6)" :key="t" :value="t.replace(/_/g, ' ')" severity="secondary" class="dcard-type-tag" />
        </div>
      </div>

      <!-- Photo Gallery -->
      <div v-if="placeDetails?.photos?.length" class="dcard dcard-photos">
        <h2>Photos <DataSourceBadge source="google" /></h2>
        <div class="dcard-photo-scroll">
          <div
            v-for="(photo, idx) in placeDetails.photos"
            :key="photo.name"
            class="dcard-photo-item"
            :class="{ active: idx === activePhotoIdx }"
            @click="activePhotoIdx = idx"
          >
            <img :src="photo.photoUrl" :alt="`Photo ${idx + 1}`" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
          </div>
        </div>
        <div v-if="placeDetails.photos[activePhotoIdx]?.attribution" class="dcard-photo-attribution">
          Photo: {{ placeDetails.photos[activePhotoIdx].attribution }}
        </div>
      </div>

      <!-- Opening Hours Card -->
      <div v-if="placeDetails?.openingHours" class="dcard dcard-hours">
        <h2>Opening Hours <DataSourceBadge source="google" /></h2>
        <div class="dcard-hours-status">
          <span :class="['dcard-hours-dot', placeDetails.openingHours.openNow ? 'open' : 'closed']" />
          <strong>{{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}</strong>
        </div>
        <div v-if="placeDetails.openingHours.weekdays?.length" class="dcard-hours-list">
          <div
            v-for="(day, i) in (showAllHours ? placeDetails.openingHours.weekdays : placeDetails.openingHours.weekdays.slice(0, 3))"
            :key="i"
            class="dcard-hours-row"
            v-html="day"
          />
          <button v-if="placeDetails.openingHours.weekdays.length > 3" class="dcard-hours-toggle" @click="showAllHours = !showAllHours">
            {{ showAllHours ? 'Show less' : `Show all ${placeDetails.openingHours.weekdays.length} days` }}
          </button>
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard">
        <div class="dcard-header-row">
          <h2>Location <DataSourceBadge source="google" /></h2>
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
          <div v-if="displayContactName" class="dcard-row"><i class="pi pi-user" /><span><strong>{{ displayContactName }}</strong><template v-if="displayContactPosition"> · {{ displayContactPosition }}</template> <DataSourceBadge source="manual" /></span></div>
          <div v-if="displayPhone" class="dcard-row"><i class="pi pi-phone" /><a :href="`tel:${displayPhone}`">{{ displayPhone }}</a> <DataSourceBadge source="google" /></div>
          <div v-if="displayEmail" class="dcard-row"><i class="pi pi-envelope" /><a :href="`mailto:${displayEmail}`">{{ displayEmail }}</a> <DataSourceBadge source="manual" /></div>
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

      <!-- Reviews -->
      <div v-if="placeDetails?.reviews?.length" class="dcard dcard-reviews">
        <h2>Reviews <DataSourceBadge source="google" /></h2>
        <div class="dcard-reviews-list">
          <div v-for="(review, i) in placeDetails.reviews.slice(0, 5)" :key="i" class="dcard-review">
            <div class="dcard-review-header">
              <img v-if="review.authorPhoto" :src="review.authorPhoto" class="dcard-review-avatar" :alt="review.authorName" @error="($event.target as HTMLImageElement).style.display='none'" />
              <div v-else class="dcard-review-avatar-placeholder">{{ review.authorName?.charAt(0) || '?' }}</div>
              <div class="dcard-review-meta">
                <strong>{{ review.authorName }}</strong>
                <div class="dcard-review-stars">
                  <i v-for="(s, j) in stars(review.rating)" :key="j" :class="['pi', s]" />
                  <span class="dcard-review-time">{{ review.time }}</span>
                </div>
              </div>
            </div>
            <p v-if="review.text" class="dcard-review-text">{{ review.text }}</p>
          </div>
        </div>
      </div>

      <!-- Service Options -->
      <div v-if="placeDetails && (placeDetails.delivery || placeDetails.dineIn || placeDetails.takeout || placeDetails.curbsidePickup)" class="dcard dcard-services">
        <h2>Service Options <DataSourceBadge source="google" /></h2>
        <div class="dcard-service-tags">
          <Tag v-if="placeDetails.dineIn" value="Dine In" severity="success" />
          <Tag v-if="placeDetails.takeout" value="Takeout" severity="info" />
          <Tag v-if="placeDetails.delivery" value="Delivery" severity="warn" />
          <Tag v-if="placeDetails.curbsidePickup" value="Curbside Pickup" severity="secondary" />
        </div>
      </div>

      <!-- Parking & Payment & Accessibility -->
      <div v-if="(placeDetails?.parkingOptions) || (placeDetails?.paymentOptions) || (placeDetails?.accessibilityOptions)" class="dcard dcard-services">
        <h2>Amenities <DataSourceBadge source="google" /></h2>
        <div class="dcard-amenities-grid">
          <div v-if="placeDetails?.parkingOptions" class="dcard-amenity-section">
            <strong><i class="pi pi-directions" /> Parking</strong>
            <div class="dcard-amenity-list">
              <span v-if="placeDetails.parkingOptions.freeParkingLot" class="dcard-amenity-chip"><i class="pi pi-check" /> Free Lot</span>
              <span v-if="placeDetails.parkingOptions.freeStreetParking" class="dcard-amenity-chip"><i class="pi pi-check" /> Free Street</span>
              <span v-if="placeDetails.parkingOptions.paidParkingLot" class="dcard-amenity-chip"><i class="pi pi-check" /> Paid Lot</span>
              <span v-if="placeDetails.parkingOptions.paidStreetParking" class="dcard-amenity-chip"><i class="pi pi-check" /> Paid Street</span>
              <span v-if="placeDetails.parkingOptions.garageParking" class="dcard-amenity-chip"><i class="pi pi-check" /> Garage</span>
              <span v-if="placeDetails.parkingOptions.valetParking" class="dcard-amenity-chip"><i class="pi pi-check" /> Valet</span>
            </div>
          </div>
          <div v-if="placeDetails?.paymentOptions && (placeDetails.paymentOptions.cashOnly || placeDetails.paymentOptions.creditCardOnly || placeDetails.paymentOptions.debitCardOnly || placeDetails.paymentOptions.nfcOnly)" class="dcard-amenity-section">
            <strong><i class="pi pi-wallet" /> Payment</strong>
            <div class="dcard-amenity-list">
              <span v-if="placeDetails.paymentOptions.cashOnly" class="dcard-amenity-chip"><i class="pi pi-check" /> Cash</span>
              <span v-if="placeDetails.paymentOptions.creditCardOnly" class="dcard-amenity-chip"><i class="pi pi-check" /> Credit Card</span>
              <span v-if="placeDetails.paymentOptions.debitCardOnly" class="dcard-amenity-chip"><i class="pi pi-check" /> Debit Card</span>
              <span v-if="placeDetails.paymentOptions.nfcOnly" class="dcard-amenity-chip"><i class="pi pi-check" /> NFC</span>
            </div>
          </div>
          <div v-if="placeDetails?.accessibilityOptions" class="dcard-amenity-section">
            <strong><i class="pi pi-verified" /> Accessibility</strong>
            <div class="dcard-amenity-list">
              <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleEntrance" class="dcard-amenity-chip"><i class="pi pi-check" /> Wheelchair Entrance</span>
              <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleParking" class="dcard-amenity-chip"><i class="pi pi-check" /> Wheelchair Parking</span>
              <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleRestroom" class="dcard-amenity-chip"><i class="pi pi-check" /> Wheelchair Restroom</span>
              <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleSeating" class="dcard-amenity-chip"><i class="pi pi-check" /> Wheelchair Seating</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Conversion Source -->
      <div class="dcard">
        <h2>Conversion Source</h2>
        <div class="dcard-rows">
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Prospect:</strong> {{ detail.sourceProspectName }} <DataSourceBadge source="prospect" /></span></div>
          <div class="dcard-row"><i class="pi pi-id-card" /><span><strong>Source ID:</strong> {{ customer?.sourceProspectId || '—' }} <DataSourceBadge source="prospect" /></span></div>
          <div v-if="customer?.sourceGooglePlaceId" class="dcard-row">
            <i class="pi pi-info-circle" />
            <span class="dcard-place-id"><span>Google Place ID <DataSourceBadge source="google" /></span><code>{{ customer.sourceGooglePlaceId }}</code></span>
            <button class="dcard-copy-btn" title="Copy Place ID" @click="handleCopy(customer.sourceGooglePlaceId)"><i class="pi pi-copy" /></button>
          </div>
          <div class="dcard-row"><i class="pi pi-calendar" /><span><strong>Converted:</strong> {{ customer?.convertedAt ? new Date(customer.convertedAt).toLocaleDateString('en', { year: 'numeric', month: 'short', day: 'numeric' }) : '—' }} <DataSourceBadge source="system" /></span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Sales Executive:</strong> {{ customer?.salesExecutiveName }} <DataSourceBadge source="system" /></span></div>
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
.sk-gallery { height: 160px; border-radius: var(--radius-xl); background: #e2e8f0; }

/* ── Empty ───────────────────────────────────────────────── */
.detail-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.detail-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.detail-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.detail-empty span { color: var(--text-muted); font-size: 0.8rem; max-width: 260px; }
.detail-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; }

/* ── Card ────────────────────────────────────────────────── */
.dcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem; min-width: 0;
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
.dcard-identity h1 { margin: 0; font-size: 1.2rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }
.dcard-identity small { color: var(--text-secondary); font-size: 0.75rem; display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.dcard-codes { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.dcard-code-item { padding: 0.55rem 0.65rem; background: rgba(255,255,255,0.7); border-radius: 10px; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-code-item span { color: var(--text-muted); font-size: 0.55rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.dcard-code-item strong { font-size: 0.78rem; color: var(--text-primary); font-weight: 700; }
.dcard-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }

/* Google Info */
.dcard-google-info { border: 1px solid #e0e7ff; background: linear-gradient(135deg, #f5f3ff 0%, var(--surface-card) 100%); }
.dcard-editorial { margin: 0; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.55; font-style: italic; }
.dcard-info-grid { display: grid; gap: 0.4rem; }
.dcard-info-item { display: flex; align-items: center; gap: 0.5rem; }
.dcard-link { display: inline-flex; align-items: center; gap: 0.3rem; color: var(--brand-blue); text-decoration: none; font-size: 0.8rem; font-weight: 600; }
.dcard-link:hover { text-decoration: underline; }
.dcard-intl-phone { color: var(--text-muted); font-size: 0.72rem; }
.dcard-rating { display: flex; align-items: center; gap: 0.35rem; }
.dcard-rating-num { font-size: 0.95rem; font-weight: 800; color: #f59e0b; }
.dcard-stars { display: flex; gap: 1px; }
.dcard-stars .pi { font-size: 0.6rem; color: #f59e0b; }
.dcard-rating-count { color: var(--text-muted); font-size: 0.72rem; }
.dcard-types { display: flex; flex-wrap: wrap; align-items: center; gap: 0.3rem; }
.dcard-types-label { color: var(--text-muted); font-size: 0.68rem; font-weight: 600; margin-right: 0.2rem; }
.dcard-type-tag { font-size: 0.62rem !important; }

/* Photo Gallery */
.dcard-photos { border: 1px solid #e0e7ff; }
.dcard-photo-scroll {
  display: flex; gap: 0.5rem; overflow-x: auto; scroll-snap-type: x mandatory;
  -webkit-overflow-scrolling: touch; padding-bottom: 0.3rem;
}
.dcard-photo-scroll::-webkit-scrollbar { height: 4px; }
.dcard-photo-scroll::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; }
.dcard-photo-item {
  flex: 0 0 200px; height: 150px; border-radius: 12px; overflow: hidden;
  cursor: pointer; scroll-snap-align: start; border: 2px solid transparent;
  transition: border-color 0.15s ease, transform 0.15s ease;
}
.dcard-photo-item:hover { transform: scale(1.02); }
.dcard-photo-item.active { border-color: var(--brand-blue); }
.dcard-photo-item img { width: 100%; height: 100%; object-fit: cover; }
.dcard-photo-attribution { color: var(--text-muted); font-size: 0.62rem; font-style: italic; }

/* Opening Hours */
.dcard-hours { border: 1px solid #fef3c7; background: linear-gradient(135deg, #fffbeb 0%, var(--surface-card) 100%); }
.dcard-hours-status { display: flex; align-items: center; gap: 0.4rem; }
.dcard-hours-dot { width: 8px; height: 8px; border-radius: 50%; }
.dcard-hours-dot.open { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.dcard-hours-dot.closed { background: #ef4444; }
.dcard-hours-list { display: grid; gap: 0.3rem; }
.dcard-hours-row { font-size: 0.78rem; color: var(--text-secondary); line-height: 1.4; }
.dcard-hours-toggle {
  background: none; border: none; color: var(--brand-blue); font-size: 0.72rem; font-weight: 600;
  cursor: pointer; padding: 0.2rem 0; text-align: left;
}
.dcard-hours-toggle:hover { text-decoration: underline; }

/* Reviews */
.dcard-reviews { border: 1px solid #e0e7ff; }
.dcard-reviews-list { display: grid; gap: 0.75rem; }
.dcard-review { padding-bottom: 0.65rem; border-bottom: 1px solid var(--border-light); }
.dcard-review:last-child { border-bottom: none; padding-bottom: 0; }
.dcard-review-header { display: flex; align-items: center; gap: 0.6rem; }
.dcard-review-avatar { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; }
.dcard-review-avatar-placeholder {
  width: 32px; height: 32px; border-radius: 50%; display: grid; place-items: center;
  background: #e2e8f0; color: var(--text-muted); font-size: 0.72rem; font-weight: 700; flex-shrink: 0;
}
.dcard-review-meta { flex: 1; min-width: 0; }
.dcard-review-meta strong { font-size: 0.78rem; color: var(--text-primary); }
.dcard-review-stars { display: flex; align-items: center; gap: 1px; }
.dcard-review-stars .pi { font-size: 0.55rem; color: #f59e0b; }
.dcard-review-time { color: var(--text-muted); font-size: 0.65rem; margin-left: 0.4rem; }
.dcard-review-text { margin: 0.3rem 0 0; color: var(--text-secondary); font-size: 0.78rem; line-height: 1.5; }

/* Services & Amenities */
.dcard-services { border: 1px solid #fef3c7; background: linear-gradient(135deg, #fffbeb 0%, var(--surface-card) 100%); }
.dcard-service-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.dcard-amenities-grid { display: grid; gap: 0.65rem; }
.dcard-amenity-section { display: flex; flex-direction: column; gap: 0.3rem; }
.dcard-amenity-section strong { display: flex; align-items: center; gap: 0.4rem; font-size: 0.75rem; color: var(--text-primary); }
.dcard-amenity-section strong i { color: var(--brand-blue); font-size: 0.7rem; }
.dcard-amenity-list { display: flex; flex-wrap: wrap; gap: 0.3rem; }
.dcard-amenity-chip {
  display: inline-flex; align-items: center; gap: 0.2rem;
  padding: 0.2rem 0.5rem; border-radius: 9999px;
  background: #f0fdf4; color: #059669;
  font-size: 0.62rem; font-weight: 600;
}
.dcard-amenity-chip i { font-size: 0.5rem; }

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

/* Place ID */
.dcard-place-id { flex: 1; display: flex; flex-direction: column; gap: 0.15rem; }
.dcard-place-id span { color: var(--text-muted); font-size: 0.62rem; font-weight: 600; display: flex; align-items: center; gap: 0.3rem; }
.dcard-place-id code { font-size: 0.72rem; color: var(--text-primary); background: #f1f5f9; padding: 0.2rem 0.4rem; border-radius: 6px; word-break: break-all; }
.dcard-copy-btn { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0.2rem; border-radius: 6px; transition: all 0.15s ease; }
.dcard-copy-btn:hover { color: var(--brand-blue); background: #eff6ff; }

/* ── Bottom Action Bar ───────────────────────────────────── */
.detail-bottom-bar {
  position: fixed; bottom: 0; left: 0; right: 0;
  width: 100%; z-index: 40;
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

/* ── Data Source Legend ─────────────────────────────────────── */
.ds-legend {
  border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-xs); overflow: hidden;
}
.ds-legend-toggle {
  width: 100%; display: flex; align-items: center; gap: 0.5rem;
  padding: 0.65rem 1rem; border: 0; background: transparent;
  color: var(--text-muted); font-size: 0.72rem; font-weight: 600;
  cursor: pointer; transition: color 0.15s ease;
}
.ds-legend-toggle:hover { color: var(--text-primary); }
.ds-legend-toggle i:first-child { font-size: 0.8rem; color: var(--brand-blue); }
.ds-legend-toggle i:last-child { margin-left: auto; font-size: 0.6rem; }
.ds-legend-body {
  display: grid; gap: 0.45rem; padding: 0 1rem 0.75rem;
}
.ds-legend-item {
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 0.72rem; color: var(--text-secondary); line-height: 1.4;
}

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 767px) {
  .detail-page { gap: 0.7rem; }
  .dcard { padding: 1rem; }
  .dcard-identity h1 { font-size: 1.05rem; }
  .dcard-photo-item { flex: 0 0 160px; height: 120px; }
}
</style>
