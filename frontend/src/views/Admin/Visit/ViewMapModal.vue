<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import type { VisitRecord } from '../../../data/visitMonitoring'

const props = defineProps<{
  open: boolean
  visit: VisitRecord | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const googleMapsUrl = computed(() => {
  if (!props.visit) return '#'
  return `https://www.google.com/maps?q=${props.visit.latitude},${props.visit.longitude}`
})

function close() {
  emit('update:open', false)
}
</script>

<template>
  <Dialog :model-value="open" @update:model-value="emit('update:open', $event)" modal header="GPS Location" :style="{ width: '420px' }">
    <div v-if="visit" class="modal-map">
      <div class="map-placeholder">
        <i class="pi pi-map-marker" />
        <strong>{{ visit.location }}</strong>
        <span>Distance from customer: {{ visit.distance }}</span>
      </div>
    </div>
    <template #footer>
      <Button label="Close" severity="secondary" text @click="close" />
      <a :href="googleMapsUrl" target="_blank" rel="noopener noreferrer">
        <Button label="Open Google Maps" icon="pi pi-external-link" />
      </a>
    </template>
  </Dialog>
</template>

<style scoped>
.modal-map .map-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem;
  border-radius: var(--radius-md);
  background: var(--surface-subtle);
  text-align: center;
}

.modal-map .map-placeholder i {
  font-size: 1.5rem;
  color: var(--brand-blue);
}

.modal-map .map-placeholder strong {
  font-size: 0.85rem;
  color: var(--text-primary);
}

.modal-map .map-placeholder span {
  font-size: 0.72rem;
  color: var(--text-muted);
}
</style>
