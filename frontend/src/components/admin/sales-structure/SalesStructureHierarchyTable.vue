<script setup lang="ts">
import { computed, ref } from 'vue'
import Button from 'primevue/button'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Menu from 'primevue/menu'
import type { MenuItem } from 'primevue/menuitem'
import type {
  SalesRoleLevel,
  SalesStructureItem,
} from '../../../types/admin'

interface HierarchyRow {
  item: SalesStructureItem
  depth: number
  hasChildren: boolean
  isExpanded: boolean
  match: boolean
}

const props = withDefaults(
  defineProps<{
    rows: HierarchyRow[]
    loading: boolean
    assignedCount: number
    monthLabel: string
    hasFilters: boolean
    selection: HierarchyRow[]

    levelClass: (level: SalesRoleLevel) => string
    employeeIdLabel: (item: SalesStructureItem) => string
    reportsToLabel: (item: SalesStructureItem) => string
    positionLabel: (item: SalesStructureItem) => string
    statusFor: (item: SalesStructureItem) => string

    /**
     * Parent dapat mengirim fungsi proteksi khusus.
     * Default tetap melindungi SUPER_ADMIN Level 1.
     */
    isProtected?: (item: SalesStructureItem) => boolean

    /**
     * Aktifkan hanya jika backend memiliki endpoint untuk
     * mengakhiri assignment.
     */
    canEndAssignment?: boolean
  }>(),
  {
    isProtected: undefined,
    canEndAssignment: false,
  },
)

const emit = defineEmits<{
  'expand-all': []
  'collapse-all': []
  'toggle-node': [userId: string]
  'open-node': [item: SalesStructureItem]
  'assign-first': []
  'reset-filters': []
  'update:selection': [selection: HierarchyRow[]]

  'move-assignment': [item: SalesStructureItem]
  'promote-assignment': [item: SalesStructureItem]
  'demote-assignment': [item: SalesStructureItem]
  'end-assignment': [item: SalesStructureItem]
}>()

const actionMenu = ref<InstanceType<typeof Menu> | null>(null)
const actionTarget = ref<SalesStructureItem | null>(null)

function isProtectedAssignment(item: SalesStructureItem): boolean {
  if (props.isProtected) {
    return props.isProtected(item)
  }

  return (
    item.systemRole === 'SUPER_ADMIN' &&
    item.salesRole.level === 1
  )
}

const actionItems = computed<MenuItem[]>(() => {
  const item = actionTarget.value

  if (!item) {
    return []
  }

  const protectedRoot = isProtectedAssignment(item)
  const level = item.salesRole.level
  const alreadyEnded = Boolean(item.effectiveTo)

  const items: MenuItem[] = [
    {
      label: 'View Details',
      icon: 'pi pi-eye',
      command: () => emit('open-node', item),
    },
    {
      separator: true,
    },
    {
      label: 'Move Assignment',
      icon: 'pi pi-arrow-right-arrow-left',
      disabled: protectedRoot || alreadyEnded,
      command: () => emit('move-assignment', item),
    },
    {
      label: 'Promote One Level',
      icon: 'pi pi-arrow-up',
      disabled:
        protectedRoot ||
        alreadyEnded ||
        level <= 1,
      command: () => emit('promote-assignment', item),
    },
    {
      label: 'Demote One Level',
      icon: 'pi pi-arrow-down',
      disabled:
        protectedRoot ||
        alreadyEnded ||
        level >= 4,
      command: () => emit('demote-assignment', item),
    },
  ]

  if (props.canEndAssignment) {
    items.push(
      {
        separator: true,
      },
      {
        label: 'End Assignment',
        icon: 'pi pi-times-circle',
        class: 'danger-menu-item',
        disabled: protectedRoot || alreadyEnded,
        command: () => emit('end-assignment', item),
      },
    )
  }

  return items
})

function openActionMenu(
  event: MouseEvent,
  item: SalesStructureItem,
): void {
  actionTarget.value = item
  actionMenu.value?.toggle(event)
}
</script>

