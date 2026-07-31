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
import ResetPasswordDialog from '../../../components/admin/ResetPasswordDialog.vue'
import type { AdminUserListItem } from '../../../types/admin'

const store = useAdminStore()
const auth = useAuthStore()
const router = useRouter()
const error = ref('')
const updating = ref(false)
const first = ref(0)
const deactivateDialogVisible = ref(false)
const deactivateTarget = ref<{ id: string; name: string } | null>(null)
const resetDialogVisible = ref(false)
const resetPasswordTarget = ref<AdminUserListItem | null>(null)

const roleOptions = [
  { label: 'All Roles', value: '' },
  { label: 'Administrator', value: 'ADMINISTRATOR' },
  { label: 'Sales Manager', value: 'SALES_MANAGER' },
  { label: 'Sales Executive', value: 'SALES_EXECUTIVE' },
]

const statusOptions = [
  { label: 'All Status', value: '' },
  { label: 'Active', value: 'ACTIVE' },
  { label: 'Inactive', value: 'INACTIVE' },
]

const selectedRole = computed({
  get: () => store.params.role,
  set: (val) => { store.setParam('role', val); store.setParam('page', 1); first.value = 0; load() },
})

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

function roleLabel(role: string) {
  switch (role) {
    case 'ADMINISTRATOR': return 'Administrator'
    case 'SALES_MANAGER': return 'Sales Manager'
    default: return 'Sales Executive'
  }
}

function roleSeverity(role: string) {
  switch (role) {
    case 'ADMINISTRATOR': return 'warn'
    case 'SALES_MANAGER': return 'info'
    default: return 'success'
  }
}

const isSelf = (id: string) => id === auth.user?.id

function openResetPassword(user: AdminUserListItem) {
  resetPasswordTarget.value = user
  resetDialogVisible.value = true
}

function confirmDeactivate(user: AdminUserListItem) {
  deactivateTarget.value = { id: user.id, name: user.fullName }
  deactivateDialogVisible.value = true
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
        <p class="muted">Manage user accounts, roles, and activation status for the CRM system.</p>
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
          <label>Role</label>
          <Select v-model="selectedRole" :options="roleOptions" optionLabel="label" optionValue="value" placeholder="All Roles" />
        </div>
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

        <Column header="Employee ID" :style="{ width: '160px' }">
          <template #body="{ data }">
            <code class="code-tag code-blue">{{ data.employeeId || '—' }}</code>
          </template>
        </Column>
        <Column header="Name" :style="{ minWidth: '220px' }">
          <template #body="{ data }">
            <div class="cell-stack">
              <span class="cell-primary">{{ data.fullName }}</span>
              <span v-if="data.mustChangePassword" class="cell-warn">
                <i class="pi pi-key" /> Password change required
              </span>
            </div>
          </template>
        </Column>
        <Column header="Email" :style="{ minWidth: '200px' }">
          <template #body="{ data }">
            <span class="cell-text">{{ data.email }}</span>
          </template>
        </Column>
        <Column header="Phone" :style="{ minWidth: '140px' }">
          <template #body="{ data }">
            <span class="cell-text">{{ data.phone || '—' }}</span>
          </template>
        </Column>
        <Column header="Role" :style="{ width: '150px' }">
          <template #body="{ data }">
            <Tag :value="roleLabel(data.role)" :severity="roleSeverity(data.role)" />
          </template>
        </Column>
        <Column header="Manager" :style="{ minWidth: '180px' }">
          <template #body="{ data }">
            <span class="cell-text">{{ data.managerName || '—' }}</span>
          </template>
        </Column>
        <Column header="Status" :style="{ width: '110px' }">
          <template #body="{ data }">
            <Tag :value="data.status" :severity="data.status === 'ACTIVE' ? 'success' : 'secondary'" size="small" />
          </template>
        </Column>
        <Column header="Actions" :style="{ width: '260px' }">
          <template #body="{ data }">
            <div class="row-actions">
              <Button icon="pi pi-eye" text rounded size="small" title="View account" @click="router.push(`/admin/accounts/${data.id}`)" />
              <Button icon="pi pi-pencil" text rounded size="small" title="Edit account" @click="router.push(`/admin/accounts/${data.id}/edit`)" />
              <Button icon="pi pi-key" text rounded size="small" title="Reset password" @click="openResetPassword(data)" />
              <Button
                v-if="data.status === 'ACTIVE'"
                label="Deactivate"
                icon="pi pi-user-minus"
                text
                rounded
                size="small"
                class="act-delete"
                :disabled="isSelf(data.id) || updating"
                :title="isSelf(data.id) ? 'You cannot deactivate your own account' : 'Deactivate account'"
                @click="confirmDeactivate(data)"
              />
              <Button
                v-else
                label="Activate"
                icon="pi pi-user-plus"
                text
                rounded
                size="small"
                class="act-activate"
                :disabled="updating"
                title="Activate account"
                @click="activate(data.id)"
              />
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

    <!-- RESET PASSWORD DIALOG -->
    <ResetPasswordDialog v-model:visible="resetDialogVisible" :user="resetPasswordTarget" />
  </section>
</template>

<style scoped>
/* ── PAGE ─────────────────────────────────────────────────────────── */
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.75rem 2rem;
  min-height: 100vh;
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
  font-size: 1.65rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 0.2rem;
  letter-spacing: -0.03em;
  line-height: 1.15;
}
.page-title-wrapper .muted {
  font-size: 0.85rem;
  color: var(--text-muted);
  max-width: 520px;
  line-height: 1.55;
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}

