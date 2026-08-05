<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import axios from 'axios'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Toast from 'primevue/toast'
import { useAdminStore } from '../../../stores/admin'
import type { ApiErrorEnvelope } from '../../../types/auth'
import type { SalesRole } from '../../../types/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()
const toast = useToast()
const error = ref('')
const notFound = ref(false)
const saving = ref(false)
const loaded = ref(false)

const id = computed(() => String(route.params.id))

const organizationalRoleOptions = computed(() => {
  const currentRole = store.selectedUser?.organizationalRole
  const options = store.salesRoles
    .filter((role) => isAssignableSalesRole(role) || (role.id === currentRole?.id && role.name.trim().toLowerCase() !== 'super admin'))
    .map((role) => ({
      label: role.name,
      value: role.id,
      role,
      searchText: `${role.name} level ${role.level} ${landingLabel(role.landingPage)} ${role.permissionCount ?? 0} permissions`,
    }))
  if (currentRole && !options.some((option) => option.value === currentRole.id)) {
    options.unshift({
      label: currentRole.name,
      value: currentRole.id,
      role: { ...currentRole, createdAt: '', updatedAt: '', permissions: [] } as SalesRole,
      searchText: `${currentRole.name} level ${currentRole.level} ${landingLabel(currentRole.landingPage)} inactive`,
    })
  }
  return options
})

const form = reactive({
  accountType: 'SALES_ACCOUNT' as 'SUPER_ADMIN' | 'SALES_ACCOUNT',
  employeeId: '',
  name: '',
  email: '',
  phone: '',
  salesRoleId: '',
})

const accountTypeOptions = [
  { label: 'Sales Account', value: 'SALES_ACCOUNT', description: 'Uses active roles from Role Management.' },
  { label: 'Super Admin', value: 'SUPER_ADMIN', description: 'System administrator account outside Sales Structure.' },
]
const isSalesAccount = computed(() => form.accountType === 'SALES_ACCOUNT')
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const selectedOrganizationalRole = computed(() => {
  const fromList = store.salesRoles.find((role) => role.id === form.salesRoleId)
  const current = store.selectedUser?.organizationalRole
  if (fromList) return fromList
  if (current?.id === form.salesRoleId) return current
  return null
})
const roleChanged = computed(() => Boolean(store.selectedUser?.organizationalRole?.id && form.salesRoleId && form.salesRoleId !== store.selectedUser.organizationalRole.id))
const validOrganizationalSelection = computed(() => {
  const role = selectedOrganizationalRole.value
  return Boolean(role?.isActive || (store.selectedUser?.organizationalRole?.id === role?.id))
})

const isFormValid = computed(() => {
  if (!form.employeeId.trim()) return false
  if (!form.name.trim()) return false
  if (!emailPattern.test(form.email.trim())) return false
  if (isSalesAccount.value && !validOrganizationalSelection.value) return false
  return true
})

function isNotFoundError(e: unknown) {
  return axios.isAxiosError<ApiErrorEnvelope>(e)
    && (e.response?.status === 404 || e.response?.data?.error?.code === 'USER_NOT_FOUND')
}

async function load() {
  error.value = ''
  notFound.value = false
  try {
    const user = await store.fetchUserById(id.value)
    form.employeeId = user.employeeId
    form.name = user.fullName
    form.email = user.email
    form.phone = user.phone
    form.accountType = user.role === 'SUPER_ADMIN' ? 'SUPER_ADMIN' : 'SALES_ACCOUNT'
    form.salesRoleId = user.organizationalRole?.id ?? ''
    loaded.value = true
  } catch (e) {
    notFound.value = isNotFoundError(e)
    error.value = store.errorMessage(e)
  }
}

function landingLabel(path?: string | null) {
  if (!path) return '-'
  return path.split('/').filter(Boolean).map((part) => part.replace(/-/g, ' ')).join(' / ') || path
}

function roleOptionMeta(role: SalesRole) {
  return `Level ${role.level} · ${role.permissionCount ?? 0} permissions · Landing: ${landingLabel(role.landingPage)}`
}

function isAssignableSalesRole(role: SalesRole) {
  return role.isActive && role.level >= 1 && role.level <= 4 && role.name.trim().toLowerCase() !== 'super admin'
}

