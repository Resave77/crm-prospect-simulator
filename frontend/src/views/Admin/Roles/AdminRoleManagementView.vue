<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAdminStore } from '../../../stores/admin'
import type { SalesRole, SalesRoleLevel } from '../../../types/admin'

const store = useAdminStore()
const router = useRouter()
const toast = useToast()
const error = ref('')
const showGuidance = ref(true)
const search = ref('')
const levelFilter = ref<SalesRoleLevel | ''>('')
const statusFilter = ref<'ACTIVE' | 'INACTIVE' | ''>('ACTIVE')

const LEVEL_ORDER: SalesRoleLevel[] = [1, 2, 3, 4]
const LEVEL_GUIDE: Record<SalesRoleLevel, string> = {
  1: 'Oversees Levels 2–4',
  2: 'Oversees Levels 3–4 in its own team',
  3: 'Oversees Level 4 in its own team',
  4: 'Self activity only',
}
const LEVEL_HIERARCHY: Record<SalesRoleLevel, string> = {
  1: 'Oversees Levels 2, 3, and 4',
  2: 'Oversees Levels 3 and 4 in its own team',
  3: 'Oversees Level 4 in its own team',
  4: 'Self activity only',
}
const LEVEL_FALLBACK_DESCRIPTION: Record<SalesRoleLevel, string> = {
  1: 'Top-level sales leader who oversees all descendant sales teams.',
  2: 'Regional or area manager who oversees Levels 3 and 4 in their own team.',
  3: 'Supervisor who oversees Level 4 sales members in their own team.',
  4: 'Individual sales member who sees and manages their own activity.',
}
const DEFAULT_ROLE_IDS = new Set([
  '00000000-0000-0000-0000-000000000101',
  '00000000-0000-0000-0000-000000000102',
  '00000000-0000-0000-0000-000000000103',
  '00000000-0000-0000-0000-000000000104',
  '00000000-0000-0000-0000-000000000105',
  '00000000-0000-0000-0000-000000000106',
  '00000000-0000-0000-0000-000000000107',
  '00000000-0000-0000-0000-000000000108',
])

