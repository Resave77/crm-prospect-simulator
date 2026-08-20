<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { checkOutProspect, transitionProspect } from '../../../api/crm'
import { useVisitLocation, type GpsCoords } from '../../../composables/sales/useVisitLocation'
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
const activeVisit = ref<{ id: string; checkInAt: string; checkInLatitude: number; checkInLongitude: number } | null>(null)
const pageState = ref<PageState>('loading')
const pageError = ref('')
const submitBusy = ref(false)
const autoCheckoutStarted = ref(false)
const CHECKOUT_TOLERANCE_METERS = 500

const location = useVisitLocation()

async function autoCheckOutIfOutside(coords: GpsCoords) {
  if (autoCheckoutStarted.value || !activeVisit.value || pageState.value !== 'ready') return
  const distance = location.distanceTo(activeVisit.value.checkInLatitude, activeVisit.value.checkInLongitude)
  if (distance === null || distance <= CHECKOUT_TOLERANCE_METERS || coords.accuracy > CHECKOUT_TOLERANCE_METERS / 2) return
  autoCheckoutStarted.value = true
  try {
    await checkOutProspect(entity.value!.entityId, activeVisit.value.id, {
      latitude: coords.latitude, longitude: coords.longitude,
      followUpNotes: '', visitResult: '', visitOutcome: '', autoCheckOut: true,
    })
    localStorage.removeItem(localStorageKey())
    router.replace({ name: resolvedEntityType.value === 'customer' ? 'SalesCustomerCheckOutSuccess' : 'SalesProspectCheckOutSuccess', params: { id: entity.value!.entityId } })
  } catch (caught) {
    autoCheckoutStarted.value = false
    pageError.value = formatErrorMessage(caught)
  }
}

watch(() => location.state.value.coords, (coords) => {
  if (coords) void autoCheckOutIfOutside(coords)
})

const storedVisitResult = ref<{ visitResult: string; visitOutcome: string; followUpNotes: string; followUpDate: string } | null>(null)
const elapsed = ref('—')
let elapsedTimer: ReturnType<typeof setInterval> | null = null

const hasValidGps = computed(() =>
  Boolean(
    location.state.value.coords &&
    Number.isFinite(location.state.value.coords.latitude) &&
    Number.isFinite(location.state.value.coords.longitude)
  )
)

const canCheckOut = computed(() =>
  Boolean(
    activeVisit.value &&
    storedVisitResult.value &&
    hasValidGps.value &&
    !submitBusy.value
  )
)

const checkoutButtonLabel = computed(() => {
  if (submitBusy.value) return 'Saving check-out\u2026'
  return 'Confirm Check-Out'
})

const checkoutHelper = computed(() => {
  if (!activeVisit.value) return 'No active visit found'
  if (!storedVisitResult.value) return 'No visit result found. Go back and complete the visit result form first.'
  if (!hasValidGps.value) return 'Current location is required'
  return 'Ready to check out'
})

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
  router.push(resolvedEntityType.value === 'customer' ? '/sales/my-customers' : '/sales/pipeline')
}

function localStorageKey() {
  return `visit-result-${resolvedEntityType.value}-${resolvedEntityId.value}`
}

function loadStoredVisitResult() {
  try {
    const raw = localStorage.getItem(localStorageKey())
    if (raw) {
      storedVisitResult.value = JSON.parse(raw)
    }
  } catch {
    storedVisitResult.value = null
  }
}

