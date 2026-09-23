import axios from 'axios'

export interface DebugApiFailure {
  id: number
  method: string
  url: string
  status: number | null
  message: string
  backendMessage: string
  time: string
}

const failures: DebugApiFailure[] = []
const listeners = new Set<() => void>()
let sequence = 0

function notify() { listeners.forEach((listener) => listener()) }

export function recordApiFailure(error: unknown) {
  if (!axios.isAxiosError(error)) return
  const responseData = error.response?.data as { error?: { message?: string }; message?: string } | undefined
  failures.unshift({
    id: ++sequence,
    method: (error.config?.method || 'GET').toUpperCase(),
    url: error.config?.url || 'Unknown URL',
    status: error.response?.status ?? null,
    message: error.message || 'Request failed',
    backendMessage: responseData?.error?.message || responseData?.message || '',
    time: new Date().toLocaleTimeString(),
  })
  if (failures.length > 30) failures.pop()
  notify()
}

export function getApiFailures() { return failures }
export function subscribeDebug(listener: () => void) { listeners.add(listener); return () => listeners.delete(listener) }
export function clearDebugFailures() { failures.splice(0); notify() }