/* ── FILTER PANEL ─────────────────────────────────────────────────── */
.filter-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.1rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.search-row {
  margin-bottom: 0.9rem;
}
.search-field {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 0.9rem;
  background: var(--surface-subtle);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast), background var(--transition-fast);
}
.search-field:focus-within {
  background: var(--surface-card);
  border-color: var(--brand-blue);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}
.search-field i {
  color: var(--text-faint);
  font-size: 0.9rem;
}
.search-field input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 0.87rem;
  color: var(--text-primary);
}
.search-field input::placeholder {
  color: var(--text-faint);
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr) auto;
  gap: 0.75rem;
  align-items: end;
}
.filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.filter-field label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.filter-action {
  justify-content: flex-end;
  padding-bottom: 0.15rem;
}

/* ── TABLE PANEL ──────────────────────────────────────────────────── */
.table-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  overflow: hidden;
}

/* ── SKELETON LOADING ─────────────────────────────────────────────── */
.skeleton-area {
  padding: 0.25rem 1rem 0.75rem;
}
.skeleton-row {
  height: 3.2rem;
  margin-top: 0.75rem;
  border-radius: var(--radius-sm);
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
  color: var(--text-muted);
}
.empty-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--surface-subtle);
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.empty-icon i {
  font-size: 1.4rem;
  color: var(--text-faint);
}
.empty-state strong {
  color: var(--text-primary);
  font-size: 0.95rem;
}

/* ── TABLE CELLS ──────────────────────────────────────────────────── */
.code-tag {
  display: inline-block;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: #f1f5f9;
  color: var(--text-secondary);
}
.code-blue {
  background: #eff6ff;
  color: #2563eb;
}

.cell-stack {
  display: flex;
  flex-direction: column;
}
.cell-primary {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
}
.cell-warn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  margin-top: 0.1rem;
  font-size: 0.68rem;
  font-weight: 600;
  color: #b45309;
}
.cell-text {
  font-size: 0.84rem;
  color: var(--text-secondary);
}

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.row-actions {
  display: flex;
  align-items: center;
  gap: 0.15rem;
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
    grid-template-columns: repeat(3, 1fr);
  }
  .filter-action {
    grid-column: 1 / -1;
    justify-content: flex-end;
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
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 480px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
