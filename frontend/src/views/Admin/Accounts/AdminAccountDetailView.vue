<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
import { getApiUsageSummary, getApiUsageHistory, getApiActivityHistory } from '../../../api/admin'

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
const usageSummaryError = ref('')
const usageHistoryError = ref('')
const activityHistoryError = ref('')
const usageProvider = ref('')
const usageFeature = ref('')
const usageDateFrom = ref('')
const usageDateTo = ref('')
const id = computed(() => String(route.params.id))
const usageParams = () => Object.fromEntries(Object.entries({ provider: usageProvider.value, feature: usageFeature.value, dateFrom: usageDateFrom.value, dateTo: usageDateTo.value }).filter(([, value]) => value))
const reloadUsage = async () => {
  if (!id.value) return
  const params = usageParams()
  const load = async (request: () => Promise<unknown>, target: { value: any[] }, failure: { value: string }, message: string) => {
    failure.value = ''
    try { target.value = (await request()) as any[] } catch { target.value = []; failure.value = message }
  }
  await Promise.all([
    load(() => getApiUsageSummary(id.value, params), usageSummary, usageSummaryError, 'Monthly summary is unavailable.'),
    load(() => getApiUsageHistory(id.value, params), usageHistory, usageHistoryError, 'Provider usage history is unavailable.'),
    load(() => getApiActivityHistory(id.value), activityHistory, activityHistoryError, 'Activity history is unavailable.'),
  ])
}

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
        <div class="detail-grid">
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
              <div class="detail-row"><span class="detail-label">Timezone</span><span class="detail-value">{{ user.timezone || 'â€”' }}</span></div><div class="detail-row"><span class="detail-label">Location</span><span class="detail-value">{{ [user.city, user.province, user.district].filter(Boolean).join(', ') || 'â€”' }}</span></div><div class="detail-row"><span class="detail-label">Job</span><span class="detail-value">{{ [user.jobTitle, user.positionGrade, user.subDepartment].filter(Boolean).join(' · ') || 'â€”' }}</span></div><div class="detail-row"><span class="detail-label">Join Date</span><span class="detail-value">{{ user.joinDate?.slice(0, 10) || 'â€”' }}</span></div><div class="detail-row"><span class="detail-label">Gender / Birth Date</span><span class="detail-value">{{ [user.gender, user.dateOfBirth?.slice(0, 10)].filter(Boolean).join(' · ') || 'â€”' }}</span></div><div class="detail-row"><span class="detail-label">Phone Numbers</span><span class="detail-value">{{ user.phones?.map((phone) => phone.phoneNumber).join(', ') || user.phone || 'â€”' }}</span></div>
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
          <div class="detail-card-header"><div class="detail-card-icon si-violet"><i class="pi pi-chart-bar" /></div><div><h3>API Usage</h3><p>Monitoring penggunaan API berdasarkan aktivitas yang tercatat di sistem. Estimasi biaya bukan merupakan tagihan resmi dari provider.</p></div></div>
          <div class="usage-filters"><select v-model="usageProvider" @change="reloadUsage"><option value="">Provider</option><option value="GOOGLE_MAPS">Google Maps</option><option value="OPENAI">OpenAI</option></select><input v-model="usageFeature" placeholder="Operation / Feature" @change="reloadUsage" /><input v-model="usageDateFrom" type="date" aria-label="From date" @change="reloadUsage" /><input v-model="usageDateTo" type="date" aria-label="To date" @change="reloadUsage" /><button type="button" class="usage-clear" @click="usageProvider=''; usageFeature=''; usageDateFrom=''; usageDateTo=''; reloadUsage()">Clear</button></div>
          <div v-if="usageSummaryError" class="usage-inline-error">{{ usageSummaryError }} <button type="button" @click="reloadUsage">Retry</button></div>
          <div v-else-if="!usageSummary.length" class="usage-empty">No API usage recorded yet.</div>
          <div v-else class="usage-table-wrap"><p class="usage-helper">Ringkasan penggunaan API pengguna ini pada periode yang dipilih.</p><table class="usage-table"><thead><tr><th>API / Operasi</th><th>Request Provider Tercatat</th><th>Kuota Gratis</th><th>Penggunaan</th><th>Estimasi Biaya</th></tr></thead><tbody><tr v-for="item in usageSummary" :key="`${item.provider}-${item.operation}`"><td><strong>{{ item.provider === 'GOOGLE_MAPS' ? 'Google Maps' : item.provider }}</strong><small>{{ item.operation.replaceAll('_', ' ') }}</small></td><td>{{ item.requests || 0 }}</td><td>{{ item.freeTierVerified ? item.freeTier : 'Belum dikonfigurasi' }}</td><td>{{ item.freeTierVerified && item.usagePercent != null ? `${item.usagePercent.toFixed(1)}%` : '—' }}</td><td>{{ item.estimatedCost ? item.estimatedCost : 'Belum dikonfigurasi' }}</td></tr></tbody></table></div>
          <div v-if="usageSummary.length" class="usage-semantics">Hit provider menunjukkan request yang benar-benar dikirim ke layanan eksternal. Aktivitas aplikasi dan cache hit dicatat secara terpisah.</div>
          <section class="usage-subsection"><h4>Riwayat Penggunaan Provider</h4><p class="usage-helper">Menampilkan riwayat request ke layanan eksternal seperti Google Maps atau OpenAI yang tercatat atas penggunaan akun ini.</p><div v-if="usageHistoryError" class="usage-inline-error">{{ usageHistoryError }} <button type="button" @click="reloadUsage">Retry</button></div><div v-else-if="!usageHistory.length" class="usage-empty">Belum ada riwayat penggunaan provider untuk pengguna ini.</div><div v-else class="usage-history"><div v-for="item in usageHistory.slice(0, 10)" :key="`${item.createdAt}-${item.operation}`" class="usage-history-row"><span>{{ formatDateTime(item.createdAt) }}</span><strong>{{ item.provider === 'GOOGLE_MAPS' ? 'Google Maps' : item.provider }}</strong><span>{{ item.operation.replaceAll('_', ' ') }}</span><span>{{ item.fieldMask || '—' }}</span><span>{{ item.success ? 'Berhasil' : 'Gagal' }}</span></div></div></section>
          <section class="usage-subsection"><h4>Riwayat Aktivitas</h4><p class="usage-helper">Menampilkan aktivitas CRM yang dilakukan langsung oleh pengguna ini.</p><div v-if="activityHistoryError" class="usage-inline-error">{{ activityHistoryError }} <button type="button" @click="reloadUsage">Retry</button></div><div v-else-if="!activityHistory.length" class="usage-empty"><strong>Belum ada aktivitas yang tercatat untuk pengguna ini.</strong><small>Riwayat aktivitas hanya mencatat tindakan yang dilakukan langsung oleh pengguna ini. Aktivitas yang dilakukan Admin saat melihat atau mengelola akun ini akan tercatat pada riwayat aktivitas Admin.</small><small v-if="usageHistory.length" class="usage-helper">Sebagian riwayat penggunaan API dapat berasal dari data sebelum fitur Activity History diterapkan.</small></div><div v-else class="activity-table-wrap"><table class="activity-table"><thead><tr><th>Waktu</th><th>Aktivitas</th><th>Operasi</th><th>SKU</th><th>Cache</th><th>Hit Provider</th><th>Status</th><th>Detail</th></tr></thead><tbody><tr v-for="item in activityHistory.slice(0, 10)" :key="`${item.createdAt}-${item.requestId}`"><td>{{ formatDateTime(item.createdAt) }}</td><td>{{ item.trace?.action || '—' }}</td><td>{{ item.trace?.operation || '—' }}</td><td>{{ item.trace?.sku || item.trace?.sku_category || '—' }}</td><td>{{ item.trace?.cache_status || '—' }}</td><td>{{ item.trace?.provider_hit_count ?? '—' }}</td><td>{{ item.responseStatus >= 400 ? 'Gagal' : 'Berhasil' }}</td><td><details><summary>Lihat</summary><small>Endpoint: {{ item.endpoint }} · Request ID: {{ item.requestId }} · Field mask: {{ item.trace?.field_mask || '—' }}</small></details></td></tr></tbody></table></div></section>
        </section>
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
  gap: 0.8rem;
  align-items: start;
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
  .detail-grid {
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
.usage-filters select,.usage-filters input,.usage-clear { min-height:30px; padding:.3rem .55rem; border:1px solid var(--border-light); border-radius:7px; background:#fff; color:var(--text-secondary); font-size:.72rem; }
.usage-clear { cursor:pointer; font-weight:700; }
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
.usage-history-row { display:grid; grid-template-columns:145px minmax(0,1fr) auto minmax(120px,auto) minmax(120px,auto); gap:.6rem; padding:.45rem 0; border-top:1px solid #eef2f6; font-size:.72rem; align-items:center; }
@media (max-width: 768px) { .usage-history-row { grid-template-columns:1fr 1fr; gap:.35rem; } .usage-history-row > :first-child { grid-column:1 / -1; } }
@media (max-width: 768px) { .activity-columns { display:none; } }
</style>