const roleList = computed(() => Array.isArray(store.salesRoles) ? store.salesRoles : [])
const levelOptions = LEVEL_ORDER.map((level) => ({ label: `Level ${level} — ${LEVEL_GUIDE[level]}`, value: level }))
const levelFilterOptions = [{ label: 'All Levels', value: '' }, ...levelOptions]
const statusOptions = [
  { label: 'Active Roles', value: 'ACTIVE' },
  { label: 'Inactive Roles', value: 'INACTIVE' },
  { label: 'All Roles', value: '' },
]
const totalRoles = computed(() => roleList.value.length)
const activeRoles = computed(() => roleList.value.filter((role) => role.isActive).length)
const levelCounts = computed(() => LEVEL_ORDER.map((level) => ({
  level,
  count: roleList.value.filter((role) => role.level === level).length,
})))
const filteredRoles = computed(() => {
  const query = search.value.trim().toLowerCase()
  return roleList.value.filter((role) => {
    const matchesSearch = !query || [role.name, role.description ?? '', role.landingPage ?? ''].some((value) => value.toLowerCase().includes(query))
    const matchesLevel = !levelFilter.value || role.level === levelFilter.value
    const matchesStatus = !statusFilter.value || (role.isActive ? 'ACTIVE' : 'INACTIVE') === statusFilter.value
    return matchesSearch && matchesLevel && matchesStatus
  })
})
const groupedRoles = computed(() => [...filteredRoles.value].sort((a, b) => a.level - b.level || a.name.localeCompare(b.name)))
const hasFilters = computed(() => Boolean(search.value.trim() || levelFilter.value || statusFilter.value))
const duplicateRoleNames = computed(() => {
  const counts = new Map<string, number>()
  for (const role of roleList.value) {
    const key = normalizeRoleName(role.name)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return counts
})

function roleDisplayDescription(role: SalesRole) {
  return role.description || LEVEL_FALLBACK_DESCRIPTION[role.level]
}

function normalizeRoleName(name: string) {
  return name.trim().toLowerCase().replace(/\s+/g, ' ')
}

function isDemoRole(role: SalesRole) {
  return role.name.toLowerCase().includes('collector') || role.name.toLowerCase().includes('billing') || role.name.toLowerCase().includes('merchandising') || role.name === 'Admin Sales'
}

function isSystemDefaultRole(role: SalesRole) {
  return DEFAULT_ROLE_IDS.has(role.id)
}

function hasDuplicateName(role: SalesRole) {
  return (duplicateRoleNames.value.get(normalizeRoleName(role.name)) ?? 0) > 1
}

function permissionCount(role: SalesRole) {
  return role.permissionCount ?? role.permissions?.length ?? null
}

function load() {
  error.value = ''
  store.fetchSalesRoles().catch((e) => { error.value = store.errorMessage(e) })
}

function goCreate() {
  router.push('/admin/role-management/create')
}

function goView(role: SalesRole) {
  router.push(`/admin/role-management/${role.id}`)
}

function goEdit(role: SalesRole) {
  router.push(`/admin/role-management/${role.id}/edit`)
}

async function toggleStatus(role: SalesRole) {
  error.value = ''
  const nextActive = !role.isActive
  try {
    await store.setSalesRoleStatus(role.id, nextActive)
    toast.add({ severity: 'success', summary: nextActive ? 'Role Activated' : 'Role Deactivated', detail: `${role.name} is now ${nextActive ? 'active' : 'inactive'}.`, life: 3000 })
  } catch (e) {
    error.value = store.errorMessage(e)
  }
}

async function deleteRole(role: SalesRole) {
  if (isSystemDefaultRole(role) || store.savingSalesRole) return
  error.value = ''
  try {
    await store.deleteSalesRole(role.id)
    toast.add({ severity: 'success', summary: 'Role Deleted', detail: `${role.name} has been deleted.`, life: 3000 })
  } catch (e) {
    error.value = store.errorMessage(e)
  }
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat('en', { year: 'numeric', month: 'short', day: '2-digit' }).format(date)
}

function levelSeverity(level: SalesRoleLevel) {
  if (level === 1) return 'info'
  if (level === 2) return 'success'
  if (level === 3) return 'warn'
  return 'secondary'
}

function clearFilters() {
  search.value = ''
  levelFilter.value = ''
  statusFilter.value = 'ACTIVE'
}

onMounted(load)
</script>

<template>
  <section class="admin-page">
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Sales Organization</span>
        <h1>Role Management</h1>
        <p class="muted">Role names are flexible. Numeric levels control assignment hierarchy and reporting rules.</p>
      </div>
      <Button label="Create Role" icon="pi pi-plus" size="small" @click="goCreate" />
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div class="summary-grid">
      <div class="summary-card"><span>Total Roles</span><strong>{{ totalRoles }}</strong></div>
      <div class="summary-card"><span>Active Roles</span><strong>{{ activeRoles }}</strong></div>
    </div>

    <div class="filter-panel">
      <div class="search-field">
        <i class="pi pi-search" />
        <InputText v-model="search" placeholder="Search role name, description, or landing page..." />
      </div>
      <div class="filter-field">
        <label>Level</label>
        <Select v-model="levelFilter" :options="levelFilterOptions" optionLabel="label" optionValue="value" />
      </div>
      <div class="filter-field">
        <label>Role View</label>
        <Select v-model="statusFilter" :options="statusOptions" optionLabel="label" optionValue="value" />
      </div>
      <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="clearFilters" />
    </div>

    <div class="table-panel">
      <div v-if="store.salesRolesLoading && !roleList.length" class="skeleton-area">
        <Skeleton v-for="n in 6" :key="n" class="skeleton-row" />
      </div>
      <DataTable
        v-else
        :value="groupedRoles"
        :loading="store.salesRolesLoading"
        dataKey="id"
        scrollable
        rowGroupMode="subheader"
        groupRowsBy="level"
        sortField="level"
        :sortOrder="1"
      >
        <template #groupheader="{ data }">
          <span class="level-group-title">Level {{ data.level }}</span>
        </template>
        <template #empty>
          <div class="empty-state">
            <div class="empty-icon"><i class="pi pi-id-card" /></div>
            <strong>No sales roles found</strong>
            <span>{{ hasFilters ? 'Adjust your filters to view roles.' : 'Create a role to begin configuring the sales hierarchy.' }}</span>
            <Button v-if="!hasFilters" label="Create Role" icon="pi pi-plus" size="small" @click="goCreate" />
          </div>
        </template>
        <Column field="level" header="Level" :style="{ width: '100px' }">
          <template #body="{ data }">
            <span class="level-chip" :class="`level-chip-${data.level}`">Level {{ data.level }}</span>
          </template>
        </Column>
        <Column field="name" header="Role Name" :style="{ minWidth: '200px' }">
          <template #body="{ data }">
            <div class="role-name-cell">
              <button type="button" class="role-name-link" @click="goView(data)"><strong class="role-name">{{ data.name }}</strong></button>
              <Tag v-if="isSystemDefaultRole(data)" value="System Default" severity="info" size="small" class="template-tag" />
              <Tag v-if="isDemoRole(data)" value="Demo" severity="warn" size="small" class="template-tag" />
              <span v-if="hasDuplicateName(data)" class="warning-chip" title="Duplicate display name"><i class="pi pi-exclamation-triangle" /></span>
            </div>
          </template>
        </Column>
        <Column header="Description" :style="{ minWidth: '260px' }">
          <template #body="{ data }">
            <template v-if="data.description">
              <span class="cell-text clamp-2">{{ data.description }}</span>
            </template>
            <template v-else>
              <div class="cell-text clamp-2">{{ roleDisplayDescription(data) }}</div>
              <span class="cell-hint">Edit this role to customize the name and description.</span>
            </template>
          </template>
        </Column>
        <Column header="Landing Page" :style="{ width: '190px' }">
          <template #body="{ data }">
            <code v-if="data.landingPage" class="route-badge">{{ data.landingPage }}</code>
            <span v-else class="cell-hint">—</span>
          </template>
        </Column>
        <Column header="Permissions" :style="{ width: '110px' }">
          <template #body="{ data }">
            <span class="perm-count">{{ permissionCount(data) ?? '—' }}</span>
          </template>
        </Column>
        <Column header="Status" :style="{ width: '100px' }"><template #body="{ data }"><Tag :value="data.isActive ? 'Active' : 'Inactive'" :severity="data.isActive ? 'success' : 'secondary'" size="small" class="soft-tag" /></template></Column>
        <Column header="Updated At" :style="{ width: '120px' }"><template #body="{ data }"><span class="date-text">{{ formatDate(data.updatedAt) }}</span></template></Column>
        <Column header="Actions" :style="{ width: '110px' }">
          <template #body="{ data }">
            <details class="action-menu">
              <summary><i class="pi pi-ellipsis-v" /><span class="sr-only">Actions</span></summary>
              <div class="action-popover">
                <button type="button" @click="goView(data)"><i class="pi pi-eye" /> View</button>
                <button type="button" @click="goEdit(data)"><i class="pi pi-pencil" /> Edit</button>
                <button type="button" :class="{ danger: data.isActive, success: !data.isActive }" :disabled="store.savingSalesRole" @click="toggleStatus(data)">
                  <i :class="data.isActive ? 'pi pi-ban' : 'pi pi-check-circle'" /> {{ data.isActive ? 'Deactivate' : 'Activate' }}
                </button>
                <button type="button" class="danger" :disabled="isSystemDefaultRole(data) || store.savingSalesRole" :title="isSystemDefaultRole(data) ? 'Default role cannot be deleted' : 'Delete role'" @click="deleteRole(data)">
                  <i class="pi pi-trash" /> Delete
                </button>
              </div>
            </details>
          </template>
        </Column>
      </DataTable>
    </div>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 0.95rem; padding: 1.35rem 1.6rem; min-height: 100vh; background: #ffffff; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: #0b7766; }
h1 { margin: 0.2rem 0 0.2rem; font-size: 1.6rem; color: #0f172a; }
.muted { margin: 0; color: #7c8798; font-size: 0.85rem; }
.summary-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.6rem; }
.summary-card, .filter-panel, .table-panel, .info-section, .guidance-banner { background: #fff; border: 1px solid #edf1f6; border-radius: 10px; box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03); }
.summary-card { padding: 0.68rem 0.8rem; display: grid; gap: 0.12rem; }
.summary-card span { color: #64748b; font-size: 0.68rem; font-weight: 750; text-transform: uppercase; }
.summary-card strong { color: #0f172a; font-size: 1.08rem; }
.level-breakdown { display: flex; flex-wrap: wrap; gap: 0.3rem; }
.guidance-banner { display: flex; align-items: flex-start; gap: 0.8rem; padding: 0.85rem 1rem; background: #fffbeb; border-color: #fde68a; }
.guidance-icon { width: 34px; height: 34px; flex: 0 0 auto; display: grid; place-content: center; border-radius: 10px; background: #fef3c7; color: #b45309; }
.guidance-banner div:nth-child(2) { display: grid; gap: 0.2rem; flex: 1; }
.guidance-banner strong { color: #92400e; font-size: 0.9rem; }
.guidance-banner span { color: #a16207; font-size: 0.8rem; line-height: 1.45; }
.info-section { display: grid; grid-template-columns: minmax(260px, 1.1fr) 1.6fr; gap: 1.1rem; padding: 1.1rem 1.2rem; }
.info-body { display: grid; align-content: start; gap: 0.55rem; }
.info-title { display: flex; align-items: center; gap: 0.5rem; font-weight: 750; color: #0f172a; font-size: 0.95rem; }
.info-title i { color: #0b7766; }
.info-list { margin: 0; padding-left: 1.15rem; display: grid; gap: 0.35rem; color: #475569; font-size: 0.82rem; line-height: 1.45; }
.hierarchy-guide { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.6rem; align-content: start; }
.hier-card { display: flex; flex-direction: column; gap: 0.2rem; padding: 0.7rem 0.8rem; border: 1px solid #edf1f6; border-radius: 10px; background: #f8fafc; }
.hier-level { font-weight: 750; color: #0f172a; font-size: 0.8rem; }
.hier-desc { color: #64748b; font-size: 0.74rem; line-height: 1.4; }
.filter-panel { display: grid; grid-template-columns: minmax(260px, 1fr) 200px 160px auto; gap: 0.65rem; align-items: end; padding: 0.8rem; }
.search-field { display: flex; align-items: center; gap: 0.6rem; min-height: 2.55rem; padding: 0 0.75rem; background: #fff; border: 1px solid #dbe3ee; border-radius: 8px; }
.search-field i { color: #94a3b8; }
.search-field :deep(.p-inputtext) { flex: 1; border: 0; padding-inline: 0; box-shadow: none; }
.filter-field { display: grid; gap: 0.35rem; }
.filter-field label { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; color: #64748b; }
.optional-label { font-weight: 500; text-transform: none; color: #94a3b8; }
.table-panel { overflow-x: auto; }
.table-panel :deep(.p-datatable), .table-panel :deep(.p-datatable-table), .table-panel :deep(.p-datatable-table-container) { background: #ffffff; color: #0f172a; }
.table-panel :deep(.p-datatable-thead > tr > th) { background: #f8fafc; color: #475569; font-size: 0.66rem; text-transform: uppercase; border-color: #edf1f6; padding: 0.55rem 0.75rem; }
.table-panel :deep(.p-datatable-tbody > tr), .table-panel :deep(.p-datatable-tbody > tr > td) { background: #ffffff; color: #1e293b; border-color: #f1f4f8; }
.table-panel :deep(.p-datatable-tbody > tr > td) { padding: 0.52rem 0.75rem; font-size: 0.8rem; }
.table-panel :deep(.p-datatable-tbody > tr:hover > td) { background: #f8fafc; }
.table-panel :deep(.p-rowgroup-header > td) { background: #f1f5f9 !important; padding: 0.42rem 0.75rem !important; border-color: #e2e8f0 !important; }
.table-panel :deep(.p-datatable-loading-overlay) { background: rgba(255, 255, 255, 0.72); }
.role-name { color: #0f172a; font-weight: 750; }
.role-name-cell { display: flex; align-items: center; gap: 0.45rem; flex-wrap: wrap; }
.role-name-link { border: 0; background: transparent; padding: 0; font: inherit; cursor: pointer; text-align: left; }
.role-name-link:hover .role-name { color: #0b7766; text-decoration: underline; }
.template-tag { opacity: 0.85; }
.warning-chip { display: inline-grid; place-content: center; width: 1.35rem; height: 1.35rem; border-radius: 999px; background: #fff7ed; color: #b45309; font-size: 0.68rem; }
.level-group-title { color: #334155; font-size: 0.68rem; font-weight: 850; letter-spacing: 0.08em; text-transform: uppercase; }
.level-chip { display: inline-flex; align-items: center; justify-content: center; min-width: 4.4rem; padding: 0.18rem 0.48rem; border-radius: 999px; font-size: 0.7rem; font-weight: 800; border: 1px solid transparent; }
.level-chip-1 { background: #eff6ff; color: #1d4ed8; border-color: #dbeafe; }
.level-chip-2 { background: #ecfdf5; color: #047857; border-color: #d1fae5; }
.level-chip-3 { background: #fff7ed; color: #9a3412; border-color: #fed7aa; }
.level-chip-4 { background: #f8fafc; color: #475569; border-color: #e2e8f0; }
.cell-text { color: #475569; line-height: 1.35; }
.clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.cell-hint { display: block; margin-top: 0.2rem; color: #94a3b8; font-size: 0.72rem; }
.route-badge { font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; font-size: 0.68rem; font-weight: 600; color: #2563eb; background: #eff6ff; border: 1px solid #dbeafe; border-radius: 6px; padding: 0.14rem 0.42rem; white-space: nowrap; }
.perm-count { font-weight: 800; color: #334155; font-size: 0.8rem; }
.date-text { color: #64748b; font-size: 0.78rem; }
.soft-tag { font-size: 0.68rem; border-radius: 999px; }
.action-menu { position: relative; width: fit-content; }
.action-menu summary { width: 2rem; height: 2rem; display: grid; place-content: center; list-style: none; border: 1px solid #e2e8f0; border-radius: 999px; background: #ffffff; color: #64748b; cursor: pointer; }
.action-menu summary::-webkit-details-marker { display: none; }
.action-menu summary:hover { background: #f8fafc; color: #0f172a; }
.action-popover { position: absolute; z-index: 20; right: 0; top: calc(100% + 0.3rem); min-width: 150px; display: grid; gap: 0.15rem; padding: 0.35rem; border: 1px solid #edf1f6; border-radius: 10px; background: #ffffff; box-shadow: 0 12px 28px -12px rgba(15, 23, 42, 0.22); }
.action-popover button { display: flex; align-items: center; gap: 0.45rem; width: 100%; border: 0; border-radius: 8px; padding: 0.5rem 0.6rem; background: transparent; color: #334155; font: inherit; font-size: 0.78rem; font-weight: 700; text-align: left; cursor: pointer; }
.action-popover button:hover { background: #f8fafc; }
.action-popover button.danger { color: #dc2626; }
.action-popover button.success { color: #059669; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
.skeleton-area { padding: 0.6rem 1rem 1rem; }
.skeleton-row { height: 3rem; margin-top: 0.65rem; border-radius: 10px; }
.empty-state { min-height: 240px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.45rem; padding: 2rem; text-align: center; color: #8492a6; }
.empty-icon { width: 52px; height: 52px; display: grid; place-content: center; border-radius: 14px; background: #f8fafc; color: #94a3b8; }
.empty-state strong { color: #0f172a; }
:deep(.p-inputtext), :deep(.p-select) { background: #ffffff; color: #0f172a; border-color: #dbe3ee; }
:deep(.p-select-label), :deep(.p-select-dropdown) { color: #64748b; }
:deep(.p-select-overlay), :deep(.p-select-list), :deep(.p-select-option) { background: #ffffff; color: #0f172a; }
:deep(.p-select-option.p-select-option-selected), :deep(.p-select-option:hover) { background: #f1f5f9; color: #0f172a; }
@media (max-width: 900px) { .filter-panel { grid-template-columns: 1fr 1fr; } .search-field { grid-column: 1 / -1; } .info-section { grid-template-columns: 1fr; } }
@media (max-width: 768px) { .admin-page { padding: 1rem; } .summary-grid, .filter-panel { grid-template-columns: 1fr; } .hierarchy-guide { grid-template-columns: 1fr; } }
</style>
