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
import type { AdminAccountType, SalesRole, SalesStructureItem } from '../../../types/admin'

const router = useRouter()
const store = useAdminStore()
const toast = useToast()
const error = ref('')
const saving = ref(false)

const organizationalRoleOptions = computed(() =>
  store.salesRoles
    .filter(isAssignableSalesRole)
    .map((role) => ({
      label: role.name,
      value: role.id,
      role,
      searchText: `${role.name} level ${role.level} ${landingLabel(role.landingPage)} ${role.permissionCount ?? 0} permissions`,
    })),
)

const form = reactive({
  accountType: 'SALES_ACCOUNT' as AdminAccountType,
  name: '',
  email: '',
  employeeId: '',
  phone: '',
  salesRoleId: '',
  reportsToUserId: '',
  temporaryPassword: '',
})

const accountTypeOptions: Array<{
  label: string
  value: AdminAccountType
  description: string
}> = [
  {
    label: 'Sales Account',
    value: 'SALES_ACCOUNT',
    description: 'Uses an active Sales Level 1–3 role (hierarchy Level 2–4) and is assigned directly into Sales Structure.',
  },
  {
    label: 'Super Admin',
    value: 'SUPER_ADMIN',
    description: 'System administrator account. Additional Super Admin accounts do not create another Level 1 hierarchy root.',
  },
]
const isSalesAccount = computed(() => form.accountType === 'SALES_ACCOUNT')
const selectedOrganizationalRole = computed(
  () => store.salesRoles.find((role) => role.id === form.salesRoleId) ?? null,
)
const requiredParentLevel = computed(() => selectedOrganizationalRole.value ? selectedOrganizationalRole.value.level - 1 : null)
const reportsToOptions = computed(() =>
  store.salesStructure
    .filter((item: SalesStructureItem) => item.salesRole.level === requiredParentLevel.value)
    .map((item: SalesStructureItem) => ({
      label: item.salesName,
      value: item.userId,
      searchText: `${item.salesName} ${item.salesRole.name} level ${item.salesRole.level}`,
      item,
    })),
)

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const requiredFields = computed(() => [
  { label: 'Account Type', complete: Boolean(form.accountType) },
  { label: 'Full Name', complete: Boolean(form.name.trim()) },
  { label: 'Email', complete: emailPattern.test(form.email.trim()) },
  { label: 'Role', complete: !isSalesAccount.value || Boolean(selectedOrganizationalRole.value?.isActive) },
  { label: 'Reports To', complete: !isSalesAccount.value || Boolean(form.reportsToUserId) },
  { label: 'Temporary Password', complete: form.temporaryPassword.length >= 8 },
])

const completedRequiredCount = computed(
  () => requiredFields.value.filter((item) => item.complete).length,
)

const isFormValid = computed(
  () =>
    Boolean(form.employeeId.trim()) &&
    requiredFields.value.every((item) => item.complete),
)

watch(
  () => form.email,
  (email) => {
    if (!form.employeeId || form.employeeId.startsWith('EMP-')) {
      form.employeeId = generateEmployeeId(email)
    }
  },
)

watch(
  () => form.accountType,
  (accountType) => {
    if (accountType === 'SUPER_ADMIN') {
      form.salesRoleId = ''
      form.reportsToUserId = ''
    }
  },
)

watch(
  () => form.salesRoleId,
  () => {
    form.reportsToUserId = ''
  },
)

function landingLabel(path?: string | null) {
  if (!path) return '-'
  return (
    path
      .split('/')
      .filter(Boolean)
      .map((part) => part.replace(/-/g, ' '))
      .join(' / ') || path
  )
}

function roleOptionMeta(role: SalesRole) {
  return `Level ${role.level} · ${role.permissionCount ?? 0} permissions · Landing: ${landingLabel(role.landingPage)}`
}

function isAssignableSalesRole(role: SalesRole) {
  return role.isActive && role.level >= 2 && role.level <= 4
}

function generateEmployeeId(seed = '') {
  const now = new Date()
  const year = now.getFullYear()
  const normalized = seed
    .split('@')[0]
    ?.replace(/[^a-zA-Z0-9]/g, '')
    .slice(0, 3)
    .toUpperCase()

  const suffix = String(
    Math.floor(1000 + Math.random() * 9000),
  )

  return `EMP-${year}-${normalized || suffix.slice(0, 3)}-${suffix}`
}

