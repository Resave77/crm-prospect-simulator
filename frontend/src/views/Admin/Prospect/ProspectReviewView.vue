<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
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

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const review = ref<ProspectReview | null>(null)
const placeDetails = ref<PlaceDetails | null>(null)
const error = ref('')
const mapElement = ref<HTMLElement | null>(null)
const activePhotoIdx = ref(0)
const menuPhotoIdx = ref(0)
let map: L.Map | null = null

const menuPhotos = computed(() => {
  if (!placeDetails.value?.photos) return []
  return placeDetails.value.photos.filter((p) => /menu/i.test(p.name) || /menu/i.test(p.attribution))
})

const regularPhotos = computed(() => {
  if (!placeDetails.value?.photos) return []
  const menuNames = new Set(menuPhotos.value.map((p) => p.name))
  return placeDetails.value.photos.filter((p) => !menuNames.has(p.name))
})

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

        <article v-if="placeDetails?.photos?.length" class="section-card review-photos">
          <p class="eyebrow"><i class="pi pi-image" /> Menu</p>
          <template v-if="menuPhotos.length">
            <div class="review-photo-scroll">
              <div v-for="(photo, idx) in menuPhotos" :key="photo.name" class="review-photo-item" :class="{ active: idx === menuPhotoIdx }" @click="menuPhotoIdx = idx">
                <img :src="photo.photoUrl" :alt="`Menu ${idx + 1}`" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
              </div>
            </div>
            <p v-if="menuPhotos[menuPhotoIdx]?.attribution" class="review-photo-attr">Photo: {{ menuPhotos[menuPhotoIdx].attribution }}</p>
          </template>
          <div v-else class="review-menu-empty"><i class="pi pi-image" /><span>Menu not found</span></div>
        </article>

        <article v-if="placeDetails?.photos?.length" class="section-card review-photos">
          <p class="eyebrow"><i class="pi pi-images" /> Photos</p>
          <template v-if="regularPhotos.length">
            <div class="review-photo-scroll">
              <div v-for="(photo, idx) in regularPhotos" :key="photo.name" class="review-photo-item" :class="{ active: idx === activePhotoIdx }" @click="activePhotoIdx = idx">
                <img :src="photo.photoUrl" :alt="`Photo ${idx + 1}`" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
              </div>
            </div>
            <p v-if="regularPhotos[activePhotoIdx]?.attribution" class="review-photo-attr">Photo: {{ regularPhotos[activePhotoIdx].attribution }}</p>
          </template>
          <div v-else class="review-menu-empty"><i class="pi pi-images" /><span>No photos available</span></div>
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
