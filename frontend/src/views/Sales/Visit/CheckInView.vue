<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { checkInProspect } from '../../../api/crm'
import { useVisitLocation } from '../../../composables/sales/useVisitLocation'
import {
  isValidEntityType,
  normalizeRouteId,
  fetchProspectVisitData,
  fetchCustomerVisitData,
  type VisitEntityContext,
} from '../../../utils/visitEntity'
import VisitLocationCard from '../../../components/sales/visit/VisitLocationCard.vue'
import VisitSelfieCapture from '../../../components/sales/visit/VisitSelfieCapture.vue'
import type { WatermarkMeta } from '../../../utils/selfieWatermark'

// Demo-only client-side override. Production must enforce radius on backend.
function envFlag(value: unknown): boolean {
  return String(value ?? '').trim().toLowerCase() === 'true'
}

const demoRadiusOverrideEnabled = envFlag(import.meta.env.VITE_DEMO_RADIUS_OVERRIDE)

console.debug('[DemoRadiusOverride]', {
  raw: import.meta.env.VITE_DEMO_RADIUS_OVERRIDE,
  enabled: demoRadiusOverrideEnabled,
  mode: import.meta.env.MODE,
})

type PageState = 'loading' | 'ready' | 'invalid-params' | 'not-found' | 'error' | 'location-unavailable'

const route = useRoute()
const router = useRouter()

const resolvedEntityType = computed(() => {
  if (isValidEntityType(route.meta.entityType)) return route.meta.entityType
  return null
})
const resolvedEntityId = computed(() => normalizeRouteId(route.params.id))

const entity = ref<VisitEntityContext | null>(null)
const pageState = ref<PageState>('loading')
const pageError = ref('')
const submitBusy = ref(false)
const visitNotes = ref('')
const selfieFile = ref<File | null>(null)
const showDemoConfirm = ref(false)

const location = useVisitLocation()

const gpsAccuracyWarning = computed(() => {
  const c = location.state.value.coords
  if (!c || !entity.value) return false
  return c.accuracy > entity.value.attendanceRadiusMeters * 0.5
})

const hasEntity = computed(() =>
  Boolean(entity.value?.entityId)
)

const hasValidGps = computed(() =>
  Boolean(
    location.state.value.coords &&
    Number.isFinite(location.state.value.coords.latitude) &&
    Number.isFinite(location.state.value.coords.longitude)
  )
)

const hasSelfie = computed(() =>
  selfieFile.value instanceof File &&
  selfieFile.value.size > 0
)

const insideRadius = computed(() => {
  if (!entity.value || !location.state.value.coords) return false
  if (entity.value.latitude == null || entity.value.longitude == null) return false
  return location.isInsideRadius(
    entity.value.latitude,
    entity.value.longitude,
    entity.value.attendanceRadiusMeters,
  )
})

const radiusAllowedForSubmit = computed(() =>
  insideRadius.value || demoRadiusOverrideEnabled
)

const canCheckIn = computed(() =>
  hasEntity.value &&
  hasValidGps.value &&
  hasSelfie.value &&
  radiusAllowedForSubmit.value &&
  !submitBusy.value
)

watchEffect(() => {
  console.debug('[CheckInEligibility]', {
    hasEntity: hasEntity.value,
    hasValidGps: hasValidGps.value,
    hasSelfie: hasSelfie.value,
    selfieSize: selfieFile.value?.size ?? 0,
    insideRadius: insideRadius.value,
    demoOverride: demoRadiusOverrideEnabled,
    radiusAllowed: radiusAllowedForSubmit.value,
    submitBusy: submitBusy.value,
    canCheckIn: canCheckIn.value,
  })
})

const usingDemoRadiusOverride = computed(() =>
  demoRadiusOverrideEnabled &&
  hasValidGps.value &&
  !insideRadius.value
)

const submitLabel = computed(() => {
  if (submitBusy.value) return 'Checking in\u2026'
  if (usingDemoRadiusOverride.value) return 'Demo Check in'
  return 'Check in'
})

const submitHelper = computed(() => {
  if (!hasValidGps.value) return 'GPS location required'
  if (!hasSelfie.value) return 'Selfie required'
  if (!insideRadius.value && !demoRadiusOverrideEnabled) {
    return 'Outside radius'
  }
  if (usingDemoRadiusOverride.value) {
    return 'Simulation mode \u2022 Outside radius override'
  }
  return 'GPS verified \u2022 Ready to check in'
})