function regenerateEmployeeId() {
  form.employeeId = generateEmployeeId(form.email)
}

function todayDate() {
  return new Date().toISOString().slice(0, 10)
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
      accountType: form.accountType,
      salesRoleId: isSalesAccount.value ? form.salesRoleId || null : null,
      managerId: isSalesAccount.value ? form.reportsToUserId || null : null,
      temporaryPassword: form.temporaryPassword,
    })

    toast.add({
      severity: 'success',
      summary: 'Account Created',
      detail: `Account for ${user.fullName} has been created and is active.`,
      life: 4000,
    })

    await new Promise((resolve) => setTimeout(resolve, 700))
    await router.push('/admin/accounts')
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  form.employeeId = generateEmployeeId()

  try {
    await Promise.all([store.fetchSalesRoles(), store.fetchSalesStructure(todayDate())])
  } catch {
    // The page-level error is handled when submitting or through the store.
  }
})
</script>
<template>
  <section class="create-account-page">
    <Toast position="top-right" />

    <header class="topbar">
      <div class="topbar-left">
        <Button
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          rounded
          title="Back to Accounts"
          @click="router.push('/admin/accounts')"
        />

        <div>
          <span class="eyebrow">Account Management</span>
          <h1>Create Account</h1>
          <p>Create a CRM account and assign access from Role Management.</p>
        </div>
      </div>

      <div class="topbar-actions">
        <Button
          label="Cancel"
          severity="secondary"
          outlined
          size="small"
          @click="router.push('/admin/accounts')"
        />
        <Button
          label="Create Account"
          icon="pi pi-check"
          size="small"
          :loading="saving"
          :disabled="!isFormValid || saving"
          @click="handleSubmit"
        />
      </div>
    </header>

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <div class="content-layout">
      <main class="form-column">
        <section class="form-section">
          <header class="section-header">
            <div>
              <h2>User Information</h2>
              <p>Primary sign-in and account identity information.</p>
            </div>
            <span class="section-status required-status">Required</span>
          </header>

          <div class="form-grid">
            <div class="form-field">
              <label>
                Full Name
                <span class="required">*</span>
              </label>
              <InputText
                v-model="form.name"
                placeholder="e.g. Budi Santoso"
                autocomplete="name"
              />
              <small>Use the employee's complete name.</small>
            </div>

            <div class="form-field">
              <label>
                Email
                <span class="required">*</span>
              </label>
              <InputText
                v-model="form.email"
                placeholder="e.g. budi@yummy.test"
                autocomplete="email"
              />
              <small>This email will be used to sign in.</small>
            </div>

            <div class="form-field">
              <label>Employee ID</label>
              <div class="generated-field">
                <InputText
                  v-model="form.employeeId"
                  readonly
                  aria-label="Generated employee ID"
                />
                <Button
                  icon="pi pi-refresh"
                  severity="secondary"
                  outlined
                  title="Generate another employee ID"
                  @click="regenerateEmployeeId"
                />
              </div>
              <small>Generated automatically. Backend generation is recommended for production.</small>
            </div>

            <div class="form-field">
              <label>Phone <span class="optional-badge">Optional</span></label>
              <InputText
                v-model="form.phone"
                placeholder="e.g. 0812-3456-7890"
                autocomplete="tel"
              />
              <small>Primary contact number for this employee.</small>
            </div>

            <div class="form-field">
              <label>
                Temporary Password
                <span class="required">*</span>
              </label>
              <Password
                v-model="form.temporaryPassword"
                toggleMask
                :feedback="true"
                promptLabel="Choose a temporary password"
                placeholder="At least 8 characters"
                inputClass="full-input"
              />
              <small>This temporary password can be changed later from account settings.</small>
            </div>

            <div class="form-field">
              <label>
                Account Type
                <span class="required">*</span>
              </label>
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
              <small>Choose system-level Super Admin access or a Sales Account with an organizational role.</small>
            </div>

            <div v-if="isSalesAccount" class="form-field">
              <label>
                Role
                <span class="required">*</span>
              </label>
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
              <small>Sales Level 1–3 maps to hierarchy Level 2–4 and controls access plus the landing page.</small>
            </div>

            <div v-if="isSalesAccount" class="form-field">
              <label>
                Reports To
                <span class="required">*</span>
              </label>
              <Select
                v-model="form.reportsToUserId"
                :options="reportsToOptions"
                optionLabel="label"
                optionValue="value"
                filter
                :filterFields="['label', 'searchText']"
                :disabled="!selectedOrganizationalRole"
                :placeholder="selectedOrganizationalRole ? `Select Level ${requiredParentLevel} parent` : 'Select a role first'"
                :loading="store.salesStructureLoading"
              >
                <template #option="{ option }">
                  <div class="role-option">
                    <strong>{{ option.item.salesName }}</strong>
                    <span>{{ option.item.salesRole.name }} - Level {{ option.item.salesRole.level }}</span>
                  </div>
                </template>
              </Select>
              <small>Parent must be exactly one hierarchy level above this role.</small>
            </div>
          </div>
        </section>

        <section class="form-section">
          <header class="section-header">
            <div>
              <h2>Role &amp; Access Preview</h2>
              <p>Review the access that will be assigned to this account.</p>
            </div>
            <span class="section-status">Automatic</span>
          </header>

          <div class="access-preview">
            <div class="preview-role">
              <span>Selected Role</span>
              <strong>
                {{ isSalesAccount ? selectedOrganizationalRole?.name || 'No role selected' : 'Super Admin' }}
              </strong>
            </div>

            <div class="access-stat">
              <span>Level</span>
              <strong>{{ isSalesAccount ? selectedOrganizationalRole?.level ?? '—' : 'System' }}</strong>
            </div>

            <div class="access-stat">
              <span>Permissions</span>
              <strong>{{ isSalesAccount ? selectedOrganizationalRole?.permissionCount ?? 0 : 'System' }}</strong>
            </div>

            <div class="access-stat">
              <span>Initial Menu</span>
              <strong>
                {{ isSalesAccount ? landingLabel(selectedOrganizationalRole?.landingPage) : 'admin / dashboard' }}
              </strong>
            </div>

            <p>
              {{
                !isSalesAccount
                  ? 'Super Admin provides system administration access. Creating an additional Super Admin does not create another Level 1 root; the existing primary Super Admin remains the hierarchy root.'
                  :
                selectedOrganizationalRole?.description ||
                'Select a role to preview its access configuration.'
              }}
            </p>
          </div>

          <div class="system-defaults">
            <div>
              <i class="pi pi-check-circle" />
              <span>
                <strong>Status:</strong>
                Active
              </span>
            </div>
            <div>
              <i class="pi pi-key" />
              <span>
                <strong>Password change:</strong>
                Not required for the current demo deployment
              </span>
            </div>
            <div>
              <i class="pi pi-sitemap" />
              <span>
                <strong>Hierarchy assignment:</strong>
                {{
                  isSalesAccount
                    ? 'Created automatically using the selected Role and Reports To'
                    : 'No additional hierarchy root will be created'
                }}
              </span>
            </div>
          </div>
        </section>
      </main>

      <aside class="sidebar-column">
        <section class="preview-card">
          <div class="avatar-preview">
            {{ form.name.trim().slice(0, 1).toUpperCase() || '?' }}
          </div>

          <h3>{{ form.name || 'Employee Name' }}</h3>
          <p>{{ form.email || 'Email address' }}</p>

          <Tag
            :value="isSalesAccount ? selectedOrganizationalRole?.name || 'Role not selected' : 'Super Admin'"
            :severity="isSalesAccount && !selectedOrganizationalRole ? 'secondary' : 'info'"
            rounded
          />

          <div class="preview-divider" />

          <dl>
            <div>
              <dt>Employee ID</dt>
              <dd>{{ form.employeeId || '—' }}</dd>
            </div>
            <div>
              <dt>Role Level</dt>
              <dd>{{ isSalesAccount ? selectedOrganizationalRole?.level ?? '—' : 'System' }}</dd>
            </div>
            <div>
              <dt>Status</dt>
              <dd>Active</dd>
            </div>
            <div>
              <dt>Initial Menu</dt>
              <dd>{{ isSalesAccount ? landingLabel(selectedOrganizationalRole?.landingPage) : 'admin / dashboard' }}</dd>
            </div>
          </dl>
        </section>

        <section class="required-card">
          <div class="required-card-header">
            <div>
              <h3>Required Fields</h3>
              <p>{{ completedRequiredCount }} / {{ requiredFields.length }} complete</p>
            </div>
            <span
              class="completion-count"
              :class="{ complete: isFormValid }"
            >
              {{ Math.round((completedRequiredCount / requiredFields.length) * 100) }}%
            </span>
          </div>

          <div class="required-list">
            <div
              v-for="item in requiredFields"
              :key="item.label"
              class="required-item"
              :class="{ complete: item.complete }"
            >
              <i
                :class="
                  item.complete
                    ? 'pi pi-check-circle'
                    : 'pi pi-circle'
                "
              />
              <span>{{ item.label }}</span>
              <small>{{ item.complete ? 'Complete' : 'Required' }}</small>
            </div>
          </div>
        </section>

        <section class="help-card">
          <i class="pi pi-info-circle" />
          <div>
            <strong>What should Admin fill?</strong>
            <p>
              Fill Full Name, Email, Temporary Password, and Account Type.
              Sales Accounts also require a Role and Reports To so their hierarchy
              assignment can be created automatically. Additional Super Admin accounts
              receive system access without creating another Level 1 root.
              Phone is optional and Employee ID is generated automatically.
            </p>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.create-account-page {
  width: 100%;
  min-width: 0;
  min-height: 100vh;
  overflow-x: hidden;
  background: #f8fafc;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 68px;
  padding: 0.7rem 1rem;
  border-bottom: 1px solid #e5eaf0;
  background: rgba(255, 255, 255, 0.97);
  backdrop-filter: blur(10px);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.topbar-left > div:last-child {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.eyebrow {
  color: #64748b;
  font-size: 0.61rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.topbar h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.12rem;
}

.topbar p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.7rem;
}

.topbar-actions {
  display: flex;
  flex: 0 0 auto;
  gap: 0.5rem;
}

.page-message {
  margin: 0.8rem 1rem 0;
}

.content-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 310px;
  align-items: start;
  gap: 1rem;
  width: min(1120px, calc(100% - 2rem));
  margin: 1rem auto 2rem;
}