<template>
  <div class="tree-panel">
    <div class="tree-toolbar">
      <div class="tree-toolbar-copy">
        <strong>Hierarchy Table</strong>

        <span>
          {{ assignedCount }} assigned sales for {{ monthLabel }}
        </span>
      </div>

      <div class="tree-actions">
        <Button
          label="Expand All"
          icon="pi pi-plus-circle"
          severity="secondary"
          text
          size="small"
          @click="emit('expand-all')"
        />

        <Button
          label="Collapse All"
          icon="pi pi-minus-circle"
          severity="secondary"
          text
          size="small"
          @click="emit('collapse-all')"
        />
      </div>
    </div>

    <div
      v-if="loading && !rows.length"
      class="skeleton-area"
    >
      <Skeleton
        v-for="n in 6"
        :key="n"
        class="skeleton-row"
      />
    </div>

    <div
      v-else-if="!rows.length"
      class="empty-state"
    >
      <div class="empty-icon">
        <i class="pi pi-sitemap" />
      </div>

      <strong>No assignments for this month</strong>

      <span>
        {{
          hasFilters
            ? 'Adjust your search or filters to view the organization.'
            : 'Assign the first eligible sales member to continue the hierarchy.'
        }}
      </span>

      <Button
        :label="hasFilters ? 'Reset Filters' : 'Assign Sales'"
        :icon="hasFilters ? 'pi pi-replay' : 'pi pi-plus'"
        size="small"
        @click="
          hasFilters
            ? emit('reset-filters')
            : emit('assign-first')
        "
      />
    </div>

    <DataTable
      v-else
      :selection="selection"
      :value="rows"
      :loading="loading"
      data-key="item.assignmentId"
      class="hierarchy-table"
      table-style="width: 100%; table-layout: fixed;"
      @update:selection="emit('update:selection', $event)"
    >
      <Column
        selection-mode="multiple"
        class="selection-column"
        header-class="selection-column"
      />

      <Column
        header="Level"
        class="level-column"
        header-class="level-column"
      >
        <template #body="{ data }">
          <span
            class="hier-level"
            :class="levelClass(data.item.salesRole.level)"
          >
            {{ data.item.salesRole.level }}
          </span>
        </template>
      </Column>

      <Column
        header="Sales"
        class="sales-column"
        header-class="sales-column"
      >
        <template #body="{ data }">
          <div
            class="hier-name-cell"
            :style="{ '--depth': data.depth }"
          >
            <span
              v-if="data.depth"
              class="hier-connector"
            />

            <button
              v-if="data.hasChildren"
              class="hier-expander"
              type="button"
              :title="
                data.isExpanded
                  ? 'Collapse row'
                  : 'Expand row'
              "
              @click.stop="
                emit('toggle-node', data.item.userId)
              "
            >
              <i
                :class="
                  data.isExpanded
                    ? 'pi pi-chevron-down'
                    : 'pi pi-chevron-right'
                "
              />
            </button>

            <span
              v-else
              class="hier-expander-spacer"
            />

            <button
              class="hier-name-button"
              type="button"
              @click="emit('open-node', data.item)"
            >
              <strong
                class="sales-name"
                :class="
                  levelClass(data.item.salesRole.level)
                "
                :title="data.item.salesName"
              >
                {{ data.item.salesName }}
              </strong>

              <small>
                {{ employeeIdLabel(data.item) }}
              </small>

              <span class="responsive-meta">
                {{ positionLabel(data.item) }}
              </span>

              <span class="responsive-meta">
                Reports to:
                {{ reportsToLabel(data.item) }}
              </span>

              <span class="responsive-role">
                {{ data.item.salesRole.name }}
              </span>
            </button>
          </div>
        </template>
      </Column>

      <Column
        header="Reports To"
        class="reports-column optional-column"
        header-class="reports-column optional-column"
      >
        <template #body="{ data }">
          <span
            class="truncate-cell"
            :title="reportsToLabel(data.item)"
          >
            {{ reportsToLabel(data.item) }}
          </span>
        </template>
      </Column>

      <Column
        header="Position"
        class="position-column optional-column"
        header-class="position-column optional-column"
      >
        <template #body="{ data }">
          <span
            class="truncate-cell"
            :title="positionLabel(data.item)"
          >
            {{ positionLabel(data.item) }}
          </span>
        </template>
      </Column>

      <Column
        header="Role"
        class="role-column"
        header-class="role-column"
      >
        <template #body="{ data }">
          <span
            class="role-name-label"
            :title="data.item.salesRole.name"
          >
            {{ data.item.salesRole.name }}
          </span>
        </template>
      </Column>

      <Column
        header="Status"
        class="status-column"
        header-class="status-column"
      >
        <template #body="{ data }">
          <span
            class="compact-status"
            :class="{ inactive: data.item.effectiveTo }"
          >
            {{ statusFor(data.item) }}
          </span>
        </template>
      </Column>

      <Column
        header=""
        class="action-column"
        header-class="action-column"
      >
        <template #body="{ data }">
          <button
            class="more-action"
            type="button"
            title="Assignment actions"
            aria-label="Open assignment actions"
            @click.stop="
              openActionMenu($event, data.item)
            "
          >
            <i class="pi pi-ellipsis-v" />
          </button>
        </template>
      </Column>
    </DataTable>

    <Menu
      ref="actionMenu"
      :model="actionItems"
      popup
      class="assignment-action-menu"
    />
  </div>
