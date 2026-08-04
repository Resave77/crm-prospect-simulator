<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Password from 'primevue/password'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Toast from 'primevue/toast'
import { useAdminStore } from '../../../stores/admin'
import type { SalesRole } from '../../../types/admin'

const router = useRouter()
const store = useAdminStore()
const toast = useToast()
const error = ref('')
const saving = ref(false)

const managerOptions = computed(() => {
  const managers = store.managerOptions ?? []
  return [{ label: 'Select manager', value: '' }, ...managers.map((m) => ({ label: m.name, value: m.id }))]
})

const organizationalRoleOptions = computed(() =>
  store.salesRoles.filter((role) => role.isActive).map((role) => ({
    label: role.name,
    value: role.id,
    role,
    searchText: `${role.name} level ${role.level} ${landingLabel(role.landingPage)} ${role.permissionCount ?? 0} permissions`,
  })),
)

const form = reactive({
  name: '',
  email: '',
  employeeId: '',
  phone: '',
  salesRoleId: '',
  managerId: '',
  temporaryPassword: '',
})

const selectedOrganizationalRole = computed(() => store.salesRoles.find((role) => role.id === form.salesRoleId) ?? null)
const showManager = computed(() => false)
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const validOrganizationalSelection = computed(() => {
  const role = selectedOrganizationalRole.value
  return Boolean(role?.isActive)
})

const isFormValid = computed(() => {
  if (!form.employeeId.trim()) return false
  if (!form.name.trim()) return false
  if (!emailPattern.test(form.email.trim())) return false
  if (!validOrganizationalSelection.value) return false
  if (form.temporaryPassword.length < 8) return false
  return true
})

watch(() => form.salesRoleId, () => {
  if (!showManager.value) form.managerId = ''
})

function landingLabel(path?: string | null) {
  if (!path) return '-'
  return path.split('/').filter(Boolean).map((part) => part.replace(/-/g, ' ')).join(' / ') || path
}

function roleOptionMeta(role: SalesRole) {
  return `Level ${role.level} · ${role.permissionCount ?? 0} permissions · Landing: ${landingLabel(role.landingPage)}`
}

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  error.value = ''
  try {
    const user = await store.createUser({
      name: form.name.trim(),
      email: form.email.trim(),
      employeeId: form.employeeId.trim(),
      phone: form.phone.trim(),
      salesRoleId: form.salesRoleId || null,
      managerId: null,
      temporaryPassword: form.temporaryPassword,
    })
    toast.add({ severity: 'success', summary: 'Account Created', detail: `Account for ${user.fullName} has been created and is active.`, life: 4000 })
    await new Promise((resolve) => setTimeout(resolve, 800))
    await router.push('/admin/accounts')
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    saving.value = false
  }
}

function managerName(id: string) {
  return store.managerOptions.find((m) => m.id === id)?.name ?? '—'
}

onMounted(async () => {
  try {
    await Promise.all([store.fetchManagers(), store.fetchSalesRoles()])
  } catch { /* optional */ }
})
</script>

<template>
  <section class="admin-page">
    <Toast position="top-right" />

    <!-- PAGE HEADER -->
    <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push('/admin/accounts')" title="Back" />
    <header class="page-heading">
      <div class="page-title-wrapper">
        <span class="eyebrow">New Account</span>
        <h1>Create Account</h1>
        <p class="muted">Register a new user account and assign its role in the CRM system.</p>
      </div>
      <div class="page-heading-actions">
        <Button label="Cancel" severity="secondary" text size="small" @click="router.push('/admin/accounts')" />
        <Button label="Create Account" icon="pi pi-check" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
      </div>
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

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
              <p>Choose the account role from Role Management.</p>
            </div>
          </div>
          <div class="form-grid">
            <div class="form-field full">
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
                  </div>
                </template>
              </Select>
            </div>
            <div class="access-preview full">
              <strong>{{ selectedOrganizationalRole?.name || 'No organizational role selected' }}</strong>
              <span>Level {{ selectedOrganizationalRole?.level ?? '-' }}</span>
              <span>Landing: {{ landingLabel(selectedOrganizationalRole?.landingPage) }}</span>
              <span>{{ selectedOrganizationalRole?.permissionCount ?? 0 }} permissions</span>
              <p>{{ selectedOrganizationalRole?.description || 'Role controls landing page, hierarchy level, and future access rules.' }}</p>
            </div>
            <div class="form-field">
              <label>Temporary Password <span class="required">*</span></label>
              <Password v-model="form.temporaryPassword" toggleMask :feedback="true" :promptLabel="'Choose a temporary password'" placeholder="At least 8 characters" />
            </div>
            <div class="form-field">
              <label>Status</label>
              <div class="readonly-value"><Tag value="ACTIVE" severity="success" /></div>
            </div>
            <div class="form-field">
              <label>Must Change Password</label>
              <div class="readonly-value"><Tag value="Yes" severity="warn" icon="pi pi-key" /></div>
            </div>
            <div v-if="showManager" class="form-field">
              <label>Reports To / Manager <span class="required">*</span></label>
              <Select v-model="form.managerId" :options="managerOptions" optionLabel="label" optionValue="value" placeholder="Select manager" />
            </div>
            <div v-if="showManager" class="form-field form-note-field">
              <div class="field-note">
                <i class="pi pi-info-circle" />
                <span>Manager represents the current direct manager; monthly history is unchanged in this phase.</span>
              </div>
            </div>
          </div>
          <div class="must-change-note">
            <i class="pi pi-key" />
            <span>This user will be required to change their password on first sign-in.</span>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: SIDEBAR -->
      <aside class="form-sidebar">
        <div class="sidebar-card">
          <h4>Role Preview</h4>
          <div class="role-scope">
            <span class="role-title">{{ selectedOrganizationalRole?.name || 'No role selected' }}</span>
            <ul>
              <li><i class="pi pi-check-circle" /> Level {{ selectedOrganizationalRole?.level ?? '-' }}</li>
              <li><i class="pi pi-check-circle" /> {{ selectedOrganizationalRole?.permissionCount ?? 0 }} permissions</li>
              <li><i class="pi pi-check-circle" /> Landing: {{ landingLabel(selectedOrganizationalRole?.landingPage) }}</li>
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
              <strong>{{ selectedOrganizationalRole?.name || '—' }}</strong>
            </div>
            <div class="summary-row">
              <span>Manager</span>
              <strong>{{ showManager ? managerName(form.managerId) : '—' }}</strong>
            </div>
            <div class="summary-row">
              <span>Status</span>
              <strong>Active</strong>
            </div>
            <div class="summary-row">
              <span>Password Change</span>
              <strong>Required</strong>
            </div>
          </div>
        </div>
        <div class="sidebar-actions">
          <Button label="Create Account" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
          <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push('/admin/accounts')" />
        </div>
      </aside>
    </div>
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
.si-green { background: #ecfdf5; color: #059669; }

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
.form-field.full {
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
  min-width: 0;
}
.role-option strong {
  font-size: 0.82rem;
  color: var(--text-primary);
}
.role-option span {
  font-size: 0.72rem;
  color: var(--text-muted);
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
  .form-field.full { grid-column: 1; }
}
</style>
