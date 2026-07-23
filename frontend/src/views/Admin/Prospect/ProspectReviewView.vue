<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
onMounted(async () => {
  try { review.value = await getProspectReview(String(route.params.id)) } catch (caught) { error.value = crm.errorMessage(caught) }
})
</script>

<template>
  <section class="admin-page review-page">
    <Button label="Back to Won queue" icon="pi pi-arrow-left" severity="secondary" text @click="router.push('/admin/prospects/won')" />
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="review" class="review-grid">
      <article class="section-card">
        <div class="page-heading"><div><p class="eyebrow">Administrator review</p><h1>{{ review.prospect.placeName }}</h1></div><Tag :value="review.prospect.status" severity="success" /></div>
        <div class="source-grid"><div><span>Category</span><strong>{{ review.prospect.placeCategory }}</strong></div><div><span>Industry Group</span><strong>{{ review.prospect.industryGroup }}</strong></div><div><span>Google Place ID</span><strong>{{ review.prospect.googlePlaceId }}</strong></div><div class="wide"><span>Address</span><strong>{{ review.prospect.formattedAddress }}</strong></div><div><span>Sales Executive</span><strong>{{ review.prospect.assignedSalesExecutive }}</strong></div><div><span>Phone</span><strong>{{ review.prospect.phoneNumber || 'Not available' }}</strong></div></div>
        <div class="review-notes"><div><span>Visit notes</span><p>{{ review.prospect.visitNotes }}</p></div><div><span>Follow-up notes</span><p>{{ review.prospect.followUpNotes }}</p></div></div>
        <Button v-if="review.prospect.status === 'WON'" label="Open conversion form" icon="pi pi-arrow-right" icon-pos="right" @click="router.push(`/admin/prospects/${review.prospect.id}/convert`)" />
      </article>
      <aside class="section-card"><p class="eyebrow">Status history</p><div class="timeline"><div v-for="item in review.history" :key="item.id"><i class="pi pi-circle-fill" /><div><strong>{{ item.fromStatus || 'START' }} → {{ item.toStatus }}</strong><span>{{ item.changedByName }} · {{ new Date(item.createdAt).toLocaleString() }}</span><p>{{ item.notes }}</p></div></div></div></aside>
    </div>
  </section>
</template>
