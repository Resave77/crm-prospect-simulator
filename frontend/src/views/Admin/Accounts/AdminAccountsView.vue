<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import Checkbox from 'primevue/checkbox'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Toast from 'primevue/toast'
import AdminPageHeader from '../../../components/admin/AdminPageHeader.vue'
import AdminTableShell from '../../../components/admin/AdminTableShell.vue'
import { useToast } from 'primevue/usetoast'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAdminStore } from '../../../stores/admin'
import { useAuthStore } from '../../../stores/auth'
import type { AdminUserListItem } from '../../../types/admin'

const store = useAdminStore()
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const toast = useToast()
const error = ref('')
const showFilters = ref(false)
const updating = ref(false)
const first = ref(0)
const goToPage = ref(1)
const selectedIds = ref<string[]>([])
const bulkDeleteVisible = ref(false)
const deactivateDialogVisible = ref(false)
const deactivateTarget = ref<{ id: string; name: string } | null>(null)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<AdminUserListItem | null>(null)
const actionDialogVisible = ref(false)
const actionTarget = ref<AdminUserListItem | null>(null)
const trashVisible = ref(route.query.view === 'trash')
const restoringId = ref<string | null>(null)
const trashedUsers = computed(() => store.users.filter((user) => user.status === 'INACTIVE'))
watch(() => route.query.view, (value) => {
  trashVisible.value = value === 'trash'
  store.setParam('includeDeleted', trashVisible.value)
  store.setParam('page', 1)
  first.value = 0
  store.users = []
  store.total = 0
  load()
}, { immediate: true })
watch(() => route.query.search, (value) => { store.setParam('search', typeof value === 'string' ? value : ''); store.setParam('page', 1); first.value = 0; load() }, { immediate: true })
watch(() => route.query.status, (value) => {
  const status = typeof value === 'string' && (value === 'ACTIVE' || value === 'INACTIVE') ? value : ''
  store.setParam('status', status)
  store.setParam('page', 1)
  first.value = 0
  load()
}, { immediate: true })

const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Inactive', value: 'INACTIVE' },
]
const roleOptions = [
  { label: 'All CRM Roles', value: '' },
  { label: 'Super Admin', value: 'SUPER_ADMIN' },
  { label: 'Administrator', value: 'ADMINISTRATOR' },
  { label: 'Sales Manager', value: 'SALES_MANAGER' },
  { label: 'Sales Executive', value: 'SALES_EXECUTIVE' },
]

const selectedStatus = computed({
  get: () => store.params.status,
  set: (val) => { store.setParam('status', val); store.setParam('page', 1); first.value = 0; load() },
})

function load() {
  error.value = ''
  store.setParam('includeDeleted', trashVisible.value)
  store.fetchUsers().catch((e) => { error.value = store.errorMessage(e) })
}

function resetAll() {
  store.resetFilters()
  first.value = 0
  load()
}

function onPage(event: { first: number; rows: number; page: number }) {
  if (event.rows !== store.limit) {
    store.setParam('limit', event.rows)
    store.setParam('page', 1)
    first.value = 0
  } else {
    store.setPage(event.page + 1)
  }
  load()
}

async function restoreAccount(user: AdminUserListItem) {
  restoringId.value = user.id
  try {
    await store.restoreUser(user.id)
    toast.add({ severity: 'success', summary: 'Account Restored', detail: `${accountDisplayName(user)} dikembalikan ke daftar akun.`, life: 3500 })
  } catch (e) { error.value = store.errorMessage(e) } finally { restoringId.value = null }
}

const allPageSelected = computed(() => store.users.length > 0 && store.users.every((user) => selectedIds.value.includes(user.id)))

function toggleSelected(id: string, checked: boolean) {
  selectedIds.value = checked
    ? Array.from(new Set([...selectedIds.value, id]))
    : selectedIds.value.filter((selectedId) => selectedId !== id)
}

function toggleSelectAll(checked: boolean) {
  selectedIds.value = checked
    ? Array.from(new Set([...selectedIds.value, ...store.users.map((user) => user.id)]))
    : selectedIds.value.filter((id) => !store.users.some((user) => user.id === id))
}

function cancelSelection() {
  selectedIds.value = []
}

async function executeBulkDelete() {
  if (!selectedIds.value.length) return
  updating.value = true
  error.value = ''
  try {
    await Promise.all(selectedIds.value.map((id) => store.deleteUser(id)))
    selectedIds.value = []
    bulkDeleteVisible.value = false
    await load()
    toast.add({ severity: 'success', summary: 'Soft Delete Berhasil', detail: 'Akun terpilih berhasil dipindahkan ke trash.', life: 3500 })
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    updating.value = false
  }
}

const pageNumbers = computed(() => Array.from({ length: store.pages }, (_, index) => index + 1))

const selectedPageSize = computed({
  get: () => store.limit,
  set: (value: number) => {
    store.setParam('limit', Number(value))
    store.setParam('page', 1)
    first.value = 0
    goToPage.value = 1
    load()
  },
})
const selectedRole = computed({
  get: () => store.params.role,
  set: (val) => { store.setParam('role', val); store.setParam('page', 1); first.value = 0; load() },
})

function goToSelectedPage() {
  const target = Math.min(Math.max(Number(goToPage.value) || 1, 1), Math.max(store.pages, 1))
  goToPage.value = target
  store.setPage(target)
  first.value = (target - 1) * store.limit
  load()
}

function selectPage(page: number) {
  goToPage.value = page
  store.setPage(page)
  first.value = (page - 1) * store.limit
  load()
}

const isSelf = (id: string) => id === auth.user?.id
const isProtectedSuperAdmin = (user: AdminUserListItem) => user.email === 'admin@yummy.test' || user.fullName === 'Yummy Super Admin'

function fallback(value?: string | null) {
  return value?.trim() || '-'
}

function accountDisplayName(user?: Pick<AdminUserListItem, 'fullName' | 'email' | 'employeeId'> | null) {
  return user?.fullName?.trim() || user?.email?.trim() || user?.employeeId?.trim() || 'this account'
}

function organizationalRoleLabel(user: AdminUserListItem) {
  return fallback(user.organizationalRole?.name)
}

