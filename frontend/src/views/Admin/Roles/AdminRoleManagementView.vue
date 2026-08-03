<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAdminStore } from '../../../stores/admin'
import type { SalesRole, SalesRoleLevel } from '../../../types/admin'

const store = useAdminStore()
const toast = useToast()
const error = ref('')
const dialogVisible = ref(false)
const editingRole = ref<SalesRole | null>(null)
const formError = ref('')
const showGuidance = ref(true)
const search = ref('')
const levelFilter = ref<SalesRoleLevel | ''>('')
const statusFilter = ref<'ACTIVE' | 'INACTIVE' | ''>('ACTIVE')

const form = reactive<{ name: string; level: SalesRoleLevel | null; description: string }>({ name: '', level: null, description: '' })

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
const LEVEL_PREVIEW: Record<SalesRoleLevel, string> = {
  1: 'Level 1 is the top organizational level. It has no parent and supervises Levels 2, 3, and 4.',
  2: 'Level 2 may report to Level 1 and supervises Levels 3 and 4 in its own team.',
  3: 'Level 3 may report to Level 2 and may supervise Level 4 in its own team.',
  4: 'Level 4 reports to Level 3 and manages its own activity only.',
}
const LEVEL_FALLBACK_DESCRIPTION: Record<SalesRoleLevel, string> = {
  1: 'Top-level sales leader who oversees all descendant sales teams.',
  2: 'Regional or area manager who oversees Levels 3 and 4 in their own team.',
  3: 'Supervisor who oversees Level 4 sales members in their own team.',
  4: 'Individual sales member who sees and manages their own activity.',
}

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
    const matchesSearch = !query || [role.name, role.description ?? ''].some((value) => value.toLowerCase().includes(query))
    const matchesLevel = !levelFilter.value || role.level === levelFilter.value
    const matchesStatus = !statusFilter.value || (role.isActive ? 'ACTIVE' : 'INACTIVE') === statusFilter.value
    return matchesSearch && matchesLevel && matchesStatus
  })
})
const hasFilters = computed(() => Boolean(search.value.trim() || levelFilter.value || statusFilter.value))
const canSubmit = computed(() => Boolean(form.name.trim() && form.level && form.level >= 1 && form.level <= 4))
const selectedLevelPreview = computed(() => (form.level ? LEVEL_PREVIEW[form.level] : ''))
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
  return /^Sales Level [1-4]$/.test(role.name)
}

function hasDuplicateName(role: SalesRole) {
  return (duplicateRoleNames.value.get(normalizeRoleName(role.name)) ?? 0) > 1
}

function load() {
  error.value = ''
  store.fetchSalesRoles().catch((e) => { error.value = store.errorMessage(e) })
}

function openCreate() {
  editingRole.value = null
  form.name = ''
  form.level = null
  form.description = ''
  formError.value = ''
  dialogVisible.value = true
}

function openEdit(role: SalesRole) {
  editingRole.value = role
  form.name = role.name
  form.level = role.level
  form.description = role.description ?? ''
  formError.value = ''
  dialogVisible.value = true
}

