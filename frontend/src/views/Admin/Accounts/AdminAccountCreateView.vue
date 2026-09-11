<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DatePicker from 'primevue/datepicker'
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
const jobInformationOpen = ref(false)
const timezoneOptions = [
  { label: 'WIB — Asia/Jakarta (UTC+7)', value: 'Asia/Jakarta' },
  { label: 'WITA — Asia/Makassar (UTC+8)', value: 'Asia/Makassar' },
  { label: 'WIT — Asia/Jayapura (UTC+9)', value: 'Asia/Jayapura' },
]
const genderOptions = [{ label: 'Male', value: 'MALE' }, { label: 'Female', value: 'FEMALE' }]

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
  timezone: 'Asia/Jakarta', city: '', province: '', district: '', jobTitle: '', positionGrade: '', subDepartment: '', joinDate: null as Date | null, gender: '', dateOfBirth: null as Date | null, phoneNumbers: [{ phoneNumber: '', label: '', isPrimary: true }],
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

function isoDate(value: Date | string | null) {
  if (!value) return null
  if (typeof value === 'string') return value
  return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}`
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
      timezone: form.timezone,
      city: form.city || null, province: form.province || null, district: form.district || null,
      jobTitle: form.jobTitle || null, positionGrade: form.positionGrade || null, subDepartment: form.subDepartment || null,
      joinDate: isoDate(form.joinDate) || null, gender: form.gender || null, dateOfBirth: isoDate(form.dateOfBirth) || null,
      phones: form.phoneNumbers.filter((phone) => phone.phoneNumber.trim()),
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
      <div class="topbar-left"><Button label="Back to Employee" icon="pi pi-arrow-left" text class="back-employee-button" @click="router.push('/admin/accounts')" /><div><h1>Create Employee</h1><p>Employee Management &gt; Employee &gt; Create</p></div></div>
      <div class="topbar-actions"><Button label="Cancel" severity="secondary" outlined size="small" @click="router.push('/admin/accounts')" /><Button label="Submit" icon="pi pi-send" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" /></div>
    </header>

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <div class="content-layout">
      <main class="form-column">
        <section class="form-section job-section">
          <header class="section-header"><div><h2>Job Information</h2><p>Optional employee and organization details.</p></div><button type="button" class="section-toggle" :aria-expanded="jobInformationOpen" @click="jobInformationOpen = !jobInformationOpen"><i class="pi" :class="jobInformationOpen ? 'pi-chevron-up' : 'pi-chevron-down'" /></button></header>
          <div v-if="jobInformationOpen" class="form-grid">
            <div class="form-field"><label>Timezone <span class="required">*</span></label><Select v-model="form.timezone" :options="timezoneOptions" optionLabel="label" optionValue="value" /></div>
            <div class="form-field"><label>City <span class="optional-badge">Optional</span></label><InputText v-model="form.city" /></div>
            <div class="form-field"><label>Province <span class="optional-badge">Optional</span></label><InputText v-model="form.province" /></div>
            <div class="form-field"><label>District <span class="optional-badge">Optional</span></label><InputText v-model="form.district" /></div>
            <div class="form-field"><label>Job Title <span class="optional-badge">Optional</span></label><InputText v-model="form.jobTitle" placeholder="e.g. Sales Supervisor" /></div>
            <div class="form-field"><label>Position Grade <span class="optional-badge">Optional</span></label><InputText v-model="form.positionGrade" /></div>
            <div class="form-field"><label>Sub Department <span class="optional-badge">Optional</span></label><InputText v-model="form.subDepartment" /></div>
            <div class="form-field"><label>Join Date <span class="optional-badge">Optional</span></label><DatePicker v-model="form.joinDate" dateFormat="dd/mm/yy" :maxDate="new Date()" showIcon /></div>
            <div class="form-field"><label>Gender <span class="optional-badge">Optional</span></label><Select v-model="form.gender" :options="genderOptions" optionLabel="label" optionValue="value" placeholder="Select gender" /></div>
            <div class="form-field"><label>Date of Birth <span class="optional-badge">Optional</span></label><DatePicker v-model="form.dateOfBirth" dateFormat="dd/mm/yy" :maxDate="new Date()" showIcon /></div>
            <div v-for="(phone, index) in form.phoneNumbers" :key="index" class="form-field"><label>Phone {{ index + 1 }}</label><InputText v-model="phone.phoneNumber" autocomplete="tel" /><Button v-if="index === form.phoneNumbers.length - 1" label="Add Phone" text size="small" @click="form.phoneNumbers.push({ phoneNumber: '', label: '', isPrimary: false })" /></div>
          </div>
        </section>

        <section class="form-section user-section">
          <button type="button" class="section-header user-information-header" aria-expanded="true" aria-label="User information is expanded">
            <div>
              <h2>User Information</h2>
              <p>Primary CRM login and access information.</p>
            </div>
            <span class="section-toggle"><i class="pi pi-chevron-up" /></span>
          </button>

          <div class="form-grid">
            <div class="form-field">
              <label>
                Employee Name
                <span class="required">*</span>
              </label>
              <InputText
                v-model="form.name"
                placeholder="e.g. Michael Carter"
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
                placeholder="e.g. michael.carter@yummydairy.com"
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
              <label>
                Password
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

        <section class="form-section additional-section">
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
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 1rem;
}
.form-column .user-section { order: 1; }
.form-column .job-section { order: 2; }
.form-column .additional-section { order: 3; }

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

/* Match the shared CRM input treatment for the required identity fields. */
.user-section .form-field :deep(.p-inputtext) {
  box-sizing: border-box;
  width: 100%;
  height: 48px;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  padding: 0 14px;
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  font-size: 13px;
  font-weight: 500;
  line-height: 48px;
  outline: none;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  transition: border-color .2s ease, box-shadow .2s ease;
}

.user-section .form-field :deep(.p-inputtext::placeholder) {
  color: #94a3b8;
  opacity: 1;
}

.user-section .form-field :deep(.p-inputtext:focus) {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
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
  background: #fff5f5;
  color: #e63946;
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
  border-color: #f4b3ba;
  background: #fff0f1;
}

.help-card > i {
  margin-top: 0.08rem;
  color: #e63946;
}

.help-card strong {
  color: #a51e2d;
  font-size: 0.73rem;
}

.help-card p {
  margin: 0.18rem 0 0;
  color: #d62839;
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
.create-account-page{min-height:100%;background:#f8fafc;color:#0f172a;font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif}.topbar{min-height:48px;height:48px;padding:0 32px;border-bottom:1px solid #e2e8f0;background:#fff;backdrop-filter:none}.topbar-left{gap:12px}.back-employee-button{height:32px;padding:0 12px 0 0;border-right:1px solid #e2e8f0;border-radius:0;color:#64748b;font-size:16px}.back-employee-button .p-button-icon{font-size:13px}.topbar-left>div:last-child{display:flex;align-items:center;gap:8px}.topbar h1{font-size:17px;font-weight:700;white-space:nowrap}.topbar p{color:#94a3b8;font-size:11px}.topbar-actions{gap:8px}.topbar-actions :deep(.p-button){height:32px;min-height:32px;border-radius:8px;padding:0 14px;font-size:12px}.topbar-actions :deep(.p-button:last-child){background:#94a3b8;border-color:#94a3b8}.content-layout{grid-template-columns:minmax(0,792px) 320px;gap:24px;width:min(1136px,calc(100% - 64px));margin:32px auto}.form-column{gap:16px}.form-section,.preview-card,.required-card,.help-card{border:1px solid #e2e8f0;border-radius:20px;box-shadow:0 6px 18px rgba(15,23,42,.035)}.section-header{padding:20px 24px;border-bottom:1px solid #e8edf3}.section-header h2{font-size:16px;font-weight:700}.section-header p{font-size:12px;color:#64748b}.form-grid{gap:18px 20px;padding:24px}.form-field{gap:6px}.form-field label{font-size:12px;font-weight:500;color:#334155}.form-field small{font-size:11px;color:#64748b}.form-field :deep(.p-inputtext),.form-field :deep(.p-select),.form-field :deep(.p-password-input),.form-field :deep(.p-datepicker-input){height:48px;border:1px solid #d6e0ec;border-radius:14px;padding:0 14px;font-size:13px}.form-field :deep(.p-select-label){padding:0 14px;line-height:46px}.preview-card{padding:28px 24px}.avatar-preview{width:84px;height:84px;margin:0 auto 16px;border-radius:50%;background:#e6f0ff;color:#64748b;font-size:28px}.preview-card h3{text-align:center;font-size:18px;font-weight:700}.preview-card>p{text-align:center;font-size:12px;color:#64748b}.preview-card :deep(.p-tag){display:flex;width:max-content;margin:12px auto;padding:5px 12px;border-radius:999px;font-size:11px}.preview-divider{margin:20px 0;border-color:#e2e8f0}.preview-card dl{gap:12px}.preview-card dt{font-size:12px;color:#64748b}.preview-card dd{font-size:12px;font-weight:500}.required-card,.help-card{padding:20px 24px}.page-message{margin:16px 32px 0}@media(max-width:1000px){.content-layout{grid-template-columns:1fr;width:min(792px,calc(100% - 32px));margin:20px auto}.sidebar-column{display:none}}@media(max-width:640px){.topbar{height:auto;min-height:56px;padding:10px 16px}.topbar-left{align-items:flex-start}.topbar-left>div:last-child{display:grid;gap:2px}.topbar h1{font-size:15px}.topbar p{font-size:10px}.content-layout{width:calc(100% - 24px);margin:12px auto}.form-grid{grid-template-columns:1fr;padding:16px}.section-header{padding:16px}.preview-card{padding:20px}}
.section-toggle{display:grid;width:34px;height:34px;place-items:center;flex:0 0 auto;border:1px solid #d6e0ec;border-radius:50%;background:#f8fafc;color:#64748b;cursor:pointer}.section-toggle:hover{background:#eef5ff;border-color:#bfd3ed;color:#334155}
.user-information-header { display:flex; width:100%; align-items:flex-start; justify-content:space-between; gap:16px; padding:16px 20px; border:0; border-bottom:1px solid #eef2f7; background:#fff; text-align:left; cursor:pointer; }
.user-information-header:hover { background:#fff; }
.user-information-header h2 { font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:17px; font-weight:700; color:#0f172a; }
.user-information-header p { margin-top:4px; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:12px; color:#64748b; }
.user-information-header .section-toggle { margin-top:2px; width:32px; height:32px; }
@media(min-width:640px){.user-information-header{padding:18px 20px}.user-information-header .section-toggle{margin-top:0}}

/* Dropdown fields in User Information match the shared CRM form control. */
.user-section .form-field :deep(.p-select) {
  box-sizing: border-box;
  width: 100%;
  height: 48px;
  min-height: 48px;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  font-size: 13px;
  transition: border-color .2s ease, box-shadow .2s ease;
}

.user-section .form-field :deep(.p-select-label) {
  display: flex;
  height: 46px;
  align-items: center;
  padding: 0 14px;
  color: #334155;
  font-size: 13px;
  font-weight: 500;
  line-height: 46px;
}

.user-section .form-field :deep(.p-select-label.p-placeholder) {
  color: #94a3b8;
}

.user-section .form-field :deep(.p-select-dropdown) {
  width: 42px;
  color: #64748b;
}

.user-section .form-field :deep(.p-select:not(.p-disabled):hover) {
  border-color: #cbd5e1;
}

.user-section .form-field :deep(.p-select.p-focus) {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}

/* Job Information uses the same controls and spacing as User Information. */
.job-section .form-field :deep(.p-inputtext),
.job-section .form-field :deep(.p-select),
.job-section .form-field :deep(.p-password-input),
.job-section .form-field :deep(.p-datepicker-input) {
  box-sizing: border-box;
  width: 100%;
  height: 48px;
  min-height: 48px;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  padding: 0 14px;
  color: #0f172a;
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  font-size: 13px;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  transition: border-color .2s ease, box-shadow .2s ease;
}

.job-section .form-field :deep(.p-select-label) {
  display: flex;
  height: 46px;
  align-items: center;
  padding: 0 14px;
  color: #334155;
  font-size: 13px;
  font-weight: 500;
  line-height: 46px;
}

.job-section .form-field :deep(.p-select-label.p-placeholder),
.job-section .form-field :deep(.p-inputtext::placeholder) {
  color: #94a3b8;
  opacity: 1;
}

.job-section .form-field :deep(.p-select-dropdown) {
  width: 42px;
  color: #64748b;
}

.job-section .form-field :deep(.p-select:not(.p-disabled):hover),
.job-section .form-field :deep(.p-inputtext:focus) {
  border-color: #94a3b8;
}

.job-section .form-field :deep(.p-select.p-focus),
.job-section .form-field :deep(.p-inputtext:focus) {
  box-shadow: 0 0 0 2px #dbeafe;
}
</style>
