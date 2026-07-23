<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { getProspectReview } from '../../../api/crm'
import { useCrmStore } from '../../../stores/crm'
import type { ProspectReview } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const review = ref<ProspectReview | null>(null)
const error = ref('')
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
    review.value = await getProspectReview(String(route.params.id))
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
    <Button label="Back to Won queue" icon="pi pi-arrow-left" severity="secondary" text @click="router.push('/admin/prospects/won')" />
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div v-if="review" class="review-grid">
      <div class="review-main">
        <article class="section-card">
          <div class="page-heading"><div><p class="eyebrow">Administrator review</p><h1>{{ review.prospect.placeName }}</h1></div><Tag :value="review.prospect.status" severity="success" /></div>
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

        <div v-if="review.prospect.status === 'WON'" class="review-convert-bar">
          <div><p class="eyebrow">Ready to convert</p><p class="muted">This Google snapshot will be pre-filled into the conversion form.</p></div>
          <Button label="Open Conversion Form" icon="pi pi-arrow-right" icon-pos="right" @click="router.push(`/admin/prospects/${review.prospect.id}/convert`)" />
        </div>
      </div>

      <aside class="section-card review-sidebar">
        <p class="eyebrow">Status history</p>
        <div class="timeline">
          <div v-for="item in review.history" :key="item.id"><i class="pi pi-circle-fill" /><div><strong>{{ item.fromStatus || 'START' }} → {{ item.toStatus }}</strong><span>{{ item.changedByName }} · {{ new Date(item.createdAt).toLocaleString() }}</span><p>{{ item.notes }}</p></div></div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.review-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(280px, 0.75fr);
  gap: 1.25rem;
}

.section-card {
  padding: clamp(1.25rem, 3vw, 2rem);
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

.timeline { display: grid; gap: 0.85rem; }
.timeline > div {
  display: grid;
  grid-template-columns: 14px 1fr;
  gap: 0.6rem;
}

.timeline i {
  margin-top: 0.3rem;
  color: var(--brand-blue);
  font-size: 0.45rem;
}

.timeline span {
  display: block;
  margin-top: 0.15rem;
  color: var(--text-muted);
  font-size: 0.72rem;
}

.timeline p {
  margin: 0.3rem 0 0;
  color: #52615d;
  font-size: 0.82rem;
  line-height: 1.5;
}

.review-main {
  display: grid;
  gap: 1.25rem;
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
  padding: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  border: 2px solid #d1fae5;
  border-radius: var(--radius-xl);
  background: linear-gradient(135deg, #ecfdf5, #f0fdf4);
}

.review-convert-bar :deep(.eyebrow) {
  margin-bottom: 0.15rem;
  color: #059669;
}

.review-convert-bar :deep(.muted) {
  margin: 0;
  font-size: 0.78rem;
}

.review-sidebar {
  position: sticky;
  top: 1rem;
  align-self: start;
}

@media (max-width: 900px) {
  .review-grid { grid-template-columns: 1fr; }
  .review-convert-bar { flex-direction: column; align-items: flex-start; }
}

@media (max-width: 560px) {
  .review-grid { grid-template-columns: 1fr; }
}
</style>