async function saveRole() {
  if (!canSubmit.value || !form.level) return
  if (store.savingSalesRole) return
  formError.value = ''
  try {
    if (editingRole.value) {
      await store.updateSalesRole(editingRole.value.id, { name: form.name.trim(), level: form.level, description: form.description.trim() || null })
      toast.add({ severity: 'success', summary: 'Role Updated', detail: 'Sales role has been updated.', life: 3000 })
    } else {
      await store.createSalesRole({ name: form.name.trim(), level: form.level, description: form.description.trim() })
      toast.add({ severity: 'success', summary: 'Role Created', detail: 'Sales role has been created.', life: 3000 })
    }
    dialogVisible.value = false
  } catch (e) {
    formError.value = store.errorMessage(e)
  }
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
      <Button label="Create Role" icon="pi pi-plus" size="small" @click="openCreate" />
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div v-if="showGuidance" class="guidance-banner">
      <div class="guidance-icon"><i class="pi pi-lightbulb" /></div>
      <div>
        <strong>Role names are display labels</strong>
        <span>Keep names unique for operators, but hierarchy and reporting still depend on numeric levels only.</span>
      </div>
      <Button icon="pi pi-times" text rounded size="small" aria-label="Dismiss guidance" @click="showGuidance = false" />
    </div>

    <div class="info-section">
      <div class="info-body">
        <div class="info-title"><i class="pi pi-info-circle" /> How sales roles work</div>
        <ul class="info-list">
          <li><span>Role Name is flexible and may be renamed.</span></li>
          <li><span>Role Level controls the reporting hierarchy.</span></li>
          <li><span>Level 1 is the highest organizational level.</span></li>
          <li><span>Level 4 is an individual/self-scope level.</span></li>
          <li><span>Position and user assignment are managed in Sales Structure.</span></li>
          <li><span>Permissions must not be inferred from the role name.</span></li>
        </ul>
      </div>
      <div class="hierarchy-guide">
        <div v-for="lv in LEVEL_ORDER" :key="lv" class="hier-card">
          <span class="hier-level">Level {{ lv }}</span>
          <span class="hier-desc">{{ LEVEL_HIERARCHY[lv] }}</span>
        </div>
      </div>
    </div>

    <div class="summary-grid">
      <div class="summary-card"><span>Total Roles</span><strong>{{ totalRoles }}</strong></div>
      <div class="summary-card"><span>Active Roles</span><strong>{{ activeRoles }}</strong></div>
      <div class="summary-card wide">
        <span>Level Breakdown</span>
        <div class="level-breakdown">
          <Tag v-for="item in levelCounts" :key="item.level" :value="`Level ${item.level}: ${item.count}`" :severity="levelSeverity(item.level)" />
        </div>
      </div>
    </div>

    <div class="filter-panel">
      <div class="search-field">
        <i class="pi pi-search" />
        <InputText v-model="search" placeholder="Search role name or description..." />
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
      <DataTable v-else :value="filteredRoles" :loading="store.salesRolesLoading" dataKey="id" scrollable>
        <template #empty>
          <div class="empty-state">
            <div class="empty-icon"><i class="pi pi-id-card" /></div>
            <strong>No sales roles found</strong>
            <span>{{ hasFilters ? 'Adjust your filters to view roles.' : 'Create a role to begin configuring the sales hierarchy.' }}</span>
            <Button v-if="!hasFilters" label="Create Role" icon="pi pi-plus" size="small" @click="openCreate" />
          </div>
        </template>
        <Column field="name" header="Role Name" :style="{ minWidth: '200px' }">
          <template #body="{ data }">
            <div class="role-name-cell">
              <strong class="role-name">{{ data.name }}</strong>
              <Tag v-if="isSystemDefaultRole(data)" value="System Default" severity="info" size="small" class="template-tag" />
              <Tag v-if="isDemoRole(data)" value="Demo" severity="warn" size="small" class="template-tag" />
              <span v-if="hasDuplicateName(data)" class="warning-chip"><i class="pi pi-exclamation-triangle" /> Name not unique</span>
            </div>
          </template>
        </Column>
        <Column header="Level" :style="{ width: '130px' }"><template #body="{ data }"><Tag :value="`Level ${data.level}`" :severity="levelSeverity(data.level)" /></template></Column>
        <Column header="Description" :style="{ minWidth: '280px' }">
          <template #body="{ data }">
            <template v-if="data.description">
              <span class="cell-text">{{ data.description }}</span>
            </template>
            <template v-else>
              <div class="cell-text">{{ roleDisplayDescription(data) }}</div>
              <span class="cell-hint">Edit this role to customize the name and description.</span>
            </template>
          </template>
        </Column>
        <Column header="Status" :style="{ width: '120px' }"><template #body="{ data }"><Tag :value="data.isActive ? 'Active' : 'Inactive'" :severity="data.isActive ? 'success' : 'secondary'" /></template></Column>
        <Column header="Updated At" :style="{ width: '150px' }"><template #body="{ data }">{{ formatDate(data.updatedAt) }}</template></Column>
        <Column header="Actions" :style="{ width: '230px' }">
          <template #body="{ data }">
            <details class="action-menu">
              <summary><i class="pi pi-ellipsis-v" /><span class="sr-only">Actions</span></summary>
              <div class="action-popover">
                <button type="button" @click="openEdit(data)"><i class="pi pi-pencil" /> Edit</button>
                <button type="button" :class="{ danger: data.isActive, success: !data.isActive }" :disabled="store.savingSalesRole" @click="toggleStatus(data)">
                  <i :class="data.isActive ? 'pi pi-ban' : 'pi pi-check-circle'" /> {{ data.isActive ? 'Deactivate' : 'Activate' }}
                </button>
              </div>
            </details>
          </template>
        </Column>
      </DataTable>
    </div>

    <Dialog v-model:visible="dialogVisible" :header="editingRole ? 'Edit Sales Role' : 'Create Sales Role'" modal :style="{ width: 'min(560px, 92vw)' }">
      <div class="dialog-form">
        <Message v-if="formError" severity="error">{{ formError }}</Message>
        <div class="field-note"><i class="pi pi-info-circle" /><span>Name is flexible. Numeric level controls hierarchy.</span></div>
        <div class="form-field"><label>Role Name</label><InputText v-model="form.name" autofocus /></div>
        <div class="form-field">
          <label>Level</label>
          <Select v-model="form.level" :options="levelOptions" optionLabel="label" optionValue="value" placeholder="Select level" />
          <div v-if="selectedLevelPreview" class="level-preview">
            <span class="level-preview-title">Selected hierarchy:</span>
            <span>{{ selectedLevelPreview }}</span>
          </div>
        </div>
        <div class="form-field"><label>Description <span class="optional-label">(optional)</span></label><Textarea v-model="form.description" rows="4" autoResize /></div>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="dialogVisible = false" :disabled="store.savingSalesRole" />
        <Button label="Save" icon="pi pi-check" :loading="store.savingSalesRole" :disabled="!canSubmit || store.savingSalesRole" @click="saveRole" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; background: #f7f9fb; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: #0b7766; }
h1 { margin: 0.2rem 0 0.2rem; font-size: 1.6rem; color: #0f172a; }
.muted { margin: 0; color: #7c8798; font-size: 0.85rem; }
.summary-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.85rem; }
.summary-card, .filter-panel, .table-panel, .info-section, .guidance-banner { background: #fff; border: 1px solid #edf1f6; border-radius: 12px; box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03); }
.summary-card { padding: 1rem; display: grid; gap: 0.25rem; }
.summary-card span { color: #64748b; font-size: 0.76rem; font-weight: 700; text-transform: uppercase; }
.summary-card strong { color: #0f172a; font-size: 1.25rem; }
.level-breakdown { display: flex; flex-wrap: wrap; gap: 0.4rem; }
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
.filter-panel { display: grid; grid-template-columns: minmax(260px, 1fr) 200px 160px auto; gap: 0.75rem; align-items: end; padding: 1rem; }
.search-field { display: flex; align-items: center; gap: 0.6rem; min-height: 2.55rem; padding: 0 0.75rem; background: #fff; border: 1px solid #dbe3ee; border-radius: 8px; }
.search-field i { color: #94a3b8; }
.search-field :deep(.p-inputtext) { flex: 1; border: 0; padding-inline: 0; box-shadow: none; }
.filter-field { display: grid; gap: 0.35rem; }
.filter-field label { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; color: #64748b; }
.optional-label { font-weight: 500; text-transform: none; color: #94a3b8; }
.table-panel { overflow-x: auto; }
.table-panel :deep(.p-datatable), .table-panel :deep(.p-datatable-table), .table-panel :deep(.p-datatable-table-container) { background: #ffffff; color: #0f172a; }
.table-panel :deep(.p-datatable-thead > tr > th) { background: #f8fafc; color: #475569; font-size: 0.68rem; text-transform: uppercase; border-color: #edf1f6; }
.table-panel :deep(.p-datatable-tbody > tr), .table-panel :deep(.p-datatable-tbody > tr > td) { background: #ffffff; color: #1e293b; border-color: #f1f4f8; }
.table-panel :deep(.p-datatable-tbody > tr:hover > td) { background: #f8fafc; }
.table-panel :deep(.p-datatable-loading-overlay) { background: rgba(255, 255, 255, 0.72); }
.role-name { color: #0f172a; font-weight: 750; }
.role-name-cell { display: flex; align-items: center; gap: 0.45rem; flex-wrap: wrap; }
.template-tag { opacity: 0.85; }
.warning-chip { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.12rem 0.45rem; border-radius: 999px; background: #fef2f2; color: #b91c1c; font-size: 0.68rem; font-weight: 800; }
.cell-text { color: #475569; }
.cell-hint { display: block; margin-top: 0.2rem; color: #94a3b8; font-size: 0.72rem; }
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
.dialog-form { display: grid; gap: 1rem; }
.form-field { display: grid; gap: 0.35rem; }
.form-field label { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; color: #64748b; }
.field-note { display: flex; gap: 0.5rem; align-items: center; padding: 0.65rem 0.75rem; border: 1px solid #bfdbfe; border-radius: 8px; background: #eff6ff; color: #1e40af; font-size: 0.8rem; }
.level-preview { display: grid; gap: 0.2rem; margin-top: 0.45rem; padding: 0.6rem 0.75rem; border: 1px solid #dbeafe; border-radius: 8px; background: #eff6ff; color: #1e40af; font-size: 0.8rem; line-height: 1.45; }
.level-preview-title { font-weight: 700; }
:deep(.p-inputtext), :deep(.p-select), :deep(.p-textarea) { background: #ffffff; color: #0f172a; border-color: #dbe3ee; }
:deep(.p-select-label), :deep(.p-select-dropdown), :deep(.p-textarea::placeholder), :deep(.p-inputtext::placeholder) { color: #64748b; }
:deep(.p-dialog), :deep(.p-dialog-header), :deep(.p-dialog-content), :deep(.p-dialog-footer) { background: #ffffff; color: #0f172a; }
:deep(.p-select-overlay), :deep(.p-select-list), :deep(.p-select-option) { background: #ffffff; color: #0f172a; }
:deep(.p-select-option.p-select-option-selected), :deep(.p-select-option:hover) { background: #f1f5f9; color: #0f172a; }
@media (max-width: 900px) { .filter-panel { grid-template-columns: 1fr 1fr; } .search-field { grid-column: 1 / -1; } .info-section { grid-template-columns: 1fr; } }
@media (max-width: 768px) { .admin-page { padding: 1rem; } .summary-grid, .filter-panel { grid-template-columns: 1fr; } .hierarchy-guide { grid-template-columns: 1fr; } }
</style>
