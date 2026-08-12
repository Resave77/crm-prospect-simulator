<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getMyProspect, getProspectPlaceDetails } from '../../../api/crm'
import { useAuthStore } from '../../../stores/auth'
import type { ProspectReview, PlaceDetails } from '../../../types/crm'
import EntityLocationMap from '../../../components/sales/EntityLocationMap.vue'
import ProspectComments from '../../../components/ProspectComments.vue'
import PlacePhotoGallery from '../../../components/PlacePhotoGallery.vue'
import AISummaryCard from '../../../components/prospect-ai/AISummaryCard.vue'
import AIMenuProfilingCard from '../../../components/prospect-ai/AIMenuProfilingCard.vue'
import TanyaAICard from '../../../components/prospect-ai/TanyaAICard.vue'
import DataSourceBadge from '../../../components/sales/detail/DataSourceBadge.vue'
import { openGoogleMapsNavigation, getDistanceTo, formatDistance } from '../../../utils/maps'
import { formatPlaceType, isValidWebsite, websiteDisplayUrl, isValidPhone, copyToClipboard } from '../../../utils/placeDetails'
import { initials, formatErrorMessage, formatVisitDate, calcDuration } from '../../../utils/format'
import { priceLevelLabel, priceLevelSeverity, businessStatusLabel, businessStatusSeverity, stars, utcOffsetLabel } from '../../../utils/placeLabels'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
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
const expandedPanel = ref<'summary' | 'discussion' | 'chat' | null>(null)
const apiBase = import.meta.env.VITE_API_BASE_URL || ''

const userCoords = ref<{ lat: number; lng: number } | null>(null)
const isDesktop = ref(false)
const photoGalleryShell = ref<HTMLElement | null>(null)
let geoWatchId: number | null = null
let desktopQuery: MediaQueryList | null = null

const openVisit = computed(() => review.value?.visits.find((v) => !v.checkOutAt) ?? null)
const canViewAISummary = computed(() => auth.hasPermission('view_ai_summary'))
const canViewAIMenuProfiling = computed(() => auth.hasPermission('view_ai_menu_profiling'))
const canUseProspectAIChat = computed(() => auth.hasPermission('use_prospect_ai_chat'))
const hasPhotos = computed(() => (placeDetails.value?.photos?.length ?? 0) > 0)

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


function openExpandedPanel(panel: 'summary' | 'discussion' | 'chat') {
  expandedPanel.value = panel
  document.body.style.overflow = 'hidden'
}

function closeExpandedPanel() {
  expandedPanel.value = null
  document.body.style.overflow = ''
}

function onGlobalKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && expandedPanel.value) closeExpandedPanel()
}

