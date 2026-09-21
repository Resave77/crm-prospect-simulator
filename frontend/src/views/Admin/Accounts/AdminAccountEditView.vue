<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import axios from 'axios'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DatePicker from 'primevue/datepicker'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Toast from 'primevue/toast'
import { useAdminStore } from '../../../stores/admin'
import type { ApiErrorEnvelope } from '../../../types/auth'
import type { SalesRole, SalesStructureItem } from '../../../types/admin'
import ResetPasswordDialog from '../../../components/admin/ResetPasswordDialog.vue'
import { updateUserProfile } from '../../../api/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()
const toast = useToast()
const error = ref('')
const notFound = ref(false)
const saving = ref(false)
const jobInformationOpen = ref(true)
const additionalDetailsOpen = ref(true)
const loaded = ref(false)
const resetPasswordDialogVisible = ref(false)
const timezoneOptions = [
  { label: 'WIB — Asia/Jakarta (UTC+7)', value: 'Asia/Jakarta' },
  { label: 'WITA — Asia/Makassar (UTC+8)', value: 'Asia/Makassar' },
  { label: 'WIT — Asia/Jayapura (UTC+9)', value: 'Asia/Jayapura' },
]
const cityOptions = ['Jakarta Timur, DKI Jakarta', 'Jakarta Barat, DKI Jakarta', 'Jakarta Selatan, DKI Jakarta', 'Bandung, Jawa Barat', 'Surabaya, Jawa Timur', 'Medan, Sumatera Utara', 'Semarang, Jawa Tengah'].map((label) => ({ label, value: label }))
const genderOptions = [{ label: 'Male', value: 'MALE' }, { label: 'Female', value: 'FEMALE' }]

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

const avatarPreview = ref('')
const avatarFileName = ref('')

const form = reactive({
  accountType: 'SALES_ACCOUNT' as 'SUPER_ADMIN' | 'SALES_ACCOUNT',
  temporaryPassword: '',
  employeeId: '',
  name: '',
  email: '',
  phone: '',
  salesRoleId: '',
  reportsToUserId: '',
  timezone: 'Asia/Jakarta', city: '', province: '', district: '', jobTitle: '', positionGrade: '', subDepartment: '', joinDate: null as Date | null, gender: '', dateOfBirth: null as Date | null, phoneNumbers: [] as Array<{ phoneNumber: string; label: string; isPrimary: boolean }>,
})

const accountTypeOptions = [
  { label: 'Sales Account', value: 'SALES_ACCOUNT', description: 'Uses an active Sales Level 1–3 role (hierarchy Level 2–4) and belongs to Sales Structure.' },
  { label: 'Super Admin', value: 'SUPER_ADMIN', description: 'System administrator access. The existing primary Super Admin remains the single Level 1 hierarchy root.' },
]
const isSalesAccount = computed(() => form.accountType === 'SALES_ACCOUNT')
const isPrimarySuperAdminRoot = computed(() =>
  store.selectedUser?.role === 'SUPER_ADMIN'
  && store.selectedUser?.organizationalRole?.level === 1
)
const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const selectedOrganizationalRole = computed(() => {
  const fromList = store.salesRoles.find((role) => role.id === form.salesRoleId)
  const current = store.selectedUser?.organizationalRole
  if (fromList) return fromList
  if (current?.id === form.salesRoleId) return current
  return null
})
const requiredParentLevel = computed(() => selectedOrganizationalRole.value ? selectedOrganizationalRole.value.level - 1 : null)
const reportsToOptions = computed(() =>
  store.salesStructure
    .filter((item: SalesStructureItem) => item.salesRole.level === requiredParentLevel.value && item.userId !== id.value)
    .map((item: SalesStructureItem) => ({
      label: item.salesName,
      value: item.userId,
      searchText: `${item.salesName} ${item.salesRole.name} level ${item.salesRole.level}`,
      item,
    })),
)
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
  if (isSalesAccount.value && !form.reportsToUserId) return false
  return true
})
const requiredFields = computed(() => [
  { label: 'Account Type', complete: Boolean(form.accountType) },
  { label: 'Full Name', complete: Boolean(form.name.trim()) },
  { label: 'Email', complete: emailPattern.test(form.email.trim()) },
  { label: 'Role', complete: !isSalesAccount.value || validOrganizationalSelection.value },
  { label: 'Reports To', complete: !isSalesAccount.value || Boolean(form.reportsToUserId) },
])
const completedRequiredCount = computed(() => requiredFields.value.filter((item) => item.complete).length)
const previewTimezone = computed(() => timezoneOptions.find((item) => item.value === form.timezone)?.label.split(' — ').pop() || form.timezone || 'Not set')
const previewCity = computed(() => [form.city, form.province].filter(Boolean).join(', ') || 'Not set')

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
    form.reportsToUserId = user.reportsToUserId ?? user.managerId ?? ''
    form.timezone = user.timezone || 'Asia/Jakarta'; form.city = user.city || ''; form.province = user.province || ''; form.district = user.district || ''
    form.jobTitle = user.jobTitle || ''; form.positionGrade = user.positionGrade || ''; form.subDepartment = user.subDepartment || ''
    form.joinDate = user.joinDate ? new Date(`${user.joinDate.slice(0, 10)}T00:00:00`) : null; form.gender = user.gender || ''; form.dateOfBirth = user.dateOfBirth ? new Date(`${user.dateOfBirth.slice(0, 10)}T00:00:00`) : null
    form.phoneNumbers = (user.phones || []).map((phone) => ({ phoneNumber: phone.phoneNumber, label: phone.label || '', isPrimary: phone.isPrimary }))
    if (!form.phoneNumbers.length && user.phone) form.phoneNumbers = [{ phoneNumber: user.phone, label: '', isPrimary: true }]
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
  return role.isActive && role.level >= 2 && role.level <= 4
}