watch(
  () => form.accountType,
  (accountType) => {
    if (accountType === 'SUPER_ADMIN') {
      form.salesRoleId = ''
    }
  },
)

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  error.value = ''
  try {
    const user = await store.updateUser(id.value, {
      employeeId: form.employeeId.trim(),
      name: form.name.trim(),
      email: form.email.trim(),
      phone: form.phone.trim(),
      accountType: form.accountType,
      salesRoleId: isSalesAccount.value ? form.salesRoleId || null : null,
    })
    toast.add({ severity: 'success', summary: 'Account Updated', detail: `Account for ${user.fullName} has been updated.`, life: 4000 })
    await new Promise((resolve) => setTimeout(resolve, 800))
    await router.push(`/admin/accounts/${id.value}`)
  } catch (e) {
    const message = store.errorMessage(e)
    if (message.toLowerCase().includes('sales executive must have a manager')) {
      error.value =
        'Role could not be updated because the backend still requires Sales Executive accounts to have a manager. The account form no longer clears or changes the existing manager.'
    } else {
      error.value = message
    }
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await load()
  if (!notFound.value) {
    try { await Promise.all([store.fetchManagers(), store.fetchSalesRoles()]) } catch { /* optional */ }
  }
})
</script>

<template>
  <section class="admin-page">
    <Toast position="top-right" />

    <!-- NOT FOUND -->
    <template v-if="notFound">
      <div class="state-box">
        <div class="state-icon-wrap"><i class="pi pi-user-slash" /></div>
        <strong>Account not found</strong>
        <span class="muted">The requested account does not exist or was removed.</span>
        <Button label="Back to Account List" icon="pi pi-arrow-left" size="small" @click="router.push('/admin/accounts')" />
      </div>
    </template>

    <template v-else>
      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

      <!-- LOADING -->
      <div v-if="store.detailLoading && !loaded" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading account...</span>
      </div>

      <template v-if="loaded">
        <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push('/admin/accounts')" title="Back to account list" />

        <!-- PAGE HEADER -->
        <header class="page-heading">
          <div class="page-title-wrapper">
            <span class="eyebrow">Edit Account</span>
            <h1>Edit Account</h1>
            <div class="subtitle-row">
              <code class="code-tag code-blue">{{ form.employeeId || '—' }}</code>
              <span class="muted">&mdash;</span>
              <span class="muted">{{ form.name || 'No name set' }}</span>
            </div>
          </div>
          <div class="page-heading-actions">
            <Button label="Cancel" severity="secondary" text size="small" @click="router.push(`/admin/accounts/${id}`)" />
            <Button label="Save Changes" icon="pi pi-check" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
          </div>
        </header>

        <div class="form-layout">
          <!-- LEFT COLUMN: FORM -->
          <div class="form-stack">
            <!-- ACCOUNT INFORMATION -->
            <div class="form-card">
              <div class="form-card-header">
                <div class="form-card-icon si-blue"><i class="pi pi-user" /></div>
                <div>
                  <h3>Account Information</h3>
                  <p>Basic details about the user account.</p>
                </div>
              </div>
              <div class="form-grid">
                <div class="form-field">
                  <label>Employee ID <span class="required">*</span></label>
                  <InputText v-model="form.employeeId" placeholder="e.g. YF-2026-001" />
                </div>
                <div class="form-field">
                  <label>Full Name <span class="required">*</span></label>
                  <InputText v-model="form.name" placeholder="e.g. Budi Santoso" />
                </div>
                <div class="form-field">
                  <label>Email <span class="required">*</span></label>
                  <InputText v-model="form.email" placeholder="e.g. budi@yummy.test" />
                </div>
                <div class="form-field">
                  <label>Phone</label>
                  <InputText v-model="form.phone" placeholder="e.g. 0812-3456-7890" />
                </div>
              </div>
            </div>

            <!-- ROLE & ACCESS -->
            <div class="form-card">
              <div class="form-card-header">
                <div class="form-card-icon si-violet"><i class="pi pi-lock" /></div>
                <div>
                  <h3>Role &amp; Access</h3>
                  <p>Adjust the account role. The existing reporting relationship is preserved.</p>
                </div>
              </div>
              <div class="form-grid">
                <div class="form-field full">
                  <label>Account Type <span class="required">*</span></label>
                  <Select
                    v-model="form.accountType"
                    :options="accountTypeOptions"
                    optionLabel="label"
                    optionValue="value"
                  >
                    <template #option="{ option }">
                      <div class="role-option">
                        <strong>{{ option.label }}</strong>
                        <span>{{ option.description }}</span>
                      </div>
                    </template>
                  </Select>
                </div>
                <div v-if="isSalesAccount" class="form-field full">
                  <label>Role <span class="required">*</span></label>
                  <Select
                    v-model="form.salesRoleId"
                    :options="organizationalRoleOptions"
                    optionLabel="label"
                    optionValue="value"
                    filter
                    :filterFields="['label', 'searchText']"
                    placeholder="Search active roles"
                    :loading="store.salesRolesLoading"
                  >
                    <template #option="{ option }">
                      <div class="role-option">
                        <strong>{{ option.role.name }}</strong>
                        <span>{{ roleOptionMeta(option.role) }}</span>
                        <em v-if="!option.role.isActive">Inactive</em>
                      </div>
                    </template>
                  </Select>
                </div>
                <div class="access-preview full">
                  <strong>{{ isSalesAccount ? selectedOrganizationalRole?.name || 'No organizational role selected' : 'Super Admin' }}</strong>
                  <span>Level {{ isSalesAccount ? selectedOrganizationalRole?.level ?? '-' : '-' }}</span>
                  <span>Landing: {{ isSalesAccount ? landingLabel(selectedOrganizationalRole?.landingPage) : 'admin / dashboard' }}</span>
                  <span>{{ isSalesAccount ? selectedOrganizationalRole?.permissionCount ?? 0 : 'System' }} permissions</span>
                  <p>{{ isSalesAccount ? selectedOrganizationalRole?.description || 'Role controls landing page, hierarchy level, and future access rules.' : 'Super Admin is a system role and is not assigned to Sales Structure.' }}</p>
                </div>
                <div v-if="roleChanged" class="must-change-note full">
                  <i class="pi pi-exclamation-triangle" />
                  <span>Changing this role may change available menus, landing page, allowed actions, and hierarchy level. The existing manager is not changed from this form.</span>
                </div>
                <div class="form-field">
                  <label>Status</label>
                  <div class="readonly-value"><Tag :value="store.selectedUser?.status ?? '—'" :severity="store.selectedUser?.status === 'ACTIVE' ? 'success' : 'secondary'" /></div>
                </div>
                <div class="form-field">
                  <label>Must Change Password</label>
                  <div class="readonly-value">
                    <Tag :value="store.selectedUser?.mustChangePassword ? 'Yes' : 'No'" :severity="store.selectedUser?.mustChangePassword ? 'warn' : 'secondary'" :icon="store.selectedUser?.mustChangePassword ? 'pi pi-key' : ''" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- RIGHT COLUMN: SIDEBAR -->
          <aside class="form-sidebar">
            <div class="sidebar-card">
              <h4>Role Preview</h4>
              <div class="role-scope">
                <span class="role-title">{{ isSalesAccount ? selectedOrganizationalRole?.name || 'No role selected' : 'Super Admin' }}</span>
                <ul>
                  <li><i class="pi pi-check-circle" /> Level {{ isSalesAccount ? selectedOrganizationalRole?.level ?? '-' : '-' }}</li>
                  <li><i class="pi pi-check-circle" /> {{ isSalesAccount ? selectedOrganizationalRole?.permissionCount ?? 0 : 'System' }} permissions</li>
                  <li><i class="pi pi-check-circle" /> Landing: {{ isSalesAccount ? landingLabel(selectedOrganizationalRole?.landingPage) : 'admin / dashboard' }}</li>
                </ul>
              </div>
            </div>
            <div class="sidebar-card">
              <h4>Submission Summary</h4>
              <div class="summary-list">
                <div class="summary-row">
                  <span>Employee ID</span>
                  <strong>{{ form.employeeId || '—' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Name</span>
                  <strong>{{ form.name || '—' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Email</span>
                  <strong>{{ form.email || '—' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Role</span>
                  <strong>{{ isSalesAccount ? selectedOrganizationalRole?.name || '—' : 'Super Admin' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Current Manager</span>
                  <strong>{{ store.selectedUser?.managerName || '—' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Status</span>
                  <strong>{{ store.selectedUser?.status ?? '—' }}</strong>
                </div>
              </div>
            </div>
            <div class="sidebar-actions">
              <Button label="Save Changes" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
              <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push(`/admin/accounts/${id}`)" />
            </div>
          </aside>
        </div>
      </template>
    </template>
  </section>
</template>

<style scoped>
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
  gap: 0.15rem;
}
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--brand-green-light, #0b7766);
  margin-top: 0.5rem;
}
.page-title-wrapper h1 {
  font-size: 1.65rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0.2rem 0 0.15rem;
  letter-spacing: -0.03em;
}
.subtitle-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.page-title-wrapper .muted {
  font-size: 0.85rem;
  color: var(--text-muted);
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}

/* ── FORM LAYOUT ──────────────────────────────────────────────────── */
.form-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
  align-items: start;
}
.form-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

/* ── FORM CARDS ────────────────────────────────────────────────────── */
.form-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  box-shadow: var(--shadow-xs);
}
.form-card-header {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f0f3f7;
}
.form-card-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: grid;
  place-content: center;
  font-size: 1rem;
  flex-shrink: 0;
}
.si-blue { background: #eff6ff; color: #2563eb; }
.si-violet { background: #eef2ff; color: #6366f1; }

.form-card-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
}
.form-card-header p {
  margin: 0.15rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* ── FORM GRID ─────────────────────────────────────────────────────── */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.form-field.full,
.full {
  grid-column: 1 / -1;
}
.form-field label {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.required {
  color: #dc2626;
}

.readonly-value {
  flex: 1;
  min-height: 2.85rem;
  display: flex;
  align-items: center;
  padding: 0 0.2rem;
}

.form-note-field {
  justify-content: flex-end;
}
.field-note {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem 0.7rem;
  font-size: 0.75rem;
  color: #1e40af;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: var(--radius-sm);
  line-height: 1.45;
}
.field-note i {
  flex-shrink: 0;
  font-size: 0.85rem;
}

.role-option {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.role-option strong {
  font-size: 0.82rem;
  color: var(--text-primary);
}
.role-option span,
.role-option em {
  font-size: 0.72rem;
  color: var(--text-muted);
}
.role-option em {
  color: #b45309;
  font-style: normal;
  font-weight: 700;
}
.access-preview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.45rem 0.75rem;
  padding: 0.8rem 0.9rem;
  border: 1px solid #dbeafe;
  border-radius: var(--radius-sm);
  background: #f8fbff;
}
.access-preview strong,
.access-preview p {
  grid-column: 1 / -1;
}
.access-preview strong {
  color: var(--text-primary);
  font-size: 0.9rem;
}
.access-preview span,
.access-preview p {
  margin: 0;
  font-size: 0.76rem;
  color: var(--text-muted);
  line-height: 1.45;
}

.must-change-note {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1.1rem;
  padding: 0.65rem 0.8rem;
  font-size: 0.76rem;
  color: #92400e;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: var(--radius-sm);
}
.must-change-note i {
  flex-shrink: 0;
  color: #d97706;
}

/* ── SIDEBAR ───────────────────────────────────────────────────────── */
.form-sidebar {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: sticky;
  top: 1.5rem;
}
.sidebar-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.15rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.sidebar-card h4 {
  margin: 0 0 0.85rem;
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.role-scope {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.role-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--brand-blue);
}
.role-scope ul {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}
.role-scope li {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  font-size: 0.78rem;
  color: var(--text-secondary);
  line-height: 1.45;
}
.role-scope li i {
  margin-top: 0.1rem;
  font-size: 0.8rem;
  color: #059669;
  flex-shrink: 0;
}
.summary-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
}
.summary-row span {
  font-size: 0.78rem;
  color: var(--text-muted);
}
.summary-row strong {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-primary);
  text-align: right;
}
.sidebar-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.full-width {
  width: 100%;
}

/* ── CODE TAG ─────────────────────────────────────────────────────── */
.code-tag {
  display: inline-block;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: #f1f5f9;
  color: var(--text-secondary);
  word-break: break-all;
}
.code-blue {
  background: #eff6ff;
  color: #2563eb;
}

/* ── STATE BOX ────────────────────────────────────────────────────── */
.state-box {
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
.state-icon {
  font-size: 1.75rem;
  color: var(--brand-blue);
  margin-bottom: 0.25rem;
}
.state-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--surface-subtle);
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.state-icon-wrap i {
  font-size: 1.4rem;
  color: var(--text-faint);
}
.state-box strong {
  color: var(--text-primary);
  font-size: 0.95rem;
}

/* ── RESPONSIVE ────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .form-layout {
    grid-template-columns: 1fr;
  }
  .form-sidebar {
    position: static;
    order: -1;
  }
}
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .form-grid { grid-template-columns: 1fr; }
  .form-field.full,
  .full {
    grid-column: 1;
  }
}
</style>
