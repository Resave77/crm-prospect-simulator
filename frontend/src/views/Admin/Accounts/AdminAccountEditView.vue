<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
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
import type { AdminUserRole } from '../../../types/admin'
import { adminScopeSummary } from '../../../utils/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()
const toast = useToast()
const error = ref('')
const notFound = ref(false)
const saving = ref(false)
const loaded = ref(false)

const id = computed(() => String(route.params.id))

const roleOptions = [
  { label: 'Select role', value: '' },
  { label: 'Administrator', value: 'ADMINISTRATOR' },
  { label: 'Sales Manager', value: 'SALES_MANAGER' },
  { label: 'Sales Executive', value: 'SALES_EXECUTIVE' },
]

const managerOptions = computed(() => {
  const managers = store.managerOptions ?? []
  const current = store.selectedUser
  const options = managers.map((m) => ({ label: m.name, value: m.id }))
  if (current?.managerId && !managers.some((m) => m.id === current.managerId)) {
    options.unshift({ label: current.managerName || 'Current manager', value: current.managerId })
  }
  return [{ label: 'Select manager', value: '' }, ...options]
})

const form = reactive({
  employeeId: '',
  name: '',
  email: '',
  phone: '',
  role: '' as AdminUserRole | '',
  managerId: '',
})

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const showManager = computed(() => form.role === 'SALES_EXECUTIVE')
const scope = computed(() => adminScopeSummary(form.role))

const isFormValid = computed(() => {
  if (!form.employeeId.trim()) return false
  if (!form.name.trim()) return false
  if (!emailPattern.test(form.email.trim())) return false
  if (!form.role) return false
  if (showManager.value && !form.managerId) return false
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
    form.role = user.role
    form.managerId = user.managerId ?? ''
    loaded.value = true
  } catch (e) {
    notFound.value = isNotFoundError(e)
    error.value = store.errorMessage(e)
  }
}

function onRoleChange() {
  form.managerId = ''
}

function managerName(managerId: string) {
  return store.managerOptions.find((m) => m.id === managerId)?.name ?? '—'
}

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
      role: form.role as AdminUserRole,
      managerId: showManager.value ? form.managerId : null,
    })
    toast.add({ severity: 'success', summary: 'Account Updated', detail: `Account for ${user.fullName} has been updated.`, life: 4000 })
    await router.push(`/admin/accounts/${id.value}`)
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await load()
  if (!notFound.value) {
    try { await store.fetchManagers() } catch { /* optional */ }
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
        <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push(`/admin/accounts/${id}`)" title="Back to account detail" />

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
                  <p>Adjust the role and reporting line for this account.</p>
                </div>
              </div>
              <div class="form-grid">
                <div class="form-field">
                  <label>Role <span class="required">*</span></label>
                  <Select v-model="form.role" :options="roleOptions" optionLabel="label" optionValue="value" placeholder="Select role" @change="onRoleChange" />
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
                <div v-if="showManager" class="form-field">
                  <label>Manager <span class="required">*</span></label>
                  <Select v-model="form.managerId" :options="managerOptions" optionLabel="label" optionValue="value" placeholder="Select manager" />
                </div>
                <div v-if="showManager" class="form-field form-note-field">
                  <div class="field-note">
                    <i class="pi pi-info-circle" />
                    <span>Sales executives must be assigned to an active Sales Manager.</span>
                  </div>
                </div>
              </div>
              <div v-if="showManager && !store.selectedUser?.managerId" class="must-change-note">
                <i class="pi pi-user-plus" />
                <span>This account currently has no manager. A Sales Executive must be assigned to a manager.</span>
              </div>
              <div v-else-if="!showManager && store.selectedUser?.managerId" class="must-change-note">
                <i class="pi pi-exclamation-triangle" />
                <span>This account has a manager. Roles Administrator and Sales Manager must not have a manager.</span>
              </div>
            </div>
          </div>

          <!-- RIGHT COLUMN: SIDEBAR -->
          <aside class="form-sidebar">
            <div class="sidebar-card">
              <h4>Role Scope</h4>
              <div class="role-scope">
                <span class="role-title">{{ scope.title }}</span>
                <ul>
                  <li v-for="(item, idx) in scope.scopes" :key="idx">
                    <i class="pi pi-check-circle" /> {{ item }}
                  </li>
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
                  <strong>{{ form.role ? scope.title : '—' }}</strong>
                </div>
                <div class="summary-row">
                  <span>Manager</span>
                  <strong>{{ showManager ? (managerName(form.managerId) || '—') : '—' }}</strong>
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
}
</style>