function syncDesktop(value: MediaQueryList | MediaQueryListEvent) {
  isDesktop.value = value.matches
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/sales/my-prospects')
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
  window.addEventListener('keydown', onGlobalKeydown)
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
  window.removeEventListener('keydown', onGlobalKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <section class="detail-page">
    <button
      class="back-link"
      type="button"
      aria-label="Back to previous page"
      title="Back"
      @click="goBack"
    >
      <i class="pi pi-arrow-left" />
    </button>

    <Message v-if="success" severity="success" closable @close="success = ''">{{ success }}</Message>
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>

    <div class="ds-legend">
      <button class="ds-legend-toggle" type="button" @click="showLegend = !showLegend">
        <i class="pi pi-info-circle" />
        <span>Data source legend</span>
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
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
      <div class="sk-gallery" />
    </div>

    <div v-else-if="!review" class="detail-empty">
      <div class="detail-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>Prospect not found</strong>
      <span>This prospect may have been removed or you don't have access.</span>
      <RouterLink class="detail-empty-btn" to="/sales/my-prospects"><i class="pi pi-arrow-left" /> Back to prospects</RouterLink>
    </div>

    <template v-else>
      <!-- Reference-style header: identity + actions, not a duplicated detail card -->
      <header class="prospect-hero">
        <div class="hero-breadcrumb"><span>Prospect</span><i class="pi pi-angle-right" /><strong>Detail</strong></div>

        <div class="hero-row">
          <div class="hero-identity">
            <div class="hero-title-row">
              <h1>{{ review.prospect.placeName || 'Unnamed Prospect' }}</h1>
              <DataSourceBadge source="google" label="Google" />
            </div>
            <div v-if="review.prospect.formattedAddress" class="hero-address">
              <i class="pi pi-map-marker" />
              <span>{{ review.prospect.formattedAddress }}</span>
            </div>
          </div>

          <div class="hero-actions">
            <button
              class="hero-action hero-action-nav"
              type="button"
              :disabled="review.prospect.latitude == null && review.prospect.longitude == null && !review.prospect.formattedAddress"
              @click="navigate"
            >
              <i class="pi pi-directions" /> Navigate
            </button>
            <a
              v-if="review.prospect.phoneNumber && isValidPhone(review.prospect.phoneNumber)"
              :href="`tel:${review.prospect.phoneNumber}`"
              class="hero-action hero-action-call"
            >
              <i class="pi pi-phone" /> Call
            </a>
            <span v-else class="hero-action hero-action-call hero-action-disabled"><i class="pi pi-phone" /> Call</span>
            <RouterLink
              v-if="openVisit"
              class="hero-action hero-action-visit"
              :to="`/sales/my-prospects/${review.prospect.id}/check-out`"
            >
              <i class="pi pi-sign-out" /> Check out
            </RouterLink>
            <RouterLink
              v-else
              class="hero-action hero-action-visit"
              :to="`/sales/my-prospects/${review.prospect.id}/check-in`"
            >
              <i class="pi pi-sign-in" /> Check in
            </RouterLink>
          </div>

          <Tag
            class="hero-status"
            :value="review.prospect.status.replaceAll('_', ' ')"
            :severity="statusSeverity"
          />
        </div>
      </header>

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
        <!-- LEFT / MAIN: kept long so empty menu/photo states can never leave a tall dead zone -->
        <main class="detail-column detail-column-main">
          <section class="dcard dcard-prospect-core">
            <div class="section-heading">
              <h2>Prospect Detail</h2>
              <DataSourceBadge source="manual" label="CRM" />
            </div>

            <div class="prospect-detail-grid">
              <div class="detail-field">
                <span><i class="pi pi-user" /> Assigned Sales</span>
                <strong>{{ review.prospect.assignedSalesExecutive || 'Unassigned' }}</strong>
                <DataSourceBadge source="manual" label="Manual" />
              </div>
              <div class="detail-field">
                <span><i class="pi pi-phone" /> Phone</span>
                <strong>{{ review.prospect.phoneNumber || 'Not provided' }}</strong>
                <DataSourceBadge source="google" label="Google" />
              </div>
              <div class="detail-field">
                <span><i class="pi pi-tag" /> Category</span>
                <strong>{{ review.prospect.placeCategory || 'Not provided' }}</strong>
                <DataSourceBadge source="google" label="Google" />
              </div>
              <div class="detail-field">
                <span><i class="pi pi-briefcase" /> Industry</span>
                <strong>{{ review.prospect.industryGroup || 'Not provided' }}</strong>
                <DataSourceBadge source="google" label="Google" />
              </div>
              <div class="detail-field detail-field-wide">
                <span><i class="pi pi-id-card" /> Google Place ID</span>
                <strong class="wrap-anywhere">{{ review.prospect.googlePlaceId || '-' }}</strong>
                <DataSourceBadge source="google" label="Google" />
              </div>
              <div class="detail-field detail-field-wide">
                <span><i class="pi pi-refresh" /> Last Updated</span>
                <strong>{{ new Date(review.prospect.updatedAt).toLocaleString() }}</strong>
                <DataSourceBadge source="system" label="CRM" />
              </div>
            </div>

            <div class="prospect-tags">
              <Tag v-if="review.prospect.placeCategory" :value="review.prospect.placeCategory" severity="secondary" />
              <Tag v-if="review.prospect.industryGroup" :value="review.prospect.industryGroup" />
              <Tag
                v-if="placeDetails?.businessStatus"
                :value="businessStatusLabel(placeDetails.businessStatus)"
                :severity="businessStatusSeverity(placeDetails.businessStatus)"
              />
              <Tag
                v-if="placeDetails?.priceLevel"
                :value="priceLevelLabel(placeDetails.priceLevel)"
                :severity="priceLevelSeverity(placeDetails.priceLevel)"
              />
              <span v-for="t in displayTypes" :key="t" class="type-chip">{{ formatPlaceType(t) }}</span>
            </div>
          </section>

          <!-- Combined Google Maps Information like the reference -->
          <section v-if="placeDetails" class="dcard dcard-google-info">
            <div class="section-heading">
              <h2>Google Maps Information</h2>
              <DataSourceBadge source="google" label="Google" />
            </div>

            <div class="google-layout">
              <div class="google-map-pane">
                <EntityLocationMap
                  :latitude="review.prospect.latitude"
                  :longitude="review.prospect.longitude"
                  :label="review.prospect.placeName"
                  :interactive="false"
                  height="245px"
                />
                <div class="map-footer">
                  <a
                    v-if="review.prospect.googleMapsUrl"
                    :href="review.prospect.googleMapsUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    View on Google Maps <i class="pi pi-external-link" />
                  </a>
                  <span
                    v-if="review.prospect.latitude != null && review.prospect.longitude != null && userCoords"
                    class="distance-pill"
                  >
                    {{ formatDistance(getDistanceTo(review.prospect.latitude, review.prospect.longitude, userCoords.lat, userCoords.lng)!) }} away
                  </span>
                </div>
              </div>

              <div class="google-detail-pane">
                <div v-if="placeDetails.rating > 0" class="rating-line">
                  <strong>{{ placeDetails.rating.toFixed(1) }}</strong>
                  <span class="stars"><i v-for="(s, i) in stars(placeDetails.rating)" :key="i" :class="['pi', s]" /></span>
                  <span>({{ placeDetails.userRatingCount.toLocaleString() }} reviews)</span>
                </div>

                <p v-if="placeDetails.editorialSummary" class="google-editorial">{{ placeDetails.editorialSummary }}</p>

                <div v-if="placeDetails.placeTypes?.length" class="category-line">
                  <strong>Categories:</strong>
                  <span v-for="t in placeDetails.placeTypes.slice(0, 6)" :key="t">{{ t.replace(/_/g, ' ') }}</span>
                </div>

                <div class="google-rows">
                  <div v-if="placeDetails.placeName"><i class="pi pi-building" /><strong>Google place:</strong><span>{{ placeDetails.placeName }}</span></div>
                  <div v-if="placeDetails.formattedAddress"><i class="pi pi-map-marker" /><strong>Google address:</strong><span>{{ placeDetails.formattedAddress }}</span></div>
                  <div v-if="placeDetails.googlePlaceId">
                    <i class="pi pi-id-card" /><strong>Google Place ID:</strong><span class="wrap-anywhere">{{ placeDetails.googlePlaceId }}</span>
                    <button class="copy-mini" type="button" aria-label="Copy Google Place ID" @click="handleCopy(placeDetails.googlePlaceId)"><i class="pi pi-copy" /></button>
                  </div>
                  <div v-if="placeDetails.latitude != null && placeDetails.longitude != null">
                    <i class="pi pi-compass" /><strong>Google GPS:</strong><span>{{ placeDetails.latitude.toFixed(6) }}, {{ placeDetails.longitude.toFixed(6) }}</span>
                  </div>
                  <div v-if="placeDetails.placeCategory"><i class="pi pi-tag" /><strong>Category:</strong><span>{{ placeDetails.placeCategory }}</span></div>
                  <div v-if="placeDetails.internationalPhone"><i class="pi pi-phone" /><strong>International phone:</strong><span>{{ placeDetails.internationalPhone }}</span></div>
                  <div v-if="placeDetails.utcOffsetMinutes != null"><i class="pi pi-globe" /><strong>Time zone:</strong><span>{{ utcOffsetLabel(placeDetails.utcOffsetMinutes) }}</span></div>
                  <div v-if="placeDetails.websiteUrl && isValidWebsite(placeDetails.websiteUrl)">
                    <i class="pi pi-external-link" /><strong>Website:</strong>
                    <a :href="placeDetails.websiteUrl" target="_blank" rel="noopener">{{ websiteDisplayUrl(placeDetails.websiteUrl) }}</a>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- Completely omit the media card if Google returned no photos: no artificial blank box -->
          <section v-if="hasPhotos" class="dcard dcard-photos">
            <div class="section-heading section-heading-between">
              <div class="section-title-inline">
                <h2><i class="pi pi-images" /> Photos</h2>
                <DataSourceBadge source="google" label="Google" />
              </div>
              <div class="photo-gallery-controls" aria-label="Photo gallery navigation">
                <button type="button" class="photo-scroll-btn" aria-label="Scroll photos left" @click="scrollPhotos(-1)"><i class="pi pi-chevron-left" /></button>
                <button type="button" class="photo-scroll-btn" aria-label="Scroll photos right" @click="scrollPhotos(1)"><i class="pi pi-chevron-right" /></button>
              </div>
            </div>
            <div ref="photoGalleryShell" class="photo-gallery-shell">
              <PlacePhotoGallery :photos="(placeDetails?.photos ?? [])" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" section="photos" />
            </div>
          </section>

          <!-- Keep a compact menu state only when photo resources exist; it collapses instead of reserving a tall blank area -->
          <section v-if="hasPhotos" class="dcard dcard-menu">
            <div class="section-heading">
              <h2><i class="pi pi-book" /> Menu</h2>
              <DataSourceBadge source="google" label="Google" />
            </div>
            <PlacePhotoGallery :photos="(placeDetails?.photos ?? [])" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" section="menu" />
          </section>

          <AIMenuProfilingCard
            v-if="canViewAIMenuProfiling"
            class="detail-ai-menu"
            :place-details="placeDetails"
          />

          <!-- Reviews and histories stay INSIDE the main stack. This prevents a shorter media/menu stack
               from creating the long blank column seen in the screenshots. -->
          <section v-if="placeDetails?.reviews?.length" class="dcard dcard-reviews">
            <div class="section-heading section-heading-between">
              <h2><i class="pi pi-comments" /> Reviews</h2>
              <span class="section-count">{{ placeDetails.reviews.length }}</span>
            </div>
            <div class="reviews-grid">
              <article v-for="(item, i) in displayedReviews" :key="i" class="review-card">
                <div class="review-head">
                  <img v-if="item.authorPhoto" :src="item.authorPhoto" :alt="item.authorName" @error="($event.target as HTMLImageElement).style.display='none'" />
                  <span v-else class="review-avatar">{{ item.authorName?.charAt(0) || '?' }}</span>
                  <div>
                    <strong>{{ item.authorName }}</strong>
                    <div class="review-rating"><span class="stars"><i v-for="(s, j) in stars(item.rating)" :key="j" :class="['pi', s]" /></span><small>{{ item.time }}</small></div>
                  </div>
                </div>
                <p v-if="item.text">{{ item.text }}</p>
              </article>
            </div>
            <button v-if="placeDetails.reviews.length > 4" class="section-toggle-btn" type="button" @click="showAllReviews = !showAllReviews">
              {{ showAllReviews ? 'Show less' : `Show all ${placeDetails.reviews.length} reviews` }}
              <i :class="showAllReviews ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
            </button>
          </section>

          <section class="dcard history-card">
            <div class="section-heading section-heading-between">
              <h2>Visit History</h2>
              <span class="section-count">{{ review.visits.length }}</span>
            </div>
            <div v-if="review.visits.length" class="visit-grid">
              <article v-for="visit in displayedVisits" :key="visit.id" class="visit-card">
                <div class="visit-head">
                  <Tag :value="visit.checkOutAt ? 'Completed' : 'Active'" :severity="visit.checkOutAt ? 'secondary' : 'success'" />
                  <span>{{ formatVisitDate(visit.checkInAt) }}</span>
                </div>
                <div class="visit-body">
                  <div><i class="pi pi-sign-in" /> Check-in: {{ visit.checkInLatitude.toFixed(4) }}, {{ visit.checkInLongitude.toFixed(4) }}</div>
                  <div v-if="visit.checkOutAt"><i class="pi pi-sign-out" /> Check-out: {{ visit.checkOutLatitude?.toFixed(4) }}, {{ visit.checkOutLongitude?.toFixed(4) }}</div>
                  <div v-if="visit.checkOutAt"><i class="pi pi-clock" /> Duration: {{ calcDuration(visit.checkInAt, visit.checkOutAt) }}</div>
                  <div v-if="visit.visitNotes"><i class="pi pi-comment" /> {{ visit.visitNotes }}</div>
                  <div v-if="visit.followUpNotes"><i class="pi pi-directions" /> Follow-up: {{ visit.followUpNotes }}</div>
                  <img
                    v-if="visit.selfieReference && visit.selfieReference !== 'SIMULATED_SELFIE_PLACEHOLDER'"
                    class="visit-selfie"
                    :src="visit.selfieReference.startsWith('/') ? `${apiBase}${visit.selfieReference}` : visit.selfieReference"
                    alt="Visit selfie"
                  />
                  <div class="visit-user"><i class="pi pi-user" /> {{ visit.salesExecutiveName }}</div>
                </div>
              </article>
            </div>
            <div v-else class="compact-empty"><i class="pi pi-calendar-times" /><span>No visits recorded yet.</span></div>
            <button v-if="review.visits.length > 2" class="section-toggle-btn" type="button" @click="showAllVisits = !showAllVisits">
              {{ showAllVisits ? 'Show less' : `Show all ${review.visits.length} visits` }}
              <i :class="showAllVisits ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
            </button>
          </section>

          <section class="dcard history-card">
            <div class="section-heading section-heading-between">
              <h2>Status History</h2>
              <span class="section-count">{{ review.history.length }}</span>
            </div>
            <div v-if="review.history.length" class="timeline">
              <div v-for="entry in displayedStatusHistory" :key="entry.id" class="timeline-entry">
                <span class="timeline-dot" />
                <div>
                  <strong>{{ (entry.fromStatus || 'Created').replaceAll('_', ' ') }} → {{ entry.toStatus.replaceAll('_', ' ') }}</strong>
                  <small>{{ new Date(entry.createdAt).toLocaleString() }} · {{ entry.changedByName }}</small>
                  <p v-if="entry.notes">{{ entry.notes }}</p>
                </div>
              </div>
            </div>
            <div v-else class="compact-empty"><i class="pi pi-history" /><span>No status changes recorded.</span></div>
            <button v-if="review.history.length > 6" class="section-toggle-btn" type="button" @click="showAllStatusHistory = !showAllStatusHistory">
              {{ showAllStatusHistory ? 'Show less' : `Show all ${review.history.length} changes` }}
              <i :class="showAllStatusHistory ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
            </button>
          </section>
        </main>

        <!-- RIGHT / SIDEBAR -->
        <aside class="detail-column detail-column-side">
          <div v-if="canViewAISummary" class="ai-shell ai-shell-summary">
            <button class="expand-control" type="button" title="Expand AI Summary" aria-label="Expand AI Summary" @click="openExpandedPanel('summary')">
              <i class="pi pi-window-maximize" /><span>Expand</span>
            </button>
            <AISummaryCard class="detail-ai-summary" :prospect-name="review.prospect.placeName" />
          </div>

          <section class="dcard business-contact-card">
            <div class="section-heading"><h2>Business &amp; Contact Details</h2></div>
            <dl class="business-list">
              <div><dt>Business name</dt><dd>{{ review.prospect.placeName || 'Not provided' }}</dd></div>
              <div><dt>Phone (Primary)</dt><dd><a v-if="review.prospect.phoneNumber" :href="`tel:${review.prospect.phoneNumber}`">{{ review.prospect.phoneNumber }}</a><span v-else>Not provided</span></dd></div>
              <div v-if="placeDetails?.internationalPhone"><dt>Phone (International)</dt><dd>{{ placeDetails.internationalPhone }}</dd></div>
              <div><dt>Category</dt><dd>{{ review.prospect.placeCategory || 'Not provided' }}</dd></div>
              <div><dt>Industry</dt><dd>{{ review.prospect.industryGroup || 'Not provided' }}</dd></div>
              <div v-if="placeDetails?.openingHours"><dt>Operating hours</dt><dd>{{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}</dd></div>
              <div v-if="placeDetails?.priceLevel"><dt>Price level</dt><dd>{{ priceLevelLabel(placeDetails.priceLevel) }}</dd></div>
              <div><dt>Source</dt><dd>Google Maps / CRM</dd></div>
              <div><dt>Added by</dt><dd>{{ review.prospect.assignedSalesExecutive || 'CRM' }}</dd></div>
              <div><dt>Status</dt><dd>{{ review.prospect.status.replaceAll('_', ' ') }}</dd></div>
            </dl>

            <div v-if="placeDetails?.openingHours?.weekdays?.length" class="hours-compact">
              <div v-for="(day, i) in (showAllHours ? placeDetails.openingHours.weekdays : placeDetails.openingHours.weekdays.slice(0, 3))" :key="i" v-html="day" />
              <button v-if="placeDetails.openingHours.weekdays.length > 3" type="button" @click="showAllHours = !showAllHours">
                {{ showAllHours ? 'Show less' : 'Show all operating hours' }}
              </button>
            </div>

            <div v-if="placeDetails && (placeDetails.delivery || placeDetails.dineIn || placeDetails.takeout || placeDetails.curbsidePickup)" class="service-chips">
              <Tag v-if="placeDetails.dineIn" value="Dine In" severity="success" />
              <Tag v-if="placeDetails.takeout" value="Takeout" severity="info" />
              <Tag v-if="placeDetails.delivery" value="Delivery" severity="warn" />
              <Tag v-if="placeDetails.curbsidePickup" value="Curbside Pickup" severity="secondary" />
            </div>
          </section>

          <div class="discussion-shell ai-shell ai-shell-discussion">
            <button class="expand-control" type="button" title="Expand Discussion" aria-label="Expand Discussion" @click="openExpandedPanel('discussion')">
              <i class="pi pi-window-maximize" /><span>Expand</span>
            </button>
            <ProspectComments class="detail-comments" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" :embedded="isDesktop" />
          </div>

          <div v-if="canUseProspectAIChat" class="ai-shell ai-shell-chat">
            <button class="expand-control" type="button" title="Expand Tanya AI" aria-label="Expand Tanya AI" @click="openExpandedPanel('chat')">
              <i class="pi pi-window-maximize" /><span>Expand</span>
            </button>
            <TanyaAICard class="detail-tanya-ai" />
          </div>

          <!-- Optional supporting data only renders when it has actual content -->
          <section v-if="review.prospect.visitNotes || review.prospect.followUpNotes" class="dcard supporting-card">
            <div class="section-heading"><h2>Sales Notes</h2></div>
            <div class="supporting-rows">
              <p v-if="review.prospect.visitNotes"><strong>Visit notes</strong>{{ review.prospect.visitNotes }}</p>
              <p v-if="review.prospect.followUpNotes"><strong>Follow-up</strong>{{ review.prospect.followUpNotes }}</p>
            </div>
          </section>

          <section
            v-if="placeDetails?.parkingOptions || placeDetails?.paymentOptions || placeDetails?.accessibilityOptions"
            class="dcard supporting-card"
          >
            <div class="section-heading"><h2>Amenities</h2><DataSourceBadge source="google" label="Google" /></div>
            <div class="amenity-groups">
              <div v-if="placeDetails?.parkingOptions">
                <strong>Parking</strong>
                <div>
                  <span v-if="placeDetails.parkingOptions.freeParkingLot">Free Lot</span>
                  <span v-if="placeDetails.parkingOptions.freeStreetParking">Free Street</span>
                  <span v-if="placeDetails.parkingOptions.paidParkingLot">Paid Lot</span>
                  <span v-if="placeDetails.parkingOptions.paidStreetParking">Paid Street</span>
                  <span v-if="placeDetails.parkingOptions.garageParking">Garage</span>
                  <span v-if="placeDetails.parkingOptions.valetParking">Valet</span>
                </div>
              </div>
              <div v-if="placeDetails?.paymentOptions">
                <strong>Payment</strong>
                <div>
                  <span v-if="placeDetails.paymentOptions.cashOnly">Cash</span>
                  <span v-if="placeDetails.paymentOptions.creditCardOnly">Credit Card</span>
                  <span v-if="placeDetails.paymentOptions.debitCardOnly">Debit Card</span>
                  <span v-if="placeDetails.paymentOptions.nfcOnly">NFC</span>
                </div>
              </div>
              <div v-if="placeDetails?.accessibilityOptions">
                <strong>Accessibility</strong>
                <div>
                  <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleEntrance">Wheelchair Entrance</span>
                  <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleParking">Wheelchair Parking</span>
                  <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleRestroom">Wheelchair Restroom</span>
                  <span v-if="placeDetails.accessibilityOptions.wheelchairAccessibleSeating">Wheelchair Seating</span>
                </div>
              </div>
            </div>
          </section>
        </aside>
      </div>

      <!-- Mobile sticky action bar only -->
      <div class="detail-bottom-bar">
        <button
          class="dbar-btn dbar-navigate"
          type="button"
          :disabled="review.prospect.latitude == null && review.prospect.longitude == null && !review.prospect.formattedAddress"
          @click="navigate"
        >
          <i class="pi pi-directions" /> Navigate
        </button>
        <a v-if="review.prospect.phoneNumber && isValidPhone(review.prospect.phoneNumber)" :href="`tel:${review.prospect.phoneNumber}`" class="dbar-btn dbar-call">
          <i class="pi pi-phone" /> Call
        </a>
        <span v-else class="dbar-btn dbar-call dbar-disabled"><i class="pi pi-phone" /> Call</span>
        <RouterLink v-if="openVisit" class="dbar-btn dbar-checkin" :to="`/sales/my-prospects/${review.prospect.id}/check-out`"><i class="pi pi-sign-out" /> Check out</RouterLink>
        <RouterLink v-else class="dbar-btn dbar-checkin" :to="`/sales/my-prospects/${review.prospect.id}/check-in`"><i class="pi pi-sign-in" /> Check in</RouterLink>
      </div>

      <!-- Functional expand overlay. No OpenAI generation is triggered by opening/closing it. -->
      <Teleport to="body">
        <div v-if="expandedPanel" class="expand-overlay" @click.self="closeExpandedPanel">
          <section class="expand-dialog" role="dialog" aria-modal="true">
            <header class="expand-dialog-header">
              <div>
                <span>{{ expandedPanel === 'summary' ? 'AI Summary' : expandedPanel === 'discussion' ? 'Discussion' : 'Tanya AI' }}</span>
                <strong>{{ review.prospect.placeName }}</strong>
              </div>
              <button type="button" aria-label="Close expanded panel" @click="closeExpandedPanel"><i class="pi pi-times" /></button>
            </header>
            <div class="expand-dialog-body">
              <AISummaryCard v-if="expandedPanel === 'summary'" :prospect-name="review.prospect.placeName" />
              <ProspectComments v-else-if="expandedPanel === 'discussion'" :prospect-id="review.prospect.id" role="SALES_EXECUTIVE" :embedded="true" />
              <TanyaAICard v-else-if="expandedPanel === 'chat'" />
            </div>
          </section>
        </div>
      </Teleport>
    </template>
  </section>
</template>

<style scoped>
.detail-page {
  --detail-accent: var(--brand-red);
  --detail-accent-strong: var(--brand-red-hover);
  --detail-soft: var(--brand-red-50);
  --detail-soft-strong: #ffeaed;
  --detail-border: #e8e4e5;
  --detail-red-border: var(--brand-red-200);
  --detail-muted: #766c6e;
  --brand-blue: var(--detail-accent);
  --brand-blue-50: var(--detail-soft);
  --brand-blue-bg: var(--detail-soft-strong);

  width: 100%;
  min-width: 0;
  display: grid;
  gap: 0.75rem;
  align-content: start;
  padding-bottom: 5.5rem;
  overflow-x: clip;
}

.back-link {
  width: 38px; height: 38px; display: inline-grid; place-items: center;
  border: 1px solid var(--detail-border); border-radius: 10px;
  padding: 0;
  background: #fff; color: var(--detail-accent); text-decoration: none;
  cursor: pointer;
  transition: border-color .18s ease, background .18s ease, color .18s ease, transform .18s ease;
}
.back-link:hover {
  border-color: var(--detail-red-border);
  background: var(--detail-soft);
  color: var(--detail-accent-strong);
}
.back-link:active { transform: translateY(1px); }
.back-link:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--detail-accent) 28%, transparent);
  outline-offset: 2px;
}

