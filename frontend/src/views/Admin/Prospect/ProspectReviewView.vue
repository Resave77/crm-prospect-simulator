<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getProspectReview, getProspectPlaceDetails } from '../../../api/crm'
import { useCrmStore } from '../../../stores/crm'
import type { ProspectReview, PlaceDetails } from '../../../types/crm'
import ProspectComments from '../../../components/ProspectComments.vue'
import PlacePhotoGallery from '../../../components/PlacePhotoGallery.vue'
import { websiteDisplayUrl } from '../../../utils/placeDetails'
import { stars, priceLevelLabel, priceLevelSeverity, businessStatusLabel, businessStatusSeverity, utcOffsetLabel } from '../../../utils/placeLabels'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const review = ref<ProspectReview | null>(null)
const placeDetails = ref<PlaceDetails | null>(null)
const error = ref('')
const showAllHours = ref(false)
const mapElement = ref<HTMLElement | null>(null)
let map: L.Map | null = null

function renderMap() {
  const p = review.value?.prospect
  if (!mapElement.value || p?.latitude == null || p?.longitude == null) return
  map?.remove()
  map = L.map(mapElement.value).setView([p.latitude, p.longitude], 16)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution: '&copy; OpenStreetMap' }).addTo(map)
  L.marker([p.latitude, p.longitude]).addTo(map).bindPopup(p.placeName).openPopup()
}

onMounted(async () => {
  try {
    const prospectId = String(route.params.id)
    const [reviewData, placeData] = await Promise.all([
      getProspectReview(prospectId),
      getProspectPlaceDetails(prospectId, 'ADMINISTRATOR').catch(() => null),
    ])
    review.value = reviewData
    placeDetails.value = placeData
    await nextTick()
    renderMap()
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  }
})

onBeforeUnmount(() => map?.remove())
</script>

