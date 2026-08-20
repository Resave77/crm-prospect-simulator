<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useAdminStore } from '../../../stores/admin'
import { useAuthStore } from '../../../stores/auth'
import type { ApiErrorEnvelope } from '../../../types/auth'
import type { AdminUserStatus } from '../../../types/admin'
import ResetPasswordDialog from '../../../components/admin/ResetPasswordDialog.vue'
import { getApiUsageSummary, getApiUsageHistory, getApiActivityHistory, getApiUsageProjectSummary, hideApiUsageHistory } from '../../../api/admin'

const route = useRoute()
const router = useRouter()
const store = useAdminStore()
const auth = useAuthStore()
const toast = useToast()
const error = ref('')
const notFound = ref(false)
const updating = ref(false)
const statusDialogVisible = ref(false)
const statusTarget = ref<{ status: AdminUserStatus; label: string } | null>(null)
const deleteDialogVisible = ref(false)
const resetPasswordDialogVisible = ref(false)
const usageSummary = ref<any[]>([])
const usageHistory = ref<any[]>([])
const activityHistory = ref<any[]>([])
const projectUsage = ref<any[]>([])
const usageSummaryError = ref('')
const projectUsageError = ref('')
const usageHistoryError = ref('')
const activityHistoryError = ref('')
const usageProvider = ref('')
const usageFeature = ref('')
const usagePeriod = ref('THIS_MONTH')
const usageDateFrom = ref('')
const usageDateTo = ref('')
const providerHistoryVisible = ref(10)
const activityHistoryVisible = ref(10)
const activityMode = ref('user')
const historyStatus = ref('ACTIVE')
const hideDialogVisible = ref(false)
const hideTarget = ref<any | null>(null)
const hideReason = ref('')
const hidingHistory = ref(false)
const historyStatusOptions = [{ value: 'ACTIVE', label: 'Aktif' }, { value: 'HIDDEN', label: 'Disembunyikan' }, { value: 'ALL', label: 'Semua' }]
const activityModeOptions = [
  { value: 'user', label: 'Aktivitas Pengguna' },
  { value: 'system', label: 'Aktivitas Sistem' },
  { value: 'all', label: 'Semua Aktivitas' },
]
const normalizeActivityMode = (value: unknown) => ({ user: 'user', system: 'system', all: 'all', MEANINGFUL: 'user', SYSTEM: 'system', TECHNICAL: 'system', ALL: 'all', allTechnical: 'all', technical: 'system', 'request-teknis': 'system' } as Record<string, string>)[String(value)] || 'user'
const expandedProviderKey = ref('')
const expandedActivityKey = ref('')
const periodOptions = [{ value: 'TODAY', label: 'Hari Ini' }, { value: 'LAST_24_HOURS', label: '24 Jam Terakhir' }, { value: 'LAST_7_DAYS', label: '7 Hari Terakhir' }, { value: 'THIS_MONTH', label: 'Bulan Ini' }, { value: 'CUSTOM', label: 'Custom' }]
const googleOperations = [{ value: 'NEARBY_SEARCH', label: 'Nearby Search' }, { value: 'TEXT_SEARCH', label: 'Text Search' }, { value: 'PLACE_DETAILS', label: 'Place Details / Core Detail' }, { value: 'BUSINESS_INFO', label: 'Business Info' }, { value: 'PLACE_PHOTO', label: 'Place Photos' }]
const openAIFeatures = [{ value: 'AI_SUMMARY', label: 'AI Summary' }, { value: 'FIND_MENU', label: 'Find Menu' }, { value: 'MENU_PROFILING', label: 'Menu Profiling' }, { value: 'TANYA_AI', label: 'Tanya AI' }]
const id = computed(() => String(route.params.id))
const operationOptions = computed(() => usageProvider.value === 'GOOGLE_MAPS' ? googleOperations : usageProvider.value === 'OPENAI' ? openAIFeatures : [...googleOperations, ...openAIFeatures])
const jakartaParts = (date = new Date()) => {
  const values = Object.fromEntries(new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Jakarta', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(date).filter((part) => part.type !== 'literal').map((part) => [part.type, Number(part.value)]))
  return { year: values.year, month: values.month, day: values.day }
}
const jakartaMidnight = (date = new Date()) => { const p = jakartaParts(date); return new Date(Date.UTC(p.year, p.month - 1, p.day) - 7 * 60 * 60 * 1000) }
const iso = (date: Date) => date.toISOString()
const periodRange = computed(() => {
  const now = new Date()
  if (usagePeriod.value === 'CUSTOM') return { from: usageDateFrom.value, to: usageDateTo.value }
  if (usagePeriod.value === 'LAST_24_HOURS') return { from: iso(new Date(now.getTime() - 24 * 60 * 60 * 1000)), to: iso(now) }
  if (usagePeriod.value === 'LAST_7_DAYS') return { from: iso(new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)), to: iso(now) }
  const start = jakartaMidnight(now)
  if (usagePeriod.value === 'TODAY') return { from: iso(start), to: iso(new Date(start.getTime() + 24 * 60 * 60 * 1000)) }
  return { from: iso(new Date(Date.UTC(jakartaParts(now).year, jakartaParts(now).month - 1, 1) - 7 * 60 * 60 * 1000)), to: iso(now) }
})
const periodContext = computed(() => ({ TODAY: 'Periode monitoring: hari ini, Asia/Jakarta.', LAST_24_HOURS: 'Periode monitoring: 24 jam terakhir.', LAST_7_DAYS: 'Periode monitoring: 7 hari terakhir.', THIS_MONTH: 'Periode monitoring: bulan ini, Asia/Jakarta.', CUSTOM: 'Periode monitoring: rentang tanggal pilihan.' } as Record<string, string>)[usagePeriod.value] || 'Periode monitoring.')
const usageParams = () => Object.fromEntries(Object.entries({ provider: usageProvider.value, feature: usageProvider.value === 'OPENAI' ? usageFeature.value : '', operation: usageProvider.value === 'GOOGLE_MAPS' ? usageFeature.value : '', dateFrom: periodRange.value.from, dateTo: periodRange.value.to }).filter(([, value]) => value))
const displayedSummary = computed(() => usageSummary.value.filter((item) => !usageFeature.value || item.operation === usageFeature.value))
const contributionFor = (item: any) => {
  const total = projectUsage.value.filter((row) => row.provider === item.provider && row.operation === item.operation).reduce((sum, row) => sum + (row.requests || 0), 0)
  return total > 0 ? `${(((item.requests || 0) / total) * 100).toFixed(1)}%` : '—'
}
const displayedProviderHistory = computed(() => usageHistory.value.slice(0, providerHistoryVisible.value))
const meaningfulActions = new Set(['SEARCH_PROSPECT', 'VIEW_PROSPECT_DETAIL', 'LOAD_BUSINESS_INFO', 'VIEW_PHOTOS', 'FIND_MENU', 'AI_SUMMARY', 'MENU_PROFILING', 'TANYA_AI', 'LOGIN', 'UPDATE_ACCOUNT', 'CHANGE_PASSWORD'])
const meaningfulActivityHistory = computed(() => activityHistory.value.filter((item) => meaningfulActions.has(item.trace?.action)))
const systemActivityHistory = computed(() => activityHistory.value.filter((item) => !meaningfulActions.has(item.trace?.action)))
const apiRelatedActions = new Set(['SEARCH_PROSPECT', 'VIEW_PROSPECT_DETAIL', 'LOAD_BUSINESS_INFO', 'VIEW_PHOTOS', 'FIND_MENU', 'AI_SUMMARY', 'MENU_PROFILING', 'TANYA_AI'])
const apiRelatedActivityHistory = computed(() => activityHistory.value.filter((item) => {
  const trace = item.trace || {}
  return ['GOOGLE_MAPS', 'OPENAI'].includes(trace.provider) || trace.provider_attempted === true ||
    (Number(trace.provider_hit_count) > 0) || apiRelatedActions.has(trace.action)
}))
const activityRows = apiRelatedActivityHistory
const activityModeDescription = computed(() => activityMode.value === 'user'
  ? 'Menampilkan tindakan yang dilakukan langsung oleh pengguna.'
  : activityMode.value === 'system'
    ? 'Menampilkan aktivitas sistem yang dianggap penting untuk audit atau monitoring.'
    : 'Menampilkan seluruh aktivitas pengguna dan sistem yang tercatat untuk audit atau monitoring.')
