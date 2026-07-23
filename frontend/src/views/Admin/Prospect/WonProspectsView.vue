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

function createdAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  return `${days}d ago`
}
</script>

<template>
  <section class="admin-page">
    <div class="page-heading"><div><p class="eyebrow">Point 11</p><h1>Won Prospect Review</h1><p class="muted">Review qualified Google snapshots before creating a Customer Existing record.</p></div><Tag :value="`${crm.wonProspects.length} waiting`" severity="success" /></div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div class="table-card">
      <div v-if="crm.loading" class="empty-state">Loading Won prospects…</div>
      <div v-else-if="!crm.wonProspects.length" class="empty-state"><i class="pi pi-inbox" /><strong>No prospects waiting</strong><span>Sales decisions marked WON will appear here.</span></div>
      <div v-else class="responsive-table">
        <table><thead><tr><th>Place Name</th><th>Category</th><th>Phone</th><th>Sales Executive</th><th>Won</th><th></th></tr></thead>
          <tbody><tr v-for="prospect in crm.wonProspects" :key="prospect.id">
            <td><strong>{{ prospect.placeName }}</strong><span>{{ prospect.formattedAddress }}</span></td>
            <td><Tag :value="prospect.placeCategory" severity="info" /><span>{{ prospect.industryGroup }}</span></td>
            <td><span>{{ prospect.phoneNumber || '—' }}</span></td>
            <td><span>{{ prospect.assignedSalesExecutive }}</span></td>
            <td><span>{{ createdAgo(prospect.updatedAt) }}</span></td>
            <td><Button label="Review & Convert" icon="pi pi-arrow-right" icon-pos="right" size="small" @click="router.push(`/admin/prospects/${prospect.id}/review`)" /></td>
          </tr></tbody>
        </table>
      </div>
    </div>
  </section>
</template>
