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