function organizationalRoleMeta(user: AdminUserListItem) {
  const role = user.organizationalRole
  if (!role) return 'No organizational role'
  return `${role.permissionCount ?? 0} permissions`
}

function reportsToLabel(user: AdminUserListItem & Record<string, unknown>) {
  return fallback(String(user.reportsToName || user.parentName || user.managerName || ''))
}

function jobTitleLabel(user: AdminUserListItem & Record<string, unknown>) {
  return fallback(String(user.jobTitle || user.position || user.reportsToName || user.managerName || ''))
}

function openActions(user: AdminUserListItem) {
  actionTarget.value = user
  actionDialogVisible.value = true
}

function closeActions() {
  actionDialogVisible.value = false
  actionTarget.value = null
}

function viewSelectedAccount() {
  if (!actionTarget.value) return
  const id = actionTarget.value.id
  closeActions()
  router.push(`/admin/accounts/${id}`)
}

function editSelectedAccount() {
  if (!actionTarget.value) return
  const id = actionTarget.value.id
  closeActions()
  router.push(`/admin/accounts/${id}/edit`)
}

function confirmDeactivate(user: AdminUserListItem) {
  deactivateTarget.value = { id: user.id, name: user.fullName }
  deactivateDialogVisible.value = true
}

function deactivateSelectedAccount() {
  if (!actionTarget.value) return
  const user = actionTarget.value
  closeActions()
  confirmDeactivate(user)
}

function activateSelectedAccount() {
  if (!actionTarget.value) return
  const id = actionTarget.value.id
  closeActions()
  void activate(id)
}

function confirmDelete(user: AdminUserListItem) {
  deleteTarget.value = user
  deleteDialogVisible.value = true
}

async function executeDeactivate() {
  if (!deactivateTarget.value) return
  updating.value = true
  error.value = ''
  try {
    await store.updateStatus(deactivateTarget.value.id, 'INACTIVE')
    deactivateDialogVisible.value = false
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    updating.value = false
  }
}

async function executeDelete() {
  if (!deleteTarget.value) return
  const name = accountDisplayName(deleteTarget.value)
  updating.value = true
  error.value = ''
  try {
    await store.deleteUser(deleteTarget.value.id)
    await load()
    deleteDialogVisible.value = false
    deleteTarget.value = null
    toast.add({
      severity: 'success',
      summary: 'Account Deleted',
      detail: `${name} has been removed from the active account list.`,
      life: 3500,
    })
  } catch (e) {
    error.value = store.errorMessage(e)
    toast.add({
      severity: 'error',
      summary: 'Delete Failed',
      detail: error.value,
      life: 5000,
    })
  } finally {
    updating.value = false
  }
}

async function activate(id: string) {
  updating.value = true
  error.value = ''
  try {
    await store.updateStatus(id, 'ACTIVE')
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    updating.value = false
  }
}

onMounted(() => { load() })
</script>