.form-column {
  display: grid;
  min-width: 0;
  gap: 1rem;
}

.form-section,
.preview-card,
.required-card,
.help-card {
  border: 1px solid #e3e9f0;
  border-radius: 13px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border-bottom: 1px solid #edf1f6;
}

.section-header h2 {
  margin: 0;
  color: #0f172a;
  font-size: 0.93rem;
}

.section-header p {
  margin: 0.14rem 0 0;
  color: #7c8798;
  font-size: 0.7rem;
}

.section-status,
.optional-badge {
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  padding: 0.14rem 0.42rem;
  border: 1px solid #dbe3ee;
  border-radius: 999px;
  background: #f8fafc;
  color: #64748b;
  font-size: 0.59rem;
  font-weight: 700;
}

.required-status {
  border-color: #fecaca;
  background: #fff7f7;
  color: #b91c1c;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem 1rem;
  padding: 1rem;
}

.form-field {
  display: grid;
  min-width: 0;
  gap: 0.3rem;
}

.form-field label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: #334155;
  font-size: 0.68rem;
  font-weight: 750;
}

.required {
  color: #dc2626;
}

.form-field small {
  color: #94a3b8;
  font-size: 0.62rem;
  line-height: 1.4;
}

.form-field :deep(.p-inputtext),
.form-field :deep(.p-select),
.form-field :deep(.p-password),
.form-field :deep(.p-password-input) {
  width: 100%;
}

