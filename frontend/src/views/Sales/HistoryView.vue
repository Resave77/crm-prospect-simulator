<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { useCrmStore } from '../../stores/crm'
import { deleteMyVisit, getMyVisits } from '../../api/crm'
import { deleteCustomerVisit, loadCustomerVisits, type CustomerVisitRecord } from '../../utils/visitEntity'
import type { VisitMonitoringItem } from '../../types/crm'
import { formatErrorMessage } from '../../utils/format'

const crm = useCrmStore()
const error = ref('')
const activeTab = ref<'visits' | 'outcomes'>('visits')

const prospectVisits = ref<VisitMonitoringItem[]>([])
const customerVisits = ref<CustomerVisitRecord[]>([])
const visitsLoading = ref(false)
const deleteTarget = ref<VisitHistoryItem | null>(null)
const deleteBusy = ref(false)
const deleteDialogVisible = computed({
  get: () => deleteTarget.value !== null,
  set: (visible: boolean) => {
    if (!visible && !deleteBusy.value) deleteTarget.value = null
  },
})

const allVisits = computed(() => {
  const prospect: VisitHistoryItem[] = prospectVisits.value.map((v) => ({
    id: v.id,
    entityName: v.customerName,
    entityType: v.entityType,
    checkInAt: v.checkInAt,
    checkOutAt: v.checkOutAt,
    duration: v.durationSeconds,
    radiusStatus: v.radiusStatus,
    visitNotes: v.visitNotes,
    followUpNotes: v.followUpNotes,
    distanceMeters: v.distanceMeters,
    industryGroup: v.industryGroup,
    prospectId: v.prospectId,
    entityId: v.customerId,
    localOnly: false,
  }))
  const customer: VisitHistoryItem[] = customerVisits.value.map((v) => ({
    id: v.id,
    entityName: v.entityName,
    entityType: 'customer' as const,
    checkInAt: v.checkInAt,
    checkOutAt: v.checkOutAt,
    duration: v.checkOutAt ? (new Date(v.checkOutAt).getTime() - new Date(v.checkInAt).getTime()) / 1000 : undefined,
    radiusStatus: 'UNKNOWN' as const,
    visitNotes: v.visitResult,
    followUpNotes: v.followUpNotes,
    distanceMeters: 0,
    industryGroup: '',
    prospectId: '',
    entityId: v.entityId,
    localOnly: true,
  }))
  return [...prospect, ...customer].sort((a, b) => new Date(b.checkInAt).getTime() - new Date(a.checkInAt).getTime())
})

const outcomes = computed(() =>
  crm.myProspects.filter((v) => v.status === 'WON' || v.status === 'LOST'),
)

interface VisitHistoryItem {
  id: string
  entityName: string
  entityType: 'prospect' | 'customer'
  checkInAt: string
  checkOutAt?: string
  duration?: number
  radiusStatus: string
  visitNotes: string
  followUpNotes: string
  distanceMeters: number
  industryGroup: string
  prospectId: string
  entityId?: string
  localOnly: boolean
}

function confirmDelete(visit: VisitHistoryItem) {
  deleteTarget.value = visit
}

async function executeDelete() {
  const target = deleteTarget.value
  if (!target) return
  deleteBusy.value = true
  error.value = ''
  try {
    if (target.localOnly) {
      deleteCustomerVisit(target.id)
      customerVisits.value = customerVisits.value.filter((visit) => visit.id !== target.id)
    } else {
      await deleteMyVisit(target.id)
      prospectVisits.value = prospectVisits.value.filter((visit) => visit.id !== target.id)
    }
    deleteTarget.value = null
  } catch (e) {
    error.value = formatErrorMessage(e)
  } finally {
    deleteBusy.value = false
  }
}

function formatDuration(seconds?: number) {
  if (!seconds) return '—'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

function formatDistance(meters: number) {
  if (meters < 1000) return `${Math.round(meters)}m`
  return `${(meters / 1000).toFixed(1)}km`
}

function formatCheckIn(dateStr: string) {
  return new Date(dateStr).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  try {
    visitsLoading.value = true
    const [visits] = await Promise.all([
      getMyVisits(),
      crm.loadMyProspects(),
    ])
    prospectVisits.value = visits
    customerVisits.value = loadCustomerVisits()
  } catch (e) {
    error.value = formatErrorMessage(e)
  } finally {
    visitsLoading.value = false
  }
})
</script>

