<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import VisitSuccessCard from '../../../components/sales/visit/VisitSuccessCard.vue'
import {
  isValidEntityType,
  normalizeRouteId,
  fetchProspectVisitData,
  fetchCustomerVisitData,
} from '../../../utils/visitEntity'

const route = useRoute()

const entityType = computed(() => {
  if (isValidEntityType(route.meta.entityType)) return route.meta.entityType
  return null
})
const entityId = computed(() => normalizeRouteId(route.params.id))

const entityName = ref('')
const loading = ref(true)

const detailRoute = computed(() => {
  if (entityType.value === 'customer') {
    return { name: 'SalesCustomerDetail', params: { id: entityId.value } }
  }
  return { name: 'SalesProspectDetail', params: { id: entityId.value } }
})

const pipelineRoute = computed(() => ({
  path: '/sales/pipeline',
  query: { prospectId: entityId.value, action: 'update' },
}))

const historyRoute = computed(() => ({ name: 'SalesHistory' }))

const isProspect = computed(() => entityType.value === 'prospect')

onMounted(async () => {
  try {
    if (entityType.value === 'prospect') {
      const { entity } = await fetchProspectVisitData(entityId.value)
      entityName.value = entity?.name ?? ''
    } else if (entityType.value === 'customer') {
      const { entity } = await fetchCustomerVisitData(entityId.value)
      entityName.value = entity?.name ?? ''
    }
  } catch {
    entityName.value = ''
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="success-page">
    <div v-if="loading" class="success-skeleton">
      <div class="sk-circle" />
      <div class="sk-line w50" />
      <div class="sk-line w70" />
    </div>
    <VisitSuccessCard
      v-else
      icon="pi pi-check-circle"
      title="Check-out successful"
      message="The visit outcome and completion time have been saved."
      :primary-label="isProspect ? 'Update Sales Pipeline' : 'View Visit History'"
      :primary-to="isProspect ? pipelineRoute : historyRoute"
      secondary-label="Back to Detail"
      :secondary-to="detailRoute"
    />
  </section>
</template>

<style scoped>
.success-page {
  display: flex; justify-content: center; padding-top: 1rem;
  padding-bottom: calc(68px + env(safe-area-inset-bottom) + 1rem);
}
.success-skeleton {
  display: flex; flex-direction: column; align-items: center; gap: 0.75rem;
  padding: 2.5rem 1.5rem; background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); width: 100%;
}
.sk-circle { width: 72px; height: 72px; border-radius: 50%; background: #e2e8f0; animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-line { height: 14px; border-radius: 7px; background: #e2e8f0; animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-line.w50 { width: 50%; }
.sk-line.w70 { width: 70%; }
@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
</style>