.generated-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 40px;
  gap: 0.45rem;
}

.generated-field :deep(.p-inputtext) {
  background: #f8fafc;
  color: #475569;
  font-family: 'SF Mono', Consolas, monospace;
}

.role-option {
  display: grid;
  min-width: 0;
  gap: 0.12rem;
}

.role-option strong {
  color: #0f172a;
  font-size: 0.8rem;
}

.role-option span {
  color: #7c8798;
  font-size: 0.69rem;
}

.access-preview {
  display: grid;
  grid-template-columns: minmax(180px, 1.6fr) repeat(3, minmax(100px, 1fr));
  gap: 0.7rem;
  padding: 1rem;
}

.preview-role,
.access-stat {
  display: grid;
  gap: 0.12rem;
  padding: 0.75rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #f8fafc;
}

.preview-role span,
.access-stat span {
  color: #94a3b8;
  font-size: 0.61rem;
  font-weight: 700;
  text-transform: uppercase;
}

.preview-role strong,
.access-stat strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.79rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.access-preview > p {
  grid-column: 1 / -1;
  margin: 0;
  color: #64748b;
  font-size: 0.71rem;
  line-height: 1.5;
}

.system-defaults {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.55rem;
  padding: 0 1rem 1rem;
}

.system-defaults > div {
  display: flex;
  align-items: flex-start;
  gap: 0.45rem;
  padding: 0.7rem;
  border: 1px solid #d1fae5;
  border-radius: 9px;
  background: #f0fdf4;
  color: #166534;
  font-size: 0.67rem;
  line-height: 1.4;
}

