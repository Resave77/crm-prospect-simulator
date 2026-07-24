import { ref, onBeforeUnmount } from 'vue'
import { haversineKm, formatDistance } from '../../utils/maps'

export interface GpsCoords {
  latitude: number
  longitude: number
  accuracy: number
  capturedAt: number
}

export interface LocationState {
  coords: GpsCoords | null
  permissionGranted: boolean
  loading: boolean
  error: string
}

export function useVisitLocation() {
  const state = ref<LocationState>({
    coords: null,
    permissionGranted: false,
    loading: false,
    error: '',
  })

  let watchId: number | null = null

  function distanceTo(targetLat: number, targetLng: number): number | null {
    if (!state.value.coords) return null
    return haversineKm(targetLat, targetLng, state.value.coords.latitude, state.value.coords.longitude) * 1000
  }

  function distanceFormatted(targetLat: number, targetLng: number): string {
    const d = distanceTo(targetLat, targetLng)
    return d !== null ? formatDistance(d / 1000) : '—'
  }

  function isInsideRadius(targetLat: number, targetLng: number, radiusMeters: number): boolean {
    const d = distanceTo(targetLat, targetLng)
    return d !== null && d <= radiusMeters
  }

  function refreshOnce(): Promise<GpsCoords> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('GPS is unavailable in this browser.'))
        return
      }
      state.value.loading = true
      state.value.error = ''
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          const coords: GpsCoords = {
            latitude: pos.coords.latitude,
            longitude: pos.coords.longitude,
            accuracy: pos.coords.accuracy,
            capturedAt: Date.now(),
          }
          state.value.coords = coords
          state.value.permissionGranted = true
          state.value.loading = false
          resolve(coords)
        },
        (err) => {
          state.value.loading = false
          state.value.permissionGranted = false
          if (err.code === err.PERMISSION_DENIED) {
            state.value.error = 'Location permission denied. Please enable location access in your browser settings.'
            reject(new Error(state.value.error))
          } else if (err.code === err.TIMEOUT) {
            state.value.error = 'GPS timed out. Please try again.'
            reject(new Error(state.value.error))
          } else {
            state.value.error = 'Unable to determine your location.'
            reject(new Error(state.value.error))
          }
        },
        { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 },
      )
    })
  }

  function startWatching() {
    if (!navigator.geolocation) return
    watchId = navigator.geolocation.watchPosition(
      (pos) => {
        state.value.coords = {
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
          accuracy: pos.coords.accuracy,
          capturedAt: Date.now(),
        }
        state.value.permissionGranted = true
        state.value.error = ''
      },
      () => {},
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 5000 },
    )
  }

  function stopWatching() {
    if (watchId !== null && navigator.geolocation) {
      navigator.geolocation.clearWatch(watchId)
      watchId = null
    }
  }

  function reset() {
    state.value = { coords: null, permissionGranted: false, loading: false, error: '' }
  }

  onBeforeUnmount(stopWatching)

  return {
    state,
    distanceTo,
    distanceFormatted,
    isInsideRadius,
    refreshOnce,
    startWatching,
    stopWatching,
    reset,
  }
}