<template>
  <section class="admin-page review-page">
    <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.back()" title="Back" />
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div v-if="review" class="review-grid">
      <div class="review-main">
        <article class="section-card">
          <div class="page-heading"><div><p class="eyebrow">Administrator review</p><h1>{{ review.prospect.placeName }}</h1></div><div class="page-heading-tags"><Tag :value="review.prospect.status" :severity="review.prospect.status === 'WON' ? 'success' : review.prospect.status === 'LOST' ? 'danger' : 'warn'" /><Tag v-if="review.prospect.status === 'WON'" value="Ready for conversion" severity="success" /></div></div>
          <div class="source-grid">
            <div><span>Category</span><strong>{{ review.prospect.placeCategory }}</strong></div>
            <div><span>Industry Group</span><strong>{{ review.prospect.industryGroup }}</strong></div>
            <div class="wide"><span>Formatted Address</span><strong>{{ review.prospect.formattedAddress }}</strong></div>
            <div><span>Phone</span><strong>{{ review.prospect.phoneNumber || 'Not available' }}</strong></div>
            <div><span>Website</span><strong><a v-if="review.prospect.websiteUrl" :href="review.prospect.websiteUrl" target="_blank" rel="noopener">{{ review.prospect.websiteUrl }}</a><span v-else>Not available</span></strong></div>
            <div><span>Google Place ID</span><strong class="mono">{{ review.prospect.googlePlaceId }}</strong></div>
            <div><span>Sales Executive</span><strong>{{ review.prospect.assignedSalesExecutive }}</strong></div>
            <div><span>Coordinates</span><strong v-if="review.prospect.latitude != null">{{ review.prospect.latitude.toFixed(6) }}, {{ review.prospect.longitude?.toFixed(6) }}</strong><strong v-else>Not available</strong></div>
          </div>
          <div v-if="review.prospect.placeTypes.length" class="tag-row" style="margin-top:0.85rem">
            <Tag v-for="t in review.prospect.placeTypes" :key="t" :value="t" severity="secondary" />
          </div>
        </article>

        <article v-if="placeDetails" class="section-card review-google">
          <p class="eyebrow"><i class="pi pi-map" /> Google Maps data</p>
          <p v-if="placeDetails.editorialSummary" class="review-editorial">{{ placeDetails.editorialSummary }}</p>

          <div class="review-google-top">
            <div v-if="placeDetails.rating > 0" class="review-rating">
              <span class="review-rating-num">{{ placeDetails.rating.toFixed(1) }}</span>
              <span class="review-stars">
                <i v-for="(s, i) in stars(placeDetails.rating)" :key="i" :class="['pi', s]" />
              </span>
              <span class="muted">({{ placeDetails.userRatingCount.toLocaleString() }} reviews)</span>
            </div>
            <div class="review-badges">
              <Tag v-if="placeDetails.businessStatus" :value="businessStatusLabel(placeDetails.businessStatus)" :severity="businessStatusSeverity(placeDetails.businessStatus)" />
              <Tag v-if="placeDetails.priceLevel" :value="priceLevelLabel(placeDetails.priceLevel)" :severity="priceLevelSeverity(placeDetails.priceLevel)" />
            </div>
          </div>

          <div class="source-grid">
            <div class="wide"><span>Address</span><strong>{{ placeDetails.formattedAddress }}</strong></div>
            <div><span>Phone</span><strong>{{ placeDetails.phoneNumber || '—' }}</strong></div>
            <div><span>International phone</span><strong>{{ placeDetails.internationalPhone || '—' }}</strong></div>
            <div><span>Website</span><strong><a v-if="placeDetails.websiteUrl" :href="placeDetails.websiteUrl" target="_blank" rel="noopener" class="review-link">{{ websiteDisplayUrl(placeDetails.websiteUrl) }}</a><span v-else>—</span></strong></div>
            <div><span>Google Maps</span><strong><a v-if="placeDetails.googleMapsUrl" :href="placeDetails.googleMapsUrl" target="_blank" rel="noopener" class="review-link">View listing <i class="pi pi-external-link" /></a><span v-else>—</span></strong></div>
            <div><span>Time zone</span><strong>{{ utcOffsetLabel(placeDetails.utcOffsetMinutes) }}</strong></div>
            <div><span>Google Place ID</span><strong class="mono review-placeid">{{ placeDetails.googlePlaceId }}</strong></div>
            <div><span>Coordinates</span><strong v-if="placeDetails.latitude != null">{{ placeDetails.latitude.toFixed(6) }}, {{ placeDetails.longitude?.toFixed(6) }}</strong><strong v-else>—</strong></div>
          </div>

          <div v-if="placeDetails.placeTypes?.length" class="tag-row" style="margin-top:0.85rem">
            <span class="review-tags-label">Categories:</span>
            <Tag v-for="t in placeDetails.placeTypes.slice(0, 8)" :key="t" :value="t.replace(/_/g, ' ')" severity="secondary" />
          </div>

          <div v-if="placeDetails.openingHours" class="review-block">
            <div class="review-block-title">
              <strong>Opening hours</strong>
              <span :class="['review-hours-dot', placeDetails.openingHours.openNow ? 'open' : 'closed']" />
              <span>{{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}</span>
            </div>
            <div v-if="placeDetails.openingHours.weekdays?.length" class="review-hours-list">
              <div v-for="(day, i) in (showAllHours ? placeDetails.openingHours.weekdays : placeDetails.openingHours.weekdays.slice(0, 4))" :key="i" class="review-hours-row" v-html="day" />
              <button v-if="placeDetails.openingHours.weekdays.length > 4" class="review-toggle" @click="showAllHours = !showAllHours">
                {{ showAllHours ? 'Show less' : `Show all ${placeDetails.openingHours.weekdays.length} days` }}
              </button>
            </div>
          </div>

          <div v-if="placeDetails.delivery || placeDetails.dineIn || placeDetails.takeout || placeDetails.curbsidePickup" class="review-block">
            <div class="review-block-title"><strong>Service options</strong></div>
            <div class="review-chips">
              <span v-if="placeDetails.dineIn" class="review-chip"><i class="pi pi-check" /> Dine In</span>
              <span v-if="placeDetails.takeout" class="review-chip"><i class="pi pi-check" /> Takeout</span>
              <span v-if="placeDetails.delivery" class="review-chip"><i class="pi pi-check" /> Delivery</span>
              <span v-if="placeDetails.curbsidePickup" class="review-chip"><i class="pi pi-check" /> Curbside Pickup</span>
            </div>
          </div>

          <div v-if="placeDetails?.parkingOptions || placeDetails?.paymentOptions || placeDetails?.accessibilityOptions" class="review-block">
            <div class="review-block-title"><strong>Amenities</strong></div>
            <div v-if="placeDetails?.parkingOptions" class="review-chips">
              <span v-if="placeDetails.parkingOptions.freeParkingLot" class="review-chip"><i class="pi pi-check" /> Free Lot</span>
              <span v-if="placeDetails.parkingOptions.freeStreetParking" class="review-chip"><i class="pi pi-check" /> Free Street</span>
              <span v-if="placeDetails.parkingOptions.paidParkingLot" class="review-chip"><i class="pi pi-check" /> Paid Lot</span>
              <span v-if="placeDetails.parkingOptions.paidStreetParking" class="review-chip"><i class="pi pi-check" /> Paid Street</span>
              <span v-if="placeDetails.parkingOptions.garageParking" class="review-chip"><i class="pi pi-check" /> Garage</span>
              <span v-if="placeDetails.parkingOptions.valetParking" class="review-chip"><i class="pi pi-check" /> Valet</span>
              <span v-if="placeDetails.paymentOptions?.cashOnly" class="review-chip"><i class="pi pi-check" /> Cash</span>
              <span v-if="placeDetails.paymentOptions?.creditCardOnly" class="review-chip"><i class="pi pi-check" /> Credit Card</span>
              <span v-if="placeDetails.paymentOptions?.debitCardOnly" class="review-chip"><i class="pi pi-check" /> Debit Card</span>
              <span v-if="placeDetails.paymentOptions?.nfcOnly" class="review-chip"><i class="pi pi-check" /> NFC</span>
              <span v-if="placeDetails.accessibilityOptions?.wheelchairAccessibleEntrance" class="review-chip"><i class="pi pi-check" /> Wheelchair Entrance</span>
              <span v-if="placeDetails.accessibilityOptions?.wheelchairAccessibleParking" class="review-chip"><i class="pi pi-check" /> Wheelchair Parking</span>
              <span v-if="placeDetails.accessibilityOptions?.wheelchairAccessibleRestroom" class="review-chip"><i class="pi pi-check" /> Wheelchair Restroom</span>
              <span v-if="placeDetails.accessibilityOptions?.wheelchairAccessibleSeating" class="review-chip"><i class="pi pi-check" /> Wheelchair Seating</span>
            </div>
          </div>

          <div v-if="placeDetails.reviews?.length" class="review-block">
            <div class="review-block-title"><strong>Reviews ({{ placeDetails.reviews.length }})</strong></div>
            <div class="review-reviews">
              <div v-for="(rv, i) in placeDetails.reviews.slice(0, 5)" :key="i" class="review-review">
                <div class="review-review-head">
                  <img v-if="rv.authorPhoto" :src="rv.authorPhoto" :alt="rv.authorName" class="review-review-avatar" @error="($event.target as HTMLImageElement).style.display='none'" />
                  <div v-else class="review-review-avatar-placeholder">{{ rv.authorName?.charAt(0) || '?' }}</div>
                  <div class="review-review-meta">
                    <strong>{{ rv.authorName }}</strong>
                    <div class="review-review-stars">
                      <i v-for="(s, j) in stars(rv.rating)" :key="j" :class="['pi', s]" />
                      <span class="muted">{{ rv.time }}</span>
                    </div>
                  </div>
                </div>
                <p v-if="rv.text">{{ rv.text }}</p>
              </div>
            </div>
          </div>
        </article>

        <article v-if="placeDetails?.photos?.length" class="section-card review-photos">
          <p class="eyebrow"><i class="pi pi-images" /> Photos</p>
          <PlacePhotoGallery :photos="placeDetails.photos" :prospect-id="review.prospect.id" role="ADMINISTRATOR" />
        </article>

        <article v-if="review.prospect.latitude != null && review.prospect.longitude != null" class="section-card">
          <p class="eyebrow">Location snapshot</p>
          <div ref="mapElement" class="review-map" />
        </article>

        <article class="section-card" v-if="review.prospect.visitNotes || review.prospect.followUpNotes">
          <p class="eyebrow">Sales notes</p>
          <div class="review-notes">
            <div v-if="review.prospect.visitNotes"><span>Visit notes</span><p>{{ review.prospect.visitNotes }}</p></div>
            <div v-if="review.prospect.followUpNotes"><span>Follow-up notes</span><p>{{ review.prospect.followUpNotes }}</p></div>
          </div>
        </article>

        <article v-if="review.visits.length" class="section-card">
          <p class="eyebrow">Visit history</p>
          <div class="visit-list">
            <div v-for="v in review.visits" :key="v.id" class="visit-item">
              <div class="visit-header">
                <strong>{{ v.salesExecutiveName }}</strong>
                <span>{{ new Date(v.checkInAt).toLocaleString() }}<template v-if="v.checkOutAt"> → {{ new Date(v.checkOutAt).toLocaleString() }}</template></span>
              </div>
              <p v-if="v.visitNotes">{{ v.visitNotes }}</p>
              <p v-if="v.followUpNotes" class="muted">{{ v.followUpNotes }}</p>
            </div>
          </div>
        </article>

        <ProspectComments :prospect-id="review.prospect.id" role="ADMINISTRATOR" />

      </div>

      <aside class="review-sidebar">
        <article class="section-card">
          <p class="eyebrow">Status history</p>
          <div class="timeline">
            <div v-for="item in review.history" :key="item.id"><i class="pi pi-circle-fill" /><div><strong>{{ item.fromStatus || 'START' }} → {{ item.toStatus }}</strong><span>{{ item.changedByName }} · {{ new Date(item.createdAt).toLocaleString() }}</span><p>{{ item.notes }}</p></div></div>
          </div>
        </article>
        <div v-if="review.prospect.status === 'WON'" class="review-convert-bar">
          <div class="review-convert-info">
            <div class="review-convert-icon"><i class="pi pi-trophy" /></div>
            <div>
              <p class="review-convert-title">Prospect Won!</p>
              <p class="review-convert-desc">Ready to convert into a customer.</p>
            </div>
          </div>
          <Button label="Convert" icon="pi pi-arrow-right" icon-pos="right" @click="router.push(`/admin/prospects/${review.prospect.id}/convert`)" />
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.review-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 1.25rem;
}

