<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import type { VisitMonitoringItem } from '../../../types/crm'

const props = defineProps<{ item: VisitMonitoringItem | null }>()
const emit = defineEmits<{ close: [] }>()
const router = useRouter()
const apiBase = import.meta.env.VITE_API_BASE_URL || ''

const visible = computed({
  get: () => props.item !== null,
  set: (v: boolean) => { if (!v) emit('close') },
})

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatDateShort(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function formatDuration(seconds: number | undefined): string {
  if (seconds == null) return '—'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

function formatDistance(meters: number): string {
  if (meters >= 1000) return `${(meters / 1000).toFixed(1)} km`
  return `${Math.round(meters)} m`
}

function statusSeverity(status: string): string {
  switch (status) {
    case 'WON': case 'CONVERTED': return 'success'
    case 'LOST': return 'danger'
    case 'NEW_LEAD': return 'info'
    case 'CONTACTED': case 'INTERESTED': return 'warn'
    default: return 'secondary'
  }
}

function statusLabel(status: string): string {
  return status.replace(/_/g, ' ')
}

function goToProspect() {
  if (!props.item) return
  emit('close')
  router.push({ name: 'AdminProspectReview', params: { id: props.item.prospectId } })
}

function downloadEvidence() {
  if (!props.item) return
  const data = {
    evidenceId: props.item.id,
    prospectId: props.item.prospectId,
    salesExecutive: props.item.salesExecutiveName,
    customer: props.item.customerName,
    industryGroup: props.item.industryGroup,
    customerCategory: props.item.customerCategory,
    formattedAddress: props.item.formattedAddress,
    phoneNumber: props.item.phoneNumber,
    checkInAt: props.item.checkInAt,
    checkOutAt: props.item.checkOutAt,
    checkInCoordinates: { lat: props.item.checkInLatitude, lng: props.item.checkInLongitude },
    distanceMeters: props.item.distanceMeters,
    radiusStatus: props.item.radiusStatus,
    prospectStatus: props.item.prospectStatus,
    visitNotes: props.item.visitNotes,
    followUpNotes: props.item.followUpNotes,
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `evidence-${props.item.id.slice(0, 8)}.json`
  a.click()
  URL.revokeObjectURL(url)
}

function viewGps() {
  if (!props.item) return
  window.open(`https://www.google.com/maps?q=${props.item.checkInLatitude},${props.item.checkInLongitude}`, '_blank')
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    header="Visit Evidence"
    :style="{ width: '480px' }"
    :closable="true"
    :breakpoints="{ '576px': '95vw' }"
    class="selfie-dialog"
  >
    <div v-if="item" class="selfie-body">
      <div class="selfie-preview">
        <img
          v-if="item.selfieReference && !item.selfieReference.startsWith('SIMULATED')"
          :src="item.selfieReference.startsWith('/') ? `${apiBase}${item.selfieReference}` : item.selfieReference"
          alt="Visit selfie"
          class="selfie-image"
        />
        <div v-else class="selfie-placeholder">
          <i class="pi pi-camera" />
          <span>Selfie captured</span>
          <small class="capture-time">captured at {{ formatTime(item.checkInAt) }}</small>
        </div>
      </div>

      <div class="selfie-info">
        <div class="info-section">
          <div class="info-section-header">Prospect Details</div>
          <div class="info-row">
            <span class="info-label">Customer</span>
            <div class="info-value-col">
              <strong class="prospect-link" @click="goToProspect">{{ item.customerName }}</strong>
              <small>{{ item.industryGroup }} · {{ item.customerCategory }}</small>
            </div>
          </div>
          <div v-if="item.formattedAddress" class="info-row">
            <span class="info-label">Address</span>
            <span class="info-text-wrap">{{ item.formattedAddress }}</span>
          </div>
          <div v-if="item.phoneNumber" class="info-row">
            <span class="info-label">Phone</span>
            <span class="mono">{{ item.phoneNumber }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Status</span>
            <Tag :value="statusLabel(item.prospectStatus)" :severity="statusSeverity(item.prospectStatus)" size="small" />
          </div>
          <div class="info-row">
            <span class="info-label">Total Visits</span>
            <strong>{{ item.visitCount }}</strong>
          </div>
        </div>

        <div class="info-section">
          <div class="info-section-header">Visit Details</div>
          <div class="info-row">
            <span class="info-label">Sales Executive</span>
            <strong>{{ item.salesExecutiveName }}</strong>
          </div>
          <div class="info-row">
            <span class="info-label">Check In</span>
            <span class="mono">{{ formatDateShort(item.checkInAt) }} {{ formatTime(item.checkInAt) }}</span>
          </div>
          <div v-if="item.checkOutAt" class="info-row">
            <span class="info-label">Check Out</span>
            <span class="mono">{{ formatDateShort(item.checkOutAt) }} {{ formatTime(item.checkOutAt) }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Duration</span>
            <strong>{{ formatDuration(item.durationSeconds) }}</strong>
          </div>
          <div class="info-row">
            <span class="info-label">Distance</span>
            <strong>{{ formatDistance(item.distanceMeters) }}</strong>
          </div>
          <div class="info-row">
            <span class="info-label">Radius</span>
            <Tag
              :value="item.radiusStatus === 'INSIDE' ? 'Inside Radius' : item.radiusStatus === 'OUTSIDE' ? 'Outside Radius' : 'Unknown'"
              :severity="item.radiusStatus === 'INSIDE' ? 'success' : item.radiusStatus === 'OUTSIDE' ? 'danger' : 'secondary'"
              size="small"
            />
          </div>
        </div>

        <div v-if="item.visitNotes || item.followUpNotes" class="info-section">
          <div class="info-section-header">Notes</div>
          <div v-if="item.visitNotes" class="note-block">
            <span class="info-label">Visit Notes</span>
            <p>{{ item.visitNotes }}</p>
          </div>
          <div v-if="item.followUpNotes" class="note-block">
            <span class="info-label">Follow Up Notes</span>
            <p>{{ item.followUpNotes }}</p>
          </div>
        </div>

        <div class="info-section">
          <div class="info-section-header">Evidence</div>
          <div class="info-row">
            <span class="info-label">Evidence ID</span>
            <code class="evidence-id">{{ item.id.slice(0, 8) }}…</code>
          </div>
          <div class="info-row">
            <span class="info-label">GPS Coordinates</span>
            <span class="mono">{{ item.checkInLatitude.toFixed(6) }}, {{ item.checkInLongitude.toFixed(6) }}</span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-actions">
        <Button label="View Prospect" icon="pi pi-external-link" severity="secondary" outlined @click="goToProspect" />
        <Button label="Download" icon="pi pi-download" severity="secondary" outlined @click="downloadEvidence" />
        <Button label="View GPS" icon="pi pi-map" @click="viewGps" />
      </div>
    </template>
  </Dialog>
</template>

<style scoped>
.selfie-preview {
  width: 100%; aspect-ratio: 4/3; border-radius: 12px; overflow: hidden;
  background: #111827; border: 1px solid var(--border-light); margin-bottom: 1rem;
}
.selfie-placeholder {
  width: 100%; height: 100%; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 0.5rem; color: #64748b;
}
.selfie-placeholder i { font-size: 2.5rem; color: #475569; }
.selfie-placeholder span { font-size: 0.85rem; font-weight: 600; }
.selfie-image { width: 100%; height: 100%; object-fit: cover; }
.capture-time { font-size: 0.75rem; color: #94a3b8; font-family: 'SF Mono', 'Fira Code', monospace; }

.selfie-info { display: flex; flex-direction: column; gap: 0.75rem; }

.info-section {
  padding: 0.75rem; background: var(--surface-subtle, #f8fafc);
  border-radius: 10px; display: flex; flex-direction: column; gap: 0.5rem;
}
.info-section-header {
  font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--brand-green-light, #0b7766); padding-bottom: 0.25rem;
  border-bottom: 1px solid var(--border-light, #e2e8f0);
}
.info-row { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
.info-label { font-size: 0.68rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text-muted); flex-shrink: 0; }
.info-value-col { display: grid; text-align: right; }
.info-value-col strong { font-size: 0.82rem; }
.info-value-col small { font-size: 0.68rem; color: var(--text-muted); }
.info-text-wrap { font-size: 0.75rem; color: var(--text-primary); text-align: right; max-width: 260px; word-break: break-word; }

.prospect-link { color: #2563eb; cursor: pointer; }
.prospect-link:hover { text-decoration: underline; }

.note-block { display: flex; flex-direction: column; gap: 0.25rem; }
.note-block p { font-size: 0.78rem; color: var(--text-primary); margin: 0; line-height: 1.5; }

.mono { font-size: 0.78rem; font-family: 'SF Mono', 'Fira Code', monospace; color: var(--text-primary); }
.evidence-id {
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: 0.78rem;
  padding: 0.1rem 0.4rem; border-radius: 4px; background: #e2e8f0;
  color: var(--text-secondary);
}

.dialog-actions { display: flex; gap: 0.5rem; justify-content: flex-end; width: 100%; }
.dialog-actions :deep(.p-button) { flex: 1; }
</style>