const selfieMeta = computed<WatermarkMeta | undefined>(() => {
  if (!entity.value) return undefined
  const c = location.state.value.coords
  return {
    salesName: '',
    entityType: entity.value.entityType as 'prospect' | 'customer',
    entityName: entity.value.name,
    latitude: c?.latitude ?? 0,
    longitude: c?.longitude ?? 0,
    accuracyMeters: c?.accuracy ?? 0,
    insideRadius: insideRadius.value,
  }
})

const gpsAgeLabel = computed(() => {
  const c = location.state.value.coords
  if (!c) return ''
  const secs = Math.round((Date.now() - c.capturedAt) / 1000)
  if (secs < 5) return 'Just now'
  if (secs < 60) return `${secs}s ago`
  return `${Math.floor(secs / 60)}m ${secs % 60}s ago`
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

function extractError(err: unknown): string {
  return (
    (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message ??
    (err instanceof Error ? err.message : 'Unable to complete the request.')
  )
}

function onSelfieCaptured(file: File) {
  selfieFile.value = file
}

function onSelfieCleared() {
  selfieFile.value = null
}

function onSelfieError(message: string) {
  pageError.value = message
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
      const { entity: ctx } = await fetchProspectVisitData(resolvedEntityId.value)
      entity.value = ctx
    } else {
      const { entity: ctx } = await fetchCustomerVisitData(resolvedEntityId.value)
      entity.value = ctx
    }

    if (!entity.value) {
      pageState.value = 'not-found'
      return
    }

    if (entity.value.latitude == null || entity.value.longitude == null) {
      pageState.value = 'location-unavailable'
      return
    }

    pageState.value = 'ready'
    location.startWatching()
  } catch (caught) {
    const status = (caught as { response?: { status?: number } })?.response?.status
    if (status === 404) {
      pageState.value = 'not-found'
    } else {
      pageError.value = extractError(caught)
      pageState.value = 'error'
    }
  }
}

async function refreshGps() {
  try {
    await location.refreshOnce()
  } catch {
    // error already set in location state
  }
}

function requestCheckIn() {
  if (!canCheckIn.value) return

  if (usingDemoRadiusOverride.value) {
    showDemoConfirm.value = true
    return
  }

  void submitCheckIn()
}

async function confirmDemoOverride() {
  showDemoConfirm.value = false
  await submitCheckIn()
}

function cancelDemoOverride() {
  showDemoConfirm.value = false
}

async function submitCheckIn() {
  if (!entity.value || !canCheckIn.value) return
  submitBusy.value = true
  pageError.value = ''

  try {
    const coords = location.state.value.coords!

    let notes = visitNotes.value

    if (!insideRadius.value && demoRadiusOverrideEnabled) {
      const c = location.state.value.coords
      const demoHeader = [
        '[DEMO RADIUS OVERRIDE]',
        `Distance: ${location.distanceFormatted(entity.value.latitude!, entity.value.longitude!)}`,
        `Allowed radius: ${entity.value.attendanceRadiusMeters}m`,
        c ? `GPS accuracy: \u00b1${Math.round(c.accuracy)}m` : '',
        'Inside radius: false',
        '',
      ].filter(Boolean).join('\n')
      notes = demoHeader + notes
    }

    if (entity.value.entityType === 'prospect') {
      await checkInProspect(entity.value.entityId, {
        latitude: coords.latitude,
        longitude: coords.longitude,
        visitNotes: notes,
        selfiePlaceholder: true,
      })
    }

    selfieFile.value = null

    router.push({
      name: resolvedEntityType.value === 'customer' ? 'SalesCustomerCheckOut' : 'SalesProspectCheckOut',
      params: { id: entity.value.entityId },
    })
  } catch (caught) {
    pageError.value = extractError(caught)
  } finally {
    submitBusy.value = false
  }
}

onMounted(initialize)

onBeforeUnmount(() => {
  location.stopWatching()
})
</script>

<template>
  <section class="checkin-page">
    <button class="back-link" @click="pageState === 'loading' ? goBackToList() : goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>

    <Message v-if="pageError" severity="error" closable @close="pageError = ''">{{ pageError }}</Message>

    <!-- Loading -->
    <div v-if="pageState === 'loading'" class="checkin-skeleton">
      <div class="sk-card"><div class="sk-line w60" /><div class="sk-line w80" /><div class="sk-line w50" /></div>
      <div class="sk-card"><div class="sk-map" /><div class="sk-line w70" /></div>
    </div>

    <!-- Invalid params -->
    <div v-else-if="pageState === 'invalid-params'" class="checkin-empty">
      <div class="checkin-empty-icon"><i class="pi pi-exclamation-triangle" /></div>
      <strong>Invalid visit parameters</strong>
      <p>The visit link is missing a valid entity type or ID.</p>
      <button class="checkin-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- Not found -->
    <div v-else-if="pageState === 'not-found'" class="checkin-empty">
      <div class="checkin-empty-icon"><i class="pi pi-inbox" /></div>
      <strong>{{ resolvedEntityType === 'customer' ? 'Customer' : 'Prospect' }} not found</strong>
      <p>This entity may have been removed or you don't have access.</p>
      <button class="checkin-empty-btn" @click="goBackToList"><i class="pi pi-arrow-left" /> Back to list</button>
    </div>

    <!-- Load error -->
    <div v-else-if="pageState === 'error'" class="checkin-empty">
      <div class="checkin-empty-icon"><i class="pi pi-exclamation-circle" /></div>
      <strong>Unable to load visit details</strong>
      <p>{{ pageError }}</p>
      <button class="checkin-empty-btn" @click="initialize"><i class="pi pi-refresh" /> Try again</button>
    </div>

    <!-- Location unavailable -->
    <div v-else-if="pageState === 'location-unavailable' && entity" class="checkin-empty">
      <div class="checkin-empty-icon"><i class="pi pi-map-marker" /></div>
      <strong>Visit location unavailable</strong>
      <p>The assigned location does not have valid coordinates.</p>
      <div class="checkin-empty-entity">{{ entity.name }}</div>
      <button class="checkin-empty-btn" @click="goBack()"><i class="pi pi-arrow-left" /> Back to detail</button>
    </div>

    <template v-else-if="pageState === 'ready' && entity">
      <!-- Entity Summary -->
      <div class="cicard cicard-summary">
        <div class="cicard-summary-top">
          <div class="cicard-avatar">{{ entity.name.split(/\s+/).slice(0, 2).map(w => w.charAt(0).toUpperCase()).join('') }}</div>
          <div class="cicard-identity">
            <p class="eyebrow">Check in — {{ entity.entityType === 'prospect' ? 'Prospect' : 'Customer' }}</p>
            <h1>{{ entity.name }}</h1>
          </div>
          <Tag v-if="entity.entityType === 'prospect'" :value="entity.status.replaceAll('_', ' ')" severity="info" />
          <Tag v-else value="ACTIVE" severity="success" />
        </div>
      </div>

      <!-- Demo Override Badge -->
      <div v-if="demoRadiusOverrideEnabled" class="cicard cicard-demo-badge">
        <div class="cicard-demo-badge-inner">
          <i class="pi pi-info-circle" />
          <span><strong>Demo mode active.</strong> Radius validation is not enforced. Location validation is for simulation only.</span>
        </div>
      </div>

      <!-- GPS Signal -->
      <div class="cicard">
        <div class="cicard-gps-header">
          <h2>GPS Signal</h2>
          <button class="cicard-refresh-btn" :disabled="location.state.value.loading" @click="refreshGps" title="Refresh location">
            <i class="pi" :class="location.state.value.loading ? 'pi-spin pi-sync' : 'pi-refresh'" />
          </button>
        </div>
        <div v-if="location.state.value.permissionGranted && location.state.value.coords" class="cicard-gps-status">
          <div class="cicard-gps-ready">
            <i class="pi pi-check-circle" /> GPS signal ready
          </div>
          <span class="cicard-gps-accuracy">Accuracy &plusmn;{{ Math.round(location.state.value.coords.accuracy) }} meters</span>
          <Tag
            :value="insideRadius ? 'Inside radius' : 'Outside radius'"
            :severity="insideRadius ? 'success' : 'danger'"
          />
        </div>
        <div v-else-if="location.state.value.error" class="cicard-gps-error">
          <i class="pi pi-exclamation-triangle" /> {{ location.state.value.error }}
        </div>
        <div v-else-if="location.state.value.loading" class="cicard-gps-loading">
          <i class="pi pi-spin pi-sync" /> Acquiring GPS signal...
        </div>
        <div v-else class="cicard-gps-idle">
          <i class="pi pi-info-circle" /> GPS will be acquired for check-in
        </div>
      </div>

      <!-- Location Status -->
      <div v-if="location.state.value.coords" class="cicard cicard-location-status">
        <div class="cicard-location-status-inner">
          <i class="pi pi-check-circle" />
          <span>Location verified &bull; Updated {{ gpsAgeLabel }}</span>
        </div>
      </div>

      <!-- Map + Location Details -->
      <div class="cicard">
        <h2>Current Location</h2>
        <VisitLocationCard
          :target-latitude="entity.latitude"
          :target-longitude="entity.longitude"
          :target-label="entity.name"
          :sales-coords="location.state.value.coords"
          :radius-meters="entity.attendanceRadiusMeters"
          height="190px"
        />
        <div class="cicard-location-rows">
          <div v-if="entity.latitude != null && entity.longitude != null" class="cicard-row">
            <i class="pi pi-compass" />
            <span>Target: {{ entity.latitude.toFixed(6) }}, {{ entity.longitude.toFixed(6) }}</span>
          </div>
          <div v-if="location.state.value.coords" class="cicard-row">
            <i class="pi pi-map-marker" />
            <span>Current: {{ location.state.value.coords.latitude.toFixed(6) }}, {{ location.state.value.coords.longitude.toFixed(6) }}</span>
          </div>
          <div v-if="location.state.value.coords && entity.latitude != null" class="cicard-row cicard-distance">
            <i class="pi pi-compass" />
            <span>{{ location.distanceFormatted(entity.latitude, entity.longitude!) }} from target</span>
          </div>
          <div class="cicard-row">
            <i class="pi pi-info-circle" />
            <span>Allowed radius: {{ entity.attendanceRadiusMeters }} meters</span>
          </div>
          <div v-if="location.state.value.coords" class="cicard-row">
            <i class="pi pi-info-circle" />
            <span>GPS accuracy: &plusmn;{{ Math.round(location.state.value.coords.accuracy) }} meters</span>
          </div>
          <div v-if="gpsAccuracyWarning" class="cicard-row cicard-warning">
            <i class="pi pi-exclamation-triangle" />
            <span>GPS accuracy is low relative to the radius. Position may be imprecise.</span>
          </div>
        </div>
      </div>

      <!-- Outside Radius Warning -->
      <div v-if="location.state.value.permissionGranted && location.state.value.coords && !insideRadius" class="cicard cicard-outside">
        <div class="cicard-outside-inner">
          <i class="pi pi-exclamation-triangle" />
          <div>
            <strong>Outside allowed radius</strong>
            <p>You are {{ location.distanceFormatted(entity.latitude!, entity.longitude!) }} from the target. Allowed radius is {{ entity.attendanceRadiusMeters }}m.</p>
            <p v-if="demoRadiusOverrideEnabled" class="cicard-outside-demo-hint">Demo mode: Check-in is permitted for simulation purposes.</p>
            <p v-else class="cicard-outside-hint">Move closer within {{ entity.attendanceRadiusMeters }}m to check in.</p>
          </div>
        </div>
      </div>

      <!-- Selfie -->
      <div class="cicard">
        <VisitSelfieCapture
          :watermark-meta="selfieMeta"
          :required="true"
          @captured="onSelfieCaptured"
          @cleared="onSelfieCleared"
          @error="onSelfieError"
        />
      </div>

      <!-- Visit Notes -->
      <div class="cicard">
        <h2>Visit Notes</h2>
        <p class="cicard-hint">Optional &mdash; what do you plan to discuss?</p>
        <label class="cicard-field">
          <span class="sr-only">Visit Notes</span>
          <Textarea v-model="visitNotes" rows="3" fluid placeholder="What do you plan to discuss?" />
        </label>
      </div>
    </template>

    <!-- Bottom Submit -->
    <div v-if="pageState === 'ready' && entity" class="checkin-bottom">
      <Button
        :label="submitLabel"
        icon="pi pi-sign-in"
        :loading="submitBusy"
        :disabled="!canCheckIn"
        class="checkin-submit-btn"
        @click="requestCheckIn"
      />
      <p class="checkin-bottom-hint">{{ submitHelper }}</p>
    </div>

    <!-- Demo Override Confirmation Dialog -->
    <Dialog
      v-model:visible="showDemoConfirm"
      header="Demo check-in outside radius"
      :modal="true"
      :closable="true"
      :style="{ width: 'min(100%, 420px)' }"
      @hide="cancelDemoOverride"
    >
      <div class="demo-confirm-body">
        <p>You are currently <strong>outside</strong> the allowed visit radius.</p>
        <p>This action is permitted only because <strong>simulation mode</strong> is enabled.</p>

        <div v-if="entity && location.state.value.coords" class="demo-confirm-details">
          <div class="demo-confirm-row">
            <span>Distance from target</span>
            <strong>{{ location.distanceFormatted(entity.latitude!, entity.longitude!) }}</strong>
          </div>
          <div class="demo-confirm-row">
            <span>Allowed radius</span>
            <strong>{{ entity.attendanceRadiusMeters }}m</strong>
          </div>
          <div class="demo-confirm-row">
            <span>GPS accuracy</span>
            <strong>&plusmn;{{ Math.round(location.state.value.coords.accuracy) }}m</strong>
          </div>
        </div>

        <p class="demo-confirm-note">This override is logged for audit. Production builds enforce radius on the server.</p>
      </div>
      <template #footer>
        <Button label="Cancel" severity="secondary" outlined @click="cancelDemoOverride" />
        <Button label="Continue Demo Check in" icon="pi pi-check" severity="warn" :loading="submitBusy" @click="confirmDemoOverride" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.checkin-page { display: grid; gap: 0.85rem; width: 100%; padding-bottom: calc(68px + 88px + env(safe-area-inset-bottom) + 1rem); }

.back-link {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0;
  border: 0; background: transparent; color: var(--brand-blue, #2563eb);
  font-size: 0.8rem; font-weight: 600; cursor: pointer; text-decoration: none;
}
.back-link:hover { text-decoration: underline; }

.checkin-skeleton { display: grid; gap: 0.85rem; }
.sk-card { padding: 1rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); display: flex; flex-direction: column; gap: 0.5rem; }
.sk-line { height: 12px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w50 { width: 50%; }
.sk-line.w60 { width: 60%; }
.sk-line.w70 { width: 70%; }
.sk-line.w80 { width: 80%; }
.sk-map { height: 180px; border-radius: 12px; background: #e2e8f0; }

.checkin-empty { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2.5rem 1rem; text-align: center; }
.checkin-empty-icon { width: 56px; height: 56px; display: grid; place-items: center; border-radius: 16px; background: #f1f5f9; color: #94a3b8; font-size: 1.4rem; }
.checkin-empty strong { color: var(--text-primary); font-size: 0.95rem; }
.checkin-empty p { margin: 0; color: var(--text-muted); font-size: 0.8rem; max-width: 280px; }
.checkin-empty-entity { color: var(--text-secondary); font-size: 0.85rem; font-weight: 600; }
.checkin-empty-btn { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.5rem 1rem; border-radius: 12px; background: var(--brand-blue); color: #fff; text-decoration: none; font-size: 0.8rem; font-weight: 600; margin-top: 0.5rem; border: 0; cursor: pointer; }

.cicard {
  padding: 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: var(--surface-card); box-shadow: var(--shadow-sm); display: grid; gap: 0.75rem;
}
.cicard h2 { margin: 0; font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted); }

.cicard-summary { background: linear-gradient(135deg, var(--brand-blue-50) 0%, var(--surface-card) 100%); }
.cicard-summary-top { display: flex; align-items: flex-start; gap: 0.85rem; }
.cicard-avatar {
  width: 48px; height: 48px; display: grid; place-items: center; border-radius: 14px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff; font-weight: 800;
  font-size: 0.9rem; flex-shrink: 0; box-shadow: 0 3px 10px rgba(37, 99, 235, 0.25);
}
.cicard-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.1rem; }
.cicard-identity .eyebrow { margin: 0; }
.cicard-identity h1 { margin: 0; font-size: 1.1rem; font-weight: 800; letter-spacing: -0.02em; color: var(--text-primary); line-height: 1.3; }

.cicard-demo-badge { border-color: #fbbf24; background: #fffbeb; }
.cicard-demo-badge-inner { display: flex; align-items: flex-start; gap: 0.5rem; color: #92400e; font-size: 0.78rem; line-height: 1.4; }
.cicard-demo-badge-inner i { color: #d97706; font-size: 0.85rem; flex-shrink: 0; margin-top: 0.1rem; }
.cicard-demo-badge-inner strong { color: #78350f; }

.cicard-gps-header { display: flex; align-items: center; justify-content: space-between; }
.cicard-gps-header h2 { margin: 0; }
.cicard-refresh-btn {
  width: 28px; height: 28px; border-radius: 8px; border: 1px solid var(--border-light);
  background: #fff; color: var(--text-muted); cursor: pointer; display: grid; place-items: center;
  font-size: 0.72rem; transition: all 0.15s ease;
}
.cicard-refresh-btn:hover { color: var(--brand-blue); border-color: #bfdbfe; background: #eff6ff; }
.cicard-refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.cicard-gps-status { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
.cicard-gps-ready { display: flex; align-items: center; gap: 0.35rem; color: #16a34a; font-size: 0.82rem; font-weight: 600; }
.cicard-gps-accuracy { color: var(--text-muted); font-size: 0.75rem; }

.cicard-gps-error { display: flex; align-items: flex-start; gap: 0.4rem; padding: 0.6rem 0.8rem; border-radius: 10px; background: #fef2f2; color: #991b1b; font-size: 0.78rem; line-height: 1.4; }
.cicard-gps-error i { flex-shrink: 0; margin-top: 0.1rem; }
.cicard-gps-loading { display: flex; align-items: center; gap: 0.4rem; color: var(--text-muted); font-size: 0.82rem; }
.cicard-gps-idle { color: var(--text-muted); font-size: 0.82rem; }

.cicard-location-status { padding: 0.65rem 1rem; background: #f0fdf4; border-color: #bbf7d0; }
.cicard-location-status-inner { display: flex; align-items: center; gap: 0.4rem; color: #166534; font-size: 0.78rem; }
.cicard-location-status-inner i { font-size: 0.85rem; }

.cicard-location-rows { display: grid; gap: 0.4rem; }
.cicard-row { display: flex; align-items: flex-start; gap: 0.5rem; color: var(--text-secondary); font-size: 0.78rem; line-height: 1.4; }
.cicard-row i { color: var(--text-muted); font-size: 0.68rem; width: 1rem; text-align: center; flex-shrink: 0; margin-top: 0.15rem; }
.cicard-distance { color: var(--brand-blue); font-weight: 600; }
.cicard-warning { color: #d97706; font-weight: 500; }

.cicard-outside { border-color: #fecaca; background: #fef2f2; }
.cicard-outside-inner { display: flex; gap: 0.6rem; align-items: flex-start; }
.cicard-outside-inner > i { color: #dc2626; font-size: 1.1rem; flex-shrink: 0; margin-top: 0.1rem; }
.cicard-outside-inner strong { color: #991b1b; font-size: 0.82rem; display: block; margin-bottom: 0.15rem; }
.cicard-outside-inner p { margin: 0; color: #7f1d1d; font-size: 0.75rem; line-height: 1.4; }
.cicard-outside-demo-hint { color: #92400e; font-weight: 600; margin-top: 0.25rem !important; }
.cicard-outside-hint { margin-top: 0.25rem !important; }

.cicard-hint { margin: -0.35rem 0 0; color: var(--text-muted); font-size: 0.75rem; }
.cicard-field { display: flex; flex-direction: column; gap: 0.3rem; }

.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }

.checkin-bottom {
  position: fixed; bottom: 68px; left: 50%; transform: translateX(-50%);
  width: min(100%, 440px); z-index: 50;
  display: flex; flex-direction: column; gap: 0.3rem;
  padding: 0.7rem 1rem; padding-bottom: calc(0.7rem + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.98); border-top: 1px solid #e2e8f0;
  backdrop-filter: blur(12px);
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.06);
}
.checkin-submit-btn { width: 100%; }
.checkin-bottom-hint { margin: 0; text-align: center; color: var(--text-muted); font-size: 0.68rem; }

.demo-confirm-body { display: grid; gap: 0.6rem; }
.demo-confirm-body > p { margin: 0; font-size: 0.85rem; line-height: 1.5; color: var(--text-primary); }
.demo-confirm-details {
  display: grid; gap: 0.4rem; padding: 0.75rem; border-radius: 10px;
  background: #f8fafc; border: 1px solid var(--border-light);
}
.demo-confirm-row { display: flex; justify-content: space-between; align-items: center; font-size: 0.8rem; }
.demo-confirm-row span { color: var(--text-muted); }
.demo-confirm-row strong { color: var(--text-primary); }
.demo-confirm-note { margin: 0; font-size: 0.72rem; color: var(--text-muted); font-style: italic; }

@media (max-width: 480px) {
  .checkin-page { gap: 0.7rem; }
  .cicard { padding: 1rem; }
  .cicard-identity h1 { font-size: 1rem; }
}
</style>
