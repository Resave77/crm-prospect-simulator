<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import Skeleton from 'primevue/skeleton'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAdminStore } from '../../../stores/admin'
import { useAuthStore } from '../../../stores/auth'
import type { AdminUserListItem } from '../../../types/admin'

const store = useAdminStore()
const auth = useAuthStore()
const router = useRouter()
const error = ref('')
const updating = ref(false)
const first = ref(0)
const deactivateDialogVisible = ref(false)
const deactivateTarget = ref<{ id: string; name: string } | null>(null)
const deleteDialogVisible = ref(false)
const deleteTarget = ref<AdminUserListItem | null>(null)

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

function organizationalRoleLabel(user: AdminUserListItem) {
  return fallback(user.organizationalRole?.name)
}

function organizationalRoleMeta(user: AdminUserListItem) {
  const role = user.organizationalRole
  if (!role) return 'No organizational role'
  return `${role.permissionCount ?? 0} permissions`
}

function reportsToLabel(user: AdminUserListItem & Record<string, unknown>) {
  return fallback(String(user.parentName || user.managerName || user.reportsToName || ''))
}

function updatedLabel(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
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
  updating.value = true
  error.value = ''
  try {
    await store.deleteUser(deleteTarget.value.id)
    deleteDialogVisible.value = false
    deleteTarget.value = null
  } catch (e) {
    error.value = store.errorMessage(e)
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
  <section class="admin-page">
    <!-- PAGE HEADER -->
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">Account Management</span>
        <h1>Account List</h1>
      </div>
      <div class="page-heading-actions">
        <Button label="Create Account" icon="pi pi-plus" size="small" @click="router.push('/admin/accounts/create')" />
      </div>
    </header>

    <!-- ERROR -->
    <Message v-if="error" severity="error">{{ error }}</Message>

    <!-- FILTERS -->
    <div class="filter-panel">
      <div class="search-row">
        <div class="search-field">
          <i class="pi pi-search" />
          <input
            type="text"
            placeholder="Search by name, email, or employee ID..."
            :value="store.params.search"
            @input="onKeywordSearch(($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <div class="filter-grid">
        <div class="filter-field">
          <label>Status</label>
          <Select v-model="selectedStatus" :options="statusOptions" optionLabel="label" optionValue="value" placeholder="All Status" />
        </div>
        <div class="filter-field filter-action">
          <Button label="Reset" icon="pi pi-replay" severity="secondary" text size="small" @click="resetAll" />
        </div>
      </div>
    </div>

    <!-- TABLE -->
    <div class="table-panel">
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
        @page="onPage"
      >
        <template #empty>
          <div class="empty-state">
            <div class="empty-icon"><i class="pi pi-users" /></div>
            <strong>No accounts found</strong>
            <span>Adjust your search or filters to view results.</span>
          </div>
        </template>

        <Column header="Employee" :style="{ width: '30%' }">
          <template #body="{ data }">
            <div class="employee-cell">
              <code class="employee-id">{{ data.employeeId || '-' }}</code>
              <span class="cell-primary">{{ data.fullName }}</span>
              <span class="cell-muted">{{ data.email }}</span>
              <span v-if="data.mustChangePassword" class="cell-warn">
                <i class="pi pi-key" /> Password change required
              </span>
            </div>
          </template>
        </Column>
        <Column header="Role" :style="{ width: '26%' }">
          <template #body="{ data }">
            <div class="cell-stack">
              <span class="cell-primary">{{ organizationalRoleLabel(data) }}</span>
              <span class="cell-muted">Level {{ data.organizationalRole?.level ?? '-' }} · {{ organizationalRoleMeta(data) }}</span>
            </div>
          </template>
        </Column>
        <Column header="Reports To" :style="{ width: '16%' }" class="optional-column">
          <template #body="{ data }">
            <span class="cell-text">{{ reportsToLabel(data) }}</span>
          </template>
        </Column>
        <Column header="Status" :style="{ width: '10%' }">
          <template #body="{ data }">
            <Tag :value="data.status === 'ACTIVE' ? 'Active' : 'Inactive'" :severity="data.status === 'ACTIVE' ? 'success' : 'secondary'" size="small" class="soft-tag" />
          </template>
        </Column>
        <Column header="Updated" :style="{ width: '10%' }" class="optional-column">
          <template #body="{ data }">
            <span class="cell-text">{{ updatedLabel(data.updatedAt) }}</span>
          </template>
        </Column>
        <Column header="Actions" :style="{ width: '8%' }">
          <template #body="{ data }">
            <div class="row-actions">
              <Button icon="pi pi-eye" text rounded size="small" title="View account" @click="router.push(`/admin/accounts/${data.id}`)" />
              <Button icon="pi pi-pencil" text rounded size="small" title="Edit account" @click="router.push(`/admin/accounts/${data.id}/edit`)" />
              <Button
                v-if="data.status === 'ACTIVE'"
                icon="pi pi-user-minus"
                text
                rounded
                size="small"
                class="act-delete"
                :disabled="isSelf(data.id) || isProtectedSuperAdmin(data) || updating"
                :title="isProtectedSuperAdmin(data) ? 'Yummy Super Admin is protected' : isSelf(data.id) ? 'You cannot deactivate your own account' : 'Deactivate account'"
                @click="confirmDeactivate(data)"
              />
              <Button
                v-else
                icon="pi pi-user-plus"
                text
                rounded
                size="small"
                class="act-activate"
                :disabled="updating"
                title="Activate account"
                @click="activate(data.id)"
              />
              <Button icon="pi pi-trash" text rounded size="small" class="act-delete" :disabled="isSelf(data.id) || isProtectedSuperAdmin(data) || updating" title="Delete account" @click="confirmDelete(data)" />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- DEACTIVATE CONFIRMATION DIALOG -->
    <Dialog v-model:visible="deactivateDialogVisible" header="Deactivate Account" modal :style="{ width: '400px' }">
      <p>Are you sure you want to deactivate <strong>{{ deactivateTarget?.name }}</strong>? The user will no longer be able to sign in.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deactivateDialogVisible = false" :disabled="updating" />
        <Button label="Deactivate" severity="danger" icon="pi pi-user-minus" :loading="updating" @click="executeDeactivate" />
      </template>
    </Dialog>

    <Dialog v-model:visible="deleteDialogVisible" header="Delete Account" modal :style="{ width: '400px' }">
      <p>Delete <strong>{{ deleteTarget?.fullName }}</strong>? This will fail if the account is still referenced by existing records.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="updating" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="updating" @click="executeDelete" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════════════
   ADMIN — ACCOUNT LIST — professional, white-surface visual pass
   ════════════════════════════════════════════════════════════════ */

/* ── PAGE ─────────────────────────────────────────────────────────── */
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.35rem 1.6rem;
  min-height: 100vh;
  background: #ffffff;
}

/* ── PAGE HEADER ──────────────────────────────────────────────────── */
.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}
.page-title-wrapper {
  display: flex;
  flex-direction: column;
}
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--brand-green-light, #0b7766);
  margin-bottom: 0.35rem;
}
.page-title-wrapper h1 {
  font-size: 1.6rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 0.25rem;
  letter-spacing: -0.03em;
  line-height: 1.15;
}
.page-title-wrapper .muted {
  font-size: 0.84rem;
  color: #7c8798;
  max-width: 520px;
  line-height: 1.55;
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}
.page-heading-actions :deep(.p-button) {
  border-radius: 10px;
  font-weight: 700;
  box-shadow: 0 6px 16px -6px rgba(37, 99, 235, 0.4);
}

