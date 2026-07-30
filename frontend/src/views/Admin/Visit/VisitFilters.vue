<script setup lang="ts">
import { computed } from 'vue'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'

const props = defineProps<{
  date: string
  sales: string
  customer: string
  radius: string
  search: string
  salesOptions: string[]
  customerOptions: string[]
}>()

const emit = defineEmits<{
  'update:date': [value: string]
  'update:sales': [value: string]
  'update:customer': [value: string]
  'update:radius': [value: string]
  'update:search': [value: string]
  reset: []
}>()

const salesSelect = computed({
  get: () => props.sales,
  set: (val: string) => emit('update:sales', val),
})
const customerSelect = computed({
  get: () => props.customer,
  set: (val: string) => emit('update:customer', val),
})
</script>

<template>
  <div class="visit-filters">
    <div class="filter-search">
      <i class="pi pi-search" />
      <input
        type="text"
        placeholder="Search by sales, customer, location..."
        :value="search"
        @input="emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div class="filter-fields">
      <label class="filter-field">
        <span>Date</span>
        <input type="date" :value="date" @input="emit('update:date', ($event.target as HTMLInputElement).value)" />
      </label>
      <label class="filter-field">
        <span>Sales</span>
        <Select v-model="salesSelect" :options="salesOptions" placeholder="All Sales" :showClear="true" />
      </label>
      <label class="filter-field">
        <span>Customer</span>
        <Select v-model="customerSelect" :options="customerOptions" placeholder="All Customers" :showClear="true" />
      </label>
      <label class="filter-field">
        <span>Radius</span>
        <Select v-model="radius" :options="['Inside Radius', 'Outside Radius']" placeholder="All" :showClear="true" />
      </label>
      <div class="filter-action">
        <button class="reset-btn" type="button" @click="emit('reset')">
          <i class="pi pi-replay" /> Reset
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.visit-filters {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1rem 1.1rem;
  box-shadow: var(--shadow-xs);
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.5rem 0.75rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  margin-bottom: 0.75rem;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.filter-search:focus-within {
  background: var(--surface-card);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}

.filter-search i {
  color: var(--text-faint);
  font-size: 0.8rem;
}

.filter-search input {
  flex: 1;
  border: 0;
  outline: 0;
  background: transparent;
  font-size: 0.82rem;
  color: var(--text-primary);
}

.filter-search input::placeholder {
  color: var(--text-faint);
}

.filter-fields {
  display: grid;
  grid-template-columns: repeat(4, 1fr) auto;
  gap: 0.65rem;
  align-items: end;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-field > span {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

.filter-field input,
.filter-field :deep(.p-select) {
  font-size: 0.78rem;
}

.filter-action {
  display: flex;
  align-items: flex-end;
}

.reset-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-card);
  color: var(--text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.reset-btn:hover {
  background: var(--surface-hover);
  color: var(--brand-blue);
}

@media (max-width: 900px) {
  .filter-fields {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