function goBackToVisitResult() {
  router.push(
    resolvedEntityType.value === 'customer'
      ? { name: 'SalesCustomerVisitResult', params: { id: resolvedEntityId.value } }
      : { name: 'SalesProspectVisitResult', params: { id: resolvedEntityId.value } }
  )
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

      activeVisit.value = { id: open.id, checkInAt: open.checkInAt, checkInLatitude: open.checkInLatitude, checkInLongitude: open.checkInLongitude }
      startElapsedTimer()

      if (entity.value.latitude == null || entity.value.longitude == null) {
        pageState.value = 'location-unavailable'
        return
      }

      pageState.value = 'ready'
      location.startWatching()
      loadStoredVisitResult()
    } else {
      const { entity: ctx, sourceProspectId } = await fetchCustomerVisitData(resolvedEntityId.value)
      entity.value = ctx

      if (!entity.value) {
        pageState.value = 'not-found'
        return
      }

      const { review } = await fetchProspectVisitData(sourceProspectId)
      const open = review.visits.find((v) => !v.checkOutAt)
      if (!open) {
        pageState.value = 'no-active-visit'
        return
      }

      activeVisit.value = { id: open.id, checkInAt: open.checkInAt, checkInLatitude: open.checkInLatitude, checkInLongitude: open.checkInLongitude }
      startElapsedTimer()

      if (entity.value.latitude == null || entity.value.longitude == null) {
        pageState.value = 'location-unavailable'
        return
      }

      pageState.value = 'ready'
      location.startWatching()
      loadStoredVisitResult()
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

async function handleSubmit() {
  if (!entity.value || !activeVisit.value || !canCheckOut.value) return
  submitBusy.value = true
  pageError.value = ''

  try {
    const coords = await location.refreshOnce().catch(() => location.state.value.coords)
    const lat = coords?.latitude ?? 0
    const lng = coords?.longitude ?? 0
    const stored = storedVisitResult.value

    if (entity.value.entityType === 'prospect') {
      await checkOutProspect(entity.value.entityId, activeVisit.value.id, {
        latitude: lat,
        longitude: lng,
        followUpNotes: stored?.followUpNotes ?? '',
        visitResult: stored?.visitResult ?? '',
        visitOutcome: stored?.visitOutcome ?? '',
      })

      try {
        await transitionProspect(entity.value.entityId, 'CONTACTED', '')
      } catch {
        // Transition may fail if already CONTACTED or beyond
      }

      localStorage.removeItem(localStorageKey())

      router.replace({
        name: 'SalesProspectCheckOutSuccess',
        params: { id: entity.value.entityId },
      })
    } else {
      await checkOutProspect(entity.value.entityId, activeVisit.value.id, {
        latitude: lat,
        longitude: lng,
        followUpNotes: stored?.followUpNotes ?? '',
        visitResult: stored?.visitResult ?? '',
        visitOutcome: stored?.visitOutcome ?? '',
      })

      localStorage.removeItem(localStorageKey())

      router.replace({
        name: 'SalesCustomerCheckOutSuccess',
        params: { id: entity.value.entityId },
      })
    }
  } catch (caught) {
    pageError.value = formatErrorMessage(caught)
  } finally {
    submitBusy.value = false
  }
}

onMounted(initialize)

onBeforeUnmount(() => {
  if (elapsedTimer) clearInterval(elapsedTimer)
  location.stopWatching()
})
</script>

<template>
  <section class="checkout-page">
    <button class="back-link" @click="pageState === 'loading' ? goBackToList() : goBack()"><i class="pi pi-arrow-left" /></button>

    <Message v-if="pageError" severity="error" closable @close="pageError = ''">{{ pageError }}</Message>

    <!-- Loading -->
    <div v-if="pageState === 'loading'" class="checkout-skeleton">
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-line w40" /><div class="sk-line w80" /></div>
    </div>

    <!-- Invalid params -->
    <div v-else-if="pageState === 'invalid-params'" class="checkout-empty">
      <div class="checkout-empty-icon"><i class="pi pi-exclamation-triangle" /></div>
      <strong>Invalid visit parameters</strong>
      <p>The visit link is missing a valid entity type or ID.</p>
      <button class="checkout-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- Not found -->
    <div v-else-if="pageState === 'not-found'" class="checkout-empty">
      <div class="checkout-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>{{ resolvedEntityType === 'customer' ? 'Customer' : 'Prospect' }} not found</strong>
      <p>This entity may have been removed or you don't have access.</p>
      <button class="checkout-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- No active visit -->
    <div v-else-if="pageState === 'no-active-visit'" class="checkout-empty">
      <div class="checkout-empty-icon"><i class="pi pi-sign-in" /></div>
      <strong>{{ resolvedEntityType === 'customer' ? 'Customer checkout not available' : 'No active visit' }}</strong>
      <p>{{ resolvedEntityType === 'customer' ? 'Customer visit check-out is not yet supported by the backend.' : 'You need to check in before you can check out.' }}</p>
      <button class="checkout-empty-btn" @click="goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <!-- Load error -->
    <div v-else-if="pageState === 'error'" class="checkout-empty">
      <div class="checkout-empty-icon"><i class="pi pi-exclamation-circle" /></div>
      <strong>Unable to load visit details</strong>
      <p>{{ pageError }}</p>
      <button class="checkout-empty-btn" @click="initialize"><i class="pi pi-refresh" /> Try again</button>
    </div>

    <!-- Location unavailable -->
    <div v-else-if="pageState === 'location-unavailable' && entity" class="checkout-empty">
      <div class="checkout-empty-icon"><i class="pi pi-map-marker" /></div>
      <strong>Visit location unavailable</strong>
      <p>The assigned location does not have valid coordinates.</p>
      <div class="checkout-empty-entity">{{ entity.name }}</div>
      <button class="checkout-empty-btn" @click="goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <template v-else-if="pageState === 'ready' && entity && activeVisit">
      <!-- Visit Summary (full width) -->
      <div class="cocard cocard-summary">
        <div class="cocard-summary-top">
          <div class="cocard-avatar">{{ entity.name.split(/\s+/).slice(0, 2).map(w => w.charAt(0).toUpperCase()).join('') }}</div>
          <div class="cocard-identity">
            <p class="eyebrow">Check out — {{ entity.entityType === 'prospect' ? 'Prospect' : 'Customer' }}</p>
            <h1>{{ entity.name }}</h1>
          </div>
          <Tag value="In Progress" severity="warn" />
        </div>
        <div class="cocard-stats">
          <div class="cocard-stat"><span>Check-in Time</span><strong>{{ new Date(activeVisit.checkInAt).toLocaleString() }}</strong></div>
          <div class="cocard-stat"><span>Duration</span><strong>{{ elapsed }}</strong></div>
        </div>
      </div>

      <!-- Two-column content grid -->
      <div class="checkout-content-grid">
        <!-- Left column: Location/map -->
        <div class="checkout-main-column">
          <div class="cocard">
            <div class="cocard-header-row">
              <h2>Location</h2>
              <button class="cocard-refresh-btn" :disabled="location.state.value.loading" @click="location.refreshOnce()" title="Refresh location">
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
            <div class="cocard-location-rows">
              <div v-if="location.state.value.coords" class="cocard-row">
                <i class="pi pi-map-marker" />
                <span>Current: {{ location.state.value.coords.latitude.toFixed(6) }}, {{ location.state.value.coords.longitude.toFixed(6) }}</span>
              </div>
              <div v-if="location.state.value.coords && entity.latitude != null" class="cocard-row cocard-distance">
                <i class="pi pi-compass" />
                <span>{{ location.distanceFormatted(entity.latitude, entity.longitude!) }} from target</span>
              </div>
              <div v-if="location.state.value.coords" class="cocard-row">
                <Tag
                  :value="insideRadius ? 'Inside radius' : 'Outside radius'"
                  :severity="insideRadius ? 'success' : 'warn'"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Right column: Visit result summary + submit -->
        <div class="checkout-side-column">
          <div v-if="storedVisitResult" class="cocard">
            <h2>Visit Result</h2>
            <div class="cocard-summary-rows">
              <div class="cocard-summary-row">
                <span>Result</span>
                <strong>{{ storedVisitResult.visitResult }}</strong>
              </div>
              <div class="cocard-summary-row">
                <span>Outcome</span>
                <strong>{{ storedVisitResult.visitOutcome }}</strong>
              </div>
              <div v-if="storedVisitResult.followUpNotes" class="cocard-summary-row">
                <span>Notes</span>
                <strong>{{ storedVisitResult.followUpNotes }}</strong>
              </div>
              <div v-if="storedVisitResult.followUpDate" class="cocard-summary-row">
                <span>Follow-Up Date</span>
                <strong>{{ new Date(storedVisitResult.followUpDate).toLocaleDateString() }}</strong>
              </div>
            </div>
            <button class="cocard-edit-btn" @click="goBackToVisitResult">
              <i class="pi pi-pencil" /> Edit visit result
            </button>
          </div>

          <div v-else class="cocard cocard-missing">
            <div class="cocard-missing-icon"><i class="pi pi-exclamation-circle" /></div>
            <strong>No visit result found</strong>
            <p>You must complete the visit result form before checking out.</p>
            <button class="checkout-empty-btn" @click="goBackToVisitResult"><i class="pi pi-arrow-left" /> Fill visit result</button>
          </div>

          <!-- Submit Action -->
          <div class="checkout-bottom">
            <Button
              :label="checkoutButtonLabel"
              icon="pi pi-check-circle"
              :loading="submitBusy"
              :disabled="!canCheckOut"
              class="checkout-submit-btn"
              @click="handleSubmit"
            />
            <p class="checkout-bottom-hint">{{ checkoutHelper }}</p>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.checkout-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: calc(68px + 92px + env(safe-area-inset-bottom) + 1rem); }