function handleAvatarUpload(event: Event) { const file = (event.target as HTMLInputElement).files?.[0]; if (!file || !file.type.startsWith('image/')) return; avatarFileName.value = file.name; const reader = new FileReader(); reader.onload = () => { avatarPreview.value = String(reader.result || '') }; reader.readAsDataURL(file) }
function clearAvatar(event?: Event) { event?.stopPropagation(); avatarPreview.value = ''; avatarFileName.value = '' }

function todayDate() {
  return new Date().toISOString().slice(0, 10)
}

function isoDate(value: Date | string | null) {
  if (!value) return null
  if (typeof value === 'string') return value
  return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}`
}

function removePhone(index: number) {
  if (form.phoneNumbers.length <= 1) {
    form.phoneNumbers[0] = { phoneNumber: '', label: '', isPrimary: true }
    return
  }
  form.phoneNumbers.splice(index, 1)
}

watch(
  () => form.accountType,
  (accountType) => {
    if (accountType === 'SUPER_ADMIN' && !isPrimarySuperAdminRoot.value) {
      form.salesRoleId = ''
      form.reportsToUserId = ''
    }
  },
)

watch(
  () => form.salesRoleId,
  () => {
    const stillValid = reportsToOptions.value.some((option) => option.value === form.reportsToUserId)
    if (!stillValid) form.reportsToUserId = ''
  },
)

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  error.value = ''
  try {
    await updateUserProfile(id.value, { timezone: form.timezone || undefined, city: form.city || undefined, province: form.province || undefined, district: form.district || undefined, jobTitle: form.jobTitle || undefined, positionGrade: form.positionGrade || undefined, subDepartment: form.subDepartment || undefined, joinDate: isoDate(form.joinDate) || undefined, gender: form.gender || undefined, dateOfBirth: isoDate(form.dateOfBirth) || undefined, phones: form.phoneNumbers.filter((phone) => phone.phoneNumber.trim()) })
    const user = await store.updateUser(id.value, {
      employeeId: form.employeeId.trim(),
      name: form.name.trim(),
      email: form.email.trim(),
      phone: form.phone.trim(),
      accountType: form.accountType,
      salesRoleId: isSalesAccount.value
        ? form.salesRoleId || null
        : isPrimarySuperAdminRoot.value
          ? store.selectedUser?.organizationalRole?.id ?? null
          : null,
      managerId: isSalesAccount.value
        ? form.reportsToUserId || null
        : isPrimarySuperAdminRoot.value
          ? store.selectedUser?.reportsToUserId ?? store.selectedUser?.managerId ?? null
          : null,
    })
    // Simpan ulang profile sebagai operasi terakhir agar update account tidak
    // menimpa field profile yang dikelola oleh endpoint profile.
    await updateUserProfile(id.value, {
      timezone: form.timezone || undefined,
      city: form.city || null,
      province: form.province || null,
      district: form.district || null,
      jobTitle: form.jobTitle || null,
      positionGrade: form.positionGrade || null,
      subDepartment: form.subDepartment || null,
      joinDate: isoDate(form.joinDate),
      gender: form.gender || null,
      dateOfBirth: isoDate(form.dateOfBirth),
      phones: form.phoneNumbers.filter((phone) => phone.phoneNumber.trim()),
    })
    await store.fetchUserById(id.value)
    await store.fetchUsers()
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
    try { await Promise.all([store.fetchManagers(), store.fetchSalesRoles(), store.fetchSalesStructure(todayDate())]) } catch { /* optional */ }
  }
})
</script>
<template>
  <section class="admin-page compact-admin-page">
    <Toast position="top-right" />

    <header class="topbar">
      <div class="topbar-left"><Button label="Back to Employee" icon="pi pi-arrow-left" text class="back-employee-button" @click="router.push('/admin/accounts')" /><div><h1>Edit Employee</h1><p>Employee Management &gt; Employee &gt; Edit</p></div></div>
      <div class="topbar-actions"><Button label="Cancel" severity="secondary" outlined size="small" @click="router.push('/admin/accounts')" /><Button label="Update" icon="pi pi-send" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" /></div>
    </header>

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <div class="content-layout">
      <main class="form-column">
        <section class="form-section job-section">
          <button type="button" class="section-header job-information-header" :aria-expanded="jobInformationOpen" aria-label="Toggle job information" @click="jobInformationOpen = !jobInformationOpen">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-[8px]">
                <h2>Job Information</h2>
                <span class="optional-badge">Optional</span>
              </div>
              <p>Employee organizational and job-related data.</p>
            </div>
            <span class="section-toggle"><i class="pi" :class="jobInformationOpen ? 'pi-chevron-up' : 'pi-chevron-down'" /></span>
          </button>
          <div v-if="jobInformationOpen" class="form-grid">
            <div class="form-field"><label>Job Title <span class="optional-badge">Optional</span></label><InputText v-model="form.jobTitle" placeholder="e.g. Finance Supervisor" /><small>Employee's organizational title, for example Finance Supervisor or Warehouse Staff.</small></div>
            <div class="form-field"><label>Join Date <span class="optional-badge">Optional</span></label><DatePicker v-model="form.joinDate" dateFormat="dd/mm/yy" placeholder="dd/mm/yyyy" :maxDate="new Date()" showIcon /><small>YYYY-MM-DD</small></div>
            <div class="form-field"><label>Employee ID <span class="optional-badge">Optional</span></label><InputText v-model="form.employeeId" placeholder="e.g. 44010" /><small>Optional internal employee identifier.</small></div>
            <div class="form-field"><label>Position Grade <span class="optional-badge">Optional</span></label><InputText v-model="form.positionGrade" placeholder="Select position grade" /></div>
            <div class="form-field"><label>Sub Department <span class="optional-badge">Optional</span></label><InputText v-model="form.subDepartment" placeholder="Select sub department" /></div>
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
              <label>Timezone <span class="required">*</span></label>
              <Select v-model="form.timezone" :options="timezoneOptions" optionLabel="label" optionValue="value" />
            </div>

            <div class="form-field">
              <label>City <span class="optional-badge">Optional</span></label>
              <Select v-model="form.city" :options="cityOptions" optionLabel="label" optionValue="value" placeholder="Select city" />
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
              <small>Sales Level 1â€“3 maps to hierarchy Level 2â€“4 and controls access plus the landing page.</small>
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
          <button type="button" class="section-header additional-details-header" :aria-expanded="additionalDetailsOpen" aria-label="Toggle additional details" @click="additionalDetailsOpen = !additionalDetailsOpen">
            <div>
              <div class="flex flex-wrap items-center gap-[8px]">
                <h2>Additional Details</h2>
                <span class="optional-badge">Optional</span>
              </div>
              <p>Additional employee profile data.</p>
            </div>
            <span class="section-toggle"><i class="pi" :class="additionalDetailsOpen ? 'pi-chevron-up' : 'pi-chevron-down'" /></span>
          </button>

          <div v-if="additionalDetailsOpen" class="access-preview">
            <div class="preview-role">
              <span>Selected Role</span>
              <strong>
                {{ isSalesAccount ? selectedOrganizationalRole?.name || 'No role selected' : 'Super Admin' }}
              </strong>
            </div>

            <div class="access-stat">
              <span>Level</span>
              <strong>{{ isSalesAccount ? selectedOrganizationalRole?.level ?? 'â€”' : 'System' }}</strong>
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

          <div class="additional-profile-fields">
            <div class="form-field"><label>Gender <span class="optional-badge">Optional</span></label><Select v-model="form.gender" :options="genderOptions" optionLabel="label" optionValue="value" placeholder="Select gender" /></div>
            <div class="form-field"><label>Date of Birth <span class="optional-badge">Optional</span></label><DatePicker v-model="form.dateOfBirth" dateFormat="dd/mm/yy" placeholder="dd/mm/yyyy" :maxDate="new Date()" showIcon /><small>YYYY-MM-DD</small></div>
            <div class="form-field phone-field"><div class="phone-heading"><div><label>Phone Numbers <span class="optional-badge">(Optional)</span></label><small>You can add more than one phone number.</small></div><Button label="Add Phone" icon="pi pi-plus" outlined size="small" @click="form.phoneNumbers.push({ phoneNumber: '', label: '', isPrimary: false })" /></div><div v-for="(phone, index) in form.phoneNumbers" :key="index" class="phone-row"><InputText v-model="phone.phoneNumber" placeholder="e.g. 0812-3456-7890" /><Button icon="pi pi-trash" text severity="danger" @click="removePhone(index)" /></div></div>
            <div class="form-field avatar-field"><label>Avatar Upload <span class="optional-badge">Optional</span></label><label class="avatar-drop"><input type="file" accept="image/png,image/jpeg,image/webp" hidden @change="handleAvatarUpload" /><button v-if="avatarPreview" type="button" class="avatar-remove" @click="clearAvatar">Remove avatar</button><span><i class="pi pi-upload" /><strong>{{ avatarPreview ? avatarFileName : 'Upload avatar image' }}</strong><small>PNG, JPG, or WEBP. Updates the preview instantly.</small></span></label><small>Upload employee profile photo. JPG or PNG recommended.</small></div>
          </div>
        </section>
      </main>

      <aside class="sidebar-column">
        <section class="preview-card">
          <h2>Employee Preview</h2>
          <div class="avatar-preview">
            <img v-if="avatarPreview" :src="avatarPreview" alt="Employee avatar" />
            <i v-if="!avatarPreview && !form.name.trim()" class="pi pi-user" />
            <span v-else-if="!avatarPreview">{{ form.name.trim().slice(0, 1).toUpperCase() }}</span>
          </div>

          <h3>{{ form.name || 'Employee Name' }}</h3>
          <p>{{ form.email || 'Email address' }}</p>

          <div class="preview-divider" />

          <dl>
            <div><dt>Role</dt><dd>{{ isSalesAccount ? selectedOrganizationalRole?.name || 'Not set' : 'Super Admin' }}</dd></div>
            <div><dt>Department</dt><dd>{{ form.subDepartment || 'Not set' }}</dd></div>
            <div>
              <dt>Status</dt>
              <dd>Active</dd>
            </div>
            <div>
              <dt>Timezone</dt>
              <dd>{{ previewTimezone }}</dd>
            </div>
            <div>
              <dt>Job Title</dt>
              <dd>{{ form.jobTitle || 'Not set' }}</dd>
            </div>
            <div>
              <dt>City</dt>
              <dd>{{ previewCity }}</dd>
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

.user-section .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 20px; padding: 20px 24px 24px; }
.user-section .form-field { min-width: 0; min-height: 88px; gap: 8px; }
.user-section .form-field > label { min-height: 16px; font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 12px; font-weight: 500; line-height: 16px; }
.user-section .form-field > small { min-height: 16px; font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 11px; line-height: 16px; }
.user-section :deep(.p-inputtext), .user-section :deep(.p-select), .user-section :deep(.p-password-input) { box-sizing: border-box; min-height: 48px; }
.user-section :deep(.p-select-label) { line-height: 46px; font-size: 13px; }

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

.preview-card h2 { align-self: flex-start; margin: 0 0 1.1rem; color: #0f172a; font-size: 1rem; font-weight: 700; }

.avatar-preview {
  display: grid;
  width: 84px;
  height: 84px;
  place-content: center;
  border-radius: 50%;
  background: linear-gradient(135deg, #dbeafe 0%, #eff6ff 100%);
  color: #1d4ed8;
  font-size: 1.5rem;
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
.job-information-header { display:flex; width:100%; align-items:flex-start; justify-content:space-between; gap:16px; background:#fff; text-align:left; cursor:pointer; }
.job-information-header:hover { background:#fff; }
.job-information-header h2 { font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:17px; font-weight:700; color:#0f172a; }
.job-information-header p { margin-top:4px; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:12px; color:#64748b; }
.job-information-header .section-toggle { margin-top:2px; width:32px; height:32px; }
.additional-details-header { display:flex; width:100%; align-items:flex-start; justify-content:space-between; gap:16px; background:#fff; text-align:left; cursor:pointer; }
.additional-details-header:hover { background:#fff; }
.additional-details-header h2 { font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:17px; font-weight:700; color:#0f172a; }
.additional-details-header p { margin-top:4px; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:12px; color:#64748b; }
.additional-details-header .section-toggle { margin-top:2px; width:32px; height:32px; }
.phone-field-heading { display:flex; align-items:center; justify-content:space-between; gap:8px; }
.phone-field-heading label { margin:0; }
.remove-phone-button { display:inline-flex; width:24px; height:24px; align-items:center; justify-content:center; padding:0; border:0; border-radius:999px; background:transparent; color:#94a3b8; cursor:pointer; font-size:12px; }
.remove-phone-button:hover { background:#fff1f2; color:#dc2626; }
.phone-field { min-width:0; gap:8px; }
.phone-field :deep(.p-inputtext) { box-sizing:border-box; width:100%; height:48px; border:1px solid #d9e2ec; border-radius:14px; padding:0 14px; background:#fff; color:#0f172a; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:13px; font-weight:500; line-height:48px; box-shadow:0 1px 2px rgba(15,23,42,.04); }
.phone-field :deep(.p-inputtext::placeholder) { color:#94a3b8; opacity:1; }
.phone-field :deep(.p-inputtext:focus) { border-color:#94a3b8; outline:0; box-shadow:0 0 0 2px #dbeafe; }
.phone-field > :deep(.p-button) { width:max-content; height:32px; padding:0 12px; border:1px solid #dbeafe; border-radius:10px; color:#1d4ed8; font-family:Inter,ui-sans-serif,system-ui,sans-serif; font-size:12px; font-weight:600; }
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

.avatar-preview i { color: #64748b; font-size: 30px; }
.avatar-preview img { width: 100%; height: 100%; object-fit: cover; }
.additional-profile-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; padding: 1rem; border-top: 1px solid #eef2f7; }
.additional-profile-fields { padding: 20px 16px 24px; gap: 20px 16px; }
.additional-profile-fields { box-sizing: border-box; width: 100%; padding: 24px 16px 28px; gap: 20px; }
.additional-profile-fields .form-field { gap: 8px; }
.additional-profile-fields .form-field > label, .phone-heading label { font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 12px; font-weight: 500; line-height: 16px; color: #172b4d; }
.additional-profile-fields .form-field small, .phone-heading small { font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 11px; line-height: 16px; color: #526783; }
.additional-profile-fields > .form-field > label { justify-content: space-between; }
.additional-profile-fields .optional-badge { border: 0; background: transparent; padding: 0; color: #7184a0; font-size: 11px; font-weight: 400; }
.additional-profile-fields .p-inputtext, .additional-profile-fields .p-select { min-height: 48px; border-radius: 14px; }
.additional-profile-fields .p-select { width: 100%; }
.additional-profile-fields :deep(.p-select-label) { font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 13px; font-weight: 500; color: #94a3b8; padding: 0 14px; line-height: 46px; }
.additional-profile-fields :deep(.p-select-label.p-placeholder) { color: #94a3b8; }
.additional-profile-fields :deep(.p-select-dropdown) { width: 42px; color: #64748b; }
.additional-profile-fields :deep(.p-select-option) { font-family: Inter,ui-sans-serif,system-ui,sans-serif; font-size: 13px; font-weight: 500; }
.additional-profile-fields .p-datepicker { width: 100%; }
.additional-profile-fields .p-datepicker-input { min-height: 48px; border-radius: 14px 0 0 14px; }
.additional-profile-fields .p-datepicker-dropdown { width: 48px; border-radius: 0 14px 14px 0; background: #f8fafc; }
.phone-heading label { display: flex; align-items: center; gap: 4px; }
.phone-heading .optional-badge { font-size: 11px; }
.avatar-drop { min-height: 120px; }
.additional-section { display: flex; flex-direction: column; }
.additional-section { border: 0; background: transparent; box-shadow: none; gap: 1rem; }
.additional-details-header, .additional-profile-fields { border: 1px solid #e3e9f0; background: #fff; }
.additional-details-header { border-radius: 13px 13px 0 0; }
.additional-profile-fields { border-radius: 0 0 13px 13px; margin-top: -1rem; }
.additional-section > .access-preview, .additional-section > .system-defaults { background: #fff; border: 1px solid #e3e9f0; }
.additional-section > .access-preview { border-radius: 13px 13px 0 0; margin-top: 0; }
.additional-section > .system-defaults { border-radius: 0 0 13px 13px; margin-top: -1rem; }
.additional-details-header { order: 1; }
.additional-profile-fields { order: 2; }
.additional-section > .access-preview { order: 3; }
.additional-section > .system-defaults { order: 4; }
.phone-field, .avatar-field { grid-column: 1 / -1; }
.phone-heading { display:flex; align-items:flex-end; justify-content:space-between; gap:1rem; }
.phone-heading small { display:block; color:#64748b; font-size:.62rem; margin-top:.25rem; }
.phone-row { display:flex; gap:.5rem; margin-top:.5rem; }
.phone-row .p-inputtext { flex:1; }
.phone-row .p-button { width:48px; min-width:48px; height:48px; border:1px solid #fecaca; border-radius:14px; background:#fff; color:#dc2626; }
.phone-row .p-button:hover { background:#fff5f5; border-color:#fca5a5; }
.phone-row .p-button .pi { font-size:15px; }
.additional-profile-fields :deep(.phone-row .p-button) { width:48px; min-width:48px; height:48px; border:1px solid #fecaca; border-radius:14px; background:#fff; color:#dc2626; box-shadow:none; }
.additional-profile-fields :deep(.phone-row .p-button:hover) { background:#fff5f5; border-color:#fca5a5; }
.additional-profile-fields :deep(.phone-row .p-button .pi) { color:#dc2626; font-size:15px; }
.avatar-drop { min-height:120px; display:grid; place-items:center; border:1px dashed #cbd5e1; border-radius:18px; background:#f8fafc; cursor:pointer; text-align:center; }
.avatar-drop { position:relative; font-family:Inter,ui-sans-serif,system-ui,sans-serif; }
.avatar-drop span { display:grid; gap:.5rem; justify-items:center; color:#334155; font-size:.8rem; }
.avatar-drop > span { width:100%; align-items:center; justify-content:center; text-align:center; font-family:Inter,ui-sans-serif,system-ui,sans-serif; }
.avatar-drop > span strong { font-size:13px; font-weight:600; line-height:18px; }
.avatar-drop > span small { font-size:11px; line-height:16px; }
.avatar-drop i { font-size:18px; color:#64748b; }
.avatar-drop span > i { display:grid; width:40px; height:40px; place-items:center; border-radius:50%; background:#fff; box-shadow:0 2px 6px rgba(15,23,42,.08); }
.avatar-drop small { color:#64748b; font-size:.68rem; }
.avatar-drop img { max-height:140px; max-width:100%; border-radius:14px; object-fit:cover; }
.avatar-remove { position:absolute; right:12px; top:12px; z-index:2; padding:7px 10px; border:1px solid #fecaca; border-radius:9px; background:#fff; color:#dc2626; font:600 11px Inter, sans-serif; cursor:pointer; }
@media (max-width:640px) { .additional-profile-fields { grid-template-columns:1fr; } .phone-field, .avatar-field { grid-column:auto; } }

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

.job-section .form-grid {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 18px 20px;
  padding: 24px;
}

.job-section {
  margin: 12px 10px 0;
}

.user-section {
  margin: 12px 10px 0;
}

.job-section .form-field label {
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
  font-size: 12px;
  font-weight: 500;
}

.job-section .optional-badge {
  display: inline-flex;
  align-items: center;
  min-height: 20px;
  padding: 2px 8px;
  border: 1px solid #dbe3ef;
  border-radius: 999px;
  background: #f8fafc;
  color: #64748b;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 10px;
  font-weight: 600;
  line-height: 14px;
}

.job-section .form-field small {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 11px;
  line-height: 16px;
}

@media(max-width:640px){
  .job-section,
  .user-section { margin: 12px 0 0; }
  .job-section .form-grid { grid-template-columns:1fr; padding:16px; }
}
</style>