<template>
  <section class="mobile-page">
    <RouterLink class="history-back" to="/sales/dashboard">
      <i class="pi pi-arrow-left" />
    </RouterLink>

    <div class="mobile-title">
      <div>
        <p class="eyebrow">Activity log</p>
        <h1>History</h1>
      </div>
    </div>

    <!-- Tabs -->
    <div class="htabs">
      <button class="htab" :class="{ active: activeTab === 'visits' }" @click="activeTab = 'visits'">
        <i class="pi pi-map-marker" /> Visits
      </button>
      <button class="htab" :class="{ active: activeTab === 'outcomes' }" @click="activeTab = 'outcomes'">
        <i class="pi pi-trophy" /> Outcomes
      </button>
    </div>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <!-- Visits Tab -->
    <template v-if="activeTab === 'visits'">
      <div v-if="visitsLoading" class="ht-list">
        <div v-for="i in 3" :key="i" class="ht-skeleton">
          <div class="sk-circle" />
          <div class="sk-lines"><div class="sk-line w60" /><div class="sk-line w40" /></div>
        </div>
      </div>

      <div v-else-if="allVisits.length" class="ht-list">
        <div v-for="visit in allVisits" :key="visit.id" class="ht-visit-card">
          <div class="ht-visit-header">
            <div class="ht-visit-avatar" :class="visit.entityType">
              {{ visit.entityName.split(/\s+/).slice(0, 2).map(w => w.charAt(0).toUpperCase()).join('') }}
            </div>
            <div class="ht-visit-info">
              <strong>{{ visit.entityName }}</strong>
              <span class="ht-visit-meta">
                <Tag :value="visit.entityType === 'prospect' ? 'Prospect' : 'Customer'" :severity="visit.entityType === 'prospect' ? 'info' : 'success'" />
                {{ visit.industryGroup }}
              </span>
            </div>
            <button class="ht-delete-btn" type="button" aria-label="Delete visit" @click="confirmDelete(visit)">
              <i class="pi pi-trash" />
            </button>
          </div>

          <div class="ht-visit-details">
            <div class="ht-detail-row">
              <i class="pi pi-clock" />
              <span>{{ formatCheckIn(visit.checkInAt) }} <template v-if="visit.checkOutAt">→ {{ formatCheckIn(visit.checkOutAt) }}</template></span>
            </div>
            <div class="ht-detail-row" v-if="visit.duration">
              <i class="pi pi-hourglass" />
              <span>{{ formatDuration(visit.duration) }}</span>
            </div>
            <div class="ht-detail-row" v-if="visit.distanceMeters">
              <i class="pi pi-compass" />
              <span>{{ formatDistance(visit.distanceMeters) }} from target</span>
            </div>
            <div class="ht-detail-row">
              <i class="pi pi-map-marker" />
              <Tag
                :value="visit.radiusStatus === 'INSIDE' ? 'Inside radius' : visit.radiusStatus === 'OUTSIDE' ? 'Outside radius' : 'Unknown'"
                :severity="visit.radiusStatus === 'INSIDE' ? 'success' : visit.radiusStatus === 'OUTSIDE' ? 'warn' : 'secondary'"
              />
            </div>
          </div>

          <div v-if="visit.visitNotes" class="ht-visit-notes">
            <span class="ht-notes-label">Visit notes:</span> {{ visit.visitNotes }}
          </div>
          <div v-if="visit.followUpNotes" class="ht-visit-notes">
            <span class="ht-notes-label">Follow-up:</span> {{ visit.followUpNotes }}
          </div>

          <RouterLink
            v-if="visit.entityType === 'prospect' && visit.prospectId"
            class="ht-visit-link"
            :to="`/sales/my-prospects/${visit.prospectId}`"
          >
            View prospect <i class="pi pi-arrow-right" />
          </RouterLink>
          <RouterLink
            v-else-if="visit.entityType === 'customer'"
            class="ht-visit-link"
            :to="`/sales/my-customers/${visit.entityId}`"
          >
            View customer <i class="pi pi-arrow-right" />
          </RouterLink>
        </div>
      </div>

      <div v-else class="ht-empty">
        <i class="pi pi-map-marker" />
        <strong>No visits yet</strong>
        <p>Check-in records from your field visits will appear here.</p>
      </div>
    </template>

    <!-- Outcomes Tab -->
    <template v-if="activeTab === 'outcomes'">
      <div v-if="outcomes.length" class="ht-list">
        <RouterLink
          v-for="item in outcomes"
          :key="item.id"
          class="ht-outcome-card"
          :to="`/sales/my-prospects/${item.id}`"
        >
          <div class="ht-outcome-info">
            <strong>{{ item.placeName }}</strong>
            <span>{{ item.industryGroup }} · {{ new Date(item.updatedAt).toLocaleDateString() }}</span>
          </div>
          <Tag :value="item.status" :severity="item.status === 'WON' ? 'success' : 'danger'" />
        </RouterLink>
      </div>

      <div v-else class="ht-empty">
        <i class="pi pi-trophy" />
        <strong>No completed outcomes yet</strong>
        <p>Prospect outcomes will appear here.</p>
      </div>
    </template>

    <Dialog v-model:visible="deleteDialogVisible" modal header="Delete History" :style="{ width: 'min(92vw, 400px)' }" :closable="!deleteBusy">
      <p v-if="deleteTarget" class="ht-delete-copy">
        Delete visit history for <strong>{{ deleteTarget.entityName }}</strong>? This action cannot be undone.
      </p>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined :disabled="deleteBusy" @click="deleteTarget = null" />
        <Button label="Delete" icon="pi pi-trash" severity="danger" :loading="deleteBusy" @click="executeDelete" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.history-back { display: inline-flex; align-items: center; justify-content: center; width: 2rem; height: 2rem; color: var(--brand-blue); background: var(--brand-blue-bg); border: 1px solid transparent; border-radius: var(--radius-md); text-decoration: none; font-size: 0.9rem; margin-bottom: 0.5rem; transition: background var(--transition-fast), border-color var(--transition-fast); }