<template>
  <section class="accounts-page">
    <Toast position="top-right" />

    <AdminPageHeader title="Employee Management" subtitle="Manage employee accounts, roles, and access status.">
      <template #actions>
        <Button label="Create Employee" icon="pi pi-plus" size="small" class="create-button" @click="router.push('/admin/accounts/create')" />
      </template>
    </AdminPageHeader>

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <header class="accounts-toolbar">
      <label class="employee-search"><i class="pi pi-search" /><input v-model="store.params.search" placeholder="Search by employee ID, name, email, phone..." @keyup.enter="load" /></label>

      <div class="toolbar-controls">
        <Select v-model="selectedRole" :options="roleOptions" optionLabel="label" optionValue="value" placeholder="All departments" class="toolbar-inline-filter" />
        <Select v-model="selectedStatus" :options="statusOptions" optionLabel="label" optionValue="value" placeholder="All statuses" class="toolbar-inline-filter" />
        <Button label="More Filters" icon="pi pi-sliders-h" severity="secondary" outlined size="small" @click="showFilters = !showFilters" />

        <Button
          label="Trash"
          icon="pi pi-trash"
          severity="secondary"
          outlined
          size="small"
          class="trash-button"
          @click="trashVisible = true; router.replace({ query: { ...route.query, view: 'trash' } })"
        />

        <Button
          label="Create Employee"
          icon="pi pi-plus"
          size="small"
          class="create-button"
          @click="router.push('/admin/accounts/create')"
        />

      </div>
    </header>

    <div v-if="showFilters" class="employee-filter-panel">
      <div class="employee-filter-field"><label>Status</label><Select v-model="selectedStatus" :options="statusOptions" optionLabel="label" optionValue="value" placeholder="All statuses" /></div>
      <div class="employee-filter-field"><label>CRM Role</label><Select v-model="selectedRole" :options="roleOptions" optionLabel="label" optionValue="value" placeholder="All CRM roles" /></div>
      <Button label="Reset" icon="pi pi-refresh" severity="secondary" text size="small" @click="resetAll" />
    </div>

    <div v-if="selectedIds.length" class="bulk-action-bar">
      <strong>{{ selectedIds.length }} selected</strong>
      <div>
        <Button label="Soft Delete" icon="pi pi-trash" size="small" class="soft-delete-button" @click="bulkDeleteVisible = true" />
        <Button label="Cancel" severity="secondary" text size="small" @click="cancelSelection" />
      </div>
    </div>

    <AdminTableShell class="table-shell">
      <div class="accounts-pagination-top">
        <div class="accounts-page-links">
          <Button
            v-for="pageNumber in pageNumbers"
            :key="pageNumber"
            :label="String(pageNumber)"
            text
            size="small"
            :class="['accounts-page-link', { active: pageNumber === store.page }]"
            @click="selectPage(pageNumber)"
          />
          <span class="accounts-page-report">Page {{ store.page }} of {{ store.pages || 1 }} / {{ store.total }} records</span>
        </div>
        <div class="accounts-page-settings">
          <label>Page size <Select v-model="selectedPageSize" :options="[10, 20, 50]" /></label>
          <label>Go to <InputNumber v-model="goToPage" :min="1" :max="Math.max(store.pages, 1)" /></label>
          <Button label="Set" outlined size="small" @click="goToSelectedPage" />
        </div>
      </div>
      <div v-if="store.loading && !store.users.length" class="skeleton-area">
        <Skeleton v-for="n in 8" :key="n" class="skeleton-row" />
      </div>

      <DataTable
        v-else
        :value="store.users"
        :loading="store.loading"
        lazy
        :totalRecords="store.total"
        v-model:first="first"
        :rows="store.limit"
        dataKey="id"
        class="accounts-table"
        @page="onPage"
        @row-click="openActions($event.data)"
      >
        <template #empty>
          <div class="empty-state">
            <span class="empty-icon"><i class="pi pi-users" /></span>
            <strong>Employee tidak ditemukan</strong>
            <span>Coba kata kunci lain atau ubah filter status.</span>
          </div>
        </template>

        <Column header="" class="select-column">
          <template #header>
            <Checkbox :modelValue="allPageSelected" binary @update:modelValue="toggleSelectAll($event)" />
          </template>
          <template #body="{ data }">
            <Checkbox :modelValue="selectedIds.includes(data.id)" binary @update:modelValue="toggleSelected(data.id, $event)" @click.stop />
          </template>
        </Column>

        <Column header="Employee" class="employee-column">
          <template #body="{ data }">
            <div class="employee-cell">
              <div class="employee-avatar">
                {{ data.fullName?.slice(0, 1)?.toUpperCase() || '?' }}
              </div>
              <div class="employee-copy">
                <strong>{{ data.fullName }}</strong>
                <small>
                  {{ data.employeeId || 'No employee ID' }}
                  <template v-if="data.mustChangePassword">
                    · <span class="password-warning">Password change required</span>
                  </template>
                </small>
              </div>
            </div>
          </template>
        </Column>

        <Column header="Contact" class="contact-column">
          <template #body="{ data }"><div class="role-cell"><span>{{ data.email }}</span><span>{{ data.phone || '-' }}</span></div></template>
        </Column>

        <Column header="Department" class="role-column">
          <template #body="{ data }">
            <div class="role-cell">
              <strong>{{ data.role.replaceAll('_', ' ') }}</strong>
              <span>
                {{ data.accountType ? data.accountType.replaceAll('_', ' ') : 'CRM department' }}
              </span>
            </div>
          </template>
        </Column>

        <Column header="Job Title" class="reports-column">
          <template #body="{ data }">
            <span class="single-line" :title="jobTitleLabel(data)">
              {{ jobTitleLabel(data) }}
            </span>
          </template>
        </Column>

        <Column header="Role" class="role-column">
          <template #body="{ data }"><div class="role-cell"><strong>{{ organizationalRoleLabel(data) }}</strong><span>Level {{ data.organizationalRole?.level ?? '-' }} · {{ organizationalRoleMeta(data) }}</span></div></template>
        </Column>

        <Column header="Status" class="status-column">
          <template #body="{ data }">
            <Tag
              :value="data.status === 'ACTIVE' ? 'Active' : 'Inactive'"
              :severity="data.status === 'ACTIVE' ? 'success' : 'secondary'"
              rounded
              class="status-tag"
            />
          </template>
        </Column>

        <Column header="Location" class="updated-column">
          <template #body="{ data }">
            <span class="single-line">{{ fallback(String(data.location || data.city || data.province || '')) }}</span>
          </template>
        </Column>

      </DataTable>
    </AdminTableShell>

    <Dialog
  v-model:visible="actionDialogVisible"
  modal
  :draggable="false"
  :closable="false"
  :dismissableMask="true"
  class="account-action-dialog"
  :style="{ width: 'min(520px, calc(100vw - 2rem))' }"
  @hide="closeActions"
>
  <div v-if="actionTarget" class="action-dialog-content">
    <div class="action-dialog-header">
      <div>
        <span class="dialog-eyebrow">Actions</span>
        <h2>{{ actionTarget.fullName }}</h2>
        <p>{{ actionTarget.email }}</p>
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
      <button class="action-card" type="button" @click="viewSelectedAccount">
        <span class="action-card-icon view-icon"><i class="pi pi-eye" /></span>
        <span>
          <strong>View Detail</strong>
          <small>View complete account profile and access information.</small>
        </span>
      </button>

      <button class="action-card" type="button" @click="editSelectedAccount">
        <span class="action-card-icon edit-icon"><i class="pi pi-pencil" /></span>
        <span>
          <strong>Edit</strong>
          <small>Update identity, role, and account information.</small>
        </span>
      </button>
    </div>

    <div class="account-summary">
      <div>
        <span>Role</span>
        <strong>{{ organizationalRoleLabel(actionTarget) }}</strong>
      </div>
      <div>
        <span>Status</span>
        <Tag
          :value="actionTarget.status === 'ACTIVE' ? 'Active' : 'Inactive'"
          :severity="actionTarget.status === 'ACTIVE' ? 'success' : 'secondary'"
          rounded
        />
      </div>
    </div>

    <div class="danger-zone">
      <div class="danger-zone-copy">
        <i class="pi pi-exclamation-triangle" />
        <div>
          <strong>
            {{ actionTarget.status === 'ACTIVE' ? 'Deactivate Account' : 'Activate Account' }}
          </strong>
          <span>
            {{
              actionTarget.status === 'ACTIVE'
                ? 'Disable sign-in access without deleting the account.'
                : 'Restore sign-in access for this account.'
            }}
          </span>
        </div>
      </div>

      <Button
        v-if="actionTarget.status === 'ACTIVE'"
        label="Deactivate"
        icon="pi pi-user-minus"
        severity="danger"
        outlined
        size="small"
        :disabled="isSelf(actionTarget.id) || isProtectedSuperAdmin(actionTarget) || updating"
        @click="deactivateSelectedAccount"
      />
      <Button
        v-else
        label="Activate"
        icon="pi pi-user-plus"
        severity="success"
        outlined
        size="small"
        :disabled="updating"
        @click="activateSelectedAccount"
      />
    </div>

    <div class="delete-zone">
      <div class="danger-zone-copy">
        <i class="pi pi-trash" />
        <div>
          <strong>Delete Account</strong>
          <span>Permanently remove this account when it is no longer referenced.</span>
        </div>
      </div>

      <Button
        label="Delete"
        icon="pi pi-trash"
        severity="danger"
        size="small"
        :disabled="!actionTarget || isSelf(actionTarget?.id) || isProtectedSuperAdmin(actionTarget) || updating"
        @click="confirmDelete(actionTarget)"
      />
    </div>
  </div>
