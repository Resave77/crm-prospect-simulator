<script setup lang="ts">
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import type { VisitRecord } from '../../../data/visitMonitoring'

defineProps<{
  visits: VisitRecord[]
  sortKey: string
  sortDirection: string
}>()

const emit = defineEmits<{
  sort: [key: string]
  'view-map': [visit: VisitRecord]
  selfie: [visit: VisitRecord]
}>()

function sortIcon(key: string, currentKey: string, direction: string) {
  if (key !== currentKey) return 'pi pi-sort-alt'
  return direction === 'asc' ? 'pi pi-sort-amount-up' : 'pi pi-sort-amount-down'
}
</script>

<template>
  <div class="table-scroll">
    <table class="data-table">
      <thead>
        <tr>
          <th class="sortable" @click="emit('sort', 'sales')">
            Sales <i :class="sortIcon('sales', sortKey, sortDirection)" />
          </th>
          <th class="sortable" @click="emit('sort', 'customer')">
            Customer <i :class="sortIcon('customer', sortKey, sortDirection)" />
          </th>
          <th class="sortable" @click="emit('sort', 'checkIn')">
            Check In <i :class="sortIcon('checkIn', sortKey, sortDirection)" />
          </th>
          <th>Check Out</th>
          <th class="sortable" @click="emit('sort', 'duration')">
            Duration <i :class="sortIcon('duration', sortKey, sortDirection)" />
          </th>
          <th class="sortable" @click="emit('sort', 'distance')">
            Distance <i :class="sortIcon('distance', sortKey, sortDirection)" />
          </th>
          <th>Radius</th>
          <th>Result</th>
          <th class="th-action">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="visit in visits" :key="visit.id">
          <td><span class="cell-primary">{{ visit.sales }}</span></td>
          <td>
            <div class="cell-stack">
              <span class="cell-primary">{{ visit.customer }}</span>
              <span class="cell-sub">{{ visit.location }}</span>
            </div>
          </td>
          <td><span class="cell-text">{{ visit.checkIn }}</span></td>
          <td><span class="cell-text">{{ visit.checkOut ?? '—' }}</span></td>
          <td><span class="cell-text">{{ visit.duration }}</span></td>
          <td><span class="cell-text">{{ visit.distance }}</span></td>
          <td>
            <Tag
              :value="visit.radiusStatus"
              :severity="visit.radiusStatus === 'Inside Radius' ? 'success' : 'warn'"
              size="small"
            />
          </td>
          <td>
            <Tag
              :value="visit.result"
              :severity="visit.result === 'Order Placed' ? 'success' : visit.result === 'No Answer' ? 'danger' : 'info'"
              size="small"
            />
          </td>
          <td class="td-action">
            <div class="row-actions">
              <Button icon="pi pi-map-marker" text rounded size="small" class="act-view" title="View Map" @click="emit('view-map', visit)" />
              <Button v-if="visit.selfie" icon="pi pi-camera" text rounded size="small" class="act-edit" title="View Evidence" @click="emit('selfie', visit)" />
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-scroll {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.82rem;
}

.data-table thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--surface-subtle);
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.7rem 0.85rem;
  border-bottom: 1px solid var(--border-light);
  white-space: nowrap;
  text-align: left;
  user-select: none;
}

.data-table thead th.sortable {
  cursor: pointer;
  transition: color var(--transition-fast);
}

.data-table thead th.sortable:hover {
  color: var(--brand-blue);
}

.data-table thead th i {
  margin-left: 0.3rem;
  font-size: 0.6rem;
}

.data-table tbody td {
  padding: 0.7rem 0.85rem;
  border-bottom: 1px solid #f0f3f7;
  color: var(--text-primary);
  vertical-align: middle;
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.data-table tbody tr {
  transition: background var(--transition-fast);
}

.data-table tbody tr:hover {
  background: #f8fafc;
}

.th-action {
  width: 80px;
  text-align: center;
}

.cell-primary {
  font-weight: 600;
  font-size: 0.82rem;
  color: var(--text-primary);
}

.cell-stack {
  display: flex;
  flex-direction: column;
}

.cell-sub {
  font-size: 0.68rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-text {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.td-action {
  text-align: center;
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
}

.act-view {
  color: #2563eb !important;
}

.act-view:hover {
  background: #eff6ff !important;
}

.act-edit {
  color: #059669 !important;
}

.act-edit:hover {
  background: #ecfdf5 !important;
}
</style>