/* ── FILTER PANEL ─────────────────────────────────────────────────── */
.filter-panel {
  background: #ffffff;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  padding: 0.85rem 1rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}
.search-row {
  margin-bottom: 0.65rem;
}
.search-field {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.62rem 0.95rem;
  background: #f8fafc;
  border: 1px solid #e6eaf0;
  border-radius: 11px;
  transition: border-color 160ms ease, box-shadow 160ms ease, background 160ms ease;
}
.search-field:focus-within {
  background: #ffffff;
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.09);
}
.search-field i {
  color: #a3adba;
  font-size: 0.9rem;
}
.search-field input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 0.87rem;
  color: #0f172a;
}
.search-field input::placeholder {
  color: #a3adba;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr) auto;
  gap: 0.8rem;
  align-items: end;
}
.filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.32rem;
}
.filter-field label {
  font-size: 0.67rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #8492a6;
}
.filter-field :deep(.p-select) {
  border-radius: 11px;
  border-color: #e6eaf0;
}
.filter-field :deep(.p-select:not(.p-disabled).p-focus) {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.09);
}
.filter-action {
  justify-content: flex-end;
  padding-bottom: 0.15rem;
}

/* ── TABLE PANEL ──────────────────────────────────────────────────── */
.table-panel {
  background: #ffffff;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
  overflow: hidden;
}

