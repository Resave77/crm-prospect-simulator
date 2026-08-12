<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { useVisitLocation } from '../../../composables/sales/useVisitLocation'
import {
  isValidEntityType,
  normalizeRouteId,
  fetchProspectVisitData,
  fetchCustomerVisitData,
  type VisitEntityContext,
} from '../../../utils/visitEntity'
import { formatErrorMessage } from '../../../utils/format'
import VisitLocationCard from '../../../components/sales/visit/VisitLocationCard.vue'

type PageState = 'loading' | 'ready' | 'invalid-params' | 'not-found' | 'no-active-visit' | 'error' | 'location-unavailable'

const route = useRoute()
const router = useRouter()

const resolvedEntityType = computed(() => {
  if (isValidEntityType(route.meta.entityType)) return route.meta.entityType
  return null
})
const resolvedEntityId = computed(() => normalizeRouteId(route.params.id))

const entity = ref<VisitEntityContext | null>(null)
const activeVisit = ref<{ id: string; checkInAt: string } | null>(null)
const pageState = ref<PageState>('loading')
const pageError = ref('')

const visitResult = ref('')
const visitOutcome = ref('')
const followUpNotes = ref('')
const followUpDate = ref('')

const location = useVisitLocation()

const elapsed = ref('—')
let elapsedTimer: ReturnType<typeof setInterval> | null = null

const VISIT_RESULT_OPTIONS = [
  'Meeting completed',
  'Contacted',
  'No response',
  'Location closed',
  'Reschedule required',
  'Not interested',
]

const PROSPECT_OUTCOME_OPTIONS = [
  'Visited',
  'Needs follow-up',
  'No follow-up',
  'Unsuccessful',
]

const CUSTOMER_OUTCOME_OPTIONS = [
  'Completed',
  'Follow-up required',
  'Issue reported',
  'Order discussion',
  'Routine visit',
]

const outcomeOptions = computed(() =>
  resolvedEntityType.value === 'customer' ? CUSTOMER_OUTCOME_OPTIONS : PROSPECT_OUTCOME_OPTIONS,
)

const needsFollowUp = computed(() => {
  const v = visitOutcome.value
  return v === 'Needs follow-up' || v === 'Follow-up required'
})

const followUpDateValid = computed(() => {
  if (!needsFollowUp.value || !followUpDate.value) return true
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return new Date(followUpDate.value) >= today
})

const hasValidGps = computed(() =>
  Boolean(
    location.state.value.coords &&
    Number.isFinite(location.state.value.coords.latitude) &&
    Number.isFinite(location.state.value.coords.longitude)
  )
)

const canSave = computed(() =>
  Boolean(
    activeVisit.value &&
    visitResult.value !== '' &&
    visitOutcome.value !== '' &&
    followUpDateValid.value &&
    hasValidGps.value
  )
)

const insideRadius = computed(() => {
  if (!entity.value || !location.state.value.coords) return true
  if (entity.value.latitude == null || entity.value.longitude == null) return true
  return location.isInsideRadius(
    entity.value.latitude,
    entity.value.longitude,
    entity.value.attendanceRadiusMeters,
  )
})

function goBack() {
  if (resolvedEntityType.value === 'customer') {
    router.push({ name: 'SalesCustomerDetail', params: { id: resolvedEntityId.value } })
  } else {
    router.push({ name: 'SalesProspectDetail', params: { id: resolvedEntityId.value } })
  }
}

function goBackToList() {
  router.push(resolvedEntityType.value === 'customer' ? '/sales/my-customers' : '/sales/my-prospects')
}

function startElapsedTimer() {
  updateElapsed()
  elapsedTimer = setInterval(updateElapsed, 30000)
}

function updateElapsed() {
  if (!activeVisit.value) { elapsed.value = '—'; return }
  const diff = Date.now() - new Date(activeVisit.value.checkInAt).getTime()
  const mins = Math.floor(diff / 60000)
  const hrs = Math.floor(mins / 60)
  const remMins = mins % 60
  elapsed.value = hrs > 0 ? `${hrs}h ${remMins}m` : `${remMins}m`
}

function localStorageKey() {
  return `visit-result-${resolvedEntityType.value}-${resolvedEntityId.value}`
}

