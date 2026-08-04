<script setup lang="ts">
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import type { AdminUserListItem, SalesRoleLevel, SalesStructureItem } from '../../../types/admin'

type ActiveTab = 'assigned' | 'unassigned' | 'all'

defineProps<{
  activeTab: ActiveTab
  assignedRows: SalesStructureItem[]
  salesUsers: AdminUserListItem[]
  unassignedUsers: AdminUserListItem[]
  salesStructureLoading: boolean
  salesUsersLoading: boolean
  salesStructureEmpty: boolean
  assignedUserIds: Set<string>
  hasFilters: boolean
  userSearch: string
  hasMoreUsers: boolean
  salesUsersTotal: number
  selection: SalesStructureItem[]
  nextGuidance: { title: string; button: string; level: SalesRoleLevel } | null
  levelClass: (level: SalesRoleLevel) => string
  positionLabel: (item: SalesStructureItem) => string
  statusFor: (item: SalesStructureItem) => string
  roleLabel: (role: string) => string
  roleSeverity: (role: string) => string
}>()

const emit = defineEmits<{
  'assign-user': [user: AdminUserListItem]
  'assign-first': [level?: SalesRoleLevel]
  'load-more': []
  'reset-filters': []
  'update:selection': [selection: SalesStructureItem[]]
}>()
</script>

<template>
  <div class="table-panel">
    <div v-if="salesStructureLoading && salesStructureEmpty" class="skeleton-area">
      <Skeleton v-for="n in 8" :key="n" class="skeleton-row" />
    </div>

    <DataTable
      v-else-if="activeTab === 'assigned'"
      :selection="selection"
      :value="assignedRows"
      :loading="salesStructureLoading"
      dataKey="assignmentId"
      scrollable
      paginator
      :rows="20"
      :rowsPerPageOptions="[20, 50, 100]"
      sortMode="multiple"
      paginatorTemplate="CurrentPageReport FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink RowsPerPageDropdown"
      currentPageReportTemplate="Showing {first} to {last} of {totalRecords} assignments"
      @update:selection="emit('update:selection', $event)"
    >
      <template #empty>
        <div class="empty-state">
          <div v-if="hasFilters" class="empty-body">
            <div class="empty-icon"><i class="pi pi-filter" /></div>
            <strong>No assignments match your filters</strong>
            <span>Adjust your filters to view assignments.</span>
            <Button label="Reset Filters" icon="pi pi-replay" severity="secondary" text size="small" @click="emit('reset-filters')" />
          </div>
          <template v-else>
            <div class="empty-icon"><i class="pi pi-sitemap" /></div>
            <template v-if="nextGuidance">
              <strong>{{ nextGuidance.title }}</strong>
              <span>Build the hierarchy from the top down to unlock higher levels.</span>
              <Button :label="nextGuidance.button" icon="pi pi-plus" size="small" @click="emit('assign-first', nextGuidance.level)" />
            </template>
            <template v-else>
              <strong>No assignments for this month</strong>
              <span>Create an assignment for the selected month.</span>
              <Button label="Assign First Leader" icon="pi pi-plus" size="small" @click="emit('assign-first', 1)" />
            </template>
          </template>
        </div>
      </template>
      <Column selectionMode="multiple" headerStyle="width: 3rem" />
      <Column field="salesRole.level" header="Role Level" sortable :style="{ width: '130px' }">
        <template #body="{ data }"><Tag :value="`Level ${data.salesRole.level}`" severity="info" /></template>
      </Column>
      <Column field="salesName" header="Sales Name" sortable :style="{ minWidth: '220px' }">
        <template #body="{ data }"><span class="sales-name" :class="levelClass(data.salesRole.level)">{{ data.salesName }}</span></template>
      </Column>
      <Column field="parentName" header="Reports To" sortable :style="{ minWidth: '180px' }">
        <template #body="{ data }">{{ data.parentName || '-' }}</template>
      </Column>
      <Column field="systemRole" header="Position" sortable :style="{ minWidth: '160px' }">
        <template #body="{ data }">{{ positionLabel(data) }}</template>
      </Column>
      <Column field="salesRole.name" header="Role Name" sortable :style="{ minWidth: '190px' }">
        <template #body="{ data }">{{ data.salesRole.name }}</template>
      </Column>
      <Column header="Status" sortable sortField="effectiveTo" :style="{ width: '120px' }">
        <template #body="{ data }"><Tag :value="statusFor(data)" :severity="data.effectiveTo ? 'secondary' : 'success'" /></template>
      </Column>
      <Column header="Actions" :style="{ width: '120px' }">
        <template #body>
          <Button icon="pi pi-ellipsis-h" text rounded size="small" disabled title="Assignment movement is deferred" />
        </template>
      </Column>
    </DataTable>

    <DataTable v-else :value="activeTab === 'unassigned' ? unassignedUsers : salesUsers" :loading="salesUsersLoading" dataKey="id" scrollable>
      <template #empty>
        <div class="empty-state">
          <div class="empty-icon"><i class="pi pi-users" /></div>
          <strong v-if="activeTab === 'unassigned'">All active sales users are assigned for this month.</strong>
          <strong v-else>No active sales users found.</strong>
          <span v-if="userSearch || activeTab === 'all'">Adjust your search to view results.</span>
        </div>
      </template>
      <Column header="Employee ID" :style="{ width: '160px' }">
        <template #body="{ data }"><code class="code-tag">{{ data.employeeId || '—' }}</code></template>
      </Column>
      <Column field="fullName" header="Sales Name" sortable :style="{ minWidth: '220px' }">
        <template #body="{ data }"><span class="sales-name">{{ data.fullName }}</span></template>
      </Column>
      <Column field="email" header="Email" sortable :style="{ minWidth: '220px' }">
        <template #body="{ data }"><span class="cell-text">{{ data.email }}</span></template>
      </Column>
      <Column field="role" header="System Role" sortable :style="{ width: '160px' }">
        <template #body="{ data }"><Tag :value="roleLabel(data.role)" :severity="roleSeverity(data.role)" /></template>
      </Column>
      <Column field="status" header="Status" sortable :style="{ width: '110px' }">
        <template #body="{ data }"><Tag :value="data.status" :severity="data.status === 'ACTIVE' ? 'success' : 'secondary'" /></template>
      </Column>
      <Column v-if="activeTab === 'all'" header="Assignment" :style="{ width: '140px' }">
        <template #body="{ data }">
          <Tag :value="assignedUserIds.has(data.id) ? 'Assigned' : 'Unassigned'" :severity="assignedUserIds.has(data.id) ? 'info' : 'secondary'" />
        </template>
      </Column>
      <Column header="Action" :style="{ width: '130px' }">
        <template #body="{ data }">
          <Button v-if="!assignedUserIds.has(data.id)" label="Assign" icon="pi pi-user-plus" text rounded size="small" @click="emit('assign-user', data)" />
          <span v-else class="cell-hint">Assigned</span>
        </template>
      </Column>
    </DataTable>

    <div v-if="activeTab !== 'assigned' && hasMoreUsers" class="load-more-row">
      <span class="cell-hint">Showing {{ salesUsers.length }} of {{ salesUsersTotal }} active accounts.</span>
      <Button label="Load more" icon="pi pi-chevron-down" severity="secondary" text size="small" :loading="salesUsersLoading" @click="emit('load-more')" />
    </div>
  </div>