/* Force a clean white surface for the PrimeVue DataTable regardless of theme */
.table-panel :deep(.p-datatable) {
  background: #ffffff;
  color: #1e293b;
}
.table-panel :deep(.p-datatable-table) {
  background: #ffffff;
  table-layout: fixed;
  width: 100%;
}
.table-panel :deep(.p-datatable-thead > tr > th) {
  background: #f8fafc;
  color: #64748b;
  font-size: 0.66rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-color: #edf1f6;
  padding: 0.58rem 0.75rem;
}
.table-panel :deep(.p-datatable-tbody > tr) {
  background: #ffffff;
  transition: background 140ms ease;
}
.table-panel :deep(.p-datatable-tbody > tr > td) {
  background: transparent;
  color: #1e293b;
  border-color: #f1f4f8;
  padding: 0.52rem 0.75rem;
  font-size: 0.8rem;
}
.table-panel :deep(.p-datatable-tbody > tr:hover) {
  background: #f8fafc;
}
.table-panel :deep(.p-datatable-tbody > tr:hover > td) {
  background: #f8fafc;
}
.table-panel :deep(.p-paginator) {
  background: #ffffff;
  border-color: #edf1f6;
  padding: 0.5rem 0.75rem;
}
.table-panel :deep(.p-paginator .p-paginator-current) {
  color: #7c8798;
  font-size: 0.76rem;
  font-weight: 500;
}
.table-panel :deep(.p-paginator-page.p-highlight) {
  background: #2563eb;
  color: #ffffff;
  border-radius: 8px;
}
.table-panel :deep(.p-datatable-loading-overlay) {
  background: rgba(255, 255, 255, 0.7);
}

/* ── SKELETON LOADING ─────────────────────────────────────────────── */
.skeleton-area {
  padding: 0.3rem 1.1rem 0.85rem;
}
.skeleton-row {
  height: 3.2rem;
  margin-top: 0.75rem;
  border-radius: 10px;
}

/* ── EMPTY STATE ──────────────────────────────────────────────────── */
.empty-state {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 2rem;
  text-align: center;
  color: #8492a6;
}
.empty-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: #f8fafc;
  display: grid;
  place-content: center;
  margin-bottom: 0.4rem;
}
.empty-icon i {
  font-size: 1.4rem;
  color: #a3adba;
}
.empty-state strong {
  color: #0f172a;
  font-size: 0.95rem;
}

/* ── TABLE CELLS ──────────────────────────────────────────────────── */
.code-tag,
.employee-id {
  display: inline-block;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.68rem;
  font-weight: 600;
  padding: 0.12rem 0.42rem;
  border-radius: 6px;
  background: #eef4fb;
  color: #35628f;
}
.code-blue {
  background: #eff6ff;
  color: #2563eb;
}

.cell-stack,
.employee-cell {
  display: flex;
  flex-direction: column;
  gap: 0.08rem;
}
.cell-primary {
  font-weight: 650;
  font-size: 0.82rem;
  color: #0f172a;
}
.cell-muted {
  color: #94a3b8;
  font-size: 0.72rem;
}
.cell-warn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  margin-top: 0.15rem;
  font-size: 0.67rem;
  font-weight: 600;
  color: #b45309;
}
.cell-text {
  font-size: 0.8rem;
  color: #475569;
}
.soft-tag {
  font-size: 0.68rem;
  border-radius: 999px;
}

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.05rem;
  white-space: nowrap;
}
.row-actions :deep(.p-button) { width: 2rem; height: 2rem; padding: 0; }
.row-actions :deep(.p-button-label) { display: none; }
.row-actions :deep(.p-button) {
  transition: transform 140ms ease, background 140ms ease;
}
.row-actions :deep(.p-button:hover) {
  transform: translateY(-1px);
}
.act-activate {
  color: #059669 !important;
}
.act-activate:hover {
  background: #ecfdf5 !important;
}
.act-delete {
  color: #dc2626 !important;
}
.act-delete:hover {
  background: #fef2f2 !important;
}

/* ── RESPONSIVE ───────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .filter-grid {
    grid-template-columns: 1fr auto;
  }
  .filter-action {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }
  .table-panel :deep(.optional-column) {
    display: none;
  }
}
@media (max-width: 768px) {
  .admin-page {
    padding: 1.25rem 1rem;
    gap: 1rem;
  }
  .page-heading {
    flex-direction: column;
    align-items: stretch;
  }
  .page-heading-actions {
    width: 100%;
    justify-content: flex-end;
  }
  .filter-grid {
    grid-template-columns: 1fr;
  }
  .table-panel :deep(.p-datatable-thead) {
    display: none;
  }
  .table-panel :deep(.p-datatable-table),
  .table-panel :deep(.p-datatable-tbody),
  .table-panel :deep(.p-datatable-tbody > tr),
  .table-panel :deep(.p-datatable-tbody > tr > td) {
    display: block;
    width: 100%;
  }
  .table-panel :deep(.p-datatable-tbody > tr) {
    border-bottom: 1px solid #edf1f6;
    padding: 0.7rem;
  }
  .table-panel :deep(.p-datatable-tbody > tr > td) {
    border: 0;
    padding: 0.35rem 0;
  }
  .row-actions {
    justify-content: flex-start;
    padding-top: 0.25rem;
  }
}
@media (max-width: 480px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