</Dialog>

    <Dialog
      v-model:visible="deactivateDialogVisible"
      header="Deactivate Account"
      modal
      :draggable="false"
      :style="{ width: 'min(420px, calc(100vw - 2rem))' }"
    >
      <p>
        Are you sure you want to deactivate
        <strong>{{ deactivateTarget?.name }}</strong>?
        The user will no longer be able to sign in.
      </p>
      <template #footer>
        <Button
          label="Cancel"
          severity="secondary"
          text
          :disabled="updating"
          @click="deactivateDialogVisible = false"
        />
        <Button
          label="Deactivate"
          severity="danger"
          icon="pi pi-user-minus"
          :loading="updating"
          @click="executeDeactivate"
        />
      </template>
    </Dialog>

    <Dialog
      v-model:visible="deleteDialogVisible"
      header="Delete Account"
      modal
      :draggable="false"
      :style="{ width: 'min(420px, calc(100vw - 2rem))' }"
    >
      <p>
        Delete <strong>{{ accountDisplayName(deleteTarget) }}</strong>?
        This will remove the account from active account views while preserving CRM records.
      </p>
      <template #footer>
        <Button
          label="Cancel"
          severity="secondary"
          text
          :disabled="updating"
          @click="deleteDialogVisible = false"
        />
        <Button
          label="Delete"
          severity="danger"
          icon="pi pi-trash"
          :loading="updating"
          @click="executeDelete"
        />
      </template>
    </Dialog>
    <Dialog v-model:visible="bulkDeleteVisible" modal header="Soft Delete Accounts" :style="{ width: 'min(420px, calc(100vw - 2rem))' }">
      <p class="bulk-delete-copy">Move {{ selectedIds.length }} selected account{{ selectedIds.length === 1 ? '' : 's' }} to trash?</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text :disabled="updating" @click="bulkDeleteVisible = false" />
        <Button label="Soft Delete" icon="pi pi-trash" class="soft-delete-button" :loading="updating" @click="executeBulkDelete" />
      </template>
    </Dialog>
    <Dialog v-model:visible="trashVisible" modal header="Trash Accounts" :style="{ width: 'min(760px, calc(100vw - 2rem))' }" :draggable="false" @hide="router.replace({ query: { ...route.query, view: undefined } })">
      <div class="trash-content">
        <p class="trash-intro">Akun yang dihapus tersimpan di sini dan bisa dikembalikan kapan saja.</p>
        <div v-if="store.loading" class="trash-loading">Loading trash...</div>
        <div v-else-if="!trashedUsers.length" class="trash-empty"><i class="pi pi-trash" /><strong>Trash masih kosong</strong><span>Tidak ada akun yang dihapus.</span></div>
        <div v-else class="trash-list">
          <div v-for="user in trashedUsers" :key="user.id" class="trash-row"><div><strong>{{ user.fullName || user.email }}</strong><span>{{ user.employeeId || '-' }} · {{ user.email }}</span></div><Button label="Restore" icon="pi pi-undo" size="small" :loading="restoringId === user.id" @click="restoreAccount(user)" /></div>
        </div>
      </div>
    </Dialog>
  </section>
</template>

<style scoped>
.accounts-page {
  width: 100%;
  min-width: 0;
  min-height: 100vh;
  padding: 0;
  overflow-x: hidden;
  background: #ffffff;
}

.trash-content{display:grid;gap:.8rem}.trash-intro{margin:0;color:#64748b;font-size:.8rem}.trash-list{display:grid;gap:.5rem;max-height:55vh;overflow:auto}.trash-row{display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:.75rem;border:1px solid #e5eaf0;border-radius:10px}.trash-row div{display:grid;gap:.2rem}.trash-row strong{color:#172033;font-size:.82rem}.trash-row span{color:#94a3b8;font-size:.7rem}.trash-empty{display:grid;justify-items:center;gap:.35rem;padding:2rem;color:#94a3b8}.trash-empty i{font-size:1.7rem;color:#ef4444}.trash-empty strong{color:#172033;font-size:.9rem}.trash-empty span,.trash-loading{font-size:.75rem}

.page-message {
  margin: 0.75rem 1rem 0;
}

.accounts-toolbar {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  min-width: 0;
  padding: 0.6rem 0.8rem;
  border-bottom: 1px solid #e5eaf0;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(10px);
}

.toolbar-title {
  display: grid;
  flex: 0 0 auto;
  gap: 0.08rem;
  min-width: 150px;
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
  line-height: 1.15;
}

.toolbar-title p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.7rem;
}

.toolbar-controls {
  display: grid;
  grid-template-columns: 165px auto auto;
  align-items: center;
  gap: 0.55rem;
  width: auto;
  min-width: 0;
}

.search-field {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.55rem;
  height: 38px;
  padding: 0 0.75rem;
  border: 1px solid #dfe5ec;
  border-radius: 9px;
  background: #ffffff;
}

.search-field:focus-within {
  border-color: #e63946;
  box-shadow: 0 0 0 3px rgba(230, 57, 70, 0.08);
}

.search-field i {
  flex: 0 0 auto;
  color: #94a3b8;
  font-size: 0.82rem;
}

.search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #0f172a;
  font-size: 0.78rem;
}

.search-field input::placeholder {
  color: #a7b1c0;
}

.status-filter {
  min-width: 0;
}

.status-filter :deep(.p-select) {
  width: 100%;
}

.status-filter :deep(.p-select-label) {
  padding-top: 0.54rem;
  padding-bottom: 0.54rem;
  font-size: 0.76rem;
}

.reset-button,
.create-button {
  white-space: nowrap;
}

.table-shell {
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border-bottom: 1px solid #e5eaf0;
  background: #ffffff;
}

.accounts-table {
  width: 100%;
  min-width: 0;
}