.checkout-skeleton { display: grid; gap: 0.85rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w40 { width: 40%; }
.sk-line.w50 { width: 50%; }
.sk-line.w60 { width: 60%; }
.sk-line.w80 { width: 80%; }

.checkout-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.checkout-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.checkout-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.checkout-empty p { margin: 0; color: var(--text-muted); font-size: 0.8rem; max-width: 280px; }
.checkout-empty-entity { color: var(--text-secondary); font-size: 0.85rem; font-weight: 600; }
.checkout-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; border: 0; cursor: pointer; }

.cocard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.cocard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

.cocard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.cocard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.cocard-avatar {
  width: 48px; height: 48px; display: grid; place-items: center; border-radius: 14px;
  background: linear-gradient(135deg, #e63946, #d62839); color: #fff; font-weight: 800;
  font-size: 0.9rem; flex-shrink: 0; box-shadow: 0 3px 10px rgba(230, 57, 70, 0.25);
}
.cocard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.cocard-identity .eyebrow { margin: 0; }
.cocard-identity h1 { margin: 0; font-size: 1.1rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

.cocard-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.6rem; }
.cocard-stat { padding: 0.75rem; border-radius: 12px; background: #f8fafc; }
.cocard-stat span { display: block; color: var(--text-muted); font-size: 0.6rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 0.15rem; }
.cocard-stat strong { font-size: 0.82rem; color: var(--text-primary); }

.cocard-header-row { display: flex; align-items: center; justify-content: space-between; }
.cocard-header-row h2 { margin: 0; }
.cocard-refresh-btn {
  width: 28px; height: 28px; border-radius: 8px; border: 1px solid var(--border-light);
  background: #fff; color: var(--text-muted); cursor: pointer; display: grid; place-items: center;
  font-size: 0.72rem; transition: all 0.15s ease;
}
.cocard-refresh-btn:hover { color: var(--brand-blue); border-color: #f4b3ba; background: #fff0f1; }
.cocard-refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.cocard-location-rows { display: grid; gap: 0.4rem; }
.cocard-row { display: flex; align-items: flex-start; gap: 0.5rem; color: var(--text-secondary); font-size: 0.78rem; line-height: 1.4; }
.cocard-row i { color: var(--text-muted); font-size: 0.68rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.15rem; }
.cocard-distance { color: var(--brand-blue); font-weight: 600; }

.cocard-form { display: grid; gap: 0.75rem; }
.cocard-summary-rows { display: grid; gap: 0.5rem; }
.cocard-summary-row { display: flex; flex-direction: column; gap: 0.1rem; }
.cocard-summary-row span { color: var(--text-muted); font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.cocard-summary-row strong { color: var(--text-primary); font-size: 0.82rem; font-weight: 600; }
.cocard-edit-btn {
  display: inline-flex; align-items: center; gap: 0.35rem;
  padding: 0.5rem 0.85rem; border-radius: 10px; border: 1px solid var(--border-light);
  background: #fff; color: var(--brand-blue); font-size: 0.75rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease; margin-top: 0.25rem;
}
.cocard-edit-btn:hover { background: var(--brand-blue-bg); border-color: #f4b3ba; }
.cocard-missing { text-align: center; }
.cocard-missing-icon { width: 40px; height: 40px; display: grid; place-items: center; border-radius: 12px; background: #fef2f2; color: #dc2626; font-size: 1rem; margin: 0 auto 0.5rem; }
.cocard-missing strong { display: block; font-size: 0.85rem; color: var(--text-primary); }
.cocard-missing p { margin: 0.25rem 0 0; font-size: 0.78rem; color: var(--text-muted); }

.checkout-bottom {
  position: fixed; bottom: 68px; left: 0; right: 0;
  width: 100%; z-index: 50;
  display: flex; flex-direction: column; gap: 0.3rem;
  padding: 0.75rem 1rem; padding-bottom: calc(0.75rem + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.98); border-top: 1px solid #e2e8f0;
  backdrop-filter: blur(12px);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.checkout-submit-btn { width: 100%; }
.checkout-bottom-hint { margin: 0; text-align: center; color: var(--text-muted); font-size: 0.68rem; }

/* ── Desktop ──────────────────────────────────────────── */
@media (min-width: 1024px) {
  .checkout-page {
    padding-bottom: 2rem;
    max-width: 1280px;
  }

  .checkout-content-grid {
    display: grid;
    grid-template-columns: minmax(0, 0.9fr) minmax(380px, 1.1fr);
    gap: 1.25rem;
    align-items: start;
  }

  .checkout-main-column,
  .checkout-side-column {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    min-width: 0;
  }

  .checkout-bottom {
    position: sticky;
    top: 1rem;
    width: 100%;
    z-index: 2;
  }

  .checkout-submit-btn { width: 100%; }

  .visit-location-map,
  .entity-map-frame,
  .vlm-wrapper { height: 320px !important; max-height: 360px; }
}

/* ── Mobile ──────────────────────────────────────────── */
@media (max-width: 767px) {
  .checkout-page { gap: 0.7rem; }
  .cocard { padding: 1rem; }
  .cocard-identity h1 { font-size: 1rem; }
}
</style>
