import { computed, nextTick, onBeforeUnmount, ref, type Ref } from 'vue'
import { applySelfieWatermark, type WatermarkMeta } from '../../utils/selfieWatermark'

export type CameraMode = 'idle' | 'opening' | 'live' | 'captured' | 'error'

export interface CameraOptions {
  videoEl: Ref<HTMLVideoElement | null>
  canvasEl: Ref<HTMLCanvasElement | null>
}

export interface CameraState {
  mode: CameraMode
  error: string
}

export function useVisitCamera(options: CameraOptions) {
  const { videoEl, canvasEl } = options

  const state = ref<CameraState>({
    mode: 'idle',
    error: '',
  })

  let activeStream: MediaStream | null = null
  const capturedFile = ref<File | null>(null)
  const previewUrl = ref('')

  const cameraReady = computed(() =>
    state.value.mode === 'live' &&
    videoEl.value != null &&
    videoEl.value.videoWidth > 0 &&
    videoEl.value.videoHeight > 0 &&
    videoEl.value.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA,
  )

  function mapCameraError(err: unknown): string {
    if (err instanceof DOMException) {
      switch (err.name) {
        case 'NotAllowedError':
          return 'Camera permission was denied. Please allow camera access in your browser settings.'
        case 'NotFoundError':
          return 'No camera was found on this device.'
        case 'NotReadableError':
          return 'Camera is being used by another application.'
        case 'SecurityError':
          return 'Camera requires HTTPS or localhost.'
        case 'AbortError':
          return 'Camera could not be started.'
        default:
          return `Camera error: ${err.message || 'Unable to access camera.'}`
      }
    }
    return 'Unable to access camera.'
  }

  function waitForVideoReady(video: HTMLVideoElement): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      if (
        video.readyState >= HTMLMediaElement.HAVE_METADATA &&
        video.videoWidth > 0 &&
        video.videoHeight > 0
      ) {
        resolve()
        return
      }

      let cleaned = false
      const cleanup = () => {
        if (cleaned) return
        cleaned = true
        window.clearTimeout(timeoutId)
        video.removeEventListener('loadedmetadata', handleReady)
        video.removeEventListener('canplay', handleReady)
        video.removeEventListener('error', handleError)
      }

      const timeoutId = window.setTimeout(() => {
        cleanup()
        reject(new Error('Camera preview timed out.'))
      }, 10000)

      const handleReady = () => {
        if (video.videoWidth <= 0 || video.videoHeight <= 0) return
        cleanup()
        resolve()
      }

      const handleError = () => {
        cleanup()
        reject(new Error('Camera preview failed to load.'))
      }

      video.addEventListener('loadedmetadata', handleReady)
      video.addEventListener('canplay', handleReady)
      video.addEventListener('error', handleError)
    })
  }

  async function openCamera() {
    state.value.error = ''

    if (!window.isSecureContext) {
      state.value.error = 'Camera access requires HTTPS or localhost.'
      state.value.mode = 'error'
      return false
    }

    if (!navigator.mediaDevices?.getUserMedia) {
      state.value.error = 'Camera is not supported by this browser.'
      state.value.mode = 'error'
      return false
    }

    stopCamera()

    state.value.mode = 'opening'

    await nextTick()

    const video = videoEl.value
    if (!video) {
      state.value.error = 'Camera preview could not be mounted.'
      state.value.mode = 'error'
      return false
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: 'user' }, width: { ideal: 1280 }, height: { ideal: 720 } },
        audio: false,
      })

      activeStream = stream

      for (const track of stream.getTracks()) {
        track.addEventListener('ended', () => {
          if (activeStream === stream) {
            state.value.mode = state.value.mode === 'live' ? 'idle' : state.value.mode
            state.value.error = 'Camera was disconnected.'
          }
        })
      }

      video.srcObject = stream
      video.muted = true
      video.playsInline = true

      await waitForVideoReady(video)
      await video.play()

      state.value.mode = 'live'
      return true
    } catch (err) {
      stopCamera()
      state.value.error = mapCameraError(err)
      state.value.mode = 'error'
      return false
    }
  }

  async function capturePhoto(meta: WatermarkMeta): Promise<File | null> {
    const video = videoEl.value
    const canvas = canvasEl.value
    if (!video || !canvas || state.value.mode !== 'live') return null

    if (video.videoWidth <= 0 || video.videoHeight <= 0 || video.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) {
      state.value.error = 'Camera frame is not available yet. Please wait a moment.'
      return null
    }

    canvas.width = video.videoWidth
    canvas.height = video.videoHeight

    const ctx = canvas.getContext('2d')
    if (!ctx) {
      state.value.error = 'Unable to capture camera frame.'
      return null
    }

    ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
    applySelfieWatermark(ctx, canvas.width, canvas.height, meta)

    const blob = await new Promise<Blob>((resolve, reject) => {
      canvas.toBlob(
        (value) => (value ? resolve(value) : reject(new Error('Unable to create selfie image.'))),
        'image/jpeg',
        0.85,
      )
    })

    const file = new File([blob], `visit-selfie-${Date.now()}.jpg`, {
      type: 'image/jpeg',
      lastModified: Date.now(),
    })

    capturedFile.value = file
    previewUrl.value = URL.createObjectURL(file)

    stopCamera()
    state.value.mode = 'captured'
    return file
  }

  function retake() {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    capturedFile.value = null
    previewUrl.value = ''
    state.value.mode = 'idle'
    state.value.error = ''
  }

  function stopCamera() {
    if (activeStream) {
      activeStream.getTracks().forEach((t) => t.stop())
      activeStream = null
    }
    if (videoEl.value) {
      videoEl.value.srcObject = null
    }
    if (state.value.mode === 'live' || state.value.mode === 'opening') {
      state.value.mode = 'idle'
    }
  }

  function reset() {
    stopCamera()
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    capturedFile.value = null
    previewUrl.value = ''
    state.value = { mode: 'idle', error: '' }
  }

  onBeforeUnmount(() => {
    stopCamera()
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  })

  return {
    state,
    cameraReady,
    capturedFile,
    previewUrl,
    openCamera,
    capturePhoto,
    retake,
    stopCamera,
    reset,
  }
}