.history-back:hover { background: #ffd9dc; border-color: var(--brand-blue); }

.htabs { display: flex; gap: 0.5rem; margin-bottom: 0.5rem; }
.htab {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 0.35rem;
  padding: 0.6rem; border-radius: var(--radius-xl); border: 1px solid var(--border-light);
  background: var(--surface-card); color: var(--text-muted); font-size: 0.82rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.htab:hover { background: #f8fafc; }
.htab.active { background: var(--brand-blue); color: #fff; border-color: var(--brand-blue); }

.ht-list { display: grid; gap: 0.6rem; }

.ht-skeleton { display: flex; align-items: center; gap: 0.75rem; padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); }
.sk-circle { width: 40px; height: 40px; border-radius: 12px; background: #e2e8f0; flex-shrink: 0; animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-lines { display: flex; flex-direction: column; gap: 0.4rem; flex: 1; }
.sk-line { height: 10px; border-radius: 5px; background: #e2e8f0; animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-line.w40 { width: 40%; }
.sk-line.w60 { width: 60%; }
@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

.ht-visit-card, .ht-outcome-card {
  padding: 0.85rem 1rem; background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); box-shadow: var(--shadow-sm); text-decoration: none;
  color: var(--text-primary); transition: all 0.2s ease; display: grid; gap: 0.5rem;
}
.ht-visit-card:hover, .ht-outcome-card:hover { border-color: var(--border-default); box-shadow: var(--shadow-md); transform: translateY(-1px); }

.ht-visit-header { display: flex; align-items: center; gap: 0.6rem; }
.ht-delete-btn {
  width: 2rem; height: 2rem; display: grid; place-items: center; flex-shrink: 0;
  border: 0; border-radius: 10px; background: #fef2f2; color: #dc2626; cursor: pointer;
}
.ht-delete-btn:hover { background: #fee2e2; }
.ht-delete-copy { margin: 0; font-size: 0.85rem; line-height: 1.5; }
.ht-visit-avatar {
  width: 36px; height: 36px; display: grid; place-items: center; border-radius: 10px;
  font-size: 0.7rem; font-weight: 800; color: #fff; flex-shrink: 0;
}
.ht-visit-avatar.prospect { background: linear-gradient(135deg, #e63946, #d62839); }
.ht-visit-avatar.customer { background: linear-gradient(135deg, #16a34a, #15803d); }
.ht-visit-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.15rem; }
.ht-visit-info strong { font-size: 0.88rem; font-weight: 700; }
.ht-visit-meta { display: flex; align-items: center; gap: 0.35rem; color: var(--text-muted); font-size: 0.75rem; }

.ht-visit-details { display: flex; flex-wrap: wrap; gap: 0.35rem 0.75rem; }
.ht-detail-row { display: flex; align-items: center; gap: 0.3rem; font-size: 0.75rem; color: var(--text-secondary); }
.ht-detail-row i { color: var(--text-muted); font-size: 0.68rem; }

.ht-visit-notes { font-size: 0.78rem; color: var(--text-secondary); line-height: 1.4; }
.ht-notes-label { font-weight: 600; color: var(--text-muted); }

.ht-visit-link {
  display: inline-flex; align-items: center; gap: 0.3rem;
  font-size: 0.75rem; font-weight: 600; color: var(--brand-blue); text-decoration: none;
}
.ht-visit-link:hover { text-decoration: underline; }

.ht-outcome-card { flex-direction: row; align-items: center; justify-content: space-between; gap: 0.5rem; }
.ht-outcome-info { display: grid; gap: 0.15rem; }
.ht-outcome-info strong { font-size: 0.9rem; font-weight: 700; }
.ht-outcome-info span { color: var(--text-muted); font-size: 0.78rem; }

.ht-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; color: var(--text-muted); }
.ht-empty i { font-size: 2rem; opacity: 0.3; }
.ht-empty strong { color: var(--text-secondary); font-size: 0.9rem; }
.ht-empty p { margin: 0; font-size: 0.78rem; max-width: 250px; }

/* ── Desktop ───────────────────────────────────────────────── */
@media (min-width: 768px) {
  .history-back { display: none; }

  .htabs { width: auto; display: inline-flex; }
  .htab { flex: none; padding: 0.6rem 1.4rem; }

  .ht-list { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.8rem; align-items: start; }
}
</style>
