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
import Dialog from 'primevue/dialog'
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
const actionDialogVisible = ref(false)
const actionTarget = ref<SalesRole | null>(null)

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
const rolesByLevel = computed(() =>
  LEVEL_ORDER.map((level) => ({
    level,
    guide: LEVEL_GUIDE[level],
    roles: groupedRoles.value.filter((role) => role.level === level),
  })).filter((group) => group.roles.length > 0),
)
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

function systemRoleLabel(role: SalesRole) {
  if (role.level === 1) return 'SUPER_ADMIN'
  if (role.level === 4) return 'SALES_EXECUTIVE'
  return 'SALES_MANAGER'
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

function openActions(role: SalesRole) {
  actionTarget.value = role
  actionDialogVisible.value = true
}

function closeActions() {
  actionDialogVisible.value = false
  actionTarget.value = null
}

function viewSelectedRole() {
  if (!actionTarget.value) return
  const role = actionTarget.value
  closeActions()
  goView(role)
}

function editSelectedRole() {
  if (!actionTarget.value) return
  const role = actionTarget.value
  closeActions()
  goEdit(role)
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
  <section class="roles-page">
    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <header class="roles-toolbar">
      <div class="toolbar-title">
        <span class="eyebrow">Sales Organization</span>
        <h1>Role Management</h1>
        <p>{{ filteredRoles.length }} visible · {{ activeRoles }} active</p>
      </div>

      <div class="toolbar-controls">
        <div class="search-field">
          <i class="pi pi-search" />
          <InputText
            v-model="search"
            placeholder="Search role name, description, or landing page"
          />
        </div>

        <Select
          v-model="levelFilter"
          :options="levelFilterOptions"
          optionLabel="label"
          optionValue="value"
          class="toolbar-select"
        />

        <Select
          v-model="statusFilter"
          :options="statusOptions"
          optionLabel="label"
          optionValue="value"
          class="toolbar-select"
        />

        <Button
          label="Reset"
          icon="pi pi-refresh"
          severity="secondary"
          outlined
          size="small"
          class="reset-button"
          @click="clearFilters"
        />

        <Button
          label="Add Role"
          icon="pi pi-plus"
          size="small"
          class="create-button"
          @click="goCreate"
        />
      </div>
    </header>

    <div class="summary-strip">
      <div class="summary-item">
        <span>Total Roles</span>
        <strong>{{ totalRoles }}</strong>
      </div>
      <div class="summary-item">
        <span>Active Roles</span>
        <strong>{{ activeRoles }}</strong>
      </div>
      <div
        v-for="item in levelCounts"
        :key="item.level"
        class="summary-item"
      >
        <span>Level {{ item.level }}</span>
        <strong>{{ item.count }}</strong>
      </div>
    </div>

    <div v-if="store.salesRolesLoading && !roleList.length" class="skeleton-area">
      <Skeleton v-for="n in 7" :key="n" class="skeleton-row" />
    </div>

    <div v-else-if="!rolesByLevel.length" class="empty-state">
      <span class="empty-icon"><i class="pi pi-id-card" /></span>
      <strong>No roles found</strong>
      <span>
        {{
          hasFilters
            ? 'Adjust the current filters.'
            : 'Create a role to begin configuring access.'
        }}
      </span>
      <Button
        v-if="!hasFilters"
        label="Add Role"
        icon="pi pi-plus"
        size="small"
        @click="goCreate"
      />
    </div>

    <div v-else class="level-sections">
      <section
        v-for="group in rolesByLevel"
        :key="group.level"
        class="level-section"
      >
        <header class="level-section-header">
          <div class="level-section-title">
            <span class="level-number" :class="`level-${group.level}`">
              {{ group.level }}
            </span>
            <div>
              <strong>Level {{ group.level }}</strong>
              <span>{{ group.guide }}</span>
            </div>
          </div>

          <span class="role-count">
            {{ group.roles.length }} role{{ group.roles.length === 1 ? '' : 's' }}
          </span>
        </header>

        <div class="role-list-header" aria-hidden="true">
          <span>System Role</span>
          <span>Role</span>
          <span>Initial Menu</span>
          <span>Permissions</span>
          <span>Status</span>
          <span>Actions</span>
        </div>

        <article
          v-for="role in group.roles"
          :key="role.id"
          class="role-row"
          tabindex="0"
          @click="openActions(role)"
          @keydown.enter="openActions(role)"
        >
          <div class="system-role-cell">
            <span class="mobile-label">System Role</span>
            <Tag
              :value="systemRoleLabel(role)"
              :severity="role.level === 1 ? 'info' : role.level === 4 ? 'secondary' : 'success'"
              rounded
              class="status-tag"
            />
          </div>

          <div class="role-primary">
            <div class="role-icon" :class="`level-${role.level}`">
              {{ role.level }}
            </div>

            <div class="role-copy">
              <div class="role-title-line">
                <strong>{{ role.name }}</strong>

                <Tag
                  v-if="isSystemDefaultRole(role)"
                  value="System"
                  severity="info"
                  rounded
                  class="mini-tag"
                />

                <Tag
                  v-if="isDemoRole(role)"
                  value="Demo"
                  severity="warn"
                  rounded
                  class="mini-tag"
                />

                <span
                  v-if="hasDuplicateName(role)"
                  class="duplicate-warning"
                  title="Duplicate role name"
                >
                  <i class="pi pi-exclamation-triangle" />
                </span>
              </div>

              <span class="role-description">
                {{ roleDisplayDescription(role) }}
              </span>
            </div>
          </div>

          <div class="role-menu">
            <span class="mobile-label">Initial Menu</span>
            <code v-if="role.landingPage" class="route-badge">
              {{ role.landingPage }}
            </code>
            <span v-else class="muted-value">Not configured</span>
          </div>

          <div class="permission-cell">
            <span class="mobile-label">Permissions</span>
            <strong>{{ permissionCount(role) ?? 0 }}</strong>
            <span>permissions</span>
          </div>

          <div class="role-status">
            <span class="mobile-label">Status</span>
            <Tag
              :value="role.isActive ? 'Active' : 'Inactive'"
              :severity="role.isActive ? 'success' : 'secondary'"
              rounded
              class="status-tag"
            />
          </div>

          <div class="row-actions" @click.stop>
            <Button
              icon="pi pi-ellipsis-v"
              text
              rounded
              size="small"
              class="more-action"
              title="Open role actions"
              aria-label="Open role actions"
              @click="openActions(role)"
            />
          </div>
        </article>
      </section>
    </div>

    <Dialog
      v-model:visible="actionDialogVisible"
      modal
      :draggable="false"
      :closable="false"
      :dismissableMask="true"
      class="role-action-dialog"
      :style="{ width: 'min(540px, calc(100vw - 2rem))' }"
      @hide="closeActions"
    >
      <div v-if="actionTarget" class="action-dialog-content">
        <div class="action-dialog-header">
          <div>
            <span class="dialog-eyebrow">Role Actions</span>
            <h2>{{ actionTarget.name }}</h2>
            <p>
              Level {{ actionTarget.level }}
              · {{ permissionCount(actionTarget) ?? 0 }} permissions
            </p>
          </div>

          <Button
            icon="pi pi-times"
            text
            rounded
            severity="secondary"
            aria-label="Close"
            @click="closeActions"
          />
        </div>

        <div class="action-grid">
          <button class="action-card" type="button" @click="viewSelectedRole">
            <span class="action-card-icon view-icon">
              <i class="pi pi-eye" />
            </span>
            <span>
              <strong>View Detail</strong>
              <small>Review role information, landing page, and permissions.</small>
            </span>
          </button>

          <button class="action-card" type="button" @click="editSelectedRole">
            <span class="action-card-icon edit-icon">
              <i class="pi pi-pencil" />
            </span>
            <span>
              <strong>Edit Role</strong>
              <small>Update role name, level, description, and permissions.</small>
            </span>
          </button>
        </div>

        <div class="role-detail-grid">
          <div>
            <span>Hierarchy Level</span>
            <strong>Level {{ actionTarget.level }}</strong>
          </div>
          <div>
            <span>Initial Open Menu</span>
            <strong>{{ actionTarget.landingPage || 'Not configured' }}</strong>
          </div>
          <div>
            <span>Permission Count</span>
            <strong>{{ permissionCount(actionTarget) ?? 0 }}</strong>
          </div>
          <div>
            <span>Status</span>
            <Tag
              :value="actionTarget.isActive ? 'Active' : 'Inactive'"
              :severity="actionTarget.isActive ? 'success' : 'secondary'"
              rounded
            />
          </div>
        </div>

        <div class="status-zone">
          <div class="zone-copy">
            <i :class="actionTarget.isActive ? 'pi pi-ban' : 'pi pi-check-circle'" />
            <div>
              <strong>
                {{ actionTarget.isActive ? 'Deactivate Role' : 'Activate Role' }}
              </strong>
              <span>
                {{
                  actionTarget.isActive
                    ? 'Prevent this role from being selected for new accounts.'
                    : 'Make this role available for account assignment.'
                }}
              </span>
            </div>
          </div>

          <Button
            :label="actionTarget.isActive ? 'Deactivate' : 'Activate'"
            :icon="actionTarget.isActive ? 'pi pi-ban' : 'pi pi-check-circle'"
            :severity="actionTarget.isActive ? 'danger' : 'success'"
            outlined
            size="small"
            :loading="store.savingSalesRole"
            @click="toggleStatus(actionTarget).then(closeActions)"
          />
        </div>

        <div class="delete-zone">
          <div class="zone-copy">
            <i class="pi pi-trash" />
            <div>
              <strong>Delete Role</strong>
              <span>
                {{
                  isSystemDefaultRole(actionTarget)
                    ? 'System default roles are protected and cannot be deleted.'
                    : 'Delete this role when it is no longer used by accounts or assignments.'
                }}
              </span>
            </div>
          </div>

          <Button
            label="Delete"
            icon="pi pi-trash"
            severity="danger"
            size="small"
            :disabled="isSystemDefaultRole(actionTarget) || store.savingSalesRole"
            @click="deleteRole(actionTarget).then(closeActions)"
          />
        </div>
      </div>
    </Dialog>
  </section>
</template>

<style scoped>
.roles-page {
  width: 100%;
  min-width: 0;
  min-height: 100vh;
  padding: 0;
  overflow-x: hidden;
  background: #f8fafc;
}

.page-message {
  margin: 0.75rem 1rem 0;
}

.roles-toolbar {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  min-width: 0;
  padding: 0.8rem 1rem;
  border-bottom: 1px solid #e5eaf0;
  background: rgba(255, 255, 255, 0.97);
  backdrop-filter: blur(10px);
}

.toolbar-title {
  display: grid;
  flex: 0 0 auto;
  min-width: 170px;
  gap: 0.08rem;
}

.eyebrow {
  color: #64748b;
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.toolbar-title h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.12rem;
}

.toolbar-title p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.7rem;
}

.toolbar-controls {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 180px 160px auto auto;
  align-items: center;
  gap: 0.55rem;
  width: min(980px, 100%);
  min-width: 0;
}

.search-field {
  display: flex;
  align-items: center;
  min-width: 0;
  height: 38px;
  gap: 0.55rem;
  padding: 0 0.75rem;
  border: 1px solid #dfe5ec;
  border-radius: 9px;
  background: #fff;
}

.search-field:focus-within {
  border-color: #d14350;
  box-shadow: 0 0 0 3px rgba(209, 67, 80, 0.08);
}

.search-field i {
  color: #94a3b8;
  font-size: 0.82rem;
}

.search-field :deep(.p-inputtext) {
  width: 100%;
  min-width: 0;
  padding: 0;
  border: 0;
  box-shadow: none;
  background: transparent;
  font-size: 0.78rem;
}

.toolbar-select {
  min-width: 0;
}

.toolbar-select :deep(.p-select-label) {
  padding-top: 0.54rem;
  padding-bottom: 0.54rem;
  font-size: 0.75rem;
}

.reset-button,
.create-button {
  white-space: nowrap;
}

.summary-strip {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  border-bottom: 1px solid #e5eaf0;
  background: #fff;
}

.summary-item {
  display: grid;
  gap: 0.12rem;
  padding: 0.7rem 1rem;
  border-right: 1px solid #edf1f6;
}

.summary-item:last-child {
  border-right: 0;
}

.summary-item span {
  color: #94a3b8;
  font-size: 0.62rem;
  font-weight: 800;
  text-transform: uppercase;
}

.summary-item strong {
  color: #0f172a;
  font-size: 0.95rem;
}

.level-sections {
  display: grid;
  gap: 0.9rem;
  padding: 0.9rem 1rem 1.4rem;
}

.level-section {
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.level-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.7rem 0.85rem;
  border-bottom: 1px solid #e5eaf0;
  background: #f8fafc;
}

.level-section-title {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.level-section-title > div {
  display: grid;
  gap: 0.06rem;
}

.level-section-title strong {
  color: #0f172a;
  font-size: 0.82rem;
}

.level-section-title span:last-child {
  color: #64748b;
  font-size: 0.68rem;
}

.level-number {
  display: grid;
  width: 30px;
  height: 30px;
  place-content: center;
  border-radius: 8px;
  font-size: 0.74rem;
  font-weight: 850;
}

.level-1 { background: #fff1f2; color: #bb3342; }
.level-2 { background: #ecfdf5; color: #047857; }
.level-3 { background: #fff7ed; color: #c2410c; }
.level-4 { background: #f1f5f9; color: #475569; }

.role-count {
  color: #64748b;
  font-size: 0.67rem;
  font-weight: 700;
}

.role-list-header,
.role-row {
  display: grid;
  grid-template-columns: minmax(130px, 0.8fr) minmax(260px, 2fr) minmax(150px, 1fr) 110px 100px 60px;
  align-items: center;
  gap: 0.7rem;
}

.system-role-cell {
  min-width: 0;
}

.role-list-header {
  padding: 0.45rem 0.8rem;
  border-bottom: 1px solid #edf1f6;
  color: #94a3b8;
  font-size: 0.62rem;
  font-weight: 800;
  text-transform: uppercase;
}

.role-row {
  min-height: 66px;
  padding: 0.58rem 0.8rem;
  border-bottom: 1px solid #edf1f6;
  cursor: pointer;
  outline: none;
}

.role-row:last-child {
  border-bottom: 0;
}

.role-row:hover,
.role-row:focus-visible {
  background: #fffbfb;
}

.role-primary {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.65rem;
}

.role-icon {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  place-content: center;
  border-radius: 9px;
  font-size: 0.75rem;
  font-weight: 850;
}

.role-copy {
  display: grid;
  min-width: 0;
  gap: 0.12rem;
}

.role-title-line {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.35rem;
}

.role-title-line strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.8rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.role-description {
  display: -webkit-box;
  overflow: hidden;
  color: #64748b;
  font-size: 0.68rem;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.mini-tag {
  flex: 0 0 auto;
  font-size: 0.58rem;
}

.duplicate-warning {
  color: #b45309;
  font-size: 0.7rem;
}

.role-menu,
.permission-cell,
.role-status {
  min-width: 0;
}

.route-badge {
  display: inline-block;
  overflow: hidden;
  max-width: 100%;
  padding: 0.16rem 0.42rem;
  border: 1px solid #ffd9dd;
  border-radius: 6px;
  background: #fff1f2;
  color: #d14350;
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 0.65rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.muted-value {
  color: #94a3b8;
  font-size: 0.68rem;
}

.permission-cell {
  display: grid;
  gap: 0.02rem;
}

.permission-cell strong {
  color: #0f172a;
  font-size: 0.8rem;
}

.permission-cell > span:last-child {
  color: #94a3b8;
  font-size: 0.62rem;
}

.status-tag {
  font-size: 0.64rem;
}

.mobile-label {
  display: none;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
}

.more-action {
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  border: 1px solid #dfe5ec !important;
  background: #fff !important;
  color: #64748b !important;
}

.more-action:hover {
  border-color: #cbd5e1 !important;
  background: #f8fafc !important;
  color: #0f172a !important;
}

.role-action-dialog :deep(.p-dialog) {
  overflow: hidden;
  border: 1px solid #dfe5ec;
  border-radius: 18px;
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.2);
}

.role-action-dialog :deep(.p-dialog-header) {
  display: none;
}

.role-action-dialog :deep(.p-dialog-content) {
  padding: 0;
  background: #fff;
}

.action-dialog-content {
  padding: 1.2rem;
}

.action-dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1rem;
}

.dialog-eyebrow {
  color: #94a3b8;
  font-size: 0.65rem;
  font-weight: 800;
  text-transform: uppercase;
}

.action-dialog-header h2 {
  margin: 0.15rem 0 0;
  color: #0f172a;
  font-size: 1.15rem;
}

.action-dialog-header p {
  margin: 0.2rem 0 0;
  color: #64748b;
  font-size: 0.76rem;
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.7rem;
}

.action-card {
  display: flex;
  align-items: flex-start;
  min-height: 92px;
  gap: 0.7rem;
  padding: 0.9rem;
  border: 1px solid #dfe5ec;
  border-radius: 12px;
  background: #fff;
  text-align: left;
  cursor: pointer;
}

.action-card:hover {
  border-color: #f3b9c0;
  background: #fffbfb;
}

.action-card-icon {
  display: grid;
  width: 32px;
  height: 32px;
  flex: 0 0 auto;
  place-content: center;
  border-radius: 8px;
}

.view-icon { background: #fff1f2; color: #d14350; }
.edit-icon { background: #fff7ed; color: #ea580c; }

.action-card > span:last-child {
  display: grid;
  gap: 0.2rem;
}

.action-card strong {
  color: #0f172a;
  font-size: 0.84rem;
}

.action-card small {
  color: #64748b;
  font-size: 0.71rem;
  line-height: 1.45;
}

.role-detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
  margin-top: 0.8rem;
  padding: 0.75rem;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  background: #f8fafc;
}

.role-detail-grid > div {
  display: grid;
  gap: 0.12rem;
}

.role-detail-grid span {
  color: #94a3b8;
  font-size: 0.62rem;
  font-weight: 700;
  text-transform: uppercase;
}

.role-detail-grid strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.76rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-zone,
.delete-zone {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 0.8rem;
  padding: 0.8rem;
  border-radius: 11px;
}

.status-zone {
  border: 1px solid #fed7aa;
  background: #fffaf2;
}

.delete-zone {
  border: 1px solid #fecaca;
  background: #fff7f7;
}

.zone-copy {
  display: flex;
  align-items: flex-start;
  min-width: 0;
  gap: 0.65rem;
}

.zone-copy > i {
  margin-top: 0.1rem;
  color: #dc2626;
}

.zone-copy > div {
  display: grid;
  gap: 0.15rem;
}

.zone-copy strong {
  color: #991b1b;
  font-size: 0.8rem;
}

.zone-copy span {
  color: #7f1d1d;
  font-size: 0.7rem;
  line-height: 1.4;
}

.skeleton-area {
  padding: 1rem;
}

.skeleton-row {
  height: 3.2rem;
  margin: 0.45rem 0;
  border-radius: 8px;
}

.empty-state {
  display: flex;
  min-height: 320px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  background: #fff;
  color: #94a3b8;
  text-align: center;
}

.empty-state strong {
  color: #0f172a;
}

.empty-icon {
  display: grid;
  width: 48px;
  height: 48px;
  place-content: center;
  border-radius: 12px;
  background: #f1f5f9;
}

@media (max-width: 1050px) {
  .roles-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-controls {
    width: 100%;
  }

  .role-list-header,
  .role-row {
    grid-template-columns: minmax(260px, 2fr) minmax(140px, 1fr) 100px 90px 52px;
  }
}

@media (max-width: 820px) {
  .toolbar-controls {
    grid-template-columns: minmax(0, 1fr) 170px 150px auto;
  }

  .create-button {
    grid-column: 1 / -1;
    justify-self: end;
  }

  .summary-strip {
    grid-template-columns: repeat(3, 1fr);
  }

  .role-list-header {
    display: none;
  }

  .role-row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.5rem 1rem;
    padding: 0.8rem;
  }

  .role-primary {
    grid-column: 1 / 2;
  }

  .row-actions {
    grid-column: 2;
    grid-row: 1;
  }

  .role-menu,
  .permission-cell,
  .role-status {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    grid-column: 1 / -1;
    padding-left: 2.65rem;
  }

  .mobile-label {
    display: inline;
    min-width: 88px;
    color: #94a3b8;
    font-size: 0.62rem;
    font-weight: 800;
    text-transform: uppercase;
  }
}

@media (max-width: 620px) {
  .roles-toolbar {
    padding: 0.75rem;
  }

  .toolbar-controls {
    grid-template-columns: 1fr 1fr;
  }

  .search-field {
    grid-column: 1 / -1;
  }

  .reset-button,
  .create-button {
    width: 100%;
  }

  .create-button {
    grid-column: auto;
  }

  .summary-strip {
    grid-template-columns: repeat(2, 1fr);
  }

  .level-sections {
    padding: 0.7rem;
  }

  .level-section-header {
    align-items: flex-start;
  }

  .level-section-title span:last-child {
    display: none;
  }

  .action-grid,
  .role-detail-grid {
    grid-template-columns: 1fr;
  }

  .status-zone,
  .delete-zone {
    align-items: stretch;
    flex-direction: column;
  }

  .status-zone :deep(.p-button),
  .delete-zone :deep(.p-button) {
    width: 100%;
  }
}
</style>