async function initialize() {
  if (!resolvedEntityType.value || !resolvedEntityId.value) {
    pageState.value = 'invalid-params'
    pageError.value = 'The visit link is missing a valid entity type or ID.'
    return
  }

  pageState.value = 'loading'
  pageError.value = ''

  try {
    if (resolvedEntityType.value === 'prospect') {
      const { entity: ctx, review } = await fetchProspectVisitData(resolvedEntityId.value)
      entity.value = ctx

      if (!entity.value) {
        pageState.value = 'not-found'
        return
      }

      const open = review.visits.find((v) => !v.checkOutAt)
      if (!open) {
        pageState.value = 'no-active-visit'
        return
      }

      activeVisit.value = { id: open.id, checkInAt: open.checkInAt }
      startElapsedTimer()

      if (entity.value.latitude == null || entity.value.longitude == null) {
        pageState.value = 'location-unavailable'
        return
      }

      pageState.value = 'ready'
      location.startWatching()
    } else {
      const { entity: ctx } = await fetchCustomerVisitData(resolvedEntityId.value)
      entity.value = ctx

      if (!entity.value) {
        pageState.value = 'not-found'
        return
      }

      if (entity.value.latitude == null || entity.value.longitude == null) {
        pageState.value = 'location-unavailable'
        return
      }

      activeVisit.value = { id: 'customer-sim', checkInAt: new Date().toISOString() }
      startElapsedTimer()

      pageState.value = 'ready'
      location.startWatching()
    }
  } catch (caught) {
    const status = (caught as { response?: { status?: number } })?.response?.status
    if (status === 404) {
      pageState.value = 'not-found'
    } else {
      pageError.value = formatErrorMessage(caught)
      pageState.value = 'error'
    }
  }
}

function handleSave() {
  if (!canSave.value) return

  const data = {
    visitResult: visitResult.value,
    visitOutcome: visitOutcome.value,
    followUpNotes: followUpNotes.value,
    followUpDate: followUpDate.value,
  }
  localStorage.setItem(localStorageKey(), JSON.stringify(data))

  const checkoutRoute = resolvedEntityType.value === 'customer'
    ? { name: 'SalesCustomerCheckOut', params: { id: resolvedEntityId.value } }
    : { name: 'SalesProspectCheckOut', params: { id: resolvedEntityId.value } }
  router.push(checkoutRoute)
}

onMounted(initialize)

onBeforeUnmount(() => {
  if (elapsedTimer) clearInterval(elapsedTimer)
  location.stopWatching()
})
</script>