.system-defaults i {
  margin-top: 0.08rem;
  color: #16a34a;
}



.sidebar-column {
  position: sticky;
  top: 84px;
  display: grid;
  gap: 1rem;
  min-width: 0;
}

.preview-card,
.required-card {
  padding: 1rem;
}

.preview-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.avatar-preview {
  display: grid;
  width: 58px;
  height: 58px;
  place-content: center;
  border-radius: 50%;
  background: #fff5f6;
  color: #d14350;
  font-size: 1rem;
  font-weight: 800;
}

.preview-card h3 {
  margin: 0.7rem 0 0;
  color: #0f172a;
  font-size: 0.94rem;
}

.preview-card > p {
  overflow: hidden;
  max-width: 100%;
  margin: 0.18rem 0 0.65rem;
  color: #7c8798;
  font-size: 0.69rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-divider {
  width: 100%;
  height: 1px;
  margin: 0.85rem 0;
  background: #edf1f6;
}

.preview-card dl {
  display: grid;
  width: 100%;
  gap: 0.58rem;
  margin: 0;
}

.preview-card dl > div {
  display: flex;
  justify-content: space-between;
  gap: 0.7rem;
}

.preview-card dt {
  color: #7c8798;
  font-size: 0.67rem;
}

.preview-card dd {
  overflow: hidden;
  margin: 0;
  color: #334155;
  font-size: 0.68rem;
  font-weight: 700;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.required-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.required-card h3 {
  margin: 0;
  color: #0f172a;
  font-size: 0.88rem;
}

.required-card p {
  margin: 0.12rem 0 0;
  color: #94a3b8;
  font-size: 0.64rem;
}

.completion-count {
  padding: 0.2rem 0.42rem;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 0.62rem;
  font-weight: 800;
}

.completion-count.complete {
  background: #dcfce7;
  color: #15803d;
}

.required-list {
  display: grid;
  gap: 0.45rem;
  margin-top: 0.8rem;
}

.required-item {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 0.65rem;
  border: 1px solid #e5eaf0;
  border-radius: 8px;
  background: #f8fafc;
}

.required-item i {
  color: #94a3b8;
  font-size: 0.76rem;
}

.required-item span {
  color: #475569;
  font-size: 0.68rem;
  font-weight: 700;
}

.required-item small {
  color: #94a3b8;
  font-size: 0.59rem;
}

.required-item.complete {
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.required-item.complete i,
.required-item.complete span {
  color: #15803d;
}

.help-card {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  padding: 0.85rem;
  border-color: #f3b9c0;
  background: #fff1f2;
}

.help-card > i {
  margin-top: 0.08rem;
  color: #d14350;
}

.help-card strong {
  color: #922833;
  font-size: 0.73rem;
}

.help-card p {
  margin: 0.18rem 0 0;
  color: #ad3040;
  font-size: 0.66rem;
  line-height: 1.5;
}

@media (max-width: 980px) {
  .content-layout {
    grid-template-columns: 1fr;
  }

  .sidebar-column {
    position: static;
    grid-template-columns: 1fr 1fr;
  }

  .help-card {
    grid-column: 1 / -1;
  }
}

@media (max-width: 760px) {
  .topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .topbar-actions {
    width: 100%;
  }

  .topbar-actions :deep(.p-button) {
    flex: 1;
  }

  .form-grid,
  .access-preview,
  .system-defaults,
  .sidebar-column {
    grid-template-columns: 1fr;
  }

  .content-layout {
    width: min(100% - 1rem, 1120px);
    margin-top: 0.5rem;
  }

  .help-card {
    grid-column: auto;
  }
}
</style>