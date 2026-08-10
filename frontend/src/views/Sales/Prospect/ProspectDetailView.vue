<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyProspect, getProspectPlaceDetails } from '../../../api/crm'
import type { ProspectReview, PlaceDetails } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import ProspectComments from '../../../components/ProspectComments.vue'
import PlacePhotoGallery from '../../../components/PlacePhotoGallery.vue'
import DataSourceBadge from '../../../components/sales/detail/DataSourceBadge.vue'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'
import { formatPlaceType, isValidWebsite, websiteDisplayUrl, isValidPhone, copyToClipboard } from '../../../utils/placeDetails'
import { initials, formatErrorMessage, formatVisitDate, calcDuration } from '../../../utils/format'
import { priceLevelLabel, priceLevelSeverity, businessStatusLabel, businessStatusSeverity, stars, utcOffsetLabel } from '../../../utils/placeLabels'

const route = useRoute()
const review = ref<ProspectReview | null>(null)
const placeDetails = ref<PlaceDetails | null>(null)
const error = ref('')
const success = ref('')
const loading = ref(true)
const showAllHours = ref(false)
const showLegend = ref(false)
const showAllVisits = ref(false)
const showAllStatusHistory = ref(false)
const showAllReviews = ref(false)
const apiBase = import.meta.env.VITE_API_BASE_URL || ''

const userCoords = ref<{ lat: number; lng: number } | null>(null)
const isDesktop = ref(false)
const photoGalleryShell = ref<HTMLElement | null>(null)
let geoWatchId: number | null = null
let desktopQuery: MediaQueryList | null = null

const openVisit = computed(() => review.value?.visits.find((v) => !v.checkOutAt) ?? null)

const statusSeverity = computed(() => {
  const s = review.value?.prospect.status
  if (s === 'WON' || s === 'CONVERTED') return 'success' as const
  if (s === 'LOST') return 'danger' as const
  if (s === 'NEW_LEAD') return 'info' as const
  if (s === 'QUALIFIED') return 'success' as const
  return 'warn' as const
})

const displayTypes = computed(() => {
  const types = review.value?.prospect.placeTypes
  if (!Array.isArray(types) || !types.length) return []
  const skip = new Set(['establishment', 'point_of_interest', 'food', 'store'])
  return types.filter((t) => !skip.has(t)).slice(0, 5)
})

const displayedVisits = computed(() => {
  const items = review.value?.visits ?? []
  return showAllVisits.value ? items : items.slice(0, 2)
})

const displayedStatusHistory = computed(() => {
  const items = review.value?.history ?? []
  return showAllStatusHistory.value ? items : items.slice(0, 6)
})

const displayedReviews = computed(() => {
  const items = placeDetails.value?.reviews ?? []
  return showAllReviews.value ? items : items.slice(0, 4)
})

function syncDesktop(value: MediaQueryList | MediaQueryListEvent) {
  isDesktop.value = value.matches
}

