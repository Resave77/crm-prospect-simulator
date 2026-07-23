<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'

const crm = useCrmStore()
const router = useRouter()
const error = ref('')
onMounted(async () => {
  try { await crm.loadWonProspects() } catch (caught) { error.value = crm.errorMessage(caught) }
})
</script>

<template>
  <section class="admin-page">
    <div class="page-heading"><div><p class="eyebrow">Point 11</p><h1>Won Prospect Review</h1><p class="muted">Review qualified prospects before creating Customer Existing.</p></div><Tag :value="`${crm.wonProspects.length} waiting`" severity="success" /></div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div class="table-card">
      <div v-if="crm.loading" class="empty-state">Loading Won prospects…</div>
      <div v-else-if="!crm.wonProspects.length" class="empty-state"><i class="pi pi-inbox" /><strong>No prospects waiting</strong><span>Sales decisions marked WON will appear here.</span></div>
      <div v-else class="responsive-table">
        <table><thead><tr><th>Prospect</th><th>Google snapshot</th><th>Sales Executive</th><th>Status</th><th></th></tr></thead>
          <tbody><tr v-for="prospect in crm.wonProspects" :key="prospect.id"><td><strong>{{ prospect.placeName }}</strong><span>{{ prospect.placeCategory }} · {{ prospect.industryGroup }}</span></td><td>{{ prospect.formattedAddress }}</td><td>{{ prospect.assignedSalesExecutive }}</td><td><Tag value="WON" severity="success" /></td><td><Button label="Review" icon="pi pi-arrow-right" icon-pos="right" size="small" @click="router.push(`/admin/prospects/${prospect.id}/review`)" /></td></tr></tbody>
        </table>
      </div>
    </div>
  </section>
</template>