const consolidatedHistory = computed(() => {
  const matchedProviderIds = new Set<string>()
  const rows = activityRows.value.map((activity: any) => {
    const matchingProviderEvents = activity.requestId ? usageHistory.value.filter((item: any) => item.requestId === activity.requestId) : []
    const providerEvent = matchingProviderEvents[0] || null
    matchingProviderEvents.forEach((item: any) => matchedProviderIds.add(`${item.createdAt}-${item.requestId || item.operation}`))
    const trace = activity.trace || {}
    return {
      ...activity,
      providerEvent,
      trace: {
        ...trace,
        provider: providerEvent?.provider || trace.provider,
        operation: providerEvent?.operation || trace.operation,
        field_mask: providerEvent?.fieldMask || trace.field_mask,
        provider_hit_count: providerEvent ? providerEvent.requestCount : trace.provider_hit_count ?? 0,
        provider_status: providerEvent?.httpStatus || trace.provider_status,
      },
      provider: providerEvent?.provider || trace.provider || '',
      operation: providerEvent?.operation || trace.operation || '',
      apiOrModel: providerEvent?.apiOrModel || trace.model || '',
      fieldMask: providerEvent?.fieldMask || trace.field_mask || '',
      responseStatus: activity.responseStatus || providerEvent?.httpStatus || 0,
      success: providerEvent ? providerEvent.success : activity.responseStatus < 400,
    }
  })
  usageHistory.value.forEach((providerEvent: any) => {
    const key = `${providerEvent.createdAt}-${providerEvent.requestId || providerEvent.operation}`
    if (matchedProviderIds.has(key)) return
    rows.push({
      createdAt: providerEvent.createdAt,
      requestId: providerEvent.requestId,
      endpoint: 'Belum tercatat',
      method: '—',
      responseStatus: providerEvent.httpStatus || 0,
      providerEvent,
      provider: providerEvent.provider,
      operation: providerEvent.operation,
      apiOrModel: providerEvent.apiOrModel,
      fieldMask: providerEvent.fieldMask,
      hiddenAt: providerEvent.hiddenAt,
      hiddenBy: providerEvent.hiddenBy,
      hiddenByName: providerEvent.hiddenByName,
      hideReason: providerEvent.hideReason,
      success: providerEvent.success,
      trace: { action: 'PROVIDER_ONLY', provider: providerEvent.provider, operation: providerEvent.operation, field_mask: providerEvent.fieldMask, provider_hit_count: providerEvent.requestCount, provider_status: providerEvent.httpStatus },
    })
  })
  return rows.sort((a: any, b: any) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
})
const activeActivityHistory = consolidatedHistory
const displayedActivityHistory = computed(() => consolidatedHistory.value.slice(0, activityHistoryVisible.value))
const activityLabel = (action?: string | null) => ({
  SEARCH_PROSPECT: 'Cari Prospek',
  VIEW_PROSPECT_DETAIL: 'Lihat Detail Prospek',
  LOAD_BUSINESS_INFO: 'Muat Informasi Bisnis',
  VIEW_PHOTOS: 'Lihat Foto',
  FIND_MENU: 'Cari Menu',
  AI_SUMMARY: 'Ringkasan AI',
  MENU_PROFILING: 'Profil Menu',
  TANYA_AI: 'Tanya AI',
  PROVIDER_ONLY: 'Riwayat Provider',
  LOGIN: 'Masuk',
  UPDATE_ACCOUNT: 'Perbarui Akun',
  CHANGE_PASSWORD: 'Ubah Password',
}[action || ''] || '—')
const providerLabel = (provider?: string | null) => ({ GOOGLE_MAPS: 'Google Maps', OPENAI: 'OpenAI' }[provider || ''] || provider || 'Internal')
const operationLabel = (operation?: string | null) => ({
  PLACE_DETAILS: 'Place Details',
  NEARBY_SEARCH: 'Nearby Search',
  TEXT_SEARCH: 'Text Search',
  BUSINESS_INFO: 'Business Info',
  PLACE_PHOTO: 'Place Photos',
  RESPONSES: 'Responses',
}[operation || ''] || operation || '—')
const systemActivityLabel = (endpoint?: string) => {
  if (!endpoint) return 'Aktivitas Sistem'
  if (endpoint.includes('/auth/me')) return 'Validasi Sesi'
  if (endpoint.includes('/auth/refresh')) return 'Perbarui Sesi'
  if (endpoint.includes('/sales/customers')) return 'Muat Data Customer'
  if (endpoint.includes('/prospects')) return 'Muat Data Prospek'
  if (endpoint.includes('/admin/accounts')) return 'Muat Data Akun'
  return 'Aktivitas Sistem'
}
const technicalActivityLabel = (item: any) => item.trace?.action === 'PROVIDER_ONLY' ? 'Riwayat Provider' : item.trace?.action ? activityLabel(item.trace.action) : systemActivityLabel(item.endpoint)
const activityProviderOperation = (item: any) => item.trace?.provider ? `${providerLabel(item.trace.provider)} / ${operationLabel(item.trace.operation)}` : 'Internal'
const providerDetailItems = (item: any) => {
  const details = [`Request ID: ${item.requestId || '—'}`]
  if (item.feature) details.push(`Feature: ${item.feature}`)
  if (item.httpStatus) details.push(`HTTP Provider: ${item.httpStatus}`)
  if (item.provider === 'OPENAI') {
    if (item.apiOrModel) details.push(`Model: ${item.apiOrModel}`)
    if (item.inputTokens != null) details.push(`Input Tokens: ${item.inputTokens}`)
    if (item.outputTokens != null) details.push(`Output Tokens: ${item.outputTokens}`)
    if (item.totalTokens != null) details.push(`Total Tokens: ${item.totalTokens}`)
  }
  if (item.estimatedCost) details.push(`Cost: ${item.estimatedCost}`)
  return details
}
const googleUsdIdrRate = Number(import.meta.env.VITE_GOOGLE_COST_USD_IDR_RATE || 0)
const formatGoogleIdr = (micros: number) => micros === 0 ? 'Rp0' : googleUsdIdrRate > 0 ? `Rp${Math.round(micros / 1e6 * googleUsdIdrRate).toLocaleString('id-ID')}` : 'Kurs IDR belum dikonfigurasi'
const costLabel = (item: { provider?: string; estimatedCost?: number; estimatedCostMicros?: number; estimatedPayableCostMicros?: number; costState?: string; costStatus?: string }) => item.costState === 'BILLING_UNKNOWN' || item.costStatus === 'BILLING_UNKNOWN' ? 'Belum dapat dipastikan' : item.costState === 'SKU_UNKNOWN' ? 'SKU belum teridentifikasi' : item.costState === 'VERIFIED_ESTIMATE' && item.provider === 'GOOGLE_MAPS' ? formatGoogleIdr(item.estimatedPayableCostMicros || 0) : item.costState === 'VERIFIED_ESTIMATE' ? `$${((item.estimatedCostMicros ?? item.estimatedPayableCostMicros ?? 0) / 1e6).toFixed(4)}` : 'Belum dikonfigurasi'
const providerRowKey = (item: any) => `${item.createdAt}-${item.requestId || item.operation}`
const activityRowKey = (item: any) => `${item.createdAt}-${item.requestId || item.endpoint}`
const toggleProviderDetail = (item: any) => { const key = providerRowKey(item); expandedProviderKey.value = expandedProviderKey.value === key ? '' : key }
const toggleActivityDetail = (item: any) => { const key = activityRowKey(item); expandedActivityKey.value = expandedActivityKey.value === key ? '' : key }
const openHideHistory = (item: any) => { hideTarget.value = item; hideReason.value = ''; hideDialogVisible.value = true }
const executeHideHistory = async () => {
  if (!hideTarget.value || !id.value || !hideTarget.value.requestId) return
  hidingHistory.value = true
  try {
    await hideApiUsageHistory(id.value, hideTarget.value.requestId, hideReason.value.trim())
    hideDialogVisible.value = false
    await reloadUsage()
  } finally { hidingHistory.value = false }
}
const resetHistoryPagination = () => { providerHistoryVisible.value = 10; activityHistoryVisible.value = 10 }
const resetUsageFilters = () => { usageProvider.value = ''; usageFeature.value = ''; usagePeriod.value = 'THIS_MONTH'; usageDateFrom.value = ''; usageDateTo.value = ''; resetHistoryPagination(); reloadUsage() }
const reloadMoreProviderHistory = () => { providerHistoryVisible.value += 10 }
const reloadMoreActivityHistory = () => { activityHistoryVisible.value += 10 }
const reloadUsage = async () => {
  if (!id.value) return
  const params = usageParams()
  const historyParams = { ...params, historyStatus: historyStatus.value }
  const load = async (request: () => Promise<unknown>, target: { value: any[] }, failure: { value: string }, message: string) => {
    failure.value = ''
    try { target.value = (await request()) as any[] } catch { target.value = []; failure.value = message }
  }
  await Promise.all([
    load(() => getApiUsageSummary(id.value, params), usageSummary, usageSummaryError, 'Monthly summary is unavailable.'),
    load(() => getApiUsageHistory(id.value, historyParams), usageHistory, usageHistoryError, 'Provider usage history is unavailable.'),
    load(() => getApiActivityHistory(id.value, historyParams), activityHistory, activityHistoryError, 'Riwayat penggunaan API belum dapat dimuat.'),
		load(async () => (await getApiUsageProjectSummary(params)).groups || [], projectUsage, projectUsageError, 'Project usage is unavailable.'),
  ])
}
watch(usageProvider, () => {
  if (usageFeature.value && !operationOptions.value.some((option) => option.value === usageFeature.value)) usageFeature.value = ''
})
watch([usageProvider, usageFeature, usagePeriod, usageDateFrom, usageDateTo], () => { resetHistoryPagination(); reloadUsage() })
watch(historyStatus, () => { resetHistoryPagination(); reloadUsage() })

const user = computed(() => (store.selectedUser?.id === id.value ? store.selectedUser : null))
const isSelf = computed(() => user.value?.id === auth.user?.id)
const isProtectedSuperAdmin = computed(() => user.value?.email === 'admin@yummy.test' || user.value?.fullName === 'Yummy Super Admin')
const organizationalRole = computed(() => user.value?.organizationalRole ?? null)

function isNotFoundError(e: unknown) {
  return axios.isAxiosError<ApiErrorEnvelope>(e)
    && (e.response?.status === 404 || e.response?.data?.error?.code === 'USER_NOT_FOUND')
}

async function executeDelete() {
  if (!user.value) return
  updating.value = true
  error.value = ''
  try {
    await store.deleteUser(id.value)
    deleteDialogVisible.value = false
    await router.push('/admin/accounts')
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    updating.value = false
  }
}

async function load() {
  error.value = ''
  notFound.value = false
  try {
    await store.fetchUserById(id.value)
    await reloadUsage()
  } catch (e) {
    notFound.value = isNotFoundError(e)
    error.value = store.errorMessage(e)
  }
}

function formatDateTime(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function landingLabel(path?: string | null) {
  if (!path) return '—'
  return path.split('/').filter(Boolean).map((part) => part.replace(/-/g, ' ')).join(' / ') || path
}

function confirmStatus(status: AdminUserStatus) {
  statusTarget.value = { status, label: status === 'INACTIVE' ? 'Deactivate' : 'Activate' }
  statusDialogVisible.value = true
}

async function executeStatus() {
  if (!statusTarget.value || !user.value) return
  updating.value = true
  error.value = ''
  try {
    await store.updateStatus(id.value, statusTarget.value.status)
    statusDialogVisible.value = false
    const label = statusTarget.value.status === 'INACTIVE' ? 'Deactivated' : 'Activated'
    toast.add({ severity: 'success', summary: label, detail: `${user.value.fullName} is now ${statusTarget.value.status}.`, life: 3000 })
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    updating.value = false
  }
}

watch(activityMode, (value) => {
  const normalized = normalizeActivityMode(value)
  if (normalized !== value) activityMode.value = normalized
})
onMounted(() => { load() })
</script>

<template>
  <section class="admin-page compact-admin-page">
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
      <div v-if="store.detailLoading && !user" class="state-box">
        <i class="pi pi-spin pi-spinner state-icon" />
        <span>Loading account details...</span>
      </div>

      <template v-if="user">
        <!-- PAGE HEADER -->
        <header class="page-heading">
          <div class="compact-heading-main">
            <Button class="header-back" icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push('/admin/accounts')" title="Back to account list" />
          <div class="page-title-wrapper">
            <span class="eyebrow">Account Detail</span>
            <h1>{{ user.fullName }}</h1>
            <div class="subtitle-row">
              <code class="code-tag code-blue">{{ user.employeeId || '—' }}</code>
              <span class="muted">&mdash;</span>
              <span class="muted">{{ user.email }}</span>
            </div>
          </div>
          </div>
          <div class="page-heading-actions">
            <Button label="Reset Password" icon="pi pi-key" severity="warning" outlined size="small" @click="resetPasswordDialogVisible = true" />
            <Button label="Edit Account" icon="pi pi-pencil" size="small" @click="router.push(`/admin/accounts/${id}/edit`)" />
            <Button label="Delete" icon="pi pi-trash" severity="danger" outlined size="small" :disabled="isSelf || isProtectedSuperAdmin || updating" @click="deleteDialogVisible = true" />
          </div>
        </header>

        <!-- DETAIL CARDS -->
        <div class="account-info-columns">
          <div class="account-info-stack left-stack">
            <!-- IDENTITY -->
            <div class="detail-card">
              <div class="detail-card-header">
                <div class="detail-card-icon si-blue"><i class="pi pi-id-card" /></div>
                <div>
                  <h3>Account Identity</h3>
                  <p>Core identification for this user.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Employee ID</span>
                  <span class="detail-value"><code class="code-tag code-blue">{{ user.employeeId || '—' }}</code></span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Full Name</span>
                  <span class="detail-value">{{ user.fullName || '—' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Email</span>
                  <span class="detail-value">{{ user.email || '—' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Phone</span>
                  <span class="detail-value">{{ user.phone || '—' }}</span>
                </div>
              </div>
            </div>

            <div class="detail-card"><div class="detail-card-header"><div class="detail-card-icon si-blue"><i class="pi pi-map-marker" /></div><div><h3>User Information</h3><p>Identity, location, timezone, and contact information.</p></div></div><div class="detail-rows">
              <div class="detail-row"><span class="detail-label">Timezone</span><span class="detail-value">{{ user.timezone || '—' }}</span></div>
              <div class="detail-row"><span class="detail-label">Location</span><span class="detail-value">{{ [user.city, user.province, user.district].filter(Boolean).join(', ') || '—' }}</span></div>
              <div class="detail-row"><span class="detail-label">Job</span><span class="detail-value">{{ [user.jobTitle, user.positionGrade, user.subDepartment].filter(Boolean).join(' · ') || '—' }}</span></div>
              <div class="detail-row"><span class="detail-label">Join Date</span><span class="detail-value">{{ user.joinDate?.slice(0, 10) || '—' }}</span></div>
              <div class="detail-row"><span class="detail-label">Gender / Birth Date</span><span class="detail-value">{{ [user.gender, user.dateOfBirth?.slice(0, 10)].filter(Boolean).join(' · ') || '—' }}</span></div>
              <div class="detail-row"><span class="detail-label">Phone Numbers</span><span class="detail-value">{{ user.phones?.map((phone) => phone.phoneNumber).join(', ') || user.phone || '—' }}</span></div>
            </div></div>

            <!-- ROLE -->
            <div class="detail-card">
              <div class="detail-card-header">
                <div class="detail-card-icon si-violet"><i class="pi pi-sitemap" /></div>
                <div>
                  <h3>Role</h3>
                  <p>Current account role from Role Management.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Role Name</span>
                  <span class="detail-value">{{ organizationalRole?.name || '—' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Level</span>
                  <span class="detail-value">{{ organizationalRole ? `Level ${organizationalRole.level}` : '—' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Landing Page</span>
                  <span class="detail-value">{{ landingLabel(organizationalRole?.landingPage) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Permission Count</span>
                  <span class="detail-value">{{ organizationalRole?.permissionCount ?? 0 }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Description</span>
                  <span class="detail-value">
                    {{ organizationalRole?.description || '—' }}
                  </span>
                </div>
              </div>
            </div>

          </div>
          <div class="account-info-stack right-stack">
            <!-- REPORTING STRUCTURE -->
            <div class="detail-card">
              <div class="detail-card-header">
                <div class="detail-card-icon si-green"><i class="pi pi-users" /></div>
                <div>
                  <h3>Reporting Structure</h3>
                  <p>Current direct manager.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Reports To</span>
                  <span class="detail-value">{{ user.managerName || '—' }}</span>
                </div>
              </div>
            </div>

            <!-- SECURITY -->
            <div class="detail-card">
              <div class="detail-card-header">
                <div class="detail-card-icon si-green"><i class="pi pi-shield" /></div>
                <div>
                  <h3>Security</h3>
                  <p>Account status and sign-in policy.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Status</span>
                  <span class="detail-value">
                    <Tag :value="user.status" :severity="user.status === 'ACTIVE' ? 'success' : 'secondary'" />
                  </span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Must Change Password</span>
                  <span class="detail-value">
                    <Tag :value="user.mustChangePassword ? 'Yes' : 'No'" :severity="user.mustChangePassword ? 'warn' : 'secondary'" :icon="user.mustChangePassword ? 'pi pi-key' : ''" />
                  </span>
                </div>
              </div>
              <div class="status-actions">
                <Button
                  v-if="user.status === 'ACTIVE'"
                  label="Deactivate"
                  icon="pi pi-user-minus"
                  severity="danger"
                  outlined
                  size="small"
                  :disabled="isSelf || isProtectedSuperAdmin || updating"
                  :title="isProtectedSuperAdmin ? 'Yummy Super Admin is protected' : isSelf ? 'You cannot deactivate your own account' : 'Deactivate account'"
                  @click="confirmStatus('INACTIVE')"
                />
                <Button
                  v-else
                  label="Activate"
                  icon="pi pi-user-plus"
                  severity="success"
                  outlined
                  size="small"
                  :disabled="updating"
                  title="Activate account"
                  @click="confirmStatus('ACTIVE')"
                />
              </div>
            </div>

            <!-- AUDIT -->
            <div class="detail-card">
              <div class="detail-card-header">
                <div class="detail-card-icon si-amber"><i class="pi pi-history" /></div>
                <div>
                  <h3>Audit</h3>
                  <p>Creation and last update metadata.</p>
                </div>
              </div>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">Created At</span>
                  <span class="detail-value">{{ formatDateTime(user.createdAt) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Updated At</span>
                  <span class="detail-value">{{ formatDateTime(user.updatedAt) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Created By</span>
                  <span class="detail-value">
                    <code v-if="user.createdBy" class="code-tag">{{ user.createdBy }}</code>
                    <span v-else>—</span>
                  </span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Updated By</span>
                  <span class="detail-value">
                    <code v-if="user.updatedBy" class="code-tag">{{ user.updatedBy }}</code>
                    <span v-else>—</span>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- STATUS CONFIRMATION DIALOG -->
        <Dialog v-model:visible="statusDialogVisible" :header="statusTarget?.label ?? ''" modal :style="{ width: '400px' }">
          <p v-if="statusTarget?.status === 'INACTIVE'">
            Are you sure you want to deactivate <strong>{{ user?.fullName }}</strong>? The user will no longer be able to sign in.
          </p>
          <p v-else>
            Are you sure you want to activate <strong>{{ user?.fullName }}</strong>? The user will regain sign-in access.
          </p>
          <template #footer>
            <Button label="Cancel" severity="secondary" text @click="statusDialogVisible = false" :disabled="updating" />
            <Button
              :label="statusTarget?.label ?? ''"
              :severity="statusTarget?.status === 'INACTIVE' ? 'danger' : 'success'"
              :icon="statusTarget?.status === 'INACTIVE' ? 'pi pi-user-minus' : 'pi pi-user-plus'"
              :loading="updating"
              @click="executeStatus"
            />
          </template>
        </Dialog>

        <Dialog v-model:visible="deleteDialogVisible" header="Delete Account" modal :style="{ width: '400px' }">
          <p>Delete <strong>{{ user?.fullName }}</strong>? This will fail if the account is still referenced by existing records.</p>
          <template #footer>
            <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="updating" />
            <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="updating" @click="executeDelete" />
          </template>
        </Dialog>

        <section class="detail-card usage-card">

          <div class="detail-card-header"><div class="detail-card-icon si-violet"><i class="pi pi-chart-bar" /></div><div><h3>Penggunaan API Pengguna</h3><p>Monitoring penggunaan layanan eksternal oleh pengguna ini.</p></div></div>
          <div class="usage-filters"><label>Provider<select v-model="usageProvider"><option value="">Semua Provider</option><option value="GOOGLE_MAPS">Google Maps</option><option value="OPENAI">OpenAI</option></select></label><label>Operasi / Fitur<select v-model="usageFeature"><option value="">Semua Operasi / Fitur</option><option v-for="option in operationOptions" :key="option.value" :value="option.value">{{ option.label }}</option></select></label><label>Periode<select v-model="usagePeriod"><option v-for="option in periodOptions" :key="option.value" :value="option.value">{{ option.label }}</option></select></label><label v-if="usagePeriod === 'CUSTOM'">Tanggal Mulai<input v-model="usageDateFrom" type="date" /></label><label v-if="usagePeriod === 'CUSTOM'">Tanggal Akhir<input v-model="usageDateTo" type="date" /></label><button type="button" class="usage-clear" @click="resetUsageFilters">Reset Filter</button></div>
          <div v-if="usageSummaryError" class="usage-inline-error">{{ usageSummaryError }} <button type="button" @click="reloadUsage">Retry</button></div>
          <div v-else-if="!usageSummary.length" class="usage-empty">No API usage recorded yet.</div>
          <div v-else class="usage-table-wrap"><p class="usage-helper">Total request mencakup request provider yang berhasil maupun gagal. Free tier berlaku pada level project, bukan per pengguna.</p><p class="usage-period-context">{{ periodContext }}</p><table class="usage-table"><thead><tr><th>API / Operasi</th><th>Total Request</th><th>Berhasil</th><th>Gagal</th><th>Kontribusi Project</th><th title="Cost dihitung dari usage internal berdasarkan konfigurasi pricing yang diverifikasi. Tagihan resmi tetap mengikuti Google Cloud Billing.">Cost</th></tr></thead><tbody><tr v-for="item in displayedSummary" :key="`${item.provider}-${item.operation}`"><td><strong>{{ item.provider === 'GOOGLE_MAPS' ? 'Google Maps' : item.provider }}</strong><small>{{ item.operation.replaceAll('_', ' ') }}</small></td><td>{{ item.requests || 0 }}</td><td class="usage-success-number">{{ item.success || 0 }}</td><td class="usage-failed-number">{{ item.failed || 0 }}</td><td>{{ contributionFor(item) }}</td><td>{{ costLabel(item) }}</td></tr></tbody></table></div>
          <section class="usage-subsection"><h4>Riwayat Penggunaan API</h4><p class="usage-helper">Menampilkan riwayat request ke layanan eksternal seperti Google Maps atau OpenAI yang tercatat atas penggunaan akun ini.</p><div v-if="usageHistoryError" class="usage-inline-error">{{ usageHistoryError }} <button type="button" @click="reloadUsage">Retry</button></div><div v-else-if="!usageHistory.length" class="usage-empty">{{ usageProvider === 'OPENAI' ? 'Belum ada penggunaan OpenAI yang tercatat pada periode ini.' : 'Belum ada penggunaan provider pada periode ini.' }}</div><div v-else class="usage-history"><table class="usage-history-table"><colgroup><col class="provider-col-time" /><col class="provider-col-provider" /><col class="provider-col-operation" /><col class="provider-col-sku" /><col class="provider-col-status" /><col class="provider-col-detail" /></colgroup><thead><tr><th>Waktu</th><th>Provider</th><th>Operasi</th><th>SKU / Model</th><th>Status Provider</th><th>Detail</th></tr></thead><tbody><template v-for="item in displayedProviderHistory" :key="providerRowKey(item)"><tr class="provider-main-row"><td>{{ formatDateTime(item.createdAt) }}</td><td><strong>{{ providerLabel(item.provider) }}</strong></td><td>{{ operationLabel(item.operation) }}</td><td>{{ item.fieldMask || item.apiOrModel || '—' }}</td><td><span class="status-badge" :class="item.success ? 'status-success' : 'status-failed'">{{ item.success ? 'Berhasil' : 'Gagal' }}</span><small>HTTP Provider: {{ item.httpStatus || '—' }}</small></td><td><button type="button" class="detail-toggle" :aria-expanded="expandedProviderKey === providerRowKey(item)" @click="toggleProviderDetail(item)">{{ expandedProviderKey === providerRowKey(item) ? 'Tutup' : 'Lihat' }}</button></td></tr><tr v-if="expandedProviderKey === providerRowKey(item)" class="detail-row-expanded"><td colspan="6"><div class="usage-detail-panel"><h5>Detail Provider</h5><div class="usage-detail-group"><span><b>Provider</b>{{ providerLabel(item.provider) }}</span><span><b>Operasi</b>{{ operationLabel(item.operation) }}</span><span><b>SKU / Model</b>{{ item.fieldMask || item.apiOrModel || '—' }}</span><span><b>Status Provider</b>{{ item.success ? 'Berhasil' : 'Gagal' }}</span><span v-if="item.httpStatus"><b>HTTP Provider</b>{{ item.httpStatus }}</span><span v-if="item.feature"><b>Feature</b>{{ item.feature }}</span><span v-if="item.requestId"><b>Request ID</b><code>{{ item.requestId }}</code></span><span v-if="item.credentialAlias"><b>Credential Alias</b>{{ item.credentialAlias }}</span><span v-if="item.environment"><b>Environment</b>{{ item.environment }}</span></div><div v-if="item.provider === 'OPENAI'" class="usage-detail-secondary"><h6>Token Usage</h6><div class="usage-detail-group"><span v-if="item.inputTokens != null"><b>Input Tokens</b>{{ item.inputTokens }}</span><span v-if="item.cachedTokens != null"><b>Cached Tokens</b>{{ item.cachedTokens }}</span><span v-if="item.outputTokens != null"><b>Output Tokens</b>{{ item.outputTokens }}</span><span v-if="item.totalTokens != null"><b>Total Tokens</b>{{ item.totalTokens }}</span><span v-if="item.estimatedCost"><b>Estimated Cost</b>{{ item.estimatedCost }}</span></div></div></div></td></tr></template></tbody></table><button v-if="providerHistoryVisible < usageHistory.length" type="button" class="usage-more" @click="reloadMoreProviderHistory">Tampilkan 10 Lagi</button></div></section>
          <section class="usage-subsection"><div class="usage-section-heading"><div><h4>Riwayat Aktivitas</h4><span class="usage-helper">Menampilkan aktivitas pengguna yang terkait dengan penggunaan Google Maps atau OpenAI. Cache HIT tetap dicatat meskipun tidak mengirim request baru ke provider.</span><div class="history-controls"><label>Status<select v-model="historyStatus"><option v-for="option in historyStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</option></select></label></div></div></div><div v-if="activityHistoryError" class="usage-inline-error">{{ activityHistoryError }} <button type="button" @click="reloadUsage">Retry</button></div><div v-else-if="!activeActivityHistory.length" class="usage-empty"><strong>Belum ada riwayat pada filter ini.</strong><small>Riwayat aktivitas hanya mencatat tindakan yang dilakukan langsung oleh pengguna ini. Aktivitas yang dilakukan Admin saat melihat atau mengelola akun ini akan tercatat pada riwayat aktivitas Admin.</small><small v-if="usageHistory.length" class="usage-helper">Sebagian riwayat penggunaan API dapat berasal dari data sebelum fitur Activity History diterapkan.</small></div><div v-else class="activity-table-wrap"><table class="activity-table"><thead><tr><th>Waktu</th><th>Aktivitas</th><th>Provider / Operasi</th><th>Cache</th><th>Request Provider</th><th>Status Aplikasi</th><th>Detail</th></tr></thead><tbody><template v-for="item in displayedActivityHistory" :key="activityRowKey(item)"><tr class="activity-main-row"><td>{{ formatDateTime(item.createdAt) }}</td><td>{{ technicalActivityLabel(item) }}</td><td>{{ activityProviderOperation(item) }}</td><td><span class="trace-badge" :class="item.trace?.cache_status === 'HIT' ? 'trace-hit' : item.trace?.cache_status === 'MISS' ? 'trace-miss' : ''">{{ item.trace?.cache_status || '—' }}</span></td><td>{{ item.trace?.provider_hit_count ?? '—' }}</td><td><span class="status-badge" :class="item.responseStatus >= 400 ? 'status-failed' : 'status-success'">{{ item.responseStatus >= 400 ? 'Gagal' : 'Berhasil' }}</span><small>HTTP Aplikasi: {{ item.responseStatus || '—' }}</small></td><td><div class="history-row-actions"><button type="button" class="detail-toggle" :aria-expanded="expandedActivityKey === activityRowKey(item)" @click="toggleActivityDetail(item)">{{ expandedActivityKey === activityRowKey(item) ? 'Tutup' : 'Lihat' }}</button><button v-if="item.requestId && !item.hiddenAt" type="button" class="history-hide-button" aria-label="Hapus dari tampilan" title="Hapus dari tampilan" @click="openHideHistory(item)"><i class="pi pi-trash" /></button></div></td></tr><tr v-if="expandedActivityKey === activityRowKey(item)" class="detail-row-expanded"><td colspan="7"><div class="usage-detail-panel"><h5>Detail Aktivitas</h5><div v-if="item.hiddenAt" class="hidden-history-meta">Disembunyikan{{ item.hiddenByName ? ` oleh ${item.hiddenByName}` : '' }}{{ item.hideReason ? ` — ${item.hideReason}` : '' }}</div><div class="usage-detail-group"><span><b>Aktivitas</b>{{ technicalActivityLabel(item) }}</span><span><b>Provider / Operasi</b>{{ activityProviderOperation(item) }}</span><span><b>Status Aplikasi</b>{{ item.responseStatus >= 400 ? 'Gagal' : 'Berhasil' }}</span><span><b>HTTP Aplikasi</b>{{ item.responseStatus || '—' }}</span><span><b>Cache</b>{{ item.trace?.cache_status || '—' }}</span><span><b>Request Provider</b>{{ item.trace?.provider_hit_count ?? '—' }}</span></div><div class="usage-detail-secondary"><h6>Informasi Teknis</h6><div class="usage-detail-group"><span><b>Endpoint</b><code>{{ item.endpoint }}</code></span><span><b>Method</b>{{ item.method }}</span><span><b>Request ID</b><code>{{ item.requestId || '—' }}</code></span><span v-if="item.trace?.provider_status"><b>HTTP Provider</b>{{ item.trace.provider_status }}</span><span v-if="item.trace?.field_mask"><b>Field Mask</b><code>{{ item.trace.field_mask }}</code></span><span v-if="item.trace?.credential_alias"><b>Credential Alias</b>{{ item.trace.credential_alias }}</span><span v-if="item.trace?.environment"><b>Environment</b>{{ item.trace.environment }}</span><span v-if="item.trace?.model"><b>Model</b>{{ item.trace.model }}</span></div></div></div></td></tr></template></tbody></table><button v-if="activityHistoryVisible < activeActivityHistory.length" type="button" class="usage-more" @click="reloadMoreActivityHistory">Tampilkan 10 Lagi</button></div></section>

        </section>
        <Dialog v-model:visible="hideDialogVisible" header="Hapus riwayat dari tampilan?" modal :style="{ width: '420px' }">
          <p>Riwayat ini akan disembunyikan dari tampilan, tetapi tetap dipertahankan untuk audit dan rekonsiliasi.</p>
          <label class="hide-reason-field">Alasan (opsional)<textarea v-model="hideReason" rows="3" maxlength="500"></textarea></label>
          <template #footer>
            <Button label="Batal" severity="secondary" text @click="hideDialogVisible = false" :disabled="hidingHistory" />
            <Button label="Hapus dari tampilan" severity="danger" icon="pi pi-trash" :loading="hidingHistory" @click="executeHideHistory" />
          </template>
        </Dialog>
        <ResetPasswordDialog v-model:visible="resetPasswordDialogVisible" :user="user" @reset-success="load" />
      </template>
    </template>
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

/* ── DETAIL GRID ──────────────────────────────────────────────────── */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.6rem;
  align-items: start;
}
.account-info-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.8rem;
  align-items: start;
}
.account-info-stack {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
  min-width: 0;
}
.detail-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

/* ── DETAIL CARDS ─────────────────────────────────────────────────── */
.detail-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1rem 1.1rem;
  box-shadow: var(--shadow-xs);
}
.detail-card-header {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  margin-bottom: 1.1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f0f3f7;
}
.detail-card-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: grid;
  place-content: center;
  font-size: 1rem;
  flex-shrink: 0;
}
.si-blue { background: #fff0f1; color: #e63946; }
.si-violet { background: #fff5f5; color: #ef4e5d; }
.si-green { background: #ecfdf5; color: #059669; }
.si-amber { background: #fffbeb; color: #d97706; }

.detail-card-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
}
.detail-card-header p {
  margin: 0.15rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.detail-rows {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}
.detail-label {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  flex-shrink: 0;
  padding-top: 0.15rem;
}
.detail-value {
  font-size: 0.87rem;
  color: var(--text-primary);
  text-align: right;
  word-break: break-word;
}

/* ── SCOPE SUMMARY ────────────────────────────────────────────────── */
.scope-summary {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.35rem;
}
.scope-title {
  font-weight: 700;
  color: var(--brand-blue);
}
.scope-summary ul {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}
.scope-summary li {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.78rem;
  color: var(--text-secondary);
}
.scope-summary li i {
  font-size: 0.78rem;
  color: #059669;
}

/* ── STATUS ACTIONS ───────────────────────────────────────────────── */
.status-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1.1rem;
  padding-top: 1rem;
  border-top: 1px solid #f0f3f7;
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
  background: #fff0f1;
  color: #e63946;
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

/* ── RESPONSIVE ───────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .detail-grid, .account-info-columns {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .page-heading-actions { width: 100%; justify-content: flex-end; }
}

/* Compact CRM workspace treatment */
.admin-page { padding: 0.9rem 1.25rem 1.5rem; }
.page-heading { margin-bottom: 0.8rem; gap: 0.75rem; }
.page-title-wrapper h1 { font-size: 1.35rem; margin: 0.15rem 0 0; }
.detail-layout { gap: 1rem; }
.detail-card { padding: 1rem 1.1rem; }
.detail-card-header { padding-bottom: 0.65rem; margin-bottom: 0.2rem; }
.detail-row { padding: 0.45rem 0; }
.compact-admin-page > .p-button { margin-bottom: 0.35rem; }
.compact-admin-page .page-heading { padding: 0.7rem 0.85rem; border: 1px solid #e3e9f0; border-radius: 12px; background: #fff; box-shadow: 0 1px 2px rgba(15,23,42,.03); }
.compact-heading-main { display: flex; align-items: center; min-width: 0; gap: 0.35rem; }
.header-back { flex: 0 0 auto; }
.usage-card { margin-top: 1rem; }
.usage-filters { display:flex; flex-wrap:wrap; gap:.45rem; margin-bottom:.7rem; }
.usage-filters label { display:grid; gap:.18rem; color:var(--text-muted); font-size:.66rem; font-weight:700; }
.usage-filters select,.usage-filters input,.usage-clear { min-height:30px; padding:.3rem .55rem; border:1px solid var(--border-light); border-radius:7px; background:#fff; color:var(--text-secondary); font-size:.72rem; }
.usage-clear { cursor:pointer; font-weight:700; }
.usage-more { margin-top:.55rem; border:0; background:transparent; color:var(--brand-blue); font-size:.72rem; font-weight:700; cursor:pointer; }
.usage-semantics { margin-top:.65rem; color:var(--text-muted); font-size:.7rem; }
.usage-subsection { margin-top:1rem; }
.usage-subsection h4 { margin:0 0 .45rem; font-size:.82rem; color:var(--text-primary); }
.usage-helper { margin:.15rem 0 .55rem; color:var(--text-muted); font-size:.7rem; line-height:1.45; }
.usage-empty small { display:block; margin-top:.35rem; color:var(--text-muted); font-size:.7rem; line-height:1.45; }
.usage-table-wrap,.activity-table-wrap { overflow-x:auto; margin-top:.75rem; }
.usage-table,.activity-table { width:100%; border-collapse:collapse; font-size:.72rem; }
.usage-table th,.usage-table td,.activity-table th,.activity-table td { padding:.55rem .45rem; border-top:1px solid #eef2f6; text-align:left; vertical-align:top; white-space:nowrap; }
.usage-table th,.activity-table th { color:var(--text-muted); font-size:.66rem; text-transform:uppercase; letter-spacing:.04em; }
.usage-table td small { display:block; margin-top:.15rem; color:var(--text-muted); }
.activity-table td small { display:block; max-width:28rem; white-space:normal; color:var(--text-muted); }
.activity-table summary { cursor:pointer; color:var(--brand-blue); font-weight:700; }
.activity-columns { display:grid; grid-template-columns:145px minmax(0,1fr) auto minmax(70px,auto) minmax(80px,auto) auto; gap:.6rem; margin-top:.8rem; padding:.35rem 0; color:var(--text-muted); font-size:.66rem; font-weight:700; text-transform:uppercase; letter-spacing:.04em; }
.usage-empty { padding: .75rem; color: var(--text-muted); background: #f8fafc; border-radius: 8px; font-size: .8rem; }
.usage-inline-error { display:flex; align-items:center; gap:.6rem; padding:.65rem .75rem; color:#991b1b; background:#fef2f2; border:1px solid #fecaca; border-radius:8px; font-size:.78rem; }
.usage-inline-error button { margin-left:auto; border:0; background:transparent; color:#b91c1c; font-weight:700; cursor:pointer; }
.usage-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(150px,1fr)); gap:.6rem; }
.usage-stat { display:grid; gap:.2rem; padding:.7rem; border:1px solid var(--border-light); border-radius:9px; }
.usage-stat span,.usage-stat small { color:var(--text-muted); font-size:.68rem; }
.usage-stat strong { font-size:1rem; }
.usage-history { display:grid; gap:.35rem; margin-top:.75rem; }
.usage-history-row { display:grid; grid-template-columns:145px minmax(0,1fr) auto minmax(120px,auto) minmax(90px,auto) minmax(150px,1fr); gap:.6rem; padding:.45rem 0; border-top:1px solid #eef2f6; font-size:.72rem; align-items:center; }
.usage-history-header { border-top:0; padding-top:0; color:var(--text-muted); font-size:.66rem; font-weight:700; text-transform:uppercase; letter-spacing:.04em; }
.usage-section-heading { display:flex; align-items:flex-start; justify-content:space-between; gap:.75rem; }
.activity-mode { display:grid; gap:.18rem; flex:0 0 auto; color:var(--text-muted); font-size:.66rem; font-weight:700; }
.activity-mode select { min-height:30px; padding:.3rem .55rem; border:1px solid var(--border-light); border-radius:7px; background:#fff; color:var(--text-secondary); font-size:.72rem; }
.usage-history-row details,.activity-table details { min-width:0; }
.usage-history-row details summary,.activity-table details summary { cursor:pointer; color:var(--brand-blue); font-weight:700; }
.usage-detail-grid { display:grid; gap:.2rem; min-width:180px; margin-top:.35rem; color:var(--text-muted); font-size:.68rem; line-height:1.35; white-space:normal; }
.usage-history-row small,.activity-table td small { display:block; margin-top:.2rem; color:var(--text-muted); font-size:.66rem; white-space:normal; }
.status-badge { display:inline-block; padding:.15rem .4rem; border-radius:999px; font-size:.66rem; font-weight:700; }
.status-success { color:#166534; background:#dcfce7; }
.status-failed { color:#991b1b; background:#fee2e2; }
@media (max-width: 768px) { .usage-history-row { grid-template-columns:1fr 1fr; gap:.35rem; } .usage-history-row > :first-child { grid-column:1 / -1; } }
@media (max-width: 768px) { .usage-section-heading { flex-direction:column; } .activity-mode { width:100%; } .activity-mode select { width:100%; } .history-controls { width:100%; justify-content:stretch; } .history-controls label { flex:1 1 140px; } .history-controls select { width:100%; } }
@media (max-width: 768px) { .activity-columns { display:none; } }

/* Final monitoring density and stable expandable rows */
.usage-card { margin-top: .45rem; display: flex; flex-direction: column; }
.usage-card > .usage-subsection:first-of-type { display: none; }
.usage-card > .usage-subsection .activity-mode { display: none; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading .usage-helper { display: none; }
.canonical-activity-mode { display: grid; justify-items: end; gap: .25rem; margin-bottom: .5rem; }
.canonical-activity-mode label { display: grid; gap: .18rem; color: var(--text-muted); font-size: .66rem; font-weight: 700; }
.canonical-activity-mode select { min-height: 30px; min-width: 180px; padding: .3rem .55rem; border: 1px solid var(--border-light); border-radius: 7px; background: #fff; color: var(--text-secondary); font-size: .72rem; }
.canonical-activity-mode select option { color: var(--text-secondary); background: #fff; }
.canonical-activity-mode p { margin: 0; max-width: 44rem; color: var(--text-muted); font-size: .7rem; line-height: 1.45; text-align: right; }
.usage-card > .canonical-activity-mode { order: 10; align-self: stretch; }
.usage-card > .usage-subsection:nth-of-type(2) { order: 11; }
.history-controls { display: flex; align-items: flex-end; justify-content: flex-end; gap: .5rem; flex-wrap: wrap; }
.history-controls { width: 100%; justify-content: flex-start; margin-top: .35rem; padding-top: .35rem; border-top: 1px solid #eef2f6; }
.history-controls label { display: grid; gap: .25rem; min-width: 150px; color: var(--text-muted); font-size: .68rem; font-weight: 700; }
.history-controls select { height: 30px; min-height: 30px; padding: 0 .55rem; border: 1px solid var(--border-light); border-radius: 6px; background: #fff; color: var(--text-primary); font: inherit; font-weight: 500; }
.usage-section-heading > div:first-child { display: grid; grid-template-columns: minmax(0, 1fr) 420px; column-gap: 1rem; align-items: start; width: 100%; }
.usage-section-heading > div:first-child > h4 { grid-column: 1; grid-row: 1; }
.usage-section-heading > div:first-child > .usage-helper { grid-column: 1; grid-row: 2; min-height: 2.1rem; }
.usage-section-heading > div:first-child > .history-controls { grid-column: 1 / -1; grid-row: 3; min-height: 0; align-content: start; }
.history-controls p { flex-basis: 100%; min-height: 0; margin: 0; color: var(--text-muted); font-size: .68rem; }
.history-controls label { text-align: left; }
.history-controls select { min-width: 150px; }
.usage-success-number { color: #15803d; font-weight: 700; }
.usage-failed-number { color: #b91c1c; font-weight: 700; }
.activity-table td small { white-space: nowrap; }
.history-row-actions { display: inline-flex; align-items: center; gap: .35rem; }
.history-hide-button { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; border: 1px solid #fecaca; border-radius: 6px; background: #fff; color: #b91c1c; cursor: pointer; }
.history-hide-button:hover { background: #fef2f2; }
.hidden-history-meta { margin: .35rem 0; padding: .35rem .5rem; color: var(--text-muted); background: #f8fafc; border-radius: 6px; font-size: .7rem; }
.hide-reason-field { display: grid; gap: .35rem; color: var(--text-secondary); font-size: .78rem; font-weight: 600; }
.hide-reason-field textarea { width: 100%; resize: vertical; border: 1px solid var(--border-light); border-radius: 7px; padding: .45rem; font: inherit; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading h4 { font-size: 0; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading h4::after { content: 'Riwayat Penggunaan API'; font-size: .82rem; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading .usage-helper { font-size: 0; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading .usage-helper::after { content: 'Menampilkan aktivitas pengguna beserta penggunaan layanan eksternal yang terjadi dari aktivitas tersebut.'; font-size: .7rem; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading .usage-helper { font-size: .7rem; }
.usage-card > .usage-subsection:nth-of-type(2) .usage-section-heading .usage-helper::after { content: none; }
.usage-card .detail-card-header { margin-bottom: .65rem; }
.usage-filters { align-items: end; margin-bottom: .55rem; gap: .4rem; }
.usage-filters label { min-width: 145px; }
.usage-filters select, .usage-filters input, .usage-clear { height: 30px; min-height: 30px; box-sizing: border-box; }
.usage-clear { display: inline-flex; align-items: center; justify-content: center; margin: 0; }
.usage-table-wrap, .activity-table-wrap { margin-top: .55rem; }
.usage-subsection { margin-top: .8rem; }
.usage-subsection h4 { margin-bottom: .3rem; }
.usage-helper { margin-bottom: .4rem; }
.usage-period-context { margin: -.2rem 0 .45rem; color: var(--text-muted); font-size: .68rem; }
.usage-semantics { margin-top: .45rem; }
.usage-history-table, .activity-table { width: 100%; table-layout: fixed; }
.usage-history-table .provider-col-time { width: 14%; }
.usage-history-table .provider-col-provider { width: 16%; }
.usage-history-table .provider-col-operation { width: 16%; }
.usage-history-table .provider-col-sku { width: 24%; }
.usage-history-table .provider-col-status { width: 18%; }
.usage-history-table .provider-col-detail { width: 12%; }
.usage-history-table, .activity-table { border: 1px solid #edf0f4; border-radius: 8px; overflow: hidden; }
.usage-history-table thead, .activity-table thead { background: #fafbfc; }
.usage-history-table th, .usage-history-table td, .activity-table th, .activity-table td { white-space: normal; overflow-wrap: anywhere; }
.usage-history-table th, .usage-history-table td, .activity-table th, .activity-table td { overflow-wrap: normal; word-break: normal; hyphens: none; }
.provider-main-row td, .activity-main-row td { height: 2.55rem; overflow-wrap: normal; word-break: normal; }
.detail-toggle { border: 0; padding: .2rem .35rem; color: var(--brand-blue); background: transparent; font-size: .7rem; font-weight: 700; cursor: pointer; }
.detail-toggle:hover { text-decoration: underline; }
.detail-row-expanded > td { padding: 0; background: #f8fafc; }
.usage-detail-panel { display: grid; gap: .55rem; padding: .7rem .8rem; border-top: 1px solid #e5eaf0; border-bottom: 1px solid #e5eaf0; }
.usage-detail-panel h5, .usage-detail-panel h6 { margin: 0; color: var(--text-primary); font-size: .75rem; }
.usage-detail-panel h6 { color: var(--text-muted); font-size: .66rem; text-transform: uppercase; letter-spacing: .05em; }
.usage-detail-group { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .45rem .8rem; }
.usage-detail-group span { display: grid; gap: .12rem; min-width: 0; color: var(--text-secondary); font-size: .7rem; overflow-wrap: anywhere; }
.usage-detail-group b { color: var(--text-muted); font-size: .62rem; font-weight: 700; text-transform: uppercase; letter-spacing: .035em; }
.usage-detail-group code { min-width: 0; color: var(--text-secondary); font-family: 'SF Mono', 'Fira Code', Consolas, monospace; font-size: .66rem; overflow-wrap: anywhere; white-space: pre-wrap; }
.usage-detail-secondary { display: grid; gap: .35rem; padding-top: .55rem; border-top: 1px solid #e5eaf0; }
.trace-badge { display: inline-block; width: fit-content; padding: .12rem .35rem; border-radius: 999px; font-size: .62rem; font-weight: 800; }
.trace-hit { color: #166534; background: #dcfce7; }
.trace-miss { color: #92400e; background: #fef3c7; }
@media (max-width: 900px) { .usage-filters label { min-width: 130px; } .usage-detail-group { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 900px) { .usage-section-heading > div:first-child { grid-template-columns: minmax(0, 1fr) 320px; } }
@media (max-width: 600px) { .usage-filters { align-items: stretch; } .usage-filters label, .usage-clear { width: 100%; } .usage-detail-group { grid-template-columns: 1fr; } .usage-history-table, .activity-table { min-width: 660px; } .usage-table-wrap, .activity-table-wrap { overflow-x: auto; } .usage-section-heading > div:first-child { display: block; } .usage-section-heading > div:first-child > .history-controls { width: 100%; min-height: 0; margin-top: .65rem; } .history-controls p { min-height: 0; } }
</style>