.ds-legend {
  overflow: hidden; border: 1px solid var(--detail-border); border-radius: 11px;
  background: #fff; box-shadow: 0 2px 10px rgba(15, 23, 42, 0.025);
}
.ds-legend-toggle {
  width: 100%; min-height: 42px; display: flex; align-items: center; gap: .5rem;
  padding: .6rem .9rem; border: 0; background: transparent; cursor: pointer;
  color: #5f585a; font-size: .72rem; font-weight: 650;
}
.ds-legend-toggle i:first-child { color: var(--detail-accent); }
.ds-legend-toggle i:last-child { margin-left: auto; font-size: .62rem; }
.ds-legend-body { display: grid; gap: .45rem; padding: 0 .9rem .8rem; }
.ds-legend-item { display: flex; align-items: center; gap: .5rem; color: var(--text-secondary); font-size: .72rem; }

.detail-skeleton { display: grid; gap: .75rem; }
.sk-header { display: flex; gap: .7rem; align-items: center; }
.sk-circle { width: 48px; height: 48px; border-radius: 50%; background: #ece8e9; }
.sk-lines { flex: 1; display: grid; gap: .4rem; }
.sk-line { height: 12px; border-radius: 6px; background: #ece8e9; }
.sk-line.w70 { width: 70%; } .sk-line.w60 { width: 60%; } .sk-line.w50 { width: 50%; }
.sk-line.w40 { width: 40%; } .sk-line.w80 { width: 80%; }
.sk-card { padding: 1rem; border: 1px solid var(--detail-border); border-radius: 11px; background: #fff; display: grid; gap: .5rem; }
.sk-map { height: 180px; border-radius: 10px; background: #ece8e9; }
.sk-gallery { height: 150px; border-radius: 11px; background: #ece8e9; }

.detail-empty {
  min-height: 260px; display: grid; place-items: center; align-content: center; gap: .5rem;
  padding: 2rem; text-align: center;
}
.detail-empty-icon { width: 52px; height: 52px; display: grid; place-items: center; border-radius: 14px; background: var(--detail-soft); color: var(--detail-accent); }
.detail-empty span { color: var(--text-muted); font-size: .8rem; }
.detail-empty-btn { padding: .55rem .9rem; border-radius: 9px; background: var(--detail-accent); color: #fff; text-decoration: none; font-size: .75rem; font-weight: 700; }

/* Header mirrors the reference: identity left, actions/status right, no duplicated detail data. */
.prospect-hero {
  display: grid; gap: .7rem; padding: .1rem .15rem .35rem;
}
.hero-breadcrumb { display: flex; align-items: center; gap: .35rem; color: var(--text-muted); font-size: .68rem; font-weight: 650; }
.hero-breadcrumb span { color: var(--detail-accent); }
.hero-breadcrumb i { font-size: .55rem; }
.hero-row { display: grid; grid-template-columns: minmax(0, 1fr) auto auto; align-items: start; gap: .8rem; }
.hero-identity { min-width: 0; }
.hero-title-row { display: flex; align-items: center; flex-wrap: wrap; gap: .45rem; }
.hero-title-row h1 { margin: 0; color: var(--text-primary); font-size: 1.28rem; line-height: 1.2; font-weight: 800; }
.hero-address { margin-top: .35rem; display: flex; align-items: flex-start; gap: .42rem; color: #625b5d; font-size: .74rem; line-height: 1.45; }
.hero-address i { margin-top: .12rem; flex-shrink: 0; color: var(--detail-accent); font-size: .72rem; }
.hero-actions { display: flex; align-items: center; gap: .45rem; }
.hero-action {
  min-height: 38px; display: inline-flex; align-items: center; justify-content: center; gap: .4rem;
  padding: .55rem .78rem; border: 1px solid var(--detail-red-border); border-radius: 8px;
  background: #fff; color: var(--detail-accent); font-size: .7rem; font-weight: 750;
  text-decoration: none; cursor: pointer; white-space: nowrap;
}
.hero-action-nav:hover, .hero-action-visit:hover { background: var(--detail-soft); }
.hero-action-call { border-color: #b8e7d0; color: #087a50; }
.hero-action-call:hover { background: #f0fdf7; }
.hero-action:disabled { opacity: .45; cursor: not-allowed; }
.hero-action-disabled { opacity: .45; pointer-events: none; }
.hero-status { align-self: start; }

/* Cards */
.dcard {
  min-width: 0; display: grid; align-content: start; gap: .65rem; padding: .9rem;
  border: 1px solid var(--detail-border); border-radius: 10px; background: #fff;
  box-shadow: 0 3px 12px rgba(15, 23, 42, .03);
}
.section-heading { min-width: 0; display: flex; align-items: center; gap: .4rem; flex-wrap: wrap; }
.section-heading h2, .dcard h2 {
  margin: 0; color: #40383a; font-size: .67rem; line-height: 1.2;
  font-weight: 800; letter-spacing: .035em; text-transform: uppercase;
}
.section-heading h2 > i, .dcard h2 > i { color: var(--detail-accent); }
.section-heading-between { justify-content: space-between; }
.section-title-inline { display: flex; align-items: center; gap: .4rem; flex-wrap: wrap; }

.dcard-active-visit { border-color: #f2cd87; background: #fffbeb; }
.dcard-active-visit-row { display: flex; align-items: center; gap: .5rem; flex-wrap: wrap; color: #8b5314; font-size: .76rem; }
.dcard-active-visit-link { margin-left: auto; color: #9a5a12; font-size: .7rem; font-weight: 800; text-decoration: none; }

/* Prospect detail card */
.prospect-detail-grid {
  display: grid; grid-template-columns: repeat(4, minmax(0, 1fr));
  overflow: hidden; border: 1px solid #eee6e7; border-radius: 9px;
}
.detail-field {
  min-width: 0; display: grid; gap: .2rem; padding: .68rem .72rem;
  border-right: 1px solid #eee6e7; border-bottom: 1px solid #eee6e7;
}
.detail-field:nth-child(4) { border-right: 0; }
.detail-field-wide { grid-column: span 2; border-bottom: 0; }
.detail-field-wide:nth-last-child(1) { border-right: 0; }
.detail-field > span { display: flex; align-items: center; gap: .3rem; color: var(--detail-muted); font-size: .55rem; font-weight: 800; text-transform: uppercase; }
.detail-field > span i { color: var(--detail-accent); font-size: .62rem; }
.detail-field strong { color: var(--text-primary); font-size: .73rem; line-height: 1.35; }
.prospect-tags { display: flex; flex-wrap: wrap; gap: .32rem; }
.type-chip { padding: .15rem .45rem; border-radius: 6px; background: var(--detail-soft); color: var(--detail-accent-strong); font-size: .57rem; font-weight: 650; }
.wrap-anywhere { overflow-wrap: anywhere; }

/* Google block */
.dcard-google-info { border-color: #eadfe1; }
.google-layout { display: grid; grid-template-columns: minmax(240px, 38%) minmax(0, 1fr); gap: .9rem; align-items: start; }
.google-map-pane { min-width: 0; }
.map-footer { margin-top: .4rem; display: flex; align-items: center; justify-content: space-between; gap: .6rem; flex-wrap: wrap; }
.map-footer a { color: var(--detail-accent); font-size: .68rem; font-weight: 700; text-decoration: none; }
.distance-pill { padding: .18rem .45rem; border-radius: 999px; background: var(--detail-soft); color: var(--detail-accent); font-size: .6rem; font-weight: 700; }
.google-detail-pane { min-width: 0; display: grid; gap: .55rem; }
.rating-line { display: flex; align-items: center; gap: .35rem; flex-wrap: wrap; font-size: .68rem; color: var(--text-muted); }
.rating-line > strong { color: #d69215; font-size: .9rem; }
.stars { display: inline-flex; gap: 1px; color: #f6a700; }
.stars i { font-size: .58rem; }
.google-editorial { margin: 0; color: var(--text-secondary); font-size: .72rem; line-height: 1.45; }
.category-line { display: flex; align-items: center; gap: .3rem; flex-wrap: wrap; font-size: .61rem; }
.category-line span { padding: .12rem .38rem; border-radius: 999px; background: #f7f5f5; color: #5f585a; }
.google-rows { display: grid; gap: .35rem; }
.google-rows > div {
  min-width: 0; display: grid; grid-template-columns: 16px max-content minmax(0, 1fr) auto;
  align-items: start; gap: .35rem; color: var(--text-secondary); font-size: .68rem; line-height: 1.4;
}
.google-rows > div > i { margin-top: .12rem; color: #8e8587; font-size: .65rem; }
.google-rows a { color: var(--detail-accent); text-decoration: none; }
.copy-mini { width: 24px; height: 24px; display: grid; place-items: center; border: 1px solid var(--detail-border); border-radius: 6px; background: #fff; color: var(--detail-accent); cursor: pointer; }

/* Media */
.photo-gallery-shell { min-width: 0; overflow: hidden; }
.photo-gallery-controls { display: inline-flex; gap: .35rem; }
.photo-scroll-btn { width: 31px; height: 31px; display: grid; place-items: center; border: 1px solid var(--detail-border); border-radius: 8px; background: #fff; color: var(--detail-accent); cursor: pointer; }
.photo-scroll-btn:hover { border-color: var(--detail-red-border); background: var(--detail-soft); }

/* Critically keep empty menu/photo states compact. */
.detail-page :deep(.ppg-empty) {
  min-height: 0 !important;
  padding: .75rem !important;
  gap: .25rem !important;
  border-radius: 8px !important;
}
.detail-page :deep(.ppg-empty i) { font-size: 1rem !important; }
.detail-page :deep(.ppg-empty span) { font-size: .7rem !important; }
.detail-page :deep(.ppg-empty small) { font-size: .62rem !important; }

/* Main/sidebar independent stacks: reviews/history belong to main, so a short empty menu cannot hold them down. */
.detail-content-grid { display: grid; grid-template-columns: minmax(0, 1fr); gap: .75rem; min-width: 0; align-items: start; }
.detail-column { min-width: 0; display: grid; gap: .75rem; align-content: start; }
.detail-column > *, .ai-shell { min-width: 0; max-width: 100%; align-self: start; }

/* AI / Discussion wrappers and functional expand */
.ai-shell {
  position: relative;
  min-width: 0;
}

/*
 * The AI prep components still contain their old decorative expand affordance.
 * A small white mask keeps that old control from showing underneath the real button.
 * This avoids editing/re-breaking the child components.
 */
.ai-shell::after {
  content: '';
  position: absolute;
  top: .42rem;
  right: .42rem;
  z-index: 24;
  width: 5.7rem;
  height: 2.05rem;
  border-radius: 8px;
  background: #fff;
  pointer-events: none;
}

.ai-shell > .expand-control {
  position: absolute;
  top: .58rem;
  right: .58rem;
  z-index: 30;
  min-height: 28px;
  display: inline-flex;
  align-items: center;
  gap: .3rem;
  padding: .28rem .48rem;
  border: 1px solid var(--detail-red-border);
  border-radius: 7px;
  background: #fff;
  color: var(--detail-accent);
  font-size: .57rem;
  font-weight: 750;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(53, 30, 35, .07);
}
.ai-shell > .expand-control:hover {
  border-color: #e4a6ae;
  background: var(--detail-soft);
  color: var(--detail-accent-strong);
}
.ai-shell > .expand-control i { font-size: .61rem; }
.ai-shell > :deep(*) { min-width: 0; }
.detail-page :deep(.ai-card),
.detail-page :deep(.pai-card),
.detail-page :deep(.tanya-ai-card) { min-width: 0; }

/* Business & Contact like reference */
.business-list { margin: 0; display: grid; gap: .48rem; }
.business-list > div { display: grid; grid-template-columns: minmax(110px, 42%) minmax(0, 1fr); gap: .5rem; font-size: .66rem; line-height: 1.35; }
.business-list dt { color: var(--text-muted); }
.business-list dd { margin: 0; color: var(--text-primary); font-weight: 600; overflow-wrap: anywhere; }
.business-list a { color: var(--detail-accent); text-decoration: none; }
.hours-compact { display: grid; gap: .25rem; padding-top: .55rem; border-top: 1px solid var(--detail-border); color: var(--text-secondary); font-size: .63rem; }
.hours-compact button { width: max-content; padding: 0; border: 0; background: transparent; color: var(--detail-accent); font-size: .62rem; font-weight: 700; cursor: pointer; }
.service-chips { display: flex; flex-wrap: wrap; gap: .3rem; }

/* Supporting sections only exist when actual values exist. */
.supporting-rows { display: grid; gap: .5rem; }
.supporting-rows p { margin: 0; display: grid; gap: .15rem; color: var(--text-secondary); font-size: .68rem; line-height: 1.45; }
.supporting-rows strong { color: var(--text-primary); font-size: .62rem; }
.amenity-groups { display: grid; gap: .55rem; }
.amenity-groups > div { display: grid; gap: .28rem; }
.amenity-groups strong { font-size: .65rem; }
.amenity-groups > div > div { display: flex; flex-wrap: wrap; gap: .25rem; }
.amenity-groups span { padding: .15rem .4rem; border-radius: 999px; background: #f7f5f5; color: #655d5f; font-size: .58rem; }

/* Reviews */
.section-count { min-width: 28px; height: 22px; display: grid; place-items: center; padding: 0 .4rem; border: 1px solid var(--detail-red-border); border-radius: 999px; background: var(--detail-soft); color: var(--detail-accent); font-size: .6rem; font-weight: 800; }
.reviews-grid { display: grid; gap: .6rem; }
.review-card { min-width: 0; padding: .68rem; border: 1px solid var(--detail-border); border-radius: 9px; background: #fff; }
.review-head { display: flex; align-items: center; gap: .5rem; }
.review-head img, .review-avatar { width: 30px; height: 30px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }
.review-avatar { display: grid; place-items: center; background: var(--detail-soft); color: var(--detail-accent); font-size: .68rem; font-weight: 800; }
.review-head strong { font-size: .7rem; }
.review-rating { display: flex; align-items: center; gap: .3rem; }
.review-rating small { color: var(--text-muted); font-size: .58rem; }
.review-card p { margin: .35rem 0 0; color: var(--text-secondary); font-size: .68rem; line-height: 1.5; }

/* Visit/status */
.visit-grid { display: grid; gap: .6rem; }
.visit-card { overflow: hidden; border: 1px solid var(--detail-border); border-radius: 9px; }
.visit-head { display: flex; align-items: center; justify-content: space-between; gap: .5rem; padding: .5rem .65rem; background: #faf8f8; border-bottom: 1px solid var(--detail-border); }
.visit-head > span { color: var(--text-muted); font-size: .62rem; }
.visit-body { display: grid; gap: .3rem; padding: .6rem .65rem; color: var(--text-secondary); font-size: .67rem; line-height: 1.4; }
.visit-body > div { display: flex; align-items: flex-start; gap: .38rem; }
.visit-body i { margin-top: .12rem; color: var(--detail-accent); font-size: .62rem; }
.visit-selfie { width: min(220px, 100%); border: 1px solid var(--detail-border); border-radius: 8px; }
.visit-user { padding-top: .35rem; border-top: 1px solid var(--detail-border); }
.timeline { display: grid; }
.timeline-entry { position: relative; display: grid; grid-template-columns: 14px minmax(0, 1fr); gap: .6rem; padding-bottom: .8rem; }
.timeline-entry:not(:last-child)::before { content: ''; position: absolute; left: 6px; top: 14px; bottom: 0; width: 2px; background: #ead9db; }
.timeline-dot { width: 13px; height: 13px; margin-top: .1rem; border: 3px solid var(--detail-soft-strong); border-radius: 50%; background: var(--detail-accent); }
.timeline-entry > div { display: grid; gap: .08rem; }
.timeline-entry strong { font-size: .69rem; }
.timeline-entry small { color: var(--text-muted); font-size: .59rem; }
.timeline-entry p { margin: .2rem 0 0; color: var(--text-secondary); font-size: .67rem; line-height: 1.45; }
.compact-empty { min-height: 62px; display: flex; align-items: center; justify-content: center; gap: .45rem; padding: .65rem; border: 1px dashed var(--detail-border); border-radius: 8px; color: var(--text-muted); font-size: .68rem; }
.compact-empty i { color: var(--detail-accent); }

.section-toggle-btn {
  width: 100%; min-height: 34px; display: inline-flex; align-items: center; justify-content: center; gap: .4rem;
  padding: .45rem .6rem; border: 1px solid var(--detail-border); border-radius: 8px; background: #faf8f8;
  color: #625b5d; font-size: .65rem; font-weight: 700; cursor: pointer;
}
.section-toggle-btn:hover { border-color: var(--detail-red-border); background: var(--detail-soft); color: var(--detail-accent); }

/* Mobile sticky bar */
.detail-bottom-bar {
  position: fixed; left: 0; right: 0; bottom: 0; z-index: 1000;
  display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .45rem;
  padding: .65rem .8rem calc(.65rem + env(safe-area-inset-bottom));
  border-top: 1px solid var(--detail-border); background: rgba(255,255,255,.97);
  box-shadow: 0 -4px 14px rgba(15,23,42,.06); backdrop-filter: blur(10px);
}
.dbar-btn {
  min-width: 0; min-height: 43px; display: inline-flex; align-items: center; justify-content: center; gap: .35rem;
  padding: .55rem .35rem; border: 1px solid var(--detail-red-border); border-radius: 9px;
  background: #fff; color: var(--detail-accent); font-size: .67rem; font-weight: 800; text-decoration: none; cursor: pointer;
}
.dbar-navigate { background: var(--detail-accent); color: #fff; }
.dbar-call { border-color: #b8e7d0; background: #f0fdf7; color: #087a50; }
.dbar-checkin { background: var(--detail-soft); }
.dbar-disabled { opacity: .45; pointer-events: none; }

/* Expand overlay */
.expand-overlay {
  position: fixed; inset: 0; z-index: 5000; display: grid; place-items: center;
  padding: 1rem; background: rgba(24, 18, 20, .55); backdrop-filter: blur(4px);
}
.expand-dialog {
  width: min(920px, 96vw); max-height: min(820px, 92vh); overflow: hidden;
  display: grid; grid-template-rows: auto minmax(0, 1fr);
  border: 1px solid #eadfe1; border-radius: 14px; background: #fff;
  box-shadow: 0 28px 80px rgba(15,23,42,.25);
}
.expand-dialog-header {
  display: flex; align-items: center; justify-content: space-between; gap: 1rem;
  padding: .8rem 1rem; border-bottom: 1px solid var(--detail-border); background: #fffafa;
}
.expand-dialog-header > div { display: grid; gap: .1rem; }
.expand-dialog-header span { color: var(--detail-accent); font-size: .64rem; font-weight: 800; text-transform: uppercase; letter-spacing: .04em; }
.expand-dialog-header strong { font-size: .82rem; }
.expand-dialog-header button { width: 34px; height: 34px; display: grid; place-items: center; border: 1px solid var(--detail-border); border-radius: 9px; background: #fff; color: var(--detail-accent); cursor: pointer; }
.expand-dialog-body { min-height: 0; overflow: auto; padding: 1rem; }
.expand-dialog-body :deep(.pc-list) { max-height: none !important; overflow: visible !important; }

/* Desktop */
@media (min-width: 1200px) {
  .detail-page {
    padding-bottom: 1rem;
    min-height: 0;
  }


  /*
   * Desktop workspace:
   * left CRM/prospect content and right AI/business sidebar scroll independently.
   * The two stacks never stretch each other, so a short/empty card cannot create
   * a fake vertical gap in the opposite column.
   */
  .detail-content-grid {
    grid-template-columns: minmax(0, 1fr) minmax(300px, 350px);
    gap: .75rem;
    height: clamp(460px, calc(100dvh - 210px), 780px);
    min-height: 0;
    overflow: hidden;
    align-items: stretch;
  }

  .detail-column {
    min-height: 0;
    height: 100%;
    gap: .75rem;
    align-content: start;
    overflow-y: auto;
    overflow-x: hidden;
    overscroll-behavior-y: contain;
    scroll-behavior: smooth;
    scrollbar-gutter: stable;
    padding-right: .34rem;

    /* Firefox */
    scrollbar-width: thin;
    scrollbar-color: rgba(209, 67, 80, .70) rgba(209, 67, 80, .07);
  }

  /* Chromium / Edge / Safari: compact modern scroll indicator */
  .detail-column::-webkit-scrollbar {
    width: 6px;
  }

  .detail-column::-webkit-scrollbar-track {
    margin-block: .35rem;
    border-radius: 999px;
    background: rgba(209, 67, 80, .06);
  }

  .detail-column::-webkit-scrollbar-thumb {
    min-height: 42px;
    border: 1px solid rgba(255, 255, 255, .78);
    border-radius: 999px;
    background: rgba(209, 67, 80, .78);
  }

  .detail-column::-webkit-scrollbar-thumb:hover {
    background: rgba(184, 50, 64, .94);
  }

  .detail-column-main { padding-left: .02rem; }
  .detail-column-side { position: static; }

  .reviews-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .visit-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .detail-bottom-bar { display: none; }

  /*
   * The sidebar itself owns the scroll. Avoid a second large nested scrollbar
   * inside Discussion, which is confusing on desktop.
   */
  .detail-comments :deep(.pc-list) {
    max-height: none !important;
    overflow-y: visible !important;
  }
}

/* Tablet */
@media (min-width: 768px) and (max-width: 1199px) {
  .detail-page { padding-bottom: 1.5rem; }
  .detail-bottom-bar { display: none; }

  .hero-row { grid-template-columns: minmax(0, 1fr) auto; }
  .hero-status { grid-column: 2; grid-row: 1; margin-right: .1rem; }
  .hero-actions { grid-column: 1 / -1; justify-content: flex-start; }

  .prospect-detail-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .detail-field:nth-child(2n) { border-right: 0; }
  .detail-field:nth-child(3), .detail-field:nth-child(4) { border-bottom: 1px solid #eee6e7; }
  .google-layout { grid-template-columns: minmax(220px, 42%) minmax(0, 1fr); }

  /*
   * Tablet becomes one coherent work stream.
   * display:contents lets children from left/right stacks share one grid order.
   */
  .detail-content-grid {
    grid-template-columns: minmax(0, 1fr);
    gap: .7rem;
  }
  .detail-column { display: contents; }

  .dcard-prospect-core { order: 10; }
  .business-contact-card { order: 20; }
  .dcard-google-info { order: 30; }
  .dcard-photos { order: 40; }
  .dcard-menu { order: 41; }
  .ai-shell-summary { order: 50; }
  .detail-ai-menu { order: 51; }
  .ai-shell-discussion { order: 60; }
  .ai-shell-chat { order: 61; }
  .supporting-card { order: 70; }
  .dcard-reviews { order: 80; }
  .history-card { order: 90; }

  .ai-shell::after { width: 5.5rem; }
}

/* Mobile */
@media (max-width: 767px) {
  .detail-page {
    gap: .6rem;
    padding-bottom: calc(5.6rem + env(safe-area-inset-bottom));
  }

  .prospect-hero { padding-inline: 0; }
  .hero-row { grid-template-columns: minmax(0, 1fr) auto; gap: .5rem; }
  .hero-title-row h1 { font-size: 1.06rem; }
  .hero-status { grid-column: 2; grid-row: 1; }
  .hero-actions { display: none; }

  /*
   * Mobile is a single sales workflow, not desktop columns stacked blindly.
   * AI insight categories are intentionally kept ABOVE reviews/visit/status history.
   */
  .detail-content-grid {
    grid-template-columns: minmax(0, 1fr);
    gap: .62rem;
  }
  .detail-column { display: contents; }

  .dcard-prospect-core { order: 10; }
  .business-contact-card { order: 20; }
  .dcard-google-info { order: 30; }
  .dcard-photos { order: 40; }
  .dcard-menu { order: 41; }

  /* One AI category block after another */
  .ai-shell-summary { order: 50; }
  .detail-ai-menu { order: 51; }
  .ai-shell-discussion { order: 60; }
  .ai-shell-chat { order: 61; }

  .supporting-card { order: 70; }
  .dcard-reviews { order: 80; }
  .history-card { order: 90; }

  .prospect-detail-grid { grid-template-columns: minmax(0, 1fr); }
  .detail-field,
  .detail-field-wide {
    grid-column: auto;
    border-right: 0;
    border-bottom: 1px solid #eee6e7;
    padding: .62rem .68rem;
  }
  .detail-field:last-child { border-bottom: 0; }

  .google-layout { grid-template-columns: minmax(0, 1fr); }
  .google-rows > div {
    grid-template-columns: 16px minmax(92px, max-content) minmax(0, 1fr) auto;
  }

  .reviews-grid,
  .visit-grid { grid-template-columns: minmax(0, 1fr); }

  .business-list > div {
    grid-template-columns: minmax(96px, 38%) minmax(0, 1fr);
  }

  /* AI cards remain compact and visually grouped on mobile. */
  .ai-shell,
  .detail-ai-menu {
    width: 100%;
    min-width: 0;
  }
  .ai-shell::after {
    top: .35rem;
    right: .35rem;
    width: 2.25rem;
    height: 2.1rem;
  }
  .ai-shell > .expand-control {
    top: .47rem;
    right: .47rem;
    width: 29px;
    height: 29px;
    min-height: 29px;
    padding: 0;
    justify-content: center;
  }
  .ai-shell > .expand-control span { display: none; }

  .expand-dialog {
    width: 100%;
    max-height: 94vh;
    border-radius: 12px;
  }

  .dcard {
    padding: .78rem;
    border-radius: 10px;
  }

  .detail-page :deep(.ppg-scroll) {
    scroll-padding-inline: .2rem;
  }
}

@media (max-width: 390px) {
  .business-list > div { grid-template-columns: minmax(0, 1fr); gap: .12rem; }
  .google-rows > div { grid-template-columns: 16px minmax(0, 1fr); }
  .google-rows > div > strong,
  .google-rows > div > span,
  .google-rows > div > a { grid-column: 2; }
  .google-rows > div > .copy-mini { grid-column: 2; }
}
</style>