<template>
  <section class="visitresult-page">
    <button class="back-link" @click="pageState === 'loading' ? goBackToList() : goBack()"><i class="pi pi-arrow-left" /></button>

    <Message v-if="pageError" severity="error" closable @close="pageError = ''">{{ pageError }}</Message>

    <!-- Loading -->
    <div v-if="pageState === 'loading'" class="visitresult-skeleton">
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /></div>
    </div>

    <!-- Invalid params -->
    <div v-else-if="pageState === 'invalid-params'" class="visitresult-empty">
      <div class="visitresult-empty-icon"><i class="pi pi-exclamation-triangle" /></div>
      <strong>Invalid visit parameters</strong>
      <p>The visit link is missing a valid entity type or ID.</p>
      <button class="visitresult-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- Not found -->
    <div v-else-if="pageState === 'not-found'" class="visitresult-empty">
      <div class="visitresult-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>{{ resolvedEntityType === 'customer' ? 'Customer' : 'Prospect' }} not found</strong>
      <p>This entity may have been removed or you don't have access.</p>
      <button class="visitresult-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- No active visit -->
    <div v-else-if="pageState === 'no-active-visit'" class="visitresult-empty">
      <div class="visitresult-empty-icon"><i class="pi pi-sign-in" /></div>
      <strong>{{ resolvedEntityType === 'customer' ? 'Customer visit not available' : 'No active visit' }}</strong>
      <p>{{ resolvedEntityType === 'customer' ? 'Customer visit flow is not yet supported.' : 'You need to check in before you can fill in visit results.' }}</p>
      <button class="visitresult-empty-btn" @click="goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <!-- Error -->
    <div v-else-if="pageState === 'error'" class="visitresult-empty">
      <div class="visitresult-empty-icon"><i class="pi pi-exclamation-circle" /></div>
      <strong>Unable to load visit details</strong>
      <p>{{ pageError }}</p>
      <button class="visitresult-empty-btn" @click="initialize"><i class="pi pi-refresh" /> Try again</button>
    </div>

    <!-- Location unavailable -->
    <div v-else-if="pageState === 'location-unavailable' && entity" class="visitresult-empty">
      <div class="visitresult-empty-icon"><i class="pi pi-map-marker" /></div>
      <strong>Visit location unavailable</strong>
      <p>The assigned location does not have valid coordinates.</p>
      <div class="visitresult-empty-entity">{{ entity.name }}</div>
      <button class="visitresult-empty-btn" @click="goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <template v-else-if="pageState === 'ready' && entity && activeVisit">
      <!-- Visit Summary -->
      <div class="vrcard vrcard-summary">
        <div class="vrcard-summary-top">
          <div class="vrcard-avatar">{{ entity.name.split(/\s+/).slice(0, 2).map(w => w.charAt(0).toUpperCase()).join('') }}</div>
          <div class="vrcard-identity">
            <p class="eyebrow">Visit Result — {{ entity.entityType === 'prospect' ? 'Prospect' : 'Customer' }}</p>
            <h1>{{ entity.name }}</h1>
          </div>
          <Tag value="In Progress" severity="warn" />
        </div>
        <div class="vrcard-stats">
          <div class="vrcard-stat"><span>Check-in Time</span><strong>{{ new Date(activeVisit.checkInAt).toLocaleString() }}</strong></div>
          <div class="vrcard-stat"><span>Duration</span><strong>{{ elapsed }}</strong></div>
        </div>
      </div>

      <!-- Two-column content grid -->
      <div class="visitresult-content-grid">
        <!-- Left column: Location/map -->
        <div class="visitresult-main-column">
          <div class="vrcard">
            <div class="vrcard-header-row">
              <h2>Location</h2>
              <button class="vrcard-refresh-btn" :disabled="location.state.value.loading" @click="location.refreshOnce()" title="Refresh location">
                <i class="pi" :class="location.state.value.loading ? 'pi-spin pi-sync' : 'pi-refresh'" />
              </button>
            </div>
            <VisitLocationCard
              :target-latitude="entity.latitude"
              :target-longitude="entity.longitude"
              :target-label="entity.name"
              :sales-coords="location.state.value.coords"
              :radius-meters="entity.attendanceRadiusMeters"
              height="170px"
            />
            <div class="vrcard-location-rows">
              <div v-if="location.state.value.coords" class="vrcard-row">
                <i class="pi pi-map-marker" />
                <span>Current: {{ location.state.value.coords.latitude.toFixed(6) }}, {{ location.state.value.coords.longitude.toFixed(6) }}</span>
              </div>
              <div v-if="location.state.value.coords && entity.latitude != null" class="vrcard-row vrcard-distance">
                <i class="pi pi-compass" />
                <span>{{ location.distanceFormatted(entity.latitude, entity.longitude!) }} from target</span>
              </div>
              <div v-if="location.state.value.coords" class="vrcard-row">
                <Tag
                  :value="insideRadius ? 'Inside radius' : 'Outside radius'"
                  :severity="insideRadius ? 'success' : 'warn'"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Right column: Visit result form + save -->
        <div class="visitresult-side-column">
          <div class="vrcard">
            <h2>Visit Result</h2>
            <div class="vrcard-form">
              <label class="vrcard-field">
                <span>Visit Result *</span>
                <select v-model="visitResult" class="vrcard-select">
                  <option value="" disabled>Select result</option>
                  <option v-for="opt in VISIT_RESULT_OPTIONS" :key="opt" :value="opt">{{ opt }}</option>
                </select>
              </label>
              <label class="vrcard-field">
                <span>Visit Outcome *</span>
                <select v-model="visitOutcome" class="vrcard-select">
                  <option value="" disabled>Select outcome</option>
                  <option v-for="opt in outcomeOptions" :key="opt" :value="opt">{{ opt }}</option>
                </select>
              </label>
              <label class="vrcard-field">
                <span>Visit Notes</span>
                <Textarea v-model="followUpNotes" rows="3" fluid placeholder="Details about the visit..." />
              </label>
              <label v-if="needsFollowUp" class="vrcard-field">
                <span>Next Follow-Up Date *</span>
                <input v-model="followUpDate" type="date" class="vrcard-input" :min="new Date().toISOString().split('T')[0]" />
                <span v-if="!followUpDateValid" class="vrcard-field-error">Follow-up date cannot be in the past.</span>
              </label>
            </div>
          </div>

          <!-- Save Action -->
          <div class="visitresult-bottom">
            <Button
              label="Save & Continue to Check-Out"
              icon="pi pi-arrow-right"
              :disabled="!canSave"
              class="visitresult-submit-btn"
              @click="handleSave"
            />
            <p v-if="!hasValidGps" class="visitresult-bottom-hint">Current location is required</p>
            <p v-else-if="!visitResult" class="visitresult-bottom-hint">Visit result is required</p>
            <p v-else-if="!visitOutcome" class="visitresult-bottom-hint">Visit outcome is required</p>
            <p v-else class="visitresult-bottom-hint">Visit details are ready to save</p>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.visitresult-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: calc(68px + 92px + env(safe-area-inset-bottom) + 1rem); }