function navigate() {
  const p = review.value?.prospect
  if (!p) return
  openGoogleMapsNavigation({
    latitude: p.latitude,
    longitude: p.longitude,
    address: p.formattedAddress,
    googleMapsUrl: p.googleMapsUrl,
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

function handleCopy(text: string) {
  copyToClipboard(text)
  success.value = 'Copied to clipboard.'
  setTimeout(() => { if (success.value === 'Copied to clipboard.') success.value = '' }, 2000)
}

function findPhotoScroller() {
  const root = photoGalleryShell.value
  if (!root) return null

  const candidates = [root, ...Array.from(root.querySelectorAll<HTMLElement>('*'))]

  return candidates.find((element) => {
    const style = window.getComputedStyle(element)
    const canOverflow = style.overflowX === 'auto' || style.overflowX === 'scroll'
    return canOverflow && element.scrollWidth > element.clientWidth + 8
  }) ?? candidates.find((element) => element.scrollWidth > element.clientWidth + 8) ?? null
}

function scrollPhotos(direction: -1 | 1) {
  const scroller = findPhotoScroller()
  if (!scroller) return

  const distance = Math.max(scroller.clientWidth * 0.8, 320)
  scroller.scrollBy({
    left: direction * distance,
    behavior: 'smooth',
  })
}

onMounted(async () => {
  desktopQuery = window.matchMedia('(min-width: 1024px)')
  syncDesktop(desktopQuery)
  desktopQuery.addEventListener('change', syncDesktop)
  acquireGPS()
  try {
    const prospectId = String(route.params.id)
    const [reviewData, placeData] = await Promise.all([
      getMyProspect(prospectId),
      getProspectPlaceDetails(prospectId, 'SALES_EXECUTIVE').catch(() => null),
    ])
    review.value = reviewData
    placeDetails.value = placeData
  } catch (caught) { error.value = formatErrorMessage(caught) } finally { loading.value = false }
})

onBeforeUnmount(() => {
  if (geoWatchId != null) navigator.geolocation?.clearWatch(geoWatchId)
  desktopQuery?.removeEventListener('change', syncDesktop)
})
</script>

<template>
  <section class="detail-page">
    <RouterLink class="back-link" to="/sales/my-prospects"><i class="pi pi-arrow-left" /></RouterLink>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <!-- Data Source Legend -->
    <div class="ds-legend">
      <button class="ds-legend-toggle" @click="showLegend = !showLegend">
        <i class="pi pi-info-circle" /> Data source legend
        <i :class="showLegend ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
      </button>
      <div v-if="showLegend" class="ds-legend-body">
        <div class="ds-legend-item"><DataSourceBadge source="google" /> <span>Data imported from Google Maps / Places</span></div>
        <div class="ds-legend-item"><DataSourceBadge source="manual" /> <span>Data entered or updated by the sales/admin team</span></div>
        <div class="ds-legend-item"><DataSourceBadge source="system" /> <span>Generated automatically by CRM</span></div>
      </div>
    </div>

    <div v-if="loading" class="detail-skeleton">
      <div class="sk-header"><div class="sk-circle" /><div class="sk-lines"><div class="sk-line w70" /><div class="sk-line w40" /></div></div>
      <div class="sk-gallery" />
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /></div>
    </div>

    <div v-else-if="!review" class="detail-empty">
      <div class="detail-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>Prospect not found</strong>
      <span>This prospect may have been removed or you don't have access.</span>
      <RouterLink class="detail-empty-btn" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> Back to prospects</RouterLink>
    </div>

    <template v-else>
      <!-- Summary Card -->
      <div class="dcard dcard-summary">
        <div class="dcard-summary-top">
          <div class="dcard-avatar dcard-avatar-prospect">{{ initials(review.prospect.placeName || 'Prospect') }}</div>

          <div class="dcard-identity">
            <div class="dcard-title-line">
              <div>
                <p class="eyebrow">Prospect Detail</p>
                <h1>
                  {{ review.prospect.placeName || 'Unnamed Prospect' }}
                  <DataSourceBadge source="google" label="Google" />
                </h1>
              </div>
              <Tag
                class="summary-status"
                :value="review.prospect.status.replaceAll('_', ' ')"
                :severity="statusSeverity"
              />
            </div>

            <div v-if="review.prospect.formattedAddress" class="summary-address">
              <i class="pi pi-map-marker" />
              <span>{{ review.prospect.formattedAddress }}</span>
            </div>
          </div>
        </div>

        <div class="summary-meta-grid">
          <div class="summary-meta-item">
            <span class="summary-meta-label">
              <i class="pi pi-user" /> Assigned Sales
              <DataSourceBadge source="manual" label="Manual" />
            </span>
            <strong>{{ review.prospect.assignedSalesExecutive || 'Unassigned' }}</strong>
          </div>

          <div class="summary-meta-item">
            <span class="summary-meta-label">
              <i class="pi pi-phone" /> Phone
              <DataSourceBadge source="google" label="Google" />
            </span>
            <strong>{{ review.prospect.phoneNumber || 'Not provided' }}</strong>
          </div>

          <div class="summary-meta-item">
            <span class="summary-meta-label">
              <i class="pi pi-tag" /> Category
              <DataSourceBadge source="google" label="Google" />
            </span>
            <strong>{{ review.prospect.placeCategory || 'Not provided' }}</strong>
          </div>

          <div class="summary-meta-item">
            <span class="summary-meta-label">
              <i class="pi pi-briefcase" /> Industry
              <DataSourceBadge source="google" label="Google" />
            </span>
            <strong>{{ review.prospect.industryGroup || 'Not provided' }}</strong>
          </div>
        </div>

        <div class="dcard-codes">
          <div class="dcard-code-item">
            <span>Google Place ID <DataSourceBadge source="google" label="Google" /></span>
            <strong>{{ review.prospect.googlePlaceId || '-' }}</strong>
          </div>
          <div class="dcard-code-item">
            <span>Last Updated <DataSourceBadge source="system" label="CRM" /></span>
            <strong>{{ new Date(review.prospect.updatedAt).toLocaleString() }}</strong>
          </div>
        </div>

        <div class="dcard-tags">
          <Tag v-if="review.prospect.placeCategory" :value="review.prospect.placeCategory" severity="secondary" />
          <Tag v-if="review.prospect.industryGroup" :value="review.prospect.industryGroup" />
          <template v-if="placeDetails">
            <Tag
              v-if="placeDetails.businessStatus"
              :value="businessStatusLabel(placeDetails.businessStatus)"
              :severity="businessStatusSeverity(placeDetails.businessStatus)"
            />
            <Tag
              v-if="placeDetails.priceLevel"
              :value="priceLevelLabel(placeDetails.priceLevel)"
              :severity="priceLevelSeverity(placeDetails.priceLevel)"
            />
          </template>
        </div>

        <div v-if="displayTypes.length" class="dcard-type-badges">
          <span v-for="t in displayTypes" :key="t" class="dcard-type-badge">{{ formatPlaceType(t) }}</span>
        </div>
      </div>

      <!-- Active Visit Alert -->
      <div v-if="openVisit" class="dcard dcard-active-visit">
        <div class="dcard-active-visit-row">
          <i class="pi pi-sign-in" />
          <span>You have an <strong>active visit</strong> in progress.</span>
          <RouterLink class="dcard-active-visit-link" :to="`/sales/my-prospects/${review.prospect.id}/check-out`">
            Continue visit <i class="pi pi-arrow-right" />
          </RouterLink>
        </div>
      </div>

      <div class="detail-content-grid">
        <div class="detail-column detail-column-main">
      <!-- Google Maps Info Card -->
      <div v-if="placeDetails" class="dcard dcard-google-info">
        <h2>Google Maps Info <DataSourceBadge source="google" label="Google" /></h2>
        <p v-if="placeDetails.editorialSummary" class="dcard-editorial">{{ placeDetails.editorialSummary }}</p>
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
        <div v-if="placeDetails.placeTypes?.length" class="dcard-types">
          <span class="dcard-types-label">Categories:</span>
          <Tag v-for="t in placeDetails.placeTypes.slice(0, 6)" :key="t" :value="t.replace(/_/g, ' ')" severity="secondary" class="dcard-type-tag" />
        </div>
        <div class="dcard-rows">
          <div v-if="placeDetails.placeName" class="dcard-row">
            <i class="pi pi-building" />
            <span><strong>Google place:</strong> {{ placeDetails.placeName }}</span>
          </div>
          <div v-if="placeDetails.placeCategory" class="dcard-row">
            <i class="pi pi-tag" />
            <span><strong>Category:</strong> {{ placeDetails.placeCategory }}</span>
          </div>
          <div v-if="placeDetails.formattedAddress" class="dcard-row">
            <i class="pi pi-map-marker" />
            <span><strong>Google address:</strong> {{ placeDetails.formattedAddress }}</span>
          </div>
          <div v-if="placeDetails.internationalPhone && placeDetails.internationalPhone !== placeDetails.phoneNumber" class="dcard-row">
            <i class="pi pi-phone" />
            <span><strong>International phone:</strong> {{ placeDetails.internationalPhone }}</span>
          </div>
          <div v-if="placeDetails.googlePlaceId" class="dcard-row">
            <i class="pi pi-id-card" />
            <span class="dcard-place-id"><span>Google Place ID</span><code>{{ placeDetails.googlePlaceId }}</code></span>
            <button class="dcard-copy-btn" title="Copy Place ID" aria-label="Copy Place ID" @click="handleCopy(placeDetails.googlePlaceId)"><i class="pi pi-copy" /></button>
          </div>
          <div v-if="placeDetails.utcOffsetMinutes != null" class="dcard-row">
            <i class="pi pi-globe" />
            <span><strong>Time zone:</strong> {{ utcOffsetLabel(placeDetails.utcOffsetMinutes) }} ({{ placeDetails.utcOffsetMinutes >= 0 ? '+' : '' }}{{ placeDetails.utcOffsetMinutes }} min from UTC)</span>
          </div>
          <div v-if="placeDetails.latitude != null && placeDetails.longitude != null" class="dcard-row dcard-row-coords">
            <i class="pi pi-compass" />
            <span>Google GPS: {{ placeDetails.latitude.toFixed(6) }}, {{ placeDetails.longitude.toFixed(6) }}</span>
          </div>
        </div>
      </div>

      <!-- Location Card -->
      <div class="dcard dcard-location">
        <div class="dcard-header-row">
          <h2>Location <DataSourceBadge source="google" label="Google" /></h2>
          <span v-if="review.prospect.latitude != null && review.prospect.longitude != null && userCoords" class="dcard-distance-pill">
            <i class="pi pi-compass" /> {{ formatDistance(getDistanceTo(review.prospect.latitude, review.prospect.longitude, userCoords.lat, userCoords.lng)!) }} away
          </span>
        </div>
        <EntityLocationMap
          :latitude="review.prospect.latitude"
          :longitude="review.prospect.longitude"
          :label="review.prospect.placeName"
          :interactive="false"
          height="200px"
        />
        <div class="dcard-location-rows">
          <div class="dcard-row"><i class="pi pi-map-marker" /><span>{{ review.prospect.formattedAddress || 'No address' }}</span></div>
          <div v-if="review.prospect.latitude != null && review.prospect.longitude != null" class="dcard-row dcard-row-coords">
            <i class="pi pi-compass" />
            <span>GPS: {{ review.prospect.latitude?.toFixed(6) }}, {{ review.prospect.longitude?.toFixed(6) }}</span>
            <button class="dcard-copy-btn" title="Copy coordinates" aria-label="Copy coordinates" @click="handleCopy(`${review.prospect.latitude}, ${review.prospect.longitude}`)"><i class="pi pi-copy" /></button>
          </div>
          <a v-if="review.prospect.googleMapsUrl" :href="review.prospect.googleMapsUrl" target="_blank" rel="noopener noreferrer" class="dcard-row dcard-row-link">
            <i class="pi pi-external-link" /><span>Open in Google Maps</span>
          </a>
        </div>
      </div>

      <!-- Contact & Address -->
      <div class="dcard dcard-contact">
        <h2>Contact & Address</h2>
        <div class="dcard-rows">
          <div class="dcard-row"><i class="pi pi-building" /><span><strong>Place name:</strong> {{ review.prospect.placeName || 'Not provided' }} <DataSourceBadge source="google" label="Google" /></span></div>
          <div class="dcard-row"><i class="pi pi-phone" /><span><strong>Phone:</strong> <a v-if="review.prospect.phoneNumber" :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a><template v-else>Not provided</template> <DataSourceBadge source="google" label="Google" /></span></div>
          <div class="dcard-row"><i class="pi pi-globe" /><span><strong>Website:</strong> <a v-if="review.prospect.websiteUrl && isValidWebsite(review.prospect.websiteUrl)" :href="review.prospect.websiteUrl" target="_blank" rel="noopener">{{ websiteDisplayUrl(review.prospect.websiteUrl) }}</a><template v-else>Not provided</template> <DataSourceBadge source="google" label="Google" /></span></div>
          <div class="dcard-row"><i class="pi pi-map-marker" /><span><strong>Address:</strong> {{ review.prospect.formattedAddress || 'Not provided' }} <DataSourceBadge source="google" label="Google" /></span></div>
        </div>
      </div>

      <!-- Business Information -->
      <div class="dcard dcard-business">
        <h2>Business Information</h2>
        <div class="dcard-rows">
          <div class="dcard-row"><i class="pi pi-tag" /><span><strong>Category:</strong> {{ review.prospect.placeCategory || 'Not provided' }} <DataSourceBadge source="google" label="Google" /></span></div>
          <div class="dcard-row"><i class="pi pi-briefcase" /><span><strong>Industry group:</strong> {{ review.prospect.industryGroup || 'Not provided' }} <DataSourceBadge source="google" label="Google" /></span></div>
          <div class="dcard-row"><i class="pi pi-user" /><span><strong>Sales executive:</strong> {{ review.prospect.assignedSalesExecutive || 'Unassigned' }} <DataSourceBadge source="manual" label="Manual" /></span></div>
          <div class="dcard-row">
            <i class="pi pi-id-card" />
            <span class="dcard-place-id"><span>Google Place ID <DataSourceBadge source="google" label="Google" /></span><code>{{ review.prospect.googlePlaceId || 'Not provided' }}</code></span>
            <button v-if="review.prospect.googlePlaceId" class="dcard-copy-btn" title="Copy Place ID" aria-label="Copy Place ID" @click="handleCopy(review.prospect.googlePlaceId)"><i class="pi pi-copy" /></button>
          </div>
          <div class="dcard-row"><i class="pi pi-calendar-plus" /><span><strong>Created:</strong> {{ new Date(review.prospect.createdAt).toLocaleString() }} <DataSourceBadge source="system" label="CRM" /></span></div>
          <div class="dcard-row"><i class="pi pi-refresh" /><span><strong>Updated:</strong> {{ new Date(review.prospect.updatedAt).toLocaleString() }} <DataSourceBadge source="system" label="CRM" /></span></div>
        </div>
      </div>

        

      <!-- Visit History -->
      <div class="dcard dcard-visits history-card">
        <div class="history-card-header">
          <h2>Visit History</h2>
          <span class="history-count">{{ review.visits.length }}</span>
        </div>
        <div v-if="review.visits.length" class="dcard-visit-list">
          <div v-for="visit in displayedVisits" :key="visit.id" class="dcard-visit">
            <div class="dcard-visit-header">
              <Tag :value="visit.checkOutAt ? 'Completed' : 'Active'" :severity="visit.checkOutAt ? 'secondary' : 'success'" />
              <span>{{ formatVisitDate(visit.checkInAt) }}</span>
            </div>
            <div class="dcard-visit-body">
              <div class="dcard-visit-detail"><i class="pi pi-sign-in" /><span>Check-in: {{ visit.checkInLatitude.toFixed(4) }}, {{ visit.checkInLongitude.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="dcard-visit-detail"><i class="pi pi-sign-out" /><span>Check-out: {{ visit.checkOutLatitude?.toFixed(4) }}, {{ visit.checkOutLongitude?.toFixed(4) }}</span></div>
              <div v-if="visit.checkOutAt" class="dcard-visit-detail"><i class="pi pi-clock" /><span>Duration: {{ calcDuration(visit.checkInAt, visit.checkOutAt) }}</span></div>
              <div v-if="visit.visitNotes" class="dcard-visit-detail"><i class="pi pi-comment" /><span>{{ visit.visitNotes }}</span></div>
              <div v-if="visit.followUpNotes" class="dcard-visit-detail"><i class="pi pi-directions" /><span>Follow-up: {{ visit.followUpNotes }}</span></div>
              <div v-if="visit.selfieReference && visit.selfieReference !== 'SIMULATED_SELFIE_PLACEHOLDER'" class="dcard-visit-selfie">
                <img :src="visit.selfieReference.startsWith('/') ? `${apiBase}${visit.selfieReference}` : visit.selfieReference" alt="Visit selfie" />
              </div>
              <div class="dcard-visit-detail dcard-visit-exec"><i class="pi pi-user" /><span>{{ visit.salesExecutiveName }}</span></div>
            </div>
          </div>
        </div>
        <p v-else class="dcard-empty-text">No visits recorded yet.</p>
        <button
          v-if="review.visits.length > 2"
          type="button"
          class="section-toggle-btn"
          @click="showAllVisits = !showAllVisits"
        >
          <span>{{ showAllVisits ? 'Show less' : `Show all ${review.visits.length} visits` }}</span>
          <i :class="showAllVisits ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
        </button>
      </div>
      </div>

        <div class="detail-column detail-column-side">
      <!-- Photos (Menu vs Photo, taggable) -->
      <div v-if="placeDetails?.photos?.length" class="dcard dcard-photos">
        <div class="photo-gallery-header">
          <h2><i class="pi pi-images" /> Photos <DataSourceBadge source="google" label="Google" /></h2>
          <div class="photo-gallery-controls" aria-label="Photo gallery navigation">
            <button type="button" class="photo-scroll-btn" aria-label="Scroll photos left" @click="scrollPhotos(-1)">
              <i class="pi pi-chevron-left" />
            </button>
            <button type="button" class="photo-scroll-btn" aria-label="Scroll photos right" @click="scrollPhotos(1)">
              <i class="pi pi-chevron-right" />
            </button>
          </div>
        </div>
        <div ref="photoGalleryShell" class="photo-gallery-shell">
          <PlacePhotoGallery :photos="placeDetails.photos" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" />
        </div>
      </div>

      <!-- Opening Hours Card -->
      <div v-if="placeDetails?.openingHours" class="dcard dcard-hours">
        <h2><i class="pi pi-clock" /> Opening Hours <DataSourceBadge source="google" label="Google" /></h2>
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

      <!-- Service Options -->
      <div v-if="placeDetails && (placeDetails.delivery || placeDetails.dineIn || placeDetails.takeout || placeDetails.curbsidePickup)" class="dcard dcard-services dcard-service-options">
        <h2><i class="pi pi-shopping-bag" /> Service Options <DataSourceBadge source="google" label="Google" /></h2>
        <div class="dcard-service-tags">
          <Tag v-if="placeDetails.dineIn" value="Dine In" severity="success" />
          <Tag v-if="placeDetails.takeout" value="Takeout" severity="info" />
          <Tag v-if="placeDetails.delivery" value="Delivery" severity="warn" />
          <Tag v-if="placeDetails.curbsidePickup" value="Curbside Pickup" severity="secondary" />
        </div>
      </div>

      <!-- Amenities -->
      <div v-if="(placeDetails?.parkingOptions) || (placeDetails?.paymentOptions) || (placeDetails?.accessibilityOptions)" class="dcard dcard-services dcard-amenities">
        <h2><i class="pi pi-building" /> Amenities <DataSourceBadge source="google" label="Google" /></h2>
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

      <!-- Sales Notes -->
      <div v-if="review.prospect.visitNotes || review.prospect.followUpNotes" class="dcard dcard-sales-notes">
        <h2>Sales Notes</h2>
        <div class="dcard-rows">
          <div v-if="review.prospect.visitNotes" class="dcard-row"><i class="pi pi-comment" /><span><strong>Visit notes:</strong> {{ review.prospect.visitNotes }} <DataSourceBadge source="manual" label="Manual" /></span></div>
          <div v-if="review.prospect.followUpNotes" class="dcard-row"><i class="pi pi-directions" /><span><strong>Follow-up:</strong> {{ review.prospect.followUpNotes }} <DataSourceBadge source="manual" label="Manual" /></span></div>
        </div>
      </div>

      <!-- Status History -->
      <div class="dcard dcard-status-history history-card">
        <div class="history-card-header">
          <h2>Status History</h2>
          <span class="history-count">{{ review.history.length }}</span>
        </div>
        <div v-if="review.history.length" class="dcard-timeline">
          <div v-for="entry in displayedStatusHistory" :key="entry.id" class="dcard-timeline-entry">
            <div class="dcard-timeline-dot" />
            <div class="dcard-timeline-content">
              <strong>{{ (entry.fromStatus || 'Created').replaceAll('_', ' ') }} → {{ entry.toStatus.replaceAll('_', ' ') }}</strong>
              <span>{{ new Date(entry.createdAt).toLocaleString() }} · {{ entry.changedByName }}</span>
              <p v-if="entry.notes">{{ entry.notes }}</p>
            </div>
          </div>
        </div>
        <p v-else class="dcard-empty-text">No status changes recorded.</p>
        <button
          v-if="review.history.length > 6"
          type="button"
          class="section-toggle-btn"
          @click="showAllStatusHistory = !showAllStatusHistory"
        >
          <span>{{ showAllStatusHistory ? 'Show less' : `Show all ${review.history.length} changes` }}</span>
          <i :class="showAllStatusHistory ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
        </button>
      </div>
      

        </div>
      </div>

      <!-- Reviews -->
      <div class="detail-reviews-section">
      <div v-if="placeDetails?.reviews?.length" class="dcard dcard-reviews">
        <div class="dcard-header-row">
          <h2><i class="pi pi-comments" /> Reviews</h2>
          <span class="section-count">{{ placeDetails.reviews.length }}</span>
        </div>
        <div class="dcard-reviews-list">
          <div v-for="(item, i) in displayedReviews" :key="i" class="dcard-review">
            <div class="dcard-review-header">
              <img v-if="item.authorPhoto" :src="item.authorPhoto" class="dcard-review-avatar" :alt="item.authorName" @error="($event.target as HTMLImageElement).style.display='none'" />
              <div v-else class="dcard-review-avatar-placeholder">{{ item.authorName?.charAt(0) || '?' }}</div>
              <div class="dcard-review-meta">
                <strong>{{ item.authorName }}</strong>
                <div class="dcard-review-stars">
                  <i v-for="(s, j) in stars(item.rating)" :key="j" :class="['pi', s]" />
                  <span class="dcard-review-time">{{ item.time }}</span>
                </div>
              </div>
            </div>
            <p v-if="item.text" class="dcard-review-text">{{ item.text }}</p>
          </div>
        </div>
        <button
          v-if="placeDetails.reviews.length > 4"
          type="button"
          class="section-toggle-btn"
          @click="showAllReviews = !showAllReviews"
        >
          <span>{{ showAllReviews ? 'Show less' : `Show all ${placeDetails.reviews.length} reviews` }}</span>
          <i :class="showAllReviews ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
        </button>

      </div>
      </div>

      <!-- Comments / Ticketing -->
      <ProspectComments class="detail-comments" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" :embedded="isDesktop" />

      <!-- Bottom Action Bar -->
      <div class="detail-bottom-bar">
        <button class="dbar-btn dbar-navigate" :disabled="review.prospect.latitude == null && review.prospect.longitude == null && !review.prospect.formattedAddress" @click="navigate">
          <i class="pi pi-directions" /> Navigate
        </button>
        <a v-if="review.prospect.phoneNumber && isValidPhone(review.prospect.phoneNumber)" :href="`tel:${review.prospect.phoneNumber}`" class="dbar-btn dbar-call">
          <i class="pi pi-phone" /> Call
        </a>
        <span v-else class="dbar-btn dbar-call dbar-disabled"><i class="pi pi-phone" /> Call</span>
        <RouterLink v-if="openVisit" class="dbar-btn dbar-checkout" :to="`/sales/my-prospects/${review.prospect.id}/check-out`">
          <i class="pi pi-sign-out" /> Check out
        </RouterLink>
        <RouterLink v-else class="dbar-btn dbar-checkin" :to="`/sales/my-prospects/${review.prospect.id}/check-in`">
          <i class="pi pi-sign-in" /> Check in
        </RouterLink>
      </div>
    </template>
  </section>
</template>

<style scoped>
.detail-page { display: grid; gap: 0.6rem; width: 100%; padding-bottom: 5.5rem; align-content: start; }
.detail-content-grid,
.detail-column { display: contents; }


.detail-reviews-section {
  min-width: 0;
}

.detail-reviews-section > .dcard {
  margin: 0;
}

.history-card {
  min-width: 0;
  align-content: start;
}

.history-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.history-card-header h2 { margin: 0; }

.history-count {
  min-width: 28px;
  height: 24px;
  padding: 0 0.45rem;
  display: inline-grid;
  place-items: center;
  border-radius: 9999px;
  background: #eff6ff;
  color: var(--brand-blue);
  font-size: 0.65rem;
  font-weight: 800;
  line-height: 1;
}


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

.detail-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.detail-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.detail-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.detail-empty span { color: var(--text-muted); font-size: 0.8rem; max-width: 260px; }
.detail-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; }

.dcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem; min-width: 0;
}
.dcard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }
.dcard h2 { display: flex; align-items: center; gap: 0.35rem; flex-wrap: wrap; }


.dcard-summary {
  background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%);
  gap: 0.7rem;
}
.dcard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.dcard-title-line {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  min-width: 0;
}
.dcard-title-line > div { min-width: 0; }
.summary-status { flex-shrink: 0; }
.summary-address {
  margin-top: 0.35rem;
  display: flex;
  align-items: flex-start;
  gap: 0.4rem;
  color: var(--text-secondary);
  font-size: 0.76rem;
  line-height: 1.45;
}
.summary-address i {
  color: var(--brand-blue);
  font-size: 0.72rem;
  margin-top: 0.15rem;
  flex-shrink: 0;
}
.summary-meta-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.45rem;
}
.summary-meta-item {
  min-width: 0;
  padding: 0.5rem 0.65rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 11px;
  background: rgba(255, 255, 255, 0.78);
  display: grid;
  gap: 0.22rem;
}
.summary-meta-label {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  flex-wrap: wrap;
  color: var(--text-muted);
  font-size: 0.57rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.summary-meta-label > i {
  color: var(--brand-blue);
  font-size: 0.64rem;
}
.summary-meta-item strong {
  color: var(--text-primary);
  font-size: 0.76rem;
  font-weight: 700;
  line-height: 1.35;
  overflow-wrap: anywhere;
}
.dcard-avatar { width: 52px; height: 52px; display: grid; place-items: center; border-radius: 16px; color: #fff; font-weight: 800; font-size: 1rem; flex-shrink: 0; }
.dcard-avatar-prospect { background: linear-gradient(135deg, #2563eb, #1d4ed8); box-shadow: 0 3px 10px rgba(37, 99, 235, 0.25); }
.dcard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.dcard-identity .eyebrow { margin: 0; }
.dcard-identity h1 { margin: 0; font-size: 1.2rem; font-weight: 800; letter-spacing: 0; color: var(--text-primary); line-height: 1.3; display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }

.dcard-codes { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.dcard-code-item { padding: 0.55rem 0.65rem; background: rgba(255,255,255,0.7); border-radius: 10px; display: flex; flex-direction: column; gap: 0.1rem; min-width: 0; }
.dcard-code-item span { color: var(--text-muted); font-size: 0.55rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.dcard-code-item strong { font-size: 0.78rem; color: var(--text-primary); font-weight: 700; overflow-wrap: anywhere; }

.dcard-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }

.dcard-type-badges { display: flex; flex-wrap: wrap; gap: 0.3rem; }
.dcard-type-badge {
  display: inline-block; padding: 0.15rem 0.5rem; border-radius: 6px;
  background: #eff6ff; color: #1d4ed8; font-size: 0.58rem; font-weight: 600; line-height: 1.5;
}

.dcard-active-visit { border-color: #fbbf24; background: #fffbeb; }
.dcard-active-visit-row { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8rem; color: #92400e; flex-wrap: wrap; }
.dcard-active-visit-row > i { font-size: 1rem; color: #f59e0b; }
.dcard-active-visit-link {
  display: inline-flex; align-items: center; gap: 0.25rem; margin-left: auto;
  padding: 0.35rem 0.75rem; border-radius: 10px; background: #f59e0b; color: #fff;
  font-size: 0.7rem; font-weight: 700; text-decoration: none; white-space: nowrap;
}
.dcard-active-visit-link:hover { background: #d97706; }

/* Google Info */
.dcard-google-info { border: 1px solid #e0e7ff; background: linear-gradient(135deg, #f5f3ff 0%, var(--surface-card) 100%); }
.dcard-editorial { margin: 0; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.55; font-style: italic; }
.dcard-info-grid { display: grid; gap: 0.4rem; }
.dcard-info-item { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
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
.dcard-photos h2 { display: flex; align-items: center; gap: 0.4rem; }
.dcard-photos h2 i { color: var(--brand-blue); font-size: 0.75rem; }

.photo-gallery-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.photo-gallery-controls {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  flex-shrink: 0;
}

.photo-scroll-btn {
  width: 34px;
  height: 34px;
  display: inline-grid;
  place-items: center;
  border: 1px solid var(--border-light);
  border-radius: 10px;
  background: #fff;
  color: var(--text-secondary);
  cursor: pointer;
  box-shadow: var(--shadow-xs);
  transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease, transform 0.15s ease;
}

.photo-scroll-btn:hover {
  color: var(--brand-blue);
  border-color: #bfdbfe;
  background: #eff6ff;
  transform: translateY(-1px);
}

.photo-scroll-btn:active { transform: translateY(0); }
.photo-scroll-btn i { font-size: 0.72rem; line-height: 1; }
.photo-gallery-shell { min-width: 0; overflow: hidden; }
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
.dcard-menu-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  padding: 2rem 1rem; color: var(--text-muted); text-align: center;
  background: var(--surface-subtle); border-radius: 12px;
}
.dcard-menu-empty i { font-size: 1.5rem; color: #cbd5e1; }
.dcard-menu-empty span { font-size: 0.82rem; font-weight: 600; }

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
.dcard-reviews {
  border: 1px solid #e0e7ff;
  min-height: 0;
  align-content: start;
}
.dcard-reviews-list { display: grid; gap: 0.75rem; min-height: 0; }
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

.dcard-header-row { display: flex; align-items: center; justify-content: space-between; }
.dcard-header-row h2 { margin: 0; }
.section-count {
  min-width: 34px;
  height: 24px;
  padding: 0 0.45rem;
  display: inline-grid;
  place-items: center;
  border-radius: 9999px;
  background: #f8fafc;
  color: var(--text-muted);
  font-size: 0.62rem;
  font-weight: 700;
  line-height: 1;
  border: 1px solid var(--border-light);
}

.dcard-distance-pill {
  display: inline-flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.55rem;
  border-radius: 9999px; background: #eff6ff; color: var(--brand-blue);
  font-size: 0.62rem; font-weight: 700; white-space: nowrap;
}
.dcard-distance-pill i,
.dcard-link i,
.dcard-amenity-section strong i,
.dcard-amenity-chip i {
  line-height: 1;
}

.dcard-location-rows, .dcard-rows { display: grid; gap: 0.45rem; }
.dcard-row { display: flex; align-items: flex-start; gap: 0.5rem; color: var(--text-secondary); font-size: 0.8rem; line-height: 1.45; }
.dcard-row i { color: var(--text-muted); font-size: 0.78rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.08rem; line-height: 1; }
.dcard-row a { color: var(--brand-blue); text-decoration: none; }
.dcard-row a:hover { text-decoration: underline; }
.dcard-row-link { cursor: pointer; }
.dcard-row-coords { color: var(--text-muted); font-size: 0.75rem; }
.dcard-row-coords code { font-size: 0.7rem; color: var(--text-muted); background: #f1f5f9; padding: 0.1rem 0.3rem; border-radius: 4px; }

.dcard-place-id { display: flex; flex-direction: column; gap: 0.15rem; flex: 1; min-width: 0; }
.dcard-place-id span { color: var(--text-muted); font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.dcard-place-id code {
  font-size: 0.7rem; color: var(--text-secondary); background: #f1f5f9;
  padding: 0.2rem 0.4rem; border-radius: 6px; word-break: break-all; line-height: 1.4;
}
.dcard-copy-btn {
  display: inline-flex; align-items: center; justify-content: center; width: 32px; height: 32px;
  border-radius: 8px; border: 1px solid var(--border-light); background: #fff;
  color: var(--text-muted); cursor: pointer; font-size: 0.82rem; flex-shrink: 0; transition: all 0.15s ease;
}
.dcard-copy-btn:hover { color: var(--brand-blue); border-color: #bfdbfe; background: #eff6ff; }

.dcard-visit-list { display: grid; gap: 0.65rem; }
.dcard-visit { border: 1px solid var(--border-light); border-radius: 14px; overflow: hidden; }
.dcard-visit-header { display: flex; align-items: center; justify-content: space-between; padding: 0.55rem 0.85rem; background: #f8fafc; border-bottom: 1px solid var(--border-light); }
.dcard-visit-header span { color: var(--text-muted); font-size: 0.72rem; }
.dcard-visit-body { padding: 0.65rem 0.85rem; display: grid; gap: 0.35rem; }
.dcard-visit-detail { display: flex; align-items: flex-start; gap: 0.5rem; font-size: 0.78rem; color: var(--text-secondary); }
.dcard-visit-detail i { color: var(--text-muted); font-size: 0.68rem; margin-top: 0.18rem; flex-shrink: 0; }
.dcard-visit-exec { margin-top: 0.25rem; padding-top: 0.35rem; border-top: 1px solid var(--border-light); }
.dcard-visit-selfie { margin-top: 0.25rem; }
.dcard-visit-selfie img { width: 100%; max-width: 240px; border-radius: 10px; border: 1px solid var(--border-light); }

.dcard-timeline { display: grid; }
.dcard-timeline-entry { display: grid; grid-template-columns: 16px 1fr; gap: 0.75rem; padding-bottom: 1rem; position: relative; }
.dcard-timeline-entry:not(:last-child)::before { content: ''; position: absolute; left: 7px; top: 18px; bottom: 0; width: 2px; background: var(--border-light); }
.dcard-timeline-dot { width: 16px; height: 16px; border-radius: 50%; background: var(--brand-blue); border: 3px solid var(--brand-blue-bg); flex-shrink: 0; }
.dcard-timeline-content { display: grid; gap: 0.1rem; }
.dcard-timeline-content strong { font-size: 0.78rem; color: var(--text-primary); text-transform: capitalize; }
.dcard-timeline-content span { color: var(--text-muted); font-size: 0.68rem; }
.dcard-timeline-content p { margin: 0.2rem 0 0; font-size: 0.78rem; color: var(--text-secondary); line-height: 1.5; }

.dcard-empty-text { margin: 0; color: var(--text-muted); font-size: 0.82rem; text-align: center; padding: 1.5rem 0; }

.section-toggle-btn {
  width: 100%;
  min-height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  margin-top: 0.15rem;
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--border-light);
  border-radius: 10px;
  background: #f8fafc;
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
  transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
}
.section-toggle-btn:hover {
  color: var(--brand-blue);
  border-color: #bfdbfe;
  background: #eff6ff;
}
.section-toggle-btn i { font-size: 0.62rem; }

.detail-comments {
  min-width: 0;
}

.detail-comments :deep(.pc-wrap) {
  margin: 0;
}

.detail-bottom-bar {
  position: fixed; bottom: 0; left: 0; right: 0;
  width: 100%; z-index: 1000;
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.5rem;
  padding: 0.75rem 1rem; padding-bottom: calc(0.75rem + env(safe-area-inset-bottom));
  background: var(--surface-card); border-top: 1px solid var(--border-light);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.dbar-btn {
  min-height: 44px;
  display: inline-flex; align-items: center; justify-content: center; gap: 0.45rem;
  padding: 0.65rem 0.45rem; border-radius: 12px; border: none;
  font-size: 0.72rem; font-weight: 700; cursor: pointer;
  line-height: 1; text-decoration: none; text-align: center; transition: all 0.15s ease;
}
.dbar-btn i { width: 16px; height: 16px; display: inline-grid; place-items: center; font-size: 0.9rem; line-height: 1; flex-shrink: 0; }
.dbar-navigate { background: var(--brand-blue); color: #fff; }
.dbar-navigate:hover { background: #1d4ed8; }
.dbar-navigate:disabled { background: #cbd5e1; cursor: not-allowed; }
.dbar-call { background: #f0fdf4; color: #059669; border: 1px solid #a7f3d0; }
.dbar-call:hover { background: #dcfce7; }
.dbar-disabled { opacity: 0.45; cursor: not-allowed; pointer-events: none; }
.dbar-checkin { background: #eff6ff; color: var(--brand-blue); border: 1px solid #bfdbfe; }
.dbar-checkin:hover { background: #dbeafe; }
.dbar-checkout { background: #fff7ed; color: #c2410c; border: 1px solid #fed7aa; }
.dbar-checkout:hover { background: #ffedd5; }

/* Data Source Legend */
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
.ds-legend-body { display: grid; gap: 0.45rem; padding: 0 1rem 0.75rem; }
.ds-legend-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.72rem; color: var(--text-secondary); line-height: 1.4; }

@media (max-width: 767px) {
  .detail-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.6rem;
  }
  .detail-column { display: contents; }
  .dcard-visit-list,
  .dcard-timeline,
  .dcard-reviews-list {
    max-height: none !important;
    overflow: visible !important;
    padding-right: 0 !important;
  }
  .detail-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.6rem;
  }
  .detail-column {
    display: grid;
    gap: 0.6rem;
  }
  .dcard-visit-list,
  .dcard-timeline,
  .dcard-reviews-list {
    max-height: none !important;
    overflow: visible !important;
    padding-right: 0 !important;
  }
  .detail-content-grid { display: block; column-count: 1; }
  .detail-column { display: contents; }

  .history-scroll-region,
  .dcard-visit-list,
  .dcard-timeline,
  .detail-reviews-section .dcard-reviews-list {
    grid-template-columns: minmax(0, 1fr);
  }
  .detail-page { gap: 0.7rem; }
  .dcard { padding: 1rem; }
  .dcard-identity h1 { font-size: 1.05rem; }
  .dcard-title-line { flex-direction: column; gap: 0.45rem; }
  .summary-status { align-self: flex-start; }
  .summary-meta-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .dcard-codes { grid-template-columns: minmax(0, 1fr); }
  .dcard-photo-item { flex: 0 0 160px; height: 120px; }
  .photo-scroll-btn { width: 32px; height: 32px; }
}

/* ── Desktop ───────────────────────────────────────────────── */
@media (min-width: 1024px) {
  .detail-page {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.6rem;
    padding-bottom: 7rem;
    align-items: start;
    align-content: start;
  }

  .back-link { display: none; }

  /*
   * Stable dense document flow.
   * We intentionally use one full-width card flow for variable-height sections.
   * This removes the dead-space problem permanently: no section has to wait
   * for a taller card in the opposite column.
   *
   * The page stays compact by using multi-column DATA GRIDS inside the cards,
   * not by placing variable-height cards side-by-side.
   */
  .detail-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.6rem;
    min-width: 0;
    align-items: start;
  }

  .detail-column {
    display: contents;
  }

  .detail-content-grid > *,
  .detail-column > .dcard {
    width: 100%;
    min-width: 0;
    min-height: 0;
    align-self: start;
  }

  /* Keep the reading order predictable even though the original template
     stores cards in two wrappers. */
  .dcard-google-info { order: 1; }
  .dcard-photos { order: 2; }
  .dcard-location { order: 3; }
  .dcard-contact { order: 4; }
  .dcard-business { order: 5; }
  .dcard-hours { order: 6; }
  .dcard-service-options { order: 7; }
  .dcard-amenities { order: 8; }
  .dcard-sales-notes { order: 9; }
  .dcard-visits { order: 10; }
  .dcard-status-history { order: 11; }

  /* Use horizontal space INSIDE cards so the page remains compact. */
  .dcard-info-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.55rem 1rem;
  }

  .dcard-google-info .dcard-rows,
  .dcard-contact .dcard-rows,
  .dcard-business .dcard-rows {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.55rem 1rem;
  }

  .dcard-location .dcard-location-rows,
  .dcard-sales-notes .dcard-rows {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.55rem 1rem;
  }

  .dcard-amenities-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.7rem 1rem;
  }

  /* Visit/Status/Reviews grow naturally. No nested scroll areas. */
  .dcard-visit-list,
  .dcard-timeline,
  .dcard-reviews-list {
    max-height: none;
    overflow: visible;
    padding-right: 0;
  }

  .dcard-visits .dcard-visit-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.65rem;
  }

  .detail-reviews-section {
    width: 100%;
    min-width: 0;
  }

  .detail-reviews-section .dcard-reviews-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.65rem;
  }

  .detail-reviews-section .dcard-review {
    min-width: 0;
    padding: 0.7rem;
    border: 1px solid var(--border-light);
    border-radius: 12px;
    background: #fff;
  }

  .detail-reviews-section .dcard-review:last-child {
    border-bottom: 1px solid var(--border-light);
    padding-bottom: 0.7rem;
  }

  .detail-comments {
    width: 100%;
    min-width: 0;
    margin-top: 0;
  }

  .detail-page :deep(.pc-wrap) {
    width: 100%;
    min-width: 0;
  }

  .photo-gallery-shell :deep(*) {
    min-width: 0;
  }
}

@media (min-width: 769px) {
  .detail-bottom-bar {
    left: var(--sidebar-width, 240px);
    right: 0;
    width: auto;
    min-height: var(--desktop-action-bar-height, 72px);
    grid-template-columns: repeat(3, minmax(0, 1fr));
    align-items: center;
    gap: 0.75rem;
    padding: 0.65rem 2rem;
  }

  .dbar-btn {
    width: 100%;
    min-width: 0;
    min-height: 44px;
    height: 44px;
    box-sizing: border-box;
    border-radius: 12px;
    white-space: nowrap;
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .detail-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.65rem;
  }
  .detail-column { display: contents; }
  .dcard-visit-list,
  .dcard-timeline,
  .dcard-reviews-list {
    max-height: none !important;
    overflow: visible !important;
    padding-right: 0 !important;
  }
  .detail-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.65rem;
  }
  .detail-column {
    display: grid;
    gap: 0.65rem;
  }
  .dcard-visit-list,
  .dcard-timeline,
  .dcard-reviews-list {
    max-height: none !important;
    overflow: visible !important;
    padding-right: 0 !important;
  }
  .detail-content-grid { display: block; column-count: 1; }
  .detail-column { display: contents; }

  .history-scroll-region,
  .dcard-visit-list,
  .dcard-timeline,
  .detail-reviews-section .dcard-reviews-list {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (min-width: 768px) and (max-width: 1199px) {
  .summary-meta-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

</style>