.accounts-table :deep(.p-datatable-table-container) {
  width: 100%;
  min-width: 0;
  overflow: hidden;
}

.accounts-table :deep(.p-datatable-table) {
  width: 100%;
  min-width: 0;
  table-layout: fixed;
}

.accounts-table :deep(.p-datatable-thead > tr > th) {
  overflow: hidden;
  padding: 0.58rem 0.7rem;
  border-color: #e5eaf0;
  background: #f8fafc;
  color: #475569;
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.045em;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.accounts-table :deep(.p-datatable-tbody > tr > td) {
  overflow: hidden;
  padding: 0.55rem 0.7rem;
  border-color: #edf1f6;
  background: #ffffff;
  color: #1e293b;
  font-size: 0.76rem;
  vertical-align: middle;
}

.accounts-table :deep(.p-datatable-tbody > tr) {
  cursor: pointer;
}

.accounts-table :deep(.p-datatable-tbody > tr:hover > td) {
  background: #fffbfb;
}

.accounts-table :deep(.employee-column) {
  width: 31%;
}

.accounts-table :deep(.role-column) {
  width: 24%;
}

.accounts-table :deep(.reports-column) {
  width: 15%;
}

.accounts-table :deep(.status-column) {
  width: 10%;
}

.accounts-table :deep(.updated-column) {
  width: 11%;
}

.accounts-table :deep(.actions-column) {
  width: 9%;
  text-align: right;
}

.employee-cell {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.65rem;
}

.employee-avatar {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  place-content: center;
  border-radius: 9px;
  background: #fff0f1;
  color: #e63946;
  font-size: 0.75rem;
  font-weight: 800;
}

.employee-copy,
.role-cell {
  display: grid;
  min-width: 0;
  gap: 0.06rem;
}

.employee-copy strong,
.role-cell strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.78rem;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.employee-copy span,
.role-cell span {
  overflow: hidden;
  color: #64748b;
  font-size: 0.69rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.employee-copy small {
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.64rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.password-warning {
  color: #b45309 !important;
  font-weight: 700;
}

.single-line {
  display: block;
  overflow: hidden;
  color: #475569;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  font-size: 0.66rem;
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  white-space: nowrap;
}

.more-action {
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  border: 1px solid #dfe5ec !important;
  background: #ffffff !important;
  color: #64748b !important;
}

.more-action:hover {
  border-color: #cbd5e1 !important;
  background: #f8fafc !important;
  color: #0f172a !important;
}

.accounts-table :deep(.p-paginator) {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem;
  min-height: 48px;
  padding: 0.45rem 0.7rem;
  border-top: 1px solid #e5eaf0;
  background: #ffffff;
}

.accounts-table :deep(.p-paginator-current) {
  margin-right: auto;
  color: #64748b;
  font-size: 0.7rem;
}

.skeleton-area {
  padding: 0.6rem 1rem;
}

.skeleton-row {
  height: 3.2rem;
  margin: 0.45rem 0;
  border-radius: 8px;
}

.empty-state {
  display: flex;
  min-height: 280px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
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


.account-action-dialog :deep(.p-dialog) {
  overflow: hidden;
  border: 1px solid #dfe5ec;
  border-radius: 18px;
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.2);
}

.account-action-dialog :deep(.p-dialog-header) {
  display: none;
}

.account-action-dialog :deep(.p-dialog-content) {
  padding: 0;
  border-radius: 18px;
  background: #ffffff;
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
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.action-dialog-header h2 {
  margin: 0.15rem 0 0;
  color: #0f172a;
  font-size: 1.15rem;
  line-height: 1.2;
}

.action-dialog-header p {
  margin: 0.2rem 0 0;
  color: #64748b;
  font-size: 0.78rem;
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.7rem;
}

.action-card {
  display: flex;
  align-items: flex-start;
  gap: 0.7rem;
  min-height: 92px;
  padding: 0.9rem;
  border: 1px solid #dfe5ec;
  border-radius: 12px;
  background: #ffffff;
  text-align: left;
  cursor: pointer;
  transition: border-color 150ms ease, background 150ms ease, transform 150ms ease;
}

.action-card:hover {
  border-color: #f4b3ba;
  background: #fffbfb;
  transform: translateY(-1px);
}

.action-card-icon {
  display: grid;
  width: 32px;
  height: 32px;
  flex: 0 0 auto;
  place-content: center;
  border-radius: 8px;
}

.view-icon {
  background: #fff0f1;
  color: #e63946;
}

.edit-icon {
  background: #fff7ed;
  color: #ea580c;
}

.action-card > span:last-child {
  display: grid;
  gap: 0.2rem;
}

.action-card strong {
  color: #0f172a;
  font-size: 0.85rem;
}

.action-card small {
  color: #64748b;
  font-size: 0.72rem;
  line-height: 1.45;
}

.account-summary {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.75rem;
  margin-top: 0.8rem;
  padding: 0.75rem 0.85rem;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  background: #f8fafc;
}

.account-summary > div {
  display: grid;
  gap: 0.15rem;
}

.account-summary span {
  color: #94a3b8;
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
}

.account-summary strong {
  color: #0f172a;
  font-size: 0.8rem;
}

.danger-zone,
.delete-zone {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 0.8rem;
  padding: 0.8rem;
  border-radius: 11px;
}

.danger-zone {
  border: 1px solid #fed7aa;
  background: #fffaf2;
}

.delete-zone {
  border: 1px solid #fecaca;
  background: #fff7f7;
}

.danger-zone-copy {
  display: flex;
  align-items: flex-start;
  min-width: 0;
  gap: 0.65rem;
}

.danger-zone-copy > i {
  margin-top: 0.1rem;
  color: #dc2626;
}

.danger-zone-copy > div {
  display: grid;
  min-width: 0;
  gap: 0.15rem;
}

.danger-zone-copy strong {
  color: #991b1b;
  font-size: 0.8rem;
}

.danger-zone-copy span {
  color: #7f1d1d;
  font-size: 0.7rem;
  line-height: 1.4;
}

@media (max-width: 1100px) {
  .accounts-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-controls {
    width: 100%;
  }

  .accounts-table :deep(.updated-column) {
    display: none;
  }

  .accounts-table :deep(.employee-column) {
    width: 34%;
  }

  .accounts-table :deep(.role-column) {
    width: 25%;
  }

  .accounts-table :deep(.reports-column) {
    width: 18%;
  }

  .accounts-table :deep(.status-column) {
    width: 12%;
  }

  .accounts-table :deep(.actions-column) {
    width: 11%;
  }
}

@media (max-width: 820px) {
  .toolbar-controls {
    grid-template-columns: minmax(0, 1fr) 150px auto;
  }

  .create-button {
    grid-column: 1 / -1;
    justify-self: end;
  }

  .accounts-table :deep(.reports-column) {
    display: none;
  }

  .accounts-table :deep(.employee-column) {
    width: 42%;
  }

  .accounts-table :deep(.role-column) {
    width: 31%;
  }

  .accounts-table :deep(.status-column) {
    width: 14%;
  }

  .accounts-table :deep(.actions-column) {
    width: 13%;
  }
}

@media (max-width: 620px) {
  .action-grid {
    grid-template-columns: 1fr;
  }

  .danger-zone,
  .delete-zone {
    align-items: stretch;
    flex-direction: column;
  }

  .danger-zone :deep(.p-button),
  .delete-zone :deep(.p-button) {
    width: 100%;
  }

  .accounts-toolbar {
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

  .accounts-table :deep(.p-datatable-thead) {
    display: none;
  }

  .accounts-table :deep(.p-datatable-table),
  .accounts-table :deep(.p-datatable-tbody),
  .accounts-table :deep(.p-datatable-tbody > tr),
  .accounts-table :deep(.p-datatable-tbody > tr > td) {
    display: block;
    width: 100%;
  }

  .accounts-table :deep(.p-datatable-tbody > tr) {
    position: relative;
    padding: 0.75rem;
    border-bottom: 1px solid #e5eaf0;
  }

  .accounts-table :deep(.p-datatable-tbody > tr > td) {
    padding: 0.2rem 0;
    border: 0;
  }

  .accounts-table :deep(.status-column) {
    position: absolute;
    top: 0.75rem;
    right: 0.75rem;
    display: block;
    width: auto;
  }

  .accounts-table :deep(.employee-column),
  .accounts-table :deep(.role-column),
  .accounts-table :deep(.actions-column) {
    display: block;
    width: 100%;
  }

  .role-cell {
    margin-top: 0.35rem;
    padding-left: 2.65rem;
  }

  .row-actions {
    justify-content: flex-start;
    padding-left: 2.55rem;
    padding-top: 0.35rem;
  }

  .accounts-table :deep(.p-paginator-current) {
    width: 100%;
    margin: 0 0 0.3rem;
    text-align: center;
  }
}
.accounts-pagination-top{display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:.55rem .7rem;border:1px solid #e5eaf0;border-radius:10px 10px 0 0;background:#fff}.accounts-page-links,.accounts-page-settings{display:flex;align-items:center;gap:.35rem}.accounts-page-link{min-width:28px;height:28px;padding:0;color:#475569}.accounts-page-link.active{border-radius:50%;background:#fff0f1;color:#d62839;font-weight:800}.accounts-page-report{margin-left:.45rem;color:#64748b;font-size:.68rem;white-space:nowrap}.accounts-page-settings{gap:.55rem}.accounts-page-settings label{display:flex;align-items:center;gap:.35rem;color:#64748b;font-size:.62rem;white-space:nowrap}.accounts-page-settings :deep(.p-select),.accounts-page-settings :deep(.p-inputnumber-input){height:32px;border:1px solid #dbe3ee;border-radius:7px;font-size:.68rem}.accounts-page-settings :deep(.p-select){width:72px}.accounts-page-settings :deep(.p-inputnumber){width:58px}.accounts-page-settings :deep(.p-inputnumber-input){width:58px;padding:.35rem}.accounts-page-settings :deep(.p-button){height:32px;padding:0 .7rem;font-size:.68rem}.table-shell>.accounts-table{border-top:0;border-radius:0 0 10px 10px}@media(max-width:700px){.accounts-pagination-top{align-items:flex-start;flex-direction:column}.accounts-page-settings{width:100%;justify-content:flex-end}.accounts-page-report{margin-left:.2rem}}
.select-column{width:46px;text-align:center}.bulk-action-bar{display:flex;align-items:center;justify-content:space-between;padding:.65rem .8rem;border:1px solid #fecdd3;background:#fff1f2;color:#9f1239}.bulk-action-bar strong{font-size:.72rem}.bulk-action-bar>div{display:flex;align-items:center;gap:.5rem}.soft-delete-button{border:0!important;background:#e11d2e!important;color:#fff!important}.bulk-delete-copy{margin:0;color:#475569;font-size:.85rem}.accounts-toolbar{flex-wrap:nowrap;min-height:56px;padding:.5rem .7rem}.toolbar-controls{display:flex;flex-wrap:nowrap;align-items:center;justify-content:flex-end;gap:.4rem}.toolbar-controls :deep(.p-select),.toolbar-controls :deep(.p-button){height:34px}.toolbar-controls :deep(.p-select){min-width:130px}.accounts-pagination-top{flex-wrap:nowrap;min-height:48px;white-space:nowrap}.accounts-page-links,.accounts-page-settings{flex-wrap:nowrap;white-space:nowrap}.accounts-page-settings label{flex-shrink:0}.accounts-page-report{overflow:hidden;text-overflow:ellipsis}.table-shell{overflow:hidden;border:1px solid #e5eaf0;border-radius:10px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.03)}.accounts-pagination-top{border:0;border-bottom:1px solid #e5eaf0;border-radius:0;background:#fff}.accounts-table{border:0!important;border-radius:0!important}.accounts-table :deep(.p-datatable-thead > tr > th){padding:.6rem .7rem;background:#f4f6f9;color:#475569;font-size:.6rem;letter-spacing:.05em;text-transform:uppercase;border-right:1px solid #e5eaf0}.accounts-table :deep(.p-datatable-tbody > tr > td){padding:.65rem .7rem;border-color:#e5eaf0;border-right:1px solid #e5eaf0}.accounts-table :deep(.p-datatable-tbody > tr:hover > td){background:#fffafa}@media(max-width:700px){.accounts-toolbar{align-items:flex-start;flex-direction:column}.toolbar-controls{width:100%;justify-content:flex-start;flex-wrap:wrap}}
@media (max-width: 620px) {
  .accounts-page { padding:.5rem; gap:.55rem; background:#f8fafc; }
  .accounts-toolbar { padding:.85rem; border-radius:14px; gap:.7rem; }
  .toolbar-title .eyebrow { color:#e63946; font-size:.55rem; }
  .toolbar-title h1 { font-size:1.15rem; }
  .toolbar-title p { font-size:.64rem; }
  .toolbar-controls { grid-template-columns:1fr 1fr; gap:.4rem; }
  .toolbar-controls :deep(.p-select), .toolbar-controls :deep(.p-button) { width:100%; min-height:38px; font-size:.67rem; }
  .toolbar-controls .status-filter { grid-column:1/-1; }
  .toolbar-controls .create-button { grid-column:1/-1; background:#e63946; border-color:#e63946; }
  .table-shell { border-radius:12px; }
  .accounts-pagination-top { padding:.6rem; gap:.45rem; }
  .accounts-page-links { width:100%; justify-content:center; flex-wrap:wrap; }
  .accounts-page-report { width:100%; margin:0; text-align:center; font-size:.58rem; }
  .accounts-page-settings { display:none; }
  .accounts-table :deep(.p-datatable-tbody > tr) { margin:0; padding:.8rem; background:#fff; border-bottom:1px solid #f4e6e8; }
  .accounts-table :deep(.p-datatable-tbody > tr > td) { padding:.18rem 0; }
  .accounts-table :deep(.select-column) { position:absolute; top:.7rem; left:.55rem; width:auto; }
  .accounts-table :deep(.employee-column) { padding-left:2rem!important; padding-right:4.5rem!important; }
  .employee-cell { min-height:42px; gap:.55rem; }
  .employee-avatar { width:34px; height:34px; font-size:.78rem; }
  .employee-copy strong { font-size:.74rem; }
  .employee-copy span, .employee-copy small { font-size:.58rem; }
  .accounts-table :deep(.role-column), .accounts-table :deep(.actions-column) { font-size:.62rem; }
  .accounts-table :deep(.status-column) { top:.7rem; right:.7rem; }
  .accounts-table :deep(.p-tag) { font-size:.55rem; padding:.18rem .35rem; }
}
.accounts-page { font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif; }
.accounts-page .accounts-table :deep(.p-datatable-thead > tr > th) { height:40px; padding:0 .5rem; background:#f8fafc; color:#0f172a; font-size:.875rem; font-weight:500; line-height:1.25; letter-spacing:0; text-transform:none; }
.accounts-page .accounts-table :deep(.p-datatable-tbody > tr > td) { height:48px; padding:.5rem; color:#0f172a; border-bottom:1px solid rgba(15,23,42,.1); font-size:.875rem; line-height:1.4; }
.accounts-page button:not(.collapse-btn):not(.hamburger-btn), .accounts-page :deep(.p-button) { min-height:36px; height:36px; padding:0 1rem; border-radius:6px; font-size:.875rem; font-weight:500; }
.accounts-page input, .accounts-page select, .accounts-page :deep(.p-inputtext), .accounts-page :deep(.p-select) { height:36px; min-height:36px; border-radius:6px; font-size:.875rem; }
.accounts-page :deep(.p-tag) { border-radius:6px; padding:2px 8px; font-size:.75rem; font-weight:500; }
.accounts-page { min-width:0; padding:0; gap:0; background:#fff; font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif; }
.accounts-toolbar { display:flex; min-height:70px; height:70px; align-items:center; gap:10px; padding:14px 20px; border-bottom:1px solid #e2e8f0; background:#fff; }
.employee-search { display:flex; flex:0 1 360px; height:40px; align-items:center; gap:8px; padding:0 12px; border:1px solid #e2e8f0; border-radius:10px; color:#94a3b8; }.employee-search i{font-size:13px}.employee-search input{width:100%;height:32px;border:0;outline:0;background:transparent;color:#0f172a;font-size:13px}.employee-search input::placeholder{color:#94a3b8}
.toolbar-title { display:none; }.toolbar-controls { display:flex; align-items:center; gap:10px; margin-left:auto; }.toolbar-controls :deep(.p-button){height:40px;min-height:40px;padding:0 14px;border-radius:10px;font-size:12px;font-weight:600}.toolbar-controls :deep(.create-button){background:#dc2626;border-color:#dc2626}.toolbar-controls :deep(.trash-button){min-width:72px}.employee-filter-panel{display:flex;align-items:flex-end;gap:14px;padding:14px 20px 16px;border-bottom:1px solid #e2e8f0;background:#fff;box-shadow:0 3px 12px rgba(15,23,42,.04)}.employee-filter-field{display:grid;gap:5px;width:210px}.employee-filter-field label{color:#64748b;font-size:10px;font-weight:600;letter-spacing:.04em;text-transform:uppercase}.employee-filter-field :deep(.p-select){height:40px;border:1px solid #e2e8f0;border-radius:10px;font-size:12px}.employee-filter-panel>:deep(.p-button){height:40px;min-height:40px;border-radius:10px;font-size:12px}
.accounts-page .table-shell{border:0;border-radius:0;box-shadow:none}.accounts-page .accounts-pagination-top{height:42px;min-height:42px;padding:0 20px;border-bottom:1px solid #e2e8f0}.accounts-page .accounts-page-link{width:28px;min-width:28px;height:28px;border-radius:50%;font-size:10px}.accounts-page .accounts-page-report{font-size:10px}.accounts-page .accounts-page-settings :deep(.p-select),.accounts-page .accounts-page-settings :deep(.p-inputnumber-input),.accounts-page .accounts-page-settings :deep(.p-button){height:34px;border-radius:8px;font-size:11px}.accounts-page .accounts-table :deep(.p-datatable-thead > tr > th){box-sizing:border-box;height:49px;padding:0 12px;border-right:1px solid #e2e2e2;border-bottom:1px solid #e0e0e0;background:#f4f4f4;color:#000;font-size:10px;font-weight:600;letter-spacing:.07em;text-transform:uppercase}.accounts-page .accounts-table :deep(.p-datatable-tbody > tr > td){box-sizing:border-box;height:84px;padding:10px 12px;border-right:1px solid #e2e2e0;border-bottom:1px solid #e0e0e0;color:#0f172a;font-size:12px}.accounts-page .accounts-table :deep(.p-datatable-tbody > tr:hover > td){background:#fff}.accounts-page .employee-avatar{display:none}.accounts-page .employee-cell,.accounts-page .employee-copy,.accounts-page .role-cell{display:grid;gap:3px}.accounts-page .employee-copy strong,.accounts-page .role-cell strong{font-size:12px;font-weight:600}.accounts-page .employee-copy span,.accounts-page .role-cell span,.accounts-page .single-line{color:#64748b;font-size:11px;line-height:15px}.accounts-page .employee-copy small{color:#64748b;font-size:10px}.accounts-page :deep(.status-tag){border-radius:6px;padding:4px 9px;font-size:10px;font-weight:500}.accounts-page .select-column{width:54px}.accounts-page .skeleton-area{padding:16px}
@media(max-width:900px){.accounts-toolbar{height:auto;min-height:70px;flex-wrap:wrap}.employee-search{flex-basis:100%;max-width:none}.toolbar-controls{width:100%;margin-left:0;flex-wrap:wrap}.toolbar-controls :deep(.p-button){flex:1}.employee-filter-panel{flex-wrap:wrap}.accounts-page .accounts-table{min-width:900px}.accounts-page .table-shell{overflow-x:auto}}
.accounts-page .accounts-page-settings{gap:8px}.accounts-page .accounts-page-settings label{gap:5px;font-size:10px}.accounts-page .accounts-page-settings :deep(.p-select){width:76px;height:36px;border-radius:9px;font-size:11px}.accounts-page .accounts-page-settings :deep(.p-select-label){overflow:hidden;padding:0 10px;white-space:nowrap}.accounts-page .accounts-page-settings :deep(.p-inputnumber){width:64px}.accounts-page .accounts-page-settings :deep(.p-inputnumber-input){width:64px;height:36px;padding:0 10px;border-radius:9px;font-size:11px}.accounts-page .accounts-page-settings :deep(.p-button){height:36px;min-width:48px;border-radius:9px;font-size:11px}
.accounts-page .toolbar-status-filter,.accounts-page .toolbar-role-filter{width:180px;height:40px;border:1px solid #e2e8f0;border-radius:10px;font-size:12px}.accounts-page .toolbar-status-filter :deep(.p-select-label),.accounts-page .toolbar-role-filter :deep(.p-select-label){padding:0 12px}.accounts-page .filter-panel{box-shadow:0 3px 12px rgba(15,23,42,.04)}
.accounts-page .accounts-table :deep(.p-datatable-thead > tr > th){font-weight:600;-webkit-font-smoothing:antialiased}.accounts-page .accounts-table :deep(.p-datatable-tbody > tr > td){font-weight:400;-webkit-font-smoothing:antialiased}.accounts-page .employee-copy strong,.accounts-page .role-cell strong{font-weight:600}.accounts-page .employee-copy span,.accounts-page .employee-copy small,.accounts-page .role-cell span,.accounts-page .single-line{font-weight:400}.accounts-page :deep(.status-tag){font-weight:500}
.accounts-page .employee-copy strong,.accounts-page .role-cell strong{font-size:12px;font-weight:700;line-height:16px}.accounts-page .employee-copy span,.accounts-page .employee-copy small,.accounts-page .role-cell span,.accounts-page .single-line{font-size:11px;font-weight:400;line-height:15px}.accounts-page .accounts-table :deep(.p-datatable-thead > tr > th){font-size:10px;font-weight:600}.accounts-page :deep(.status-tag){font-size:10px;font-weight:500}
.accounts-page>.admin-page-header{display:none}.accounts-page .accounts-toolbar{gap:10px;padding:14px 24px}.accounts-page .employee-search{flex:0 1 360px;max-width:360px}.accounts-page .toolbar-controls{flex:1;gap:10px}.accounts-page .toolbar-inline-filter{width:190px}.accounts-page .toolbar-inline-filter :deep(.p-select-label){padding:0 12px;color:#64748b}.accounts-page .toolbar-controls .trash-button{margin-left:auto}.accounts-page .toolbar-controls .create-button{min-width:160px;background:#dc2626;border-color:#dc2626}@media(max-width:900px){.accounts-page .toolbar-inline-filter{flex:1;min-width:150px}}@media(max-width:640px){.accounts-page .accounts-toolbar{padding:10px 12px}.accounts-page .employee-search{max-width:none}.accounts-page .toolbar-inline-filter{width:100%;flex-basis:100%}.accounts-page .toolbar-controls .trash-button{margin-left:0}}
.accounts-page .accounts-toolbar{padding-left:24px;padding-right:24px;overflow:visible}.accounts-page .employee-search{width:auto;min-width:220px;flex:1 1 360px}.accounts-page .toolbar-controls{min-width:0;flex:1 1 auto;flex-wrap:nowrap}.accounts-page .toolbar-inline-filter{flex:1 1 150px;min-width:130px}.accounts-page .toolbar-controls :deep(.p-button){white-space:nowrap;flex-shrink:0}.accounts-page .toolbar-controls .create-button{width:160px;min-width:160px}
@media(max-width:1100px){.accounts-page .employee-search{width:300px;min-width:260px;flex-basis:300px}.accounts-page .toolbar-inline-filter{flex-basis:160px}}
@media(max-width:800px){.accounts-page .accounts-toolbar{flex-wrap:wrap}.accounts-page .employee-search{width:100%;min-width:0;flex-basis:100%;max-width:none}.accounts-page .toolbar-controls{width:100%;flex-wrap:wrap}.accounts-page .toolbar-inline-filter{flex:1 1 150px}.accounts-page .toolbar-controls .trash-button{margin-left:auto}}
</style>