</template>

<style scoped>
.table-panel { overflow-x: auto; background: #fff; border: 1px solid #edf1f6; border-radius: 12px; box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03); }
.table-panel :deep(.p-datatable), .table-panel :deep(.p-datatable-table), .table-panel :deep(.p-datatable-table-container) { background: #ffffff; color: #0f172a; }
.table-panel :deep(.p-datatable-thead > tr > th) { background: #f8fafc; color: #475569; font-size: 0.68rem; text-transform: uppercase; border-color: #edf1f6; }
.table-panel :deep(.p-datatable-tbody > tr), .table-panel :deep(.p-datatable-tbody > tr > td) { background: #ffffff; color: #1e293b; border-color: #f1f4f8; }
.table-panel :deep(.p-datatable-tbody > tr > td) { white-space: nowrap; }
.table-panel :deep(.p-datatable-tbody > tr:hover > td) { background: #f8fafc; }
.table-panel :deep(.p-paginator) { background: #ffffff; border-color: #edf1f6; color: #475569; }
.table-panel :deep(.p-paginator-current) { color: #64748b; }
.table-panel :deep(.p-datatable-loading-overlay) { background: rgba(255, 255, 255, 0.72); }
.skeleton-area { padding: 0.6rem 1rem 1rem; }
.skeleton-row { height: 3rem; margin-top: 0.65rem; border-radius: 10px; }
.empty-state { min-height: 240px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.45rem; padding: 2rem; text-align: center; color: #8492a6; }
.empty-body { display: flex; flex-direction: column; align-items: center; gap: 0.45rem; }
.empty-icon { width: 52px; height: 52px; display: grid; place-content: center; border-radius: 14px; background: #f8fafc; color: #94a3b8; }
.empty-state strong { color: #0f172a; }
.code-tag { display: inline-block; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; font-size: 0.76rem; font-weight: 600; padding: 0.18rem 0.55rem; border-radius: 6px; background: #f1f5f9; color: #475569; }
.cell-text { color: #475569; }
.cell-hint { color: #94a3b8; font-size: 0.75rem; }
.load-more-row { display: flex; align-items: center; justify-content: center; gap: 1rem; padding: 0.9rem 1rem; border-top: 1px solid #f1f4f8; }
.load-more-row .cell-hint { margin: 0; }
.sales-name { font-weight: 750; }
.sales-name.level-1 { color: #2563eb; }
.sales-name.level-2 { color: #059669; }
.sales-name.level-3 { color: #ea580c; }
.sales-name.level-4 { color: #b45309; }
</style>
