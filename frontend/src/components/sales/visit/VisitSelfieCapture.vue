<script setup lang="ts">
import { ref, computed } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useVisitCamera } from '../../../composables/sales/useVisitCamera'
import type { WatermarkMeta } from '../../../utils/selfieWatermark'

const props = defineProps<{
  watermarkMeta?: WatermarkMeta
  required?: boolean
}>()

const emit = defineEmits<{
  captured: [file: File]
  cleared: []
  error: [message: string]
}>()

const videoEl = ref<HTMLVideoElement | null>(null)
const canvasEl = ref<HTMLCanvasElement | null>(null)

const camera = useVisitCamera({ videoEl, canvasEl })

const captureBusy = ref(false)

const showPreview = computed(() => camera.state.value.mode === 'captured' && camera.previewUrl.value)
const showVideo = computed(() => camera.state.value.mode === 'opening' || camera.state.value.mode === 'live')
const showPlaceholder = computed(() => camera.state.value.mode === 'idle' || camera.state.value.mode === 'error')
const showOpening = computed(() => camera.state.value.mode === 'opening')

async function handleOpenCamera() {
  const ok = await camera.openCamera()
  if (!ok && camera.state.value.error) {
    emit('error', camera.state.value.error)
  }
}

async function handleCapture() {
  if (!camera.cameraReady.value || captureBusy.value) return
  captureBusy.value = true
  try {
    const meta = props.watermarkMeta ?? {
      salesName: '',
      entityType: 'prospect' as const,
      entityName: '',
      latitude: 0,
      longitude: 0,
      accuracyMeters: 0,
      insideRadius: false,
    }
    const file = await camera.capturePhoto(meta)
    if (file) {
      emit('captured', file)
    } else if (camera.state.value.error) {
      emit('error', camera.state.value.error)
    }
  } finally {
    captureBusy.value = false
  }
}

function handleRetake() {
  camera.retake()
  emit('cleared')
}

function handleCancel() {
  camera.stopCamera()
}
</script>

<template>
  <div class="selfie-card">
    <div class="selfie-card-header">
      <h2>Visit Selfie</h2>
      <Tag v-if="required" value="Required" severity="danger" />
      <Tag v-else-if="showPreview" value="Captured" severity="success" />
    </div>

    <p class="selfie-hint">Align your face inside the frame. A timestamped watermark will be added automatically.</p>

    <div class="selfie-preview-area">
      <video
        v-show="showVideo"
        ref="videoEl"
        class="selfie-video"
        playsinline
        muted
        autoplay
        aria-label="Camera preview"
      />
      <canvas ref="canvasEl" class="selfie-canvas" />
      <img
        v-if="showPreview"
        :src="camera.previewUrl.value"
        class="selfie-captured"
        alt="Captured selfie"
      />
      <div v-if="showPlaceholder" class="selfie-placeholder">
        <i class="pi pi-camera" />
        <span>No photo yet</span>
      </div>
      <div v-if="showOpening" class="selfie-loading-overlay">
        <i class="pi pi-spin pi-sync" />
        <span>Starting camera...</span>
      </div>
    </div>

    <div v-if="camera.state.value.error" class="selfie-error">
      <i class="pi pi-exclamation-triangle" /> {{ camera.state.value.error }}
    </div>

    <div class="selfie-actions">
      <template v-if="camera.state.value.mode === 'idle' || camera.state.value.mode === 'error'">
        <Button label="Open Camera" icon="pi pi-camera" @click="handleOpenCamera" />
      </template>
      <template v-else-if="camera.state.value.mode === 'opening' || camera.state.value.mode === 'live'">
        <Button
          label="Capture Selfie"
          icon="pi pi-camera"
          severity="success"
          :disabled="!camera.cameraReady.value || captureBusy"
          :loading="captureBusy"
          @click="handleCapture"
        />
        <Button label="Cancel" icon="pi pi-times" severity="secondary" outlined @click="handleCancel" />
      </template>
      <template v-else-if="camera.state.value.mode === 'captured'">
        <Button label="Retake" icon="pi pi-refresh" severity="secondary" outlined @click="handleRetake" />
      </template>
    </div>

    <p v-if="showOpening || (camera.state.value.mode === 'live' && !camera.cameraReady.value)" class="selfie-status">
      Camera is still starting. Please wait a moment.
    </p>
  </div>
</template>

<style scoped>
.selfie-card { display: grid; gap: 0.65rem; }
.selfie-card-header { display: flex; align-items: center; gap: 0.5rem; }
.selfie-card-header h2 { margin: 0; flex: 1; }
.selfie-hint { margin: 0; color: var(--text-muted); font-size: 0.75rem; }

.selfie-preview-area {
  position: relative; width: 100%; aspect-ratio: 4/3;
  border-radius: 12px; overflow: hidden; background: #111827;
  border: 1px solid var(--border-light);
}
.selfie-video { width: 100%; height: 100%; display: block; object-fit: cover; transform: scaleX(-1); }
.selfie-canvas { display: none; }
.selfie-captured { width: 100%; height: 100%; display: block; object-fit: cover; transform: scaleX(-1); }
.selfie-placeholder {
  position: absolute; inset: 0; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 0.4rem;
  color: #64748b; font-size: 0.78rem;
}
.selfie-placeholder i { font-size: 1.8rem; }

.selfie-loading-overlay {
  position: absolute; inset: 0; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 0.4rem;
  background: rgba(17, 24, 39, 0.7); color: #fff; font-size: 0.82rem;
  font-weight: 600;
}
.selfie-loading-overlay i { font-size: 1.6rem; }

.selfie-error {
  display: flex; align-items: flex-start; gap: 0.4rem; padding: 0.6rem 0.8rem;
  border-radius: 10px; background: #fef2f2; color: #991b1b; font-size: 0.75rem; line-height: 1.4;
}
.selfie-error i { flex-shrink: 0; margin-top: 0.1rem; }

.selfie-actions { display: flex; gap: 0.5rem; }
.selfie-actions :deep(.p-button) { flex: 1; }

.selfie-status { margin: 0; text-align: center; color: var(--text-muted); font-size: 0.72rem; font-style: italic; }
</style>