</template>

<style scoped>
.tree-panel {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid #edf1f6;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.tree-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0.9rem;
  border-bottom: 1px solid #edf1f6;
}

.tree-toolbar-copy {
  display: grid;
  gap: 0.15rem;
  min-width: 0;
}

.tree-toolbar-copy strong {
  color: #0f172a;
  font-size: 0.92rem;
}

.tree-toolbar-copy span {
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.76rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-actions {
  display: flex;
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.35rem;
}

.hierarchy-table {
  width: 100%;
  min-width: 0;
}

.hierarchy-table :deep(.p-datatable-table-container) {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  background: #ffffff;
}

.hierarchy-table :deep(.p-datatable-table) {
  width: 100%;
  min-width: 0;
  table-layout: fixed;
  background: #ffffff;
}

.hierarchy-table :deep(.p-datatable-thead > tr > th) {
  overflow: hidden;
  padding: 0.55rem 0.6rem;
  border-color: #e5eaf0;
  background: #f3f5f8;
  color: #475569;
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.hierarchy-table :deep(.p-datatable-tbody > tr > td) {
  overflow: hidden;
  height: 52px;
  padding: 0.46rem 0.6rem;
  border-color: #edf1f6;
  background: #ffffff;
  color: #1e293b;
  font-size: 0.78rem;
  text-overflow: ellipsis;
  vertical-align: middle;
}

.hierarchy-table :deep(.p-datatable-tbody > tr:hover > td) {
  background: #f8fbff;
}

.hierarchy-table :deep(.selection-column) {
  width: 4%;
  min-width: 38px;
}

.hierarchy-table :deep(.level-column) {
  width: 6%;
  text-align: center;
}

.hierarchy-table :deep(.sales-column) {
  width: 27%;
}

.hierarchy-table :deep(.reports-column) {
  width: 15%;
}

.hierarchy-table :deep(.position-column) {
  width: 16%;
}

.hierarchy-table :deep(.role-column) {
  width: 16%;
}

.hierarchy-table :deep(.status-column) {
  width: 10%;
}

.hierarchy-table :deep(.action-column) {
  width: 6%;
  text-align: center;
}

.hier-level {
  font-size: 0.82rem;
  font-weight: 850;
}

.hier-level.level-1,
.sales-name.level-1 {
  color: #1d4ed8;
}

.hier-level.level-2,
.sales-name.level-2 {
  color: #047857;
}

.hier-level.level-3,
.sales-name.level-3 {
  color: #9a3412;
}

.hier-level.level-4,
.sales-name.level-4 {
  color: #a16207;
}

.hier-name-cell {
  --depth: 0;

  position: relative;
  display: flex;
  align-items: center;
  min-width: 0;
  min-height: 34px;
  gap: 0.28rem;
  padding-left: min(
    calc(var(--depth) * 1.05rem),
    4.2rem
  );
}

.hier-connector {
  position: absolute;
  top: 0;
  bottom: 0;
  left: max(
    calc(min(var(--depth), 4) * 1.05rem - 0.62rem),
    0px
  );
  width: 0.68rem;
  border-bottom: 1px solid #dbe3ee;
  border-left: 1px solid #dbe3ee;
  transform: translateY(-50%);
}

.hier-expander,
.hier-expander-spacer {
  width: 1.4rem;
  height: 1.4rem;
  flex: 0 0 auto;
}

.hier-expander {
  display: grid;
  place-content: center;
  border: 1px solid #dbe3ee;
  border-radius: 6px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
}

.hier-expander:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.hier-expander i {
  font-size: 0.68rem;
}

.hier-name-button {
  display: grid;
  min-width: 0;
  gap: 0.03rem;
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  font: inherit;
  cursor: pointer;
}

.sales-name {
  display: block;
  overflow: hidden;
  max-width: 100%;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hier-name-button:hover .sales-name {
  text-decoration: underline;
  text-underline-offset: 2px;
}

.hier-name-button small {
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.67rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.responsive-meta,
.responsive-role {
  display: none;
  overflow: hidden;
  color: #64748b;
  font-size: 0.66rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.responsive-role {
  color: #475569;
  font-weight: 650;
}

.truncate-cell {
  display: block;
  overflow: hidden;
  max-width: 100%;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-name-label {
  display: -webkit-box;
  overflow: hidden;
  max-width: 100%;
  color: #334155;
  line-height: 1.3;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.compact-status {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  gap: 0.3rem;
  color: #047857;
  font-size: 0.7rem;
  font-weight: 800;
  white-space: nowrap;
}

.compact-status::before {
  width: 0.4rem;
  height: 0.4rem;
  flex: 0 0 auto;
  border-radius: 999px;
  background: #10b981;
  content: '';
}

.compact-status.inactive {
  color: #64748b;
}

.compact-status.inactive::before {
  background: #94a3b8;
}

.more-action {
  display: grid;
  width: 2rem;
  height: 2rem;
  margin: 0 auto;
  place-content: center;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
  background: #ffffff;
  color: #64748b;
  cursor: pointer;
}

.more-action:hover {
  border-color: #cbd5e1;
  background: #f8fafc;
  color: #0f172a;
}

.assignment-action-menu :deep(.danger-menu-item .p-menu-item-content) {
  color: #dc2626;
}

.skeleton-area {
  padding: 0.6rem 1rem 1rem;
}

.skeleton-row {
  height: 3rem;
  margin-top: 0.65rem;
  border-radius: 10px;
}

.empty-state {
  display: flex;
  min-height: 240px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 2rem;
  color: #8492a6;
  text-align: center;
}

.empty-icon {
  display: grid;
  width: 52px;
  height: 52px;
  place-content: center;
  border-radius: 14px;
  background: #f8fafc;
  color: #94a3b8;
}

.empty-state strong {
  color: #0f172a;
}

/*
 * Tablet dan desktop sempit:
 * sembunyikan Reports To dan Position.
 * Informasinya otomatis muncul di bawah nama.
 */
@media (max-width: 1100px) {
  .optional-column {
    display: none;
  }

  .hierarchy-table :deep(.selection-column) {
    width: 5%;
  }

  .hierarchy-table :deep(.level-column) {
    width: 8%;
  }

  .hierarchy-table :deep(.sales-column) {
    width: 44%;
  }

  .hierarchy-table :deep(.role-column) {
    width: 25%;
  }

  .hierarchy-table :deep(.status-column) {
    width: 11%;
  }

  .hierarchy-table :deep(.action-column) {
    width: 7%;
  }

  .responsive-meta {
    display: block;
  }
}

/*
 * Layar kecil:
 * hanya tampilkan Level, Sales, Status dan Actions.
 * Tidak ada horizontal scrolling.
 */
@media (max-width: 720px) {
  .tree-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .tree-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .hierarchy-table :deep(.selection-column),
  .hierarchy-table :deep(.role-column) {
    display: none;
  }

  .hierarchy-table :deep(.level-column) {
    width: 12%;
  }

  .hierarchy-table :deep(.sales-column) {
    width: 65%;
  }

  .hierarchy-table :deep(.status-column) {
    width: 13%;
  }

  .hierarchy-table :deep(.action-column) {
    width: 10%;
  }

  .responsive-role {
    display: block;
  }

  .hierarchy-table :deep(.p-datatable-thead > tr > th),
  .hierarchy-table :deep(.p-datatable-tbody > tr > td) {
    padding-right: 0.38rem;
    padding-left: 0.38rem;
  }

  .hier-name-cell {
    padding-left: min(
      calc(var(--depth) * 0.7rem),
      2.1rem
    );
  }

  .compact-status {
    font-size: 0;
  }

  .compact-status::before {
    width: 0.55rem;
    height: 0.55rem;
  }

  .tree-actions :deep(.p-button-label) {
    display: none;
  }
}
</style>