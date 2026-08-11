<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Toast from 'primevue/toast'
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
const toast = useToast()
const error = ref('')
const updating = ref(false)
const first = ref(0)
const deactivateDialogVisible = ref(false)
const deactivateTarget = ref<{ id: string; name: string } | null>(null)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<AdminUserListItem | null>(null)
const actionDialogVisible = ref(false)
const actionTarget = ref<AdminUserListItem | null>(null)

const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Inactive', value: 'INACTIVE' },
]

const selectedStatus = computed({
  get: () => store.params.status,
  set: (val) => { store.setParam('status', val); store.setParam('page', 1); first.value = 0; load() },
})

let searchTimeout: ReturnType<typeof setTimeout> | null = null
function onKeywordSearch(value: string) {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    store.setParam('search', value.trim())
    store.setParam('page', 1)
    first.value = 0
    load()
  }, 350)
}

function load() {
  error.value = ''
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

function updatedLabel(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
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

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <header class="accounts-toolbar">
      <div class="toolbar-title">
        <span class="eyebrow">Account Management</span>
        <h1>Account List</h1>
        <p>{{ store.total }} account{{ store.total === 1 ? '' : 's' }}</p>
      </div>

      <div class="toolbar-controls">
        <div class="search-field">
          <i class="pi pi-search" />
          <input
            type="text"
            placeholder="Search name, email, or employee ID"
            :value="store.params.search"
            @input="onKeywordSearch(($event.target as HTMLInputElement).value)"
          />
        </div>

        <Select
          v-model="selectedStatus"
          :options="statusOptions"
          optionLabel="label"
          optionValue="value"
          placeholder="All Status"
          class="status-filter"
        />

        <Button
          label="Reset"
          icon="pi pi-refresh"
          severity="secondary"
          outlined
          size="small"
          class="reset-button"
          @click="resetAll"
        />

        <Button
          label="Create Account"
          icon="pi pi-plus"
          size="small"
          class="create-button"
          @click="router.push('/admin/accounts/create')"
        />
      </div>
    </header>

    <div class="table-shell">
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
        :rowsPerPageOptions="[10, 20, 50]"
        paginator
        paginatorTemplate="CurrentPageReport FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink RowsPerPageDropdown"
        currentPageReportTemplate="Showing {first} to {last} of {totalRecords} accounts"
        dataKey="id"
        class="accounts-table"
        @page="onPage"
        @row-click="openActions($event.data)"
      >
        <template #empty>
          <div class="empty-state">
            <span class="empty-icon"><i class="pi pi-users" /></span>
            <strong>No accounts found</strong>
            <span>Try another keyword or change the status filter.</span>
          </div>
        </template>

        <Column header="Employee" class="employee-column">
          <template #body="{ data }">
            <div class="employee-cell">
              <div class="employee-avatar">
                {{ data.fullName?.slice(0, 1)?.toUpperCase() || '?' }}
              </div>
              <div class="employee-copy">
                <strong>{{ data.fullName }}</strong>
                <span>{{ data.email }}</span>
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

        <Column header="Role" class="role-column">
          <template #body="{ data }">
            <div class="role-cell">
              <strong>{{ organizationalRoleLabel(data) }}</strong>
              <span>
                Level {{ data.organizationalRole?.level ?? '-' }}
                · {{ organizationalRoleMeta(data) }}
              </span>
            </div>
          </template>
        </Column>

        <Column header="Reports To" class="reports-column">
          <template #body="{ data }">
            <span class="single-line" :title="reportsToLabel(data)">
              {{ reportsToLabel(data) }}
            </span>
          </template>
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

        <Column header="Updated" class="updated-column">
          <template #body="{ data }">
            <span class="single-line">{{ updatedLabel(data.updatedAt) }}</span>
          </template>
        </Column>

        <Column header="Actions" class="actions-column">
          <template #body="{ data }">
            <div class="row-actions" @click.stop>
              <Button
                icon="pi pi-ellipsis-v"
                text
                rounded
                size="small"
                class="more-action"
                title="Open account actions"
                aria-label="Open account actions"
                @click="openActions(data)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

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
        @click="closeActions(); confirmDeactivate(actionTarget)"
      />
      <Button
        v-else
        label="Activate"
        icon="pi pi-user-plus"
        severity="success"
        outlined
        size="small"
        :disabled="updating"
        @click="closeActions(); activate(actionTarget.id)"
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
  padding: 0.8rem 1rem;
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
  grid-template-columns: minmax(230px, 1fr) 165px auto auto;
  align-items: center;
  gap: 0.55rem;
  width: min(850px, 100%);
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
  border-color: #d14350;
  box-shadow: 0 0 0 3px rgba(209, 67, 80, 0.08);
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
  background: #fff1f2;
  color: #d14350;
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
  border-color: #f3b9c0;
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
  background: #fff1f2;
  color: #d14350;
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
</style>
