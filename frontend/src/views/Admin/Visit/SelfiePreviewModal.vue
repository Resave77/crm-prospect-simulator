<script setup lang="ts">
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

function getInitials(name: string) {
  return name.split(' ').filter(Boolean).slice(0, 2).map(w => w.charAt(0)).join('').toUpperCase()
}

function close() {
  emit('update:open', false)
}
</script>

<template>
  <Dialog :model-value="open" @update:model-value="emit('update:open', $event)" modal header="Visit Evidence" :style="{ width: '400px' }">
    <div v-if="visit" class="modal-evidence">
      <div class="avatar-large">{{ getInitials(visit.sales) }}</div>
      <span class="muted">{{ visit.sales }} &middot; {{ visit.customer }}</span>
    </div>
    <template #footer>
      <Button label="Close" severity="secondary" text @click="close" />
    </template>
  </Dialog>
</template>

<style scoped>
.modal-evidence {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem;
}

.avatar-large {
  width: 80px;
  height: 80px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: var(--brand-blue-bg, #fff1f2);
  color: var(--brand-blue);
  font-size: 1.4rem;
  font-weight: 800;
}
</style>