.review-page {
  overflow-x: hidden;
}

.section-card {
  padding: clamp(1.25rem, 3vw, 2rem);
  overflow: hidden;
  min-width: 0;
}

.source-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.source-grid > div {
  padding: 0.9rem;
  border-radius: var(--radius-md);
  background: var(--surface-subtle);
}

.source-grid .wide { grid-column: 1 / -1; }

.source-grid span {
  display: block;
  margin-bottom: 0.3rem;
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.review-photos { border: 1px solid #e0e7ff; }
.review-photos .eyebrow { display: flex; align-items: center; gap: 0.4rem; }
.review-photos .eyebrow i { color: var(--brand-blue); font-size: 0.75rem; }
.review-photo-scroll {
  display: flex; gap: 0.5rem; overflow-x: auto; scroll-snap-type: x mandatory;
  -webkit-overflow-scrolling: touch; padding-bottom: 0.3rem;
}
.review-photo-scroll::-webkit-scrollbar { height: 4px; }
.review-photo-scroll::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; }
.review-photo-item {
  flex: 0 0 200px; height: 150px; border-radius: 12px; overflow: hidden;
  cursor: pointer; scroll-snap-align: start; border: 2px solid transparent;
  transition: border-color 0.15s ease, transform 0.15s ease;
}
.review-photo-item:hover { transform: scale(1.02); }
.review-photo-item.active { border-color: var(--brand-blue); }
.review-photo-item img { width: 100%; height: 100%; object-fit: cover; }
.review-photo-attr { margin: 0.5rem 0 0; color: var(--text-muted); font-size: 0.62rem; font-style: italic; }
.review-menu-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  padding: 2rem 1rem; color: var(--text-muted); text-align: center;
  background: var(--surface-subtle); border-radius: 12px;
}
.review-menu-empty i { font-size: 1.5rem; color: #cbd5e1; }
.review-menu-empty span { font-size: 0.82rem; font-weight: 600; }

.review-google { border: 1px solid #e0e7ff; }
.review-google .eyebrow { display: flex; align-items: center; gap: 0.4rem; }
.review-google .eyebrow i { color: var(--brand-blue); font-size: 0.75rem; }
.review-editorial { margin: 0 0 1rem; color: var(--text-secondary); font-size: 0.84rem; line-height: 1.55; font-style: italic; }
.review-google-top { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.6rem; margin-bottom: 0.85rem; }
.review-rating { display: flex; align-items: center; gap: 0.45rem; }
.review-rating-num { font-size: 1.05rem; font-weight: 800; color: #f59e0b; }
.review-rating .review-stars { display: flex; gap: 1px; }
.review-rating .review-stars .pi { font-size: 0.65rem; color: #f59e0b; }
.review-badges { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.review-link { color: var(--brand-blue); text-decoration: none; font-weight: 600; display: inline-flex; align-items: center; gap: 0.3rem; }
.review-link:hover { text-decoration: underline; }
.review-link i { font-size: 0.6rem; }
.review-placeid { word-break: break-all; font-size: 0.72rem; }
.review-tags-label { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; margin-right: 0.25rem; }
.review-block { margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border-light); }
.review-block-title { display: flex; align-items: center; gap: 0.4rem; margin-bottom: 0.5rem; font-size: 0.78rem; }
.review-block-title strong { color: var(--text-primary); text-transform: uppercase; font-size: 0.68rem; letter-spacing: 0.06em; }
.review-hours-dot { width: 8px; height: 8px; border-radius: 50%; }
.review-hours-dot.open { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.review-hours-dot.closed { background: #ef4444; }
.review-hours-list { display: grid; gap: 0.3rem; }
.review-hours-row { font-size: 0.8rem; color: var(--text-secondary); line-height: 1.4; }
.review-toggle { background: none; border: none; color: var(--brand-blue); font-size: 0.72rem; font-weight: 600; cursor: pointer; padding: 0.2rem 0; text-align: left; }
.review-toggle:hover { text-decoration: underline; }
.review-chips { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.review-chip {
  display: inline-flex; align-items: center; gap: 0.2rem;
  padding: 0.2rem 0.5rem; border-radius: 9999px;
  background: #f0fdf4; color: #059669;
  font-size: 0.62rem; font-weight: 600;
}
.review-chip i { font-size: 0.5rem; }
.review-reviews { display: grid; gap: 0.75rem; }
.review-review { padding-bottom: 0.65rem; border-bottom: 1px solid var(--border-light); }
.review-review:last-child { border-bottom: none; padding-bottom: 0; }
.review-review-head { display: flex; align-items: center; gap: 0.55rem; }
.review-review-avatar { width: 30px; height: 30px; border-radius: 50%; object-fit: cover; }
.review-review-avatar-placeholder {
  width: 30px; height: 30px; border-radius: 50%; display: grid; place-items: center;
  background: #e2e8f0; color: var(--text-muted); font-size: 0.7rem; font-weight: 700; flex-shrink: 0;
}
.review-review-meta { flex: 1; min-width: 0; }
.review-review-meta strong { font-size: 0.75rem; color: var(--text-primary); }
.review-review-stars { display: flex; align-items: center; gap: 1px; }
.review-review-stars .pi { font-size: 0.55rem; color: #f59e0b; }
.review-review p { margin: 0.3rem 0 0; color: var(--text-secondary); font-size: 0.78rem; line-height: 1.5; }

.review-notes {
  margin: 1rem 0;
  display: grid;
  gap: 0.85rem;
}

.review-notes > div {
  padding: 0.9rem;
  border-radius: var(--radius-md);
  background: var(--surface-subtle);
}

.review-notes span {
  display: block;
  margin-bottom: 0.3rem;
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.review-notes p { margin: 0; line-height: 1.6; }

.review-main {
  display: grid;
  gap: 1.25rem;
}

.page-heading-tags {
  display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;
}

.review-map {
  height: 280px;
  margin-top: 0.5rem;
  border-radius: var(--radius-md);
  overflow: hidden;
  background: #e8eef5;
}

.visit-list {
  display: grid;
  gap: 0.65rem;
  margin-top: 0.5rem;
}

.visit-item {
  padding: 0.75rem;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  background: var(--surface-subtle);
}

.visit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.3rem;
}

.visit-header strong {
  font-size: 0.78rem;
}

.visit-header span {
  color: var(--text-muted);
  font-size: 0.65rem;
}

.visit-item p {
  margin: 0.2rem 0 0;
  font-size: 0.8rem;
  line-height: 1.5;
}

.review-convert-bar {
  padding: 1.25rem 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  border: 1px solid #a7f3d0;
  border-radius: var(--radius-xl);
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdf4 50%, #dcfce7 100%);
  box-shadow: 0 1px 3px rgba(5, 150, 105, 0.08);
}

.review-convert-info {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.review-convert-icon {
  width: 44px; height: 44px;
  display: grid; place-items: center; flex-shrink: 0;
  border-radius: 12px; background: #059669; color: #fff;
  font-size: 1.1rem;
}

.review-convert-title {
  margin: 0; font-size: 0.92rem; font-weight: 700; color: #065f46;
}

.review-convert-desc {
  margin: 0.15rem 0 0; font-size: 0.75rem; color: #047857; line-height: 1.45;
}

.review-sidebar {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  position: sticky;
  top: 1rem;
  align-self: start;
}

.review-sidebar .timeline { display: grid; gap: 0.65rem; }
.review-sidebar .timeline > div {
  display: grid; grid-template-columns: 12px 1fr; gap: 0.5rem;
}
.review-sidebar .timeline i {
  margin-top: 0.35rem; color: var(--brand-blue); font-size: 0.4rem;
}
.review-sidebar .timeline strong {
  font-size: 0.72rem; text-transform: capitalize;
}
.review-sidebar .timeline span {
  display: block; margin-top: 0.1rem; color: var(--text-muted); font-size: 0.65rem;
}
.review-sidebar .timeline p {
  margin: 0.15rem 0 0; color: #52615d; font-size: 0.75rem; line-height: 1.45;
}

.review-sidebar .review-convert-bar {
  flex-direction: column; align-items: stretch;
}

.review-page {
  overflow-x: hidden;
}

@media (max-width: 900px) {
  .review-grid { grid-template-columns: 1fr; }
  .review-sidebar { position: static; }
  .review-convert-bar { flex-direction: column; align-items: stretch; }
  .review-convert-info { gap: 0.65rem; }
}

@media (max-width: 560px) {
  .review-grid { grid-template-columns: 1fr; }
}
</style>