.visitresult-skeleton { display: grid; gap: 0.85rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w40 { width: 40%; }
.sk-line.w50 { width: 50%; }
.sk-line.w60 { width: 60%; }
.sk-line.w80 { width: 80%; }

.visitresult-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.visitresult-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.visitresult-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.visitresult-empty p { margin: 0; color: var(--text-muted); font-size: 0.8rem; max-width: 280px; }
.visitresult-empty-entity { color: var(--text-secondary); font-size: 0.85rem; font-weight: 600; }
.visitresult-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; border: 0; cursor: pointer; }

.vrcard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.vrcard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

.vrcard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.vrcard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.vrcard-avatar {
  width: 48px; height: 48px; display: grid; place-items: center; border-radius: 14px;
  background: linear-gradient(135deg, #e63946, #d62839); color: #fff; font-weight: 800;
  font-size: 0.9rem; flex-shrink: 0; box-shadow: 0 3px 10px rgba(230, 57, 70, 0.25);
}
.vrcard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.vrcard-identity .eyebrow { margin: 0; }
.vrcard-identity h1 { margin: 0; font-size: 1.1rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

.vrcard-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
.vrcard-stat { padding: 0.75rem; border-radius: 12px; background: #f8fafc; }
.vrcard-stat span { display: block; color: var(--text-muted); font-size: 0.6rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.15rem; }
.vrcard-stat strong { font-size: 0.82rem; color: var(--text-primary); }

.vrcard-header-row { display: flex; align-items: center; justify-content: space-between; }
.vrcard-header-row h2 { margin: 0; }
.vrcard-refresh-btn {
  width: 28px; height: 28px; border-radius: 8px; border: 1px solid var(--border-light);
  background: #fff; color: var(--text-muted); cursor: pointer; display: grid; place-items: center;
  font-size: 0.72rem; transition: all 0.15s ease;
}
.vrcard-refresh-btn:hover { color: var(--brand-blue); border-color: #f4b3ba; background: #fff0f1; }
.vrcard-refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.vrcard-location-rows { display: grid; gap: 0.4rem; }
.vrcard-row { display: flex; align-items: flex-start; gap: 0.5rem; color: var(--text-secondary); font-size: 0.78rem; line-height: 1.4; }
.vrcard-row i { color: var(--text-muted); font-size: 0.68rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.15rem; }
.vrcard-distance { color: var(--brand-blue); font-weight: 600; }

.vrcard-form { display: grid; gap: 0.75rem; }
.vrcard-field { display: flex; flex-direction: column; gap: 0.3rem; }
.vrcard-field span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.vrcard-field-error { color: #dc2626; font-size: 0.7rem; }
.vrcard-select, .vrcard-input {
  padding: 0.65rem 0.85rem; border: 1px solid var(--border-light); border-radius: 12px;
  background: #f8fafc; color: var(--text-primary); font-size: 0.82rem; font-family: inherit;
  width: 100%; box-sizing: border-box;
}
.vrcard-select:focus, .vrcard-input:focus { outline: 0; border-color: var(--brand-blue); }

.visitresult-bottom {
  position: fixed; bottom: 68px; left: 0; right: 0;
  width: 100%; z-index: 50;
  display: flex; flex-direction: column; gap: 0.3rem;
  padding: 0.75rem 1rem; padding-bottom: calc(0.75rem + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.98); border-top: 1px solid #e2e8f0;
  backdrop-filter: blur(12px);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.visitresult-submit-btn { width: 100%; }
.visitresult-bottom-hint { margin: 0; text-align: center; color: var(--text-muted); font-size: 0.68rem; }

/* ── Desktop ──────────────────────────────────────────── */
@media (min-width: 1024px) {
  .visitresult-page {
    padding-bottom: 2rem;
    max-width: 1280px;
  }

  .visitresult-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 0.9fr) minmax(380px, 1.1fr);
    gap: 1.25rem;
    align-items: start;
  }

  .visitresult-main-column,
  .visitresult-side-column {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    min-width: 0;
  }

  .visitresult-bottom {
    position: sticky;
    top: 1rem;
    width: 100%;
    z-index: 2;
  }

  .visitresult-submit-btn { width: 100%; }
}

/* ── Mobile ──────────────────────────────────────────── */
@media (max-width: 767px) {
  .visitresult-page { gap: 0.7rem; }
  .vrcard { padding: 1rem; }
  .vrcard-identity h1 { font-size: 1rem; }
}
</style>
